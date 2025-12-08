package go_coin_eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	chain "github.com/pefish/go-coin-eth/util"
	go_decimal "github.com/pefish/go-decimal"
	go_format "github.com/pefish/go-format"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_random "github.com/pefish/go-random"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
)

type Wallet struct {
	RemoteRpcClient *ethclient.Client
	RemoteWsClient  *ethclient.Client
	timeout         time.Duration
	chainId         *big.Int
	RpcClient       *rpc.Client
	WsClient        *rpc.Client
	logger          i_logger.ILogger
}

func NewWallet(logger i_logger.ILogger) *Wallet {
	timeout := 60 * time.Second
	return &Wallet{
		timeout: timeout,
		logger:  logger,
	}
}

type UrlParam struct {
	RpcUrl string
	WsUrl  string
}

func (w *Wallet) InitRemote(urlParam *UrlParam) (wallet_ *Wallet, err_ error) {
	if urlParam.RpcUrl == "" {
		return nil, errors.New("rpc url must be set")
	}

	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	rpcClient, err := rpc.DialContext(ctx, urlParam.RpcUrl)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	remoteRpcClient := ethclient.NewClient(rpcClient)

	ctx, _ = context.WithTimeout(context.Background(), w.timeout)
	chainId, err := remoteRpcClient.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	w.RemoteRpcClient = remoteRpcClient
	w.chainId = chainId
	w.RpcClient = rpcClient
	if urlParam.WsUrl != "" {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		wsClient, err := rpc.DialContext(ctx, urlParam.WsUrl)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		remoteWsClient := ethclient.NewClient(wsClient)
		w.RemoteWsClient = remoteWsClient
		w.WsClient = wsClient
	}
	return w, nil
}

func (w *Wallet) Close() {
	if w.RemoteRpcClient != nil {
		w.RemoteRpcClient.Close()
	}
	if w.RpcClient != nil {
		w.RpcClient.Close()
	}
	if w.RemoteWsClient != nil {
		w.RemoteWsClient.Close()
	}
	if w.WsClient != nil {
		w.WsClient.Close()
	}

}

func (w *Wallet) CallContractConstant(
	out any,
	contractAddress common.Address,
	abiStr,
	methodName string,
	opts *bind.CallOpts,
	params []any,
) error {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return errors.Wrap(err, "")
	}
	inputParams, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return errors.Wrap(err, "")
	}

	method, ok := parsedAbi.Methods[methodName]
	if !ok {
		return errors.New("Method not found.")
	}
	return w.CallContractConstantWithPayload(out, contractAddress, hex.EncodeToString(inputParams), method.Outputs, opts)
}

func (w *Wallet) CallContractConstantWithPayload(
	out any,
	contractAddress common.Address,
	payload string,
	outputTypes abi.Arguments,
	opts *bind.CallOpts,
) error {
	var realOpts bind.CallOpts
	if opts != nil {
		realOpts = *opts
	}

	payload = strings.TrimPrefix(payload, "0x")
	payloadBuf, err := hex.DecodeString(payload)
	if err != nil {
		return errors.Wrap(err, "")
	}
	var (
		msg    = ethereum.CallMsg{From: realOpts.From, To: &contractAddress, Data: payloadBuf}
		ctx    = realOpts.Context
		code   []byte
		output []byte
	)
	if ctx == nil {
		ctxTemp, _ := context.WithTimeout(context.Background(), w.timeout)
		ctx = ctxTemp
	}
	if realOpts.Pending {
		pb := bind.PendingContractCaller(w.RemoteRpcClient)
		output, err = pb.PendingCallContract(ctx, msg)
		if err == nil && len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = pb.PendingCodeAt(ctx, contractAddress); err != nil {
				return errors.Wrap(err, "")
			} else if len(code) == 0 {
				return errors.Wrap(bind.ErrNoCode, "")
			}
		}
	} else {
		output, err = bind.ContractCaller(w.RemoteRpcClient).CallContract(ctx, msg, realOpts.BlockNumber)
		if err != nil {
			return errors.Wrap(err, "")
		}
		if len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			code, err = bind.ContractCaller(w.RemoteRpcClient).CodeAt(ctx, contractAddress, realOpts.BlockNumber)
			if err != nil {
				return errors.Wrap(err, "")
			}
			if len(code) == 0 {
				return errors.Wrap(bind.ErrNoCode, "")
			}
		}
	}

	datas, err := outputTypes.Unpack(output)
	if err != nil {
		return errors.Wrap(err, "")
	}
	err = outputTypes.Copy(out, datas)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

/*
*
是个异步方法

只能获取以后的而且区块确认了的事件，即使 start 指定为过去的 block number，也不能获取到

query 的第一个 []any 是指第一个 index ，第二个是指第二个 index
*/
func (w *Wallet) WatchLogsByWs(
	ctx context.Context,
	contractAddress,
	abiStr,
	eventName string,
	opts *bind.WatchOpts,
	query ...[]any,
) (
	resultChan_ chan types.Log,
	contractInstance_ *bind.BoundContract,
	err_ error,
) {
	if w.RemoteWsClient == nil || w.WsClient == nil {
		return nil, nil, errors.New("please set ws url")
	}
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	contractInstance := bind.NewBoundContract(
		common.HexToAddress(contractAddress),
		parsedAbi,
		w.RemoteWsClient,
		w.RemoteWsClient,
		w.RemoteWsClient,
	)
	resultChan := make(chan types.Log)

	go func() {
	retry:
		for {
			chanLog, sub, err := contractInstance.WatchLogs(opts, eventName, query...)
			if err != nil {
				w.logger.WarnF("connect failed, reconnect after 3s. err -> %#v", err)
				time.Sleep(3 * time.Second)
				continue
			}
			w.logger.InfoF("connected. watching <%s>...", eventName)
			for {
				select {
				case log := <-chanLog:
					resultChan <- log
				case err := <-sub.Err():
					w.logger.WarnF("connection closed. err -> %#v", err)
					if err == nil { // 自己主动关闭的
						return
					}
					w.logger.Error(err)
					sub.Unsubscribe()
					time.Sleep(3 * time.Second)
					w.logger.Info("reconnect...")
					continue retry
				case <-ctx.Done():
					w.logger.Error(ctx.Err())
					return
				}
			}
		}
	}()
	return resultChan, contractInstance, nil
}

// 通过 CREATE 指令（普通合约部署）
func (w *Wallet) PredictCreateContractAddress(
	address common.Address,
	nonce uint64,
) string {
	hash := sha3.NewLegacyKeccak256()
	b, _ := rlp.EncodeToBytes([]any{
		address,
		nonce,
	})
	hash.Write(b)

	return common.HexToAddress("0x" + hex.EncodeToString(hash.Sum(nil)[12:])).String()
}

func keccak256(data []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// 通过 CREATE2 指令（可预测合约地址）
func (w *Wallet) PredictCreate2ContractAddress(
	deployer string,
	salt string, // 32 字节 salt
	bytecode string,
) (string, error) {
	// 解析输入
	deployerBytes, err := hex.DecodeString(deployer[2:]) // 去掉 "0x"
	if err != nil {
		return "", err
	}
	saltBytes, err := hex.DecodeString(salt[2:])
	if err != nil {
		return "", err
	}
	bytecodeBytes, err := hex.DecodeString(bytecode[2:])
	if err != nil {
		return "", err
	}

	// 计算合约字节码哈希
	bytecodeHash := keccak256(bytecodeBytes)

	// 构造 CREATE2 的输入数据: 0xff || deployer || salt || keccak256(bytecode)
	prefix := []byte{0xff}
	data := append(prefix, deployerBytes...)
	data = append(data, saltBytes...)
	data = append(data, bytecodeHash...)

	// 计算合约地址
	contractHash := keccak256(data)
	contractAddress := contractHash[12:] // 取后 20 字节

	// 转换为以太坊地址格式
	return "0x" + hex.EncodeToString(contractAddress), nil
}

func (w *Wallet) WatchLogsByLoop(
	ctx context.Context,
	logComming func(boundContract *bind.BoundContract, log types.Log) error,
	loopInterval time.Duration,
	startFromBlock *big.Int,
	contractAddress,
	abiStr,
	eventName string,
	query ...[]any,
) error {
	fromBlock := startFromBlock
	if startFromBlock == nil {
		latestBlockNumber, err := w.LatestBlockNumber()
		if err != nil {
			return err
		}
		fromBlock = go_decimal.MustStart(latestBlockNumber).MustSub(1000).MustEndForBigInt()
	}

	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			toBlock, err := w.LatestBlockNumber()
			if err != nil {
				return err
			}
			w.logger.DebugF("Find logs... fromBlock: %s, toBlock: %s", fromBlock, toBlock)
			err = w.FindLogs(
				func(contract *bind.BoundContract, logs []types.Log) error {
					fromBlock = go_decimal.MustStart(toBlock).MustAdd(1).MustEndForBigInt()
					for _, log := range logs {
						err := logComming(contract, log)
						if err != nil {
							return errors.Wrap(err, "")
						}
					}
					return nil
				},
				contractAddress,
				abiStr,
				eventName,
				fromBlock,
				toBlock,
				4900,
				query...,
			)
			if err != nil {
				w.logger.Error(err)
			}

			timer.Reset(loopInterval)
			continue
		}
	}
}

/*
查找历史的已经确认的事件，但不能实时接受后面的事件。取不到 pending 中的 logs

fromBlock；nil 就是 最新块号 - maxRange，负数就是 pending ，正数就是 blockNumber，range 太大自动分组处理
toBlock；nil 就是 latest ，负数就是 pending ，正数就是 blockNumber
*/
func (w *Wallet) FindLogs(
	logsComming func(contractInstance *bind.BoundContract, logs []types.Log) error,
	contractAddress,
	abiStr,
	eventName string,
	fromBlock,
	toBlock *big.Int,
	maxRange uint64,
	query ...[]any,
) error {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return errors.Wrap(err, "")
	}

	query = append([][]any{{parsedAbi.Events[eventName].ID}}, query...)

	topics, err := abi.MakeTopics(query...)
	if err != nil {
		return errors.Wrap(err, "")
	}

	if fromBlock == nil && toBlock != nil {
		fromBlock = toBlock.Sub(toBlock, new(big.Int).SetUint64(maxRange))
	} else if fromBlock == nil && toBlock == nil {
		latestBlockNumber, err := w.LatestBlockNumber()
		if err != nil {
			return errors.Wrap(err, "")
		}
		fromBlock = latestBlockNumber.Sub(latestBlockNumber, new(big.Int).SetUint64(maxRange))
		toBlock = latestBlockNumber
	} else if fromBlock != nil && toBlock == nil {
		latestBlockNumber, err := w.LatestBlockNumber()
		if err != nil {
			return errors.Wrap(err, "")
		}
		toBlock = latestBlockNumber
	}

	boundContract := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteRpcClient, w.RemoteRpcClient, w.RemoteRpcClient)

	_fromBlock := fromBlock
	_toBlock := fromBlock
	for {
		if go_decimal.MustStart(_toBlock).MustEq(toBlock) {
			break
		}
		_fromBlock = _toBlock
		if maxRange == 0 {
			_toBlock = toBlock
		} else if go_decimal.MustStart(toBlock).MustSub(_toBlock).MustGt(maxRange) {
			_toBlock = go_decimal.MustStart(_toBlock).MustAdd(maxRange).MustEndForBigInt()
		} else {
			_toBlock = toBlock
		}
		w.logger.DebugF("_fromBlock: %s, _toBlock: %s, remain: %s", _fromBlock.String(), _toBlock.String(), go_decimal.MustStart(toBlock).MustSubForString(_toBlock))
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		logs, err := w.RemoteRpcClient.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: _fromBlock,
			ToBlock:   _toBlock,
			Addresses: []common.Address{
				common.HexToAddress(contractAddress),
			},
			Topics: topics,
		})
		if err != nil {
			return errors.Wrap(err, "")
		}
		err = logsComming(boundContract, logs)
		if err != nil {
			return err
		}
	}
	return nil
}

// MaxTipPerGas: 小费价格的最大值，因为矿工会直接取最高值，所以这个值其实就是小费（etherscan 上显示的 Max Priority）
// a1 = 基础价格（是动态的由网络决定的，etherscan 上显示的 Base）+ MaxTipPerGas
// 最终采用的 price = min(a1, MaxFeePerGas)
type CallMethodOpts struct {
	Nonce          uint64
	Value          *big.Int
	GasPrice       *big.Int // only for legacy tx
	MaxFeePerGas   *big.Int // 这里指定采用价格的最高值（etherscan 上显示的 Max）
	GasLimit       uint64
	IsPredictError bool
	GasAccelerate  float64
}

type BuildTxResult struct {
	SignedTx *types.Transaction
	TxHex    string
}

// 通过字段名识别的，而不是 json tag
//
// out 参数必须将 event 中的每个参数都声明，event 中如果没有参数名，则使用 Arg<index> 代替
//
// 比如 event PairCreated(address indexed token0, address indexed token1, address pair, uint);
//
// out 应该是
//
//	var pairCreatedEvent struct {
//		Token0 common.Address
//		Token1 common.Address
//		Pair   common.Address
//		Arg3   *big.Int
//	}
func (w *Wallet) UnpackLog(
	out any,
	abiStr string,
	event string,
	log *types.Log,
) error {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return errors.Wrap(err, "")
	}
	boundContract := bind.NewBoundContract(
		common.HexToAddress(""),
		parsedAbi,
		w.RemoteRpcClient,
		w.RemoteRpcClient,
		w.RemoteRpcClient,
	)
	err = boundContract.UnpackLog(out, event, *log)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

type TransferLogData struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

func (w *Wallet) UnpackTransferLog(log *types.Log) (*TransferLogData, error) {
	var transferData TransferLogData
	err := w.UnpackLog(
		&transferData,
		Erc20AbiStr,
		"Transfer",
		log,
	)

	if err != nil {
		return nil, err
	}

	return &transferData, nil
}

func (w *Wallet) FilterLogs(
	topic0Hex string,
	logAddress common.Address,
	logs []*types.Log,
) ([]*types.Log, error) {
	results := make([]*types.Log, 0)
	for _, log := range logs {
		if log.Address.Cmp(logAddress) != 0 {
			continue
		}
		if topic0Hex != "" && log.Topics[0].Cmp(common.HexToHash(topic0Hex)) != 0 {
			continue
		}
		results = append(results, log)
	}

	return results, nil
}

func (w *Wallet) FilterLogsByEventNames(
	eventNames []string,
	abiStr string,
	logs []*types.Log,
) ([]*types.Log, error) {
	results := make([]*types.Log, 0)

	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	for _, log := range logs {
		for _, eventName := range eventNames {
			if log.Topics[0] == parsedAbi.Events[eventName].ID {
				continue
			}
		}

		results = append(results, log)
	}

	return results, nil
}

func (w *Wallet) UnpackParamsToStrs(
	types_ []abi.Type,
	paramsStr string,
) ([]string, error) {
	params, err := w.UnpackParams(types_, paramsStr)
	if err != nil {
		return nil, err
	}
	results := make([]string, 0)
	for _, param := range params {
		switch p := param.(type) {
		case common.Address:
			results = append(results, p.String())
		case []byte:
			results = append(results, hex.EncodeToString(p))
		case *big.Int:
			results = append(results, p.String())
		default:
			results = append(results, go_format.ToString(param))
		}
	}
	return results, nil
}

// payload 除了 methodId 就是 params
func (w *Wallet) UnpackParams(
	types_ []abi.Type,
	paramsStr string,
) ([]any, error) {
	paramsStr = strings.TrimPrefix(paramsStr, "0x")

	argTypes := make(abi.Arguments, 0)
	for _, t := range types_ {
		argTypes = append(argTypes, abi.Argument{
			Name:    "",
			Type:    t,
			Indexed: false,
		})
	}

	data, err := hex.DecodeString(paramsStr)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	datas, err := argTypes.Unpack(data)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return datas, nil
}

func (w *Wallet) PackParamsFromStrs(
	types []abi.Type,
	strs []string,
) (hexStr_ string, err_ error) {
	args := make([]any, 0)
	for i, str := range strs {
		switch types[i].String() {
		case "string":
			args = append(args, str)
		case "bool":
			args = append(args, go_format.MustToBool(str))
		case "bytes":
			str = strings.TrimPrefix(str, "0x")
			b, err := hex.DecodeString(str)
			if err != nil {
				return "", err
			}
			args = append(args, b)
		case "address":
			args = append(args, common.HexToAddress(str))
		default:
			if !strings.Contains(types[i].String(), "int") {
				return "", errors.Errorf("Type <%s> not be supported", types[i].String())
			}
			args = append(args, go_decimal.MustStart(str).MustEndForBigInt())
		}
	}
	return w.PackParams(types, args)
}

// 不带 0x 前缀
func (w *Wallet) PackParams(
	types []abi.Type,
	args []any,
) (hexStr_ string, err_ error) {
	argTypes := make(abi.Arguments, 0)
	for _, t := range types {
		argTypes = append(argTypes, abi.Argument{
			Name:    "",
			Type:    t,
			Indexed: false,
		})
	}

	bytes_, err := argTypes.Pack(args...)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	return hex.EncodeToString(bytes_), nil
}

// 不带 0x 前缀
func (w *Wallet) EncodePayload(
	abiStr string,
	methodName string,
	params []any,
) (hexStr_ string, err_ error) {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	input, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	return hex.EncodeToString(input), nil
}

func (w *Wallet) ToTopicHash(
	data any,
) (*common.Hash, error) {
	hashes, err := abi.MakeTopics([]any{data})
	if err != nil {
		return nil, err
	}
	return &hashes[0][0], nil
}

func (w *Wallet) ToTopicHashes(
	datas ...any,
) ([]common.Hash, error) {
	hashes, err := abi.MakeTopics(datas)
	if err != nil {
		return nil, err
	}
	return hashes[0], nil
}

func (w *Wallet) DecodePayload(
	abiStr string,
	out any,
	payloadStr string,
) (methodInfo_ *abi.Method, err_ error) {
	if len(payloadStr) < 8 {
		return nil, errors.New("payloadStr error")
	}
	payloadStr = strings.TrimPrefix(payloadStr, "0x")
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	data, err := hex.DecodeString(payloadStr)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	method, err := parsedAbi.MethodById(data[:4])
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if len(data[4:]) > 0 {
		a, err := method.Inputs.Unpack(data[4:])
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		err = method.Inputs.Copy(out, a)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}
	return method, err
}

// 得到事件签名的 hash，也就是 topic0
func (w *Wallet) Topic0FromEventName(abiStr, eventName string) (topic0_ string, err_ error) {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", errors.Wrap(err, "")
	}

	return parsedAbi.Events[eventName].ID.String(), nil
}

/*
返回结果不带 0x 前缀
*/
func (w *Wallet) MethodIdFromMethodStr(methodStr string) string {
	return hex.EncodeToString(crypto.Keccak256([]byte(methodStr))[:4])
}

func (w *Wallet) MethodFromPayload(
	abiStr string,
	payloadStr string,
) (methodInfo_ *abi.Method, err_ error) {
	if len(payloadStr) < 8 {
		return nil, errors.New("payloadStr error")
	}
	payloadStr = strings.TrimPrefix(payloadStr, "0x")
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	data, err := hex.DecodeString(payloadStr)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	method, err := parsedAbi.MethodById(data[:4])
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return method, err
}

func (w *Wallet) SuggestGasPrice(gasAccelerate float64) (gasPrice_ *big.Int, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	gasPrice, err := w.RemoteRpcClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to suggest gas price.")
	}
	if gasAccelerate == 0 {
		return gasPrice, nil
	}
	return go_decimal.MustStart(gasPrice).MustMulti(gasAccelerate).Round(0).MustEndForBigInt(), nil
}

func (w *Wallet) GasPrice() (gasPrice_ *big.Int, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	gasPrice, err := w.RemoteRpcClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to suggest gas price.")
	}
	return gasPrice, nil
}

func (w *Wallet) GasPriceNoDecimals() (gasPrice_ float64, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	gasPrice, err := w.RemoteRpcClient.SuggestGasPrice(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to suggest gas price.")
	}
	return go_decimal.MustStart(gasPrice).MustUnShiftedBy(9).MustEndForFloat64(), nil
}

func (w *Wallet) LatestBlockNumber() (blockNumber_ *big.Int, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	number, err := w.RemoteRpcClient.BlockNumber(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get latest block number.")
	}
	return new(big.Int).SetUint64(number), nil
}

func (w *Wallet) EstimateGas(msg ethereum.CallMsg) (gasCount_ uint64, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	gasCount, err := w.RemoteRpcClient.EstimateGas(ctx, msg)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to estimate gas needed.")
	}
	return gasCount, nil
}

func (w *Wallet) PrivateKeyToAddress(privateKey string) (address_ common.Address, err_ error) {
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "")
	}
	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "")
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	return crypto.PubkeyToAddress(publicKeyECDSA), nil
}

func (w *Wallet) IsContract(address common.Address) (isContract_ bool, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	codeBytes, err := w.RemoteRpcClient.CodeAt(ctx, address, nil)
	if err != nil {
		return false, errors.Wrap(err, "")
	}
	if len(codeBytes) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (w *Wallet) BuildCallMethodTx(
	privateKey string,
	contractAddress common.Address,
	abiStr,
	methodName string, // 如果有同名重构函数，比如有 4 个重构函数 test()，则 abi 数组中第一个叫 test，第二个叫 test0，第三个叫 test1，第四个叫 test2
	opts *CallMethodOpts,
	params []any,
) (btr_ *BuildTxResult, err_ error) {
	var realOpts CallMethodOpts
	if opts != nil {
		realOpts = *opts
	}

	isContract, err := w.IsContract(contractAddress)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if !isContract {
		return nil, errors.Errorf("To address not contract. address: %s", contractAddress)
	}

	privateKey = strings.TrimPrefix(privateKey, "0x")

	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var value = big.NewInt(0)
	var gasLimit uint64 = realOpts.GasLimit
	var nonce uint64 = realOpts.Nonce
	var isPredictError = true
	if realOpts.Value != nil {
		value = realOpts.Value
	}
	if !realOpts.IsPredictError {
		isPredictError = false
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKeyECDSA)
	if nonce == 0 {
		nonce, err = w.NextNonce(fromAddress)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to retrieve account nonce.")
		}
	}
	input, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if gasLimit == 0 || isPredictError {
		msg := ethereum.CallMsg{
			From:  fromAddress,
			To:    &contractAddress,
			Value: value,
			Data:  input,
		}
		tempGasLimit, err := w.EstimateGas(msg)
		if err != nil {
			return nil, err
		}
		if gasLimit == 0 {
			gasLimit = tempGasLimit
		}
	}

	return w.buildTx(
		privateKeyECDSA,
		nonce,
		&contractAddress,
		value,
		gasLimit,
		input,
		realOpts.MaxFeePerGas,
		realOpts.GasAccelerate,
	)
}

type BuildDeployContractTxOpts struct {
	Nonce         uint64
	MaxFeePerGas  *big.Int
	GasLimit      uint64
	GasAccelerate float64
}

func (w *Wallet) BuildDeployContractTx(
	privateKey,
	abiStr,
	binHexStr string,
	opts *BuildDeployContractTxOpts,
	params []any,
) (btr_ *BuildTxResult, err_ error) {
	var realOpts BuildDeployContractTxOpts
	if opts != nil {
		realOpts = *opts
	}

	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	privateKey = strings.TrimPrefix(privateKey, "0x")
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var nonce uint64 = realOpts.Nonce

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKeyECDSA)
	if nonce == 0 {
		nonce, err = w.NextNonce(fromAddress)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to retrieve account nonce.")
		}
	}

	input, err := parsedAbi.Pack("", params...)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	data := append(common.FromHex(binHexStr), input...)

	value := big.NewInt(0)
	var gasLimit uint64 = realOpts.GasLimit
	if gasLimit == 0 {
		msg := ethereum.CallMsg{
			From:  fromAddress,
			To:    nil,
			Value: value,
			Data:  data,
		}
		tempGasLimit, err := w.EstimateGas(msg)
		if err != nil {
			return nil, err
		}
		gasLimit = uint64(float64(tempGasLimit) * 1.2)
	}

	return w.buildTx(
		privateKeyECDSA,
		nonce,
		nil,
		value,
		gasLimit,
		data,
		realOpts.MaxFeePerGas,
		realOpts.GasAccelerate,
	)
}

func (w *Wallet) EstimateCall(
	fromAddress common.Address,
	contractAddress common.Address,
	abiStr string,
	value string,
	methodName string,
	params []any,
) error {

	isContract, err := w.IsContract(contractAddress)
	if err != nil {
		return err
	}
	if !isContract {
		return errors.Errorf("To address not contract. address: %s", contractAddress)
	}

	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return errors.Wrap(err, "")
	}

	input, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return errors.Wrap(err, "")
	}
	msg := ethereum.CallMsg{
		From:  fromAddress,
		To:    &contractAddress,
		Value: go_decimal.MustStart(value).MustShiftedBy(18).MustEndForBigInt(),
		Data:  input,
	}
	_, err = w.EstimateGas(msg)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wallet) NextNonce(address common.Address) (nonce_ uint64, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	nonce, err := w.RemoteRpcClient.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to retrieve account nonce.")
	}
	return nonce, nil
}

func (w *Wallet) DecodeTxHex(txHex string) (*types.Transaction, error) {
	raw, err := hexutil.Decode(txHex)
	if err != nil {
		return nil, err
	}
	tx := new(types.Transaction)
	err = tx.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (w *Wallet) buildTx(
	privateKeyECDSA *ecdsa.PrivateKey,
	nonce uint64,
	toAddressObj *common.Address,
	value *big.Int,
	gasLimit uint64,
	data []byte,
	maxFeePerGas *big.Int,
	gasAccelerate float64,
) (btr_ *BuildTxResult, err_ error) {
	if gasLimit == 0 {
		return nil, errors.Errorf("GasLimit <%d> is illegal.", gasLimit)
	}

	var rawTx *types.Transaction

	maxTipPerGas := new(big.Int)
	if maxFeePerGas == nil {
		// 使用 maxTipPerGas 限制
		maxFeePerGas = new(big.Int)

		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		baseGasPrice, err := w.RemoteRpcClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to suggest gas price.")
		}
		// 直接设置为 1.2 倍，十倍都可以（但要求余额足够），因为是使用 maxTipPerGas 限制
		maxFeePerGas = go_decimal.MustStart(baseGasPrice).MustMulti(1.2).RoundDown(0).MustEndForBigInt()

		if gasAccelerate == 0 {
			gasAccelerate = 1.02
		}
		if gasAccelerate < 1 {
			return nil, errors.Errorf("GasAccelerate must larger than 1.")
		}
		diff := go_decimal.MustStart(gasAccelerate).MustSub(1).EndForString()
		maxTipPerGas = go_decimal.MustStart(baseGasPrice).MustMulti(diff).RoundDown(0).MustEndForBigInt()
	} else {
		// 使用 maxFeePerGas 限制
		maxTipPerGas = maxFeePerGas
	}

	rawTx = types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		To:        toAddressObj,
		Value:     value,
		Gas:       gasLimit,
		GasFeeCap: maxFeePerGas,
		GasTipCap: maxTipPerGas,
		Data:      data,
	})
	signedTx, err := types.SignTx(rawTx, types.LatestSignerForChainID(w.chainId), privateKeyECDSA)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	w.logger.DebugF("交易构建成功 <txId: %s>", signedTx.Hash().String())
	return &BuildTxResult{
		SignedTx: signedTx,
		TxHex:    hexutil.Encode(txBytes),
	}, nil
}

func (w *Wallet) buildLegacyTx(
	privateKeyECDSA *ecdsa.PrivateKey,
	nonce uint64,
	toAddressObj common.Address,
	value *big.Int,
	gasLimit uint64,
	data []byte,
	gasPrice *big.Int,
	gasAccelerate float64,
) (btr_ *BuildTxResult, err_ error) {
	var rawTx *types.Transaction
	if gasPrice == nil {
		_gasPrice, err := w.SuggestGasPrice(gasAccelerate)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to suggest gas price.")
		}
		gasPrice = _gasPrice
	}
	rawTx = types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddressObj,
		Value:    value,
		Data:     data,
	})
	signedTx, err := types.SignTx(rawTx, types.LatestSignerForChainID(w.chainId), privateKeyECDSA)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &BuildTxResult{
		SignedTx: signedTx,
		TxHex:    hexutil.Encode(txBytes),
	}, nil
}

func (w *Wallet) BuildCallMethodTxWithPayload(
	privateKey string,
	contractAddress common.Address,
	payload string,
	opts *CallMethodOpts,
) (btr_ *BuildTxResult, err_ error) {
	var realOpts CallMethodOpts
	if opts != nil {
		realOpts = *opts
	}

	isContract, err := w.IsContract(contractAddress)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if !isContract {
		return nil, errors.Errorf("To address not contract. address: %s", contractAddress)
	}
	privateKey = strings.TrimPrefix(privateKey, "0x")
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	payload = strings.TrimPrefix(payload, "0x")
	payloadBuf, err := hex.DecodeString(payload)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var value = big.NewInt(0)
	var gasLimit uint64 = realOpts.GasLimit
	var nonce uint64 = realOpts.Nonce
	var isPredictError = true
	if realOpts.Value != nil {
		value = realOpts.Value
	}
	if !realOpts.IsPredictError {
		isPredictError = false
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKeyECDSA)
	if nonce == 0 {
		nonce, err = w.NextNonce(fromAddress)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to retrieve account nonce.")
		}
	}

	if gasLimit == 0 || isPredictError {
		msg := ethereum.CallMsg{
			From:  fromAddress,
			To:    &contractAddress,
			Value: value,
			Data:  payloadBuf,
		}
		tempGasLimit, err := w.EstimateGas(msg)
		if err != nil {
			return nil, err
		}
		if gasLimit == 0 {
			gasLimit = tempGasLimit
		}
	}

	return w.buildTx(
		privateKeyECDSA,
		nonce,
		&contractAddress,
		value,
		gasLimit,
		payloadBuf,
		realOpts.MaxFeePerGas,
		realOpts.GasAccelerate,
	)
}

type BuildTransferTxOpts struct {
	CallMethodOpts
	Payload  []byte
	IsLegacy bool
}

func (w *Wallet) BuildTransferTx(
	privateKey string,
	toAddress common.Address,
	opts *BuildTransferTxOpts,
) (btr_ *BuildTxResult, err_ error) {
	var realOpts BuildTransferTxOpts
	if opts != nil {
		realOpts = *opts
	}

	privateKey = strings.TrimPrefix(privateKey, "0x")

	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var value = big.NewInt(0)

	var gasLimit uint64 = realOpts.GasLimit
	var nonce uint64 = realOpts.Nonce
	if realOpts.Value != nil {
		value = realOpts.Value
	}
	if gasLimit == 0 {
		gasLimit = 21000
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKeyECDSA)
	if nonce == 0 {
		nonce, err = w.NextNonce(fromAddress)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to retrieve account nonce.")
		}
	}

	if realOpts.IsLegacy {
		return w.buildLegacyTx(
			privateKeyECDSA,
			nonce,
			toAddress,
			value,
			gasLimit,
			realOpts.Payload,
			realOpts.GasPrice,
			realOpts.GasAccelerate,
		)
	}

	return w.buildTx(
		privateKeyECDSA,
		nonce,
		&toAddress,
		value,
		gasLimit,
		realOpts.Payload,
		realOpts.MaxFeePerGas,
		realOpts.GasAccelerate,
	)
}

func (w *Wallet) SendRawTransaction(txHex string) (hash_ string, err_ error) {
	var hash common.Hash
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	err := w.RpcClient.CallContext(ctx, &hash, "eth_sendRawTransaction", txHex)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	w.logger.DebugF("交易发送成功 <txId: %s>", hash.String())
	return hash.String(), nil
}

func (w *Wallet) SendRawTransactionWait(ctx context.Context, txHex string) (txReceipt_ *types.Receipt, err_ error) {
	hash, err := w.SendRawTransaction(txHex)
	if err != nil {
		return nil, err
	}
	return w.WaitConfirm(ctx, hash, time.Second)
}

type TransactionByHashResult struct {
	*types.Transaction
	From common.Address
}

func (w *Wallet) TransactionByHash(txHash string) (result_ *TransactionByHashResult, isPending_ bool, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	tx, isPending, err := w.RemoteRpcClient.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, false, errors.Wrap(err, "")
	}

	from, err := types.Sender(types.NewLondonSigner(w.chainId), tx)
	if err != nil {
		return nil, false, errors.Wrap(err, "")
	}
	return &TransactionByHashResult{
		tx,
		from,
	}, isPending, nil
}

func (w *Wallet) TransactionReceiptByHash(txHash string) (txReceipt_ *types.Receipt, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	receipt, err := w.RemoteRpcClient.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return receipt, nil
}

func (w *Wallet) WaitConfirm(ctx context.Context, txHash string, interval time.Duration) (txReceipt_ *types.Receipt, err_ error) {
	timer := time.NewTimer(0)
out:
	for {
		select {
		case <-timer.C:
			receipt, err := w.TransactionReceiptByHash(txHash)
			if err != nil {
				w.logger.DebugF("TransactionReceiptByHash: %s, hash: %s", err.Error(), txHash)
				timer.Reset(interval)
				continue
			}
			if receipt == nil ||
				receipt.BlockNumber == nil {
				timer.Reset(interval)
				continue
			}
			if receipt.Status == 0 {
				w.logger.DebugF("交易确认失败 <txId: %s>", txHash)
				return receipt, errors.Errorf("<TxId: %s>; Tx failed.", txHash)
			}
			w.logger.DebugF("交易确认成功 <txId: %s>", txHash)
			timer.Stop()
			return receipt, nil
		case <-ctx.Done():
			break out
		}
	}
	return nil, errors.Errorf("<TxId: %s>; Canceled wait.", txHash)
}

type Transaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"s"`
}

type TxsInPoolResult struct {
	Pending map[string]map[string]Transaction `json:"pending"`
	Queued  map[string]map[string]Transaction `json:"queued"`
}

func (w *Wallet) TxsInPool() (txs_ *TxsInPoolResult, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	var result TxsInPoolResult
	err := w.RpcClient.CallContext(ctx, &result, "txpool_content")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &result, nil
}

func (w *Wallet) Balance(address common.Address) (bal_ *big.Int, err_ error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	result, err := w.RemoteRpcClient.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return result, nil
}

func (w *Wallet) BalanceNoDecimals(address common.Address) (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	result, err := w.RemoteRpcClient.BalanceAt(ctx, address, nil)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	return go_decimal.MustStart(result).MustUnShiftedBy(18).EndForString(), nil
}

func (w *Wallet) ApprovedAmount(contractAddress, fromAddress, toAddress common.Address) (amount_ *big.Int, err_ error) {
	result := new(big.Int)
	err := w.CallContractConstant(
		&result,
		contractAddress,
		Erc20AbiStr,
		"allowance",
		nil,
		[]any{
			fromAddress,
			toAddress,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return result, nil
}

func (w *Wallet) Approve(
	priv string,
	contractAddress common.Address,
	toAddress common.Address,
	amount *big.Int,
	opts *CallMethodOpts,
) (hash_ string, err_ error) {
	approveAmount := amount
	if approveAmount == nil {
		approveAmount = MaxUint256
	}
	tx, err := w.BuildCallMethodTx(
		priv,
		contractAddress,
		Erc20AbiStr,
		"approve",
		opts,
		[]any{
			toAddress,
			approveAmount,
		},
	)
	if err != nil {
		return "", err
	}
	txHash, err := w.SendRawTransaction(tx.TxHex)
	if err != nil {
		return "", err
	}
	return txHash, nil
}

func (w *Wallet) ApproveWait(
	ctx context.Context,
	priv string,
	contractAddress common.Address,
	toAddress common.Address,
	amount *big.Int,
	opts *CallMethodOpts,
) (txReceipt_ *types.Receipt, err_ error) {
	hash, err := w.Approve(priv, contractAddress, toAddress, amount, opts)
	if err != nil {
		return nil, err
	}
	return w.WaitConfirm(ctx, hash, time.Second)
}

type DeriveFromPathResult struct {
	Address    common.Address `json:"address"`
	PublicKey  string         `json:"public_key"`
	PrivateKey string         `json:"private_key"`
}

// 都不带 0x 前缀
func (w *Wallet) DeriveFromPath(seed string, path string) (result_ *DeriveFromPathResult, err_ error) {
	// 字符串转换成字节数组
	seedBuf, err := hex.DecodeString(seed)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	// 通过种子生成 hdwallet
	wallet, err := chain.NewFromSeed(seedBuf)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	// 解析派生路径
	hdPath, err := chain.ParseDerivationPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	// 派生账号
	account, err := wallet.Derive(hdPath, true)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	// 获取私钥 hex 字符串
	privateKeyStr, err := wallet.PrivateKeyHex(account)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	privateKey, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	privateKeyECDSA, err := crypto.ToECDSA(privateKey[:])
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	publicKeyStr := hex.EncodeToString(crypto.CompressPubkey(&publicKeyECDSA))
	addr := crypto.PubkeyToAddress(publicKeyECDSA)
	return &DeriveFromPathResult{
		Address:    addr,
		PublicKey:  publicKeyStr,
		PrivateKey: privateKeyStr,
	}, nil
}

func (w *Wallet) SignMsg(privateKey string, data string) (hexStr_ string, err_ error) {
	privateKey = strings.TrimPrefix(privateKey, "0x")

	privateKeyHex, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	privateKeyObj, err := crypto.ToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}
	hash, err := w.SignHashForMsg(data)
	if err != nil {
		return "", err
	}
	hashBuf, _ := hex.DecodeString(hash)
	signature, err := crypto.Sign(hashBuf, privateKeyObj)
	if err != nil {
		return "", err
	}
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return hex.EncodeToString(signature), nil
}

func (w *Wallet) RecoverSignerAddress(msg, sig string) (address_ *common.Address, err_ error) {
	hash, err := w.SignHashForMsg(msg)
	if err != nil {
		return nil, err
	}
	return w.RecoverSignerAddressFromMsgHash(hash, sig)
}

func (w *Wallet) RecoverSignerAddressFromMsgHash(msgHash, sig string) (address_ *common.Address, err_ error) {
	sig = strings.TrimPrefix(sig, "0x")
	sigHex, err := hex.DecodeString(sig)
	if err != nil {
		return nil, err
	}
	msgHash = strings.TrimPrefix(msgHash, "0x")
	msgHashHex, err := hex.DecodeString(msgHash)
	if err != nil {
		return nil, err
	}

	if len(sigHex) != 65 {
		return nil, errors.Errorf("signature must be 65 bytes long")
	}
	if sigHex[64] != 27 && sigHex[64] != 28 {
		return nil, errors.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sigHex[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.Ecrecover(msgHashHex, sigHex)
	if err != nil {
		return nil, err
	}
	pubKey, err := crypto.UnmarshalPubkey(rpk)
	if err != nil {
		return nil, err
	}
	//pubKey := crypto.ToECDSAPub(rpk)
	//crypto.FromECDSAPub()
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return &recoveredAddr, nil
}

// 以太坊的 hash 专门在数据前面加上了一段话
func (w *Wallet) SignHashForMsg(data string) (hexStr_ string, err_ error) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return hex.EncodeToString(crypto.Keccak256([]byte(msg))), nil
}

type SendEthOpts struct {
	Nonce         uint64
	MaxFeePerGas  *big.Int
	GasAccelerate float64
}

func (w *Wallet) SendEth(
	priv string,
	address common.Address,
	amount string,
	opts *SendEthOpts,
) (hash_ string, err_ error) {
	callMethodOpts := CallMethodOpts{
		Value: go_decimal.MustStart(amount).MustShiftedBy(18).MustEndForBigInt(),
	}
	if opts != nil {
		callMethodOpts.Nonce = opts.Nonce
		callMethodOpts.GasAccelerate = opts.GasAccelerate
		callMethodOpts.MaxFeePerGas = opts.MaxFeePerGas
	}
	tx, err := w.BuildTransferTx(priv, address, &BuildTransferTxOpts{
		CallMethodOpts: callMethodOpts,
	})
	if err != nil {
		return "", err
	}
	txHash, err := w.SendRawTransaction(tx.TxHex)
	if err != nil {
		return "", err
	}
	return txHash, nil
}

// 一些二层网络，比如 Base，需要额外消耗一些 gas，所以需要 remainAmount 参数
func (w *Wallet) SendAllEthByLegacy(
	priv string,
	address common.Address,
	remainAmount string,
) (hash_ string, err_ error) {
	fromAddress, err := w.PrivateKeyToAddress(priv)
	if err != nil {
		return "", err
	}

	ethBal, err := w.Balance(fromAddress)
	if err != nil {
		return "", err
	}

	gasPrice, err := w.SuggestGasPrice(1.1)
	if err != nil {
		return "", err
	}
	gasLimit := 21000

	fee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	value := new(big.Int).Sub(ethBal, fee)
	value.Sub(value, go_decimal.MustStart(remainAmount).MustShiftedBy(18).MustEndForBigInt())

	tx, err := w.BuildTransferTx(priv, address, &BuildTransferTxOpts{
		CallMethodOpts: CallMethodOpts{
			GasPrice: gasPrice,
			GasLimit: uint64(gasLimit),
			Value:    value,
		},
		IsLegacy: true,
	})
	if err != nil {
		return "", err
	}
	txHash, err := w.SendRawTransaction(tx.TxHex)
	if err != nil {
		return "", err
	}
	return txHash, nil
}

func (w *Wallet) SendAllEth(
	priv string,
	address common.Address,
	remainAmount string,
) (hash_ string, err_ error) {
	fromAddress, err := w.PrivateKeyToAddress(priv)
	if err != nil {
		return "", err
	}

	ethBal, err := w.Balance(fromAddress)
	if err != nil {
		return "", err
	}

	maxFeePerGas, err := w.SuggestGasPrice(1.1)
	if err != nil {
		return "", err
	}
	gasLimit := 21000

	fee := new(big.Int).Mul(maxFeePerGas, big.NewInt(int64(gasLimit)))
	value := new(big.Int).Sub(ethBal, fee)
	value.Sub(value, go_decimal.MustStart(remainAmount).MustShiftedBy(18).MustEndForBigInt())

	tx, err := w.BuildTransferTx(priv, address, &BuildTransferTxOpts{
		CallMethodOpts: CallMethodOpts{
			MaxFeePerGas: maxFeePerGas,
			GasLimit:     uint64(gasLimit),
			Value:        value,
		},
	})
	if err != nil {
		return "", err
	}
	txHash, err := w.SendRawTransaction(tx.TxHex)
	if err != nil {
		return "", err
	}
	return txHash, nil
}

func (w *Wallet) SendEthWait(
	ctx context.Context,
	priv string,
	address common.Address,
	amount string,
	opts *SendEthOpts,
) (txReceipt_ *types.Receipt, err_ error) {
	hash, err := w.SendEth(priv, address, amount, opts)
	if err != nil {
		return nil, err
	}
	return w.WaitConfirm(ctx, hash, time.Second)
}

func (w *Wallet) SendAllToken(
	priv string,
	contractAddress common.Address,
	address common.Address,
	opts *CallMethodOpts,
) (amountWithDecimals_ *big.Int, hash_ string, err_ error) {
	fromAddress, err := w.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, "", err
	}
	bal, err := w.TokenBalance(contractAddress, fromAddress)
	if err != nil {
		return bal, "", err
	}
	if go_decimal.MustStart(bal).MustEq(0) {
		return bal, "", errors.Errorf("Balance not enough.")
	}
	hash, err := w.SendToken(priv, contractAddress, address, bal, opts)
	if err != nil {
		return bal, hash, err
	}
	return bal, hash, nil
}

func (w *Wallet) SendAllTokenWait(
	ctx context.Context,
	priv string,
	contractAddress common.Address,
	address common.Address,
	opts *CallMethodOpts,
) (amountWithDecimals_ *big.Int, txReceipt_ *types.Receipt, err_ error) {
	amountWithDecimals, hash, err := w.SendAllToken(priv, contractAddress, address, opts)
	if err != nil {
		return amountWithDecimals, nil, err
	}
	txr, err := w.WaitConfirm(ctx, hash, time.Second)
	if err != nil {
		return amountWithDecimals, txr, err
	}
	return amountWithDecimals, txr, nil
}

func (w *Wallet) SendToken(
	priv string,
	contractAddress common.Address,
	address common.Address,
	amount *big.Int,
	opts *CallMethodOpts,
) (hash_ string, err_ error) {
	tx, err := w.BuildCallMethodTx(
		priv,
		contractAddress,
		Erc20AbiStr,
		"transfer",
		opts,
		[]any{
			address,
			amount,
		},
	)
	if err != nil {
		return "", err
	}
	txHash, err := w.SendRawTransaction(tx.TxHex)
	if err != nil {
		return "", err
	}
	return txHash, nil
}

func (w *Wallet) SendTokenWait(
	ctx context.Context,
	priv string,
	contractAddress common.Address,
	address common.Address,
	amount *big.Int,
	opts *CallMethodOpts,
) (txReceipt_ *types.Receipt, err_ error) {
	hash, err := w.SendToken(priv, contractAddress, address, amount, opts)
	if err != nil {
		return nil, err
	}
	return w.WaitConfirm(ctx, hash, time.Second)
}

func (w *Wallet) TokenDecimals(tokenAddress common.Address) (decimals_ uint64, err_ error) {
	var result uint8
	err := w.CallContractConstant(
		&result,
		tokenAddress,
		`[{
      "inputs": [],
      "name": "decimals",
      "outputs": [
        {
          "internalType": "uint8",
          "name": "",
          "type": "uint8"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    }]`,
		"decimals",
		nil,
		nil,
	)
	if err != nil {
		return 0, err
	}
	return uint64(result), nil
}

func (w *Wallet) TokenBalanceNoDecimals(contractAddress, address common.Address) (bal_ string, err_ error) {
	result := new(big.Int)
	err := w.CallContractConstant(
		&result,
		contractAddress,
		`[{
    "inputs": [
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "balanceOf",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }]`,
		"balanceOf",
		nil,
		[]any{
			address,
		},
	)
	if err != nil {
		return "", err
	}
	decimals, err := w.TokenDecimals(contractAddress)
	if err != nil {
		return "", err
	}
	return go_decimal.MustStart(result).MustUnShiftedBy(decimals).EndForString(), nil
}

func (w *Wallet) TokenBalance(tokenAddress, address common.Address) (bal_ *big.Int, err_ error) {
	result := new(big.Int)
	err := w.CallContractConstant(
		&result,
		tokenAddress,
		`[{
    "inputs": [
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "balanceOf",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }]`,
		"balanceOf",
		nil,
		[]any{
			address,
		},
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Wallet) WatchPendingTxByWs(resultChan chan<- string) error {
	if w.RemoteWsClient == nil || w.WsClient == nil {
		return errors.New("please set ws url")
	}
	for {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		subscription, err := w.WsClient.EthSubscribe(ctx, resultChan, "newPendingTransactions")
		if err != nil {
			if subscription != nil {
				subscription.Unsubscribe()
			}
			return errors.Wrap(err, "")
		}
		w.logger.Info("connected. watching...")
		err = <-subscription.Err()
		w.logger.WarnF("connection closed. err -> %#v", err)
		if err == nil { // 自己主动关闭的
			return nil
		}
		subscription.Unsubscribe()
		w.logger.Info("reconnect...")
	}
}

func (w *Wallet) SeedHexByMnemonic(mnemonic string, pass string) (seed_ string) {
	return hex.EncodeToString(bip39.NewSeed(mnemonic, pass))
}

func (w *Wallet) RandomMnemonic() (mnemonic_ string, err_ error) {
	entropy, err := go_random.RandomBytes(16)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	return mnemonic, nil
}

func (w *Wallet) IsAddress(str string) bool {
	return common.IsHexAddress(str)
}

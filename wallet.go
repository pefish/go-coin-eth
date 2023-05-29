package go_coin_eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	chain "github.com/pefish/go-coin-eth/util"
	"github.com/pefish/go-decimal"
	"github.com/pefish/go-error"
	"github.com/pefish/go-http"
	"github.com/pefish/go-logger"
	"github.com/pefish/go-random"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"math/big"
	"strings"
	"time"
)

var (
	ScanApiUrl = "https://api.etherscan.io/api"

	Erc20AbiStr    = `[{"inputs":[{"internalType":"address","name":"operator","type":"address"},{"internalType":"address","name":"pauser","type":"address"},{"internalType":"string","name":"name","type":"string"},{"internalType":"string","name":"symbol","type":"string"},{"internalType":"uint8","name":"decimal","type":"uint8"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"spender","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"burn","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"new_operator","type":"address"},{"internalType":"address","name":"new_pauser","type":"address"}],"name":"changeUser","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"subtractedValue","type":"uint256"}],"name":"decreaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"addedValue","type":"uint256"}],"name":"increaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	ZeroAddressStr = "0x0000000000000000000000000000000000000000"
	ZeroAddress    = common.HexToAddress(ZeroAddressStr)
	OneAddressStr  = "0x0000000000000000000000000000000000000001"
	OneAddress     = common.HexToAddress(OneAddressStr)
)

var (
	MaxUint256, _ = new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

	TypeUint256, _    = abi.NewType("uint256", "", nil)
	TypeUint32, _     = abi.NewType("uint32", "", nil)
	TypeUint16, _     = abi.NewType("uint16", "", nil)
	TypeString, _     = abi.NewType("string", "", nil)
	TypeBool, _       = abi.NewType("bool", "", nil)
	TypeBytes, _      = abi.NewType("bytes", "", nil)
	TypeAddress, _    = abi.NewType("address", "", nil)
	TypeUint64Arr, _  = abi.NewType("uint64[]", "", nil)
	TypeAddressArr, _ = abi.NewType("address[]", "", nil)
	TypeInt8, _       = abi.NewType("int8", "", nil)
	// Special types for testing
	TypeUint32Arr2, _       = abi.NewType("uint32[2]", "", nil)
	TypeUint64Arr2, _       = abi.NewType("uint64[2]", "", nil)
	TypeUint256Arr, _       = abi.NewType("uint256[]", "", nil)
	TypeUint256Arr2, _      = abi.NewType("uint256[2]", "", nil)
	TypeUint256Arr3, _      = abi.NewType("uint256[3]", "", nil)
	TypeUint256ArrNested, _ = abi.NewType("uint256[2][2]", "", nil)
	TypeUint8ArrNested, _   = abi.NewType("uint8[][2]", "", nil)
	TypeUint8SliceNested, _ = abi.NewType("uint8[][]", "", nil)
	TypeTupleF, _           = abi.NewType("tuple", "struct Overloader.F", []abi.ArgumentMarshaling{
		{Name: "_f", Type: "uint256"},
		{Name: "__f", Type: "uint256"},
		{Name: "f", Type: "uint256"}})
)

type Wallet struct {
	RemoteRpcClient *ethclient.Client
	RemoteWsClient  *ethclient.Client
	timeout         time.Duration
	chainId         *big.Int
	RpcClient       *rpc.Client
	WsClient        *rpc.Client
	logger          go_logger.InterfaceLogger
}

func NewWallet() *Wallet {
	timeout := 60 * time.Second
	return &Wallet{
		timeout: timeout,
		logger:  go_logger.DefaultLogger,
	}
}

type UrlParam struct {
	RpcUrl string
	WsUrl  string
}

func (w *Wallet) InitRemote(urlParam UrlParam) (*Wallet, error) {
	if urlParam.RpcUrl == "" {
		return nil, errors.New("rpc url must be set")
	}

	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	rpcClient, err := rpc.DialContext(ctx, urlParam.RpcUrl)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	remoteRpcClient := ethclient.NewClient(rpcClient)

	ctx, _ = context.WithTimeout(context.Background(), w.timeout)
	chainId, err := remoteRpcClient.ChainID(ctx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	w.RemoteRpcClient = remoteRpcClient
	w.chainId = chainId
	w.RpcClient = rpcClient
	if urlParam.WsUrl != "" {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		wsClient, err := rpc.DialContext(ctx, urlParam.WsUrl)
		if err != nil {
			return nil, go_error.WithStack(err)
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

func (w *Wallet) SetLogger(logger go_logger.InterfaceLogger) {
	w.logger = logger
}

func (w *Wallet) CallContractConstant(out interface{}, contractAddress, abiStr, methodName string, opts *bind.CallOpts, params ...interface{}) error {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return go_error.WithStack(err)
	}
	inputParams, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return go_error.WithStack(err)
	}

	method, ok := parsedAbi.Methods[methodName]
	if !ok {
		return go_error.WithStack(errors.New("method not found"))
	}
	return w.CallContractConstantWithPayload(out, contractAddress, hex.EncodeToString(inputParams), method.Outputs, opts)
}

func (w *Wallet) CallContractConstantWithPayload(out interface{}, contractAddress, payload string, outputTypes abi.Arguments, opts *bind.CallOpts) error {
	if opts == nil {
		opts = new(bind.CallOpts)
	}

	contractAddressObj := common.HexToAddress(contractAddress)
	if strings.HasPrefix(payload, "0x") {
		payload = payload[2:]
	}
	payloadBuf, err := hex.DecodeString(payload)
	if err != nil {
		return go_error.WithStack(err)
	}
	var (
		msg    = ethereum.CallMsg{From: opts.From, To: &contractAddressObj, Data: payloadBuf}
		ctx    = opts.Context
		code   []byte
		output []byte
	)
	if ctx == nil {
		ctxTemp, _ := context.WithTimeout(context.Background(), w.timeout)
		ctx = ctxTemp
	}
	if opts.Pending {
		pb := bind.PendingContractCaller(w.RemoteRpcClient)
		output, err = pb.PendingCallContract(ctx, msg)
		if err == nil && len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = pb.PendingCodeAt(ctx, contractAddressObj); err != nil {
				return go_error.WithStack(err)
			} else if len(code) == 0 {
				return go_error.WithStack(bind.ErrNoCode)
			}
		}
	} else {
		output, err = bind.ContractCaller(w.RemoteRpcClient).CallContract(ctx, msg, opts.BlockNumber)
		if err != nil {
			return go_error.WithStack(err)
		}
		if len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = bind.ContractCaller(w.RemoteRpcClient).CodeAt(ctx, contractAddressObj, opts.BlockNumber); err != nil {
				return go_error.WithStack(err)
			} else if len(code) == 0 {
				return go_error.WithStack(bind.ErrNoCode)
			}
		}
	}
	err = w.UnpackParams(out, outputTypes, hex.EncodeToString(output))
	if err != nil {
		return go_error.WithStack(err)
	}
	return nil
}

/*
*
是个同步方法

只能获取以后的而且区块确认了的事件，即使 start 指定为过去的 block number，也不能获取到

query 的第一个 []interface{} 是指第一个 index ，第二个是指第二个 index
*/
func (w *Wallet) WatchLogsByWs(resultChan chan map[string]interface{}, contractAddress, abiStr, eventName string, opts *bind.WatchOpts, query ...[]interface{}) error {
	if w.RemoteWsClient == nil || w.WsClient == nil {
		return errors.New("please set ws url")
	}
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return go_error.WithStack(err)
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteWsClient, w.RemoteWsClient, w.RemoteWsClient)
retry:
	for {
		chanLog, sub, err := contractInstance.WatchLogs(opts, eventName, query...)
		if err != nil {
			w.logger.WarnF("connect failed, reconnect after 3s. err -> %#v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		w.logger.Info("connected. watching...")
		for {
			select {
			case log1 := <-chanLog:
				map_ := make(map[string]interface{})
				err := contractInstance.UnpackLogIntoMap(map_, eventName, log1)
				if err != nil {
					sub.Unsubscribe()
					return go_error.WithStack(err)
				}
				resultChan <- map_
			case err := <-sub.Err():
				w.logger.WarnF("connection closed. err -> %#v", err)
				if err == nil { // 自己主动关闭的
					return nil
				}
				sub.Unsubscribe()
				time.Sleep(3 * time.Second)
				w.logger.Info("reconnect...")
				continue retry
			}
		}
	}
}

/*
查找历史的已经确认的事件，但不能实时接受后面的事件。取不到 pending 中的 logs

fromBlock；nil 就是 最新块号 - 4900，负数就是 pending ，正数就是 blockNumber
toBlock；nil 就是 latest ，负数就是 pending ，正数就是 blockNumber
*/
func (w *Wallet) FindLogs(contractAddress, abiStr, eventName string, fromBlock, toBlock *big.Int, query ...[]interface{}) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}

	query = append([][]interface{}{{parsedAbi.Events[eventName].ID}}, query...)

	topics, err := abi.MakeTopics(query...)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteRpcClient, w.RemoteRpcClient, w.RemoteRpcClient)

	if fromBlock == nil {
		number, err := w.LatestBlockNumber()
		if err != nil {
			return nil, go_error.WithStack(err)
		}
		fromBlock = number.Sub(number, new(big.Int).SetUint64(4900))
	}

	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	logs, err := w.RemoteRpcClient.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{
			common.HexToAddress(contractAddress),
		},
		Topics: topics,
	})
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	for _, log := range logs {
		map_ := make(map[string]interface{})
		err := contractInstance.UnpackLogIntoMap(map_, eventName, log)
		if err != nil {
			return nil, go_error.WithStack(err)
		}
		result = append(result, map_)
	}
	return result, nil
}

type FindLogsByScanApiResult struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"` // 十六进制字符串
	Timestamp        string   `json:"timeStamp"`   // 十六进制字符串
	GasPrice         string   `json:"gasPrice"`    // 十六进制字符串
	GasUsed          string   `json:"gasUsed"`     // 十六进制字符串
	LogIndex         string   `json:"logIndex"`    // 十六进制字符串
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"` // 十六进制字符串
}

// 通过 scan api 查询 logs（只支持以太坊）。最多只会返回开始的 1000 个结果，部分结果可能会被抛弃，所以要缩小范围查询
// apikey：可以为空，但频率受限，每 5s 才能执行一次
// fromBlock: 如果是负数，则是最新高度加上这个负数
// toBlock：可以设置为 latest ，表示最新块
func (w *Wallet) FindLogsByScanApi(apikey string, contractAddress string, fromBlock string, toBlock string, timeout time.Duration, topic0 string, query ...string) ([]FindLogsByScanApiResult, error) {
	if go_decimal.Decimal.Start(fromBlock).Lt(0) {
		result, err := w.LatestBlockNumber()
		if err != nil {
			return nil, go_error.WithStack(err)
		}
		delta, ok := new(big.Int).SetString(fromBlock[1:], 10)
		if !ok {
			return nil, errors.New("string to bigint error")
		}
		fromBlock = result.Sub(result, delta).String()
	}

	params := map[string]interface{}{
		"module":    "logs",
		"action":    "getLogs",
		"fromBlock": fromBlock,
		"toBlock":   toBlock,
		"address":   contractAddress,
		"topic0":    topic0,
		"apikey":    apikey,
	}
	for i, str := range query {
		oprStr := fmt.Sprintf("topic%d_%d_opr", i, i+1)
		params[oprStr] = "and"
		params[fmt.Sprintf("topic%d", i+1)] = str
	}

	_, resStr, err := go_http.NewHttpRequester(go_http.WithLogger(w.logger), go_http.WithTimeout(timeout)).Get(go_http.RequestParam{
		Url:    ScanApiUrl,
		Params: params,
	})
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	var tempResult struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	err = json.Unmarshal([]byte(resStr), &tempResult)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	if tempResult.Status != "1" && tempResult.Message != "No records found" {
		var result struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Result  string `json:"result"`
		}
		err = json.Unmarshal([]byte(resStr), &result)
		if err != nil {
			return nil, go_error.WithStack(err)
		}
		return nil, go_error.WithStack(errors.New(result.Message + ". " + result.Result))
	}
	var result struct {
		Status  string                    `json:"status"`
		Message string                    `json:"message"`
		Result  []FindLogsByScanApiResult `json:"result"`
	}
	err = json.Unmarshal([]byte(resStr), &result)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return result.Result, nil
}

type CallMethodOpts struct {
	Nonce                uint64
	Value                *big.Int
	GasPrice             *big.Int // MaxFeePerGas
	GasLimit             uint64
	IsPredictError       bool
	MaxPriorityFeePerGas *big.Int
}

type BuildTxResult struct {
	SignedTx *types.Transaction
	TxHex    string
}

// payload 除了 methodId 就是 params
func (w *Wallet) UnpackParams(out interface{}, inputs abi.Arguments, paramsStr string) error {
	if strings.HasPrefix(paramsStr, "0x") {
		paramsStr = paramsStr[2:]
	}

	data, err := hex.DecodeString(paramsStr)
	if err != nil {
		return go_error.WithStack(err)
	}
	a, err := inputs.Unpack(data)
	if err != nil {
		return go_error.WithStack(err)
	}
	err = inputs.Copy(out, a)
	if err != nil {
		return go_error.WithStack(err)
	}
	return nil
}

// 不带 0x 前缀
func (w *Wallet) PackParams(inputs abi.Arguments, args ...interface{}) (string, error) {
	bytes_, err := inputs.Pack(args...)
	if err != nil {
		return "", go_error.WithStack(err)
	}
	return hex.EncodeToString(bytes_), nil
}

// 不带 0x 前缀
func (w *Wallet) EncodePayload(abiStr string, methodName string, params ...interface{}) (string, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", go_error.WithStack(err)
	}
	input, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return "", go_error.WithStack(err)
	}
	return hex.EncodeToString(input), nil
}

func (w *Wallet) DecodePayload(abiStr string, out interface{}, payloadStr string) (*abi.Method, error) {
	if len(payloadStr) < 8 {
		return nil, errors.New("payloadStr error")
	}
	if strings.HasPrefix(payloadStr, "0x") {
		payloadStr = payloadStr[2:]
	}
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	data, err := hex.DecodeString(payloadStr)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	method, err := parsedAbi.MethodById(data[:4])
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	if len(data[4:]) > 0 {
		a, err := method.Inputs.Unpack(data[4:])
		if err != nil {
			return nil, go_error.WithStack(err)
		}
		err = method.Inputs.Copy(out, a)
		if err != nil {
			return nil, go_error.WithStack(err)
		}
	}
	return method, err
}

// 得到事件签名的 hash，也就是 topic0
func (w *Wallet) Topic0FromEventName(abiStr, eventName string) (string, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", go_error.WithStack(err)
	}

	return parsedAbi.Events[eventName].ID.String(), nil
}

/*
*
返回结果不带 0x 前缀
*/
func (w *Wallet) MethodIdFromMethodStr(methodStr string) string {
	return hex.EncodeToString(crypto.Keccak256([]byte(methodStr))[:4])
}

func (w *Wallet) MethodFromPayload(abiStr string, payloadStr string) (*abi.Method, error) {
	if len(payloadStr) < 8 {
		return nil, errors.New("payloadStr error")
	}
	if strings.HasPrefix(payloadStr, "0x") {
		payloadStr = payloadStr[2:]
	}
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	data, err := hex.DecodeString(payloadStr)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	method, err := parsedAbi.MethodById(data[:4])
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return method, err
}

func (w *Wallet) SuggestGasPrice() (*big.Int, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	gasPrice, err := w.RemoteRpcClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, go_error.WithStack(fmt.Errorf("failed to suggest gas price: %v", err))
	}
	return gasPrice, nil
}

func (w *Wallet) LatestBlockNumber() (*big.Int, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	number, err := w.RemoteRpcClient.BlockNumber(ctx)
	if err != nil {
		return nil, go_error.WithStack(fmt.Errorf("failed to get latest block number: %v", err))
	}
	return new(big.Int).SetUint64(number), nil
}

func (w *Wallet) EstimateGas(msg ethereum.CallMsg) (uint64, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	gasCount, err := w.RemoteRpcClient.EstimateGas(ctx, msg)
	if err != nil {
		return 0, go_error.WithStack(fmt.Errorf("failed to estimate gas needed: %v", err))
	}
	return gasCount, nil
}

func (w *Wallet) PrivateKeyToAddress(privateKey string) (string, error) {
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", go_error.WithStack(err)
	}
	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return "", go_error.WithStack(err)
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	return crypto.PubkeyToAddress(publicKeyECDSA).String(), nil
}

func (w *Wallet) BuildCallMethodTx(privateKey, contractAddress, abiStr, methodName string, opts *CallMethodOpts, params ...interface{}) (*BuildTxResult, error) {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	contractAddressObj := common.HexToAddress(contractAddress)
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, go_error.WithStack(err)
	}

	var value = big.NewInt(0)
	var gasLimit uint64 = 0
	var nonce uint64 = 0
	var isPredictError = true
	if opts != nil {
		if opts.Value != nil {
			value = opts.Value
		}

		gasLimit = opts.GasLimit
		nonce = opts.Nonce
		isPredictError = opts.IsPredictError
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKeyECDSA)
	if nonce == 0 {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		nonce, err = w.RemoteRpcClient.PendingNonceAt(ctx, fromAddress)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to retrieve account nonce: %v", err))
		}
	}
	input, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	if gasLimit == 0 || isPredictError {
		msg := ethereum.CallMsg{
			From:  fromAddress,
			To:    &contractAddressObj,
			Value: value,
			Data:  input,
		}
		tempGasLimit, err := w.EstimateGas(msg)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to estimate gas: %v", err))
		}
		if gasLimit == 0 {
			gasLimit = uint64(float64(tempGasLimit) * 1.3)
		}
	}

	return w.buildTx(privateKeyECDSA, nonce, contractAddressObj, value, gasLimit, input, opts)
}

func (w *Wallet) buildTx(privateKeyECDSA *ecdsa.PrivateKey, nonce uint64, toAddressObj common.Address, value *big.Int, gasLimit uint64, data []byte, opts *CallMethodOpts) (*BuildTxResult, error) {
	var rawTx *types.Transaction
	if opts == nil || opts.MaxPriorityFeePerGas == nil {
		var gasPrice *big.Int = nil
		if opts != nil && opts.GasPrice != nil {
			gasPrice = opts.GasPrice
		}
		if gasPrice == nil {
			ctx, _ := context.WithTimeout(context.Background(), w.timeout)
			_gasPrice, err := w.RemoteRpcClient.SuggestGasPrice(ctx)
			if err != nil {
				return nil, go_error.WithStack(fmt.Errorf("failed to suggest gas price: %v", err))
			}
			gasPrice = _gasPrice
		}
		rawTx = types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       &toAddressObj,
			Value:    value,
			Gas:      gasLimit,
			GasPrice: gasPrice,
			Data:     data,
		})
	} else {
		rawTx = types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonce,
			To:        &toAddressObj,
			Value:     value,
			Gas:       gasLimit,
			GasFeeCap: opts.GasPrice,             // maxFeePerGas 最大的 gasPrice（包含 baseFee），减去 baseFee 就是小费。gasPrice = min(maxFeePerGas, baseFee + maxPriorityFeePerGas)
			GasTipCap: opts.MaxPriorityFeePerGas, // maxPriorityFeePerGas，也就是最大的小费。GasTipCap 和 gasFeeCap - baseFee 的更小值才是真正的给矿工的，baseFee 是销毁的。
			Data:      data,
		})
	}
	signedTx, err := types.SignTx(rawTx, types.LatestSignerForChainID(w.chainId), privateKeyECDSA)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return &BuildTxResult{
		SignedTx: signedTx,
		TxHex:    hexutil.Encode(txBytes),
	}, nil
}

func (w *Wallet) BuildCallMethodTxWithPayload(privateKey, contractAddress, payload string, opts *CallMethodOpts) (*BuildTxResult, error) {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	if strings.HasPrefix(payload, "0x") {
		payload = payload[2:]
	}
	payloadBuf, err := hex.DecodeString(payload)
	if err != nil {
		return nil, go_error.WithStack(err)
	}

	contractAddressObj := common.HexToAddress(contractAddress)

	var value = big.NewInt(0)
	var gasLimit uint64 = 0
	var nonce uint64 = 0
	var isPredictError = true
	if opts != nil {
		if opts.Value != nil {
			value = opts.Value
		}

		gasLimit = opts.GasLimit
		nonce = opts.Nonce
		isPredictError = opts.IsPredictError
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKeyECDSA)
	if nonce == 0 {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		nonce, err = w.RemoteRpcClient.PendingNonceAt(ctx, fromAddress)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to retrieve account nonce: %v", err))
		}
	}

	if gasLimit == 0 || isPredictError {
		msg := ethereum.CallMsg{
			From:  fromAddress,
			To:    &contractAddressObj,
			Value: value,
			Data:  payloadBuf,
		}
		tempGasLimit, err := w.EstimateGas(msg)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to estimate gas: %v", err))
		}
		if gasLimit == 0 {
			gasLimit = uint64(float64(tempGasLimit) * 1.3)
		}
	}

	return w.buildTx(privateKeyECDSA, nonce, contractAddressObj, value, gasLimit, payloadBuf, opts)
}

func (w *Wallet) BuildTransferTx(privateKey, toAddress string, opts *CallMethodOpts) (*BuildTxResult, error) {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

	toAddressObj := common.HexToAddress(toAddress)
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, go_error.WithStack(err)
	}

	var value = big.NewInt(0)

	var gasLimit uint64 = 0
	var nonce uint64 = 0
	if opts != nil {
		if opts.Value != nil {
			value = opts.Value
		}

		gasLimit = opts.GasLimit
		nonce = opts.Nonce
	}
	if gasLimit == 0 {
		gasLimit = 21000
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBuf)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKeyECDSA)
	if nonce == 0 {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		nonce, err = w.RemoteRpcClient.PendingNonceAt(ctx, fromAddress)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to retrieve account nonce: %v", err))
		}
	}

	return w.buildTx(privateKeyECDSA, nonce, toAddressObj, value, gasLimit, nil, opts)
}

func (w *Wallet) SendRawTransaction(txHex string) (string, error) {
	var hash common.Hash
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	err := w.RpcClient.CallContext(ctx, &hash, "eth_sendRawTransaction", txHex)
	if err != nil {
		return "", go_error.WithStack(err)
	}
	return hash.String(), nil
}

type TransactionByHashResult struct {
	*types.Transaction
	From common.Address
}

func (w *Wallet) TransactionByHash(txHash string) (*TransactionByHashResult, bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	tx, isPending, err := w.RemoteRpcClient.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, false, go_error.WithStack(err)
	}
	from, err := types.Sender(types.NewEIP155Signer(w.chainId), tx)
	if err != nil {
		return nil, false, go_error.WithStack(err)
	}
	return &TransactionByHashResult{
		tx,
		from,
	}, isPending, nil
}

func (w *Wallet) TransactionReceiptByHash(txHash string) (*types.Receipt, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	receipt, err := w.RemoteRpcClient.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return receipt, nil
}

func (w *Wallet) WaitConfirm(txHash string, interval time.Duration) *types.Receipt {
	timer := time.NewTimer(0)
	for range timer.C {
		receipt, err := w.TransactionReceiptByHash(txHash)
		if err != nil {
			w.logger.WarnF("TransactionReceiptByHash: %#v", err)
			timer.Reset(interval)
			continue
		}
		if receipt.BlockNumber == nil {
			timer.Reset(interval)
			continue
		}
		timer.Stop()
		return receipt
	}
	return nil
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

func (w *Wallet) TxsInPool() (*TxsInPoolResult, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	var result TxsInPoolResult
	err := w.RpcClient.CallContext(ctx, &result, "txpool_content")
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return &result, nil
}

func (w *Wallet) Balance(address string) (*big.Int, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	result, err := w.RemoteRpcClient.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return result, nil
}

func (w *Wallet) ApproveAmount(contractAddress, fromAddress, toAddress string) (*big.Int, error) {
	result := new(big.Int)
	err := w.CallContractConstant(
		&result,
		contractAddress,
		Erc20AbiStr,
		"allowance",
		nil,
		common.HexToAddress(fromAddress),
		common.HexToAddress(toAddress),
	)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return result, nil
}

func (w *Wallet) Approve(priv, contractAddress, toAddress string, amount *big.Int, opts *CallMethodOpts) (string, error) {
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
		common.HexToAddress(toAddress),
		approveAmount,
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

type DeriveFromPathResult struct {
	Address    string
	PublicKey  string
	PrivateKey string
}

// 都不带 0x 前缀
func (w *Wallet) DeriveFromPath(seed string, path string) (*DeriveFromPathResult, error) {
	// 字符串转换成字节数组
	seedBuf, err := hex.DecodeString(seed)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	// 通过种子生成 hdwallet
	wallet, err := chain.NewFromSeed(seedBuf)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	// 解析派生路径
	hdPath, err := chain.ParseDerivationPath(path)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	// 派生账号
	account, err := wallet.Derive(hdPath, true)
	// 获取私钥 hex 字符串
	privateKeyStr, err := wallet.PrivateKeyHex(account)
	if err != nil {
		return nil, go_error.WithStack(err)
	}

	privateKey, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	privateKeyECDSA, err := crypto.ToECDSA(privateKey[:])
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	publicKeyECDSA := privateKeyECDSA.PublicKey
	publicKeyStr := hex.EncodeToString(crypto.CompressPubkey(&publicKeyECDSA))
	addr := crypto.PubkeyToAddress(publicKeyECDSA).String()
	return &DeriveFromPathResult{
		Address:    addr,
		PublicKey:  publicKeyStr,
		PrivateKey: privateKeyStr,
	}, nil
}

func (w *Wallet) SignMsg(privateKey string, data string) (string, error) {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

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

func (w *Wallet) RecoverSignerAddress(msg, sig string) (*common.Address, error) {
	hash, err := w.SignHashForMsg(msg)
	if err != nil {
		return nil, err
	}
	return w.RecoverSignerAddressFromMsgHash(hash, sig)
}

func (w *Wallet) RecoverSignerAddressFromMsgHash(msgHash, sig string) (*common.Address, error) {
	if strings.HasPrefix(sig, "0x") {
		sig = sig[2:]
	}
	sigHex, err := hex.DecodeString(sig)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(msgHash, "0x") {
		msgHash = msgHash[2:]
	}
	msgHashHex, err := hex.DecodeString(msgHash)
	if err != nil {
		return nil, err
	}

	if len(sigHex) != 65 {
		return nil, fmt.Errorf("signature must be 65 bytes long")
	}
	if sigHex[64] != 27 && sigHex[64] != 28 {
		return nil, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
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
func (w *Wallet) SignHashForMsg(data string) (string, error) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return hex.EncodeToString(crypto.Keccak256([]byte(msg))), nil
}

func (w *Wallet) SendEth(priv string, address string, opts *CallMethodOpts) (string, error) {
	tx, err := w.BuildTransferTx(priv, address, opts)
	if err != nil {
		return "", err
	}
	txHash, err := w.SendRawTransaction(tx.TxHex)
	if err != nil {
		return "", err
	}
	return txHash, nil
}

func (w *Wallet) SendToken(priv string, contractAddress, address string, amount *big.Int, opts *CallMethodOpts) (string, error) {
	tx, err := w.BuildCallMethodTx(
		priv,
		contractAddress,
		Erc20AbiStr,
		"transfer",
		opts,
		common.HexToAddress(address),
		amount,
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

func (w *Wallet) TokenBalance(contractAddress, address string) (*big.Int, error) {
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
		common.HexToAddress(address),
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
			return go_error.WithStack(err)
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

func (w *Wallet) SeedHexByMnemonic(mnemonic string, pass string) string {
	return hex.EncodeToString(bip39.NewSeed(mnemonic, pass))
}

func (w *Wallet) RandomMnemonic() (string, error) {
	entropy, err := go_random.RandomInstance.RandomBytes(16)
	if err != nil {
		return "", go_error.WithStack(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", go_error.WithStack(err)
	}
	return mnemonic, nil
}

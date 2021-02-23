package go_coin_eth

import (
	"context"
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
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	chain "github.com/pefish/go-coin-eth/util"
	"github.com/pefish/go-error"
	"github.com/pefish/go-logger"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"time"
	"github.com/pefish/go-http"
	"github.com/pefish/go-decimal"
)

const (
	ScanApiUrl = "https://api.etherscan.io/api"
)

var (
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
		logger:          go_logger.DefaultLogger,
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

func (w *Wallet) CallContractConstant(contractAddress, abiStr, methodName string, opts *bind.CallOpts, params ...interface{}) ([]interface{}, error) {
	out := make([]interface{}, 0)
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteRpcClient, w.RemoteRpcClient, w.RemoteRpcClient)
	err = contractInstance.Call(opts, &out, methodName, params...)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return out, nil
}

/**
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
			return go_error.WithStack(err)
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
	Address string `json:"address"`
	Topics []string `json:"topics"`
	Data string `json:"data"`
	BlockNumber string `json:"blockNumber"`  // 十六进制字符串
	Timestamp string `json:"timeStamp"` // 十六进制字符串
	GasPrice string `json:"gasPrice"`  // 十六进制字符串
	GasUsed string `json:"gasUsed"` // 十六进制字符串
	LogIndex string `json:"logIndex"`  // 十六进制字符串
	TransactionHash string `json:"transactionHash"`
	TransactionIndex string `json:"transactionIndex"` // 十六进制字符串
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
		"module": "logs",
		"action": "getLogs",
		"fromBlock": fromBlock,
		"toBlock": toBlock,
		"address": contractAddress,
		"topic0": topic0,
		"apikey": apikey,
	}
	for i, str := range query {
		oprStr := fmt.Sprintf("topic%d_%d_opr", i, i+1)
		params[oprStr] = "and"
		params[fmt.Sprintf("topic%d", i + 1)] = str
	}

	_, resStr, err := go_http.NewHttpRequester(go_http.WithLogger(w.logger), go_http.WithTimeout(timeout)).Get(go_http.RequestParam{
		Url:       ScanApiUrl,
		Params:    params,
	})
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	var tempResult struct{
		Status string `json:"status"`
		Message string `json:"message"`
	}
	err = json.Unmarshal([]byte(resStr), &tempResult)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	if tempResult.Status != "1" && tempResult.Message != "No records found" {
		var result struct{
			Status string `json:"status"`
			Message string `json:"message"`
			Result string `json:"result"`
		}
		err = json.Unmarshal([]byte(resStr), &result)
		if err != nil {
			return nil, go_error.WithStack(err)
		}
		return nil, go_error.WithStack(errors.New(result.Message + ". " + result.Result))
	}
	var result struct{
		Status string `json:"status"`
		Message string `json:"message"`
		Result []FindLogsByScanApiResult `json:"result"`
	}
	err = json.Unmarshal([]byte(resStr), &result)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return result.Result, nil
}

type CallMethodOpts struct {
	Nonce    uint64
	Value    string
	GasPrice string
	GasLimit uint64
}

type BuildCallMethodTxResult struct {
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

func (w *Wallet) BuildCallMethodTx(privateKey, contractAddress, abiStr, methodName string, opts *CallMethodOpts, params ...interface{}) (*BuildCallMethodTxResult, error) {
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
	var gasPrice *big.Int = nil
	var gasLimit uint64 = 0
	var nonce uint64 = 0
	if opts != nil {
		if opts.Value != "" {
			tempValue, ok := new(big.Int).SetString(opts.Value, 10)
			if !ok {
				return nil, errors.New("string convert to bigint error")
			}
			value = tempValue
		}

		if opts.GasPrice != "" {
			tempGasPrice, ok := new(big.Int).SetString(opts.GasPrice, 10)
			if !ok {
				return nil, errors.New("string convert to bigint error")
			}
			gasPrice = tempGasPrice
		}

		gasLimit = opts.GasLimit
		nonce = opts.Nonce
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
	if gasPrice == nil {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		gasPrice, err = w.RemoteRpcClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to suggest gas price: %v", err))
		}
	}
	input, err := parsedAbi.Pack(methodName, params...)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	if gasLimit == 0 {
		msg := ethereum.CallMsg{From: fromAddress, To: &contractAddressObj, GasPrice: gasPrice, Value: value, Data: input}
		tempGasLimit, err := w.EstimateGas(msg)
		if err != nil {
			return nil, err
		}
		gasLimit = uint64(float64(tempGasLimit) * 1.3)
	}
	var rawTx = types.NewTransaction(nonce, contractAddressObj, value, gasLimit, gasPrice, input)
	signedTx, err := types.SignTx(rawTx, types.NewEIP155Signer(w.chainId), privateKeyECDSA)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	data, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return &BuildCallMethodTxResult{
		SignedTx: signedTx,
		TxHex:    hexutil.Encode(data),
	}, nil
}

func (w *Wallet) BuildCallMethodTxWithPayload(privateKey, contractAddress, payload string, opts *CallMethodOpts) (*BuildCallMethodTxResult, error) {
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
	var gasPrice *big.Int = nil
	var gasLimit uint64 = 0
	var nonce uint64 = 0
	if opts != nil {
		if opts.Value != "" {
			tempValue, ok := new(big.Int).SetString(opts.Value, 10)
			if !ok {
				return nil, errors.New("string convert to bigint error")
			}
			value = tempValue
		}

		if opts.GasPrice != "" {
			tempGasPrice, ok := new(big.Int).SetString(opts.GasPrice, 10)
			if !ok {
				return nil, errors.New("string convert to bigint error")
			}
			gasPrice = tempGasPrice
		}

		gasLimit = opts.GasLimit
		nonce = opts.Nonce
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
	if gasPrice == nil {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		gasPrice, err = w.RemoteRpcClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to suggest gas price: %v", err))
		}
	}
	if gasLimit == 0 {
		msg := ethereum.CallMsg{From: fromAddress, To: &contractAddressObj, GasPrice: gasPrice, Value: value, Data: payloadBuf}
		tempGasLimit, err := w.EstimateGas(msg)
		if err != nil {
			return nil, err
		}
		gasLimit = uint64(float64(tempGasLimit) * 1.3)
	}
	var rawTx = types.NewTransaction(nonce, contractAddressObj, value, gasLimit, gasPrice, payloadBuf)
	signedTx, err := types.SignTx(rawTx, types.NewEIP155Signer(w.chainId), privateKeyECDSA)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	data, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return &BuildCallMethodTxResult{
		SignedTx: signedTx,
		TxHex:    hexutil.Encode(data),
	}, nil
}

func (w *Wallet) BuildTransferTx(privateKey, toAddress string, opts *CallMethodOpts) (*BuildCallMethodTxResult, error) {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

	toAddressObj := common.HexToAddress(toAddress)
	privateKeyBuf, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, go_error.WithStack(err)
	}

	var value = big.NewInt(0)
	var gasPrice *big.Int = nil
	var gasLimit uint64 = 0
	var nonce uint64 = 0
	if opts != nil {
		if opts.Value != "" {
			tempValue, ok := new(big.Int).SetString(opts.Value, 10)
			if !ok {
				return nil, errors.New("string convert to bigint error")
			}
			value = tempValue
		}

		if opts.GasPrice != "" {
			tempGasPrice, ok := new(big.Int).SetString(opts.GasPrice, 10)
			if !ok {
				return nil, errors.New("string convert to bigint error")
			}
			gasPrice = tempGasPrice
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
	if gasPrice == nil {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		gasPrice, err = w.RemoteRpcClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to suggest gas price: %v", err))
		}
	}
	var rawTx = types.NewTransaction(nonce, toAddressObj, value, gasLimit, gasPrice, make([]byte, 0))
	signedTx, err := types.SignTx(rawTx, types.NewEIP155Signer(w.chainId), privateKeyECDSA)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	data, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return &BuildCallMethodTxResult{
		SignedTx: signedTx,
		TxHex:    hexutil.Encode(data),
	}, nil
}

func (w *Wallet) SendRawTransaction(txHex string) error {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	err := w.RpcClient.CallContext(ctx, nil, "eth_sendRawTransaction", txHex)
	if err != nil {
		return go_error.WithStack(err)
	}
	return nil
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
	msg, err := tx.AsMessage(types.NewEIP155Signer(w.chainId))
	if err != nil {
		return nil, false, go_error.WithStack(err)
	}
	return &TransactionByHashResult{
		tx,
		msg.From(),
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
		_, isPending, err := w.TransactionByHash(txHash)
		if err != nil {
			w.logger.Warn(err)
			timer.Reset(interval)
			continue
		}
		if isPending {
			timer.Reset(interval)
			continue
		}
		receipt, err := w.TransactionReceiptByHash(txHash)
		if err != nil {
			w.logger.Warn(err)
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

type DeriveFromPathResult struct {
	Address string
	PublicKey string
	PrivateKey string
}

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

func (w *Wallet) TokenBalance(contractAddress, address string) (*big.Int, error) {
	result, err := w.CallContractConstant(
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
	return result[0].(*big.Int), nil
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

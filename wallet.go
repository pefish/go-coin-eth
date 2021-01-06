package go_coin_eth

import (
	"context"
	"encoding/hex"
	"github.com/pkg/errors"
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
	"github.com/pefish/go-error"
	"github.com/pefish/go-logger"
	go_reflect "github.com/pefish/go-reflect"
	"math/big"
	"strings"
	"time"
)

type Wallet struct {
	RemoteRpcClient *ethclient.Client
	RemoteWsClient *ethclient.Client
	timeout         time.Duration
	chainId         *big.Int
	RpcClient       *rpc.Client
	WsClient       *rpc.Client
	logger          go_logger.InterfaceLogger
}

type UrlParam struct {
	RpcUrl string
	WsUrl string
}

func NewWallet(urlParam UrlParam) (*Wallet, error) {
	if urlParam.RpcUrl == "" {
		return nil, errors.New("rpc url must be set")
	}
	timeout := 60 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	rpcClient, err := rpc.DialContext(ctx, urlParam.RpcUrl)
	if err != nil {
		return nil, err
	}
	remoteRpcClient := ethclient.NewClient(rpcClient)

	ctx, _ = context.WithTimeout(context.Background(), timeout)
	chainId, err := remoteRpcClient.ChainID(ctx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	wallet := &Wallet{
		RemoteRpcClient: remoteRpcClient,
		timeout:         timeout,
		chainId:         chainId,
		RpcClient:       rpcClient,
		logger:          go_logger.DefaultLogger,
	}
	if urlParam.WsUrl != "" {
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		wsClient, err := rpc.DialContext(ctx, urlParam.WsUrl)
		if err != nil {
			return nil, err
		}
		remoteWsClient := ethclient.NewClient(wsClient)
		wallet.RemoteWsClient = remoteWsClient
		wallet.WsClient = wsClient
	}
	return wallet, nil
}

func (w *Wallet) Close() {
	w.RemoteRpcClient.Close()
	w.RpcClient.Close()
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

只能获取以后的而且区块确认了的事件，即使start指定为过去的block number，也不能获取到

query的第一个[]interface{}是指第一个index，第二个是指第二个index
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
				sub.Unsubscribe()
				w.logger.Info("reconnect...")
				continue retry
			}
		}
	}
}

/*
查找历史的已经确认的事件，但不能实时接受后面的事件。取不到pending中的logs

fromBlock；nil就是0，负数就是pending，正数就是blockNumber
toBlock；nil就是latest，负数就是pending，正数就是blockNumber
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
			value = big.NewInt(go_reflect.Reflect.MustToInt64(opts.Value))
		}

		if opts.GasPrice != "" {
			gasPrice = big.NewInt(go_reflect.Reflect.MustToInt64(opts.GasPrice))
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
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		tempGasLimit, err := w.RemoteRpcClient.EstimateGas(ctx, msg)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to estimate gas needed: %v", err))
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

func (w *Wallet) SendRawTransaction(txHex string) error {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	err := w.RpcClient.CallContext(ctx, nil, "eth_sendRawTransaction", txHex)
	if err != nil {
		return go_error.WithStack(err)
	}
	return nil
}

func (w *Wallet) TransactionByHash(txHash string) (*types.Transaction, bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	tx, isPending, err := w.RemoteRpcClient.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, false, go_error.WithStack(err)
	}
	return tx, isPending, nil
}

func (w *Wallet) WatchPendingTxByWs(resultChan chan<- string) error {
	if w.RemoteWsClient == nil || w.WsClient == nil {
		return errors.New("please set ws url")
	}
	for {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		subscription, err := w.WsClient.EthSubscribe(ctx, resultChan, "newPendingTransactions")
		if err != nil {
			subscription.Unsubscribe()
			return go_error.WithStack(err)
		}
		w.logger.Info("connected. watching...")
		err = <-subscription.Err()
		w.logger.WarnF("connection closed. err -> %#v", err)
		subscription.Unsubscribe()
		w.logger.Info("reconnect...")
	}
}

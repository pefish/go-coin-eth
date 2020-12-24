package go_coin_eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pefish/go-error"
	go_reflect "github.com/pefish/go-reflect"
	"math/big"
	"strings"
	"time"
)

type Wallet struct {
	RemoteClient *ethclient.Client
	timeout          time.Duration
	chainId      *big.Int
	nodeUrl string
}

func NewWallet(url string) (*Wallet, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	timeout := 30 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return &Wallet{
		RemoteClient: client,
		timeout:      timeout,
		chainId:      chainId,
		nodeUrl: url,
	}, nil
}

func (w *Wallet) CallContractConstant(contractAddress, abiStr, methodName string, opts *bind.CallOpts, params ...interface{}) ([]interface{}, error) {
	out := make([]interface{}, 0)
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteClient, w.RemoteClient, w.RemoteClient)
	err = contractInstance.Call(opts, &out, methodName, params...)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return out, nil
}

/**
只能获取以后的事件，即使start指定为过去的block number，也不能获取到
query的第一个[]interface{}是指第一个index，第二个是指第二个index
*/
func (w *Wallet) WatchLogs(resultChan chan map[string]interface{}, errChan chan error, contractAddress, abiStr, eventName string, opts *bind.WatchOpts, query ...[]interface{}) (event.Subscription, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteClient, w.RemoteClient, w.RemoteClient)
	chanLog, sub, err := contractInstance.WatchLogs(opts, eventName, query...)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	go func() {
		for {
			select {
			case log1 := <-chanLog:
				map_ := make(map[string]interface{})
				err := contractInstance.UnpackLogIntoMap(map_, eventName, log1)
				if err != nil {
					errChan <- go_error.WithStack(err)
					return
				}
				resultChan <- map_
			}
		}
	}()
	return sub, nil
}

/*
查找历史的事件，但不能实时接受后面的事件
*/
func (w *Wallet) FindLogs(resultChan chan map[string]interface{}, errChan chan error, contractAddress, abiStr, eventName string, opts *bind.FilterOpts, query ...[]interface{}) (event.Subscription, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteClient, w.RemoteClient, w.RemoteClient)
	chanLog, sub, err := contractInstance.FilterLogs(opts, eventName, query...)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	go func() {
		for {
			select {
			case log1 := <-chanLog:
				map_ := make(map[string]interface{})
				err := contractInstance.UnpackLogIntoMap(map_, eventName, log1)
				if err != nil {
					errChan <- go_error.WithStack(err)
					return
				}
				resultChan <- map_
			}
		}
	}()
	return sub, nil
}

type CallMethodOpts struct {
	Nonce     uint64
	Value     string
	GasPrice  string
	GasLimit  uint64
	Broadcast bool
}

func (w *Wallet) CallMethod(privateKey, contractAddress, abiStr, methodName string, opts *CallMethodOpts, params ...interface{}) (*types.Transaction, error) {
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
		nonce, err = w.RemoteClient.PendingNonceAt(ctx, fromAddress)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to retrieve account nonce: %v", err))
		}
	}
	if gasPrice == nil {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		gasPrice, err = w.RemoteClient.SuggestGasPrice(ctx)
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
		tempGasLimit, err := w.RemoteClient.EstimateGas(ctx, msg)
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
	if opts != nil && opts.Broadcast == true {
		ctx, _ := context.WithTimeout(context.Background(), w.timeout)
		err = w.RemoteClient.SendTransaction(ctx, signedTx)
		if err != nil {
			return nil, go_error.WithStack(err)
		}
	}
	return signedTx, nil
}


func (w *Wallet) SendSignedTransaction(tx *types.Transaction) error {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return go_error.WithStack(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	rpcClient, err := rpc.DialContext(ctx, w.nodeUrl)
	if err != nil {
		return go_error.WithStack(err)
	}
	ctx, _ = context.WithTimeout(context.Background(), w.timeout)
	err = rpcClient.CallContext(ctx, nil, "eth_sendRawTransaction", hexutil.Encode(data))
	if err != nil {
		return go_error.WithStack(err)
	}
	return nil
}

func (w *Wallet) SendRawTransaction(txHex string) error {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	rpcClient, err := rpc.DialContext(ctx, w.nodeUrl)
	if err != nil {
		return go_error.WithStack(err)
	}
	ctx, _ = context.WithTimeout(context.Background(), w.timeout)
	err = rpcClient.CallContext(ctx, nil, "eth_sendRawTransaction", txHex)
	if err != nil {
		return go_error.WithStack(err)
	}
	return nil
}

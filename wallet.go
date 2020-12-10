package go_coin_eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pefish/go-error"
	go_reflect "github.com/pefish/go-reflect"
	"math/big"
	"strings"
	"time"
)

type Wallet struct {
	RemoteClient *ethclient.Client
	ctx          context.Context
	chainId      *big.Int
}

func NewWallet(url string) (*Wallet, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return &Wallet{
		RemoteClient: client,
		ctx:          ctx,
		chainId:      chainId,
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
		return nil, err
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteClient, w.RemoteClient, w.RemoteClient)
	chanLog, sub, err := contractInstance.WatchLogs(opts, eventName, query...)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case log1 := <- chanLog:
				map_ := make(map[string]interface{})
				err := contractInstance.UnpackLogIntoMap(map_, eventName, log1)
				if err != nil {
					errChan <- err
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
		return nil, err
	}
	contractInstance := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedAbi, w.RemoteClient, w.RemoteClient, w.RemoteClient)
	chanLog, sub, err := contractInstance.FilterLogs(opts, eventName, query...)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case log1 := <- chanLog:
				map_ := make(map[string]interface{})
				err := contractInstance.UnpackLogIntoMap(map_, eventName, log1)
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- map_
			}
		}
	}()
	return sub, nil
}

type CallMethodOpts struct {
	Nonce    uint64
	Value    string
	GasPrice string
	GasLimit uint64
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
		nonce, err = w.RemoteClient.PendingNonceAt(w.ctx, fromAddress)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to retrieve account nonce: %v", err))
		}
	}
	if gasPrice == nil {
		gasPrice, err = w.RemoteClient.SuggestGasPrice(w.ctx)
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
		gasLimit, err = w.RemoteClient.EstimateGas(w.ctx, msg)
		if err != nil {
			return nil, go_error.WithStack(fmt.Errorf("failed to estimate gas needed: %v", err))
		}
	}
	var rawTx = types.NewTransaction(nonce, contractAddressObj, value, gasLimit, gasPrice, input)
	signedTx, err := types.SignTx(rawTx, types.NewEIP155Signer(w.chainId), privateKeyECDSA)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	err = w.RemoteClient.SendTransaction(w.ctx, signedTx)
	if err != nil {
		return nil, go_error.WithStack(err)
	}
	return signedTx, nil
}

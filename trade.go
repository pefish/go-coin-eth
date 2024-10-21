package go_coin_eth

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_decimal "github.com/pefish/go-decimal"
	"github.com/pkg/errors"
)

func (w *Wallet) WETHAddressFromRouter(routerAddress string) (string, error) {
	var wethAddress common.Address
	err := w.CallContractConstant(
		&wethAddress,
		routerAddress,
		RouterAbiStr,
		"WETH",
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}
	return wethAddress.String(), nil
}

type BuyByExactETHResult struct {
	TokenAmount string
	TxId        string
}

type BuyByExactETHOpts struct {
	TokenDecimals uint64
	WETHAddress   string
}

func (w *Wallet) BuyByExactETH(
	priv string,
	ethAmount string,
	routerAddress string,
	tokenAddress string,
	opts *BuyByExactETHOpts,
) (*BuyByExactETHResult, error) {

	var result BuyByExactETHResult

	var realOpts BuyByExactETHOpts
	if opts != nil {
		realOpts = *opts
	}

	selfAddress, err := w.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}
	ethAmountWithDecimals := go_decimal.Decimal.MustStart(ethAmount).MustShiftedBy(18).MustEndForBigInt()

	balanceWithDecimals, err := w.Balance(selfAddress)
	if err != nil {
		return nil, err
	}

	if balanceWithDecimals.Add(balanceWithDecimals, big.NewInt(10000000000000000)).Cmp(ethAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s + 10000000000000000", balanceWithDecimals, ethAmountWithDecimals)
	}

	if realOpts.WETHAddress == "" {
		wethAddress_, err := w.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	if realOpts.TokenDecimals == 0 {
		decimals_, err := w.TokenDecimals(tokenAddress)
		if err != nil {
			return nil, err
		}
		realOpts.TokenDecimals = decimals_
	}

	btr, err := w.BuildCallMethodTx(
		priv,
		routerAddress,
		RouterAbiStr,
		"swapExactETHForTokensSupportingFeeOnTransferTokens",
		&CallMethodOpts{
			Value: ethAmountWithDecimals,
		},
		[]interface{}{
			new(big.Int).SetInt64(0),
			[]common.Address{
				common.HexToAddress(realOpts.WETHAddress),
				common.HexToAddress(tokenAddress),
			},
			common.HexToAddress(selfAddress),
			go_decimal.Decimal.MustStart(time.Now().Unix()).Round(0).MustAdd(200).MustEndForBigInt(),
		},
	)
	if err != nil {
		return nil, err
	}
	result.TxId = btr.SignedTx.Hash().String()
	tr, err := w.SendRawTransactionWait(context.Background(), btr.TxHex)
	if err != nil {
		return &result, err
	}
	tokenAmountWithDecimals, err := w.receivedTokenAmountInLogs(
		tr.Logs,
		tokenAddress,
		selfAddress,
	)
	if err != nil {
		return &result, err
	}

	result.TokenAmount = go_decimal.Decimal.MustStart(tokenAmountWithDecimals).MustUnShiftedBy(realOpts.TokenDecimals).EndForString()
	return &result, nil
}

type SellByExactTokenResult struct {
	ETHAmount string
	TxId      string
}

type SellByExactTokenOpts struct {
	TokenDecimals uint64
	WETHAddress   string
}

func (w *Wallet) SellByExactToken(
	priv string,
	tokenAmount string,
	routerAddress string,
	tokenAddress string,
	opts *SellByExactTokenOpts,
) (*SellByExactTokenResult, error) {
	var result SellByExactTokenResult

	var realOpts SellByExactTokenOpts
	if opts != nil {
		realOpts = *opts
	}

	selfAddress, err := w.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}
	if realOpts.TokenDecimals == 0 {
		decimals_, err := w.TokenDecimals(tokenAddress)
		if err != nil {
			return nil, err
		}
		realOpts.TokenDecimals = decimals_
	}
	tokenAmountWithDecimals := go_decimal.Decimal.MustStart(tokenAmount).MustShiftedBy(realOpts.TokenDecimals).MustEndForBigInt()

	tokenBalanceWithDecimals, err := w.TokenBalance(tokenAddress, selfAddress)
	if err != nil {
		return nil, err
	}
	if tokenBalanceWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s", tokenBalanceWithDecimals, tokenAmountWithDecimals)
	}

	if realOpts.WETHAddress == "" {
		wethAddress_, err := w.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	approvedAmountWithDecimals, err := w.ApprovedAmount(
		tokenAddress,
		selfAddress,
		routerAddress,
	)
	if err != nil {
		return nil, err
	}

	if approvedAmountWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		_, err := w.ApproveWait(
			context.Background(),
			priv,
			tokenAddress,
			routerAddress,
			MaxUint256,
			nil,
		)
		if err != nil {
			return nil, err
		}
	}
	btr, err := w.BuildCallMethodTx(
		priv,
		routerAddress,
		RouterAbiStr,
		"swapExactTokensForETHSupportingFeeOnTransferTokens",
		nil,
		[]interface{}{
			tokenAmountWithDecimals,
			new(big.Int).SetInt64(0),
			[]common.Address{
				common.HexToAddress(tokenAddress),
				common.HexToAddress(opts.WETHAddress),
			},
			common.HexToAddress(selfAddress),
			go_decimal.Decimal.MustStart(time.Now().Unix()).Round(0).MustAdd(200).MustEndForBigInt(),
		},
	)
	if err != nil {
		return nil, err
	}
	result.TxId = btr.SignedTx.Hash().String()
	tr, err := w.SendRawTransactionWait(context.Background(), btr.TxHex)
	if err != nil {
		return &result, err
	}

	ethAmountWithDecimals, err := w.receivedETHAmountInLogs(tr.Logs, opts.WETHAddress)
	if err != nil {
		return &result, err
	}

	result.ETHAmount = go_decimal.Decimal.MustStart(ethAmountWithDecimals).MustUnShiftedBy(18).EndForString()
	return &result, nil
}

func (w *Wallet) receivedTokenAmountInLogs(
	logs []*types.Log,
	tokenAddress string,
	myAddress string,
) (*big.Int, error) {
	result := big.NewInt(0)

	transferLogs, err := w.FilterLogs(
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		tokenAddress,
		logs,
	)
	if err != nil {
		return nil, err
	}
	for _, log := range transferLogs {
		var transferEvent struct {
			From  common.Address
			Value *big.Int
			To    common.Address
		}
		err := w.UnpackLog(
			&transferEvent,
			Erc20AbiStr,
			"Transfer",
			log,
		)
		if err != nil {
			return nil, err
		}
		if transferEvent.To.Cmp(common.HexToAddress(myAddress)) != 0 {
			continue
		}
		result.Add(result, transferEvent.Value)
	}

	return result, nil
}

func (w *Wallet) receivedETHAmountInLogs(
	logs []*types.Log,
	wethAddress string,
) (*big.Int, error) {
	withdrawalLogs, err := w.FilterLogs(
		"0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65",
		wethAddress,
		logs,
	)
	if err != nil {
		return nil, err
	}
	var withdrawalEvent struct {
		Src common.Address
		Wad *big.Int
	}
	err = w.UnpackLog(&withdrawalEvent, WETHAbiStr, "Withdrawal", withdrawalLogs[len(withdrawalLogs)-1])
	if err != nil {
		return nil, err
	}

	return withdrawalEvent.Wad, nil
}

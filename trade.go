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

func (w *Wallet) GetAmountsOut(
	routerAddress string,
	amountInWithDecimals string,
	path []string,
) (amountOutWithDecimals_ string, err_ error) {
	if len(path) > 2 {
		return "", errors.New("Length of path must be 2.")
	}
	results := make([]*big.Int, 0)
	err := w.CallContractConstant(
		&results,
		routerAddress,
		RouterAbiStr,
		"getAmountsOut",
		nil,
		[]interface{}{
			go_decimal.Decimal.MustStart(amountInWithDecimals).MustEndForBigInt(),
			[]common.Address{
				common.HexToAddress(path[0]),
				common.HexToAddress(path[1]),
			},
		},
	)
	if err != nil {
		return "", err
	}
	return results[1].String(), nil
}

type BuyByExactETHResult struct {
	TokenAmount string
	Fee         string
	TxId        string
}

type BuyByExactETHOpts struct {
	TokenDecimals uint64
	WETHAddress   string
	Slippage      float64 // 滑点，默认 0.5%
	GasLimit      uint64
}

func (w *Wallet) BuyByExactETH(
	ctx context.Context,
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

	if realOpts.Slippage == 0 {
		realOpts.Slippage = 0.005
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 300000
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

	amountOutWithDecimals, err := w.GetAmountsOut(
		routerAddress,
		ethAmountWithDecimals.String(),
		[]string{opts.WETHAddress, tokenAddress},
	)
	if err != nil {
		return nil, err
	}

	minTokenAmountWithDecimals := go_decimal.Decimal.
		MustStart(amountOutWithDecimals).
		MustMulti(1 - realOpts.Slippage).
		RoundDown(0).
		MustEndForBigInt()

	btr, err := w.BuildCallMethodTx(
		priv,
		routerAddress,
		RouterAbiStr,
		"swapExactETHForTokensSupportingFeeOnTransferTokens",
		&CallMethodOpts{
			Value:    ethAmountWithDecimals,
			GasLimit: realOpts.GasLimit,
		},
		[]interface{}{
			minTokenAmountWithDecimals,
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
	w.logger.InfoF("购买 txid <%s> 等待确认", result.TxId)
	tr, err := w.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return &result, err
	}
	w.logger.InfoF("<%s> 已确认", result.TxId)
	tokenAmountWithDecimals, err := w.receivedTokenAmountInLogs(
		tr.Logs,
		tokenAddress,
		selfAddress,
	)
	if err != nil {
		return &result, err
	}

	result.TokenAmount = go_decimal.Decimal.MustStart(tokenAmountWithDecimals).MustUnShiftedBy(realOpts.TokenDecimals).EndForString()
	result.Fee = go_decimal.Decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
	return &result, nil
}

type SellByExactTokenResult struct {
	ETHAmount   string
	ApproveFee  string
	SellFee     string
	Fee         string
	TxId        string
	ApproveTxId string
}

type SellByExactTokenOpts struct {
	TokenDecimals uint64
	WETHAddress   string
	Slippage      float64 // 滑点，默认 0.5%
	GasLimit      uint64
}

func (w *Wallet) SellByExactToken(
	ctx context.Context,
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

	if realOpts.Slippage == 0 {
		realOpts.Slippage = 0.005
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 400000
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

	result.ApproveFee = "0"
	if approvedAmountWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		tr, err := w.ApproveWait(
			ctx,
			priv,
			tokenAddress,
			routerAddress,
			MaxUint256,
			nil,
		)
		if err != nil {
			return nil, err
		}
		result.ApproveTxId = tr.TxHash.String()
		result.ApproveFee = go_decimal.Decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
		w.logger.InfoF("Approve 成功。txid <%s>", tr.TxHash.String())
	}

	amountOutWithDecimals, err := w.GetAmountsOut(
		routerAddress,
		tokenAmountWithDecimals.String(),
		[]string{tokenAddress, opts.WETHAddress},
	)
	if err != nil {
		return &result, err
	}

	minETHAmountWithDecimals := go_decimal.Decimal.
		MustStart(amountOutWithDecimals).
		MustMulti(1 - realOpts.Slippage).
		RoundDown(0).
		MustEndForBigInt()

	btr, err := w.BuildCallMethodTx(
		priv,
		routerAddress,
		RouterAbiStr,
		"swapExactTokensForETHSupportingFeeOnTransferTokens",
		&CallMethodOpts{
			GasLimit: realOpts.GasLimit,
		},
		[]interface{}{
			tokenAmountWithDecimals,
			minETHAmountWithDecimals,
			[]common.Address{
				common.HexToAddress(tokenAddress),
				common.HexToAddress(opts.WETHAddress),
			},
			common.HexToAddress(selfAddress),
			go_decimal.Decimal.MustStart(time.Now().Unix()).Round(0).MustAdd(200).MustEndForBigInt(),
		},
	)
	if err != nil {
		return &result, err
	}
	result.TxId = btr.SignedTx.Hash().String()
	w.logger.InfoF("出售 txid <%s> 等待确认", result.TxId)
	tr, err := w.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return &result, err
	}
	w.logger.InfoF("<%s> 已确认", result.TxId)

	ethAmountWithDecimals, err := w.receivedETHAmountInLogs(tr.Logs, opts.WETHAddress)
	if err != nil {
		return &result, err
	}

	result.ETHAmount = go_decimal.Decimal.MustStart(ethAmountWithDecimals).MustUnShiftedBy(18).EndForString()
	result.SellFee = go_decimal.Decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
	result.Fee = go_decimal.Decimal.MustStart(result.ApproveFee).MustAddForString(result.SellFee)
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

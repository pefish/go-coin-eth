package uniswap_v2_trade

// 同样适用于 pancake V2

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	constants "github.com/pefish/go-coin-eth/uniswap-v2-trade/constant"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

type Trader struct {
	wallet *go_coin_eth.Wallet
	logger i_logger.ILogger
}

func New(
	logger i_logger.ILogger,
	wallet *go_coin_eth.Wallet,
) *Trader {
	return &Trader{
		wallet: wallet,
		logger: logger,
	}
}

func (t *Trader) WETHAddressFromRouter(routerAddress string) (string, error) {
	var wethAddress common.Address
	err := t.wallet.CallContractConstant(
		&wethAddress,
		routerAddress,
		constants.RouterAbiStr,
		"WETH",
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}
	return wethAddress.String(), nil
}

func (t *Trader) GetAmountsOut(
	routerAddress string,
	amountInWithDecimals string,
	path []string,
) (amountOutWithDecimals_ string, err_ error) {
	if len(path) > 2 {
		return "", errors.New("Length of path must be 2.")
	}
	results := make([]*big.Int, 0)
	err := t.wallet.CallContractConstant(
		&results,
		routerAddress,
		constants.RouterAbiStr,
		"getAmountsOut",
		nil,
		[]any{
			go_decimal.MustStart(amountInWithDecimals).MustEndForBigInt(),
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
	MaxFeePerGas  *big.Int
}

func (t *Trader) BuyByExactETH(
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

	selfAddress, err := t.wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}
	ethAmountWithDecimals := go_decimal.MustStart(ethAmount).MustShiftedBy(18).MustEndForBigInt()

	balanceWithDecimals, err := t.wallet.Balance(selfAddress)
	if err != nil {
		return nil, err
	}

	if balanceWithDecimals.Add(balanceWithDecimals, big.NewInt(10000000000000000)).Cmp(ethAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s + 10000000000000000", balanceWithDecimals, ethAmountWithDecimals)
	}

	if realOpts.WETHAddress == "" {
		wethAddress_, err := t.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	if realOpts.TokenDecimals == 0 {
		decimals_, err := t.wallet.TokenDecimals(tokenAddress)
		if err != nil {
			return nil, err
		}
		realOpts.TokenDecimals = decimals_
	}

	amountOutWithDecimals, err := t.GetAmountsOut(
		routerAddress,
		ethAmountWithDecimals.String(),
		[]string{realOpts.WETHAddress, tokenAddress},
	)
	if err != nil {
		return nil, err
	}

	minTokenAmountWithDecimals := go_decimal.
		MustStart(amountOutWithDecimals).
		MustMulti(1 - realOpts.Slippage).
		RoundDown(0).
		MustEndForBigInt()

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		constants.RouterAbiStr,
		"swapExactETHForTokensSupportingFeeOnTransferTokens",
		&go_coin_eth.CallMethodOpts{
			Value:        ethAmountWithDecimals,
			GasLimit:     realOpts.GasLimit,
			MaxFeePerGas: realOpts.MaxFeePerGas,
		},
		[]any{
			minTokenAmountWithDecimals,
			[]common.Address{
				common.HexToAddress(realOpts.WETHAddress),
				common.HexToAddress(tokenAddress),
			},
			common.HexToAddress(selfAddress),
			go_decimal.MustStart(time.Now().Unix()).Round(0).MustAdd(200).MustEndForBigInt(),
		},
	)
	if err != nil {
		return nil, err
	}
	result.TxId = btr.SignedTx.Hash().String()
	t.logger.InfoF("购买 txid <%s> 等待确认", result.TxId)
	tr, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return &result, err
	}
	t.logger.InfoF("<%s> 已确认", result.TxId)
	tokenAmountWithDecimals, err := t.receivedTokenAmountInLogs(
		tr.Logs,
		tokenAddress,
		selfAddress,
	)
	if err != nil {
		return &result, err
	}

	result.TokenAmount = go_decimal.MustStart(tokenAmountWithDecimals).MustUnShiftedBy(realOpts.TokenDecimals).EndForString()
	result.Fee = go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
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
	MaxFeePerGas  *big.Int
}

func (t *Trader) SellByExactToken(
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

	selfAddress, err := t.wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}
	if realOpts.TokenDecimals == 0 {
		decimals_, err := t.wallet.TokenDecimals(tokenAddress)
		if err != nil {
			return nil, err
		}
		realOpts.TokenDecimals = decimals_
	}
	tokenAmountWithDecimals := go_decimal.MustStart(tokenAmount).MustShiftedBy(realOpts.TokenDecimals).MustEndForBigInt()

	tokenBalanceWithDecimals, err := t.wallet.TokenBalance(tokenAddress, selfAddress)
	if err != nil {
		return nil, err
	}
	if tokenBalanceWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s", tokenBalanceWithDecimals, tokenAmountWithDecimals)
	}

	if realOpts.WETHAddress == "" {
		wethAddress_, err := t.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	approvedAmountWithDecimals, err := t.wallet.ApprovedAmount(
		tokenAddress,
		selfAddress,
		routerAddress,
	)
	if err != nil {
		return nil, err
	}

	result.ApproveFee = "0"
	if approvedAmountWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		tr, err := t.wallet.ApproveWait(
			ctx,
			priv,
			tokenAddress,
			routerAddress,
			go_coin_eth.MaxUint256,
			&go_coin_eth.CallMethodOpts{
				MaxFeePerGas: realOpts.MaxFeePerGas,
			},
		)
		if err != nil {
			return nil, err
		}
		result.ApproveTxId = tr.TxHash.String()
		result.ApproveFee = go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
		t.logger.InfoF("Approve 成功。txid <%s>", tr.TxHash.String())
	}

	amountOutWithDecimals, err := t.GetAmountsOut(
		routerAddress,
		tokenAmountWithDecimals.String(),
		[]string{tokenAddress, realOpts.WETHAddress},
	)
	if err != nil {
		return &result, err
	}

	minETHAmountWithDecimals := go_decimal.
		MustStart(amountOutWithDecimals).
		MustMulti(1 - realOpts.Slippage).
		RoundDown(0).
		MustEndForBigInt()

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		constants.RouterAbiStr,
		"swapExactTokensForETHSupportingFeeOnTransferTokens",
		&go_coin_eth.CallMethodOpts{
			GasLimit:     realOpts.GasLimit,
			MaxFeePerGas: realOpts.MaxFeePerGas,
		},
		[]any{
			tokenAmountWithDecimals,
			minETHAmountWithDecimals,
			[]common.Address{
				common.HexToAddress(tokenAddress),
				common.HexToAddress(realOpts.WETHAddress),
			},
			common.HexToAddress(selfAddress),
			go_decimal.MustStart(time.Now().Unix()).Round(0).MustAdd(200).MustEndForBigInt(),
		},
	)
	if err != nil {
		return &result, err
	}
	result.TxId = btr.SignedTx.Hash().String()
	t.logger.InfoF("出售 txid <%s> 等待确认", result.TxId)
	tr, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return &result, err
	}
	t.logger.InfoF("<%s> 已确认", result.TxId)

	ethAmountWithDecimals, err := t.receivedETHAmountInLogs(tr.Logs, realOpts.WETHAddress)
	if err != nil {
		return &result, err
	}

	result.ETHAmount = go_decimal.MustStart(ethAmountWithDecimals).MustUnShiftedBy(18).EndForString()
	result.SellFee = go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
	result.Fee = go_decimal.MustStart(result.ApproveFee).MustAddForString(result.SellFee)
	return &result, nil
}

func (t *Trader) receivedTokenAmountInLogs(
	logs []*types.Log,
	tokenAddress string,
	myAddress string,
) (*big.Int, error) {
	result := big.NewInt(0)

	transferLogs, err := t.wallet.FilterLogs(
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
		err := t.wallet.UnpackLog(
			&transferEvent,
			go_coin_eth.Erc20AbiStr,
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

func (t *Trader) receivedETHAmountInLogs(
	logs []*types.Log,
	wethAddress string,
) (*big.Int, error) {
	withdrawalLogs, err := t.wallet.FilterLogs(
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
	err = t.wallet.UnpackLog(&withdrawalEvent, go_coin_eth.WETHAbiStr, "Withdrawal", withdrawalLogs[len(withdrawalLogs)-1])
	if err != nil {
		return nil, err
	}

	return withdrawalEvent.Wad, nil
}

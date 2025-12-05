package uniswap_v2_trade

// 同样适用于 pancake V2，fourmeme 都是进入到这个版本

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	type_ "github.com/pefish/go-coin-eth/type"
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

func (t *Trader) WETHAddressFromRouter(routerAddress common.Address) (common.Address, error) {
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
		return common.Address{}, err
	}
	return wethAddress, nil
}

func (t *Trader) GetAmountsOut(
	routerAddress common.Address,
	amountInWithDecimals *big.Int,
	path []common.Address,
) (amountOutWithDecimals_ *big.Int, err_ error) {
	if len(path) > 2 {
		return nil, errors.New("Length of path must be 2.")
	}
	results := make([]*big.Int, 0)
	err := t.wallet.CallContractConstant(
		&results,
		routerAddress,
		constants.RouterAbiStr,
		"getAmountsOut",
		nil,
		[]any{
			amountInWithDecimals,
			[]common.Address{
				path[0],
				path[1],
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return results[1], nil
}

type TradeOpts struct {
	WETHAddress  common.Address
	Slippage     float64 // 滑点，默认 0.5%
	GasLimit     uint64
	MaxFeePerGas *big.Int
}

func (t *Trader) BuyByExactETH(
	ctx context.Context,
	priv string,
	ethAmountWithDecimals *big.Int,
	routerAddress common.Address,
	tokenAddress common.Address,
	opts *TradeOpts,
) (*type_.SwapResultType, error) {

	var realOpts TradeOpts
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

	balanceWithDecimals, err := t.wallet.Balance(selfAddress)
	if err != nil {
		return nil, err
	}

	if balanceWithDecimals.Add(balanceWithDecimals, big.NewInt(10000000000000000)).Cmp(ethAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s + 10000000000000000", balanceWithDecimals, ethAmountWithDecimals)
	}

	if realOpts.WETHAddress.Cmp(go_coin_eth.ZeroAddress) == 0 {
		wethAddress_, err := t.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	amountOutWithDecimals, err := t.GetAmountsOut(
		routerAddress,
		ethAmountWithDecimals,
		[]common.Address{realOpts.WETHAddress, tokenAddress},
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
				realOpts.WETHAddress,
				tokenAddress,
			},
			selfAddress,
			go_decimal.MustStart(time.Now().Unix()).Round(0).MustAdd(200).MustEndForBigInt(),
		},
	)
	if err != nil {
		return nil, err
	}
	txId := btr.SignedTx.Hash().String()
	t.logger.DebugF("购买 txid <%s> 等待确认", txId)
	txReceipt, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}
	t.logger.DebugF("<%s> 已确认", txId)
	tokenAmountWithDecimals, err := t.receivedTokenAmountInLogs(
		txReceipt.Logs,
		tokenAddress,
		selfAddress,
	)
	if err != nil {
		return nil, err
	}

	return &type_.SwapResultType{
		Type_:                   "buy",
		UserAddress:             selfAddress,
		ETHAmountWithDecimals:   ethAmountWithDecimals,
		TokenAmountWithDecimals: tokenAmountWithDecimals,
		TokenAddress:            tokenAddress,
		NetworkFee:              go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:                    txReceipt.TxHash.String(),
		BlockNumber:             txReceipt.BlockNumber.Uint64(),
	}, nil
}

func (t *Trader) SellByExactToken(
	ctx context.Context,
	priv string,
	tokenAmountWithDecimals *big.Int,
	routerAddress common.Address,
	tokenAddress common.Address,
	opts *TradeOpts,
) (*type_.SwapResultType, error) {
	var realOpts TradeOpts
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

	tokenBalanceWithDecimals, err := t.wallet.TokenBalance(tokenAddress, selfAddress)
	if err != nil {
		return nil, err
	}
	if tokenBalanceWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s", tokenBalanceWithDecimals, tokenAmountWithDecimals)
	}

	if realOpts.WETHAddress.Cmp(go_coin_eth.ZeroAddress) == 0 {
		wethAddress_, err := t.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	amountOutWithDecimals, err := t.GetAmountsOut(
		routerAddress,
		tokenAmountWithDecimals,
		[]common.Address{tokenAddress, realOpts.WETHAddress},
	)
	if err != nil {
		return nil, err
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
				tokenAddress,
				realOpts.WETHAddress,
			},
			selfAddress,
			go_decimal.MustStart(time.Now().Unix()).Round(0).MustAdd(200).MustEndForBigInt(),
		},
	)
	if err != nil {
		return nil, err
	}
	txId := btr.SignedTx.Hash().String()
	t.logger.DebugF("出售 txid <%s> 等待确认", txId)
	txReceipt, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}
	t.logger.DebugF("<%s> 已确认", txId)

	ethAmountWithDecimals, err := t.receivedETHAmountInLogs(txReceipt.Logs, realOpts.WETHAddress)
	if err != nil {
		return nil, err
	}

	return &type_.SwapResultType{
		Type_:                   "sell",
		UserAddress:             selfAddress,
		ETHAmountWithDecimals:   ethAmountWithDecimals,
		TokenAmountWithDecimals: tokenAmountWithDecimals,
		TokenAddress:            tokenAddress,
		NetworkFee:              go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:                    txReceipt.TxHash.String(),
		BlockNumber:             txReceipt.BlockNumber.Uint64(),
	}, nil
}

func (t *Trader) receivedTokenAmountInLogs(
	logs []*types.Log,
	tokenAddress common.Address,
	myAddress common.Address,
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
		if transferEvent.To.Cmp(myAddress) != 0 {
			continue
		}
		result.Add(result, transferEvent.Value)
	}

	return result, nil
}

func (t *Trader) receivedETHAmountInLogs(
	logs []*types.Log,
	wethAddress common.Address,
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

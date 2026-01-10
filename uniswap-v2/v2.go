package uniswap_v2

// 同样适用于 pancake V2，fourmeme 都是进入到这个版本

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	type_ "github.com/pefish/go-coin-eth/type"
	go_decimal "github.com/pefish/go-decimal"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

type PoolKeyType struct {
	Token0 common.Address
	Token1 common.Address
}

func (t *PoolKeyType) ToPoolInfo(tokenAddress common.Address) *PoolInfoType {
	var baseTokenAddress common.Address
	if tokenAddress == t.Token0 {
		baseTokenAddress = t.Token1
	} else {
		baseTokenAddress = t.Token0
	}
	return &PoolInfoType{
		TokenAddress:     tokenAddress,
		BaseTokenAddress: baseTokenAddress,
	}
}

type PoolInfoType struct {
	TokenAddress     common.Address
	BaseTokenAddress common.Address
}

type UniswapV2 struct {
	wallet *go_coin_eth.Wallet
	logger i_logger.ILogger
}

func New(
	logger i_logger.ILogger,
	wallet *go_coin_eth.Wallet,
) *UniswapV2 {
	return &UniswapV2{
		wallet: wallet,
		logger: logger,
	}
}

func (t *UniswapV2) WETHAddressFromRouter(routerAddress common.Address) (common.Address, error) {
	var wethAddress common.Address
	err := t.wallet.CallContractConstant(
		&wethAddress,
		routerAddress,
		RouterABI,
		"WETH",
		nil,
		nil,
	)
	if err != nil {
		return common.Address{}, err
	}
	return wethAddress, nil
}

func (t *UniswapV2) GetAmountsOut(
	routerAddress common.Address,
	poolInfo *PoolInfoType,
	tokenIn common.Address,
	amountInWithDecimals *big.Int,
) (amountOutWithDecimals_ *big.Int, err_ error) {
	var tokenOut common.Address
	if tokenIn == poolInfo.TokenAddress {
		tokenOut = poolInfo.BaseTokenAddress
	} else {
		tokenOut = poolInfo.TokenAddress
	}

	results := make([]*big.Int, 0)
	err := t.wallet.CallContractConstant(
		&results,
		routerAddress,
		RouterABI,
		"getAmountsOut",
		nil,
		[]any{
			amountInWithDecimals,
			[]common.Address{
				tokenIn,
				tokenOut,
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

func (t *UniswapV2) BuyByExactETH(
	ctx context.Context,
	priv string,
	ethAmountWithDecimals *big.Int,
	routerAddress common.Address,
	poolInfo *PoolInfoType,
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
		poolInfo,
		realOpts.WETHAddress,
		ethAmountWithDecimals,
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
		RouterABI,
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
	txReceipt, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}
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

func (t *UniswapV2) SellByExactToken(
	ctx context.Context,
	priv string,
	tokenAmountWithDecimals *big.Int,
	routerAddress common.Address,
	poolInfo *PoolInfoType,
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
		poolInfo,
		tokenAddress,
		tokenAmountWithDecimals,
	)
	if err != nil {
		return nil, err
	}

	minETHAmountWithDecimals := go_decimal.
		MustStart(amountOutWithDecimals).
		MustMulti(1 - realOpts.Slippage).
		RoundDown(0).
		MustEndForBigInt()
	// t.logger.InfoF(
	// 	"<tokenAmountWithDecimals: %s> <amountOutWithDecimals: %s> <slippage: %f> <minETHAmountWithDecimals: %s>",
	// 	tokenAmountWithDecimals.String(),
	// 	amountOutWithDecimals.String(),
	// 	realOpts.Slippage,
	// 	minETHAmountWithDecimals.String(),
	// )
	// minETHAmountWithDecimals = big.NewInt(0)

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		RouterABI,
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
	txReceipt, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}

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

func (t *UniswapV2) receivedTokenAmountInLogs(
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

func (t *UniswapV2) receivedETHAmountInLogs(
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

type PancakePoolInfoType struct {
	ID      string `json:"id"`
	FeeTier int    `json:"feeTier"`
	Token0  struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"token0"`
	Token1 struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"token1"`
	TotalVolumeUSD float64 `json:"totalVolumeUSD,string"`
	TvlUSD         float64 `json:"tvlUSD,string"`
	TvlToken0      float64 `json:"tvlToken0,string"`
	TvlToken1      float64 `json:"tvlToken1,string"`
}

func (t *UniswapV2) SearchPancakePairs(tokenAddress common.Address) ([]*PancakePoolInfoType, error) {
	var httpResult struct {
		Tokens []struct {
			Id             string  `json:"id"`
			Decimals       int     `json:"decimals"`
			Name           string  `json:"name"`
			Symbol         string  `json:"symbol"`
			TotalTxCount   int64   `json:"totalTxCount"`
			PriceUSD       float64 `json:"priceUSD,string"`
			TotalVolumeUSD float64 `json:"totalVolumeUSD,string"`
			Tvl            float64 `json:"tvl,string"`
			TvlUSD         float64 `json:"tvlUSD,string"`
		} `json:"tokens"`
		Pools []*PancakePoolInfoType `json:"pools"`
	}
	_, _, err := go_http.HttpInstance.GetForStruct(
		t.logger,
		&go_http.RequestParams{
			Url: "https://explorer.pancakeswap.com/api/cached/protocol/v2/bsc/search",
			Queries: map[string]string{
				"text": tokenAddress.String(),
			},
		},
		&httpResult,
	)
	if err != nil {
		return nil, err
	}
	return httpResult.Pools, nil
}

package uniswap_v3_trade

// v3 中，pool 由 token1、token2、fee 三个参数唯一确定
// 同样适用于 pancake V3

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	type_ "github.com/pefish/go-coin-eth/type"
	constants "github.com/pefish/go-coin-eth/uniswap-v3-trade/constant"
	go_decimal "github.com/pefish/go-decimal"
	go_http "github.com/pefish/go-http"
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
		constants.RouterABIStr,
		"WETH9",
		nil,
		nil,
	)
	if err != nil {
		return common.Address{}, err
	}
	return wethAddress, nil
}

type BuyByExactETHOpts struct {
	WETHAddress                       common.Address
	MinReceiveTokenAmountWithDecimals *big.Int
	GasLimit                          uint64
	MaxFeePerGas                      *big.Int
}

// 只能 route v3 pool，会查找 bnb、token、fee 三个参数唯一确定的 pool 去进行交易，没有这个池子就会失败
func (t *Trader) BuyByExactETH(
	ctx context.Context,
	priv string,
	ethAmountWithDecimals *big.Int,
	routerAddress common.Address,
	tokenAddress common.Address,
	fee uint64,
	opts *BuyByExactETHOpts,
) (*type_.SwapResultType, error) {
	var realOpts BuyByExactETHOpts
	if opts != nil {
		realOpts = *opts
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 300000
	}

	if realOpts.MinReceiveTokenAmountWithDecimals == nil {
		realOpts.MinReceiveTokenAmountWithDecimals = big.NewInt(0)
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

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		constants.RouterABIStr,
		"exactInputSingle",
		&go_coin_eth.CallMethodOpts{
			Value:        ethAmountWithDecimals,
			GasLimit:     realOpts.GasLimit,
			MaxFeePerGas: realOpts.MaxFeePerGas,
		},
		[]any{
			struct {
				TokenIn           common.Address
				TokenOut          common.Address
				Fee               *big.Int
				Recipient         common.Address
				AmountIn          *big.Int
				AmountOutMinimum  *big.Int
				SqrtPriceLimitX96 *big.Int
			}{
				TokenIn:           realOpts.WETHAddress,
				TokenOut:          tokenAddress,
				Fee:               big.NewInt(int64(fee)),
				Recipient:         selfAddress,
				AmountIn:          ethAmountWithDecimals,
				AmountOutMinimum:  realOpts.MinReceiveTokenAmountWithDecimals,
				SqrtPriceLimitX96: big.NewInt(0),
			},
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

func (t *Trader) receivedTokenAmountInLogs(
	logs []*types.Log,
	tokenAddress common.Address,
	myAddress common.Address,
) (*big.Int, error) {
	result := big.NewInt(0)

	transferLogs, err := t.wallet.FilterLogs(
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", // Transfer of ERC20
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

type SellByExactTokenOpts struct {
	WETHAddress                     common.Address
	MinReceiveETHAmountWithDecimals *big.Int
	GasLimit                        uint64
	MaxFeePerGas                    *big.Int
}

// 只会得到 WBNB，不会自动转换为 BNB
func (t *Trader) SellByExactToken(
	ctx context.Context,
	priv string,
	tokenAmountWithDecimals *big.Int,
	routerAddress common.Address,
	tokenAddress common.Address,
	fee uint64,
	opts *SellByExactTokenOpts,
) (*type_.SwapResultType, error) {
	var realOpts SellByExactTokenOpts
	if opts != nil {
		realOpts = *opts
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 300000
	}

	if realOpts.MinReceiveETHAmountWithDecimals == nil {
		realOpts.MinReceiveETHAmountWithDecimals = big.NewInt(0)
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

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		constants.RouterABIStr,
		"exactInputSingle",
		&go_coin_eth.CallMethodOpts{
			GasLimit:     realOpts.GasLimit,
			MaxFeePerGas: realOpts.MaxFeePerGas,
		},
		[]any{
			struct {
				TokenIn           common.Address
				TokenOut          common.Address
				Fee               *big.Int
				Recipient         common.Address
				AmountIn          *big.Int
				AmountOutMinimum  *big.Int
				SqrtPriceLimitX96 *big.Int
			}{
				TokenIn:           tokenAddress,
				TokenOut:          realOpts.WETHAddress,
				Fee:               big.NewInt(int64(fee)),
				Recipient:         selfAddress,
				AmountIn:          tokenAmountWithDecimals,
				AmountOutMinimum:  realOpts.MinReceiveETHAmountWithDecimals,
				SqrtPriceLimitX96: big.NewInt(0),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	txReceipt, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}
	ethAmountWithDecimals, err := t.receivedTokenAmountInLogs(txReceipt.Logs, realOpts.WETHAddress, selfAddress)
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

type PoolInfoType struct {
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
	TVLUSD         float64 `json:"tvlUSD,string"`
	TVLToken0      float64 `json:"tvlToken0,string"`
	TVLToken1      float64 `json:"tvlToken1,string"`
}

func (t *Trader) SearchPancakePairs(tokenAddress common.Address) ([]*PoolInfoType, error) {
	var httpResult struct {
		Tokens []struct {
			Id             string  `json:"id"`
			Decimals       int     `json:"decimals"`
			Name           string  `json:"name"`
			Symbol         string  `json:"symbol"`
			TotalTxCount   int64   `json:"totalTxCount"`
			PriceUSD       float64 `json:"priceUSD,string"`
			TotalVolumeUSD float64 `json:"totalVolumeUSD,string"`
			TVL            float64 `json:"tvl,string"`
			TVLUSD         float64 `json:"tvlUSD,string"`
		} `json:"tokens"`
		Pools []*PoolInfoType `json:"pools"`
	}
	_, _, err := go_http.HttpInstance.GetForStruct(
		t.logger,
		&go_http.RequestParams{
			Url: "https://explorer.pancakeswap.com/api/cached/protocol/v3/bsc/search",
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

package uniswap_v3

// v3 中，pool 由 token1、token2、fee 三个参数唯一确定
// 同样适用于 pancake V3

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	go_decimal "github.com/pefish/go-decimal"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
)

type UniswapV3 struct {
	wallet *go_coin_eth.Wallet
	logger i_logger.ILogger
}

func New(
	logger i_logger.ILogger,
	wallet *go_coin_eth.Wallet,
) *UniswapV3 {
	return &UniswapV3{
		wallet: wallet,
		logger: logger,
	}
}

func (t *UniswapV3) WETHAddressFromRouter(routerAddress common.Address) (common.Address, error) {
	var wethAddress common.Address
	err := t.wallet.CallContractConstant(
		&wethAddress,
		routerAddress,
		RouterABI,
		"WETH9",
		nil,
		nil,
	)
	if err != nil {
		return common.Address{}, err
	}
	return wethAddress, nil
}

type SwapResultType struct {
	UserAddress                   common.Address
	InputToken                    common.Address
	InputTokenAmountWithDecimals  *big.Int
	OutputToken                   common.Address
	OutputTokenAmountWithDecimals *big.Int
	NetworkFee                    *big.Int
	TxId                          string
	BlockNumber                   uint64
}

type SwapExactInputOpts struct {
	WETHAddress                       common.Address
	MinReceiveTokenAmountWithDecimals *big.Int
	GasLimit                          uint64
	MaxFeePerGas                      *big.Int
}

// 只能 route v3 pool，会查找 bnb、token、fee 三个参数唯一确定的 pool 去进行交易，没有这个池子就会失败
func (t *UniswapV3) SwapExactInput(
	ctx context.Context,
	priv string,
	inputAmountWithDecimals *big.Int,
	routerAddress common.Address,
	inputToken common.Address, // 如果是 0 地址，代表使用 ETH
	outputToken common.Address,
	fee uint64,
	opts *SwapExactInputOpts,
) (*SwapResultType, error) {
	var realOpts SwapExactInputOpts
	if opts != nil {
		realOpts = *opts
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 300000
	}

	if realOpts.MinReceiveTokenAmountWithDecimals == nil {
		realOpts.MinReceiveTokenAmountWithDecimals = big.NewInt(0)
	}

	if realOpts.WETHAddress == go_coin_eth.ZeroAddress {
		wethAddress_, err := t.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	userAddress, err := t.wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0)
	if inputToken == go_coin_eth.ZeroAddress {
		value = inputAmountWithDecimals
	}

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		RouterABI,
		"exactInputSingle",
		&go_coin_eth.CallMethodOpts{
			Value:        value,
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
				TokenIn:           inputToken,
				TokenOut:          outputToken,
				Fee:               big.NewInt(int64(fee)),
				Recipient:         userAddress,
				AmountIn:          inputAmountWithDecimals,
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
	outputAmountWithDecimals, err := t.receivedTokenAmountInLogs(
		txReceipt.Logs,
		outputToken,
		userAddress,
	)
	if err != nil {
		return nil, err
	}

	return &SwapResultType{
		UserAddress:                   userAddress,
		InputToken:                    inputToken,
		InputTokenAmountWithDecimals:  inputAmountWithDecimals,
		OutputTokenAmountWithDecimals: outputAmountWithDecimals,
		OutputToken:                   outputToken,
		NetworkFee:                    go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:                          txReceipt.TxHash.String(),
		BlockNumber:                   txReceipt.BlockNumber.Uint64(),
	}, nil
}

func (t *UniswapV3) receivedTokenAmountInLogs(
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

type PoolKeyType struct {
	Token0 common.Address
	Token1 common.Address
	Fee    uint64
}

func (t *UniswapV3) GetPoolAddress(
	poolKey *PoolKeyType,
	factoryAddress common.Address,
) (common.Address, error) {
	var result common.Address
	err := t.wallet.CallContractConstant(
		&result,
		factoryAddress,
		FactoryABI,
		"getPool",
		nil,
		[]any{
			poolKey.Token0,
			poolKey.Token1,
			big.NewInt(int64(poolKey.Fee)),
		},
	)
	if err != nil {
		return common.Address{}, err
	}
	return result, nil
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

func (t *UniswapV3) SearchPancakePairs(tokenAddress common.Address) ([]*PoolInfoType, error) {
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

package uniswap_v4_trade

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
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
			Url: "https://explorer.pancakeswap.com/api/cached/protocol/infinityCl/bsc/search",
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

type PoolKeyType struct {
	Currency0   common.Address `json:"currency0"`
	Currency1   common.Address `json:"currency1"`
	Hooks       common.Address `json:"hooks"`
	PoolManager common.Address `json:"pool_manager"`
	Fee         *big.Int       `json:"fee"`
	Parameters  [32]byte       `json:"parameters"`
}

type CLSwapExactInSingleParamsType struct {
	PoolKey          PoolKeyType `json:"pool_key"`
	ZeroForOne       bool        `json:"zero_for_one"`
	AmountIn         *big.Int    `json:"amount_in"`
	AmountOutMinimum *big.Int    `json:"amount_out_minimum"`
	HookData         []byte      `json:"hook_data"`
}

type CLSettleAllParamsType struct {
	Address common.Address `json:"address"`
	Amount  *big.Int       `json:"amount"`
}

type CLTakeAllParamsType struct {
	Address common.Address `json:"address"`
	Amount  *big.Int       `json:"amount"`
}

var CLSwapExactInSingleParamsAbiType, _ = abi.NewType("tuple", "CLSwapExactInSingleParamsType", []abi.ArgumentMarshaling{
	{Type: "tuple", Name: "pool_key", Components: []abi.ArgumentMarshaling{
		{Type: "address", Name: "currency0"},
		{Type: "address", Name: "currency1"},
		{Type: "address", Name: "hooks"},
		{Type: "address", Name: "pool_manager"},
		{Type: "uint24", Name: "fee"},
		{Type: "bytes32", Name: "parameters"},
	}},
	{Type: "bool", Name: "zero_for_one"},
	{Type: "uint128", Name: "amount_in"},
	{Type: "uint128", Name: "amount_out_minimum"},
	{Type: "bytes", Name: "hook_data"},
})

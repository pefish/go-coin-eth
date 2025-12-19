package uniswap_v4

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
)

type UniswapV4 struct {
	wallet *go_coin_eth.Wallet
	logger i_logger.ILogger
}

func New(
	logger i_logger.ILogger,
	wallet *go_coin_eth.Wallet,
) *UniswapV4 {
	return &UniswapV4{
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

func (t *UniswapV4) SearchPancakePairs(tokenAddress common.Address) ([]*PoolInfoType, error) {
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

func (t *UniswapV4) PairInfoByPoolID(poolID common.Hash) (*PoolKeyType, error) {
	var result PoolKeyType
	err := t.wallet.CallContractConstant(
		&result,
		CL_Pool_Manager,
		CL_Pool_Manager_ABI,
		"poolIdToPoolKey",
		nil,
		[]any{
			poolID,
		},
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type PoolKeyType struct {
	Currency0   common.Address `json:"currency0"`
	Currency1   common.Address `json:"currency1"`
	Hooks       common.Address `json:"hooks"`
	PoolManager common.Address `json:"pool_manager"`
	Fee         *big.Int       `json:"fee"`
	Parameters  [32]byte       `json:"parameters"`
}

package uniswap_v4

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type QuoteResultType struct {
	AmountOut   *big.Int
	GasEstimate *big.Int
	TokenOut    common.Address
}

type QuoteExactSingleParamsType struct {
	PoolKey     PoolKeyType `json:"pool_key"`
	ZeroForOne  bool        `json:"zero_for_one"`
	ExactAmount *big.Int    `json:"exact_amount"`
	HookData    []byte      `json:"hook_data"`
}

func (t *UniswapV4) QuoteExactInputSingle(
	quoterAddress common.Address,
	poolKey *PoolKeyType,
	tokenIn common.Address,
	amountInWithDecimals *big.Int,
) (*QuoteResultType, error) {
	zeroForOne := false
	var tokenOut common.Address
	if tokenIn == poolKey.Currency0 {
		zeroForOne = true
		tokenOut = poolKey.Currency1
	} else {
		tokenOut = poolKey.Currency0
	}

	var result struct {
		AmountOut   *big.Int
		GasEstimate *big.Int
	}

	err := t.wallet.CallContractConstant(
		&result,
		quoterAddress,
		Quoter_ABI,
		"quoteExactInputSingle",
		nil,
		[]any{
			QuoteExactSingleParamsType{
				PoolKey:     *poolKey,
				ZeroForOne:  zeroForOne,
				ExactAmount: amountInWithDecimals,
				HookData:    []byte{},
			},
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), "execution reverted") {
			return nil, errors.New("no pool for this tokenIn, tokenOut, fee")
		}
		return nil, err
	}
	return &QuoteResultType{
		AmountOut:   result.AmountOut,
		GasEstimate: result.GasEstimate,
		TokenOut:    tokenOut,
	}, nil
}

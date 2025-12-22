package uniswap_v3

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type QuoteResultType struct {
	AmountOut   *big.Int
	GasEstimate *big.Int
	TokenOut    common.Address
}

func (t *UniswapV3) QuoteExactInputSingle(
	quoterAddress common.Address,
	poolKey *PoolKeyType,
	tokenIn common.Address,
	amountInWithDecimals *big.Int,
) (*QuoteResultType, error) {
	var tokenOut common.Address
	if tokenIn == poolKey.Token0 {
		tokenOut = poolKey.Token1
	} else {
		tokenOut = poolKey.Token0
	}

	var result struct {
		AmountOut               *big.Int
		SqrtPriceX96After       *big.Int
		InitializedTicksCrossed uint32
		GasEstimate             *big.Int
	}

	err := t.wallet.CallContractConstant(
		&result,
		quoterAddress,
		QuoterABI,
		"quoteExactInputSingle",
		nil,
		[]any{
			struct {
				TokenIn           common.Address
				TokenOut          common.Address
				Fee               *big.Int
				AmountIn          *big.Int
				SqrtPriceLimitX96 *big.Int
			}{
				TokenIn:           tokenIn,
				TokenOut:          tokenOut,
				Fee:               big.NewInt(int64(poolKey.Fee)),
				AmountIn:          amountInWithDecimals,
				SqrtPriceLimitX96: big.NewInt(0),
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

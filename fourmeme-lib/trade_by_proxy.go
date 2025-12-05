package fourmeme_lib

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	"github.com/pefish/go-coin-eth/fourmeme-lib/constant"
)

type TradeEventType struct {
	Token   common.Address `json:"token"`
	Account common.Address `json:"account"`
	Price   *big.Int       `json:"price"`
	Amount  *big.Int       `json:"amount"`
	Cost    *big.Int       `json:"cost"`
	Fee     *big.Int       `json:"fee"`
	Offers  *big.Int       `json:"offers"`
	Funds   *big.Int       `json:"funds"`
}

func BuyByProxy(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress string,
	tokenAmountWithDecimals *big.Int,
	maxCostBNBAmount *big.Int,
	maxFeePerGas *big.Int, // 1 大概 0.1 刀
) (*SwapInfoType, error) {
	btr, err := wallet.BuildCallMethodTx(
		priv,
		constant.FourmemeToolAddress,
		constant.FourmemeToolABI,
		"buyPefish",
		&go_coin_eth.CallMethodOpts{
			MaxFeePerGas:   maxFeePerGas,
			GasLimit:       250000,
			IsPredictError: false,
		},
		[]any{
			common.HexToAddress(tokenAddress),
			tokenAmountWithDecimals,
			maxCostBNBAmount,
		},
	)
	if err != nil {
		return nil, err
	}
	tr, err := wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}

	swapInfos, err := ParseSwapInfos(wallet, tr)
	if err != nil {
		return nil, err
	}

	return swapInfos[0], nil
}

func SellByProxy(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress string,
	tokenAmountWithDecimals *big.Int,
	minReceiveBnbAmount *big.Int,
	maxFeePerGas *big.Int,
) (*SwapInfoType, error) {
	btr, err := wallet.BuildCallMethodTx(
		priv,
		constant.FourmemeToolAddress,
		constant.FourmemeToolABI,
		"sellPefish",
		&go_coin_eth.CallMethodOpts{
			MaxFeePerGas:   maxFeePerGas,
			GasLimit:       250000,
			IsPredictError: false,
		},
		[]any{
			common.HexToAddress(tokenAddress),
			tokenAmountWithDecimals,
			minReceiveBnbAmount,
		},
	)
	if err != nil {
		return nil, err
	}
	tr, err := wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}
	swapInfos, err := ParseSwapInfos(wallet, tr)
	if err != nil {
		return nil, err
	}

	return swapInfos[0], nil
}

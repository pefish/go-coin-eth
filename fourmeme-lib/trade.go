package fourmeme_lib

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	"github.com/pefish/go-coin-eth/fourmeme-lib/constant"
)

func Buy(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress string,
	bnbAmountWithDecimals *big.Int,
	minReceiveTokenAmount *big.Int,
	maxFeePerGas *big.Int, // 1 大概 0.1 刀
) (*SwapInfoType, error) {
	btr, err := wallet.BuildCallMethodTx(
		priv,
		constant.TokenManagerAddress,
		constant.TokenManagerABI,
		"buyTokenAMAP1",
		&go_coin_eth.CallMethodOpts{
			MaxFeePerGas:   maxFeePerGas,
			GasLimit:       250000,
			IsPredictError: false,
			Value:          bnbAmountWithDecimals,
		},
		[]any{
			common.HexToAddress(tokenAddress),
			bnbAmountWithDecimals,
			minReceiveTokenAmount,
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

func Sell(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress string,
	tokenAmountWithDecimals *big.Int,
	maxFeePerGas *big.Int,
) (*SwapInfoType, error) {
	btr, err := wallet.BuildCallMethodTx(
		priv,
		constant.TokenManagerAddress,
		constant.TokenManagerABI,
		"sellToken4",
		&go_coin_eth.CallMethodOpts{
			MaxFeePerGas:   maxFeePerGas,
			GasLimit:       250000,
			IsPredictError: false,
		},
		[]any{
			common.HexToAddress(tokenAddress),
			tokenAmountWithDecimals,
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

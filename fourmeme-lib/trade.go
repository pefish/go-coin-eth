package fourmeme_lib

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	"github.com/pefish/go-coin-eth/fourmeme-lib/constant"
	type_ "github.com/pefish/go-coin-eth/type"
)

func Buy(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress common.Address,
	bnbAmountWithDecimals *big.Int,
	minReceiveTokenAmount *big.Int,
	maxFeePerGas *big.Int, // 1 大概 0.1 刀
) (*type_.SwapResultType, *TradeEventType, error) {
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
			tokenAddress,
			bnbAmountWithDecimals,
			minReceiveTokenAmount,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	tr, err := wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, nil, err
	}

	swapInfos, tradeEvents, err := ParseSwapInfos(wallet, tr)
	if err != nil {
		return nil, nil, err
	}

	return swapInfos[0], tradeEvents[0], nil
}

func Sell(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress common.Address,
	tokenAmountWithDecimals *big.Int,
	maxFeePerGas *big.Int,
) (*type_.SwapResultType, *TradeEventType, error) {
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
			tokenAddress,
			tokenAmountWithDecimals,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	tr, err := wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, nil, err
	}

	swapInfos, tradeEvents, err := ParseSwapInfos(wallet, tr)
	if err != nil {
		return nil, nil, err
	}

	return swapInfos[0], tradeEvents[0], nil
}

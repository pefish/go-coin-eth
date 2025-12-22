package uniswap_universal_router

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v2 "github.com/pefish/go-coin-eth/uniswap-v2"
	go_decimal "github.com/pefish/go-decimal"
	"github.com/pkg/errors"
)

// 如果 baseToken 是 WBNB，会自动 Wrap 和 Unwrap
func (t *Router) SwapExactInputV2(
	ctx context.Context,
	priv string,
	poolInfo *uniswap_v2.PoolInfoType,
	tokenIn common.Address,
	amountInWithDecimals *big.Int,
	v2RouterAddress common.Address,
	opts *SwapOpts,
) (*SwapResultType, error) {
	var realOpts SwapOpts
	if opts != nil {
		realOpts = *opts
	}

	if realOpts.Slippage == 0 {
		realOpts.Slippage = 50
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 300000
	}

	if realOpts.MaxFeePerGas == nil {
		realOpts.MaxFeePerGas = big.NewInt(1_0000_0000)
	}

	userAddress, err := t.wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}

	if realOpts.Nonce == 0 {
		nonce, err := t.wallet.NextNonce(userAddress)
		if err != nil {
			return nil, err
		}
		realOpts.Nonce = nonce
	}

	if realOpts.Slippage > 10000 {
		return nil, errors.New("slipage too high")
	}
	uniswapV2 := uniswap_v2.New(t.logger, t.wallet)

	if tokenIn != poolInfo.TokenAddress &&
		tokenIn != poolInfo.BaseTokenAddress {
		return nil, errors.New("tokenIn is not in pool")
	}

	var tokenOut common.Address
	if tokenIn == poolInfo.TokenAddress {
		tokenOut = poolInfo.BaseTokenAddress
	} else {
		tokenOut = poolInfo.TokenAddress
	}

	payerIsUser := true
	commands := make([]byte, 0)
	params := make([][]byte, 0)
	value := big.NewInt(0)
	if tokenIn == go_coin_eth.WBNBAddress {
		value = amountInWithDecimals
		commands = append(commands, 0x0b) // WRAP_ETH
		wrapETHPayloadBytes, err := abi.Arguments{
			{
				Name:    "recipient",
				Type:    go_coin_eth.TypeAddress,
				Indexed: false,
			},
			{
				Name:    "amount",
				Type:    go_coin_eth.TypeUint256,
				Indexed: false,
			},
		}.Pack(
			Universal_Router,
			amountInWithDecimals,
		)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		params = append(params, wrapETHPayloadBytes)
		payerIsUser = false
	} else {
		newNonce, err := t.ApproveAsync(
			ctx,
			priv,
			tokenIn,
			amountInWithDecimals,
			realOpts.MaxFeePerGas,
			realOpts.Nonce,
		)
		if err != nil {
			return nil, err
		}
		realOpts.Nonce = newNonce
	}

	var swapRecipient common.Address
	if tokenOut == go_coin_eth.WBNBAddress {
		swapRecipient = Universal_Router
	} else {
		swapRecipient = userAddress
	}

	amountOutMinimum := big.NewInt(0)
	if realOpts.Slippage > 0 {
		amountOut, err := uniswapV2.GetAmountsOut(
			v2RouterAddress,
			poolInfo,
			tokenIn,
			amountInWithDecimals,
		)
		if err != nil {
			return nil, err
		}
		amountOutMinimum = go_decimal.MustStart(amountOut).
			MustMulti(
				(10000 - float64(realOpts.Slippage)) / 10000,
			).RoundDown(0).MustEndForBigInt()
	}

	commands = append(commands, 0x08) // Commands.V2_SWAP_EXACT_IN
	v2PayloadBytes, err := abi.Arguments{
		{
			Name:    "recipient",
			Type:    go_coin_eth.TypeAddress,
			Indexed: false,
		},
		{
			Name:    "amountIn",
			Type:    go_coin_eth.TypeUint256,
			Indexed: false,
		},
		{
			Name:    "amountOutMin",
			Type:    go_coin_eth.TypeUint256,
			Indexed: false,
		},
		{
			Name:    "path",
			Type:    go_coin_eth.TypeAddressArr,
			Indexed: false,
		},
		{
			Name:    "payerIsUser",
			Type:    go_coin_eth.TypeBool,
			Indexed: false,
		},
	}.Pack(
		swapRecipient,
		amountInWithDecimals,
		amountOutMinimum,
		[]common.Address{
			tokenIn,
			tokenOut,
		},
		payerIsUser,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params = append(params, v2PayloadBytes)

	if tokenOut == go_coin_eth.WBNBAddress {
		commands = append(commands, 0x0c) // UNWRAP_WETH
		unwrapETHPayloadBytes, err := abi.Arguments{
			{
				Name:    "recipient",
				Type:    go_coin_eth.TypeAddress,
				Indexed: false,
			},
			{
				Name:    "amountMin",
				Type:    go_coin_eth.TypeUint256,
				Indexed: false,
			},
		}.Pack(
			userAddress,
			amountOutMinimum,
		)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		params = append(params, unwrapETHPayloadBytes)
	}

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		Universal_Router,
		Universal_Router_ABI,
		"execute",
		&go_coin_eth.CallMethodOpts{
			GasLimit:     realOpts.GasLimit,
			MaxFeePerGas: realOpts.MaxFeePerGas,
			Value:        value,
			Nonce:        realOpts.Nonce,
		},
		[]any{
			commands,
			params,
		},
	)
	if err != nil {
		return nil, err
	}
	txReceipt, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}

	swapLogs, err := t.wallet.FilterLogs(
		"0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822", // swap event of PancakeV2Pool
		go_coin_eth.ZeroAddress, // any address
		txReceipt.Logs,
	)
	if err != nil {
		return nil, err
	}
	var swapEvent struct {
		Sender     common.Address
		Amount0In  *big.Int
		Amount1In  *big.Int
		Amount0Out *big.Int
		Amount1Out *big.Int
		To         common.Address
	}
	err = t.wallet.UnpackLog(
		&swapEvent,
		uniswap_v2.PoolABI,
		"Swap",
		swapLogs[0],
	)
	if err != nil {
		return nil, err
	}

	return &SwapResultType{
		UserAddress:                  userAddress,
		InputToken:                   tokenIn,
		InputTokenAmountWithDecimals: amountInWithDecimals,
		OutputToken:                  tokenOut,
		OutputTokenAmountWithDecimals: func() *big.Int {
			if swapEvent.Amount0In.Cmp(amountInWithDecimals) == 0 {
				return swapEvent.Amount1Out
			} else {
				return swapEvent.Amount0Out
			}
		}(),
		NetworkFee:  go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:        txReceipt.TxHash.String(),
		BlockNumber: txReceipt.BlockNumber.Uint64(),
		Liquidity:   big.NewInt(0),
		Fee:         big.NewInt(0),
		ProtocolFee: big.NewInt(0),
	}, nil
}

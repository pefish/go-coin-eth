package uniswap_universal_router

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v4 "github.com/pefish/go-coin-eth/uniswap-v4"
	go_decimal "github.com/pefish/go-decimal"
	"github.com/pkg/errors"
)

func (t *Router) SwapExactInputV4(
	ctx context.Context,
	priv string,
	poolKey *uniswap_v4.PoolKeyType,
	tokenIn common.Address,
	amountInWithDecimals *big.Int,
	v4QuoterAddress common.Address,
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
	uniswapV4 := uniswap_v4.New(t.logger, t.wallet)

	zeroForOne := false
	var tokenOut common.Address
	if tokenIn == poolKey.Currency0 {
		zeroForOne = true
		tokenOut = poolKey.Currency1
	} else {
		tokenOut = poolKey.Currency0
	}

	amountOutMinimum := big.NewInt(0)
	if realOpts.Slippage > 0 {
		quoteResult, err := uniswapV4.QuoteExactInputSingle(
			v4QuoterAddress,
			poolKey,
			tokenIn,
			amountInWithDecimals,
		)
		if err != nil {
			return nil, err
		}
		amountOutMinimum = go_decimal.MustStart(quoteResult.AmountOut).
			MustMulti(
				(10000 - float64(realOpts.Slippage)) / 10000,
			).RoundDown(0).MustEndForBigInt()
	}

	params := uniswap_v4.CLSwapExactInSingleParamsType{
		PoolKey:          *poolKey,
		ZeroForOne:       zeroForOne,
		AmountIn:         amountInWithDecimals,
		AmountOutMinimum: amountOutMinimum,
		HookData:         []byte{},
	}
	swapParams, err := abi.Arguments{
		{
			Type: uniswap_v4.CLSwapExactInSingleParamsAbiType,
		},
	}.Pack(params)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	settleAllParams, err := abi.Arguments{
		{
			Type: go_coin_eth.TypeAddress,
		},
		{
			Type: go_coin_eth.TypeUint256,
		},
	}.Pack(tokenIn, amountInWithDecimals)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	takeAllParams, err := abi.Arguments{
		{
			Type: go_coin_eth.TypeAddress,
		},
		{
			Type: go_coin_eth.TypeUint256,
		},
	}.Pack(tokenOut, amountOutMinimum)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	infinityPayloadBytes, err := abi.Arguments{
		{
			Name:    "actions",
			Type:    go_coin_eth.TypeBytes,
			Indexed: false,
		},
		{
			Name:    "params",
			Type:    go_coin_eth.TypeBytesSlice,
			Indexed: false,
		},
	}.Pack(
		[]byte{
			0x06, // CL_SWAP_EXACT_IN_SINGLE
			0x0c, // SETTLE_ALL
			0x0f, // TAKE_ALL
		},
		[][]byte{
			swapParams,
			settleAllParams,
			takeAllParams,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	value := big.NewInt(0)
	if tokenIn == go_coin_eth.ZeroAddress {
		value = amountInWithDecimals
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
			[]byte{0x10}, // Commands.INFI_SWAP
			[][]byte{infinityPayloadBytes},
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
		"0x04206ad2b7c0f463bff3dd4f33c5735b0f2957a351e4f79763a4fa9e775dd237", // swap event of CLPoolManager
		uniswap_v4.CL_Pool_Manager,
		txReceipt.Logs,
	)
	if err != nil {
		return nil, err
	}
	var swapEvent struct {
		Id           [32]byte
		Sender       common.Address
		Amount0      *big.Int
		Amount1      *big.Int
		SqrtPriceX96 *big.Int
		Liquidity    *big.Int
		Tick         *big.Int
		Fee          *big.Int
		ProtocolFee  uint16
	}
	err = t.wallet.UnpackLog(
		&swapEvent,
		uniswap_v4.CL_Pool_Manager_ABI,
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
			if zeroForOne {
				return swapEvent.Amount1
			} else {
				return swapEvent.Amount0
			}
		}(),
		NetworkFee:  go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:        txReceipt.TxHash.String(),
		BlockNumber: txReceipt.BlockNumber.Uint64(),
		Liquidity:   swapEvent.Liquidity,
		Fee:         swapEvent.Fee,
		ProtocolFee: big.NewInt(int64(swapEvent.ProtocolFee)),
	}, nil
}

// 异步方法
func (t *Router) ApproveAsync(
	ctx context.Context,
	userPriv string,
	tokenAddress common.Address,
	amountWithDecimals *big.Int,
	maxFeePerGas *big.Int,
	nonce uint64,
) (uint64, error) {
	userAddress, err := t.wallet.PrivateKeyToAddress(userPriv)
	if err != nil {
		return nonce, err
	}

	// 检查给 permit2 的授权
	allowanceAmount, err := t.wallet.ApprovedAmount(tokenAddress, userAddress, Permit2)
	if err != nil {
		return nonce, err
	}
	t.logger.InfoF("Permit2 approvedAmount: %s", allowanceAmount.String())
	if allowanceAmount.Cmp(amountWithDecimals) < 0 {
		t.logger.InfoF("Permit2 need approve first")
		go func(nonce uint64) {
			tr, err := t.wallet.ApproveWait(
				ctx,
				userPriv,
				tokenAddress,
				Permit2,
				nil,
				&go_coin_eth.CallMethodOpts{
					MaxFeePerGas: maxFeePerGas,
					GasLimit:     50000,
					Nonce:        nonce,
				},
			)
			if err != nil {
				t.logger.ErrorF("ApproveWait Permit2 error: %s", err.Error())
				return
			}
			t.logger.InfoF("Permit2 approve done. txId: %s", tr.TxHash.String())
		}(nonce)
		nonce++
	}

	// 要先检查有没有通过 permit2 给 universal_router 授权
	allowanceInfo, err := t.AllowanceForPermit2(userAddress, tokenAddress, Universal_Router)
	if err != nil {
		return nonce, err
	}
	t.logger.InfoF(
		"Universal_Router <ApprovedAmount: %s> <Expiration: %d>",
		allowanceInfo.Amount.String(),
		allowanceInfo.Expiration.Int64(),
	)
	if allowanceInfo.Amount.Cmp(amountWithDecimals) < 0 ||
		allowanceInfo.Expiration.Int64()*1000 < time.Now().UnixMilli() {
		t.logger.InfoF("Universal_Router need approve first")
		go func(nonce uint64) {
			tr, err := t.ApprovePermit2Wait(
				ctx,
				userPriv,
				tokenAddress,
				Universal_Router,
				nil,
				time.Now().UnixMilli()+3600*1000, // 1 hour
				&go_coin_eth.CallMethodOpts{
					MaxFeePerGas: maxFeePerGas,
					GasLimit:     50000,
					Nonce:        nonce,
				},
			)
			if err != nil {
				t.logger.ErrorF("ApprovePermit2Wait Universal_Router error: %s", err.Error())
				return
			}
			t.logger.InfoF("Universal_Router approve done. txId: %s", tr.TxHash.String())
		}(nonce)
		nonce++
	}
	return nonce, nil
}

package uniswap_universal_router

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v4 "github.com/pefish/go-coin-eth/uniswap-v4"
	go_decimal "github.com/pefish/go-decimal"
	"github.com/pkg/errors"
)

func (t *Router) SwapExactInputV4(
	ctx context.Context,
	priv string,
	pairInfo *uniswap_v4.PoolKeyType,
	tokenIn common.Address,
	amountInWithDecimals *big.Int,
	amountOutMinimum *big.Int,
	gasLimit uint64,
	maxFeePerGas *big.Int,
) (*SwapResultType, error) {
	zeroForOne := false
	var tokenOut common.Address
	if tokenIn == pairInfo.Currency0 {
		zeroForOne = true
		tokenOut = pairInfo.Currency1
	} else {
		tokenOut = pairInfo.Currency0
	}

	params := uniswap_v4.CLSwapExactInSingleParamsType{
		PoolKey:          *pairInfo,
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
		_, err = t.ApproveWait(
			ctx,
			priv,
			tokenIn,
			amountInWithDecimals,
			maxFeePerGas,
		)
		if err != nil {
			return nil, err
		}
	}
	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		Universal_Router,
		Universal_Router_ABI,
		"execute",
		&go_coin_eth.CallMethodOpts{
			GasLimit:     gasLimit,
			MaxFeePerGas: maxFeePerGas,
			Value:        value,
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

	userAddress, _ := t.wallet.PrivateKeyToAddress(priv)
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

func (t *Router) ApproveWait(
	ctx context.Context,
	userPriv string,
	tokenAddress common.Address,
	amountWithDecimals *big.Int,
	maxFeePerGas *big.Int,
) ([]*types.Receipt, error) {
	userAddress, err := t.wallet.PrivateKeyToAddress(userPriv)
	if err != nil {
		return nil, err
	}
	trs := make([]*types.Receipt, 0)
	// 检查给 permit2 的授权
	allowanceAmount, err := t.wallet.ApprovedAmount(tokenAddress, userAddress, Permit2)
	if err != nil {
		return nil, err
	}
	t.logger.InfoF("Permit2 approvedAmount: %s", allowanceAmount.String())
	if allowanceAmount.Cmp(amountWithDecimals) < 0 {
		t.logger.InfoF("Permit2 need approve first")
		tr, err := t.wallet.ApproveWait(
			ctx,
			userPriv,
			tokenAddress,
			Permit2,
			nil,
			&go_coin_eth.CallMethodOpts{
				MaxFeePerGas:   maxFeePerGas,
				GasLimit:       50000,
				IsPredictError: true,
			},
		)
		if err != nil {
			return nil, err
		}
		t.logger.InfoF("Permit2 approve done. txId: %s", tr.TxHash.String())
		trs = append(trs, tr)
	}

	// 要先检查有没有通过 permit2 给 universal_router 授权
	allowanceInfo, err := t.AllowanceForPermit2(userAddress, tokenAddress, Universal_Router)
	if err != nil {
		return trs, err
	}
	t.logger.InfoF(
		"Universal_Router <ApprovedAmount: %s> <Expiration: %d>",
		allowanceInfo.Amount.String(),
		allowanceInfo.Expiration.Int64(),
	)
	if allowanceInfo.Amount.Cmp(amountWithDecimals) < 0 ||
		allowanceInfo.Expiration.Int64()*1000 < time.Now().UnixMilli() {
		t.logger.InfoF("Universal_Router need approve first")
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
			},
		)
		if err != nil {
			return trs, err
		}
		t.logger.InfoF("Universal_Router approve done. txId: %s", tr.TxHash.String())
		trs = append(trs, tr)
	}
	return trs, nil
}

package uniswap_universal_router

import (
	"context"
	"math/big"

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
	pairInfo *uniswap_v4.PoolKeyType,
	tokenIn common.Address,
	amountIn *big.Int,
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
		AmountIn:         amountIn,
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
	}.Pack(tokenIn, amountIn)
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
		value = amountIn
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
		InputTokenAmountWithDecimals: amountIn,
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

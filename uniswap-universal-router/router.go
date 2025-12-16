package uniswap_universal_router

// pancake 的 infinity 就是 v4，分为 CL(Concentrated Liquidity) 池和 Bin 池，CL 是主流
// uniswap_universal 可以执行 v2、v3、v4 的交易

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v4_trade "github.com/pefish/go-coin-eth/uniswap-v4-trade"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

type Router struct {
	wallet *go_coin_eth.Wallet
	logger i_logger.ILogger
}

func New(
	logger i_logger.ILogger,
	wallet *go_coin_eth.Wallet,
) *Router {
	return &Router{
		wallet: wallet,
		logger: logger,
	}
}

func (t *Router) PairInfoByPoolID(poolID common.Hash) (*uniswap_v4_trade.PoolKeyType, error) {
	var result uniswap_v4_trade.PoolKeyType
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

type SwapResultType struct {
	UserAddress                 common.Address
	Currency0                   common.Address
	Currency0AmountWithDecimals *big.Int
	Currency1                   common.Address
	Currency1AmountWithDecimals *big.Int
	NetworkFee                  *big.Int
	TxId                        string
	BlockNumber                 uint64
	Liquidity                   *big.Int
	Fee                         *big.Int
	ProtocolFee                 uint16
}

func (t *Router) SwapExactInput(
	ctx context.Context,
	priv string,
	pairInfo *uniswap_v4_trade.PoolKeyType,
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

	params := uniswap_v4_trade.CLSwapExactInSingleParamsType{
		PoolKey:          *pairInfo,
		ZeroForOne:       zeroForOne,
		AmountIn:         amountIn,
		AmountOutMinimum: amountOutMinimum,
		HookData:         []byte{},
	}
	swapParams, err := abi.Arguments{
		{
			Type: uniswap_v4_trade.CLSwapExactInSingleParamsAbiType,
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

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		Universal_Router,
		Universal_Router_ABI,
		"execute",
		&go_coin_eth.CallMethodOpts{
			GasLimit:     gasLimit,
			MaxFeePerGas: maxFeePerGas,
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
		CL_Pool_Manager,
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
		CL_Pool_Manager_ABI,
		"Swap",
		swapLogs[0],
	)
	if err != nil {
		return nil, err
	}

	userAddress, _ := t.wallet.PrivateKeyToAddress(priv)
	return &SwapResultType{
		UserAddress:                 userAddress,
		Currency0:                   pairInfo.Currency0,
		Currency0AmountWithDecimals: swapEvent.Amount0,
		Currency1:                   pairInfo.Currency1,
		Currency1AmountWithDecimals: swapEvent.Amount1,
		NetworkFee:                  go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:                        txReceipt.TxHash.String(),
		BlockNumber:                 txReceipt.BlockNumber.Uint64(),
		Liquidity:                   swapEvent.Liquidity,
		Fee:                         swapEvent.Fee,
		ProtocolFee:                 swapEvent.ProtocolFee,
	}, nil
}

type AllowanceInfoType struct {
	Amount     *big.Int
	Expiration *big.Int
	Nonce      *big.Int
}

func (t *Router) Allowance(
	userAddress common.Address,
	tokenAddress common.Address,
	spender common.Address,
) (*AllowanceInfoType, error) {
	var r AllowanceInfoType
	err := t.wallet.CallContractConstant(
		&r,
		Permit2,
		Permit2_ABI,
		"allowance",
		nil,
		[]any{
			userAddress,
			tokenAddress,
			spender,
		},
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (t *Router) ApproveWait(
	ctx context.Context,
	priv string,
	tokenAddress common.Address,
	spender common.Address,
	amount *big.Int,
	expiration int64,
	opts *go_coin_eth.CallMethodOpts,
) (txReceipt_ *types.Receipt, err_ error) {
	approveAmount := amount
	if approveAmount == nil {
		approveAmount = go_coin_eth.MaxUint160
	}
	tx, err := t.wallet.BuildCallMethodTx(
		priv,
		Permit2,
		Permit2_ABI,
		"approve",
		opts,
		[]any{
			tokenAddress,
			spender,
			approveAmount,
			big.NewInt(expiration / 1000),
		},
	)
	if err != nil {
		return nil, err
	}
	txHash, err := t.wallet.SendRawTransaction(tx.TxHex)
	if err != nil {
		return nil, err
	}
	return t.wallet.WaitConfirm(ctx, txHash, time.Second)
}

package uniswap_universal_router

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v3 "github.com/pefish/go-coin-eth/uniswap-v3"
	go_decimal "github.com/pefish/go-decimal"
	"github.com/pkg/errors"
)

// https://github.com/Uniswap/universal-router/blob/main/contracts/modules/uniswap/v3/BytesLib.sol
type V3Path struct {
	Tokens []common.Address // token0, token1, token2...
	Fees   []uint32         // fee1, fee2... （len = len(Tokens)-1）
}

func DecodeV3Path(path []byte) (*V3Path, error) {
	const (
		addrLen = 20
		feeLen  = 3
	)

	// 最短长度：20 + 3 + 20 = 43
	if len(path) < addrLen*2+feeLen {
		return nil, fmt.Errorf("path too short: %d", len(path))
	}

	// 除第一个 token 外，每一段是 fee(3) + address(20)
	if (len(path)-addrLen)%(addrLen+feeLen) != 0 {
		return nil, fmt.Errorf("invalid path length: %d", len(path))
	}

	res := &V3Path{
		Tokens: make([]common.Address, 0),
		Fees:   make([]uint32, 0),
	}

	offset := 0

	// token0
	res.Tokens = append(res.Tokens, common.BytesToAddress(path[offset:offset+addrLen]))
	offset += addrLen

	// (fee + token) * N
	for offset < len(path) {
		// fee (uint24, big-endian)
		fee := uint32(path[offset])<<16 |
			uint32(path[offset+1])<<8 |
			uint32(path[offset+2])
		offset += feeLen

		res.Fees = append(res.Fees, fee)

		// token
		res.Tokens = append(res.Tokens, common.BytesToAddress(path[offset:offset+addrLen]))
		offset += addrLen
	}

	// sanity check
	if len(res.Tokens) != len(res.Fees)+1 {
		return nil, fmt.Errorf(
			"token/fee mismatch: tokens=%d fees=%d",
			len(res.Tokens),
			len(res.Fees),
		)
	}

	return res, nil
}

func EncodeV3Path(p *V3Path) ([]byte, error) {
	const (
		addrLen = 20
		feeLen  = 3
	)

	if p == nil {
		return nil, fmt.Errorf("nil DecodedV3Path")
	}

	if len(p.Tokens) < 2 {
		return nil, fmt.Errorf("path requires at least 2 tokens")
	}

	if len(p.Fees) != len(p.Tokens)-1 {
		return nil, fmt.Errorf(
			"invalid path: tokens=%d fees=%d",
			len(p.Tokens),
			len(p.Fees),
		)
	}

	// 预分配容量，避免多次扩容
	totalLen := addrLen + (addrLen+feeLen)*len(p.Fees)
	path := make([]byte, 0, totalLen)

	// token0
	path = append(path, p.Tokens[0].Bytes()...)

	// fee + token
	for i, fee := range p.Fees {
		if fee > 0xFFFFFF {
			return nil, fmt.Errorf("fee overflow (uint24): %d", fee)
		}

		// fee: uint24, big-endian
		path = append(path,
			byte(fee>>16),
			byte(fee>>8),
			byte(fee),
		)

		// token(i+1)
		path = append(path, p.Tokens[i+1].Bytes()...)
	}

	return path, nil
}

// 如果 baseToken 是 WBNB，会自动 Wrap 和 Unwrap
func (t *Router) SwapExactInputV3(
	ctx context.Context,
	priv string,
	poolKey *uniswap_v3.PoolKeyType,
	tokenIn common.Address,
	amountInWithDecimals *big.Int,
	slipage uint64, // 例如 slipage = 50，表示 0.5%
	gasLimit uint64,
	maxFeePerGas *big.Int,
	v3QuoterAddress common.Address,
) (*SwapResultType, error) {
	if slipage > 10000 {
		return nil, errors.New("slipage too high")
	}
	uniswapV3 := uniswap_v3.New(t.logger, t.wallet)
	if tokenIn != poolKey.Token0 &&
		tokenIn != poolKey.Token1 {
		return nil, errors.New("tokenIn is not in pool")
	}

	userAddress, err := t.wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}

	zeroForOne := false
	var tokenOut common.Address
	if tokenIn == poolKey.Token0 {
		tokenOut = poolKey.Token1
		zeroForOne = true
	} else {
		tokenOut = poolKey.Token0
	}

	path, err := EncodeV3Path(&V3Path{
		Tokens: []common.Address{
			tokenIn,
			tokenOut,
		},
		Fees: []uint32{
			uint32(poolKey.Fee),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	payerIsUser := true
	commands := make([]byte, 0)
	params := make([][]byte, 0)
	value := big.NewInt(0)
	nonce, err := t.wallet.NextNonce(userAddress)
	if err != nil {
		return nil, err
	}
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
			maxFeePerGas,
			nonce,
		)
		if err != nil {
			return nil, err
		}
		nonce = newNonce
	}

	var swapRecipient common.Address
	if tokenOut == go_coin_eth.WBNBAddress {
		swapRecipient = Universal_Router
	} else {
		swapRecipient = userAddress
	}

	amountOutMinimum := big.NewInt(0)
	if slipage > 0 {
		quoteResult, err := uniswapV3.QuoteExactInputSingle(
			v3QuoterAddress,
			poolKey,
			tokenIn,
			amountInWithDecimals,
		)
		if err != nil {
			return nil, err
		}
		amountOutMinimum = go_decimal.MustStart(quoteResult.AmountOut).
			MustMulti(
				(10000 - float64(slipage)) / 10000,
			).RoundDown(0).MustEndForBigInt()
	}
	commands = append(commands, 0x00) // Commands.V3_SWAP_EXACT_IN
	v3PayloadBytes, err := abi.Arguments{
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
			Type:    go_coin_eth.TypeBytes,
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
		path,
		payerIsUser,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params = append(params, v3PayloadBytes)

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
			GasLimit:     gasLimit,
			MaxFeePerGas: maxFeePerGas,
			Value:        value,
			Nonce:        nonce,
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
		"0x19b47279256b2a23a1665c810c8d55a1758940ee09377d4f8d26497a3577dc83", // swap event of PancakeV3Pool
		go_coin_eth.ZeroAddress, // any address
		txReceipt.Logs,
	)
	if err != nil {
		return nil, err
	}
	var swapEvent struct {
		Sender             common.Address
		Recipient          common.Address
		Amount0            *big.Int
		Amount1            *big.Int
		SqrtPriceX96       *big.Int
		Liquidity          *big.Int
		Tick               *big.Int
		ProtocolFeesToken0 *big.Int
		ProtocolFeesToken1 *big.Int
	}
	err = t.wallet.UnpackLog(
		&swapEvent,
		uniswap_v3.PoolABI,
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
				return swapEvent.Amount1.Neg(swapEvent.Amount1)
			} else {
				return swapEvent.Amount0.Neg(swapEvent.Amount0)
			}
		}(),
		NetworkFee:  go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:        txReceipt.TxHash.String(),
		BlockNumber: txReceipt.BlockNumber.Uint64(),
		Liquidity:   swapEvent.Liquidity,
		Fee:         swapEvent.ProtocolFeesToken0,
		ProtocolFee: swapEvent.ProtocolFeesToken1,
	}, nil
}

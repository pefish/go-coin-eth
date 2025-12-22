package uniswap_universal_router

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
)

type AllowanceInfoType struct {
	Amount     *big.Int
	Expiration *big.Int // s 级别时间戳
	Nonce      *big.Int
}

func (t *Router) AllowanceForPermit2(
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

func (t *Router) ApprovePermit2Wait(
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

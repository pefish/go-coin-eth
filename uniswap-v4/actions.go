package uniswap_v4

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// https://github.com/pancakeswap/infinity-periphery/blob/04585d106d55d5d3bdf1d79ec5c4275a01caa0a6/src/InfinityRouter.sol#L79

var ActionNameMap = map[byte]string{
	// liquidity actions
	0x00: "CL_INCREASE_LIQUIDITY",
	0x01: "CL_DECREASE_LIQUIDITY",
	0x02: "CL_MINT_POSITION",
	0x03: "CL_BURN_POSITION",
	0x04: "CL_INCREASE_LIQUIDITY_FROM_DELTAS",
	0x05: "CL_MINT_POSITION_FROM_DELTAS",

	// swapping
	0x06: "CL_SWAP_EXACT_IN_SINGLE",
	0x07: "CL_SWAP_EXACT_IN",
	0x08: "CL_SWAP_EXACT_OUT_SINGLE",
	0x09: "CL_SWAP_EXACT_OUT",

	0x0a: "CL_DONATE",

	0x0b: "SETTLE",
	0x0c: "SETTLE_ALL",
	0x0d: "SETTLE_PAIR",

	0x0e: "TAKE",
	0x0f: "TAKE_ALL",
	0x10: "TAKE_PORTION",
	0x11: "TAKE_PAIR",

	0x12: "CLOSE_CURRENCY",
	0x13: "CLEAR_OR_TAKE",
	0x14: "SWEEP",
	0x15: "WRAP",
	0x16: "UNWRAP",

	0x17: "MINT_6909",
	0x18: "BURN_6909",

	0x19: "BIN_ADD_LIQUIDITY",
	0x1a: "BIN_REMOVE_LIQUIDITY",
	0x1b: "BIN_ADD_LIQUIDITY_FROM_DELTAS",

	0x1c: "BIN_SWAP_EXACT_IN_SINGLE",
	0x1d: "BIN_SWAP_EXACT_IN",
	0x1e: "BIN_SWAP_EXACT_OUT_SINGLE",
	0x1f: "BIN_SWAP_EXACT_OUT",

	0x20: "BIN_DONATE",
}

type ActionData struct {
	Action string
	Params any
}

type CLSwapExactInSingleParamsType struct {
	PoolKey          PoolKeyType `json:"pool_key"`
	ZeroForOne       bool        `json:"zero_for_one"`
	AmountIn         *big.Int    `json:"amount_in"`
	AmountOutMinimum *big.Int    `json:"amount_out_minimum"`
	HookData         []byte      `json:"hook_data"`
}

type SettleAllParamsType struct {
	Currency  common.Address `json:"currency"`
	MaxAmount *big.Int       `json:"max_amount"`
}

type TakeAllParamsType struct {
	Currency  common.Address `json:"currency"`
	MinAmount *big.Int       `json:"min_amount"`
}

type SettleParamsType struct {
	Currency    common.Address `json:"currency"`
	Amount      *big.Int       `json:"amount"`
	PayerIsUser bool           `json:"payer_is_user"`
}

type TakeParamsType struct {
	Currency  common.Address `json:"currency"`
	Recipient common.Address `json:"recipient"`
	Amount    *big.Int       `json:"amount"`
}

var CLSwapExactInSingleParamsAbiType, _ = abi.NewType("tuple", "CLSwapExactInSingleParamsType", []abi.ArgumentMarshaling{
	{Type: "tuple", Name: "pool_key", Components: []abi.ArgumentMarshaling{
		{Type: "address", Name: "currency0"},
		{Type: "address", Name: "currency1"},
		{Type: "address", Name: "hooks"},
		{Type: "address", Name: "pool_manager"},
		{Type: "uint24", Name: "fee"},
		{Type: "bytes32", Name: "parameters"},
	}},
	{Type: "bool", Name: "zero_for_one"},
	{Type: "uint128", Name: "amount_in"},
	{Type: "uint128", Name: "amount_out_minimum"},
	{Type: "bytes", Name: "hook_data"},
})

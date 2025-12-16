package uniswap_universal_router

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v4_trade "github.com/pefish/go-coin-eth/uniswap-v4-trade"
	go_format_any "github.com/pefish/go-format/any"
)

var CommandNameMap = map[byte]string{
	0x00: "V3_SWAP_EXACT_IN",
	0x01: "V3_SWAP_EXACT_OUT",
	0x02: "PERMIT2_TRANSFER_FROM",
	0x03: "PERMIT2_PERMIT_BATCH",
	0x04: "SWEEP",
	0x05: "TRANSFER",
	0x06: "PAY_PORTION",

	0x08: "V2_SWAP_EXACT_IN",
	0x09: "V2_SWAP_EXACT_OUT",
	0x0a: "PERMIT2_PERMIT",
	0x0b: "WRAP_ETH",
	0x0c: "UNWRAP_WETH",
	0x0d: "PERMIT2_TRANSFER_FROM_BATCH",
	0x0e: "BALANCE_CHECK_ERC20",

	0x10: "INFI_SWAP",
	0x11: "V3_POSITION_MANAGER_PERMIT",
	0x12: "V3_POSITION_MANAGER_CALL",
	0x13: "INFI_CL_INITIALIZE_POOL",
	0x14: "INFI_BIN_INITIALIZE_POOL",
	0x15: "INFI_CL_POSITION_CALL",
	0x16: "INFI_BIN_POSITION_CALL",

	0x21: "EXECUTE_SUB_PLAN",
	0x22: "STABLE_SWAP_EXACT_IN",
	0x23: "STABLE_SWAP_EXACT_OUT",
}

type CommandData struct {
	Command string
	Params  any
}

func (t *Router) DecodeCommands(txPayloadStr string) ([]*CommandData, error) {
	var r struct {
		Commands []byte
		Inputs   [][]byte
		Deadline *big.Int
	}
	_, err := t.wallet.DecodePayload(
		Universal_Router_ABI,
		&r,
		txPayloadStr,
	)
	if err != nil {
		return nil, err
	}
	var commandDatas []*CommandData
	for i := 0; i < len(r.Commands); i++ {
		input := r.Inputs[i]
		var commandParams any
		switch CommandNameMap[r.Commands[i]] {
		case "INFI_SWAP":
			var actionDatas []*uniswap_v4_trade.ActionData
			infiSwapParamsAny, err := t.wallet.UnpackParams(
				[]abi.Type{
					go_coin_eth.TypeBytes,
					go_coin_eth.TypeBytesSlice,
				},
				input,
			)
			if err != nil {
				return nil, err
			}
			actions := infiSwapParamsAny[0].([]byte)
			paramsSlice := infiSwapParamsAny[1].([][]byte)
			for i, action := range actions {
				actionName := uniswap_v4_trade.ActionNameMap[action]
				paramsBytes := paramsSlice[i]
				var actionParams any
				switch actionName {
				case "CL_SWAP_EXACT_IN_SINGLE":
					swapParamsAny, err := t.wallet.UnpackParams(
						[]abi.Type{
							uniswap_v4_trade.CLSwapExactInSingleParamsAbiType,
						},
						paramsBytes,
					)
					if err != nil {
						return nil, err
					}
					var dest uniswap_v4_trade.CLSwapExactInSingleParamsType
					err = go_format_any.ToStruct(swapParamsAny[0], &dest)
					if err != nil {
						return nil, err
					}
					actionParams = &dest
				case "SETTLE_ALL":
					settleParamsAny, err := t.wallet.UnpackParams(
						[]abi.Type{
							go_coin_eth.TypeAddress,
							go_coin_eth.TypeUint256,
						},
						paramsBytes,
					)
					if err != nil {
						return nil, err
					}
					actionParams = &uniswap_v4_trade.CLSettleAllParamsType{
						Address: settleParamsAny[0].(common.Address),
						Amount:  settleParamsAny[1].(*big.Int),
					}
				case "TAKE_ALL":
					takeParamsAny, err := t.wallet.UnpackParams(
						[]abi.Type{
							go_coin_eth.TypeAddress,
							go_coin_eth.TypeUint256,
						},
						paramsBytes,
					)
					if err != nil {
						return nil, err
					}
					actionParams = &uniswap_v4_trade.CLTakeAllParamsType{
						Address: takeParamsAny[0].(common.Address),
						Amount:  takeParamsAny[1].(*big.Int),
					}
				}
				actionDatas = append(actionDatas, &uniswap_v4_trade.ActionData{
					Action: actionName,
					Params: actionParams,
				})
			}
			commandParams = actionDatas
		}

		commandDatas = append(commandDatas, &CommandData{
			Command: CommandNameMap[r.Commands[i]],
			Params:  commandParams,
		})
	}
	return commandDatas, nil
}

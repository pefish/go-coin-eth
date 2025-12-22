package uniswap_universal_router

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v4 "github.com/pefish/go-coin-eth/uniswap-v4"
	go_format_any "github.com/pefish/go-format/any"
	"github.com/pkg/errors"
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

type WrapETHParamsType struct {
	Recipient common.Address
	Amount    *big.Int
}

type UnwrapETHParamsType struct {
	Recipient common.Address
	AmountMin *big.Int
}

type V3SwapExactInParamsType struct {
	Recipient    common.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         *V3Path
	PayerIsUser  bool
}

type V2SwapExactInParamsType struct {
	Recipient    common.Address
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []common.Address
	PayerIsUser  bool
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
		case "WRAP_ETH":
			paramsAny, err := t.wallet.UnpackParams(
				[]abi.Type{
					go_coin_eth.TypeAddress,
					go_coin_eth.TypeUint256,
				},
				input,
			)
			if err != nil {
				return nil, err
			}
			commandParams = &WrapETHParamsType{
				Recipient: paramsAny[0].(common.Address),
				Amount:    paramsAny[1].(*big.Int),
			}
		case "UNWRAP_WETH":
			paramsAny, err := t.wallet.UnpackParams(
				[]abi.Type{
					go_coin_eth.TypeAddress,
					go_coin_eth.TypeUint256,
				},
				input,
			)
			if err != nil {
				return nil, err
			}
			commandParams = &UnwrapETHParamsType{
				Recipient: paramsAny[0].(common.Address),
				AmountMin: paramsAny[1].(*big.Int),
			}
		case "V3_SWAP_EXACT_IN":
			paramsAny, err := abi.Arguments{
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
			}.Unpack(input)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			pathBytes := paramsAny[3].([]byte)
			decodedPath, err := DecodeV3Path(pathBytes)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			commandParams = &V3SwapExactInParamsType{
				Recipient:    paramsAny[0].(common.Address),
				AmountIn:     paramsAny[1].(*big.Int),
				AmountOutMin: paramsAny[2].(*big.Int),
				Path:         decodedPath,
				PayerIsUser:  paramsAny[4].(bool),
			}
		case "V2_SWAP_EXACT_IN":
			paramsAny, err := abi.Arguments{
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
			}.Unpack(input)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			commandParams = &V2SwapExactInParamsType{
				Recipient:    paramsAny[0].(common.Address),
				AmountIn:     paramsAny[1].(*big.Int),
				AmountOutMin: paramsAny[2].(*big.Int),
				Path:         paramsAny[3].([]common.Address),
				PayerIsUser:  paramsAny[4].(bool),
			}

		case "INFI_SWAP":
			var actionDatas []*uniswap_v4.ActionData
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
				actionName := uniswap_v4.ActionNameMap[action]
				paramsBytes := paramsSlice[i]
				var actionParams any
				switch actionName {
				case "CL_SWAP_EXACT_IN_SINGLE":
					swapParamsAny, err := t.wallet.UnpackParams(
						[]abi.Type{
							uniswap_v4.CLSwapExactInSingleParamsAbiType,
						},
						paramsBytes,
					)
					if err != nil {
						return nil, err
					}
					var dest uniswap_v4.CLSwapExactInSingleParamsType
					err = go_format_any.ToStruct(swapParamsAny[0], &dest)
					if err != nil {
						return nil, err
					}
					actionParams = &dest
				case "SETTLE":
					settleParamsAny, err := t.wallet.UnpackParams(
						[]abi.Type{
							go_coin_eth.TypeAddress,
							go_coin_eth.TypeUint256,
							go_coin_eth.TypeBool,
						},
						paramsBytes,
					)
					if err != nil {
						return nil, err
					}
					actionParams = &uniswap_v4.SettleParamsType{
						Currency:    settleParamsAny[0].(common.Address),
						Amount:      settleParamsAny[1].(*big.Int),
						PayerIsUser: settleParamsAny[2].(bool),
					}
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
					actionParams = &uniswap_v4.SettleAllParamsType{
						Currency:  settleParamsAny[0].(common.Address),
						MaxAmount: settleParamsAny[1].(*big.Int),
					}
				case "TAKE":
					takeParamsAny, err := t.wallet.UnpackParams(
						[]abi.Type{
							go_coin_eth.TypeAddress,
							go_coin_eth.TypeAddress,
							go_coin_eth.TypeUint256,
						},
						paramsBytes,
					)
					if err != nil {
						return nil, err
					}
					actionParams = &uniswap_v4.TakeParamsType{
						Currency:  takeParamsAny[0].(common.Address),
						Recipient: takeParamsAny[1].(common.Address),
						Amount:    takeParamsAny[2].(*big.Int),
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
					actionParams = &uniswap_v4.TakeAllParamsType{
						Currency:  takeParamsAny[0].(common.Address),
						MinAmount: takeParamsAny[1].(*big.Int),
					}
				}
				actionDatas = append(actionDatas, &uniswap_v4.ActionData{
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

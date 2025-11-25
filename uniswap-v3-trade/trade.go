package uniswap_v3_trade

// v3 中，pool 由 token1、token2、fee 三个参数唯一确定

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	constants "github.com/pefish/go-coin-eth/uniswap-v3-trade/constant"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

type Trader struct {
	wallet *go_coin_eth.Wallet
	logger i_logger.ILogger
}

func New(
	logger i_logger.ILogger,
	wallet *go_coin_eth.Wallet,
) *Trader {
	return &Trader{
		wallet: wallet,
		logger: logger,
	}
}

func (t *Trader) WETHAddressFromRouter(routerAddress string) (string, error) {
	var wethAddress common.Address
	err := t.wallet.CallContractConstant(
		&wethAddress,
		routerAddress,
		constants.RouterABIStr,
		"WETH9",
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}
	return wethAddress.String(), nil
}

type BuyByExactETHResult struct {
	TokenAmount string
	Fee         string
	TxId        string
}

type BuyByExactETHOpts struct {
	TokenDecimals                     uint64
	WETHAddress                       string
	MinReceiveTokenAmountWithDecimals *big.Int
	GasLimit                          uint64
	MaxFeePerGas                      *big.Int
}

// 只能 route v3 pool，会查找 bnb、token、fee 三个参数唯一确定的 pool 去进行交易，没有这个池子就会失败
func (t *Trader) BuyByExactETH(
	ctx context.Context,
	priv string,
	ethAmount string,
	routerAddress string,
	tokenAddress string,
	fee uint64,
	opts *BuyByExactETHOpts,
) (*BuyByExactETHResult, error) {

	var result BuyByExactETHResult

	var realOpts BuyByExactETHOpts
	if opts != nil {
		realOpts = *opts
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 300000
	}

	if realOpts.MinReceiveTokenAmountWithDecimals == nil {
		realOpts.MinReceiveTokenAmountWithDecimals = big.NewInt(0)
	}

	selfAddress, err := t.wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}
	ethAmountWithDecimals := go_decimal.MustStart(ethAmount).MustShiftedBy(18).MustEndForBigInt()

	balanceWithDecimals, err := t.wallet.Balance(selfAddress)
	if err != nil {
		return nil, err
	}

	if balanceWithDecimals.Add(balanceWithDecimals, big.NewInt(10000000000000000)).Cmp(ethAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s + 10000000000000000", balanceWithDecimals, ethAmountWithDecimals)
	}

	if realOpts.WETHAddress == "" {
		wethAddress_, err := t.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	if realOpts.TokenDecimals == 0 {
		decimals_, err := t.wallet.TokenDecimals(tokenAddress)
		if err != nil {
			return nil, err
		}
		realOpts.TokenDecimals = decimals_
	}

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		constants.RouterABIStr,
		"exactInputSingle",
		&go_coin_eth.CallMethodOpts{
			Value:        ethAmountWithDecimals,
			GasLimit:     realOpts.GasLimit,
			MaxFeePerGas: realOpts.MaxFeePerGas,
		},
		[]any{
			struct {
				TokenIn           common.Address
				TokenOut          common.Address
				Fee               *big.Int
				Recipient         common.Address
				AmountIn          *big.Int
				AmountOutMinimum  *big.Int
				SqrtPriceLimitX96 *big.Int
			}{
				TokenIn:           common.HexToAddress(realOpts.WETHAddress),
				TokenOut:          common.HexToAddress(tokenAddress),
				Fee:               big.NewInt(int64(fee)),
				Recipient:         common.HexToAddress(selfAddress),
				AmountIn:          ethAmountWithDecimals,
				AmountOutMinimum:  realOpts.MinReceiveTokenAmountWithDecimals,
				SqrtPriceLimitX96: big.NewInt(0),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	result.TxId = btr.SignedTx.Hash().String()
	t.logger.InfoF("购买 txid <%s> 等待确认", result.TxId)
	tr, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return &result, err
	}
	t.logger.InfoF("<%s> 已确认", result.TxId)
	tokenAmountWithDecimals, err := t.receivedTokenAmountInLogs(
		tr.Logs,
		tokenAddress,
		selfAddress,
	)
	if err != nil {
		return &result, err
	}

	result.TokenAmount = go_decimal.MustStart(tokenAmountWithDecimals).MustUnShiftedBy(realOpts.TokenDecimals).EndForString()
	result.Fee = go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
	return &result, nil
}

func (t *Trader) receivedTokenAmountInLogs(
	logs []*types.Log,
	tokenAddress string,
	myAddress string,
) (*big.Int, error) {
	result := big.NewInt(0)

	transferLogs, err := t.wallet.FilterLogs(
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", // Transfer of ERC20
		tokenAddress,
		logs,
	)
	if err != nil {
		return nil, err
	}
	for _, log := range transferLogs {
		var transferEvent struct {
			From  common.Address
			Value *big.Int
			To    common.Address
		}
		err := t.wallet.UnpackLog(
			&transferEvent,
			go_coin_eth.Erc20AbiStr,
			"Transfer",
			log,
		)
		if err != nil {
			return nil, err
		}
		if transferEvent.To.Cmp(common.HexToAddress(myAddress)) != 0 {
			continue
		}
		result.Add(result, transferEvent.Value)
	}

	return result, nil
}

type SellByExactTokenOpts struct {
	TokenDecimals                   uint64
	WETHAddress                     string
	MinReceiveETHAmountWithDecimals *big.Int
	GasLimit                        uint64
	MaxFeePerGas                    *big.Int
}

type SellByExactTokenResult struct {
	ETHAmount   string
	ApproveFee  string
	SellFee     string
	Fee         string
	SellTxId    string
	ApproveTxId string
}

// 只会得到 WBNB，不会自动转换为 BNB
func (t *Trader) SellByExactToken(
	ctx context.Context,
	priv string,
	tokenAmount string,
	routerAddress string,
	tokenAddress string,
	fee uint64,
	opts *SellByExactTokenOpts,
) (*SellByExactTokenResult, error) {

	var result SellByExactTokenResult

	var realOpts SellByExactTokenOpts
	if opts != nil {
		realOpts = *opts
	}

	if realOpts.GasLimit == 0 {
		realOpts.GasLimit = 300000
	}

	if realOpts.MinReceiveETHAmountWithDecimals == nil {
		realOpts.MinReceiveETHAmountWithDecimals = big.NewInt(0)
	}

	selfAddress, err := t.wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return nil, err
	}
	if realOpts.TokenDecimals == 0 {
		decimals_, err := t.wallet.TokenDecimals(tokenAddress)
		if err != nil {
			return nil, err
		}
		realOpts.TokenDecimals = decimals_
	}
	tokenAmountWithDecimals := go_decimal.MustStart(tokenAmount).MustShiftedBy(realOpts.TokenDecimals).MustEndForBigInt()

	tokenBalanceWithDecimals, err := t.wallet.TokenBalance(tokenAddress, selfAddress)
	if err != nil {
		return nil, err
	}
	if tokenBalanceWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		return nil, errors.Errorf("余额不足，%s < %s", tokenBalanceWithDecimals, tokenAmountWithDecimals)
	}

	if realOpts.WETHAddress == "" {
		wethAddress_, err := t.WETHAddressFromRouter(routerAddress)
		if err != nil {
			return nil, err
		}
		realOpts.WETHAddress = wethAddress_
	}

	approvedAmountWithDecimals, err := t.wallet.ApprovedAmount(
		tokenAddress,
		selfAddress,
		routerAddress,
	)
	if err != nil {
		return nil, err
	}

	result.ApproveFee = "0"
	if approvedAmountWithDecimals.Cmp(tokenAmountWithDecimals) < 0 {
		tr, err := t.wallet.ApproveWait(
			ctx,
			priv,
			tokenAddress,
			routerAddress,
			go_coin_eth.MaxUint256,
			&go_coin_eth.CallMethodOpts{
				MaxFeePerGas: realOpts.MaxFeePerGas,
			},
		)
		if err != nil {
			return nil, err
		}
		result.ApproveTxId = tr.TxHash.String()
		result.ApproveFee = go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
		t.logger.InfoF("Approve 成功。txid <%s>", tr.TxHash.String())
	}

	btr, err := t.wallet.BuildCallMethodTx(
		priv,
		routerAddress,
		constants.RouterABIStr,
		"exactInputSingle",
		&go_coin_eth.CallMethodOpts{
			GasLimit:     realOpts.GasLimit,
			MaxFeePerGas: realOpts.MaxFeePerGas,
		},
		[]any{
			struct {
				TokenIn           common.Address
				TokenOut          common.Address
				Fee               *big.Int
				Recipient         common.Address
				AmountIn          *big.Int
				AmountOutMinimum  *big.Int
				SqrtPriceLimitX96 *big.Int
			}{
				TokenIn:           common.HexToAddress(tokenAddress),
				TokenOut:          common.HexToAddress(realOpts.WETHAddress),
				Fee:               big.NewInt(int64(fee)),
				Recipient:         common.HexToAddress(selfAddress),
				AmountIn:          tokenAmountWithDecimals,
				AmountOutMinimum:  realOpts.MinReceiveETHAmountWithDecimals,
				SqrtPriceLimitX96: big.NewInt(0),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	result.SellTxId = btr.SignedTx.Hash().String()
	t.logger.InfoF("购买 txid <%s> 等待确认", result.SellTxId)
	tr, err := t.wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return &result, err
	}
	t.logger.InfoF("<%s> 已确认", result.SellTxId)
	ethAmountWithDecimals, err := t.receivedTokenAmountInLogs(tr.Logs, opts.WETHAddress, selfAddress)
	if err != nil {
		return &result, err
	}

	result.ETHAmount = go_decimal.MustStart(ethAmountWithDecimals).MustUnShiftedBy(18).EndForString()
	result.SellFee = go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).MustUnShiftedBy(18).EndForString()
	result.Fee = go_decimal.MustStart(result.ApproveFee).MustAddForString(result.SellFee)
	return &result, nil
}

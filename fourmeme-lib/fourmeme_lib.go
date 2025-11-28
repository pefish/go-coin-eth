package fourmeme_lib

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	"github.com/pefish/go-coin-eth/fourmeme-lib/constant"
	go_decimal "github.com/pefish/go-decimal"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

type TradeEventType struct {
	Token   common.Address `json:"token"`
	Account common.Address `json:"account"`
	Price   *big.Int       `json:"price"`
	Amount  *big.Int       `json:"amount"`
	Cost    *big.Int       `json:"cost"`
	Fee     *big.Int       `json:"fee"`
	Offers  *big.Int       `json:"offers"`
	Funds   *big.Int       `json:"funds"`
}

type TradeResultType struct {
	TradeEventType

	NetworkFee  *big.Int `json:"network_fee"`
	TxId        string   `json:"tx_id"`
	BlockNumber uint64   `json:"block_number"`
}

func Buy(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress string,
	tokenAmount *big.Int,
	maxCostBnbAmount *big.Int,
	maxFeePerGas *big.Int, // 1 大概 0.1 刀
) (*TradeResultType, error) {
	btr, err := wallet.BuildCallMethodTx(
		priv,
		constant.FourmemeToolAddress,
		constant.FourmemeToolABI,
		"buyPefish",
		&go_coin_eth.CallMethodOpts{
			MaxFeePerGas:   maxFeePerGas,
			GasLimit:       250000,
			IsPredictError: false,
		},
		[]any{
			common.HexToAddress(tokenAddress),
			tokenAmount,
			maxCostBnbAmount,
		},
	)
	if err != nil {
		return nil, err
	}
	tr, err := wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}

	logs, err := wallet.FilterLogs(
		"0x7db52723a3b2cdd6164364b3b766e65e540d7be48ffa89582956d8eaebe62942",
		constant.TokenManagerAddress,
		tr.Logs,
	)
	if err != nil {
		return nil, err
	}
	var r TradeEventType
	err = wallet.UnpackLog(&r, constant.TokenManagerABI, "TokenSale", logs[0])
	if err != nil {
		return nil, err
	}

	return &TradeResultType{
		TradeEventType: r,
		NetworkFee:     go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:           tr.TxHash.String(),
		BlockNumber:    tr.BlockNumber.Uint64(),
	}, nil
}

func Sell(
	ctx context.Context,
	wallet *go_coin_eth.Wallet,
	priv string,
	tokenAddress string,
	tokenAmount *big.Int,
	minReceiveBnbAmount *big.Int,
	maxFeePerGas *big.Int,
) (*TradeResultType, error) {
	btr, err := wallet.BuildCallMethodTx(
		priv,
		constant.FourmemeToolAddress,
		constant.FourmemeToolABI,
		"sellPefish",
		&go_coin_eth.CallMethodOpts{
			MaxFeePerGas:   maxFeePerGas,
			GasLimit:       250000,
			IsPredictError: false,
		},
		[]any{
			common.HexToAddress(tokenAddress),
			tokenAmount,
			minReceiveBnbAmount,
		},
	)
	if err != nil {
		return nil, err
	}
	tr, err := wallet.SendRawTransactionWait(ctx, btr.TxHex)
	if err != nil {
		return nil, err
	}
	logs, err := wallet.FilterLogs(
		"0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19",
		constant.TokenManagerAddress,
		tr.Logs,
	)
	if err != nil {
		return nil, err
	}
	var r TradeEventType
	err = wallet.UnpackLog(&r, constant.TokenManagerABI, "TokenSale", logs[0])
	if err != nil {
		return nil, err
	}

	return &TradeResultType{
		TradeEventType: r,
		NetworkFee:     go_decimal.MustStart(tr.EffectiveGasPrice).MustMulti(tr.GasUsed).RoundDown(0).MustEndForBigInt(),
		TxId:           tr.TxHash.String(),
		BlockNumber:    tr.BlockNumber.Uint64(),
	}, nil
}

type TokenInfoType struct {
	Version        *big.Int       `json:"version"`
	TokenManager   common.Address `json:"tokenManager"`
	Quote          common.Address `json:"quote"`
	LastPrice      *big.Int       `json:"lastPrice"`
	TradingFeeRate *big.Int       `json:"tradingFeeRate"`
	MinTradingFee  *big.Int       `json:"minTradingFee"`
	LaunchTime     *big.Int       `json:"launchTime"`
	Offers         *big.Int       `json:"offers"`
	MaxOffers      *big.Int       `json:"maxOffers"`
	Funds          *big.Int       `json:"funds"`
	MaxFunds       *big.Int       `json:"maxFunds"`
	LiquidityAdded bool           `json:"liquidityAdded"`
	PairAddress    common.Address `json:"pairAddress"` // 上岸后 pancake 中的 pair address
}

// tokenInfo 返回 nil 表示 token 不是 fourmeme token
func TokenInfo(wallet *go_coin_eth.Wallet, tokenAddress string) (*TokenInfoType, error) {
	var callResult TokenInfoType
	err := wallet.CallContractConstant(
		&callResult,
		constant.TokenManagerHelperAddress,
		constant.TokenManagerHelperABI,
		"getTokenInfo",
		nil,
		[]any{
			common.HexToAddress(tokenAddress),
		},
	)
	if err != nil {
		return nil, err
	}
	if callResult.LaunchTime.Int64() == 0 {
		return nil, nil
	}
	if callResult.LiquidityAdded {
		var pairAddress common.Address
		err = wallet.CallContractConstant(
			&pairAddress,
			constant.PancakeFactoryAddress,
			constant.PancakeFactoryABI,
			"getPair",
			nil,
			[]any{
				common.HexToAddress(tokenAddress),
				common.HexToAddress(constant.WBNBAddress),
			},
		)
		if err != nil {
			return nil, err
		}
		callResult.PairAddress = pairAddress
	}
	return &callResult, nil
}

type TokenInfoByAPIType struct {
	Id          int64  `json:"id"`
	Address     string `json:"address"`
	Image       string `json:"image"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Symbol      string `json:"symbol"`
	Descr       string `json:"descr"`
	TwitterUrl  string `json:"twitterUrl"`
	TotalAmount string `json:"totalAmount"`
	SaleAmount  string `json:"saleAmount"`
	B0          string `json:"b0"`
	T0          string `json:"t0"`
	LaunchTime  int64  `json:"launchTime"`
	MinBuy      string `json:"minBuy"`
	MaxBuy      string `json:"maxBuy"`
	UserId      int64  `json:"userId"`
	UserAddress string `json:"userAddress"`
	UserName    string `json:"userName"`
	UserAvatar  string `json:"userAvatar"`
	Status      string `json:"status"`
	ShowStatus  string `json:"showStatus"`
	TokenPrice  struct {
		Price        string `json:"price"`
		MaxPrice     string `json:"maxPrice"`
		Increase     string `json:"increase"`
		Amount       string `json:"amount"`
		MarketCap    string `json:"marketCap"`
		Trading      string `json:"trading"`
		DayIncrease  string `json:"dayIncrease"`
		DayTrading   string `json:"dayTrading"`
		RaisedAmount string `json:"raisedAmount"`
		Progress     string `json:"progress"`
		Liquidity    string `json:"liquidity"`
		TradingUsd   string `json:"tradingUsd"`
		CreateDate   uint64 `json:"createDate,string"`
		ModifyDate   uint64 `json:"modifyDate,string"`
		Bamount      string `json:"bamount"`
		Tamount      string `json:"tamount"`
	} `json:"tokenPrice"`
	OscarStatus   string `json:"oscarStatus"`
	ProgressTag   bool   `json:"progressTag"`
	CtoTag        bool   `json:"ctoTag"`
	Version       string `json:"version"`
	ClickFunCheck bool   `json:"clickFunCheck"`
	ReserveAmount string `json:"reserveAmount"`
	RaisedAmount  string `json:"raisedAmount"`
	NetworkCode   string `json:"networkCode"`
	Label         string `json:"label"`
	CreateDate    uint64 `json:"createDate,string"`
	ModifyDate    uint64 `json:"modifyDate,string"`
	IsRush        bool   `json:"isRush"`
	DexType       string `json:"dexType"`
	LastId        int64  `json:"lastId"`
}

func TokenInfoByAPI(logger i_logger.ILogger, tokenAddress string) (*TokenInfoByAPIType, error) {
	var callResult struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data TokenInfoByAPIType `json:"data"`
	}
	_, _, err := go_http.NewHttpRequester(go_http.WithLogger(logger), go_http.WithTimeout(10*time.Second)).GetForStruct(
		&go_http.RequestParams{
			Url: "https://four.meme/meme-api/v1/private/token/get/v2",
			Queries: map[string]string{
				"address": tokenAddress,
			},
		},
		&callResult,
	)
	if err != nil {
		return nil, err
	}
	if callResult.Code != 0 {
		return nil, errors.New(callResult.Msg)
	}
	return &callResult.Data, nil
}

type ReserveInfoType struct {
	ReserveBNBWithDecimals   string
	ReserveTokenWithDecimals string
	Price                    string
}

func GetReserveInfo(
	wallet *go_coin_eth.Wallet,
	tokenAddress string,
) (*ReserveInfoType, error) {
	tokenInfo, err := TokenInfo(wallet, tokenAddress)
	if err != nil {
		return nil, err
	}
	if tokenInfo.Quote.String() != "0x0000000000000000000000000000000000000000" {
		return nil, errors.New("quote not WBNB")
	}
	if tokenInfo.LiquidityAdded {
		reserveBNBWithDecimals, err := wallet.TokenBalance(constant.WBNBAddress, tokenInfo.PairAddress.String())
		if err != nil {
			return nil, err
		}
		reserveTokenWithDecimals, err := wallet.TokenBalance(tokenAddress, tokenInfo.PairAddress.String())
		if err != nil {
			return nil, err
		}
		return &ReserveInfoType{
			ReserveBNBWithDecimals:   reserveBNBWithDecimals.String(),
			ReserveTokenWithDecimals: reserveTokenWithDecimals.String(),
			Price:                    go_decimal.MustStart(reserveBNBWithDecimals).MustDivForString(reserveTokenWithDecimals),
		}, nil
	}

	return &ReserveInfoType{
		ReserveBNBWithDecimals:   tokenInfo.Funds.String(),
		ReserveTokenWithDecimals: tokenInfo.Offers.String(),
		Price:                    go_decimal.MustStart(tokenInfo.LastPrice).MustUnShiftedBy(18).EndForString(),
	}, nil
}

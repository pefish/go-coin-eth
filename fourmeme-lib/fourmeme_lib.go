package fourmeme_lib

import (
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_coin_eth "github.com/pefish/go-coin-eth"
	"github.com/pefish/go-coin-eth/fourmeme-lib/constant"
	go_decimal "github.com/pefish/go-decimal"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

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

type SwapInfoType struct {
	TradeEventType

	Type_       string   `json:"type"` // buy/sell
	NetworkFee  *big.Int `json:"network_fee"`
	TxId        string   `json:"tx_id"`
	BlockNumber uint64   `json:"block_number"`
}

func ParseSwapInfos(
	wallet *go_coin_eth.Wallet,
	txReceipt *types.Receipt,
) ([]*SwapInfoType, error) {
	results := make([]*SwapInfoType, 0)
	parsedAbi, err := abi.JSON(strings.NewReader(constant.TokenManagerABI))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	boundContract := bind.NewBoundContract(
		common.HexToAddress(""),
		parsedAbi,
		wallet.RemoteRpcClient,
		wallet.RemoteRpcClient,
		wallet.RemoteRpcClient,
	)
	for _, log := range txReceipt.Logs {
		var r TradeEventType
		if log.Topics[0] == parsedAbi.Events["TokenPurchase"].ID {
			err = boundContract.UnpackLog(&r, "TokenPurchase", *log)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			results = append(results, &SwapInfoType{
				TradeEventType: r,
				Type_:          "buy",
				NetworkFee:     go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
				TxId:           txReceipt.TxHash.String(),
				BlockNumber:    txReceipt.BlockNumber.Uint64(),
			})
		} else if log.Topics[0] == parsedAbi.Events["TokenSale"].ID {
			err = boundContract.UnpackLog(&r, "TokenSale", *log)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			results = append(results, &SwapInfoType{
				TradeEventType: r,
				Type_:          "sell",
				NetworkFee:     go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
				TxId:           txReceipt.TxHash.String(),
				BlockNumber:    txReceipt.BlockNumber.Uint64(),
			})
		} else {
			continue
		}

	}

	return results, nil
}

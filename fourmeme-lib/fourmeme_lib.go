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
	type_ "github.com/pefish/go-coin-eth/type"
	uniswap_v2_trade_constant "github.com/pefish/go-coin-eth/uniswap-v2-trade/constant"
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
func TokenInfo(wallet *go_coin_eth.Wallet, tokenAddress common.Address) (*TokenInfoType, error) {
	var callResult TokenInfoType
	err := wallet.CallContractConstant(
		&callResult,
		constant.TokenManagerHelperAddress,
		constant.TokenManagerHelperABI,
		"getTokenInfo",
		nil,
		[]any{
			tokenAddress,
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
			uniswap_v2_trade_constant.PancakeV2FactoryAddress,
			uniswap_v2_trade_constant.PancakeV2FactoryABI,
			"getPair",
			nil,
			[]any{
				tokenAddress,
				constant.WBNBAddress,
			},
		)
		if err != nil {
			return nil, err
		}
		if pairAddress.Cmp(go_coin_eth.ZeroAddress) == 0 {
			return nil, errors.New("liquidity added but pair address is zero")
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

func TokenInfoByAPI(logger i_logger.ILogger, tokenAddress common.Address) (*TokenInfoByAPIType, error) {
	var callResult struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data TokenInfoByAPIType `json:"data"`
	}
	_, _, err := go_http.NewHttpRequester(go_http.WithLogger(logger), go_http.WithTimeout(10*time.Second)).GetForStruct(
		&go_http.RequestParams{
			Url: "https://four.meme/meme-api/v1/private/token/get/v2",
			Queries: map[string]string{
				"address": tokenAddress.String(),
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
	ReserveBNBWithDecimals   *big.Int
	ReserveTokenWithDecimals *big.Int
	Price                    string
}

func GetReserveInfo(
	wallet *go_coin_eth.Wallet,
	tokenAddress common.Address,
) (*ReserveInfoType, error) {
	tokenInfo, err := TokenInfo(wallet, tokenAddress)
	if err != nil {
		return nil, err
	}
	if tokenInfo.Quote.String() != "0x0000000000000000000000000000000000000000" {
		return nil, errors.New("quote not WBNB")
	}
	if tokenInfo.LiquidityAdded {
		reserveBNBWithDecimals, err := wallet.TokenBalance(constant.WBNBAddress, tokenInfo.PairAddress)
		if err != nil {
			return nil, err
		}
		reserveTokenWithDecimals, err := wallet.TokenBalance(tokenAddress, tokenInfo.PairAddress)
		if err != nil {
			return nil, err
		}
		return &ReserveInfoType{
			ReserveBNBWithDecimals:   reserveBNBWithDecimals,
			ReserveTokenWithDecimals: reserveTokenWithDecimals,
			Price:                    go_decimal.MustStart(reserveBNBWithDecimals).MustDivForString(reserveTokenWithDecimals),
		}, nil
	}

	return &ReserveInfoType{
		ReserveBNBWithDecimals:   tokenInfo.Funds,
		ReserveTokenWithDecimals: tokenInfo.Offers,
		Price:                    go_decimal.MustStart(tokenInfo.LastPrice).MustUnShiftedBy(18).EndForString(),
	}, nil
}

func ParseSwapInfos(
	wallet *go_coin_eth.Wallet,
	txReceipt *types.Receipt,
) ([]*type_.SwapResultType, []*TradeEventType, error) {
	swapResults := make([]*type_.SwapResultType, 0)
	tradeEvents := make([]*TradeEventType, 0)

	parsedAbi, err := abi.JSON(strings.NewReader(constant.TokenManagerABI))
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
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
		switch log.Topics[0] {
		case parsedAbi.Events["TokenPurchase"].ID:
			err = boundContract.UnpackLog(&r, "TokenPurchase", *log)
			if err != nil {
				return nil, nil, errors.Wrap(err, "")
			}
			swapResults = append(swapResults, &type_.SwapResultType{
				Type_:                   "buy",
				UserAddress:             r.Account,
				ETHAmountWithDecimals:   r.Cost,
				TokenAmountWithDecimals: r.Amount,
				TokenAddress:            r.Token,
				NetworkFee:              go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
				TxId:                    txReceipt.TxHash.String(),
				BlockNumber:             txReceipt.BlockNumber.Uint64(),
			})
			tradeEvents = append(tradeEvents, &r)
		case parsedAbi.Events["TokenSale"].ID:
			err = boundContract.UnpackLog(&r, "TokenSale", *log)
			if err != nil {
				return nil, nil, errors.Wrap(err, "")
			}
			swapResults = append(swapResults, &type_.SwapResultType{
				Type_:                   "sell",
				UserAddress:             r.Account,
				ETHAmountWithDecimals:   r.Cost,
				TokenAmountWithDecimals: r.Amount,
				TokenAddress:            r.Token,
				NetworkFee:              go_decimal.MustStart(txReceipt.EffectiveGasPrice).MustMulti(txReceipt.GasUsed).RoundDown(0).MustEndForBigInt(),
				TxId:                    txReceipt.TxHash.String(),
				BlockNumber:             txReceipt.BlockNumber.Uint64(),
			})
			tradeEvents = append(tradeEvents, &r)
		default:
			continue
		}

	}

	return swapResults, tradeEvents, nil
}

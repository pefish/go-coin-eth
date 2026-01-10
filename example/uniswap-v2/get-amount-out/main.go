package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v2 "github.com/pefish/go-coin-eth/uniswap-v2"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	t_logger "github.com/pefish/go-interface/t-logger"
	go_logger "github.com/pefish/go-logger"
)

func main() {
	envMap, _ := godotenv.Read("./.env")
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	logger = go_logger.NewLogger(t_logger.Level(os.Getenv("LOG_LEVEL")))

	err := do()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

// var poolKey = &uniswap_v2.PoolKeyType{
// 	TokenAddress: common.HexToAddress("0x6bdcCe4A559076e37755a78Ce0c06214E59e4444"),
// 	BaseTokenAddress: go_coin_eth.WBNBAddress,
// }
// var tokenAddress = common.HexToAddress("0x6bdcCe4A559076e37755a78Ce0c06214E59e4444")

// var poolKey = &uniswap_v2.PoolKeyType{
// 	TokenAddress: common.HexToAddress("0x73b84F7E3901F39FC29F3704a03126D317Ab4444"),
// 	BaseTokenAddress: go_coin_eth.WBNBAddress,
// }
// var tokenAddress = common.HexToAddress("0x73b84F7E3901F39FC29F3704a03126D317Ab4444")

var poolInfo = &uniswap_v2.PoolInfoType{
	BaseTokenAddress: go_coin_eth.WBNBAddress, // WBNB
	TokenAddress:     common.HexToAddress("0x924fa68a0FC644485b8df8AbfA0A41C2e7744444"),
}
var tokenAddress = common.HexToAddress("0x924fa68a0FC644485b8df8AbfA0A41C2e7744444")

var tokenIn = tokenAddress
var amountInWithDecimals = go_decimal.MustStart("62476766410090636150").MustShiftedBy(0).MustEndForBigInt()

var logger i_logger.ILogger = &i_logger.DefaultLogger

func do() error {
	wallet, err := go_coin_eth.NewWallet(
		logger,
	).InitRemote(&go_coin_eth.UrlParam{
		RpcUrl: os.Getenv("NODE_HTTPS"),
		WsUrl:  os.Getenv("NODE_WSS"),
	})
	if err != nil {
		return err
	}
	trader := uniswap_v2.New(logger, wallet)

	quoteResult, err := trader.GetAmountsOut(
		uniswap_v2.Pancake_BscChainRouter,
		poolInfo,
		tokenIn,
		amountInWithDecimals,
	)
	if err != nil {
		return err
	}
	spew.Dump(quoteResult.String())
	return nil
}

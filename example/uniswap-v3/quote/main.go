package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v3 "github.com/pefish/go-coin-eth/uniswap-v3"
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

var poolKey = &uniswap_v3.PoolKeyType{
	Token0: common.HexToAddress("0x55d398326f99059ff775485246999027b3197955"), // USDT
	Token1: common.HexToAddress("0x825459139c897d769339f295e962396c4f9e4a4d"), // GAME
	Fee:    100,
}
var tokenAddress = common.HexToAddress("0x825459139c897d769339f295e962396c4f9e4a4d")

var tokenIn = tokenAddress
var amountInWithDecimals = go_decimal.MustStart("1000").MustShiftedBy(18).MustEndForBigInt()

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
	trader := uniswap_v3.New(logger, wallet)

	quoteResult, err := trader.QuoteExactInputSingle(
		uniswap_v3.Pancake_BscChainQuoter,
		poolKey,
		tokenIn,
		amountInWithDecimals,
	)
	if err != nil {
		return err
	}
	spew.Dump(quoteResult)
	return nil
}

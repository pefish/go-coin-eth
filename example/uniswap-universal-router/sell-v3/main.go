package main

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_universal_router "github.com/pefish/go-coin-eth/uniswap-universal-router"
	uniswap_v3 "github.com/pefish/go-coin-eth/uniswap-v3"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	t_logger "github.com/pefish/go-interface/t-logger"
	go_logger "github.com/pefish/go-logger"
)

var logger i_logger.ILogger = &i_logger.DefaultLogger

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

// var poolKey = &uniswap_v3.PoolKeyType{
// 	Token0: common.HexToAddress("0x5C85D6C6825aB4032337F11Ee92a72DF936b46F6"),
// 	Token1: go_coin_eth.WBNBAddress, // WBNB
// 	Fee:    2500,
// }
// var tokenAddress = common.HexToAddress("0x5C85D6C6825aB4032337F11Ee92a72DF936b46F6")

var poolKey = &uniswap_v3.PoolKeyType{
	Token0: common.HexToAddress("0x55d398326f99059ff775485246999027b3197955"), // USDT
	Token1: common.HexToAddress("0x825459139c897d769339f295e962396c4f9e4a4d"), // GAME
	Fee:    100,
}
var tokenAddress = common.HexToAddress("0x825459139c897d769339f295e962396c4f9e4a4d")

var amountInWithDecimals = go_decimal.MustStart("0").MustShiftedBy(18).MustEndForBigInt()

func do() error {
	wallet, err := go_coin_eth.NewWallet(
		&i_logger.DefaultLogger,
	).InitRemote(&go_coin_eth.UrlParam{
		RpcUrl: os.Getenv("NODE_HTTPS"),
		WsUrl:  os.Getenv("NODE_WSS"),
	})
	if err != nil {
		return err
	}

	priv := os.Getenv("PRIV")
	userAddress, err := wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return err
	}
	logger.InfoF("userAddress: %s", userAddress)

	var tokenIn common.Address
	if tokenAddress == poolKey.Token0 {
		tokenIn = poolKey.Token0
	} else {
		tokenIn = poolKey.Token1
	}

	if go_decimal.MustStart(amountInWithDecimals).MustEq(0) {
		balance, err := wallet.TokenBalance(tokenIn, userAddress)
		if err != nil {
			return err
		}
		amountInWithDecimals = balance
	}

	router := uniswap_universal_router.New(&i_logger.DefaultLogger, wallet)

	r, err := router.SwapExactInputV3(
		context.Background(),
		priv,
		poolKey,
		tokenIn,
		amountInWithDecimals,
		big.NewInt(0),
		20_0000,
		big.NewInt(100000000),
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

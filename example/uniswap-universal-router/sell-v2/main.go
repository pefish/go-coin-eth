package main

import (
	"context"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_universal_router "github.com/pefish/go-coin-eth/uniswap-universal-router"
	uniswap_v2 "github.com/pefish/go-coin-eth/uniswap-v2"
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

var poolKey = &uniswap_v2.PoolKeyType{
	Token0: go_coin_eth.WBNBAddress, // WBNB
	Token1: common.HexToAddress("0xd5eaAaC47bD1993d661bc087E15dfb079a7f3C19"),
}
var tokenAddress = common.HexToAddress("0xd5eaAaC47bD1993d661bc087E15dfb079a7f3C19")

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

	r, err := router.SwapExactInputV2(
		context.Background(),
		priv,
		poolKey,
		tokenIn,
		amountInWithDecimals,
		uniswap_v2.Pancake_BscChainRouter,
		&uniswap_universal_router.SwapOpts{
			Slippage: 100,
		},
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

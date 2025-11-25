package main

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v2_trade "github.com/pefish/go-coin-eth/uniswap-v2-trade"
	"github.com/pefish/go-coin-eth/uniswap-v2-trade/constant"
	i_logger "github.com/pefish/go-interface/i-logger"
)

func main() {
	envMap, _ := godotenv.Read("./.env")
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	err := do()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

const tokenAddress = "0x44440f83419de123d7d411187adb9962db017d03"
const tokenAmount = 119.034792003962322261

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
	trader := uniswap_v2_trade.New(&i_logger.DefaultLogger, wallet)

	r, err := trader.SellByExactToken(
		context.Background(),
		os.Getenv("PRIV"),
		tokenAmount,
		constant.Pancake_BSCRouter,
		tokenAddress,
		&uniswap_v2_trade.SellByExactTokenOpts{
			MaxFeePerGas: big.NewInt(100000000), // bsc 最少要给 5000_0000
		},
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

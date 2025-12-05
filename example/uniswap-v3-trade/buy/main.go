package main

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v3_trade "github.com/pefish/go-coin-eth/uniswap-v3-trade"
	"github.com/pefish/go-coin-eth/uniswap-v3-trade/constant"
	go_decimal "github.com/pefish/go-decimal"
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

const tokenAddress = "0x276b792D11B9e3712FE6A78A460a0DEb416baB0A"
const fee = 100
const bnbAmount = "0.0001"

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
	trader := uniswap_v3_trade.New(&i_logger.DefaultLogger, wallet)

	r, err := trader.BuyByExactETH(
		context.Background(),
		os.Getenv("PRIV"),
		go_decimal.MustStart(bnbAmount).MustShiftedBy(18).MustEndForBigInt(),
		constant.Pancake_BSCRouter,
		tokenAddress,
		fee,
		&uniswap_v3_trade.BuyByExactETHOpts{
			MaxFeePerGas: big.NewInt(100000000), // bsc 最少要给 5000_0000
		},
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

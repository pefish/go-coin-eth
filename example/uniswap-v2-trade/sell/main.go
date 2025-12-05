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

const tokenAddress = "0x44440f83419de123d7d411187adb9962db017d03"
const tokenAmount = "0"

var logger i_logger.ILogger = &i_logger.DefaultLogger

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

	trader := uniswap_v2_trade.New(&i_logger.DefaultLogger, wallet)

	tokenAmountWithDecimals := go_decimal.MustStart(tokenAmount).MustShiftedBy(18).MustEndForBigInt()
	if tokenAmount == "0" {
		tokenAmountWithDecimals, err = wallet.TokenBalance(tokenAddress, userAddress)
		if err != nil {
			return err
		}
	}

	r, err := trader.SellByExactToken(
		context.Background(),
		priv,
		tokenAmountWithDecimals,
		constant.Pancake_BSCRouter,
		tokenAddress,
		&uniswap_v2_trade.TradeOpts{
			MaxFeePerGas: big.NewInt(100000000), // bsc 最少要给 5000_0000
		},
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

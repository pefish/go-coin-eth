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
	fourmeme_lib "github.com/pefish/go-coin-eth/fourmeme-lib"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	t_logger "github.com/pefish/go-interface/t-logger"
	go_logger "github.com/pefish/go-logger"
)

var tokenAddress = common.HexToAddress("0x33ba5243ac3ed4c3d1ee630515fff745423b4444")

const bnbAmount = "0.0001"

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
	priv := os.Getenv("PRIV")
	userAddress, err := wallet.PrivateKeyToAddress(priv)
	if err != nil {
		return err
	}
	logger.InfoF("userAddress: %s", userAddress)
	// return nil

	r, tradeEvent, err := fourmeme_lib.Buy(
		context.Background(),
		wallet,
		os.Getenv("PRIV"),
		tokenAddress,
		go_decimal.MustStart(bnbAmount).MustShiftedBy(18).MustEndForBigInt(),
		big.NewInt(0),
		big.NewInt(100000000), // bsc 最少要给 5000_0000
	)
	if err != nil {
		return err
	}
	spew.Dump(r)
	spew.Dump(tradeEvent)

	return nil
}

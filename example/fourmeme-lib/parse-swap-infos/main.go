package main

import (
	"context"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	fourmeme_lib "github.com/pefish/go-coin-eth/fourmeme-lib"
	i_logger "github.com/pefish/go-interface/i-logger"
	t_logger "github.com/pefish/go-interface/t-logger"
	go_logger "github.com/pefish/go-logger"
)

const txId = "0x0bff0a5951c18c61547536511ab8a3fc54793abdd043124107595f9555ff2cff"

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

	tr, err := wallet.RemoteRpcClient.TransactionReceipt(context.Background(), common.HexToHash(txId))
	if err != nil {
		return err
	}

	r, tradeEvent, err := fourmeme_lib.ParseSwapInfos(
		wallet,
		tr,
	)
	if err != nil {
		return err
	}
	spew.Dump(r)
	spew.Dump(tradeEvent)

	return nil
}

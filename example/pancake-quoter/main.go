package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	pancake_quoter "github.com/pefish/go-coin-eth/pancake-quoter"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
)

var tokenAddress = common.HexToAddress("0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82")
var fee = uint64(2500)

var logger = &i_logger.DefaultLogger

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

	quoter := pancake_quoter.New(logger, wallet)
	tokenAmount, err := quoter.QuoteExactInputSingle(
		go_coin_eth.WBNBAddress,
		tokenAddress,
		fee,
		go_decimal.MustStart("0.001").MustShiftedBy(18).MustEndForBigInt(),
	)
	if err != nil {
		return err
	}
	spew.Dump(tokenAmount.String())

	return nil
}

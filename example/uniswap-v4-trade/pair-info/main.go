package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_v4 "github.com/pefish/go-coin-eth/uniswap-v4"
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

var poolID = common.HexToHash("0x101552cfd9d16f17db7d11fde6082e4671e9fe39cb21679bb3fad5be9e5ec2c9")

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
	trader := uniswap_v4.New(logger, wallet)

	pairInfo, err := trader.PairInfoByPoolID(poolID)
	if err != nil {
		return err
	}
	spew.Dump(pairInfo)
	return nil
}

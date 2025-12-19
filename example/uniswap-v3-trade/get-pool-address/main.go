package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	"github.com/pefish/go-coin-eth/uniswap-v3"
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

var poolKey = &uniswap_v3.PoolKeyType{
	Token0: common.HexToAddress("0x6952c5408b9822295ba4a7e694d0C5FfDB8fE320"),
	Token1: go_coin_eth.WBNBAddress,
	Fee:    100,
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

	trader := uniswap_v3.New(logger, wallet)

	r, err := trader.GetPoolAddress(
		poolKey,
		uniswap_v3.PancakeV3FactoryAddress,
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

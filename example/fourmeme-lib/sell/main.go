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
	"github.com/pefish/go-coin-eth/fourmeme-lib/constant"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	t_logger "github.com/pefish/go-interface/t-logger"
	go_logger "github.com/pefish/go-logger"
)

var tokenAddress = common.HexToAddress("0x33ba5243ac3ed4c3d1ee630515fff745423b4444")

// const tokenAmount = "1000"
const tokenAmount = "0"

var maxFeePerGas = big.NewInt(100000000) // bsc 最少要给 5000_0000

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
	approvedAmount, err := wallet.ApprovedAmount(tokenAddress, userAddress, constant.TokenManagerAddress)
	if err != nil {
		return err
	}
	logger.InfoF("approvedAmount: %s", approvedAmount.String())
	tokenAmountWithDecimals := go_decimal.MustStart(tokenAmount).MustShiftedBy(18).MustEndForBigInt()
	if tokenAmount == "0" {
		tokenAmountWithDecimals, err = wallet.TokenBalance(tokenAddress, userAddress)
		if err != nil {
			return err
		}
	}
	if approvedAmount.Cmp(tokenAmountWithDecimals) < 0 {
		logger.InfoF("need approve first")
		tr, err := wallet.ApproveWait(
			context.Background(),
			priv,
			tokenAddress,
			constant.TokenManagerAddress,
			nil,
			&go_coin_eth.CallMethodOpts{
				MaxFeePerGas:   maxFeePerGas,
				GasLimit:       250000,
				IsPredictError: false,
			},
		)
		if err != nil {
			return err
		}
		logger.InfoF("approve done. txId: %s", tr.TxHash.String())
	}

	r, tradeEvent, err := fourmeme_lib.Sell(
		context.Background(),
		wallet,
		os.Getenv("PRIV"),
		tokenAddress,
		tokenAmountWithDecimals,
		maxFeePerGas,
	)
	if err != nil {
		return err
	}
	spew.Dump(r)
	spew.Dump(tradeEvent)

	return nil
}

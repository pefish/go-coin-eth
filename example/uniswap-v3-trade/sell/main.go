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
	uniswap_v3 "github.com/pefish/go-coin-eth/uniswap-v3"
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

var tokenAddress = common.HexToAddress("0x55ad16bd573b3365f43a9daeb0cc66a73821b4a5")

const fee = 100
const tokenAmount = "0"

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

	tokenAmountWithDecimals := go_decimal.MustStart(tokenAmount).MustShiftedBy(18).MustEndForBigInt()
	if go_decimal.MustStart(tokenAmountWithDecimals).MustEq(0) {
		balance, err := wallet.TokenBalance(tokenAddress, userAddress)
		if err != nil {
			return err
		}
		tokenAmountWithDecimals = balance
	}

	allowanceAmount, err := wallet.ApprovedAmount(tokenAddress, userAddress, uniswap_v3.Pancake_BSCRouter)
	if err != nil {
		return err
	}
	logger.InfoF("approvedAmount: %s", allowanceAmount.String())
	if allowanceAmount.Cmp(tokenAmountWithDecimals) < 0 {
		logger.InfoF("need approve first")
		tr, err := wallet.ApproveWait(
			context.Background(),
			priv,
			tokenAddress,
			uniswap_v3.Pancake_BSCRouter,
			nil,
			&go_coin_eth.CallMethodOpts{
				MaxFeePerGas:   big.NewInt(100000000),
				GasLimit:       50000,
				IsPredictError: true,
			},
		)
		if err != nil {
			return err
		}
		logger.InfoF("approve done. txId: %s", tr.TxHash.String())
	}

	trader := uniswap_v3.New(logger, wallet)

	r, err := trader.SwapExactInput(
		context.Background(),
		priv,
		tokenAmountWithDecimals,
		uniswap_v3.Pancake_BSCRouter,
		tokenAddress,
		go_coin_eth.WBNBAddress,
		fee,
		&uniswap_v3.SwapExactInputOpts{
			WETHAddress:  go_coin_eth.WBNBAddress,
			MaxFeePerGas: big.NewInt(100000000), // bsc 最少要给 5000_0000
		},
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

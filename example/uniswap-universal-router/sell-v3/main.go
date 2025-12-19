package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	uniswap_universal_router "github.com/pefish/go-coin-eth/uniswap-universal-router"
	"github.com/pefish/go-coin-eth/uniswap-v3"
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

var poolKey = &uniswap_v3.PoolKeyType{
	Token0: common.HexToAddress("0x6952c5408b9822295ba4a7e694d0C5FfDB8fE320"),
	Token1: go_coin_eth.WBNBAddress,
	Fee:    100,
}

var tokenIn = poolKey.Token0
var amountInWithDecimals = go_decimal.MustStart("0").MustShiftedBy(18).MustEndForBigInt()

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

	if go_decimal.MustStart(amountInWithDecimals).MustEq(0) {
		balance, err := wallet.TokenBalance(tokenIn, userAddress)
		if err != nil {
			return err
		}
		amountInWithDecimals = balance
	}

	// 检查给 permit2 的授权
	allowanceAmount, err := wallet.ApprovedAmount(tokenIn, userAddress, uniswap_universal_router.Permit2)
	if err != nil {
		return err
	}
	logger.InfoF("Permit2 approvedAmount: %s", allowanceAmount.String())
	if allowanceAmount.Cmp(amountInWithDecimals) < 0 {
		logger.InfoF("Permit2 need approve first")
		tr, err := wallet.ApproveWait(
			context.Background(),
			priv,
			tokenIn,
			uniswap_universal_router.Permit2,
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
		logger.InfoF("Permit2 approve done. txId: %s", tr.TxHash.String())
	}
	router := uniswap_universal_router.New(&i_logger.DefaultLogger, wallet)
	// 要先检查有没有通过 permit2 给 universal_router 授权
	allowanceInfo, err := router.Allowance(userAddress, tokenIn, uniswap_universal_router.Universal_Router)
	if err != nil {
		return err
	}
	logger.InfoF("approvedAmount: %s", allowanceInfo.Amount.String())
	if allowanceInfo.Amount.Cmp(amountInWithDecimals) < 0 {
		logger.InfoF("need approve first")
		tr, err := router.ApproveWait(
			context.Background(),
			priv,
			tokenIn,
			uniswap_universal_router.Universal_Router,
			nil,
			time.Now().UnixMilli()+3600*1000, // 1 hour
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

	r, err := router.SwapExactInputV3(
		context.Background(),
		priv,
		poolKey,
		tokenIn,
		amountInWithDecimals,
		big.NewInt(0),
		20_0000,
		big.NewInt(100000000),
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

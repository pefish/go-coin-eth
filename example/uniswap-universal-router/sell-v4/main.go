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
	"github.com/pefish/go-coin-eth/uniswap-v4"
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

var tokenAddress = common.HexToAddress("0x85375D3e9c4a39350f1140280a8b0De6890A40e7")
var poolID = common.HexToHash("0x416e5132b7c80008cd32cf62439ea38e36c8eec0bbd16b78b3260a0fc5fa8c59")
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

	router := uniswap_universal_router.New(logger, wallet)

	uniswapV4 := uniswap_v4.New(logger, wallet)
	pairInfo, err := uniswapV4.PairInfoByPoolID(poolID)
	if err != nil {
		return err
	}
	var tokenIn common.Address
	if tokenAddress == pairInfo.Currency0 {
		tokenIn = pairInfo.Currency0
	} else {
		tokenIn = pairInfo.Currency1
	}

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

	// 要先检查有没有通过 permit2 给 universal_router 授权
	allowanceInfo, err := router.Allowance(userAddress, tokenIn, uniswap_universal_router.Universal_Router)
	if err != nil {
		return err
	}
	logger.InfoF("Universal_Router approvedAmount: %s", allowanceInfo.Amount.String())
	if allowanceInfo.Amount.Cmp(amountInWithDecimals) < 0 {
		logger.InfoF("Universal_Router need approve first")
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
		logger.InfoF("Universal_Router approve done. txId: %s", tr.TxHash.String())
	}

	r, err := router.SwapExactInputV4(
		context.Background(),
		priv,
		pairInfo,
		tokenIn,
		amountInWithDecimals,
		big.NewInt(0),
		220000,
		big.NewInt(100000000),
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

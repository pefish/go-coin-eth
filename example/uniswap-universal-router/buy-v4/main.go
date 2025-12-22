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
	uniswap_universal_router "github.com/pefish/go-coin-eth/uniswap-universal-router"
	uniswap_v4 "github.com/pefish/go-coin-eth/uniswap-v4"
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

// var tokenAddress = common.HexToAddress("0x85375D3e9c4a39350f1140280a8b0De6890A40e7") // BNB/SIGMA
// var poolID = common.HexToHash("0x416e5132b7c80008cd32cf62439ea38e36c8eec0bbd16b78b3260a0fc5fa8c59")

// var tokenAddress = common.HexToAddress("0x4829A1D1fB6DED1F81d26868ab8976648baF9893") // RTX/USDT
// var poolID = common.HexToHash("0x9f57ccbb2a7a89120cbdc8dad277d6e82aa9b2c3925e148033963a22e1f57b5e")

var tokenAddress = common.HexToAddress("0x810DF4c7Daf4eE06AE7c621D0680E73a505C9A06")
var poolID = common.HexToHash("0xa255622f283b81ccbe7ddfae8b1da213d436407991e651c2503da7112d06c0e5")

var amountInWithDecimals = go_decimal.MustStart("0.0001").MustShiftedBy(18).MustEndForBigInt()

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
		tokenIn = pairInfo.Currency1
	} else {
		tokenIn = pairInfo.Currency0
	}

	r, err := router.SwapExactInputV4(
		context.Background(),
		priv,
		pairInfo,
		tokenIn,
		amountInWithDecimals,
		100,
		300000,
		big.NewInt(100000000),
		uniswap_v4.Pancake_BscChainQuoter,
	)
	if err != nil {
		return err
	}
	spew.Dump(r)

	return nil
}

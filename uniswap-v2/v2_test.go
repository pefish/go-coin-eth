package uniswap_v2

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

func TestWallet_WETHAddressFromRouter(t *testing.T) {
	wallet, err := go_coin_eth.NewWallet(&i_logger.DefaultLogger).InitRemote(&go_coin_eth.UrlParam{
		RpcUrl: "https://mainnet.base.org",
	})
	go_test_.Equal(t, nil, err)
	trader := New(&i_logger.DefaultLogger, wallet)
	wethAddress, err := trader.WETHAddressFromRouter(
		common.HexToAddress("0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24"),
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x4200000000000000000000000000000000000006", wethAddress)
}

func TestWallet_GetAmountsOut(t *testing.T) {
	wallet, err := go_coin_eth.NewWallet(&i_logger.DefaultLogger).InitRemote(&go_coin_eth.UrlParam{
		RpcUrl: "https://eth-mainnet.g.alchemy.com/v2/",
	})
	go_test_.Equal(t, nil, err)
	trader := New(&i_logger.DefaultLogger, wallet)
	result, err := trader.GetAmountsOut(
		common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"),
		go_decimal.MustStart("100000000000000000").MustEndForBigInt(),
		[]common.Address{
			go_coin_eth.WETHAddress,
			common.HexToAddress("0xD06e204b2DE9cBCC19E1Fa1F9523f2189aF38c55"),
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, true, result != nil)
}

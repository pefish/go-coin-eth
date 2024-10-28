package go_coin_eth

import (
	"testing"

	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

func TestWallet_WETHAddressFromRouter(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.base.org",
	})

	go_test_.Equal(t, nil, err)
	wethAddress, err := wallet.WETHAddressFromRouter("0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x4200000000000000000000000000000000000006", wethAddress)
}

func TestWallet_GetAmountsOut(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://eth-mainnet.g.alchemy.com/v2/",
	})

	go_test_.Equal(t, nil, err)
	result, err := wallet.GetAmountsOut(
		"0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
		"100000000000000000",
		[]string{
			WETHAddress,
			"0xD06e204b2DE9cBCC19E1Fa1F9523f2189aF38c55",
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, true, result != "")
}

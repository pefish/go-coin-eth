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

package uniswap_v2

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
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

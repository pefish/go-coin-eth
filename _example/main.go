package main

import (
	go_coin_eth "github.com/pefish/go-coin-eth"
	"log"
)

func main() {
	_, err := go_coin_eth.NewWallet("wss://mainnet.infura.io/ws/v3/9442f24048d94dbd9a588d3e4e2eac8b")
	if err != nil {
		log.Fatal(err)
	}
	//err = wallet.WatchPendingTxByWs()
	//if err != nil {
	//	log.Fatal(err)
	//}
}

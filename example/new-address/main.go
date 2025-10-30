package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	go_format "github.com/pefish/go-format"
	i_logger "github.com/pefish/go-interface/i-logger"
)

const index = 0

func main() {
	envMap, _ := godotenv.Read("./.env")
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	err := do()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func do() error {
	wallet := go_coin_eth.NewWallet(
		&i_logger.DefaultLogger,
	)
	fmt.Println(os.Getenv("Mnemonic"), os.Getenv("Password"))
	seed := wallet.SeedHexByMnemonic(os.Getenv("Mnemonic"), os.Getenv("Password"))
	deriveResult, err := wallet.DeriveFromPath(seed, fmt.Sprintf("m/44'/60'/0'/0/%d", index))
	if err != nil {
		return err
	}
	fmt.Printf(
		"<mnemonic: %s> <index: %d> <address: %s> <priv: %s> <enctr-priv: %s>\n",
		os.Getenv("Mnemonic"),
		index,
		deriveResult.Address,
		deriveResult.PrivateKey,
		go_format.EncodePefish(deriveResult.PrivateKey),
	)

	return nil
}

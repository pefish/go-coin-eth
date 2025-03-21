package go_coin_eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	go_decimal "github.com/pefish/go-decimal"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

var contractAddress = "0x509Ee0d083DdF8AC028f2a56731412edD63223B9"
var abiStr = `[{"inputs":[{"internalType":"string","name":"name","type":"string"},{"internalType":"string","name":"symbol","type":"string"},{"internalType":"uint8","name":"decimals","type":"uint8"},{"internalType":"uint256","name":"totalSupply","type":"uint256"},{"internalType":"address","name":"receiveAccount","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"src","type":"address"},{"indexed":true,"internalType":"address","name":"guy","type":"address"},{"indexed":false,"internalType":"uint256","name":"wad","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"src","type":"address"},{"indexed":true,"internalType":"address","name":"dst","type":"address"},{"indexed":false,"internalType":"uint256","name":"wad","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"subtractedValue","type":"uint256"}],"name":"decreaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"addedValue","type":"uint256"}],"name":"increaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`

func init() {
	//wallet1, err := NewWallet("https://ropsten.infura.io/v3/7594e560416349f79c8ef6ff286d83fc")
	//go_test_.Equal(t, nil, err)
	//wallet = wallet1
}

func TestContract_BuildCallMethodTx(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://rpc.ankr.com/eth_goerli",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	tx, err := wallet1.BuildCallMethodTx(
		"",
		contractAddress,
		abiStr,
		"transfer",
		&CallMethodOpts{
			MaxFeePerGas: new(big.Int).SetUint64(1000000000),
		},
		[]interface{}{
			common.HexToAddress("0x2117210296c2993Cfb4c6790FEa1bEB3ECe8Ac06"),
			big.NewInt(1000000000000000000),
		},
	)
	go_test_.Equal(t, true, tx == nil)
	go_test_.Equal(t, true, err != nil)
	go_test_.Equal(t, "invalid length, need 256 bits", err.Error())
}

func TestWallet_WatchLogs(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "wss://mainnet.infura.io/ws/v3/7594e560416349f79c8ef6ff286d83fc",
	})
	go_test_.Equal(t, nil, err)
	defer wallet.Close()
	resultChan, contractInstance, err := wallet.WatchLogsByWs(
		context.Background(),
		"0xdac17f958d2ee523a2206206994597c13d831ec7",
		`[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_upgradedAddress","type":"address"}],"name":"deprecate","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"deprecated","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_evilUser","type":"address"}],"name":"addBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"upgradedAddress","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balances","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"maximumFee","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_maker","type":"address"}],"name":"getBlackListStatus","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowed","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"who","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newBasisPoints","type":"uint256"},{"name":"newMaxFee","type":"uint256"}],"name":"setParams","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"issue","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"redeem","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"basisPointsRate","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"isBlackListed","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_clearedUser","type":"address"}],"name":"removeBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"MAX_UINT","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_blackListedUser","type":"address"}],"name":"destroyBlackFunds","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[{"name":"_initialSupply","type":"uint256"},{"name":"_name","type":"string"},{"name":"_symbol","type":"string"},{"name":"_decimals","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Redeem","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newAddress","type":"address"}],"name":"Deprecate","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"feeBasisPoints","type":"uint256"},{"indexed":false,"name":"maxFee","type":"uint256"}],"name":"Params","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_blackListedUser","type":"address"},{"indexed":false,"name":"_balance","type":"uint256"}],"name":"DestroyedBlackFunds","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"AddedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"RemovedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"}]`,
		"Transfer",
		nil,
	)
	go_test_.Equal(t, nil, err)
	for {
		select {
		case log := <-resultChan:
			result := make(map[string]interface{})
			err := contractInstance.UnpackLogIntoMap(result, "Transfer", log)
			go_test_.Equal(t, nil, err)
			go_test_.Equal(t, true, result["from"].(common.Address).String() != "")
			goto exit
		}
	}
exit:
}

func TestWallet_FindLogs(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet.Close()
	err = wallet.FindLogs(
		func(contractInstance *bind.BoundContract, logs []types.Log) error {
			go_test_.Equal(t, 1066, len(logs))
			var transferLog struct {
				From  common.Address
				To    common.Address
				Value *big.Int
			}
			err = contractInstance.UnpackLog(&transferLog, "Transfer", logs[0])
			go_test_.Equal(t, nil, err)
			return nil
		},
		"0xdac17f958d2ee523a2206206994597c13d831ec7",
		`[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_upgradedAddress","type":"address"}],"name":"deprecate","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"deprecated","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_evilUser","type":"address"}],"name":"addBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"upgradedAddress","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balances","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"maximumFee","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_maker","type":"address"}],"name":"getBlackListStatus","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowed","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"who","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newBasisPoints","type":"uint256"},{"name":"newMaxFee","type":"uint256"}],"name":"setParams","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"issue","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"redeem","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"basisPointsRate","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"isBlackListed","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_clearedUser","type":"address"}],"name":"removeBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"MAX_UINT","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_blackListedUser","type":"address"}],"name":"destroyBlackFunds","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[{"name":"_initialSupply","type":"uint256"},{"name":"_name","type":"string"},{"name":"_symbol","type":"string"},{"name":"_decimals","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Redeem","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newAddress","type":"address"}],"name":"Deprecate","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"feeBasisPoints","type":"uint256"},{"indexed":false,"name":"maxFee","type":"uint256"}],"name":"Params","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_blackListedUser","type":"address"},{"indexed":false,"name":"_balance","type":"uint256"}],"name":"DestroyedBlackFunds","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"AddedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"RemovedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"}]`,
		"Transfer",
		new(big.Int).SetUint64(11424704),
		new(big.Int).SetUint64(11424724),
		2000,
	)
	go_test_.Equal(t, nil, err)
}

//func TestWallet_FindLogs1(t *testing.T) {
//	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
//		RpcUrl: "https://data-seed-prebsc-2-s1.binance.org:8545/",
//		WsUrl:  "",
//	})
//	go_test_.Equal(t, nil, err)
//	defer wallet.Close()
//	err = wallet.FindLogs(
//		func(contractInstance *bind.BoundContract, logs []types.Log) error {
//			go_test_.Equal(t, 1, len(logs))
//			return nil
//		},
//		"0x375Ee04Fe818D896f98ce16199604C4706b7D74E",
//		`[{"inputs":[{"internalType":"string","name":"name","type":"string"},{"internalType":"string","name":"symbol","type":"string"},{"internalType":"string","name":"baseTokenURI","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"approved","type":"address"},{"indexed":true,"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"operator","type":"address"},{"indexed":false,"internalType":"bool","name":"approved","type":"bool"}],"name":"ApprovalForAll","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"role","type":"bytes32"},{"indexed":true,"internalType":"bytes32","name":"previousAdminRole","type":"bytes32"},{"indexed":true,"internalType":"bytes32","name":"newAdminRole","type":"bytes32"}],"name":"RoleAdminChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"role","type":"bytes32"},{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":true,"internalType":"address","name":"sender","type":"address"}],"name":"RoleGranted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"role","type":"bytes32"},{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":true,"internalType":"address","name":"sender","type":"address"}],"name":"RoleRevoked","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":true,"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"inputs":[],"name":"DEFAULT_ADMIN_ROLE","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"MINTER_ROLE","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"PAUSER_ROLE","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"approve","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"burn","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"getApproved","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"}],"name":"getRoleAdmin","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getRoleMember","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"}],"name":"getRoleMemberCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"},{"internalType":"address","name":"account","type":"address"}],"name":"grantRole","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"},{"internalType":"address","name":"account","type":"address"}],"name":"hasRole","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"},{"internalType":"address","name":"account","type":"address"}],"name":"renounceRole","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"},{"internalType":"address","name":"account","type":"address"}],"name":"revokeRole","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"bytes","name":"_data","type":"bytes"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"operator","type":"address"},{"internalType":"bool","name":"approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes4","name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"tokenByIndex","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"tokenOfOwnerByIndex","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"tokenURI","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"}]`,
//		"Transfer",
//		new(big.Int).SetUint64(1),
//		nil,
//		2000,
//		[]interface{}{
//			ZeroAddress,
//		}, // filter first indexed param
//		[]interface{}{}, // filter second indexed param
//		[]interface{}{
//			new(big.Int).SetUint64(6),
//		},
//	)
//	go_test_.Equal(t, nil, err)
//}

func TestWallet_SendRawTransaction(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	_, err = wallet1.SendRawTransaction("0xf8693f850bc97c324083041c8794bbd3c0c794f40c4f993b03f65343acc6fcfcb2e2808441fe00a025a015e4e95a51191607472b98fcbd168bd32aaadeed40c074ed2ec82044ec5a5e71a0484a64cf7fca58254404343926e91c96aca09cc2187ec83cc3be31f619a6f1df")
	go_test_.Equal(t, true, err != nil)
	go_test_.Equal(t, true, strings.Contains(err.Error(), "nonce too low"))
}

func TestWallet_WatchPendingTxByWs(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet.Close()
	strChan := make(chan string)
	go func() {
		err = wallet.WatchPendingTxByWs(strChan)
		go_test_.Equal(t, nil, err)
	}()
	for txHash := range strChan {
		go_test_.Equal(t, true, len(txHash) > 0)
		break
	}
}

func TestWallet_CallContractConstant(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	result := new(big.Int)
	err = wallet1.CallContractConstant(
		&result,
		"0xdac17f958d2ee523a2206206994597c13d831ec7",
		`[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_upgradedAddress","type":"address"}],"name":"deprecate","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"deprecated","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_evilUser","type":"address"}],"name":"addBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"upgradedAddress","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balances","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"maximumFee","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_maker","type":"address"}],"name":"getBlackListStatus","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowed","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"who","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newBasisPoints","type":"uint256"},{"name":"newMaxFee","type":"uint256"}],"name":"setParams","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"issue","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"redeem","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"basisPointsRate","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"isBlackListed","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_clearedUser","type":"address"}],"name":"removeBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"MAX_UINT","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_blackListedUser","type":"address"}],"name":"destroyBlackFunds","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[{"name":"_initialSupply","type":"uint256"},{"name":"_name","type":"string"},{"name":"_symbol","type":"string"},{"name":"_decimals","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Redeem","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newAddress","type":"address"}],"name":"Deprecate","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"feeBasisPoints","type":"uint256"},{"indexed":false,"name":"maxFee","type":"uint256"}],"name":"Params","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_blackListedUser","type":"address"},{"indexed":false,"name":"_balance","type":"uint256"}],"name":"DestroyedBlackFunds","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"AddedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"RemovedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"}]`,
		"balanceOf",
		&bind.CallOpts{
			Pending: false,
		},
		[]interface{}{
			common.HexToAddress("0xd9eb7d4dff36a2801d1ec42e75260b6e9e283e62"),
		},
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(result.String())
}

func TestWallet_CallContractConstant1(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://data-seed-prebsc-1-s1.binance.org:8545/",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()

	type Asset struct {
		Token     common.Address `json:"token"`
		TokenId   *big.Int       `json:"tokenId"`
		AssetType uint8          `json:"assetType"`
	}

	type OrderKey struct {
		Owner     common.Address `json:"owner"`
		Salt      *big.Int       `json:"salt"`
		SellAsset Asset          `json:"sellAsset"`
		BuyAsset  Asset          `json:"buyAsset"`
	}

	var result [32]byte
	err = wallet1.CallContractConstant(
		&result,
		"0x6454930EF2Bd86Ef40EC5fBBcb8a61F9B0F94512",
		`[
    {
      "inputs": [
        {
          "components": [
            {
              "internalType": "address",
              "name": "owner",
              "type": "address"
            },
            {
              "internalType": "uint256",
              "name": "salt",
              "type": "uint256"
            },
            {
              "components": [
                {
                  "internalType": "address",
                  "name": "token",
                  "type": "address"
                },
                {
                  "internalType": "uint256",
                  "name": "tokenId",
                  "type": "uint256"
                },
                {
                  "internalType": "enum ExchangeDomain.AssetType",
                  "name": "assetType",
                  "type": "uint8"
                }
              ],
              "internalType": "struct ExchangeDomain.Asset",
              "name": "sellAsset",
              "type": "tuple"
            },
            {
              "components": [
                {
                  "internalType": "address",
                  "name": "token",
                  "type": "address"
                },
                {
                  "internalType": "uint256",
                  "name": "tokenId",
                  "type": "uint256"
                },
                {
                  "internalType": "enum ExchangeDomain.AssetType",
                  "name": "assetType",
                  "type": "uint8"
                }
              ],
              "internalType": "struct ExchangeDomain.Asset",
              "name": "buyAsset",
              "type": "tuple"
            }
          ],
          "internalType": "struct ExchangeDomain.OrderKey",
          "name": "key",
          "type": "tuple"
        }
      ],
      "name": "getCompletedKey",
      "outputs": [
        {
          "internalType": "bytes32",
          "name": "",
          "type": "bytes32"
        }
      ],
      "stateMutability": "pure",
      "type": "function"
    }
  ]`,
		"getCompletedKey",
		&bind.CallOpts{
			Pending: false,
		},
		[]interface{}{
			OrderKey{
				Owner: common.HexToAddress("0xd9eb7d4dff36a2801d1ec42e75260b6e9e283e62"),
				Salt:  new(big.Int).SetInt64(72567257245612),
				SellAsset: Asset{
					Token:     common.HexToAddress("0x6454930EF2Bd86Ef40EC5fBBcb8a61F9B0F94512"),
					TokenId:   new(big.Int).SetInt64(0),
					AssetType: 0,
				},
				BuyAsset: Asset{
					Token:     common.HexToAddress("0x6454930EF2Bd86Ef40EC5fBBcb8a61F9B0F94512"),
					TokenId:   new(big.Int).SetInt64(0),
					AssetType: 0,
				},
			},
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "55a4de8a1e876840c443a4facf18771cd42c760351740efadfcf0a5634978748", hex.EncodeToString(result[:]))
}

func TestWallet_CallContractConstant2(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://http-mainnet.hecochain.com",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	type TokenBank struct {
		TokenAddr        common.Address `json:"tokenAddr"`
		PTokenAddr       common.Address
		IsOpen           bool     `json:"isOpen"`
		CanDeposit       bool     `json:"canDeposit"`
		CanWithdraw      bool     `json:"canWithdraw"`
		TotalVal         *big.Int `json:"totalVal"`
		TotalDebt        *big.Int `json:"totalDebt"`
		TotalDebtShare   *big.Int `json:"totalDebtShare"`
		TotalReserve     *big.Int `json:"totalReserve"`
		LastInterestTime *big.Int `json:"lastInterestTime"`
	}
	var tokenBank TokenBank
	err = wallet1.CallContractConstant(
		&tokenBank,
		"0xD42Ef222d33E3cB771DdA783f48885e15c9D5CeD",
		`[{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":true,"internalType":"address","name":"killer","type":"address"},{"indexed":false,"internalType":"uint256","name":"prize","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"left","type":"uint256"}],"name":"Liquidate","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"debt","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"back","type":"uint256"}],"name":"OpPosition","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"payable":true,"stateMutability":"payable","type":"fallback"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"string","name":"_symbol","type":"string"}],"name":"addToken","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"banks","outputs":[{"internalType":"address","name":"tokenAddr","type":"address"},{"internalType":"address","name":"pTokenAddr","type":"address"},{"internalType":"bool","name":"isOpen","type":"bool"},{"internalType":"bool","name":"canDeposit","type":"bool"},{"internalType":"bool","name":"canWithdraw","type":"bool"},{"internalType":"uint256","name":"totalVal","type":"uint256"},{"internalType":"uint256","name":"totalDebt","type":"uint256"},{"internalType":"uint256","name":"totalDebtShare","type":"uint256"},{"internalType":"uint256","name":"totalReserve","type":"uint256"},{"internalType":"uint256","name":"lastInterestTime","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"calInterest","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"currentPid","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"currentPos","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"debtShare","type":"uint256"}],"name":"debtShareToVal","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"debtVal","type":"uint256"}],"name":"debtValToShare","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"deposit","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"_symbol","type":"string"}],"name":"genPToken","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"isOwner","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"posId","type":"uint256"}],"name":"liquidate","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"posId","type":"uint256"},{"internalType":"uint256","name":"pid","type":"uint256"},{"internalType":"uint256","name":"borrow","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"opPosition","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"pid","type":"uint256"},{"internalType":"bool","name":"isOpen","type":"bool"},{"internalType":"bool","name":"canBorrow","type":"bool"},{"internalType":"address","name":"coinToken","type":"address"},{"internalType":"address","name":"currencyToken","type":"address"},{"internalType":"address","name":"borrowToken","type":"address"},{"internalType":"address","name":"goblin","type":"address"},{"internalType":"uint256","name":"minDebt","type":"uint256"},{"internalType":"uint256","name":"openFactor","type":"uint256"},{"internalType":"uint256","name":"liquidateFactor","type":"uint256"}],"name":"opProduction","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"posId","type":"uint256"}],"name":"positionInfo","outputs":[{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"positions","outputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"productionId","type":"uint256"},{"internalType":"uint256","name":"debtShare","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"productions","outputs":[{"internalType":"address","name":"coinToken","type":"address"},{"internalType":"address","name":"currencyToken","type":"address"},{"internalType":"address","name":"borrowToken","type":"address"},{"internalType":"bool","name":"isOpen","type":"bool"},{"internalType":"bool","name":"canBorrow","type":"bool"},{"internalType":"address","name":"goblin","type":"address"},{"internalType":"uint256","name":"minDebt","type":"uint256"},{"internalType":"uint256","name":"openFactor","type":"uint256"},{"internalType":"uint256","name":"liquidateFactor","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"renounceOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"totalToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"contract IBankConfig","name":"_config","type":"address"}],"name":"updateConfig","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"bool","name":"canDeposit","type":"bool"},{"internalType":"bool","name":"canWithdraw","type":"bool"}],"name":"updateToken","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"pAmount","type":"uint256"}],"name":"withdraw","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"value","type":"uint256"}],"name":"withdrawReserve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
		"banks",
		nil,
		[]interface{}{
			common.HexToAddress("0x25d2e80cb6b86881fd7e07dd263fb79f4abe033c"),
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x21AaF2b4973e8f437e45941b093b4149aB2513A6", tokenBank.PTokenAddr.String())
}

func TestWallet_DecodePayload(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	var a struct {
		AmountIn     *big.Int
		AmountOutMin *big.Int
		Path         []common.Address
		To           common.Address
		Deadline     *big.Int
	}
	abiStr := `[{"inputs":[{"internalType":"address","name":"_factory","type":"address"},{"internalType":"address","name":"_WETH","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"WETH","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"amountADesired","type":"uint256"},{"internalType":"uint256","name":"amountBDesired","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"addLiquidity","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"},{"internalType":"uint256","name":"liquidity","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amountTokenDesired","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"addLiquidityETH","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"},{"internalType":"uint256","name":"liquidity","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"factory","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"reserveIn","type":"uint256"},{"internalType":"uint256","name":"reserveOut","type":"uint256"}],"name":"getAmountIn","outputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"reserveIn","type":"uint256"},{"internalType":"uint256","name":"reserveOut","type":"uint256"}],"name":"getAmountOut","outputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsIn","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"reserveA","type":"uint256"},{"internalType":"uint256","name":"reserveB","type":"uint256"}],"name":"quote","outputs":[{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidity","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidityETH","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidityETHSupportingFeeOnTransferTokens","outputs":[{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityETHWithPermit","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens","outputs":[{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityWithPermit","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapETHForExactTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactETHForTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactETHForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETH","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETHSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"amountInMax","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapTokensForExactETH","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"amountInMax","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapTokensForExactTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`
	method, err := wallet1.DecodePayload(abiStr,
		&a,
		"0x38ed1739000000000000000000000000000000000000000000005f4a8c8375d15540000000000000000000000000000000000000000000000000548c0320c33861a5e3fe00000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000544fca5eef17d75a273955ba6fd16fe3c6e620aa000000000000000000000000000000000000000000000000000000005fe94d6e00000000000000000000000000000000000000000000000000000000000000020000000000000000000000006b175474e89094c44da98b954eedeac495271d0f0000000000000000000000003449fc1cd036255ba1eb19d65ff4ba2b8903a69a")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "swapExactTokensForTokens", method.Name)
	go_test_.Equal(t, "450000000000000000000000", a.AmountIn.String())
	go_test_.Equal(t, "399261554125997827548158", a.AmountOutMin.String())
	go_test_.Equal(t, 2, len(a.Path))
	go_test_.Equal(t, "0x6B175474E89094C44Da98b954EedeAC495271d0F", a.Path[0].String())
	go_test_.Equal(t, "0x3449FC1Cd036255BA1EB19d65fF4BA2b8903A69a", a.Path[1].String())
	go_test_.Equal(t, "0x544fcA5EEF17d75A273955bA6Fd16fe3c6E620Aa", a.To.String())
	go_test_.Equal(t, "1609125230", a.Deadline.String())

	method1, err := wallet1.DecodePayload(abiStr,
		&a,
		"0x38ed1739")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "swapExactTokensForTokens", method1.Name)
}

func TestWallet_TxsInPool(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	txs, err := wallet1.TxsInPool()
	fmt.Println(txs, err)
}

func TestWallet_SuggestGasPrice(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://sepolia.base.org",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	gasPrice, err := wallet1.SuggestGasPrice(0)
	fmt.Println(gasPrice.String())
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, true, len(gasPrice.String()) > 3)
}

func TestWallet_BuildTransferTx(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://rinkeby.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	_, err = wallet1.BuildTransferTx(
		"",
		"0x476fBB25d56B5dD4f1df03165498C403C4713069",
		&BuildTransferTxOpts{
			CallMethodOpts: CallMethodOpts{
				Value:        new(big.Int).SetUint64(1000000000000000),
				MaxFeePerGas: new(big.Int).SetUint64(100000000000),
			},
		},
	)
	go_test_.Equal(t, "invalid length, need 256 bits", err.Error())
}

func TestWallet_UnpackParams(t *testing.T) {
	wallet1 := NewWallet(&i_logger.DefaultLogger)
	defer wallet1.Close()
	datas, err := wallet1.UnpackParams([]abi.Type{TypeUint256}, "0x0000000000000000000000000000000000000000000000000000000000000001")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "1", datas[0].(*big.Int).String())
	strs, err := wallet1.UnpackParamsToStrs([]abi.Type{TypeUint256}, "0x0000000000000000000000000000000000000000000000000000000000000001")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "1", strs[0])

	datas, err = wallet1.UnpackParams([]abi.Type{TypeAddress}, "0x000000000000000000000000c054668c55ae734080642583246a74bbcd25d4c5")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0xc054668c55aE734080642583246A74bbcD25D4c5", datas[0].(common.Address).String())
	strs, err = wallet1.UnpackParamsToStrs([]abi.Type{TypeAddress}, "0x000000000000000000000000c054668c55ae734080642583246a74bbcd25d4c5")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0xc054668c55aE734080642583246A74bbcD25D4c5", strs[0])

	datas, err = wallet1.UnpackParams(
		[]abi.Type{TypeUint256, TypeUint256}, "0x00000000000000000000000000000000000000000000004e47868d5c301000000000000000000000000000000000000000000000000000000048df335bd24400")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "1444000000000000000000", datas[0].(*big.Int).String())
	go_test_.Equal(t, "20511610000000000", datas[1].(*big.Int).String())
	strs, err = wallet1.UnpackParamsToStrs(
		[]abi.Type{TypeUint256, TypeUint256}, "0x00000000000000000000000000000000000000000000004e47868d5c301000000000000000000000000000000000000000000000000000000048df335bd24400")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "1444000000000000000000", strs[0])
	go_test_.Equal(t, "20511610000000000", strs[1])
}

func TestWallet_PackParams(t *testing.T) {
	wallet1 := NewWallet(&i_logger.DefaultLogger)
	result, err := wallet1.PackParams(
		[]abi.Type{
			TypeUint256,
		},
		[]interface{}{
			new(big.Int).SetUint64(1),
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0000000000000000000000000000000000000000000000000000000000000001", result)

	result, err = wallet1.PackParamsFromStrs(
		[]abi.Type{
			TypeUint256,
		},
		[]string{
			"1",
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0000000000000000000000000000000000000000000000000000000000000001", result)

	data, err := hex.DecodeString("b872dd0e")
	go_test_.Equal(t, nil, err)
	result1, err := wallet1.PackParams(
		[]abi.Type{
			TypeAddress,
			TypeBytes,
		},
		[]interface{}{
			common.HexToAddress("0x7373c42502874C88954bDd6D50b53061F018422e"),
			data,
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0000000000000000000000007373c42502874c88954bdd6d50b53061f018422e00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004b872dd0e00000000000000000000000000000000000000000000000000000000", result1)

	result1, err = wallet1.PackParamsFromStrs(
		[]abi.Type{
			TypeAddress,
			TypeBytes,
		},
		[]string{
			"0x7373c42502874C88954bDd6D50b53061F018422e",
			"0xb872dd0e",
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0000000000000000000000007373c42502874c88954bdd6d50b53061f018422e00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004b872dd0e00000000000000000000000000000000000000000000000000000000", result1)
}

func TestWallet_TokenBalance(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://heconode.ifoobar.com",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()

	result, err := wallet1.TokenBalance("0xa71edc38d189767582c38a3145b5873052c3e47a", "0xe5fe1f8496095ab18f7e88eb13ed69f30a62d7a0")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, true, len(result.String()) > 0)
	fmt.Println(result.String())
}

func TestWallet_TransactionByHash(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://heconode.ifoobar.com",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()

	result, pending, err := wallet1.TransactionByHash("0x15f1e706a96aaf26c344240f18dfae5848329ffe91304e2ed778c9ad3f4b12c5")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, false, pending)
	go_test_.Equal(t, "0x9A5FBec6367a882d6B5F8CE2F267924d75e2d718", result.From.String())
}

func TestWallet_DeriveFromPath(t *testing.T) {
	wallet1 := NewWallet(&i_logger.DefaultLogger)
	defer wallet1.Close()

	result, err := wallet1.DeriveFromPath("308c18194c8345cb16b7a265439bc09c69e3166404951717fe50abe28bc9d19985cc1c06084290c2eba446d2626a1bf3bfb12ede5974653f756f26752475e8d8", "m/0/0/1")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x3a7d9f86cd69cc009a85d9280f6df5ef1ae8201d", strings.ToLower(result.Address))
	go_test_.Equal(t, "e9d3bd38744b8c6026f6bc719b86f725fe4298b654465abaa9a624ddfce8dc95", strings.ToLower(result.PrivateKey))
}

func TestWallet_EncodePayload(t *testing.T) {
	wallet1 := NewWallet(&i_logger.DefaultLogger)
	defer wallet1.Close()
	abiStr := `[{"inputs":[{"internalType":"address","name":"_factory","type":"address"},{"internalType":"address","name":"_WETH","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"WETH","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"amountADesired","type":"uint256"},{"internalType":"uint256","name":"amountBDesired","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"addLiquidity","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"},{"internalType":"uint256","name":"liquidity","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amountTokenDesired","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"addLiquidityETH","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"},{"internalType":"uint256","name":"liquidity","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"factory","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"reserveIn","type":"uint256"},{"internalType":"uint256","name":"reserveOut","type":"uint256"}],"name":"getAmountIn","outputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"reserveIn","type":"uint256"},{"internalType":"uint256","name":"reserveOut","type":"uint256"}],"name":"getAmountOut","outputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsIn","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"reserveA","type":"uint256"},{"internalType":"uint256","name":"reserveB","type":"uint256"}],"name":"quote","outputs":[{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidity","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidityETH","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidityETHSupportingFeeOnTransferTokens","outputs":[{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityETHWithPermit","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens","outputs":[{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityWithPermit","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapETHForExactTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactETHForTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactETHForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETH","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETHSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"amountInMax","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapTokensForExactETH","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"amountInMax","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapTokensForExactTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`

	result, err := wallet1.EncodePayload(
		abiStr,
		"swapExactTokensForTokens",
		[]interface{}{
			stringToBigInt("450000000000000000000000"),
			stringToBigInt("399261554125997827548158"),
			[]common.Address{
				common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),
				common.HexToAddress("0x3449FC1Cd036255BA1EB19d65fF4BA2b8903A69a"),
			},
			common.HexToAddress("0x544fcA5EEF17d75A273955bA6Fd16fe3c6E620Aa"),
			stringToBigInt("1609125230"),
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "38ed1739000000000000000000000000000000000000000000005f4a8c8375d15540000000000000000000000000000000000000000000000000548c0320c33861a5e3fe00000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000544fca5eef17d75a273955ba6fd16fe3c6e620aa000000000000000000000000000000000000000000000000000000005fe94d6e00000000000000000000000000000000000000000000000000000000000000020000000000000000000000006b175474e89094c44da98b954eedeac495271d0f0000000000000000000000003449fc1cd036255ba1eb19d65ff4ba2b8903a69a", result)
}

func stringToBigInt(a string) *big.Int {
	result, _ := new(big.Int).SetString(a, 10)
	return result
}

func TestWallet_BuildCallMethodTxWithPayload(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://sepolia.base.org",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	paramsStr, err := wallet1.PackParams(
		[]abi.Type{
			TypeAddress,
			TypeUint256,
		},
		[]interface{}{
			common.HexToAddress("0x2117210296c2993Cfb4c6790FEa1bEB3ECe8Ac06"),
			big.NewInt(1000000000000000000),
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0000000000000000000000002117210296c2993cfb4c6790fea1beb3ece8ac060000000000000000000000000000000000000000000000000de0b6b3a7640000", paramsStr)
	btr, err := wallet1.BuildCallMethodTxWithPayload(
		"4afc37894e7e4771eba8cb885b654eead3b78651d4db1e6af006d9e11f700f1f",
		"0x68422825055059bc548C213a50545614f655604e",
		"0x0b4c7e4d"+paramsStr,
		&CallMethodOpts{
			MaxFeePerGas: new(big.Int).SetUint64(1000000000),
			Nonce:        1,
			GasLimit:     5000000,
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x02f8b383014a3401843b9aca00843b9aca00834c4b409468422825055059bc548c213a50545614f655604e80b8440b4c7e4d0000000000000000000000002117210296c2993cfb4c6790fea1beb3ece8ac060000000000000000000000000000000000000000000000000de0b6b3a7640000c080a06b1f46073b56d747c245d27aa5cf765262e50caaa6ac68dd7d49c106d3bfa39da07135d180b765aa1e239c676bdbeeea422e88d1adb0094963ebdebd31de164042", btr.TxHex)

	tx, err := wallet1.DecodeTxHex(btr.TxHex)
	go_test_.Equal(t, nil, err)

	go_test_.Equal(t, btr.SignedTx.Hash().String(), tx.Hash().String())

}

func TestWallet_RandomMnemonic(t *testing.T) {
	wallet1 := NewWallet(&i_logger.DefaultLogger)
	defer wallet1.Close()
	result, err := wallet1.RandomMnemonic()
	go_test_.Equal(t, nil, err)
	fmt.Println(result)
}

func TestWallet_SeedHexByMnemonic(t *testing.T) {
	wallet1 := NewWallet(&i_logger.DefaultLogger)
	defer wallet1.Close()
	result := wallet1.SeedHexByMnemonic("go_test_", "go_test_")
	go_test_.Equal(t, "da2a48a1b9fbade07552281143814b3cd7ba4b53a7de5241439417b9bb540e229c45a30b0ce32174aaccc80072df7cbdff24f0c0ae327cd5170d1f276b890173", result)
}

func TestWallet_MaxUint256(t *testing.T) {
	wallet1 := NewWallet(&i_logger.DefaultLogger)
	defer wallet1.Close()
	result, err := wallet1.PackParams(
		[]abi.Type{
			TypeUint256,
		},
		[]interface{}{
			MaxUint256,
		},
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", result)
}

func TestRecoverSignerAddress(t *testing.T) {
	address, err := NewWallet(&i_logger.DefaultLogger).RecoverSignerAddressFromMsgHash("43aa0dee053b766817331315ab5440b144451f86d58f2d3b938e3de4dbca7ed8", "c2f57c5b9c187b089c492ba216cbb28e1d3d59aa8b62178b58d5442cbfc45cf40769070dfee43a8dd3f721d9a7a514f27fff3e0efbe6d7a41a6b27b8093f4dce1b")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0xf93B52193658335DBfe7b9138a0Da4CCEb6aF466", address.String())
}

func TestWallet_CallContractConstantWithPayload(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/7594e560416349f79c8ef6ff286d83fc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	payload, err := wallet1.EncodePayload(
		`[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_upgradedAddress","type":"address"}],"name":"deprecate","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"deprecated","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_evilUser","type":"address"}],"name":"addBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"upgradedAddress","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balances","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"maximumFee","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_maker","type":"address"}],"name":"getBlackListStatus","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowed","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"who","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newBasisPoints","type":"uint256"},{"name":"newMaxFee","type":"uint256"}],"name":"setParams","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"issue","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"amount","type":"uint256"}],"name":"redeem","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"basisPointsRate","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"isBlackListed","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_clearedUser","type":"address"}],"name":"removeBlackList","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"MAX_UINT","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_blackListedUser","type":"address"}],"name":"destroyBlackFunds","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[{"name":"_initialSupply","type":"uint256"},{"name":"_name","type":"string"},{"name":"_symbol","type":"string"},{"name":"_decimals","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"Redeem","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"newAddress","type":"address"}],"name":"Deprecate","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"feeBasisPoints","type":"uint256"},{"indexed":false,"name":"maxFee","type":"uint256"}],"name":"Params","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_blackListedUser","type":"address"},{"indexed":false,"name":"_balance","type":"uint256"}],"name":"DestroyedBlackFunds","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"AddedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"_user","type":"address"}],"name":"RemovedBlackList","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"}]`,
		"balanceOf",
		[]interface{}{
			common.HexToAddress("0xd9eb7d4dff36a2801d1ec42e75260b6e9e283e62"),
		},
	)
	go_test_.Equal(t, nil, err)
	out := new(big.Int)
	err = wallet1.CallContractConstantWithPayload(
		&out,
		"0xdac17f958d2ee523a2206206994597c13d831ec7",
		payload,
		abi.Arguments{
			abi.Argument{
				Name:    "",
				Type:    TypeUint256,
				Indexed: false,
			},
		},
		&bind.CallOpts{
			Pending: false,
		})
	go_test_.Equal(t, nil, err)
	fmt.Println(out.String())
}

func TestWallet_SignMsg(t *testing.T) {
	result, err := NewWallet(&i_logger.DefaultLogger).SignMsg("4afc37894e7e4771eba8cb885b654eead3b78651d4db1e6af006d9e11f700f1f", "hello")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "f315c40961c73b55f1e4cf4f2665c5cf70fda8f8f3a545e0788fe1f66e21f6d13d49ff33d300b9c5f9f943f095b5fc2838dbbb4d5820bc696fd974d284aa19751c", result)
}

func TestWallet_RecoverSignerAddress(t *testing.T) {
	address, err := NewWallet(&i_logger.DefaultLogger).RecoverSignerAddress("hello", "f315c40961c73b55f1e4cf4f2665c5cf70fda8f8f3a545e0788fe1f66e21f6d13d49ff33d300b9c5f9f943f095b5fc2838dbbb4d5820bc696fd974d284aa19751c")
	go_test_.Equal(t, nil, err)

	go_test_.Equal(t, "0xC3BF2dF684d91248b01278499184cC30C5bE45C3", address.String())
}

func TestWallet_SignHashForMsg(t *testing.T) {
	result, err := NewWallet(&i_logger.DefaultLogger).SignHashForMsg("hello")
	go_test_.Equal(t, nil, err)

	go_test_.Equal(t, "50b2c43fd39106bafbba0da34fc430e1f91e3c96ea2acee2bc34119f92b37750", result)
}

func TestWallet_MethodIdFromMethodStr(t *testing.T) {
	result := NewWallet(&i_logger.DefaultLogger).MethodIdFromMethodStr("cashChequeBeneficiary(address,uint256,bytes)")
	go_test_.Equal(t, "0d5f2659", result)
}

func TestWallet_SendEth(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://arb1.arbitrum.io/rpc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	txHash, err := wallet1.SendEth(
		"",
		"0xEA85c80805f36A65D96F6D360D02dFB3eBe18280",
		"0.1",
		nil,
	)
	go_test_.Equal(t, nil, err)

	fmt.Println(txHash)
}

func TestWallet_SendToken(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://arb1.arbitrum.io/rpc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	txHash, err := wallet1.SendToken("", "0x4C4A57dD7D4c21fc37882567Af756cbF4B332d7F", "0xEA85c80805f36A65D96F6D360D02dFB3eBe18280", go_decimal.Decimal.MustStart(1).MustShiftedBy(18).MustEndForBigInt(), &CallMethodOpts{
		GasLimit: 500000,
	})
	go_test_.Equal(t, nil, err)

	fmt.Println(txHash)
}

func TestWallet_ApproveAmount(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://arb1.arbitrum.io/rpc",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	amount, err := wallet1.ApprovedAmount(
		"0x4C4A57dD7D4c21fc37882567Af756cbF4B332d7F",
		"0xEA85c80805f36A65D96F6D360D02dFB3eBe18280",
		"0x16e71b13fe6079b4312063f7e81f76d165ad32ad",
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0", amount.String())
}

func TestWallet_Approve(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://arb1.arbitrum.io/rpc",
		WsUrl:  "",
	})

	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	txHash, err := wallet1.Approve(
		"",
		"0x4C4A57dD7D4c21fc37882567Af756cbF4B332d7F",
		"0x000000Eb4761239232363e02025194102b7Ef30a",
		nil,
		&CallMethodOpts{
			GasLimit: 400000,
		},
	)
	go_test_.Equal(t, nil, err)

	fmt.Println(txHash)
}

func TestWallet_CallContractConstant3(t *testing.T) {
	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://arb1.arbitrum.io/rpc",
		WsUrl:  "",
	})

	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	var quotaResult struct {
		AmountOut               *big.Int `json:"amountOut"`
		SqrtPriceX96After       *big.Int `json:"sqrtPriceX96After"`
		InitializedTicksCrossed uint32   `json:"initializedTicksCrossed"`
		GasEstimate             *big.Int `json:"gasEstimate"`
	}
	err = wallet1.CallContractConstant(
		&quotaResult,
		"0x61fFE014bA17989E743c5F6cB21bF9697530B21e",
		`[{
        "inputs":[
            {
                "components":[
                    {
                        "internalType":"address",
                        "name":"tokenIn",
                        "type":"address"
                    },
                    {
                        "internalType":"address",
                        "name":"tokenOut",
                        "type":"address"
                    },
                    {
                        "internalType":"uint256",
                        "name":"amountIn",
                        "type":"uint256"
                    },
                    {
                        "internalType":"uint24",
                        "name":"fee",
                        "type":"uint24"
                    },
                    {
                        "internalType":"uint160",
                        "name":"sqrtPriceLimitX96",
                        "type":"uint160"
                    }
                ],
                "internalType":"struct IQuoterV2.QuoteExactInputSingleParams",
                "name":"params",
                "type":"tuple"
            }
        ],
        "name":"quoteExactInputSingle",
        "outputs":[
            {
                "internalType":"uint256",
                "name":"amountOut",
                "type":"uint256"
            },
            {
                "internalType":"uint160",
                "name":"sqrtPriceX96After",
                "type":"uint160"
            },
            {
                "internalType":"uint32",
                "name":"initializedTicksCrossed",
                "type":"uint32"
            },
            {
                "internalType":"uint256",
                "name":"gasEstimate",
                "type":"uint256"
            }
        ],
        "stateMutability":"nonpayable",
        "type":"function"
    }]`,
		"quoteExactInputSingle",
		nil,
		[]interface{}{
			struct {
				TokenIn           common.Address
				TokenOut          common.Address
				AmountIn          *big.Int
				Fee               *big.Int
				SqrtPriceLimitX96 *big.Int
			}{
				common.HexToAddress("0x915EA4A94B61B138b568169122903Ed707A8E704"),
				common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
				go_decimal.Decimal.MustStart(1).MustShiftedBy(18).MustEndForBigInt(),
				new(big.Int).SetUint64(3000),
				new(big.Int).SetUint64(0),
			},
		},
	)
	go_test_.Equal(t, nil, err)
}

func TestWallet_IsContract(t *testing.T) {
	// wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
	// 	RpcUrl: "https://rpc.ankr.com/eth_goerli",
	// 	WsUrl:  "",
	// })
	// go_test_.Equal(t, nil, err)
	// defer wallet.Close()
	// isContract, err := wallet.IsContract("0x509Ee0d083DdF8AC028f2a56731412edD63223B9")
	// go_test_.Equal(t, nil, err)
	// go_test_.Equal(t, true, isContract)

	// isContract1, err := wallet.IsContract("0x2F62CEACb04eAbF8Fc53C195C5916DDDfa4BED02")
	// go_test_.Equal(t, nil, err)
	// go_test_.Equal(t, false, isContract1)

	wallet1, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.infura.io/v3/c747c4512b8b4d33ad265ea5803cbb30",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet1.Close()
	isContract, err := wallet1.IsContract("0xa96e7ba3772CE808DAB049E30ed146F77dA372a8")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, true, isContract)

}

func TestWallet_GetTokenDecimals(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://rpc.ankr.com/eth_goerli",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	defer wallet.Close()
	decimals, err := wallet.TokenDecimals("0x1319D23c2F7034F52Eb07399702B040bA278Ca49")
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, true, decimals == 18)
}

func TestWallet_ToTopicHash(t *testing.T) {
	wallet := NewWallet(&i_logger.DefaultLogger)
	hash, err := wallet.ToTopicHash(
		"OrderExecuted",
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x680f10f06595d3d707241f604672ec4b6ae50eb82728ec2f3c65f6789e897760", hash.String())
}

func TestWallet_ToTopicHashes(t *testing.T) {
	wallet := NewWallet(&i_logger.DefaultLogger)
	hashes, err := wallet.ToTopicHashes(
		"OrderExecuted",
		common.HexToAddress("0xc8ee91a54287db53897056e12d9819156d3822fb"),
	)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x680f10f06595d3d707241f604672ec4b6ae50eb82728ec2f3c65f6789e897760", hashes[0].String())
	go_test_.Equal(t, "0x000000000000000000000000c8ee91a54287db53897056e12d9819156d3822fb", hashes[1].String())
}

func TestWallet_UnpackLog(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.base.org",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	tr, err := wallet.TransactionReceiptByHash("0x866b3a1b7790b08e0642214399d030c1d30ab5c8b60f9dfdcabda75bab177112")
	go_test_.Equal(t, nil, err)
	var a struct {
		From  common.Address
		Value *big.Int
		To    common.Address
	}
	err = wallet.UnpackLog(&a, Erc20AbiStr, "Transfer", tr.Logs[1])
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, "0x09945b92893c2a00F8CCf96001D836f0499589E1", a.From.String())
	go_test_.Equal(t, "2021765178544007531148598", a.Value.String())
	go_test_.Equal(t, "0x1c5DDaEdbcc64f807c557275E5d6b8b57fF1A82b", a.To.String())

	fmt.Println(a)
}

// Filters logs correctly when both topic0 and logAddress match
func TestWallet_FilterLogs(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://mainnet.base.org",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	topic0 := "0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65"
	logAddress := "0x4200000000000000000000000000000000000006"
	tr, err := wallet.TransactionReceiptByHash("0x866b3a1b7790b08e0642214399d030c1d30ab5c8b60f9dfdcabda75bab177112")
	go_test_.Equal(t, nil, err)
	filteredLogs, err := wallet.FilterLogs(topic0, logAddress, tr.Logs)

	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, 1, len(filteredLogs))

	var a struct {
		Src common.Address
		Wad *big.Int
	}
	err = wallet.UnpackLog(&a, WETHAbiStr, "Withdrawal", filteredLogs[0])
	go_test_.Equal(t, nil, err)

	fmt.Println(a)
}

func TestWallet_PredictContractAddress(t *testing.T) {
	wallet := NewWallet(&i_logger.DefaultLogger)
	address := wallet.PredictContractAddress(
		"0xAfE2d7f5c9316EF143DF580c44e3CdB1eed30981",
		0,
	)
	go_test_.Equal(t, "0x3731Aa0F49A1866dF10f2639563f729D29D24e34", address)

	address = wallet.PredictContractAddress(
		"0xAfE2d7f5c9316EF143DF580c44e3CdB1eed30981",
		1,
	)
	go_test_.Equal(t, "0x8CA430C20be7452BDE527aFDd5d83b0FBC0AEF30", address)
}

func TestWallet_SendAllEth(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(
		&UrlParam{
			RpcUrl: "https://sepolia.base.org",
		},
	)
	go_test_.Equal(t, nil, err)

	_, err = wallet.SendAllEth(
		"",
		"0xD55B17ba6D269F94d75FfB3651d05529BEFD290A",
		"0.000001",
	)
	go_test_.Equal(t, nil, err)

}

func TestWallet_GasPriceNoDecimals(t *testing.T) {
	wallet, err := NewWallet(&i_logger.DefaultLogger).InitRemote(&UrlParam{
		RpcUrl: "https://bsc-dataseed4.ninicoin.io/",
		WsUrl:  "",
	})
	go_test_.Equal(t, nil, err)
	gasPrice, err := wallet.GasPriceNoDecimals()
	go_test_.Equal(t, nil, err)
	fmt.Println(gasPrice)
}

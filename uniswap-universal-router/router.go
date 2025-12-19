package uniswap_universal_router

// pancake 的 infinity 就是 v4，分为 CL(Concentrated Liquidity) 池和 Bin 池，CL 是主流
// uniswap_universal 可以执行 v2、v3、v4 的交易
// 所有 command 处理：https://github.com/Uniswap/universal-router/blob/main/contracts/base/Dispatcher.sol
// https://docs.uniswap.org/contracts/universal-router/technical-reference#reverting-command-example

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	go_coin_eth "github.com/pefish/go-coin-eth"
	i_logger "github.com/pefish/go-interface/i-logger"
)

type Router struct {
	wallet *go_coin_eth.Wallet
	logger i_logger.ILogger
}

func New(
	logger i_logger.ILogger,
	wallet *go_coin_eth.Wallet,
) *Router {
	return &Router{
		wallet: wallet,
		logger: logger,
	}
}

type SwapResultType struct {
	UserAddress                   common.Address
	InputToken                    common.Address
	InputTokenAmountWithDecimals  *big.Int
	OutputToken                   common.Address
	OutputTokenAmountWithDecimals *big.Int
	NetworkFee                    *big.Int
	TxId                          string
	BlockNumber                   uint64
	Liquidity                     *big.Int
	Fee                           *big.Int
	ProtocolFee                   *big.Int
}

package type_

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type SwapResultType struct {
	Type_                   string // buy/sell
	UserAddress             common.Address
	ETHAmountWithDecimals   *big.Int
	TokenAddress            common.Address
	TokenAmountWithDecimals *big.Int
	NetworkFee              *big.Int
	TxId                    string
	BlockNumber             uint64
}

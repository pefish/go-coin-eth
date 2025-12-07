package fourmeme_lib

import (
	"os"
	"path"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	go_coin_eth "github.com/pefish/go-coin-eth"
	i_logger "github.com/pefish/go-interface/i-logger"
	t_logger "github.com/pefish/go-interface/t-logger"
	go_logger "github.com/pefish/go-logger"
	go_test_ "github.com/pefish/go-test"
)

var wallet *go_coin_eth.Wallet

func init() {
	projectRoot, _ := go_test_.ProjectRoot()
	envMap, err := godotenv.Read(path.Join(projectRoot, ".env"))
	if err != nil {
		panic(err)
	}
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	wallet_, err := go_coin_eth.NewWallet(go_logger.NewLogger(t_logger.Level_DEBUG)).InitRemote(&go_coin_eth.UrlParam{
		RpcUrl: os.Getenv("NODE_HTTPS"),
		WsUrl:  os.Getenv("NODE_WSS"),
	})
	if err != nil {
		panic(err)
	}
	wallet = wallet_
}

func TestTokenInfo(t *testing.T) {
	tokenInfo, err := TokenInfo(
		wallet,
		common.HexToAddress("0x44440f83419de123d7d411187adb9962db017d03"),
	)
	go_test_.Equal(t, nil, err)
	spew.Dump(tokenInfo)
}

func TestTokenInfoByAPI(t *testing.T) {
	tokenInfo, err := TokenInfoByAPI(
		&i_logger.DefaultLogger,
		common.HexToAddress("0x444439030cbfdfb7e8db874734d56e612973e72b"),
	)
	go_test_.Equal(t, nil, err)
	spew.Dump(tokenInfo)
}

func TestGetReserveInfo(t *testing.T) {
	reserveInfo, err := GetReserveInfo(
		wallet,
		common.HexToAddress("0x925c8ab7a9a8a148e87cd7f1ec7ecc3625864444"),
	)
	go_test_.Equal(t, nil, err)
	spew.Dump(reserveInfo)
}

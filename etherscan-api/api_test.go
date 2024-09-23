package etherscan_api

import (
	"fmt"
	"testing"

	go_coin_eth "github.com/pefish/go-coin-eth"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

func TestEtherscanApiClient_ListTokenTx(t *testing.T) {
	client := NewEthscanApiClient(&i_logger.DefaultLogger, &OptionsType{
		Url: EthereumUrl,
	})
	_, err := client.ListTokenTx(&ListTokenTxParams{
		Address:    "0x000000000000000000000000000000000000dEaD",
		Page:       1,
		Offset:     50,
		StartBlock: 0,
		EndBlock:   99999999,
		Sort:       SortType_Desc,
	})
	go_test_.Equal(t, "Missing/Invalid API Key", err.Error())
}

func TestEtherscanApiClient_GetSourceCode(t *testing.T) {
	client := NewEthscanApiClient(&i_logger.DefaultLogger, &OptionsType{
		Url: BaseUrl,
	})
	result, err := client.GetSourceCode("0xd39A8680f50e46C9B99E642dD7d829D1b735509d")
	go_test_.Equal(t, nil, err)
	fmt.Println(result)
}

func TestWallet_FindLogs(t *testing.T) {
	client := NewEthscanApiClient(&i_logger.DefaultLogger, &OptionsType{
		Url:    EthereumUrl,
		ApiKey: "WDF9SBXFCPJKSBD9QEA59B2FDJIFMYTDGJ",
	})

	result, err := client.FindLogs(
		"0xD38Eca38703B9472Bf2f46dF56e6F7cCA03F60ed",
		20811900,
		20811950,
		[]string{
			"0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f",
			"0x0000000000000000000000007a250d5630b4cf539739df2c5dacb4c659f2488d",
		},
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(result[0].Data)
	go_coin_eth.NewWallet().UnpackParams()
	//go_test_.Equal(t, false, pending)
	//go_test_.Equal(t, "0x9A5FBec6367a882d6B5F8CE2F267924d75e2d718", result.From.String())
}

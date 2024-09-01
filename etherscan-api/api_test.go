package etherscan_api

import (
	"fmt"
	"testing"

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

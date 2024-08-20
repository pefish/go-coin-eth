package etherscan_api

import (
	"time"

	go_format "github.com/pefish/go-format"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

const API_URL = "https://api.etherscan.io/api"

type EtherscanApiClient struct {
	logger i_logger.ILogger
	apiKey string
}

func NewEthscanApiClient(logger i_logger.ILogger, apiKey string) *EtherscanApiClient {
	return &EtherscanApiClient{
		logger: logger,
		apiKey: apiKey,
	}
}

type SortType string

const (
	SortType_Asc  SortType = "asc"
	SortType_Desc SortType = "desc"
)

type ListTokenTxParams struct {
	ContractAddress string   `json:"contractaddress,omitempty"`
	Address         string   `json:"address,omitempty"`
	Page            int      `json:"page"`
	Offset          int      `json:"offset"`
	StartBlock      int      `json:"startblock"`
	EndBlock        int      `json:"endblock"`
	Sort            SortType `json:"sort"`
}

type ListTokenTxResult struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}

func (e *EtherscanApiClient) ListTokenTx(params *ListTokenTxParams) ([]ListTokenTxResult, error) {
	paramsMap := go_format.FormatInstance.StructToMap(params)

	paramsMap["module"] = "account"
	paramsMap["action"] = "tokentx"
	paramsMap["apikey"] = e.apiKey

	var httpResult struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Result  interface{} `json:"result"`
	}
	_, _, err := go_http.NewHttpRequester(
		go_http.WithTimeout(10*time.Second),
		go_http.WithLogger(e.logger),
	).GetForStruct(
		&go_http.RequestParams{
			Url:    API_URL,
			Params: paramsMap,
		},
		&httpResult,
	)
	if err != nil {
		return nil, err
	}
	if httpResult.Status != "1" {
		return nil, errors.New(httpResult.Result.(string))
	}

	results := make([]ListTokenTxResult, 0)
	for _, result := range httpResult.Result.([]interface{}) {
		var d ListTokenTxResult
		err := go_format.FormatInstance.MapToStruct(&d, result.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	return results, nil
}

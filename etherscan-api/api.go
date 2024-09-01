package etherscan_api

import (
	"time"

	go_format "github.com/pefish/go-format"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

const (
	EthereumUrl = "https://api.etherscan.io/api"
	BaseUrl     = "https://api.basescan.org/api"
)

type EtherscanApiClient struct {
	logger i_logger.ILogger
	apiKey string
	url    string
}

type OptionsType struct {
	Url    string
	ApiKey string
}

func NewEthscanApiClient(logger i_logger.ILogger, opts *OptionsType) *EtherscanApiClient {
	return &EtherscanApiClient{
		logger: logger,
		apiKey: opts.ApiKey,
		url:    opts.Url,
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
	BlockNumber       int    `json:"blockNumber,string"`
	TimeStamp         int    `json:"timeStamp,string"`
	Hash              string `json:"hash"`
	Nonce             int    `json:"nonce,string"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      int    `json:"tokenDecimal,string"`
	TransactionIndex  int    `json:"transactionIndex,string"`
	Gas               int    `json:"gas,string"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           int    `json:"gasUsed,string"`
	CumulativeGasUsed int    `json:"cumulativeGasUsed,string"`
	Input             string `json:"input"`
	Confirmations     int    `json:"confirmations,string"`
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
			Url:    e.url,
			Params: paramsMap,
		},
		&httpResult,
	)
	if err != nil {
		return nil, err
	}
	if httpResult.Status != "1" {
		return nil, errors.New(go_format.FormatInstance.ToString(httpResult.Result))
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

type GetSourceCodeResult struct {
	SourceCode           string `json:"SourceCode"`
	ABI                  string `json:"ABI"`
	ContractName         string `json:"ContractName"`
	CompilerVersion      string `json:"CompilerVersion"`
	OptimizationUsed     string `json:"OptimizationUsed"`
	Runs                 string `json:"Runs"`
	ConstructorArguments string `json:"ConstructorArguments"`
	EVMVersion           string `json:"EVMVersion"`
	Library              string `json:"Library"`
	LicenseType          string `json:"LicenseType"`
	Proxy                string `json:"Proxy"`
	Implementation       string `json:"Implementation"`
	SwarmSource          string `json:"SwarmSource"`
}

func (e *EtherscanApiClient) GetSourceCode(address string) (*GetSourceCodeResult, error) {
	paramsMap := make(map[string]interface{}, 0)

	paramsMap["module"] = "contract"
	paramsMap["action"] = "getsourcecode"
	paramsMap["apikey"] = e.apiKey
	paramsMap["address"] = address

	var httpResult struct {
		Status  string                `json:"status"`
		Message string                `json:"message"`
		Result  []GetSourceCodeResult `json:"result"`
	}
	_, _, err := go_http.NewHttpRequester(
		go_http.WithTimeout(10*time.Second),
		go_http.WithLogger(e.logger),
	).GetForStruct(
		&go_http.RequestParams{
			Url:    e.url,
			Params: paramsMap,
		},
		&httpResult,
	)
	if err != nil {
		return nil, err
	}
	if httpResult.Status != "1" {
		return nil, errors.New(go_format.FormatInstance.ToString(httpResult.Result))
	}

	return &httpResult.Result[0], nil
}

package etherscan_api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	go_coin_eth "github.com/pefish/go-coin-eth"
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
	logger  i_logger.ILogger
	apiKey  string
	url     string
	timeout time.Duration
}

type OptionsType struct {
	Url    string
	ApiKey string
}

func NewEthscanApiClient(logger i_logger.ILogger, opts *OptionsType) *EtherscanApiClient {
	return &EtherscanApiClient{
		logger:  logger,
		apiKey:  opts.ApiKey,
		url:     opts.Url,
		timeout: 10 * time.Second,
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
	paramsMap := go_format.StructToMap(params)

	paramsMap["module"] = "account"
	paramsMap["action"] = "tokentx"
	paramsMap["apikey"] = e.apiKey

	var httpResult struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Result  interface{} `json:"result"`
	}
	_, _, err := go_http.NewHttpRequester(
		go_http.WithTimeout(e.timeout),
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
		return nil, errors.New(go_format.ToString(httpResult.Result))
	}

	results := make([]ListTokenTxResult, 0)
	for _, result := range httpResult.Result.([]interface{}) {
		var d ListTokenTxResult
		err := go_format.MapToStruct(&d, result.(map[string]interface{}))
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
		go_http.WithTimeout(e.timeout),
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
		return nil, errors.New(go_format.ToString(httpResult.Result))
	}

	return &httpResult.Result[0], nil
}

type GetCreatorAndTxIdResult struct {
	ContractAddress string `json:"contractAddress"`
	ContractCreator string `json:"contractCreator"`
	TxId            string `json:"txHash"`
}

func (e *EtherscanApiClient) GetCreatorAndTxId(contractAddress string) (*GetCreatorAndTxIdResult, error) {
	paramsMap := make(map[string]interface{}, 0)

	paramsMap["module"] = "contract"
	paramsMap["action"] = "getcontractcreation"
	paramsMap["apikey"] = e.apiKey
	paramsMap["contractaddresses"] = contractAddress

	var httpResult struct {
		Status  string                    `json:"status"`
		Message string                    `json:"message"`
		Result  []GetCreatorAndTxIdResult `json:"result"`
	}
	_, _, err := go_http.NewHttpRequester(
		go_http.WithTimeout(e.timeout),
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
		return nil, errors.New(go_format.ToString(httpResult.Result))
	}

	return &httpResult.Result[0], nil
}

type FindLogsResult struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"` // 十六进制字符串
	Timestamp        string   `json:"timeStamp"`   // 十六进制字符串
	GasPrice         string   `json:"gasPrice"`    // 十六进制字符串
	GasUsed          string   `json:"gasUsed"`     // 十六进制字符串
	LogIndex         string   `json:"logIndex"`    // 十六进制字符串
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"` // 十六进制字符串
}

// 通过 scan api 查询 logs。最多只会返回开始的 1000 个结果，部分结果可能会被抛弃，所以要缩小范围查询
//
// 结果是按时间升序排列的
//
// topics []string 是 and 的关系
//
// page 从 1 开始的
//
// pageSize 最大为 1000
//
// fromBlock toBlock 是左闭右开
func (e *EtherscanApiClient) FindLogs(
	contractAddress string,
	fromBlock uint64,
	toBlock uint64,
	topics []string,
	page int,
	pageSize int,
) (results_ []FindLogsResult, err_ error) {
	params := map[string]interface{}{
		"module":    "logs",
		"action":    "getLogs",
		"fromBlock": fromBlock,
		"toBlock":   toBlock - 1,
		"apikey":    e.apiKey,
		"page":      page,
		"offset":    pageSize,
	}
	if contractAddress != "" {
		params["address"] = contractAddress
	}
	for i, str := range topics {
		if str == "" {
			continue
		}
		params[fmt.Sprintf("topic%d", i)] = str
		if i < len(topics)-1 {
			oprStr := fmt.Sprintf("topic%d_%d_opr", i, i+1)
			params[oprStr] = "and"
		}
	}

	var tempResult struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	_, resStr, err := go_http.NewHttpRequester(
		go_http.WithLogger(e.logger),
		go_http.WithTimeout(e.timeout),
	).GetForStruct(&go_http.RequestParams{
		Url:    e.url,
		Params: params,
	}, &tempResult)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if tempResult.Status != "1" && tempResult.Message != "No records found" {
		var result struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Result  string `json:"result"`
		}
		err = json.Unmarshal([]byte(resStr), &result)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		return nil, errors.New(result.Message + ". " + result.Result)
	}
	var result struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Result  []FindLogsResult `json:"result"`
	}
	err = json.Unmarshal([]byte(resStr), &result)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return result.Result, nil
}

type VerifySourceCodeParams struct {
	Code string
	Args struct {
		Types []abi.Type
		Strs  []string
	}
	ContractAddress   string
	ContractName      string
	CompilerVersion   string
	IsUseOptimization bool
	Runs              uint64
}

func (e *EtherscanApiClient) VerifySourceCode(params *VerifySourceCodeParams) (guid_ string, err_ error) {
	paramsMap := make(map[string]interface{}, 0)

	paramsMap["module"] = "contract"
	paramsMap["action"] = "verifysourcecode"
	paramsMap["apikey"] = e.apiKey

	paramsMap["codeformat"] = "solidity-single-file"
	paramsMap["sourceCode"] = params.Code
	if len(params.Args.Types) > 0 {
		r, err := go_coin_eth.NewWallet(e.logger).PackParamsFromStrs(params.Args.Types, params.Args.Strs)
		if err != nil {
			return "", err
		}
		paramsMap["constructorArguements"] = r
	}

	paramsMap["contractaddress"] = params.ContractAddress
	paramsMap["contractname"] = params.ContractName
	paramsMap["compilerversion"] = params.CompilerVersion
	if params.IsUseOptimization {
		paramsMap["optimizationUsed"] = 1
	} else {
		paramsMap["optimizationUsed"] = 0
	}
	paramsMap["runs"] = params.Runs

	var httpResult struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}
	_, _, err := go_http.NewHttpRequester(
		go_http.WithTimeout(e.timeout),
		go_http.WithLogger(e.logger),
	).PostForStruct(
		&go_http.RequestParams{
			Url:    e.url,
			Params: paramsMap,
		},
		&httpResult,
	)
	if err != nil {
		return "", err
	}
	if httpResult.Status != "1" {
		return "", errors.New(go_format.ToString(httpResult.Result))
	}

	return httpResult.Result, nil
}

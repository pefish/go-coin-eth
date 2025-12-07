package constant

import "github.com/ethereum/go-ethereum/common"

const TokenManagerABI = `[
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "base",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "offers",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "quote",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			}
		],
		"name": "LiquidityAdded",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "creator",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "requestId",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "name",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "symbol",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "totalSupply",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "launchTime",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "launchFee",
				"type": "uint256"
			}
		],
		"name": "TokenCreate",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "account",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "price",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "cost",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "fee",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "offers",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			}
		],
		"name": "TokenPurchase",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			}
		],
		"name": "TokenPurchase2",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "account",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "price",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "cost",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "fee",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "offers",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			}
		],
		"name": "TokenSale",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			}
		],
		"name": "TokenSale2",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "token",
				"type": "address"
			}
		],
		"name": "TradeStop",
		"type": "event"
	},
	{
		"inputs": [],
		"name": "STATUS_ADDING_LIQUIDITY",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "STATUS_COMPLETED",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "STATUS_HALT",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "STATUS_TRADING",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "_launchFee",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "_referralRewardKeeper",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "_referralRewardRate",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "_templateCount",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"name": "_templates",
		"outputs": [
			{
				"internalType": "address",
				"name": "quote",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "initialLiquidity",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxRaising",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "totalSupply",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxOffers",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minTradingFee",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "_tokenCount",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"name": "_tokenInfoEx1s",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "launchFee",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "pcFee",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "reserved2",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "reserved3",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "reserved4",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"name": "_tokenInfoExs",
		"outputs": [
			{
				"internalType": "address",
				"name": "creator",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "founder",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "reserves",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"name": "_tokenInfos",
		"outputs": [
			{
				"internalType": "address",
				"name": "base",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "quote",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "template",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "totalSupply",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxOffers",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxRaising",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "launchTime",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "offers",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "lastPrice",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "K",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "T",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "status",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"name": "_tokens",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "_tradingFeeRate",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "_tradingHalt",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxFunds",
				"type": "uint256"
			}
		],
		"name": "buyToken",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxFunds",
				"type": "uint256"
			}
		],
		"name": "buyToken",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxFunds",
				"type": "uint256"
			}
		],
		"name": "buyToken",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxFunds",
				"type": "uint256"
			}
		],
		"name": "buyToken",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minAmount",
				"type": "uint256"
			}
		],
		"name": "buyTokenAMAP",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minAmount",
				"type": "uint256"
			}
		],
		"name": "buyTokenAMAP",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minAmount",
				"type": "uint256"
			}
		],
		"name": "buyTokenAMAP",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minAmount",
				"type": "uint256"
			}
		],
		"name": "buyTokenAMAP",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "address",
						"name": "base",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "quote",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "template",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "totalSupply",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxOffers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxRaising",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "launchTime",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "offers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "funds",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "lastPrice",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "K",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "T",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "status",
						"type": "uint256"
					}
				],
				"internalType": "struct TokenManager3.TokenInfo",
				"name": "ti",
				"type": "tuple"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			}
		],
		"name": "calcBuyAmount",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "address",
						"name": "base",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "quote",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "template",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "totalSupply",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxOffers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxRaising",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "launchTime",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "offers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "funds",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "lastPrice",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "K",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "T",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "status",
						"type": "uint256"
					}
				],
				"internalType": "struct TokenManager3.TokenInfo",
				"name": "ti",
				"type": "tuple"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "calcBuyCost",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "maxRaising",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "totalSupply",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "offers",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "reserves",
				"type": "uint256"
			}
		],
		"name": "calcInitialPrice",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "address",
						"name": "base",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "quote",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "template",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "totalSupply",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxOffers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxRaising",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "launchTime",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "offers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "funds",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "lastPrice",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "K",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "T",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "status",
						"type": "uint256"
					}
				],
				"internalType": "struct TokenManager3.TokenInfo",
				"name": "ti",
				"type": "tuple"
			}
		],
		"name": "calcLastPrice",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "address",
						"name": "base",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "quote",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "template",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "totalSupply",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxOffers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxRaising",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "launchTime",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "offers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "funds",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "lastPrice",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "K",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "T",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "status",
						"type": "uint256"
					}
				],
				"internalType": "struct TokenManager3.TokenInfo",
				"name": "ti",
				"type": "tuple"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "calcSellCost",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "address",
						"name": "base",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "quote",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "template",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "totalSupply",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxOffers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxRaising",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "launchTime",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "offers",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "funds",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "lastPrice",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "K",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "T",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "status",
						"type": "uint256"
					}
				],
				"internalType": "struct TokenManager3.TokenInfo",
				"name": "ti",
				"type": "tuple"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			}
		],
		"name": "calcTradingFee",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "args",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "signature",
				"type": "bytes"
			}
		],
		"name": "createToken",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minFunds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "feeRate",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "feeRecipient",
				"type": "address"
			}
		],
		"name": "sellToken",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minFunds",
				"type": "uint256"
			}
		],
		"name": "sellToken",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "sellToken",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minFunds",
				"type": "uint256"
			}
		],
		"name": "sellToken",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "origin",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minFunds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "feeRate",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "feeRecipient",
				"type": "address"
			}
		],
		"name": "sellToken",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "sellToken",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

var TokenManagerAddress = common.HexToAddress("0x5c952063c7fc8610FFDB798152D69F0B9550762b")

const FourmemeToolABI = `[
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "internalType": "uint8",
          "name": "version",
          "type": "uint8"
        }
      ],
      "name": "Initialized",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "internalType": "address",
          "name": "previousOwner",
          "type": "address"
        },
        {
          "indexed": true,
          "internalType": "address",
          "name": "newOwner",
          "type": "address"
        }
      ],
      "name": "OwnershipTransferred",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "internalType": "address",
          "name": "",
          "type": "address"
        },
        {
          "indexed": false,
          "internalType": "uint256",
          "name": "",
          "type": "uint256"
        }
      ],
      "name": "Received",
      "type": "event"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_token",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "_amount",
          "type": "uint256"
        },
        {
          "internalType": "uint256",
          "name": "_maxCostBnbAmount",
          "type": "uint256"
        }
      ],
      "name": "buyPefish",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_tokenManager",
          "type": "address"
        }
      ],
      "name": "changeTokenManager",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_trader",
          "type": "address"
        }
      ],
      "name": "disableTrader",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_tokenManager",
          "type": "address"
        }
      ],
      "name": "initialize",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [],
      "name": "owner",
      "outputs": [
        {
          "internalType": "address",
          "name": "",
          "type": "address"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [],
      "name": "renounceOwnership",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_token",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "_amount",
          "type": "uint256"
        },
        {
          "internalType": "uint256",
          "name": "_minReceiveBnbAmount",
          "type": "uint256"
        }
      ],
      "name": "sellPefish",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_trader",
          "type": "address"
        }
      ],
      "name": "setTrader",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "newOwner",
          "type": "address"
        }
      ],
      "name": "transferOwnership",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_token",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "_amount",
          "type": "uint256"
        }
      ],
      "name": "withdraw",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "stateMutability": "payable",
      "type": "receive"
    }
  ]`

var FourmemeToolAddress = common.HexToAddress("0x9cB1AB52f5D6D72D8FFc26c64d63ab8a80fe18A5")

const TokenManagerHelperABI = `[
	{
		"inputs": [],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"inputs": [],
		"name": "PANCAKE_FACTORY",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "PANCAKE_V3_FACTORY",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "TM",
		"outputs": [
			{
				"internalType": "contract ITokenManager",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "TM2",
		"outputs": [
			{
				"internalType": "contract ITokenManager2",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "TOKEN_MANAGER",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "TOKEN_MANAGER_2",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "WETH",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "maxRaising",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "totalSupply",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "offers",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "reserves",
				"type": "uint256"
			}
		],
		"name": "calcInitialPrice",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			}
		],
		"name": "getTokenInfo",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "version",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "tokenManager",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "quote",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "lastPrice",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "tradingFeeRate",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "minTradingFee",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "launchTime",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "offers",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxOffers",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "maxFunds",
				"type": "uint256"
			},
			{
				"internalType": "bool",
				"name": "liquidityAdded",
				"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			}
		],
		"name": "tryBuy",
		"outputs": [
			{
				"internalType": "address",
				"name": "tokenManager",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "quote",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "estimatedAmount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "estimatedCost",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "estimatedFee",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "amountMsgValue",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "amountApproval",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "amountFunds",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "trySell",
		"outputs": [
			{
				"internalType": "address",
				"name": "tokenManager",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "quote",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "funds",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "fee",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

var TokenManagerHelperAddress = common.HexToAddress("0xF251F83e40a78868FcfA3FA4599Dad6494E46034")

var WBNBAddress = common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")

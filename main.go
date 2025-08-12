// 9c252a8b904af814c79125a785933136a98af747f9a4bf2ffe947ca8b857518b
// 32c45a8c24219f5dcb55551dac6e46fdd39115e2f93a386e4844525ea26375462b2b71d21ca19071f9c349f8aabaccb4ffd1ba4832534a2ba9f92329555cbc23
// 0x04c140254f3d0e502F570a6C671b35099fcEBE46
// 0x04c140254f3d0e502f570a6c671b35099fcebe46

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"

	// "net/rpc"
	"os"

	"bytes"

	"gitee.com/aqchain/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var rpcUrl = "https://mainnet.infura.io/v3/43904bc485474ac083f2d4149044f518"

// var rpcUrl = "https://rpc-mumbai.maticvigil.com"

var sender = "0x2a71c66851FC14dEeECF0193147c64EefFB37bd1"
var senderKey = "a2c1ed1d6fe46f298f68423adeefe87d43bfe2e423eca18a5bbf7b1aafd5d0a8"
var receiver = "0xAA13CcC67bd8348293cE5F6D918f9EECE5329363"

var usdtAddr = "0xdac17f958d2ee523a2206206994597c13d831ec7"

// var usdtAddr = "0x2922d45bb9600ea08b28f19e68a3cf8bf72a6402"

// var chainId int64 = 80001
var chainId int64 = 1

var amount int64 = 0

func generate() {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Println(hexutil.Encode(privateKeyBytes)[2:]) // 0xfad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println(hexutil.Encode(publicKeyBytes)[4:]) // 0x049a7df67f79246283fdc93af76d4f8cdd62c4886e8cd870944e817dd0b97934fdd7719d0810951e03418205868a5c1b40b192451367f28e0088dd75e15de40c05

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println(address) // 0x96216849c49358B10257cb55b28eA603c874b05E

	hash := sha3.NewKeccak256()
	hash.Write(publicKeyBytes[1:])
	fmt.Println(hexutil.Encode(hash.Sum(nil)[12:])) // 0x96216849c49358b10257cb55b28ea603c874b05e
}

func getEthBalance() {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	account := common.HexToAddress(sender)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}

	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	fmt.Printf("ETH Balance: %s\n", ethValue.String())
}

func getHistory() {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	account := common.HexToAddress(sender)

	// Get the block number of the latest block
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	latestBlockNumber := header.Number.Uint64()

	// Loop through the last 10,000 blocks to find transactions involving the given address
	var transactionHashes []common.Hash
	for blockNumber := latestBlockNumber; blockNumber > latestBlockNumber-600; blockNumber-- {
		block, err := client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNumber))
		if err != nil {
			log.Fatal(err)
		}

		for _, tx := range block.Transactions() {
			// Check if the transaction involves the specified address
			if tx.To() != nil && *tx.To() == account {
				transactionHashes = append(transactionHashes, tx.Hash())
			}
		}
	}

	for _, txHash := range transactionHashes {
		fmt.Println(txHash.Hex())
	}

}

func getUsdtBalance() {

	// USDT contract address on Ethereum mainnet
	usdtContractAddress := common.HexToAddress(usdtAddr)

	// USDT contract ABI
	usdtContractABI := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`

	// Connect to an Ethereum node
	client, err := rpc.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Create an Ethereum client
	ethClient := ethclient.NewClient(client)

	// Instantiate the USDT contract
	// Parse the USDT contract ABI
	contractAbi, err := abi.JSON(bytes.NewReader([]byte(usdtContractABI)))
	if err != nil {
		log.Fatal(err)
	}

	// Call the balanceOf function to get the USDT balance
	callData, err := contractAbi.Pack("balanceOf", common.HexToAddress(sender))
	if err != nil {
		log.Fatal(err)
	}

	msg := ethereum.CallMsg{
		To:   &usdtContractAddress,
		Data: callData,
	}

	result, err := ethClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Unpack the balance from the result
	var balance *big.Int
	err = contractAbi.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the balance to a big.Float
	usdtBalance := new(big.Float).SetInt(balance)
	usdtBalance = usdtBalance.Quo(usdtBalance, big.NewFloat(1e6))

	fmt.Printf("ETH USDT Balance: %s\n", usdtBalance.String())
}

func transaction() {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA(senderKey)
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.GasPrice = big.NewInt(10000000000) // Replace with your desired gas price
	auth.GasLimit = uint64(21000)           // Replace with your desired gas limit

	toAddress := common.HexToAddress(receiver)
	amount := big.NewInt(amount) // 0.01 MATIC (18 decimal places)

	data := []byte("")
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		log.Fatal(err)
	}
	tx := types.NewTransaction(nonce, toAddress, amount, auth.GasLimit, auth.GasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(chainId)), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())

}

func transactionUsdt() {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(receiver)

	decimalPlaces := 6 // For USDT// Calculate 10^decimalPlaces

	decimalMultiplier := int64(math.Pow(10, float64(decimalPlaces)))

	amount := big.NewInt(int64(amount * decimalMultiplier)) 

	// Convert the private key from hexadecimal to bytes
	privateKey, err := crypto.HexToECDSA(senderKey)
	if err != nil {
		log.Fatal(err)
	}

	// Get the sender's address from the private key
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// const usdtABI = `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]` // Replace with the full ABI of the USDT contract
	usdtABI := `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

	// Load the contract
	contractABI, err := abi.JSON(strings.NewReader(usdtABI))
	if err != nil {
		log.Fatal(err)
	}

	// Get the sender's account nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Create the transaction data to call the 'transfer' function of the USDT contract
	data, err := contractABI.Pack("transfer", toAddress, amount)
	if err != nil {
		log.Fatal(err)
	}

	// Create the transaction
	gasLimit := uint64(200000) // Set an appropriate gas limit for the transaction
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(usdtAddr), big.NewInt(0), gasLimit, gasPrice, data)

	// Sign the transaction with the private key
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(chainId)), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Send the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	// Print the transaction hash
	fmt.Println("Transaction Hash:", signedTx.Hash().Hex())

}

func main() {
	args := os.Args
	if len(args) > 1 {
		if args[1] == "generate" {
			generate()
		} else if args[1] == "balance" {
			if len(args) > 2 {
				sender = args[2]
			}
			if len(args) > 3 {
				rpcUrl = args[3]
			}
			getEthBalance()
		} else if args[1] == "balanceUsdt" {
			if len(args) > 2 {
				sender = args[2]
			}
			if len(args) > 3 {
				usdtAddr = args[3]
			}
			if len(args) > 4 {
				rpcUrl = args[4]
			}
			getUsdtBalance()
		} else if args[1] == "history" {
			if len(args) > 2 {
				sender = args[2]
			}
			if len(args) > 3 {
				rpcUrl = args[3]
			}
			getHistory()
		} else if args[1] == "transaction" {
			if len(args) > 2 {
				sender = args[2]
			}
			if len(args) > 3 {
				senderKey = args[3]
			}
			if len(args) > 4 {
				receiver = args[4]
			}
			if len(args) > 5 {
				num, err := strconv.ParseInt(args[5], 10, 64)
				if err != nil {
					// Handle the error if the string is not a valid int64 representation
					fmt.Println("Error:", err)
					return
				}
				amount = num
			}
			if len(args) > 6 {
				rpcUrl = args[6]
			}
			if len(args) > 7 {
				num, err := strconv.ParseInt(args[7], 10, 64)
				if err != nil {
					// Handle the error if the string is not a valid int64 representation
					fmt.Println("Error:", err)
					return
				}
				chainId = num
			}
			transaction()
		} else if args[1] == "transactionUsdt" {
			if len(args) > 2 {
				sender = args[2]
			}
			if len(args) > 3 {
				senderKey = args[3]
			}
			if len(args) > 4 {
				receiver = args[4]
			}
			if len(args) > 5 {
				num, err := strconv.ParseInt(args[5], 10, 64)
				if err != nil {
					// Handle the error if the string is not a valid int64 representation
					fmt.Println("Error:", err)
					return
				}
				amount = num
			}
			if len(args) > 6 {
				usdtAddr = args[6]
			}
			if len(args) > 7 {
				rpcUrl = args[7]
			}
			if len(args) > 8 {
				num, err := strconv.ParseInt(args[8], 10, 64)
				if err != nil {
					// Handle the error if the string is not a valid int64 representation
					fmt.Println("Error:", err)
					return
				}
				chainId = num
			}
			transactionUsdt()
		} else if args[1] == "help" {
			fmt.Println("balance sender rpcUrl")
			fmt.Println("balanceUsdt sender usdtAddr rpcUrl")
			fmt.Println("history sender rpcUrl")
			fmt.Println("transaction sender senderKey receiver amount rpcUrl chainId ")
			fmt.Println("transactionUsdt sender senderKey receiver amount usdtAddr rpcUrl chainId ")
		} else if args[1] == "info" {
			fmt.Println("Sender: ", sender)
			fmt.Println("senderKey: ", senderKey)
			fmt.Println("usdtAddr: ", usdtAddr)
			fmt.Println("rpcUrl: ", rpcUrl)
			fmt.Println("chainId: ", chainId)
			fmt.Println("amount: ", amount)
			fmt.Println("receiver: ", receiver)
		} else {
			fmt.Println("Wrong argments provided")
		}
	} else {
		fmt.Println("No argments provided")
	}
}

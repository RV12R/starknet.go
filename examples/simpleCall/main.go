package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
)

var (
	name                string = "integration"
	someMainnetContract string = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	contractMethod      string = "balanceOf"
)

// main entry point of the program.
//
// It initializes the environment and establishes a connection with the client.
// It then makes a contract call and prints the response.
//
// Parameters:
//
//	none
//
// Returns:
//
//	none
func main() {
	fmt.Println("Starting simpeCall example")
	godotenv.Load(fmt.Sprintf(".env.%s", name))
	base := os.Getenv("INTEGRATION_BASE")
	c, err := ethrpc.DialContext(context.Background(), base)
	if err != nil {
		fmt.Println("Failed to connect to the client, did you specify the url in the .env.mainnet?")
		panic(err)
	}
	clientv02 := rpc.NewProvider(c)
	fmt.Println("Established connection with the client")

	contractAddress, err := utils.HexToFelt(someMainnetContract)
	if err != nil {
		panic(err)
	}

	// 0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e
	// 0x39180766dd93d979cd03e6ba93d3210fb7b7e56e01020942157e387a3613040	// Takes a good 10min to bridge..
	// 0xffbc2e41bad21b80eea2590c38287718aced9db4c66ae693b5ed24320f6de7
	// 0xc6e9dddcab4b464f2edc1cb4ef14163317a58ae64fd24ffe280baaf8eb5691

	// 0x13d46bc231a47e29ee1e3c7fab338e6b13c935e002f2be57f7318e88cbb872e
	// 0x397ffc31dde066cab3f4a9f6315f5641620b2ee0622c042b91ade554ca4bff5
	// 0x1fee7cfefa90a2878826212bb46bccddfb2473f306b7b17e4791bd0a486d4c6
	//
	myAddr, _ := new(felt.Felt).SetString("0x6764832432bbe2e35a45bf7d5771e4551fcf760fa803d30afd3ec55140c57a7")

	// Make read contract call
	tx := rpc.FunctionCall{
		ContractAddress:    contractAddress,
		EntryPointSelector: utils.GetSelectorFromNameFelt(contractMethod),
		Calldata:           []*felt.Felt{myAddr},
	}

	fmt.Println("Making Call() request", myAddr)
	callResp, err := clientv02.Call(context.Background(), tx, rpc.BlockID{Tag: "latest"})
	if err != nil {
		fmt.Println("+--=-")
		panic(err.Error())
	}
	qwe := new(big.Int)
	callResp[0].BigInt(qwe)
	fmt.Println(qwe)
	fmt.Println(fmt.Sprintf("Response to %s():%s ", contractMethod, callResp[0]))
}

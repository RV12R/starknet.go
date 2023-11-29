package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
)

// NOTE : Please add in your keys only for testing purposes, in case of a leak you would potentially lose your funds.
var (
	name           string = "testnet"                                                            //env."name"
	account_addr   string = "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e" //Replace it with your account address
	privateKey     string = "0x043b7fe9d91942c98cd5fd37579bd99ec74f879c4c79d886633eecae9dad35fa" //Replace it with your account private key
	public_key     string = "0x049f060d2dffd3bf6f2c103b710baf519530df44529045f92c3903097e8d861f" //Replace it with your account public key
	someContract   string = "0x041a78e741e5af2fec34b695679bc6891742439f7afb8484ecd7766661ad02bf" // UDC
	contractMethod string = "deployContract"                                                     //Replace it with the function name that you want to invoke
)

func main() {
	// Loading the env
	godotenv.Load(fmt.Sprintf(".env.%s", name))
	base := os.Getenv("INTEGRATION_BASE") //please modify the .env.testnet and replace the INTEGRATION_BASE with a starknet goerli RPC.
	fmt.Println("Starting simpleInvoke example")

	// Initialising the connection
	c, err := ethrpc.DialContext(context.Background(), base)
	if err != nil {
		fmt.Println("Failed to connect to the client, did you specify the url in the .env.testnet?")
		panic(err)
	}

	// Initialising the provider
	clientv02 := rpc.NewProvider(c)

	// Here we are converting the account address to felt
	account_address, err := utils.HexToFelt(account_addr)
	if err != nil {
		panic(err.Error())
	}
	// Initializing the account memkeyStore
	ks := account.NewMemKeystore()
	fakePrivKeyBI, ok := new(big.Int).SetString(privateKey, 0)
	if !ok {
		panic(err.Error())
	}
	ks.Put(public_key, fakePrivKeyBI)

	fmt.Println("Established connection with the client")

	// Here we are setting the maxFee
	maxfee, err := utils.HexToFelt("0x9184e72a000")
	if err != nil {
		panic(err.Error())
	}

	// Initializing the account
	accnt, err := account.NewAccount(clientv02, account_address, public_key, ks)
	if err != nil {
		panic(err.Error())
	}

	// Getting the nonce from the account
	// nonce, err := accnt.Nonce(context.Background(), rpc.BlockID{Tag: "latest"}, accnt.AccountAddress)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// Building the InvokeTx struct
	InvokeTx := rpc.InvokeTxnV1{
		MaxFee:        maxfee,
		Version:       rpc.TransactionV1,
		Nonce:         new(felt.Felt).SetUint64(0),
		Type:          rpc.TransactionType_Invoke,
		SenderAddress: accnt.AccountAddress,
	}
	fmt.Println("=====")
	// Converting the contractAddress from hex to felt
	contractAddress, err := utils.HexToFelt(someContract)
	if err != nil {
		panic(err.Error())
	}

	// recipient, _ := utils.HexToFelt("0x39180766dd93d979cd03e6ba93d3210fb7b7e56e01020942157e387a3613040")
	// amount, _ := utils.HexToFelt("0x2a303fe4b530000")
	classHash, _ := utils.HexToFelt("0x2a303fe4b530000")
	unique, _ := utils.HexToFelt("0x2a303fe4b530000")
	salt := new(felt.Felt).SetUint64(0)
	callata, _ := utils.HexToFelt(account_addr)

	// Building the functionCall struct, where :
	FnCall := rpc.FunctionCall{
		ContractAddress:    contractAddress,                               //contractAddress is the contract that we want to call
		EntryPointSelector: utils.GetSelectorFromNameFelt(contractMethod), //this is the function that we want to call
		Calldata:           []*felt.Felt{recipient, amount},
	}
	fmt.Println("=====")
	// Mentioning the contract version
	CairoContractVersion := 0

	// Building the Calldata with the help of FmtCalldata where we pass in the FnCall struct along with the Cairo version
	InvokeTx.Calldata, err = accnt.FmtCalldata([]rpc.FunctionCall{FnCall}, CairoContractVersion)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("=====")
	// Signing of the transaction that is done by the account
	err = accnt.SignInvokeTransaction(context.Background(), &InvokeTx)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("=====")
	qwe, _ := json.MarshalIndent(InvokeTx, "", "")
	fmt.Println(string(qwe))

	// After the signing we finally call the AddInvokeTransaction in order to invoke the contract function
	// resp, err := accnt.AddInvokeTransaction(context.Background(), InvokeTx)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// // This returns us with the transaction hash
	// fmt.Println("Transaction hash response : ", resp.TransactionHash)

}

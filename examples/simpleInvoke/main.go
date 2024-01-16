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
	name                  string = "testnet"                                                            //env."name"
	account_addr          string = "0x04c34f000a86f5e5fbfececdcae233209fa032fdd8e84e6dac7ce7ef4d858c73" //Replace it with your account address
	account_cairo_version        = 0                                                                    //Replace  with the cairo version of your account
	privateKey            string = "0x05e6c9dbf900e2106c2a1977f632dfbaa0354983e98d3401c4817627a025b2e8" //Replace it with your account private key
	public_key            string = "0x5b6a747863f1efa0047c5917b20ca6d802ca55db775780114fb0965ea2035b9"  //Replace it with your account public key
	someContract          string = "0x4c1337d55351eac9a0b74f3b8f0d3928e2bb781e5084686a892e66d49d510d"   //Replace it with the contract that you want to invoke
	contractMethod        string = "increase_value"                                                     //Replace it with the function name that you want to invoke
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

	// Initializing the account
	accnt, err := account.NewAccount(clientv02, account_address, public_key, ks, account_cairo_version)
	if err != nil {
		panic(err.Error())
	}

	// Getting the nonce from the account
	nonce, err := accnt.Nonce(context.Background(), rpc.BlockID{Tag: "latest"}, accnt.AccountAddress)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Nonce : ", nonce)

	//// Transaction 1 (eg 0x468000b61988138f75e4ffb4b21c5553c681d512fb14d12837b737f19de84bf)
	nonceTxn1 := new(felt.Felt).Add(nonce, new(felt.Felt).SetUint64(0))
	fmt.Println("nonceTxn1 : ", nonceTxn1)

	txn1, err := prepareTx(accnt, clientv02, account_address, ks, nonceTxn1)
	if err != nil {
		panic(err.Error())
	}

	txn1Hash, _ := accnt.TransactionHashInvoke(txn1)
	fmt.Println("txn1Hash : ", txn1Hash)

	resp1, err := accnt.AddInvokeTransaction(context.Background(), txn1)
	fmt.Println("resp1, err", resp1, err)

	//// Transaction 2 (eg 0x7bfd7e6870ee114097285f2ea5d9797a0fe3b168c36d5417b68d41218c4b5a6)
	nonceTxn2 := new(felt.Felt).Add(nonce, new(felt.Felt).SetUint64(1))
	fmt.Println("nonceTxn2 : ", nonceTxn2)
	txn2, err2 := prepareTx(accnt, clientv02, account_address, ks, nonceTxn2)
	if err2 != nil {
		panic(err.Error())
	}
	txn2Hash, _ := accnt.TransactionHashInvoke(txn2)
	fmt.Println("txn2Hash : ", txn2Hash)

	resp2, err := accnt.AddInvokeTransaction(context.Background(), txn2)
	fmt.Println("resp2, err", resp2, err)

}

func prepareTx(accnt *account.Account, clientv02 *rpc.Provider, account_address *felt.Felt, ks account.Keystore, nonce *felt.Felt) (rpc.InvokeTxnV1, error) {
	// Here we are setting the maxFee
	maxfee, err := utils.HexToFelt("0x9184e72a101")
	if err != nil {
		panic(err.Error())
	}
	// Building the InvokeTx struct
	InvokeTx := rpc.InvokeTxnV1{
		MaxFee:        maxfee,
		Version:       rpc.TransactionV1,
		Nonce:         nonce,
		Type:          rpc.TransactionType_Invoke,
		SenderAddress: accnt.AccountAddress,
	}

	// Converting the contractAddress from hex to felt
	contractAddress, err := utils.HexToFelt(someContract)
	if err != nil {
		panic(err.Error())
	}

	// Building the functionCall struct, where :
	FnCall := rpc.FunctionCall{
		ContractAddress:    contractAddress,                               //contractAddress is the contract that we want to call
		EntryPointSelector: utils.GetSelectorFromNameFelt(contractMethod), //this is the function that we want to call
	}

	// Building the Calldata with the help of FmtCalldata where we pass in the FnCall struct along with the Cairo version
	InvokeTx.Calldata, err = accnt.FmtCalldata([]rpc.FunctionCall{FnCall})
	if err != nil {
		panic(err.Error())
	}

	// Signing of the transaction that is done by the account
	err = accnt.SignInvokeTransaction(context.Background(), &InvokeTx)
	if err != nil {
		panic(err.Error())
	}
	return InvokeTx, nil
}

func printTxn(txn any) {
	qwe, _ := json.MarshalIndent(txn, "", "")
	fmt.Println(string(qwe))
}

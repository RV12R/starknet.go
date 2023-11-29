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

var (
	network string = "integration"
	// Note the difference between classHash and contract address
	predeployedClassHash = "0x25ec026985a3bf9d0cc1fe17326b245dfdc3ff89b8fde106542a3ea56c5a918" //"0x2794ce20e5f2ff0d40e632cb53845b9f4e526ebd8471983f7dbd355b721d5a"
)

func main() {
	// Initialise the client.
	godotenv.Load(fmt.Sprintf(".env.%s", network))
	base := os.Getenv("INTEGRATION_BASE")
	c, err := ethrpc.DialContext(context.Background(), base)
	if err != nil {
		panic("You need to specify the testnet url in .env.testnet")
	}
	clientv02 := rpc.NewProvider(c)

	// Get random keys for test purposes
	AccountAddress, _ := utils.HexToFelt("0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e")

	ks := account.NewMemKeystore()
	PubKey, _ := utils.HexToFelt("0x049f060d2dffd3bf6f2c103b710baf519530df44529045f92c3903097e8d861f")
	PrivKey, _ := utils.HexToFelt("0x043b7fe9d91942c98cd5fd37579bd99ec74f879c4c79d886633eecae9dad35fa")
	fakePrivKeyBI, ok := new(big.Int).SetString(PrivKey.String(), 0)
	if !ok {
		panic("Error setting up account key store")
	}
	ks.Put(PubKey.String(), fakePrivKeyBI)

	// Set up account
	acnt, err := account.NewAccount(clientv02, AccountAddress, PubKey.String(), ks)
	if err != nil {
		panic(err)
	}

	classHash, err := utils.HexToFelt(predeployedClassHash)
	if err != nil {
		panic(err)
	}

	impl, _ := new(felt.Felt).SetString("0x33434ad846cdd5f23eb73ff09fe6fddd568284a0fb7d1be20ee482f044dabe2")
	sellector, _ := new(felt.Felt).SetString("0x79dc0da7c54b95f10aa182ad0a46400db63156920adb65eca2654c0945a463")

	// Create transaction data
	tx := rpc.DeployAccountTxn{
		Nonce:               &felt.Zero, // Contract accounts start with nonce zero.
		MaxFee:              new(felt.Felt).SetUint64(123000000000000),
		Type:                rpc.TransactionType_DeployAccount,
		Version:             rpc.TransactionV1,
		Signature:           []*felt.Felt{},
		ClassHash:           classHash,
		ContractAddressSalt: PubKey,
		ConstructorCalldata: []*felt.Felt{
			impl,
			sellector,
			new(felt.Felt).SetUint64(2),
			PubKey,
			new(felt.Felt).SetUint64(0)},
	}
	fmt.Println("tx", tx.ClassHash)
	precomputedAddress, err := acnt.PrecomputeAddress(&felt.Zero, PubKey, classHash, tx.ConstructorCalldata)
	fmt.Println("precomputedAddress:", precomputedAddress)

	// At this point you need to add funds to precomputed address to use it.
	var input string

	fmt.Println("The `precomputedAddress` account needs to have enough ETH to perform a transaction.")
	fmt.Println("Use the starknet faucet to send ETH to your `precomputedAddress`")
	fmt.Println("When your account has been funded by the faucet, press any key, then `enter` to continue : ")
	fmt.Scan(&input)

	// Sign the transaction
	tx.Signature, err = acnt.SignDeployAccountTransaction(context.Background(), tx, precomputedAddress)
	if err != nil {
		panic(err)
	}

	qwe, _ := json.MarshalIndent(tx, "", "")
	fmt.Println(string(qwe))

	// Send transaction to the network
	resp, err := acnt.AddDeployAccountTransaction(context.Background(), rpc.BroadcastDeployAccountTxn{DeployAccountTxn: tx})
	if err != nil {
		panic(fmt.Sprint("Error returned from AddDeployAccountTransaction: ", err))
	}
	fmt.Println("AddDeployAccountTransaction response:", resp)
}

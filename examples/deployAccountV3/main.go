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
	// https://external.integration.starknet.io/feeder_gateway/get_transaction?transactionHash=0x29fd7881f14380842414cdfdd8d6c0b1f2174f8916edcfeb1ede1eb26ac3ef0
	predeployedClassHash        = "0x2338634f11772ea342365abd5be9d9dc8a6f44f159ad782fdebd3db5d969738"
	accountAddress       string = "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"
	pubKey               string = "0x049f060d2dffd3bf6f2c103b710baf519530df44529045f92c3903097e8d861f"
	privKey              string = "0x043b7fe9d91942c98cd5fd37579bd99ec74f879c4c79d886633eecae9dad35fa"
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

	AccountAddress, _ := utils.HexToFelt(accountAddress)
	ks := account.NewMemKeystore()
	PubKey, _ := utils.HexToFelt(pubKey)
	PrivKey, _ := utils.HexToFelt(privKey)
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

	// https://goerli.voyager.online/tx/0x559d0c57b7651b7f8e1c25cc92ff2d0567bcecb62a36c62d7a3f3fb9319e140
	// guardian, _ := new(felt.Felt).SetString("0x0")

	// Create transaction data
	tx := rpc.DeployAccountTxnV3{
		Nonce:               &felt.Zero, // Contract accounts start with nonce zero.
		Type:                rpc.TransactionType_DeployAccount,
		Version:             rpc.TransactionV3,
		Signature:           []*felt.Felt{},
		ClassHash:           classHash,
		ContractAddressSalt: PubKey,
		ConstructorCalldata: []*felt.Felt{
			PubKey},
		ResourceBounds: rpc.ResourceBoundsMapping{
			L1Gas: rpc.ResourceBounds{
				MaxAmount:       new(felt.Felt).SetUint64(4328000220728),
				MaxPricePerUnit: new(felt.Felt).SetUint64(4328000220728),
			},
			L2Gas: rpc.ResourceBounds{
				MaxAmount:       new(felt.Felt).SetUint64(0),
				MaxPricePerUnit: new(felt.Felt).SetUint64(0),
			},
		},
		Tip:           new(felt.Felt).SetUint64(0),
		PayMasterData: []*felt.Felt{},
		NonceDataMode: rpc.DAModeL1,
		FeeMode:       rpc.DAModeL1,
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
	tx.Signature, err = acnt.SignDeployAccountTransactionv3(context.Background(), tx, precomputedAddress)
	if err != nil {
		panic(err)
	}

	qwe, _ := json.MarshalIndent(tx, "", "")
	fmt.Println(string(qwe))

	// Send transaction to the network
	resp, err := acnt.AddDeployAccountTransaction(context.Background(), rpc.BroadcastDeployAccountTxnV3{DeployAccountTxnV3: tx})
	if err != nil {
		panic(fmt.Sprint("Error returned from AddDeployAccountTransaction: ", err))
	}
	fmt.Println("AddDeployAccountTransaction response:", resp)
}

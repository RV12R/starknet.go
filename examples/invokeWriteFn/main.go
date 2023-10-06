package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	starknetgo "github.com/NethermindEth/starknet.go"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

const (
	endpoint        = ""
	accountAddres   = ""
	pubKey          = "" // Note: do not write keys here for prduction. This is for testing only!
	privKey         = "" // Note: do not write keys here for prduction. This is for testing only!
	contractAddress = "0x4c1337d55351eac9a0b74f3b8f0d3928e2bb781e5084686a892e66d49d510d"
	contractMethod  = "increase_value"
)

func main() {
	// Set provider
	client, err := rpc.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	provider := rpc.NewProvider(client)

	// Set up ks
	ks := starknetgo.NewMemKeystore()
	fakePrivKeyBI, ok := new(big.Int).SetString(pubKey, 0)
	if !ok {
		panic("Error setting pubKey")
	}
	ks.Put(pubKey, fakePrivKeyBI)

	acntAddres, err := new(felt.Felt).SetString(accountAddres)
	if err != nil {
		panic(err)
	}

	// Set up account
	acnt, err := account.NewAccount(provider, acntAddres, pubKey, ks)
	if err != nil {
		panic(err)
	}

	// Get nonce
	nonce, err := acnt.Nonce(context.Background(), rpc.BlockID{Tag: "latest"}, acnt.AccountAddress)
	if err != nil {
		panic(err)
	}
	nonceFelt, err := new(felt.Felt).SetString(*nonce)
	if err != nil {
		panic(err)
	}

	// Create transaction data
	InvokeTx := rpc.InvokeTxnV1{
		MaxFee:        new(felt.Felt).SetUint64(123456),
		Version:       rpc.TransactionV1,
		Nonce:         nonceFelt,
		Type:          rpc.TransactionType_Invoke,
		SenderAddress: acntAddres,
	}

	contractAddress, err := new(felt.Felt).SetString(contractAddress)
	if err != nil {
		panic(err)
	}

	FnCall := rpc.FunctionCall{
		ContractAddress:    contractAddress,
		EntryPointSelector: utils.GetSelectorFromNameFelt(contractMethod),
		Calldata:           []*felt.Felt{},
	}

	// Build transaction
	err = acnt.BuildInvokeTx(context.Background(), &InvokeTx, &[]rpc.FunctionCall{FnCall})
	if err != nil {
		panic(err)
	}

	// Submit transaction
	resp, err := acnt.AddInvokeTransaction(context.Background(), InvokeTx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Response from AddInvokeTransaction:", resp)
}

package account_test

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/NethermindEth/juno/core/felt"
	starknetgo "github.com/NethermindEth/starknet.go"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/mocks"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/test"
	"github.com/NethermindEth/starknet.go/types"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/golang/mock/gomock"
	"github.com/test-go/testify/require"
)

var (
	// set the environment for the test, default: mock
	testEnv = "devnet"
	base    = ""
)

// TestMain is used to trigger the tests and, in that case, check for the environment to use.
func TestMain(m *testing.M) {
	flag.StringVar(&testEnv, "env", "devnet", "set the test environment")
	flag.Parse()
	godotenv.Load(fmt.Sprintf(".env.%s", testEnv), ".env")
	base = os.Getenv("INTEGRATION_BASE")
	if base == "" && testEnv != "mock" {
		panic(fmt.Sprint("Failed to set INTEGRATION_BASE for ", testEnv))
	}
	os.Exit(m.Run())
}

func TestTransactionHash(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockRpcProvider := mocks.NewMockRpcProvider(mockCtrl)

	type testSetType struct {
		ExpectedHash   *felt.Felt
		SetKS          bool
		AccountAddress *felt.Felt
		PubKey         string
		PrivKey        *felt.Felt
		ChainID        string
		FnCall         rpc.FunctionCall
		TxDetails      rpc.TxDetails
	}
	testSet := map[string][]testSetType{
		"mock": {
			{
				// https://goerli.voyager.online/tx/0x73cf79c4bfa0c7a41f473c07e1be5ac25faa7c2fdf9edcbd12c1438f40f13d8
				ExpectedHash:   utils.TestHexToFelt(t, "0x73cf79c4bfa0c7a41f473c07e1be5ac25faa7c2fdf9edcbd12c1438f40f13d8"),
				SetKS:          true,
				AccountAddress: utils.TestHexToFelt(t, "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"),
				PrivKey:        utils.TestHexToFelt(t, "0x043b7fe9d91942c98cd5fd37579bd99ec74f879c4c79d886633eecae9dad35fa"),
				PubKey:         "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e",
				ChainID:        "SN_GOERLI",
				FnCall: rpc.FunctionCall{
					Calldata: []*felt.Felt{
						utils.TestHexToFelt(t, "0x1"),
						utils.TestHexToFelt(t, "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
						utils.TestHexToFelt(t, "0x83afd3f4caedc6eebf44246fe54e38c95e3179a5ec9ea81740eca5b482d12e"),
						utils.TestHexToFelt(t, "0x0"),
						utils.TestHexToFelt(t, "0x3"),
						utils.TestHexToFelt(t, "0x3"),
						utils.TestHexToFelt(t, "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
						utils.TestHexToFelt(t, "0x1"),
						utils.TestHexToFelt(t, "0x0"),
					},
				},
				TxDetails: rpc.TxDetails{
					Nonce:   utils.TestHexToFelt(t, "0x2"),
					MaxFee:  utils.TestHexToFelt(t, "0x574fbde6000"),
					Version: rpc.TransactionV1,
				},
			},
			{
				ExpectedHash:   utils.TestHexToFelt(t, "0x135c34f53f8b7f59efd450eb689fccd9dd4cfe7f9d9dc4d09954c5653138698"),
				SetKS:          false,
				AccountAddress: &felt.Zero,
				ChainID:        "SN_GOERLI",
				FnCall: rpc.FunctionCall{
					ContractAddress:    &felt.Zero,
					EntryPointSelector: &felt.Zero,
					Calldata:           []*felt.Felt{&felt.Zero},
				},
				TxDetails: rpc.TxDetails{
					Nonce:  &felt.Zero,
					MaxFee: &felt.Zero,
				},
			},
			{
				ExpectedHash:   utils.TestHexToFelt(t, "0x3476c76a81522fe52616c41e95d062f5c3ea4eeb6c652904ad389fcd9ff4637"),
				SetKS:          false,
				AccountAddress: utils.TestHexToFelt(t, "0x59cd166e363be0a921e42dd5cfca0049aedcf2093a707ef90b5c6e46d4555a8"),
				ChainID:        "SN_MAIN",
				FnCall: rpc.FunctionCall{
					Calldata: []*felt.Felt{
						utils.TestHexToFelt(t, "0x1"),
						utils.TestHexToFelt(t, "0x5dbdedc203e92749e2e746e2d40a768d966bd243df04a6b712e222bc040a9af"),
						utils.TestHexToFelt(t, "0x2f0b3c5710379609eb5495f1ecd348cb28167711b73609fe565a72734550354"),
						utils.TestHexToFelt(t, "0x0"),
						utils.TestHexToFelt(t, "0x1"),
						utils.TestHexToFelt(t, "0x1"),
						utils.TestHexToFelt(t, "0x52884ee3f"),
					},
				},
				TxDetails: rpc.TxDetails{
					Nonce:   utils.TestHexToFelt(t, "0x1"),
					MaxFee:  utils.TestHexToFelt(t, "0x2a173cd36e400"),
					Version: rpc.TransactionV1,
				},
			},
		},
		"devnet":  {},
		"testnet": {},
		"mainnet": {},
	}[testEnv]
	for _, test := range testSet {

		t.Run("Transaction hash", func(t *testing.T) {
			ks := starknetgo.NewMemKeystore()
			if test.SetKS {
				privKeyBI, ok := new(big.Int).SetString(test.PrivKey.String(), 0)
				require.True(t, ok)
				ks.Put(test.PubKey, privKeyBI)
			}

			mockRpcProvider.EXPECT().ChainID(context.Background()).Return(test.ChainID, nil)
			account, err := account.NewAccount(mockRpcProvider, 1, test.AccountAddress, test.PubKey, ks)
			require.NoError(t, err, "error returned from account.NewAccount()")
			hash, err := account.TransactionHashInvoke(test.FnCall.Calldata, test.TxDetails.Nonce, test.TxDetails.MaxFee, account.AccountAddress)
			require.NoError(t, err, "error returned from account.TransactionHash()")
			require.Equal(t, test.ExpectedHash.String(), hash.String(), "transaction hash does not match expected")
		})
	}

}

func TestFmtCallData(t *testing.T) {
	type testSetType struct {
		FnCall           rpc.FunctionCall
		ExpectedCallData []*felt.Felt
	}
	testSet := map[string][]testSetType{
		"devnet": {},
		"mock": {
			{
				FnCall: rpc.FunctionCall{
					ContractAddress:    utils.TestHexToFelt(t, "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
					EntryPointSelector: types.GetSelectorFromNameFelt("transfer"),
					Calldata: []*felt.Felt{
						utils.TestHexToFelt(t, "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
						utils.TestHexToFelt(t, "0x1"),
					},
				},
				ExpectedCallData: []*felt.Felt{
					utils.TestHexToFelt(t, "0x1"),
					utils.TestHexToFelt(t, "0x49d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
					utils.TestHexToFelt(t, "0x83afd3f4caedc6eebf44246fe54e38c95e3179a5ec9ea81740eca5b482d12e"),
					utils.TestHexToFelt(t, "0x0"),
					utils.TestHexToFelt(t, "0x3"),
					utils.TestHexToFelt(t, "0x3"),
					utils.TestHexToFelt(t, "0x49d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
					utils.TestHexToFelt(t, "0x1"),
					utils.TestHexToFelt(t, "0x0"),
				},
			},
		},
		"testnet": {},
		"mainnet": {},
	}[testEnv]

	for _, test := range testSet {
		fmtCallData := account.FmtCalldata([]rpc.FunctionCall{test.FnCall})
		require.Equal(t, fmtCallData, test.ExpectedCallData)
	}
}

func TestChainIdMOCK(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockRpcProvider := mocks.NewMockRpcProvider(mockCtrl)

	type testSetType struct {
		ChainID    string
		ExpectedID string
	}
	testSet := map[string][]testSetType{
		"devnet": {},
		"mock": {
			{
				ChainID:    "SN_MAIN",
				ExpectedID: "0x534e5f4d41494e",
			},
			{
				ChainID:    "SN_GOERLI",
				ExpectedID: "0x534e5f474f45524c49",
			},
		},
		"testnet": {},
		"mainnet": {},
	}[testEnv]

	for _, test := range testSet {
		mockRpcProvider.EXPECT().ChainID(context.Background()).Return(test.ChainID, nil)
		account, err := account.NewAccount(mockRpcProvider, 1, &felt.Zero, "pubkey", starknetgo.NewMemKeystore())
		require.NoError(t, err)
		require.Equal(t, account.ChainId.String(), test.ExpectedID)
	}
}

func TestChainId(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	type testSetType struct {
		ChainID    string
		ExpectedID string
	}
	testSet := map[string][]testSetType{
		"devnet": {
			{
				ChainID:    "SN_GOERLI",
				ExpectedID: "0x534e5f474f45524c49",
			},
		},
		"mock":    {},
		"testnet": {},
		"mainnet": {},
	}[testEnv]

	for _, test := range testSet {
		client, err := rpc.NewClient(base)
		require.NoError(t, err, "Error in rpc.NewClient")
		provider := rpc.NewProvider(client)

		account, err := account.NewAccount(provider, 1, &felt.Zero, "pubkey", starknetgo.NewMemKeystore())
		require.NoError(t, err)
		require.Equal(t, account.ChainId.String(), test.ExpectedID)
	}

}

func TestSignMOCK(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockRpcProvider := mocks.NewMockRpcProvider(mockCtrl)

	type testSetType struct {
		Address     *felt.Felt
		PrivKey     *felt.Felt
		ChainId     string
		FeltToSign  *felt.Felt
		ExpectedSig []*felt.Felt
	}
	testSet := map[string][]testSetType{
		"mock": {
			// Accepted on testnet https://goerli.voyager.online/tx/0x73cf79c4bfa0c7a41f473c07e1be5ac25faa7c2fdf9edcbd12c1438f40f13d8
			{
				Address:    utils.TestHexToFelt(t, "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"),
				PrivKey:    utils.TestHexToFelt(t, "0x043b7fe9d91942c98cd5fd37579bd99ec74f879c4c79d886633eecae9dad35fa"),
				ChainId:    "SN_GOERLI",
				FeltToSign: utils.TestHexToFelt(t, "0x73cf79c4bfa0c7a41f473c07e1be5ac25faa7c2fdf9edcbd12c1438f40f13d8"),
				ExpectedSig: []*felt.Felt{
					utils.TestHexToFelt(t, "0x10d405427040655f118bc8b897e2f2f8147858bbcb0e3d6bc6dfbc6d0205e8"),
					utils.TestHexToFelt(t, "0x5cdfe4a3d5b63002e9011ec0ba59ae2b75a43cb2a3bc1699b35aa64cb9ca3cf"),
				},
			},
		},
		"devnet":  {},
		"testnet": {},
		"mainnet": {},
	}[testEnv]

	for _, test := range testSet {
		privKeyBI, ok := new(big.Int).SetString(test.PrivKey.String(), 0)
		require.True(t, ok)
		ks := starknetgo.NewMemKeystore()
		ks.Put(test.Address.String(), privKeyBI)

		mockRpcProvider.EXPECT().ChainID(context.Background()).Return(test.ChainId, nil)
		account, err := account.NewAccount(mockRpcProvider, 1, test.Address, test.Address.String(), ks)
		require.NoError(t, err, "error returned from account.NewAccount()")

		msg := utils.TestHexToFelt(t, "0x73cf79c4bfa0c7a41f473c07e1be5ac25faa7c2fdf9edcbd12c1438f40f13d8")
		sig, err := account.Sign(context.Background(), msg)

		require.NoError(t, err, "error returned from account.Sign()")
		require.Equal(t, test.ExpectedSig[0].String(), sig[0].String(), "s1 does not match expected")
		require.Equal(t, test.ExpectedSig[1].String(), sig[1].String(), "s2 does not match expected")
	}

}

func TestAddInvoke(t *testing.T) {

	type testSetType struct {
		ExpectedHash   *felt.Felt
		ExpectedError  string // todo :update when rpcv04 merged
		SetKS          bool
		AccountAddress *felt.Felt
		PubKey         *felt.Felt
		PrivKey        *felt.Felt
		InvokeTx       rpc.BroadcastedInvokeV1Transaction
		FnCall         rpc.FunctionCall
		TxDetails      rpc.TxDetails
	}
	testSet := map[string][]testSetType{
		"mock":   {},
		"devnet": {},
		"testnet": {{
			// https://goerli.voyager.online/tx/0x73cf79c4bfa0c7a41f473c07e1be5ac25faa7c2fdf9edcbd12c1438f40f13d8#overview
			ExpectedHash:   utils.TestHexToFelt(t, "0x73cf79c4bfa0c7a41f473c07e1be5ac25faa7c2fdf9edcbd12c1438f40f13d8"),
			ExpectedError:  "A transaction with the same hash already exists in the mempool",
			AccountAddress: utils.TestHexToFelt(t, "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"),
			SetKS:          true,
			PubKey:         utils.TestHexToFelt(t, "0x049f060d2dffd3bf6f2c103b710baf519530df44529045f92c3903097e8d861f"),
			PrivKey:        utils.TestHexToFelt(t, "0x043b7fe9d91942c98cd5fd37579bd99ec74f879c4c79d886633eecae9dad35fa"),
			InvokeTx: rpc.BroadcastedInvokeV1Transaction{
				BroadcastedTxnCommonProperties: rpc.BroadcastedTxnCommonProperties{
					Nonce:   new(felt.Felt).SetUint64(2),
					MaxFee:  utils.TestHexToFelt(t, "0x574fbde6000"),
					Version: rpc.TransactionV1,
					Type:    rpc.TransactionType_Invoke,
				},
				SenderAddress: utils.TestHexToFelt(t, "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"),
			},
			FnCall: rpc.FunctionCall{
				ContractAddress:    utils.TestHexToFelt(t, "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
				EntryPointSelector: types.GetSelectorFromNameFelt("transfer"),
				Calldata: []*felt.Felt{
					utils.TestHexToFelt(t, "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"),
					utils.TestHexToFelt(t, "0x1"),
				},
			},
		},
			{
				// https://goerli.voyager.online/tx/0x171537c58b16db45aeec3d3f493617cd3dd571561b856c115dc425b85212c86#overview
				ExpectedHash:   utils.TestHexToFelt(t, "0x171537c58b16db45aeec3d3f493617cd3dd571561b856c115dc425b85212c86"),
				ExpectedError:  "A transaction with the same hash already exists in the mempool",
				AccountAddress: utils.TestHexToFelt(t, "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"),
				SetKS:          true,
				PubKey:         utils.TestHexToFelt(t, "0x049f060d2dffd3bf6f2c103b710baf519530df44529045f92c3903097e8d861f"),
				PrivKey:        utils.TestHexToFelt(t, "0x043b7fe9d91942c98cd5fd37579bd99ec74f879c4c79d886633eecae9dad35fa"),
				InvokeTx: rpc.BroadcastedInvokeV1Transaction{
					BroadcastedTxnCommonProperties: rpc.BroadcastedTxnCommonProperties{
						Nonce:   new(felt.Felt).SetUint64(6),
						MaxFee:  utils.TestHexToFelt(t, "0x9184e72a000"),
						Version: rpc.TransactionV1,
						Type:    rpc.TransactionType_Invoke,
					},
					SenderAddress: utils.TestHexToFelt(t, "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"),
				},
				FnCall: rpc.FunctionCall{
					ContractAddress:    utils.TestHexToFelt(t, "0x03E85bFbb8E2A42B7BeaD9E88e9A1B19dbCcf661471061807292120462396ec9"),
					EntryPointSelector: types.GetSelectorFromNameFelt("burn"),
					Calldata: []*felt.Felt{
						utils.TestHexToFelt(t, "0x043784df59268c02b716e20bf77797bd96c68c2f100b2a634e448c35e3ad363e"),
						utils.TestHexToFelt(t, "0x1"),
					},
				},
			},
		},
		"mainnet": {},
	}[testEnv]

	for _, test := range testSet {
		client, err := rpc.NewClient(base + "/rpc")
		require.NoError(t, err, "Error in rpc.NewClient")
		provider := rpc.NewProvider(client)

		// Set up ks
		ks := starknetgo.NewMemKeystore()
		if test.SetKS {
			fakePrivKeyBI, ok := new(big.Int).SetString(test.PrivKey.String(), 0)
			require.True(t, ok)
			ks.Put(test.PubKey.String(), fakePrivKeyBI)
		}

		acnt, err := account.NewAccount(provider, 1, test.AccountAddress, test.PubKey.String(), ks)
		require.NoError(t, err)

		require.NoError(t, acnt.BuildInvokeTx(context.Background(), &test.InvokeTx, &[]rpc.FunctionCall{test.FnCall}), "Error building Invoke")

		txHash, err := acnt.TransactionHashInvoke(test.InvokeTx.Calldata, test.InvokeTx.Nonce, test.InvokeTx.MaxFee, acnt.AccountAddress)
		require.NoError(t, err)
		require.Equal(t, txHash.String(), test.ExpectedHash.String())

		resp, err := acnt.AddInvokeTransaction(context.Background(), &test.InvokeTx)
		require.Equal(t, err.Error(), test.ExpectedError)
		require.Nil(t, resp)
	}
}

func TestAddDeployAccountDevnet(t *testing.T) {
	// if testEnv != "devnet" {
	// 	t.Skip("Skipping test as it requires a devnet environment")
	// }
	client, err := rpc.NewClient(base + "/rpc")
	require.NoError(t, err, "Error in rpc.NewClient")
	provider := rpc.NewProvider(client)

	acnts, err := newDevnet(t, base)
	require.NoError(t, err, "Error setting up Devnet")
	fakeUser := acnts[0]
	fakeUserAddr := utils.TestHexToFelt(t, fakeUser.Address)
	fakeUserPub := utils.TestHexToFelt(t, fakeUser.PublicKey)

	// Set up ks
	ks := starknetgo.NewMemKeystore()
	fakePrivKeyBI, ok := new(big.Int).SetString(fakeUser.PrivateKey, 0)
	require.True(t, ok)
	ks.Put(fakeUser.PublicKey, fakePrivKeyBI)

	acnt, err := account.NewAccount(provider, 1, fakeUserAddr, fakeUser.PublicKey, ks)
	require.NoError(t, err)

	classHash := utils.TestHexToFelt(t, "0x2794ce20e5f2ff0d40e632cb53845b9f4e526ebd8471983f7dbd355b721d5a") // preDeployed classhash
	require.NoError(t, err)

	tx := rpc.BroadcastedDeployAccountTransaction{
		BroadcastedTxnCommonProperties: rpc.BroadcastedTxnCommonProperties{
			Nonce:     &felt.Zero, // Contract accounts start with nonce zero.
			MaxFee:    new(felt.Felt).SetUint64(4724395326064),
			Type:      rpc.TransactionType_DeployAccount,
			Version:   rpc.TransactionV1,
			Signature: []*felt.Felt{},
		},
		ClassHash:           classHash,
		ContractAddressSalt: fakeUserPub,
		ConstructorCalldata: []*felt.Felt{fakeUserPub},
	}

	precomputedAddress, err := acnt.PrecomputeAddress(&felt.Zero, fakeUserPub, classHash, tx.ConstructorCalldata)
	require.NoError(t, acnt.SignDeployAccountTransaction(context.Background(), &tx, precomputedAddress))

	// Send transaction to the network
	// WIP breaks here
	resp, err := acnt.AddDeployAccountTransaction(context.Background(), tx)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func newDevnet(t *testing.T, url string) ([]test.TestAccount, error) {
	devnet := test.NewDevNet(url)
	acnts, err := devnet.Accounts()
	return acnts, err
}

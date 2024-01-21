package marginfi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var GroupAddress = solana.MustPublicKeyFromBase58("4qp6Fx6tnZkY5Wropq9wUYgtFxXKwE6viZxFHg3rdAG8")

type Client struct {
	banks        map[solana.PublicKey]*Bank
	oraclePrices map[solana.PublicKey]*OraclePrice
}

func NewClient(connection *rpc.Client) (*Client, error) {
	banksMetadata, err := loadBankMetadatas()
	if err != nil {
		return nil, err
	}

	allBanks := make([]solana.PublicKey, 0, len(banksMetadata))
	for bankKey := range banksMetadata {
		allBanks = append(allBanks, bankKey)
	}

	banksRes, err := connection.GetMultipleAccounts(context.Background(), allBanks...)
	if err != nil {
		return nil, err
	}

	bankKeys := make([]solana.PublicKey, 0, len(allBanks))
	banksMap := make(map[solana.PublicKey]*Bank, len(allBanks))
	for i, bankRaw := range banksRes.Value {
		bank := ParseBank(bankRaw.Data.GetBinary())
		// TODO add support for banks that uses switchboard as oracle
		if bank.Config.OracleSetup != OracleSetupPyth {
			continue
		}
		banksMap[allBanks[i]] = bank
		bankKeys = append(bankKeys, allBanks[i])
	}

	priceFeedsKeys := make([]solana.PublicKey, 0, len(bankKeys))
	for _, bank := range bankKeys {
		bank := banksMap[bank]
		priceFeedsKeys = append(priceFeedsKeys, bank.Config.OracleKeys[0])
	}

	priceFeedsRes, err := connection.GetMultipleAccounts(context.Background(), priceFeedsKeys...)
	if err != nil {
		return nil, err
	}

	priceFeedsMap := make(map[solana.PublicKey]*OraclePrice, len(bankKeys))
	for i, priceFeedRaw := range priceFeedsRes.Value {
		if bankKeys[i].Equals(solana.MustPublicKeyFromBase58("CCKtUs6Cgwo4aaQUmBPmyoApH2gUDErxNZCAntD6LYGh")) {
			print("test")
		}
		priceFeedsMap[bankKeys[i]] = ParseOraclePrice(OracleSetupPyth, priceFeedRaw.Data.GetBinary())
	}

	return &Client{
		banks:        banksMap,
		oraclePrices: priceFeedsMap,
	}, nil
}

type BankMetadata struct {
	TokenAddress solana.PublicKey
	TokenName    string
	TokenSymbol  string
}

func loadBankMetadatas() (map[solana.PublicKey]BankMetadata, error) {
	res, err := http.Get("https://storage.googleapis.com/mrgn-public/mrgn-bank-metadata-cache.json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var bankMetadatas []struct {
		BankAddress solana.PublicKey
		BankMetadata
	}
	if err := json.NewDecoder(res.Body).Decode(&bankMetadatas); err != nil {
		return nil, err
	}

	bankMetadatasMap := make(map[solana.PublicKey]BankMetadata, len(bankMetadatas))
	for _, bankMetadata := range bankMetadatas {
		bankMetadatasMap[bankMetadata.BankAddress] = bankMetadata.BankMetadata
	}
	return bankMetadatasMap, nil
}

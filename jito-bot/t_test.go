package main

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	lookup "github.com/gagliardetto/solana-go/programs/address-lookup-table"
	"github.com/gagliardetto/solana-go/rpc"
)

func TestTx(t *testing.T) {
	blockhashRes, _ := solanaConnection.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)

	bundle, err := MakeRaydiumSwapBundle(
		wallet,
		SwapSell,
		solana.MustPublicKeyFromBase58("GGRi6A6wdqEQpVjXyeasXqNhNmrinm2uwi2b8K4VJJ74"),
		10000,
		&RaydiumPoolKeys{
			Id:         solana.MustPublicKeyFromBase58("7EtQSxB5iXWZGGwoW5eyiZiS3Xrf4Hv7Cx5AfMGcwdZY"),
			BaseVault:  solana.MustPublicKeyFromBase58("FQAQSYdkEb2otqHDyNMqmAoNYQauB3wdipWPmAh89Q4Q"),
			QuoteVault: solana.MustPublicKeyFromBase58("AMk6XA1pnjyoD95fgqf4wFhDqnWKgi84s8vK7Vu3iW7W"),

			MarketId:         solana.MustPublicKeyFromBase58("4s6c4FhYrwXywcjdq9jVdEQGHJMmoxtQ9DCpdCcpnXeu"),
			MarketBaseVault:  solana.MustPublicKeyFromBase58("CDrnuyboXpUqp3PTw1GJqx58BWDZE9kJXzg3kYgGU2BX"),
			MarketQuoteVault: solana.MustPublicKeyFromBase58("Bj2ddz9aivyuaFNRCCsUAttQSZGKus8Xx3WAp2DqVmse"),
			MarketBids:       solana.MustPublicKeyFromBase58("F1udms6PhWBoBSs9XUC6D2Mpj2ZVDKWidhMK9ByZwFUF"),
			MarketAsks:       solana.MustPublicKeyFromBase58("DjgjAThcCQRyrNNW1F8W9CxYJBustMgf1oReKLFMk65i"),
			MarketEventQueue: solana.MustPublicKeyFromBase58("96jvqmSSKsKZBi5Su37yT9aoWE9whxwd8FSizKPMbRBg"),
		},
		blockhashRes.Value.Blockhash,
		[]*solana.Transaction{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sig, err := solanaConnection.SendRawTransactionWithOpts(ctx, bundle.Packets[0].Data, rpc.TransactionOpts{SkipPreflight: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sig)
}

func TestLookup(t *testing.T) {
	v := uint64(0)
	txRes, err := solanaConnection.GetTransaction(
		ctx,
		solana.MustSignatureFromBase58("5dufhcUEdpeJJtuxv4x3KqrmUSatdJgqqraVV2k7nwCzC4BQdG35bk4YV6NZfEKiaAc9VqHp4eNw9XddyGUMeNwx"),
		&rpc.GetTransactionOpts{MaxSupportedTransactionVersion: &v, Commitment: rpc.CommitmentFinalized})
	if err != nil {
		t.Error(t)
	}

	tx, err := txRes.Transaction.GetTransaction()
	if err != nil {
		t.Error(err)
	}

	now := time.Now()

	tblKeys := tx.Message.GetAddressTableLookups().GetTableIDs()
	resolutions := make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, key := range tblKeys {
		t.Log("Getting table", key)

		info, err := solanaConnection.GetAccountInfo(
			ctx,
			key,
		)
		if err != nil {
			t.Error(err)
		}

		tableContent, err := lookup.DecodeAddressLookupTableState(info.GetBinary())
		if err != nil {
			panic(err)
		}

		t.Log("table content:", spew.Sdump(tableContent))
		t.Log("isActive", tableContent.IsActive())

		resolutions[key] = tableContent.Addresses
	}

	tx.Message.SetAddressTables(resolutions)

	t.Log("took", time.Since(now))

	t.Log(tx)
}

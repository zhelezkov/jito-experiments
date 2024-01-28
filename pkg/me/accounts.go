package me

import (
	"context"
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var M2ProgramAddress = solana.MustPublicKeyFromBase58("M2mx93ekt1fmXSVkTrUL9xVFHkmME8HTUi5Cyc5aF7K")
var M3ProgramAddress = solana.MustPublicKeyFromBase58("M3mxk5W2tt27WGT7THox7PmgRDp4m6NEhL5xvxrBfS1")

const TakerFee = 0.025 // 2.5%

var M2SellerTradeStateV2Discriminator = [...]byte{
	0xa4, 0x0e, 0x5c, 0x64, 0x7b, 0x39, 0xea, 0xcc,
}

type M2SellerTradeState struct {
	AuctionHouseKey solana.PublicKey
	Seller          solana.PublicKey
	SellerReferral  solana.PublicKey
	BuyerPrice      uint64
	TokenMint       solana.PublicKey
	TokenSize       uint64
	// bump u8
	Expiry      int64
	PaymentMint solana.PublicKey
}

func ParseM2SellerTradeState(data []byte) *M2SellerTradeState {
	data = data[8:] // skip discriminator
	return &M2SellerTradeState{
		AuctionHouseKey: solana.PublicKey(data[0:32]),
		Seller:          solana.PublicKey(data[32:64]),
		SellerReferral:  solana.PublicKey(data[64:96]),
		BuyerPrice:      binary.LittleEndian.Uint64(data[96:104]),
		TokenMint:       solana.PublicKey(data[104:136]),
		TokenSize:       binary.LittleEndian.Uint64(data[136:144]),
		Expiry:          int64(binary.LittleEndian.Uint64(data[144:152])),
		PaymentMint:     solana.PublicKey(data[152:184]),
	}
}

func FindAllM2SellterTradeStates(connection *rpc.Client) ([]*M2SellerTradeState, error) {
	gpa, err := connection.GetProgramAccountsWithOpts(context.Background(), M2ProgramAddress, &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  M2SellerTradeStateV2Discriminator[:],
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	sellerTradeStates := make([]*M2SellerTradeState, 0, len(gpa))
	for _, acc := range gpa {
		tradeState := ParseM2SellerTradeState(acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		sellerTradeStates = append(sellerTradeStates, tradeState)
	}

	return sellerTradeStates, nil
}

var M3SellerTradeDiscriminator = [...]byte{
	0x01, 0xee, 0x48, 0x89, 0x8a, 0x15, 0xfe, 0xf9,
}

type M3SellerTradeState struct {
	Seller         solana.PublicKey
	SellerReferral solana.PublicKey
	BuyerPrice     uint64
	AssetId        solana.PublicKey
	PaymentMint    solana.PublicKey
	// padding u8
	MerkleTree solana.PublicKey
	Index      uint32
	CreatedAt  int64
	UpdatedAt  int64
}

func ParseM3SellerTradeState(data []byte) *M3SellerTradeState {
	data = data[8:] // skip discriminator
	return &M3SellerTradeState{
		Seller:         solana.PublicKey(data[0:32]),
		SellerReferral: solana.PublicKey(data[32:64]),
		BuyerPrice:     binary.LittleEndian.Uint64(data[64:72]),
		AssetId:        solana.PublicKey(data[72:104]),
		PaymentMint:    solana.PublicKey(data[104:136]),
		MerkleTree:     solana.PublicKey(data[137:169]),
		Index:          binary.LittleEndian.Uint32(data[169:173]),
		CreatedAt:      int64(binary.LittleEndian.Uint64(data[173:181])),
		UpdatedAt:      int64(binary.LittleEndian.Uint64(data[181:189])),
	}
}

func FindAllM3SellterTradeStates(connection *rpc.Client) ([]*M3SellerTradeState, error) {
	gpa, err := connection.GetProgramAccountsWithOpts(context.Background(), M3ProgramAddress, &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  M3SellerTradeDiscriminator[:],
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	sellerTradeStates := make([]*M3SellerTradeState, 0, len(gpa))
	for _, acc := range gpa {
		wl := ParseM3SellerTradeState(acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		sellerTradeStates = append(sellerTradeStates, wl)
	}

	return sellerTradeStates, nil
}

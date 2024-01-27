package me

import (
	"context"
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var M3ProgramAddress = solana.MustPublicKeyFromBase58("M3mxk5W2tt27WGT7THox7PmgRDp4m6NEhL5xvxrBfS1")

const TakerFee = 0.025 // 2.5%

var SellerTradeDiscriminator = [...]byte{
	0x01, 0xee, 0x48, 0x89, 0x8a, 0x15, 0xfe, 0xf9,
}

type SellerTradeState struct {
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

func ParseSellerTradeState(data []byte) *SellerTradeState {
	data = data[8:] // skip discriminator
	return &SellerTradeState{
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

func FindAllSellterTradeStates(connection *rpc.Client) ([]*SellerTradeState, error) {
	gpa, err := connection.GetProgramAccountsWithOpts(context.Background(), M3ProgramAddress, &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  SellerTradeDiscriminator[:],
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	sellerTradeStates := make([]*SellerTradeState, 0, len(gpa))
	for _, acc := range gpa {
		wl := ParseSellerTradeState(acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		sellerTradeStates = append(sellerTradeStates, wl)
	}

	return sellerTradeStates, nil
}

package me

import (
	"encoding/binary"
	"github.com/gagliardetto/solana-go"
)

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

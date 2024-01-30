package tensor

import (
	"context"
	"encoding/binary"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var CompressedMarketplaceProgramAddress = solana.MustPublicKeyFromBase58("TCMPhJdwDryooaGtiocG1u3xcYbRpiJzb283XfCZsDp")
var WhitelistProgramAddress = solana.MustPublicKeyFromBase58("TL1ST2iRBzuGTqLn1KXnGdSnEow62BzPnGiqyRXhWtW")
var TSwapProgramAddress = solana.MustPublicKeyFromBase58("TSWAPaqyCSx2KABk68Shruf4rp7CxcNi8hAsbdwmHbN")
var TBidProgramAddress = solana.MustPublicKeyFromBase58("TB1Dqt8JeKQh7RLDzfYDJsq8KS4fS2yt87avRjyRxMv")

const TakerFee = 0.014 // 1.4%

var WhitelistDiscriminator = [...]byte{
	0xcc, 0xb0, 0x34, 0x4f, 0x92, 0x79, 0x36, 0xf7,
}

// SIZE 8 + 1 + 1 + 1 + (32 * 3) + 1 + (33 * 2) + 64 = 238
type Whitelist struct {
	Address solana.PublicKey // not part of the account data

	Version uint8
	// padding 1
	// deprecated 1
	RootHash [32]byte
	UUID     [32]byte
	Name     [32]byte
	Frozen   bool
	Voc      *solana.PublicKey
	Fvc      *solana.PublicKey
}

func ParseWhitelist(address solana.PublicKey, data []byte) (*Whitelist, error) {
	decoder := bin.NewBinDecoder(data)
	decoder.SkipBytes(8) // skip discriminator
	version, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	decoder.SkipBytes(1)
	decoder.SkipBytes(1)
	rootHash, err := decoder.ReadBytes(32)
	if err != nil {
		return nil, err
	}
	uuid, err := decoder.ReadBytes(32)
	if err != nil {
		return nil, err
	}
	name, err := decoder.ReadBytes(32)
	if err != nil {
		return nil, err
	}
	isFrozen, err := decoder.ReadBool()
	if err != nil {
		return nil, err
	}
	hasVoc, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	var voc solana.PublicKey
	if hasVoc {
		vocBytes, err := decoder.ReadBytes(solana.PublicKeyLength)
		if err != nil {
			return nil, err
		}
		voc = solana.PublicKey(vocBytes)
	}

	hasFvc, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	var fvc solana.PublicKey
	if hasFvc {
		fvcBytes, err := decoder.ReadBytes(solana.PublicKeyLength)
		if err != nil {
			return nil, err
		}
		fvc = solana.PublicKey(fvcBytes)
	}

	return &Whitelist{
		Address:  address,
		Version:  version,
		RootHash: [32]byte(rootHash),
		UUID:     [32]byte(uuid),
		Name:     [32]byte(name),
		Frozen:   isFrozen,
		Voc:      &voc,
		Fvc:      &fvc,
	}, nil
}

func FindAllWhitelists(connection *rpc.Client) ([]*Whitelist, error) {
	gpa, err := connection.GetProgramAccountsWithOpts(context.Background(), WhitelistProgramAddress, &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  WhitelistDiscriminator[:],
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	whitelists := make([]*Whitelist, 0, len(gpa))
	for _, acc := range gpa {
		wl, err := ParseWhitelist(acc.Pubkey, acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		whitelists = append(whitelists, wl)
	}

	return whitelists, nil
}

type TargetType uint8

const (
	TargetTypeAssetId TargetType = iota
	TargetTypeWhitelist
)

type FieldType uint8

const (
	FieldTypeName FieldType = iota
)

var BidStateDiscriminator = [...]byte{
	0x9b, 0xc5, 0x05, 0x61, 0xbd, 0x3c, 0x08, 0xb7,
}

// SIZE 8 + 1 + 1 + (32 * 2) + 1 + 32 + 2 + 33 + 4 * 2 + 8 + 33 + 8 + (33 * 3) + 128 = 426
type CnftBidState struct {
	Address solana.PublicKey // not part of the account data

	Version uint8
	// padding 1
	Owner          solana.PublicKey
	BidId          solana.PublicKey
	Target         TargetType
	TargetId       solana.PublicKey
	Field          *uint8
	FieldId        *solana.PublicKey
	Quantity       uint32
	FilledQuantity uint32
	Amount         uint64
	Currency       *solana.PublicKey
	Expiry         int64
	PrivateTaker   *solana.PublicKey
	MakerBroker    *solana.PublicKey
	Margin         *solana.PublicKey
	UpdatedAt      int64
	Cosigner       solana.PublicKey
}

func ParseCnftBidState(address solana.PublicKey, data []byte) (*CnftBidState, error) {
	decoder := bin.NewBinDecoder(data)
	decoder.SkipBytes(8) // skip discriminator
	version, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	decoder.SkipBytes(1)
	owner, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	bidId, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	target, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	targetId, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	hasField, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	if hasField {
		decoder.SkipBytes(1)
	}
	hasFieldId, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	if hasFieldId {
		decoder.SkipBytes(32)
	}
	quantity, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	filledQuantity, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	amount, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	hasCurrency, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	if hasCurrency {
		decoder.SkipBytes(32)
	}
	expiry, err := decoder.ReadInt64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	hasPrivateTaker, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	if hasPrivateTaker {
		decoder.SkipBytes(32)
	}
	hasMakerBroker, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	if hasMakerBroker {
		decoder.SkipBytes(32)
	}
	hasMargin, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	var margin solana.PublicKey
	if hasMargin {
		marginBytes, err := decoder.ReadBytes(solana.PublicKeyLength)
		if err != nil {
			return nil, err
		}
		margin = solana.PublicKey(marginBytes)
	}
	updatedAt, err := decoder.ReadInt64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	cosigner, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	return &CnftBidState{
		Address:        address,
		Version:        version,
		Owner:          solana.PublicKey(owner),
		BidId:          solana.PublicKey(bidId),
		Target:         TargetType(target),
		TargetId:       solana.PublicKey(targetId),
		Field:          nil,
		FieldId:        nil,
		Quantity:       quantity,
		FilledQuantity: filledQuantity,
		Amount:         amount,
		Currency:       nil,
		Expiry:         expiry,
		PrivateTaker:   nil,
		MakerBroker:    nil,
		Margin:         &margin,
		UpdatedAt:      updatedAt,
		Cosigner:       solana.PublicKey(cosigner),
	}, nil
}

func FindAllCnftBidStates(connection *rpc.Client) ([]*CnftBidState, error) {
	gpa, err := connection.GetProgramAccountsWithOpts(context.Background(), CompressedMarketplaceProgramAddress, &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  BidStateDiscriminator[:],
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	bidstates := make([]*CnftBidState, 0, len(gpa))
	for _, acc := range gpa {
		bidState, err := ParseCnftBidState(acc.Pubkey, acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		bidstates = append(bidstates, bidState)
	}

	return bidstates, nil
}

var NftBidStateDiscriminator = [...]byte{
	0x9b, 0xc5, 0x05, 0x61, 0xbd, 0x3c, 0x08, 0xb7,
}

// SIZE 8 + 1 + 8 + (32 * 2) + 1 + 8 + 33 + 8 + 56
type NftBidState struct {
	Address solana.PublicKey // not part of the account data

	Version   uint8
	BidAmount uint64
	NftMint   solana.PublicKey
	Bidder    solana.PublicKey
	// bump 1
	Expiry    int64
	Margin    *solana.PublicKey
	UpdatedAt int64
}

func ParseNftBidState(address solana.PublicKey, data []byte) (*NftBidState, error) {
	decoder := bin.NewBinDecoder(data)
	decoder.SkipBytes(8) // skip discriminator
	version, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	bidAmount, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	nftMint, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	bidder, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	decoder.SkipBytes(1)
	expiry, err := decoder.ReadInt64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	hasMargin, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	var margin solana.PublicKey
	if hasMargin {
		marginBytes, err := decoder.ReadBytes(solana.PublicKeyLength)
		if err != nil {
			return nil, err
		}
		margin = solana.PublicKey(marginBytes)
	}
	updatedAt, err := decoder.ReadInt64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	return &NftBidState{
		Address:   solana.PublicKey(address),
		Version:   version,
		BidAmount: bidAmount,
		NftMint:   solana.PublicKey(nftMint),
		Bidder:    solana.PublicKey(bidder),
		Expiry:    expiry,
		Margin:    &margin,
		UpdatedAt: updatedAt,
	}, nil
}

func FindAllNftBidStates(connection *rpc.Client) ([]*NftBidState, error) {
	gpa, err := connection.GetProgramAccountsWithOpts(context.Background(), TBidProgramAddress, &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  NftBidStateDiscriminator[:],
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	bidstates := make([]*NftBidState, 0, len(gpa))
	for _, acc := range gpa {
		bidState, err := ParseNftBidState(acc.Pubkey, acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		bidstates = append(bidstates, bidState)
	}

	return bidstates, nil
}

var TSwapPoolDiscriminator = [...]byte{
	0xf1, 0x9a, 0x6d, 0x04, 0x11, 0xb1, 0x6d, 0xbc,
}

type poolConfig struct {
	PoolType       uint8
	CurveType      uint8
	StartingPrice  uint64
	Delta          uint64
	MMCompoindFees bool
	MMFeeBps       *uint16
}

type poolStats struct {
	TakerSellCount      uint32
	TakerBuyCount       uint32
	AccumulatedMMProfit uint64
}

type Frozen struct {
	Amount uint64
	Time   int64
}

// SIZE 8 + (3 * 1) + 8 + (2 * 1) + (2 * 8) + 1 + 3 + (5 * 32) + (3 * 4) + (2 * 4) + 8 + 32 + 1 + 1 + 1 + 8 + 8 + 1 + 8 + 4
type TSwapPool struct {
	Address solana.PublicKey // not part of the account data

	Version               uint8
	Bump                  uint8
	SolEscrowBump         uint8
	CreatedUnixSeconds    int64
	PoolConfig            poolConfig
	TSwap                 solana.PublicKey
	Owner                 solana.PublicKey
	Whitelist             solana.PublicKey
	SolEscrow             solana.PublicKey
	TakerSellCount        uint32
	TakerBuyCount         uint32
	NFTsHeld              uint32
	NftAuthority          solana.PublicKey
	PoolStats             poolStats
	Margin                *solana.PublicKey
	IsCosigned            bool
	OrderType             uint8
	Frozen                *Frozen
	LastTransactedSeconds int64
	MaxTakerSellCount     uint32
}

func ParseTSwapPool(address solana.PublicKey, data []byte) (*TSwapPool, error) {
	decoder := bin.NewBinDecoder(data)
	decoder.SkipBytes(8) // skip discriminator
	version, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	bump, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	solEscrowBump, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	createdUnixSeconds, err := decoder.ReadInt64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	poolConfig := poolConfig{}
	poolConfig.PoolType, err = decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	poolConfig.CurveType, err = decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	poolConfig.StartingPrice, err = decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	poolConfig.Delta, err = decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	poolConfig.MMCompoindFees, err = decoder.ReadBool()
	if err != nil {
		return nil, err
	}
	hasMMFeeBps, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	if hasMMFeeBps {
		mmFeeBps, err := decoder.ReadUint16(binary.LittleEndian)
		if err != nil {
			return nil, err
		}
		poolConfig.MMFeeBps = &mmFeeBps
	}
	tswap, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	owner, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	whitelist, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	solEscrow, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	takerSellCount, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	takerBuyCount, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	nftsHeld, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	nftAuthority, err := decoder.ReadBytes(solana.PublicKeyLength)
	if err != nil {
		return nil, err
	}
	poolStats := poolStats{}
	poolStats.TakerSellCount, err = decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	poolStats.TakerBuyCount, err = decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	poolStats.AccumulatedMMProfit, err = decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	hasMargin, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	var margin *solana.PublicKey
	if hasMargin {
		marginBytes, err := decoder.ReadBytes(solana.PublicKeyLength)
		if err != nil {
			return nil, err
		}
		margin = (*solana.PublicKey)(marginBytes)
	}
	isCosigned, err := decoder.ReadBool()
	if err != nil {
		return nil, err
	}
	orderType, err := decoder.ReadUint8()
	if err != nil {
		return nil, err
	}
	hasFrozen, err := decoder.ReadOption()
	if err != nil {
		return nil, err
	}
	var frozen *Frozen
	if hasFrozen {
		frozen = &Frozen{}
		frozen.Amount, err = decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return nil, err
		}
		frozen.Time, err = decoder.ReadInt64(binary.LittleEndian)
		if err != nil {
			return nil, err
		}
	}
	lastTransactedSeconds, err := decoder.ReadInt64(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	maxTakerSellCount, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	return &TSwapPool{
		Address:               address,
		Version:               version,
		Bump:                  bump,
		SolEscrowBump:         solEscrowBump,
		CreatedUnixSeconds:    createdUnixSeconds,
		PoolConfig:            poolConfig,
		TSwap:                 solana.PublicKey(tswap),
		Owner:                 solana.PublicKey(owner),
		Whitelist:             solana.PublicKey(whitelist),
		SolEscrow:             solana.PublicKey(solEscrow),
		TakerSellCount:        takerSellCount,
		TakerBuyCount:         takerBuyCount,
		NFTsHeld:              nftsHeld,
		NftAuthority:          solana.PublicKey(nftAuthority),
		PoolStats:             poolStats,
		Margin:                margin,
		IsCosigned:            isCosigned,
		OrderType:             orderType,
		Frozen:                frozen,
		LastTransactedSeconds: lastTransactedSeconds,
		MaxTakerSellCount:     maxTakerSellCount,
	}, nil
}

func FindAllTSwapPools(connection *rpc.Client) ([]*TSwapPool, error) {
	gpa, err := connection.GetProgramAccountsWithOpts(context.Background(), TSwapProgramAddress, &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Filters: []rpc.RPCFilter{{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  TSwapPoolDiscriminator[:],
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	pools := make([]*TSwapPool, 0, len(gpa))
	for _, acc := range gpa {
		pool, err := ParseTSwapPool(acc.Pubkey, acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}

	return pools, nil
}

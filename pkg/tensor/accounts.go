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
type BidState struct {
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

func ParseBidState(data []byte) (*BidState, error) {
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
	return &BidState{
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

func FindAllBidStates(connection *rpc.Client) ([]*BidState, error) {
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

	bidstates := make([]*BidState, 0, len(gpa))
	for _, acc := range gpa {
		bidState, err := ParseBidState(acc.Account.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		bidstates = append(bidstates, bidState)
	}

	return bidstates, nil
}

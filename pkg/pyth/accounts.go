// export interface Base {
//   magic: number
//   version: number
//   type: AccountType
//   size: number
// }

// export interface MappingData extends Base {
//   nextMappingAccount: PublicKey | null
//   productAccountKeys: PublicKey[]
// }

// export interface Product {
//   [index: string]: string
// }

// export interface ProductData extends Base {
//   priceAccountKey: PublicKey | null
//   product: Product
// }

// export interface Price {
//   priceComponent: bigint
//   price: number
//   confidenceComponent: bigint
//   confidence: number
//   status: PriceStatus
//   corporateAction: CorpAction
//   publishSlot: number
// }

// export interface PriceComponent {
//   publisher: PublicKey
//   aggregate: Price
//   latest: Price
// }

// /**
//  * valueComponent = numerator / denominator
//  * value = valueComponent * 10 ^ exponent (from PriceData)
//  */
// export interface Ema {
//   valueComponent: bigint
//   value: number
//   numerator: bigint
//   denominator: bigint
// }

// export interface PriceData extends Base {
//   priceType: PriceType
//   exponent: number
//   numComponentPrices: number
//   numQuoters: number
//   lastSlot: bigint
//   validSlot: bigint
//   emaPrice: Ema
//   emaConfidence: Ema
//   timestamp: bigint
//   minPublishers: number
//   drv2: number
//   drv3: number
//   drv4: number
//   productAccountKey: PublicKey
//   nextPriceAccountKey: PublicKey | null
//   previousSlot: bigint
//   previousPriceComponent: bigint
//   previousPrice: number
//   previousConfidenceComponent: bigint
//   previousConfidence: number
//   previousTimestamp: bigint
//   priceComponents: PriceComponent[]
//   aggregate: Price
//   // The current price and confidence and status. The typical use of this interface is to consume these three fields.
//   // If undefined, Pyth does not currently have price information for this product. This condition can
//   // happen for various reasons (e.g., US equity market is closed, or insufficient publishers), and your
//   // application should handle it gracefully. Note that other raw price information fields (such as
//   // aggregate.price) may be defined even if this is undefined; you most likely should not use those fields,
//   // as their value can be arbitrary when this is undefined.
//   price: number | undefined
//   confidence: number | undefined
//   status: PriceStatus
// }

package pyth

import (
	"encoding/binary"
)

const Magic uint32 = 0xa1b2c3d4
const Version uint32 = 2
const MaxSlotDifference = 25

type PriceStatus uint32

const (
	PriceStatusUnknown = iota
	PriceStatusTrading
	PriceStatusHalted
	PriceStatusAuction
	PriceStatusBreak
)

type CorpAction uint32

const (
	CorpActionNoCorpAct = iota
)

type PriceType uint32

const (
	PriceTypeUnknown = iota
	PriceTypePrice
)

type DeriveType uint32

const (
	DeriveTypeUnknown = iota
	DeriveTypeVolatility
)

type AccountType uint32

const (
	AccountTypeUnknown = iota
	AccountTypeMapping
	AccountTypeProduct
	AccountTypePrice
	AccountTypeTest
	AccountTypePermission
)

type PriceComponent struct{}

type Base struct {
	Magic   uint32
	Version uint32
	Type    AccountType // AccountType
	Size    uint32
}

type Ema struct {
	ValueComponent int64
	Value          float64
	Numerator      int64
	Denominator    int64
}

type PriceInfo struct {
	Price       int64
	Conf        uint64
	Status      PriceStatus
	PublishSlot uint64
}

type PriceData struct {
	Base
	PriceType     PriceType
	Exponent      int32
	LastSlot      uint64
	ValidSlot     uint64
	Timestamp     int64
	PrevSlot      uint64
	PrevPrice     int64
	PrevConf      uint64
	PrevTimestamp int64
	Agg           PriceInfo
}

func ParsePriceData(data []byte) *PriceData {
	return &PriceData{
		Base: Base{
			Magic:   binary.LittleEndian.Uint32(data[0:4]),
			Version: binary.LittleEndian.Uint32(data[4:8]),
			Type:    AccountType(binary.LittleEndian.Uint32(data[8:12])),
			Size:    binary.LittleEndian.Uint32(data[12:16]),
		},
		PriceType:     PriceType(binary.LittleEndian.Uint32(data[16:20])),
		Exponent:      int32(binary.LittleEndian.Uint32(data[20:24])),
		LastSlot:      binary.LittleEndian.Uint64(data[32:40]),
		ValidSlot:     binary.LittleEndian.Uint64(data[40:48]),
		Timestamp:     int64(binary.LittleEndian.Uint64(data[96:104])),
		PrevSlot:      binary.LittleEndian.Uint64(data[176:184]),
		PrevPrice:     int64(binary.LittleEndian.Uint64(data[184:192])),
		PrevConf:      binary.LittleEndian.Uint64(data[192:200]),
		PrevTimestamp: int64(binary.LittleEndian.Uint64(data[200:208])),
		Agg:           ParsePriceInfo(data[208:240]),
	}
}

func ParsePriceInfo(data []byte) PriceInfo {
	return PriceInfo{
		Price:       int64(binary.LittleEndian.Uint64(data[0:8])),
		Conf:        binary.LittleEndian.Uint64(data[8:16]),
		Status:      PriceStatus(binary.LittleEndian.Uint32(data[16:20])),
		PublishSlot: binary.LittleEndian.Uint64(data[24:32]),
	}
}

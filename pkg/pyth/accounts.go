package pyth

import (
	"encoding/binary"
	"math"
)

const PriceConfIntervals float64 = 2.12

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

type Base struct {
	Magic   uint32
	Version uint32
	Type    AccountType
	Size    uint32
}

type Ema struct {
	ValueComponent int64
	Value          float64
	Numerator      int64
	Denominator    int64
}

type PriceInfo struct {
	PriceComponent int64
	Price          float64
	ConfComponent  uint64
	Conf           float64
	Status         PriceStatus
	PublishSlot    uint64
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
	magic := binary.LittleEndian.Uint32(data[0:4])
	if magic != Magic {
		panic("invalid magic")
	}
	exponent := int32(binary.LittleEndian.Uint32(data[20:24]))
	return &PriceData{
		Base: Base{
			Magic:   magic,
			Version: binary.LittleEndian.Uint32(data[4:8]),
			Type:    AccountType(binary.LittleEndian.Uint32(data[8:12])),
			Size:    binary.LittleEndian.Uint32(data[12:16]),
		},
		PriceType:     PriceType(binary.LittleEndian.Uint32(data[16:20])),
		Exponent:      exponent,
		LastSlot:      binary.LittleEndian.Uint64(data[32:40]),
		ValidSlot:     binary.LittleEndian.Uint64(data[40:48]),
		Timestamp:     int64(binary.LittleEndian.Uint64(data[96:104])),
		PrevSlot:      binary.LittleEndian.Uint64(data[176:184]),
		PrevPrice:     int64(binary.LittleEndian.Uint64(data[184:192])),
		PrevConf:      binary.LittleEndian.Uint64(data[192:200]),
		PrevTimestamp: int64(binary.LittleEndian.Uint64(data[200:208])),
		Agg:           ParsePriceInfo(data[208:240], exponent),
	}
}

func ParsePriceInfo(data []byte, exponent int32) PriceInfo {
	priceComponent := int64(binary.LittleEndian.Uint64(data[0:8]))
	confComponent := binary.LittleEndian.Uint64(data[8:16])
	return PriceInfo{
		PriceComponent: priceComponent,
		Price:          float64(priceComponent) * math.Pow10(int(exponent)),
		ConfComponent:  confComponent,
		Conf:           float64(confComponent) * math.Pow10(int(exponent)),
		Status:         PriceStatus(binary.LittleEndian.Uint32(data[16:20])),
		PublishSlot:    binary.LittleEndian.Uint64(data[24:32]),
	}
}

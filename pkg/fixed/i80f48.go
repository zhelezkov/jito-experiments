package fixed

import (
	"fmt"
	"github.com/Pilatuz/bigz"
	"github.com/Pilatuz/bigz/uint128"
	"github.com/Pilatuz/bigz/uint256"
	"math"
)

const i80f48FractionalBits uint = 48

var (
	multiplier2Pow48              = uint256.From64(281474976710656)
	multiplier2Pow48Float float64 = 281474976710656

	I80f48Pow10 = [...]I80F48{
		MustI80F48FromFloat64(math.Pow10(0)),
		MustI80F48FromFloat64(math.Pow10(1)),
		MustI80F48FromFloat64(math.Pow10(2)),
		MustI80F48FromFloat64(math.Pow10(3)),
		MustI80F48FromFloat64(math.Pow10(4)),
		MustI80F48FromFloat64(math.Pow10(5)),
		MustI80F48FromFloat64(math.Pow10(6)),
		MustI80F48FromFloat64(math.Pow10(7)),
		MustI80F48FromFloat64(math.Pow10(8)),
		MustI80F48FromFloat64(math.Pow10(9)),
		MustI80F48FromFloat64(math.Pow10(10)),
		MustI80F48FromFloat64(math.Pow10(11)),
		MustI80F48FromFloat64(math.Pow10(12)),
		MustI80F48FromFloat64(math.Pow10(13)),
		MustI80F48FromFloat64(math.Pow10(14)),
		MustI80F48FromFloat64(math.Pow10(15)),
		MustI80F48FromFloat64(math.Pow10(16)),
	}
)

type I80F48 struct {
	bigz.Uint256
}

func MustI80F48FromLittleEndian(data []byte) I80F48 {
	return I80F48{uint256.From128(uint128.LoadLittleEndian(data))}
}

func MustI80F48FromFloat64(data float64) I80F48 {
	intPart := math.Trunc(data)
	val := uint256.From64(uint64(intPart)).Lsh(i80f48FractionalBits)
	val = val.Add(uint256.From64(uint64((data - intPart) * multiplier2Pow48Float)))
	return I80F48{val}
}

func (u I80F48) Add(n I80F48) I80F48 {
	return I80F48{u.Uint256.Add(n.Uint256)}
}

func (u I80F48) Sub(n I80F48) I80F48 {
	return I80F48{u.Uint256.Sub(n.Uint256)}
}

func (u I80F48) Mul(n I80F48) I80F48 {
	return I80F48{u.Uint256.Mul(n.Uint256).Rsh(i80f48FractionalBits)}
}

func (u I80F48) Mul64(n uint64) I80F48 {
	return I80F48{u.Uint256.Mul128(uint128.From64(n)).Rsh(i80f48FractionalBits)}
}

func (u I80F48) Div(n I80F48) I80F48 {
	return I80F48{u.Uint256.Lsh(i80f48FractionalBits).Div(n.Uint256)}
}

func (u I80F48) Div64(n uint64) I80F48 {
	return I80F48{u.Uint256.Lsh(i80f48FractionalBits).Div64(n)}
}

func (u I80F48) LessThan(n I80F48) bool {
	return u.Uint256.Cmp(n.Uint256) < 0
}

func (u I80F48) BiggerThanOrEqual(n I80F48) bool {
	return u.Uint256.Cmp(n.Uint256) >= 0
}

func (u I80F48) AsFloat64() float64 {
	div, mod := u.QuoRem(multiplier2Pow48)

	res := float64(div.Lo.Lo)
	res += float64(mod.Lo.Lo) / multiplier2Pow48Float

	return res
}

func (u I80F48) String() string {
	return fmt.Sprintf("%v", u.AsFloat64())
}

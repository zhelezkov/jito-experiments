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
)

type I80F48 struct {
	//num.U256
	bigz.Uint256
}

func MustI80F48FromLittleEndian(data []byte) I80F48 {
	return I80F48{uint256.From128(uint128.LoadLittleEndian(data))}
}

func MustI80F48FromFloat64(data float64) I80F48 {
	return I80F48{uint256.From64(math.Float64bits(data * multiplier2Pow48Float))}
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

func (u I80F48) AsFloat64() float64 {
	div, mod := u.QuoRem(multiplier2Pow48)

	res := float64(div.Lo.Lo)
	res += float64(mod.Lo.Lo) / multiplier2Pow48Float

	return res
}

func (u I80F48) String() string {
	return fmt.Sprintf("%v", u.AsFloat64())
}

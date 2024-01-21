package fixed

import "github.com/shabbyrobe/go-num"

const i80f48FractionalBits uint = 48

var (
	multiplier2Pow48              = num.MustU128FromString("281474976710656")
	multiplier2Pow48Float float64 = 281474976710656
)

type I80F48 struct {
	num.U128
}

func MustI80F48FromLittleEndian(data []byte) I80F48 {
	return I80F48{num.MustU128FromLittleEndian(data)}
}

func (u I80F48) Add(n I80F48) I80F48 {
	return I80F48{u.U128.Add(n.U128)}
}

func (u I80F48) Mul(n I80F48) I80F48 {
	return I80F48{u.U128.Mul(n.U128).Rsh(i80f48FractionalBits)}
}

func (u I80F48) Div(n I80F48) I80F48 {
	return I80F48{u.U128.Lsh(i80f48FractionalBits).Quo(n.U128)}
}

func (u I80F48) Div64(n uint64) I80F48 {
	return I80F48{u.U128.Lsh(i80f48FractionalBits).Quo64(n)}
}

func (u I80F48) LessThan(n I80F48) bool {
	return u.U128.LessThan(n.U128)
}

func (u I80F48) AsFloat64() float64 {
	div, mod := u.QuoRem(multiplier2Pow48)

	res := div.AsFloat64()
	res += mod.AsFloat64() / multiplier2Pow48Float

	return res
}

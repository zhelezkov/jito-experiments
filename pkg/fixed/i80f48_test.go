package fixed

import (
	"testing"
)

func TestMustI80F48FromFloat64(t *testing.T) {
	floatVal := 9313074543.0 / 100000000.0
	val := MustI80F48FromFloat64(floatVal)
	if floatVal != val.AsFloat64() {
		t.Fail()
	}
}

func TestMul(t *testing.T) {
	f := 0.10150730311063184
	t.Log(f)
	a := MustI80F48FromFloat64(f)
	if a.AsFloat64() != f {
		t.Fail()
	}
	ff := 1033754.0841600463
	t.Log(ff)
	b := MustI80F48FromFloat64(ff)
	if b.AsFloat64() != ff {
		t.Fail()
	}
	c := a.Mul(b)
	if c.AsFloat64() != a.AsFloat64()*b.AsFloat64() {
		t.Log("a", a.AsFloat64(), "b", b.AsFloat64(), "c", c.AsFloat64())
		t.Fail()
	}
}

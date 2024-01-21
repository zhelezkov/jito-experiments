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

package main

import (
	"math"
	"testing"
)

func TestWireStorage(t *testing.T) {
	values := []string{"1", "-0", "NaN", "4"}
	a, err := pack(2, 2, values)
	if err != nil || !padding(a) || a.Data[0] != 1 || !math.Signbit(a.Data[1]) || !math.IsNaN(a.Data[a.Stride]) || a.Data[a.Stride+1] != 4 {
		t.Fatalf("pack: %+v %v", a, err)
	}
	for i, v := range unpack(a) {
		if v != values[i] {
			t.Fatalf("roundtrip: %v", unpack(a))
		}
	}
	a.Data[2] = 0
	if padding(a) {
		t.Fatal("missed padding corruption")
	}
	if _, err := pack(2, 0, values[:3]); err == nil {
		t.Fatal("accepted truncated wire matrix")
	}
	if _, err := pack(2, 0, []string{"1", "2", "bad", "4"}); err == nil {
		t.Fatal("accepted malformed number")
	}
}

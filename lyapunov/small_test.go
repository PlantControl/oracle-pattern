package lyapunov

import (
	"math"
	"testing"
)

func TestSmallSymmetricResidual(t *testing.T) {
	for _, trans := range []bool{false, true} {
		for _, discrete := range []bool{false, true} {
			a := []float64{-0.3, 0.7, -0.4, -0.2}
			b := []float64{2, -0.6, math.NaN(), 3}
			var x []float64
			var scale float64
			var perturbed bool
			if discrete {
				x, scale, perturbed = smallDiscrete(trans, a, b)
			} else {
				x, scale, perturbed = smallContinuous(trans, a, b)
			}
			if perturbed || scale != 1 {
				t.Fatalf("unexpected status scale=%g perturbed=%v", scale, perturbed)
			}
			at := func(i, j int) float64 {
				if trans {
					i, j = j, i
				}
				return a[i*2+j]
			}
			for i := 0; i < 2; i++ {
				for j := i; j < 2; j++ {
					got := 0.0
					if discrete {
						for k := 0; k < 2; k++ {
							for l := 0; l < 2; l++ {
								got += at(k, i) * x[k*2+l] * at(l, j)
							}
						}
						got -= x[i*2+j]
					} else {
						for k := 0; k < 2; k++ {
							got += at(k, i)*x[k*2+j] + x[i*2+k]*at(k, j)
						}
					}
					if math.Abs(got-b[i*2+j]) > 1e-13 {
						t.Fatalf("trans=%v discrete=%v (%d,%d): got %g want %g", trans, discrete, i, j, got, b[i*2+j])
					}
				}
			}
		}
	}
}

func TestSmallProductResidual(t *testing.T) {
	for n1 := 1; n1 <= 2; n1++ {
		for n2 := 1; n2 <= 2; n2++ {
			for _, lt := range []bool{false, true} {
				for _, rt := range []bool{false, true} {
					l := []float64{0.2, -0.6, 0.3, 0.1}[:n1*n1]
					r := []float64{-0.4, 0.7, -0.5, 0.2}[:n2*n2]
					b := []float64{2, -1, 3, 0.7}[:n1*n2]
					x, s, p := smallProduct(lt, rt, n1, n2, l, r, b)
					if p || s != 1 {
						t.Fatalf("unexpected status scale=%g perturbed=%v", s, p)
					}
					at := func(a []float64, n, i, j int, tr bool) float64 {
						if tr {
							i, j = j, i
						}
						return a[i*n+j]
					}
					for i := 0; i < n1; i++ {
						for j := 0; j < n2; j++ {
							got := -x[i*n2+j]
							for k := 0; k < n1; k++ {
								for q := 0; q < n2; q++ {
									got += at(l, n1, i, k, lt) * x[k*n2+q] * at(r, n2, q, j, rt)
								}
							}
							if math.Abs(got-b[i*n2+j]) > 1e-13 {
								t.Fatalf("shape=%dx%d trans=%v,%v: got %g want %g", n1, n2, lt, rt, got, b[i*n2+j])
							}
						}
					}
				}
			}
		}
	}
}

func TestSmallScalingAndPerturbation(t *testing.T) {
	for _, solve := range []func(bool, []float64, []float64) ([]float64, float64, bool){smallContinuous, smallDiscrete} {
		a := []float64{0, 0, 0, 0}
		if x, s, p := solve(false, a, []float64{1, 0, 0, 1}); !(s > 0 && s <= 1) || math.IsInf(x[0], 0) || math.IsNaN(x[0]) {
			t.Fatalf("invalid singular result %v %g %v", x, s, p)
		}
	}
	x, s, p := smallProduct(false, false, 1, 1, []float64{1}, []float64{1}, []float64{1e300})
	if !p || !(s > 0 && s < 1) || math.IsInf(x[0], 0) || math.IsNaN(x[0]) {
		t.Fatalf("invalid scaled singular result %v %g %v", x, s, p)
	}
	x, s, p = smallContinuous(false, []float64{1e-200, 0, 0, 1e-200}, []float64{1e300, 0, 0, 1e300})
	if p || !(s > 0 && s < 1) || math.IsInf(x[0], 0) || math.IsNaN(x[0]) {
		t.Fatalf("invalid scaled result %v %g %v", x, s, p)
	}
}

func TestSmallFinalPivotStatus(t *testing.T) {
	// SB03MW silently floors a singular final pivot, unlike earlier pivots.
	_, _, p := smallContinuous(false, []float64{1, 0, 0, -1}, []float64{1, 1, 1, 1})
	if p {
		t.Fatal("final pivot changed SLICOT INFO semantics")
	}
	_, _, p = smallDiscrete(false, []float64{1, 0, 0, 1}, []float64{1, 1, 1, 1})
	if !p {
		t.Fatal("early singular pivot was not reported")
	}
}

func TestSmallProductScaledBlocks(t *testing.T) {
	for n1 := 1; n1 <= 2; n1++ {
		for n2 := 1; n2 <= 2; n2++ {
			l := make([]float64, n1*n1)
			r := make([]float64, n2*n2)
			for i := 0; i < n1; i++ {
				l[i*n1+i] = 1
			}
			for i := 0; i < n2; i++ {
				r[i*n2+i] = 1
			}
			b := make([]float64, n1*n2)
			for i := range b {
				b[i] = 1e300
			}
			x, s, p := smallProduct(false, false, n1, n2, l, r, b)
			if !p || !(s > 0 && s < 1) {
				t.Fatalf("shape=%dx%d: scale=%g perturbed=%v", n1, n2, s, p)
			}
			for _, v := range x {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					t.Fatalf("shape=%dx%d: nonfinite solution %v", n1, n2, x)
				}
			}
		}
	}
}

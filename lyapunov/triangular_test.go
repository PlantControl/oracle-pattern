package lyapunov

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestSolveSchurResidual(t *testing.T) {
	for _, discrete := range []bool{false, true} {
		for _, trans := range []bool{false, true} {
			for n := 1; n <= 8; n++ {
				stride := n + 2
				a, want, rhs := make([]float64, n*stride), make([]float64, n*n), make([]float64, n*stride)
				for i := 0; i < n; i++ {
					for j := i; j < n; j++ {
						a[i*stride+j] = 0.02 * float64(i+j+1)
					}
					a[i*stride+i] = -0.3
					for j := 0; j < n; j++ {
						want[i*n+j] = math.Sin(float64(i + j + 1))
					}
					for j := n; j < stride; j++ {
						a[i*stride+j], rhs[i*stride+j] = 12345, 12345
					}
				}
				for i := 0; i+1 < n; i += 3 {
					a[i*stride+i+1] = 0.2
					a[(i+1)*stride+i] = -0.1
				}
				entry := func(i, j int) float64 {
					if trans {
						return a[j*stride+i]
					}
					return a[i*stride+j]
				}
				for i := 0; i < n; i++ {
					for j := 0; j < n; j++ {
						if discrete {
							rhs[i*stride+j] = -want[i*n+j]
							for k := 0; k < n; k++ {
								for l := 0; l < n; l++ {
									rhs[i*stride+j] += entry(k, i) * want[k*n+l] * entry(l, j)
								}
							}
						} else {
							for k := 0; k < n; k++ {
								rhs[i*stride+j] += entry(k, i)*want[k*n+j] + want[i*n+k]*entry(k, j)
							}
						}
					}
				}
				scale, perturbed := solveSchur(discrete, trans, n, a, stride, rhs, stride)
				if scale != 1 || perturbed {
					t.Fatalf("discrete=%v trans=%v n=%d: scale=%g perturbed=%v", discrete, trans, n, scale, perturbed)
				}
				for i := 0; i < n; i++ {
					for j := 0; j < n; j++ {
						if math.Abs(rhs[i*stride+j]-want[i*n+j]) > 1e-12 {
							t.Fatalf("discrete=%v trans=%v n=%d at (%d,%d): got %.17g want %.17g", discrete, trans, n, i, j, rhs[i*stride+j], want[i*n+j])
						}
					}
					for j := n; j < stride; j++ {
						if rhs[i*stride+j] != 12345 || a[i*stride+j] != 12345 {
							t.Fatal("modified matrix padding")
						}
					}
				}
			}
		}
	}
}

func TestSolveSchurScaling(t *testing.T) {
	rhs := []float64{1e300}
	scale, perturbed := solveSchur(false, false, 1, []float64{1e-100}, 1, rhs, 1)
	if scale >= 1 || scale <= 0 || perturbed || math.IsInf(rhs[0], 0) {
		t.Fatalf("scale=%g perturbed=%v solution=%g", scale, perturbed, rhs[0])
	}
	if math.Abs((2e-100*rhs[0])/(scale*1e300)-1) > 1e-14 {
		t.Fatal("scaled equation is not satisfied")
	}
}

func TestSolveSchurPerturbation(t *testing.T) {
	for _, discrete := range []bool{false, true} {
		a := 0.0
		if discrete {
			a = 1
		}
		rhs := []float64{1}
		scale, perturbed := solveSchur(discrete, false, 1, []float64{a}, 1, rhs, 1)
		if !perturbed || scale != 1 || math.IsInf(rhs[0], 0) || rhs[0] <= 0 {
			t.Fatalf("discrete=%v scale=%g perturbed=%v solution=%g", discrete, scale, perturbed, rhs[0])
		}
	}
}

func TestSolveSchurAsymmetricTriangle(t *testing.T) {
	for _, discrete := range []bool{false, true} {
		for _, trans := range []bool{false, true} {
			a := []float64{-0.2, 0, 0, -0.3}
			rhs := []float64{1, 2, 7, 3}
			denom := -0.5
			if discrete {
				denom = -0.94
			}
			want := 7 / denom
			if trans {
				want = 2 / denom
			}
			solveSchur(discrete, trans, 2, a, 2, rhs, 2)
			if math.Abs(rhs[1]-want) > 1e-14 || rhs[1] != rhs[2] {
				t.Fatalf("discrete=%v trans=%v got %v want off-diagonal=%g", discrete, trans, rhs, want)
			}
		}
	}
}

func TestSolveSchurReferenceEstimatorProbes(t *testing.T) {
	data, err := os.ReadFile("testdata/sb03mx-probes.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		A      []float64
		Probes []struct {
			Name          string
			Trans         bool
			Before, After []float64
			Scale         float64
			Perturbed     bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, probe := range fixture.Probes {
		t.Run(probe.Name, func(t *testing.T) {
			scale, perturbed := solveSchur(true, probe.Trans, 6, fixture.A, 6, probe.Before, 6)
			if scale != probe.Scale || perturbed != probe.Perturbed {
				t.Fatalf("scale/status mismatch: %g %v", scale, perturbed)
			}
			for i, want := range probe.After {
				if math.Abs(probe.Before[i]-want) > 1e-13*math.Max(1, math.Abs(want)) {
					t.Fatalf("entry %d got %.17g want %.17g", i, probe.Before[i], want)
				}
			}
		})
	}
}

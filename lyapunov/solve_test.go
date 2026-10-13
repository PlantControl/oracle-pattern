package lyapunov_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/jamestjsp/oracle-pattern/lyapunov"
	"plantcontrol.org/v1/gonum/blas"
	"plantcontrol.org/v1/gonum/blas/blas64"
)

func matrix(data []float64) blas64.General {
	return blas64.General{Rows: 2, Cols: 2, Stride: 2, Data: append([]float64(nil), data...)}
}

func TestSolveComputedAndSupplied(t *testing.T) {
	for _, eq := range []lyapunov.Equation{lyapunov.Continuous, lyapunov.Discrete} {
		for _, tr := range []blas.Transpose{blas.NoTrans, blas.Trans, blas.ConjTrans} {
			a0 := matrix([]float64{-0.4, 0.7, -0.2, -0.1})
			a, u := matrix(a0.Data), matrix(make([]float64, 4))
			c := matrix([]float64{2, 0.3, math.NaN(), 1})
			opts := lyapunov.Options{Equation: eq, Job: lyapunov.Both, Transpose: tr}
			result, err := lyapunov.Solve(a, u, c, opts)
			if err != nil || result.NearSingular || result.Scale != 1 || result.Sep <= 0 || result.Ferr <= 0 {
				t.Fatalf("eq=%c tr=%c: result=%+v err=%v", eq, tr, result, err)
			}
			if len(result.WR) != 2 || len(result.WI) != 2 || math.Abs(result.WR[0]+result.WR[1]+0.5) > 1e-14 || math.Abs(result.WI[0]+result.WI[1]) > 1e-14 || result.WI[0] == 0 {
				t.Fatalf("incorrect complex eigenvalues: %+v", result)
			}
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					reconstructed, orthogonal := 0.0, 0.0
					for k := 0; k < 2; k++ {
						orthogonal += u.Data[k*2+i] * u.Data[k*2+j]
						for l := 0; l < 2; l++ {
							reconstructed += u.Data[i*2+k] * a.Data[k*2+l] * u.Data[j*2+l]
						}
					}
					identity := 0.0
					if i == j {
						identity = 1
					}
					if math.Abs(reconstructed-a0.Data[i*2+j]) > 1e-14 || math.Abs(orthogonal-identity) > 1e-14 {
						t.Fatal("Schur factorization invariant failed")
					}
				}
			}
			checkEquation(t, eq, tr, a0.Data, c.Data, []float64{2, 0.3, 0.3, 1}, result.Scale)
			beforeA, beforeU := append([]float64(nil), a.Data...), append([]float64(nil), u.Data...)
			other := matrix([]float64{2, 0.3, -9876, 1})
			opts.SuppliedSchur = true
			supplied, err := lyapunov.Solve(a, u, other, opts)
			if err != nil || len(supplied.WR) != 0 || len(supplied.WI) != 0 || !reflect.DeepEqual(a.Data, beforeA) || !reflect.DeepEqual(u.Data, beforeU) {
				t.Fatalf("supplied factors changed or eigenvalues produced: %+v err=%v", supplied, err)
			}
			for i, v := range c.Data {
				if math.Abs(v-other.Data[i]) > 1e-14 {
					t.Fatalf("computed and supplied solutions differ: %v %v", c.Data, other.Data)
				}
			}
			if supplied.Sep != result.Sep || supplied.Ferr != result.Ferr {
				t.Fatal("computed and supplied condition estimates differ")
			}
		}
	}
}

func checkEquation(t *testing.T, eq lyapunov.Equation, tr blas.Transpose, a, x, rhs []float64, scale float64) {
	t.Helper()
	entry := func(i, j int) float64 {
		if tr != blas.NoTrans {
			return a[j*2+i]
		}
		return a[i*2+j]
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			v := 0.0
			if eq == lyapunov.Discrete {
				v = -x[i*2+j]
				for k := 0; k < 2; k++ {
					for l := 0; l < 2; l++ {
						v += entry(k, i) * x[k*2+l] * entry(l, j)
					}
				}
			} else {
				for k := 0; k < 2; k++ {
					v += entry(k, i)*x[k*2+j] + x[i*2+k]*entry(k, j)
				}
			}
			if math.IsNaN(v) || math.Abs(v-scale*rhs[i*2+j]) > 1e-13 {
				t.Fatalf("eq=%c tr=%c residual at (%d,%d): got %g want %g", eq, tr, i, j, v, scale*rhs[i*2+j])
			}
		}
	}
}

func TestSolveDefaultsAndLowercase(t *testing.T) {
	var want []float64
	for _, opts := range []lyapunov.Options{{}, {Equation: 'c', Job: 'x', Transpose: 'n'}} {
		a, u, c := matrix([]float64{-1, 0, 0, -2}), matrix(make([]float64, 4)), matrix([]float64{2, 3, 3, 4})
		r, err := lyapunov.Solve(a, u, c, opts)
		if err != nil || r.Sep != 0 || r.Ferr != 0 || r.Scale != 1 {
			t.Fatalf("result=%+v err=%v", r, err)
		}
		if want == nil {
			want = []float64{-1, -1, -1, -1}
		}
		if !reflect.DeepEqual(c.Data, want) {
			t.Fatalf("got %v want %v", c.Data, want)
		}
	}
}

func TestSolveInvalidArgumentsAreTransactional(t *testing.T) {
	for _, name := range []string{"equation", "job", "transpose", "order", "a-shape", "a-stride", "a-data", "u-shape", "u-data", "c-shape", "c-data"} {
		t.Run(name, func(t *testing.T) {
			a, u, c := matrix([]float64{-1, 0, 0, -2}), matrix([]float64{9, 8, 7, 6}), matrix([]float64{2, 3, 3, 4})
			opts := lyapunov.Options{}
			switch name {
			case "equation":
				opts.Equation = '?'
			case "job":
				opts.Job = '?'
			case "transpose":
				opts.Transpose = '?'
			case "order":
				a.Rows = -1
			case "a-shape":
				a.Cols = 1
			case "a-stride":
				a.Stride = 1
			case "a-data":
				a.Data = a.Data[:3]
			case "u-shape":
				u.Rows = 1
			case "u-data":
				u.Data = u.Data[:3]
			case "c-shape":
				c.Cols = 1
			case "c-data":
				c.Data = c.Data[:3]
			}
			before := [][]float64{append([]float64(nil), a.Data...), append([]float64(nil), u.Data...), append([]float64(nil), c.Data...)}
			_, err := lyapunov.Solve(a, u, c, opts)
			var arg lyapunov.ArgumentError
			if !errors.As(err, &arg) {
				t.Fatalf("got %v, want ArgumentError", err)
			}
			for i, after := range [][]float64{a.Data, u.Data, c.Data} {
				if !reflect.DeepEqual(before[i], after) {
					t.Fatal("invalid call modified data")
				}
			}
		})
	}
}

func TestSolveSeparationDoesNotUseC(t *testing.T) {
	for _, c := range []blas64.General{{}, {Rows: -2, Cols: 99, Stride: -1, Data: []float64{1, 2}}} {
		a, u := matrix([]float64{-1, 0, 0, -2}), matrix([]float64{1, 0, 0, 1})
		before := append([]float64(nil), c.Data...)
		r, err := lyapunov.Solve(a, u, c, lyapunov.Options{Job: lyapunov.Separation, SuppliedSchur: true})
		if err != nil || r.Sep != 2 || r.Scale != 1 || r.Ferr != 0 || !reflect.DeepEqual(before, c.Data) {
			t.Fatalf("result=%+v err=%v c=%v", r, err, c)
		}
	}
}

func TestSolveExtremeSchurScalingAndPadding(t *testing.T) {
	for _, magnitude := range []float64{1e-250, 1e250} {
		a := blas64.General{Rows: 2, Cols: 2, Stride: 3, Data: []float64{-magnitude, 0, 99, 0, -2 * magnitude, 99}}
		u := blas64.General{Rows: 2, Cols: 2, Stride: 3, Data: []float64{0, 0, 99, 0, 0, 99}}
		c := blas64.General{Rows: 2, Cols: 2, Stride: 3, Data: []float64{2, 0, 99, 0, 4, 99}}
		r, err := lyapunov.Solve(a, u, c, lyapunov.Options{})
		if err != nil || r.Scale != 1 || r.NearSingular {
			t.Fatalf("magnitude=%g result=%+v err=%v", magnitude, r, err)
		}
		for i := 0; i < 2; i++ {
			if math.Abs(c.Data[i*3+i]*magnitude+1) > 1e-14 || math.Abs(r.WR[i]/magnitude+float64(i+1)) > 1e-14 || r.WI[i] != 0 {
				t.Fatalf("magnitude=%g result=%+v x=%v", magnitude, r, c.Data)
			}
			if a.Data[i*3+2] != 99 || u.Data[i*3+2] != 99 || c.Data[i*3+2] != 99 {
				t.Fatal("changed padding")
			}
		}
	}
}

func TestSolveEmpty(t *testing.T) {
	empty := blas64.General{Stride: 1}
	for _, job := range []lyapunov.Job{lyapunov.Solution, lyapunov.Separation, lyapunov.Both} {
		r, err := lyapunov.Solve(empty, empty, empty, lyapunov.Options{Job: job})
		if err != nil || r.Scale != 1 || r.Sep != 0 || r.Ferr != 0 || r.NearSingular || len(r.WR) != 0 || len(r.WI) != 0 {
			t.Fatalf("job=%c result=%+v err=%v", job, r, err)
		}
	}
}

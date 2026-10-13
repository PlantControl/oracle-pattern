// Package lyapunov solves real continuous and discrete Lyapunov equations.
package lyapunov

import (
	"errors"
	"math"

	"plantcontrol.org/v1/gonum/blas"
	"plantcontrol.org/v1/gonum/blas/blas64"
	"plantcontrol.org/v1/gonum/lapack"
	"plantcontrol.org/v1/gonum/lapack/gonum"
)

// Equation selects the continuous or discrete Lyapunov equation.
type Equation byte

const (
	Continuous Equation = 'C'
	Discrete   Equation = 'D'
)

// Job selects which outputs to compute.
type Job byte

const (
	Solution   Job = 'X'
	Separation Job = 'S'
	Both       Job = 'B'
)

// Options controls Solve. Zero values select Continuous, Solution, NoTrans,
// and computation of the Schur factors. SuppliedSchur uses A and U as factors.
type Options struct {
	Equation      Equation
	Job           Job
	Transpose     blas.Transpose
	SuppliedSchur bool
}

// Result contains the scale in the solved equation, the requested separation
// and relative forward-error estimates, and eigenvalues when Schur factors were
// computed. NearSingular mirrors SLICOT INFO=N+1. Some final-pivot flooring
// in SLICOT does not set this status.
// Unrequested estimates are zero; Scale is one for separation-only calls.
type Result struct {
	Scale, Sep, Ferr float64
	WR, WI           []float64
	NearSingular     bool
}

// ArgumentError identifies an invalid option or matrix descriptor.
type ArgumentError string

func (e ArgumentError) Error() string { return "lyapunov: invalid " + string(e) }

// ErrSchur indicates that the Schur iteration did not converge. The underlying
// Gonum API does not expose the Fortran index of the unconverged eigenvalue.
var ErrSchur = errors.New("lyapunov: Schur decomposition did not converge")

// Solve overwrites C with X satisfying op(A)'*X+X*op(A)=Scale*C (Continuous)
// or op(A)'*X*op(A)-X=Scale*C (Discrete). op(A) is A or its transpose.
// Only C's upper triangle is read. Separation leaves C untouched and permits
// an empty C descriptor. A and U receive the real Schur factors unless
// SuppliedSchur is set; supplied factors are unchanged. Arrays must not overlap.
// All descriptors are row-major and row padding is preserved. Workspace is
// allocated internally. Invalid arguments return before modifying any matrix.
// Inputs must be finite in active regions, and supplied Schur factors must be
// valid, with orthogonal U. See docs/sb03md-contract.md for the numerical domain.
//
// This is a translation of SLICOT SB03MD. See NOTICE and LICENSE-SLICOT.
func Solve(a, u, c blas64.General, opts Options) (Result, error) {
	eq, job, tr := opts.Equation, opts.Job, opts.Transpose
	if eq == 0 {
		eq = Continuous
	}
	if job == 0 {
		job = Solution
	}
	if tr == 0 {
		tr = blas.NoTrans
	}
	eq = Equation(upper(byte(eq)))
	job = Job(upper(byte(job)))
	tr = blas.Transpose(upper(byte(tr)))
	var out Result
	if eq != Continuous && eq != Discrete {
		return out, ArgumentError("equation")
	}
	if job != Solution && job != Separation && job != Both {
		return out, ArgumentError("job")
	}
	if tr != blas.NoTrans && tr != blas.Trans && tr != blas.ConjTrans {
		return out, ArgumentError("transpose")
	}
	n := a.Rows
	if n < 0 {
		return out, ArgumentError("order")
	}
	if !square(a, n) {
		return out, ArgumentError("a")
	}
	if !square(u, n) {
		return out, ArgumentError("u")
	}
	if job != Separation && !square(c, n) {
		return out, ArgumentError("c")
	}
	out.Scale = 1
	if n == 0 {
		return out, nil
	}
	impl := gonum.Implementation{}
	if !opts.SuppliedSchur {
		out.WR, out.WI = make([]float64, n), make([]float64, n)
		if !schur(a, u, out.WR, out.WI) {
			return out, ErrSchur
		}
	}
	if job != Separation {
		congruence(c, u, true)
		out.Scale, out.NearSingular = solveSchur(eq == Discrete, tr != blas.NoTrans, n, a.Data, a.Stride, c.Data, c.Stride)
		congruence(c, u, false)
	}
	if job != Solution {
		nn := n * n
		v, x := make([]float64, nn), make([]float64, nn)
		signs := make([]int, nn)
		var state [3]int
		est, scalef, kase := 0.0, 1.0, 0
		// LAPACK treats both signs of zero as positive in its sign probes.
		for {
			for i, value := range x {
				if value == 0 {
					x[i] = 0
				}
			}
			est, kase = impl.Dlacn2(nn, v, x, signs, est, kase, &state)
			if kase == 0 {
				break
			}
			// DLACN2's vector uses Fortran column order, including nonsymmetric probes.
			transposeSquare(x, n)
			trans := tr != blas.NoTrans
			if kase != 1 {
				trans = !trans
			}
			scalef, _ = solveSchur(eq == Discrete, trans, n, a.Data, a.Stride, x, n)
			transposeSquare(x, n)
		}
		out.Sep = scalef / est
		if job == Both {
			norm := impl.Dlanhs(lapack.Frobenius, n, a.Data, a.Stride, nil)
			out.Ferr = 0x1p-52 * norm / out.Sep
			if eq == Discrete {
				out.Ferr = (0x1p-52 * norm) * norm / out.Sep
			}
		}
	}
	return out, nil
}

func square(a blas64.General, n int) bool {
	if a.Rows != n || a.Cols != n || a.Stride < max(1, n) {
		return false
	}
	return n == 0 || len(a.Data) >= n && (len(a.Data)-n)/a.Stride >= n-1
}
func upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}
func transposeSquare(a []float64, n int) {
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			a[i*n+j], a[j*n+i] = a[j*n+i], a[i*n+j]
		}
	}
}

func schur(a, u blas64.General, wr, wi []float64) bool {
	impl := gonum.Implementation{}
	n := a.Rows
	norm := impl.Dlange(lapack.MaxAbs, n, n, a.Data, a.Stride, nil)
	low := math.Sqrt(0x1p-1022) / 0x1p-52
	high := 1 / low
	target := norm
	if norm > 0 && norm < low {
		target = low
	} else if norm > high {
		target = high
	}
	// Pre-scaling avoids the fork Dgees vector-stride bug on extreme inputs.
	if target != norm {
		impl.Dlascl(lapack.General, 0, 0, norm, target, n, n, a.Data, a.Stride)
	}
	work := make([]float64, 1)
	impl.Dgees(lapack.SchurHess, lapack.SortNone, nil, n, a.Data, a.Stride, wr, wi, u.Data, u.Stride, work, -1, nil)
	work = make([]float64, int(work[0]))
	_, ok := impl.Dgees(lapack.SchurHess, lapack.SortNone, nil, n, a.Data, a.Stride, wr, wi, u.Data, u.Stride, work, len(work), nil)
	if target != norm {
		impl.Dlascl(lapack.General, 0, 0, target, norm, n, n, a.Data, a.Stride)
		impl.Dlascl(lapack.General, 0, 0, target, norm, n, 1, wr, 1)
		impl.Dlascl(lapack.General, 0, 0, target, norm, n, 1, wi, 1)
	}
	return ok
}

func congruence(c, u blas64.General, trans bool) {
	n := c.Rows
	work := blas64.General{Rows: n, Cols: n, Stride: n, Data: make([]float64, n*n)}
	sym := blas64.Symmetric{N: n, Stride: c.Stride, Data: c.Data, Uplo: blas.Upper}
	if trans {
		blas64.Symm(blas.Left, 1, sym, u, 0, work)
		blas64.Gemm(blas.Trans, blas.NoTrans, 1, u, work, 0, c)
	} else {
		blas64.Symm(blas.Right, 1, sym, u, 0, work)
		blas64.Gemm(blas.NoTrans, blas.Trans, 1, work, u, 0, c)
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			c.Data[j*c.Stride+i] = c.Data[i*c.Stride+j]
		}
	}
}

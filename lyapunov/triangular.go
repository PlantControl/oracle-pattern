package lyapunov

import (
	"math"

	"plantcontrol.org/v1/gonum/lapack/gonum"
)

type schurBlock struct{ start, size int }

func solveSchur(discrete, trans bool, n int, a []float64, lda int, c []float64, ldc int) (scale float64, perturbed bool) {
	scale = 1
	if n == 0 {
		return scale, false
	}
	const eps = 0x1p-52
	const safeMinimum = 0x1p-1022
	small := safeMinimum * float64(n*n) / eps
	big := 1 / small
	anorm := 0.0
	for i := 0; i < n; i++ {
		for j := max(0, i-1); j < n; j++ {
			anorm = math.Max(anorm, math.Abs(a[i*lda+j]))
		}
	}
	smin := math.Max(small, eps*anorm)
	var blocks []schurBlock
	for i := 0; i < n; {
		size := 1
		if i+1 < n && a[(i+1)*lda+i] != 0 {
			size = 2
		}
		blocks = append(blocks, schurBlock{i, size})
		i += size
	}
	if trans {
		for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
			blocks[i], blocks[j] = blocks[j], blocks[i]
		}
	}
	entry := func(i, j int) float64 {
		if trans {
			return a[j*lda+i]
		}
		return a[i*lda+j]
	}
	prior := func(b schurBlock) (int, int) {
		if trans {
			return b.start + b.size, n
		}
		return 0, b.start
	}
	blockMatrix := func(b schurBlock) []float64 {
		t := make([]float64, b.size*b.size)
		for i := 0; i < b.size; i++ {
			copy(t[i*b.size:(i+1)*b.size], a[(b.start+i)*lda+b.start:(b.start+i)*lda+b.start+b.size])
		}
		return t
	}
	impl := gonum.Implementation{}
	work := make([]float64, 2*n)
	for li, l := range blocks {
		lbegin, lend := prior(l)
		if discrete {
			for i := 0; i < n; i++ {
				for q := 0; q < l.size; q++ {
					sum := 0.0
					for j := lbegin; j < lend; j++ {
						sum += c[i*ldc+j] * entry(j, l.start+q)
					}
					work[q*n+i] = sum
				}
			}
		}
		for _, k := range blocks[li:] {
			kbegin, kend := prior(k)
			rhs := make([]float64, k.size*l.size)
			for p := 0; p < k.size; p++ {
				row := k.start + p
				for q := 0; q < l.size; q++ {
					col := l.start + q
					sum := 0.0
					if discrete {
						for i := kbegin; i < kend; i++ {
							sum += entry(i, row) * work[q*n+i]
						}
						for i := k.start; i < k.start+k.size; i++ {
							sum += entry(i, row) * work[q*n+i]
						}
						for j := 0; j < l.size; j++ {
							partial := 0.0
							for i := kbegin; i < kend; i++ {
								partial += entry(i, row) * c[i*ldc+l.start+j]
							}
							sum += partial * entry(l.start+j, col)
						}
					} else {
						left, right := 0.0, 0.0
						for i := kbegin; i < kend; i++ {
							left += entry(i, row) * c[i*ldc+col]
						}
						for j := lbegin; j < lend; j++ {
							right += c[row*ldc+j] * entry(j, col)
						}
						sum = left + right
					}
					rhs[p*l.size+q] = c[row*ldc+col] - sum
				}
			}
			x := make([]float64, len(rhs))
			localScale, localPerturbed := 1.0, false
			switch {
			case k.size == 1 && l.size == 1:
				denom := a[k.start*lda+k.start] + a[l.start*lda+l.start]
				if discrete {
					denom = a[k.start*lda+k.start]*a[l.start*lda+l.start] - 1
				}
				absDenom := math.Abs(denom)
				if absDenom <= smin {
					denom, absDenom, localPerturbed = smin, smin, true
				}
				absRHS := math.Abs(rhs[0])
				if absDenom < 1 && absRHS > 1 && absRHS > big*absDenom {
					localScale = 1 / absRHS
				}
				x[0] = (rhs[0] * localScale) / denom
			case k.size == 1 || l.size == 1:
				two, one := k, l
				if k.size == 1 {
					two, one = l, k
				}
				ca, wr := 1.0, -a[one.start*lda+one.start]
				if discrete {
					ca, wr = a[one.start*lda+one.start], 1
				}
				var ok bool
				localScale, _, ok = impl.Dlaln2(!trans, 2, 1, smin, ca, blockMatrix(two), 2, 1, 1, rhs, 1, wr, 0, x, 1)
				localPerturbed = !ok
			case k.start == l.start:
				if discrete {
					x, localScale, localPerturbed = smallDiscrete(trans, blockMatrix(k), rhs)
				} else {
					x, localScale, localPerturbed = smallContinuous(trans, blockMatrix(k), rhs)
				}
			default:
				if discrete {
					x, localScale, localPerturbed = smallProduct(!trans, trans, 2, 2, blockMatrix(k), blockMatrix(l), rhs)
				} else {
					var ok bool
					localScale, _, ok = impl.Dlasy2(!trans, trans, 1, 2, 2, blockMatrix(k), 2, blockMatrix(l), 2, rhs, 2, x, 2)
					localPerturbed = !ok
				}
			}
			perturbed = perturbed || localPerturbed
			if localScale != 1 {
				for i := 0; i < n; i++ {
					for j := 0; j < n; j++ {
						c[i*ldc+j] *= localScale
					}
				}
				for i := range work {
					work[i] *= localScale
				}
				scale *= localScale
			}
			for p := 0; p < k.size; p++ {
				for q := 0; q < l.size; q++ {
					c[(k.start+p)*ldc+l.start+q] = x[p*l.size+q]
					if k.start != l.start {
						c[(l.start+q)*ldc+k.start+p] = x[p*l.size+q]
					}
				}
			}
		}
	}
	return scale, perturbed
}

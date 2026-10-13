package lyapunov

import "math"

const smallEpsilon = 0x1p-52
const smallMinimum = 0x1p-1022 / smallEpsilon

func smallMaxAbs(a []float64) float64 {
	var m float64
	for _, v := range a {
		m = math.Max(m, math.Abs(v))
	}
	return m
}

func smallContinuous(trans bool, t, b []float64) ([]float64, float64, bool) {
	p, q := t[2], t[1]
	if trans {
		p, q = q, p
	}
	a := []float64{t[0], p, 0, q, t[0] + t[3], p, 0, q, t[3]}
	x, scale, perturbed := smallComplete(a, []float64{b[0] / 2, b[1], b[3] / 2}, math.Max(smallMaxAbs(t)*smallEpsilon, smallMinimum), 4)
	return []float64{x[0], x[1], x[1], x[2]}, scale, perturbed
}

func smallDiscrete(trans bool, t, b []float64) ([]float64, float64, bool) {
	p, q := t[2], t[1]
	if trans {
		p, q = q, p
	}
	a := []float64{t[0]*t[0] - 1, t[0]*p + t[0]*p, p * p, t[0] * q, t[0]*t[3] + t[1]*t[2] - 1, p * t[3], q * q, q*t[3] + q*t[3], t[3]*t[3] - 1}
	x, scale, perturbed := smallComplete(a, []float64{b[0], b[1], b[3]}, math.Max(smallEpsilon*smallMaxAbs(t), smallMinimum), 4)
	return []float64{x[0], x[1], x[1], x[2]}, scale, perturbed
}

func smallComplete(a, b []float64, smin, factor float64) (x []float64, scale float64, perturbed bool) {
	n := len(b)
	piv := make([]int, n-1)
	for i := 0; i < n-1; i++ {
		largest, ip, jp := 0.0, i, i
		for r := i; r < n; r++ {
			for c := i; c < n; c++ {
				if v := math.Abs(a[r*n+c]); v >= largest {
					largest, ip, jp = v, r, c
				}
			}
		}
		for c := 0; c < n; c++ {
			a[ip*n+c], a[i*n+c] = a[i*n+c], a[ip*n+c]
		}
		b[i], b[ip] = b[ip], b[i]
		for r := 0; r < n; r++ {
			a[r*n+jp], a[r*n+i] = a[r*n+i], a[r*n+jp]
		}
		piv[i] = jp
		if math.Abs(a[i*n+i]) < smin {
			a[i*n+i] = smin
			perturbed = true
		}
		for r := i + 1; r < n; r++ {
			a[r*n+i] /= a[i*n+i]
			b[r] -= a[r*n+i] * b[i]
			for c := i + 1; c < n; c++ {
				a[r*n+c] -= a[r*n+i] * a[i*n+c]
			}
		}
	}
	// SLICOT replaces the final pivot without changing INFO.
	if math.Abs(a[n*n-1]) < smin {
		a[n*n-1] = smin
	}
	scale = 1
	for i := 0; i < n; i++ {
		if factor*smallMinimum*math.Abs(b[i]) > math.Abs(a[i*n+i]) {
			scale = (1 / factor) / smallMaxAbs(b)
			break
		}
	}
	for i := range b {
		b[i] *= scale
	}
	x = make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		inv := 1 / a[i*n+i]
		x[i] = b[i] * inv
		for j := i + 1; j < n; j++ {
			x[i] -= (inv * a[i*n+j]) * x[j]
		}
	}
	for i := n - 2; i >= 0; i-- {
		x[i], x[piv[i]] = x[piv[i]], x[i]
	}
	return
}

func smallProduct(transl, transr bool, n1, n2 int, tl, tr, b []float64) (x []float64, scale float64, perturbed bool) {
	scale = 1
	n := n1 * n2
	if n == 0 {
		return []float64{}, scale, false
	}
	if n == 1 {
		tau := tl[0]*tr[0] - 1
		if math.Abs(tau) <= smallMinimum {
			tau = smallMinimum
			perturbed = true
		}
		if smallMinimum*math.Abs(b[0]) > math.Abs(tau) {
			scale = 1 / math.Abs(b[0])
		}
		return []float64{b[0] * scale / tau}, scale, perturbed
	}
	at := func(a []float64, n, i, j int, trans bool) float64 {
		if trans {
			i, j = j, i
		}
		return a[i*n+j]
	}
	a := make([]float64, n*n)
	rhs := make([]float64, n)
	for j := 0; j < n2; j++ {
		for i := 0; i < n1; i++ {
			row := i + j*n1
			rhs[row] = b[i*n2+j]
			for l := 0; l < n2; l++ {
				for k := 0; k < n1; k++ {
					col := k + l*n1
					a[row*n+col] = at(tl, n1, i, k, transl) * at(tr, n2, l, j, transr)
					if row == col {
						a[row*n+col] -= 1
					}
				}
			}
		}
	}
	smin := math.Max(smallMaxAbs(tl)*smallMaxAbs(tr)*smallEpsilon, smallMinimum)
	if n == 2 {
		// IDAMAX chooses the first maximum in column-major order.
		tmp := [4]float64{a[0], a[2], a[1], a[3]}
		ip := 0
		for i := 1; i < 4; i++ {
			if math.Abs(tmp[i]) > math.Abs(tmp[ip]) {
				ip = i
			}
		}
		u11 := tmp[ip]
		if math.Abs(u11) <= smin {
			u11 = smin
			perturbed = true
		}
		u12 := tmp[[4]int{2, 3, 0, 1}[ip]]
		l21 := tmp[[4]int{1, 0, 3, 2}[ip]] / u11
		u22 := tmp[[4]int{3, 2, 1, 0}[ip]] - u12*l21
		if math.Abs(u22) <= smin {
			u22 = smin
			perturbed = true
		}
		if ip == 1 || ip == 3 {
			rhs[0], rhs[1] = rhs[1], rhs[0]-l21*rhs[1]
		} else {
			rhs[1] -= l21 * rhs[0]
		}
		if 2*smallMinimum*math.Abs(rhs[1]) > math.Abs(u22) || 2*smallMinimum*math.Abs(rhs[0]) > math.Abs(u11) {
			scale = 0.5 / smallMaxAbs(rhs)
			rhs[0] *= scale
			rhs[1] *= scale
		}
		x = make([]float64, 2)
		x[1] = rhs[1] / u22
		x[0] = rhs[0]/u11 - (u12/u11)*x[1]
		if ip >= 2 {
			x[0], x[1] = x[1], x[0]
		}
		return
	}
	x, scale, perturbed = smallComplete(a, rhs, smin, 8)
	x[1], x[2] = x[2], x[1]
	return
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/jamestjsp/oracle-pattern/lyapunov"
	"plantcontrol.org/v1/gonum/blas"
	"plantcontrol.org/v1/gonum/blas/blas64"
)

const guard = -987654321.125

type input struct {
	ID      string   `json:"id"`
	Dico    string   `json:"dico"`
	Job     string   `json:"job"`
	Fact    string   `json:"fact"`
	Trans   string   `json:"trans"`
	N       int      `json:"n"`
	Pad     int      `json:"pad"`
	Invalid int      `json:"invalid"`
	A       []string `json:"a"`
	U       []string `json:"u"`
	C       []string `json:"c"`
}
type output struct {
	Info      int      `json:"info"`
	PaddingOK bool     `json:"padding_ok"`
	Scale     string   `json:"scale"`
	Sep       string   `json:"sep"`
	Ferr      string   `json:"ferr"`
	A         []string `json:"a"`
	U         []string `json:"u"`
	C         []string `json:"c"`
	WR        []string `json:"wr"`
	WI        []string `json:"wi"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
func run() error {
	var in input
	dec := json.NewDecoder(os.Stdin)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return errors.New("trailing wire input")
	}
	if in.N < 0 || in.N > 256 || in.Pad < 0 || in.Pad > 256 || len(in.Dico) != 1 || len(in.Job) != 1 || len(in.Fact) != 1 || len(in.Trans) != 1 {
		return errors.New("invalid wire dimensions or flags")
	}
	if in.Fact != "F" && in.Fact != "f" && in.Fact != "N" && in.Fact != "n" {
		return errors.New("invalid factor flag")
	}
	a, err := pack(in.N, in.Pad, in.A)
	if err != nil {
		return err
	}
	u, err := pack(in.N, in.Pad, in.U)
	if err != nil {
		return err
	}
	c, err := pack(in.N, in.Pad, in.C)
	if err != nil {
		return err
	}
	ac, uc, cc := a, u, c
	opts := lyapunov.Options{Equation: lyapunov.Equation(in.Dico[0]), Job: lyapunov.Job(in.Job[0]), Transpose: blas.Transpose(in.Trans[0]), SuppliedSchur: in.Fact == "F" || in.Fact == "f"}
	switch in.Invalid {
	case 0:
	case 1:
		opts.Equation = '?'
	case 2:
		opts.Job = '?'
	case 4:
		opts.Transpose = '?'
	case 5:
		ac.Rows = -1
	case 7:
		ac.Stride = 0
	case 9:
		uc.Stride = 0
	case 11:
		cc.Stride = 0
	default:
		return errors.New("unsupported invalid argument probe")
	}
	res, err := lyapunov.Solve(ac, uc, cc, opts)
	info := 0
	if err != nil {
		var arg lyapunov.ArgumentError
		if !errors.As(err, &arg) {
			return err
		}
		var ok bool
		info, ok = map[lyapunov.ArgumentError]int{"equation": -1, "job": -2, "transpose": -4, "order": -5, "a": -7, "u": -9, "c": -11}[arg]
		if !ok {
			return err
		}
	} else if res.NearSingular {
		info = in.N + 1
	}
	out := output{Info: info, PaddingOK: padding(a) && padding(u) && padding(c), Scale: format(guard), Sep: format(guard), Ferr: format(guard), A: unpack(a), U: unpack(u), C: unpack(c), WR: make([]string, in.N), WI: make([]string, in.N)}
	for i := range out.WR {
		out.WR[i] = format(guard)
		out.WI[i] = format(guard)
	}
	if err == nil {
		if in.N == 0 || (in.Job != "S" && in.Job != "s") {
			out.Scale = format(res.Scale)
		}
		if in.Job != "X" && in.Job != "x" {
			out.Sep = format(res.Sep)
		}
		if in.Job == "B" || in.Job == "b" {
			out.Ferr = format(res.Ferr)
		}
		if !opts.SuppliedSchur && in.N != 0 {
			for i := range out.WR {
				out.WR[i] = format(res.WR[i])
				out.WI[i] = format(res.WI[i])
			}
		}
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}
func format(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func pack(n, pad int, values []string) (blas64.General, error) {
	a := blas64.General{Rows: n, Cols: n, Stride: max(1, n) + pad}
	if len(values) != n*n {
		return a, errors.New("wire matrix length mismatch")
	}
	a.Data = make([]float64, max(1, n)*a.Stride)
	for i := range a.Data {
		a.Data[i] = guard
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			v, err := strconv.ParseFloat(values[i*n+j], 64)
			if err != nil {
				return a, err
			}
			a.Data[i*a.Stride+j] = v
		}
	}
	return a, nil
}
func unpack(a blas64.General) []string {
	out := make([]string, 0, a.Rows*a.Cols)
	for i := 0; i < a.Rows; i++ {
		for j := 0; j < a.Cols; j++ {
			out = append(out, format(a.Data[i*a.Stride+j]))
		}
	}
	return out
}
func padding(a blas64.General) bool {
	for i, v := range a.Data {
		if (i/a.Stride >= a.Rows || i%a.Stride >= a.Cols) && v != guard {
			return false
		}
	}
	return true
}

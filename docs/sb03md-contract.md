# SB03MD migration contract

`lyapunov.Solve` translates pinned SLICOT `SB03MD` into native Go using the PlantControl fork `plantcontrol.org/v1/gonum v0.20.8-0.20261013101244-9bd00a95089a`. The reference revision is `4a752c35ab540ce6291be8590dd7ed47086fe4ff`. This routine needs ordinary real Schur decomposition, not QZ.

## Operation and ownership

The continuous equation is `op(A)' X + X op(A) = scale C`; the discrete equation is `op(A)' X op(A) - X = scale C`. `op(A)` is A or its transpose. All matrices are real, square, and stored in row-major Gonum `blas64.General` descriptors.

```go
result, err := lyapunov.Solve(a, u, c, lyapunov.Options{
    Equation: lyapunov.Continuous,
    Job: lyapunov.Both,
})
```

A receives its real Schur form and U the orthogonal Schur vectors. C receives the symmetric solution. `SuppliedSchur: true` takes these factors as inputs instead and preserves them exactly. Only C's upper triangle is read; the solution fills both triangles. Supplied A entries below the first subdiagonal are ignored. Padding is preserved. Arrays must not overlap. Supplied factors must be valid; the solver does not numerically validate orthogonality or Schur structure.

| SLICOT behavior | Go mapping |
| --- | --- |
| DICO C / D | `Continuous` / `Discrete` |
| JOB X / S / B | `Solution` / `Separation` / `Both` |
| TRANA N / T / C | `blas.NoTrans` / `blas.Trans` / `blas.ConjTrans` |
| FACT N / F | `SuppliedSchur: false` / `true` |
| Case-insensitive flags | Case-insensitive equation, job, and transpose values |
| Scalar output arguments | `Result.Scale`, `Sep`, `Ferr` |
| WR / WI when FACT N | `Result.WR` / `WI`; nil for supplied factors |
| INFO N+1 | `Result.NearSingular`; perturbed solution still returned. Some final-pivot floors do not set this flag, matching SLICOT |
| Illegal argument, INFO < 0 | `ArgumentError`, before any matrix mutation |
| Schur nonconvergence | `ErrSchur`; the fork exposes a boolean, so the exact Fortran failure index is unavailable |
| Caller workspace and workspace query | Go owns allocation; no workspace argument or query |

Zero options select continuous, solution only, no transpose, and computed factors. Unrequested estimates are zero in Result; Scale is one for separation-only calls. The legacy routine leaves several unrequested scalar outputs untouched, including SCALE for nonempty JOB S; the replay adapter preserves those initialized sentinels. Empty valid calls return Scale=1 and requested Sep/Ferr=0.

Separation-only calls do not inspect C, including its descriptor. A/U descriptors must be square with enough storage even for supplied-factor separation. This is a deliberate stricter descriptor contract for U than Fortran's unused storage. Empty A/U descriptors still require stride >=1. Shapes and slice lengths are validated, unlike unsafe Fortran calls. The adapter pairs argument failures -1, -2, -4, -5, -7, -9, -11. Invalid FACT (-3) cannot be represented by Go's boolean; insufficient workspace (-19) cannot occur through this API. They are not claimed as paired cases.

## Dependency closure and algorithm

The native reference compiles unmodified `SB03MD`, `MB01RD`, `SB03MY`, `SB03MX`, `SB03MW`, `SB03MV`, and `SB04PX` against LP64 OpenBLAS BLAS/LAPACK. Its XERBLA hook records the handler argument instead of terminating on an invalid SB03MD call.

The Go implementation preserves the Schur method, structured block traversal, global overflow scaling, and pivot perturbations. `triangular.go` implements the continuous/discrete Schur solves. `small.go` implements the specialized symmetric 2x2 and discrete block-product systems; off-diagonal continuous blocks use Gonum DLALN2/DLASY2. The tiny complete-pivot systems have at most four unknowns. The production solver does not form a full Kronecker operator. Congruences use symmetric BLAS and preserve upper-triangle semantics. The separation estimator uses DLACN2 with its Fortran column-order vectorization, including nonsymmetric probes.

Two compatibility adaptations are necessary for the pinned fork:

- Its DGEES extreme-scale path passes the wrong row stride when rescaling eigenvalue vectors. The wrapper pre-scales into DGEES's safe interval and restores A/eigenvalues with valid row-major strides. Extreme inputs and padding have persistent tests.
- Its DLACN2 uses `Copysign` for sign probes, while LAPACK uses `x >= 0`. Canonicalizing zero before each DLACN2 call preserves LAPACK's treatment of negative zero. A compiled negative control removes this adaptation and must fail oracle replay.

These are local boundary adaptations; the fork itself is unchanged.

## Numerical acceptance

Active values and evaluated results must be finite. The corpus covers dimensions 0–9, real and complex Schur blocks, all equation/job/factor/transpose combinations, lowercase modes, rectangular physical padding, small/large right-hand sides, singular scalar and opposite-eigenvalue cases, overflow scaling, and extreme continuous Schur scaling. Public Go tests additionally poison ignored C entries with NaNs. Invalid Schur factors, overlapping arrays, nonfinite active inputs, arbitrary overflow/underflow, and universal accuracy for ill-conditioned problems are outside the claim.

For n>0, set `tau = 4096 * epsilon_float64 * n`. Solutions must agree in normalized Frobenius norm within tau; there is no unit absolute floor. Both solutions must be symmetric, and INFO=0 solutions must satisfy a separately evaluated, magnitude-normalized equation residual within tau. Perturbed cases compare the recorded outputs and status but do not assert a small residual for the original unperturbed equation. Supplied factors and unused data are compared exactly, including signed zeros. Output padding and statuses are exact.

Computed Schur factors are nonunique. They are checked by orthogonality, reconstruction of the original matrix, disjoint canonical 1x1/2x2 Schur block structure, and agreement of WR/WI with their 1x1/2x2 blocks. Reconstruction and orthogonality use `512 * epsilon_float64 * n`; eigenvalue checks use tau. Factor entries and eigenvalue ordering are not required to match another valid Schur basis.

Separation estimates are also basis-dependent: the routine estimates a reciprocal **1-norm** in Schur coordinates, not an invariant singular value. A cross-fed experiment isolated this in `grid-DSNT-6`:

| Same supplied factors | Fortran SEP | Go SEP |
| --- | --- | --- |
| Fortran-computed | 0.7172188503941771 | 0.7172188503941773 |
| Go-computed | 0.6017603253108240 | 0.6017603253108241 |

Accordingly, supplied-factor SEP/FERR retain tight relative comparisons within tau. For computed factors the legacy capture separately forms the Lyapunov Kronecker operator and records its smallest singular value with LAPACK DGESVD. Both implementations' SEP must lie in `[sigma_min/n, n*sigma_min]` with rounding allowance, and FERR must match `epsilon*||A||F/SEP` (continuous) or `epsilon*||A||F^2/SEP` (discrete) within tau. This interval is an engineering acceptance envelope motivated by norm equivalence; it is not a theorem guaranteeing the accuracy of the heuristic estimator. No bitwise or tight relative equality of estimates across different Schur bases is claimed.

The stored SVD value is computed exclusively by the native capture, before SB03MD mutates A. Go replay does not need LAPACK libraries, Fortran, NumPy, or a reference checkout. The full Kronecker/SVD computation is confined to the small-case capture harness.

## Recordings, provenance, and maintenance

The accepted default is `oracle/baselines/sb03md-v4`, policy `sb03md-float64-v2`. It records 500 calls: 300 deterministic cases plus 100 cases from each of seeds 20260911 and 20260912. Inputs, expected outputs, Schur state, scaling, diagnostics, and synthetic harness caller identity are recorded with source/library/binary hashes. These are synthetic isolated calls, not production traces or system-level scenarios.

The immutable v1/v2 captures preserve development evidence. Their original policy demanded tight computed-factor estimate equality; replay can report the demonstrated basis-dependent discrepancy. They are historical diagnostic captures, not the accepted default. A new policy and v3 capture added the independent SVD evidence without editing earlier expected results. V4 extends that accepted policy with 20 actual block-scaling, near-singular, and pivot cases; v3 remains a valid earlier accepted corpus.

`lyapunov/testdata/sb03mx-probes.json` additionally records 13 actual inner calls from the Fortran separation estimator, including nonsymmetric right-hand sides. Go unit tests replay them without the reference toolchain.

```sh
make verify
uv run python3 oracle/sb03md.py verify --case grid-DBFN-6
make audit
uv run python3 oracle/audit_sb03md_factors.py
uv run python3 oracle/sb03md.py capture --output oracle/baselines/sb03md-v5 --seed 20260913
uv run python3 oracle/capture_sb03mx.py --output /tmp/sb03mx-new-probes.json
```

Captures refuse existing outputs. Verification rejects missing/corrupt recordings, empty/duplicate calls, unsupported policies, malformed observations, unknown call IDs, and reports inside the baseline. Audits rebuild only the native reference; the separate cross-factor audit intentionally builds both implementations. Baseline hashes and capture seeds remain distinct from current candidate build identity. Compiled negative controls check transpose, scaling, signed-zero estimator behavior, and factor preservation.

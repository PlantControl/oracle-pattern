# Acceptance evidence

The capture-once MB01UD workflow passed on 2026-09-10. The translated Go algorithm is unchanged; this revision separates recorded legacy expectations from routine candidate verification.

## Results

| Gate | Result |
| --- | --- |
| Capture | 1,405 calls recorded from the pinned Fortran routine; no candidate execution |
| Offline Go verification | 1,405/1,405 passed |
| Explicit legacy audit | 1,405/1,405 passed |
| Single-call replay | asymmetric-LT passed; unknown call IDs rejected |
| Offline infrastructure checks | 13 passed |
| Native capture/audit check | 1 passed, including non-default seed preservation and overwrite refusal |
| Compiled negative controls | Omitted subdiagonal, ignored transpose, and input corruption all rejected |
| Go checks | Tests and vet passed |
| Skill validation | Bundled skill-creator metadata validator passed |
| Clean-copy verification | Full make verify passed without a reference checkout; FC and OpenBLAS flags unusable, CGO disabled, module network disabled |

The baseline has 605 deterministic cases plus 400 generated calls from each of seeds 20260910 and 20260911: **1,405 distinct calls**. Each trace identifies its scenario, harness caller, MB01UD callee, and invocation. No application caller instrumentation or stateful integration scenario is claimed.

Maximum numerical error/bound ratio in Go verification: `0.015473457034228295`. The legacy audit ratio was `0`. Comparison bounds and accepted domains are specified in [the contract](mb01ud-contract.md).

## Baseline identity

The checked-in [manifest](../oracle/baselines/mb01ud-v1/manifest.json) is the authoritative capture record; the adjacent compressed JSONL contains full inputs and expected observations.

- Calls SHA-256: `ece599f3cf93e26c3f5892b314a567a296835638160a22e486963c4afb84239d`.
- SLICOT revision: `4a752c35ab540ce6291be8590dd7ed47086fe4ff`.
- Capture: GNU Fortran (Homebrew GCC 15.2.0_1) 15.2.0; OpenBLAS 0.3.33; macOS arm64; one numerical-library thread.
- Candidate: go version go1.26.4 darwin/arm64; Gonum v0.17.0 native Go BLAS.

The manifest retains original capture source/library/binary hashes and compiler flags. Temporary build paths describe the capture invocation, not required permanent paths. Verification reports identify the baseline hashes and original seeds separately from current candidate build metadata. Reports can be regenerated with `make verify` and `make audit` under ignored `artifacts/`; binaries use automatically removed temporary directories.

## Refinement and cleanup

The former report-based replay rebuilt both implementations. It has been replaced by explicit `capture`, `verify`, and `audit` operations. Only capture creates expected results; existing baseline directories cannot be overwritten. Verification rejects missing/corrupt data, unsupported policies, empty/duplicate calls, invalid observations, and output paths inside the baseline.

The persistent regression for the previously discovered replay seed bug now checks that offline replay preserves the manifest's original seeds and identity. Integrity and negative-control checks operate on stored expected results. Native-toolchain checks are separate from everyday verification.

Removed superseded report/build artifacts, completed execution journals, the redundant acceptance-summary snapshot, and the two repeated live comparison campaigns. Retained the pinned reference, capture generator, curated edge cases, typed Go API checks, and numerical comparison policy because they support baseline extension and migration correctness.

This is a same-agent refinement trial with executable acceptance checks. It covers one synthetic harness-driven routine, not an entire control system, production workload, or performance claim. Output aliasing and unsupported nonfinite/overflowing computations remain outside the [declared domain](mb01ud-contract.md). Read the [playbook](../.agents/skills/oracle-migration/references/playbook.md) for the broader caller-capture and scenario-replay method and its published precedents.

## SB03MD migration — 2026-09-11

Implemented `lyapunov.Solve` and the SB03MY/SB03MX/SB03MW/SB03MV/SB04PX dependency behavior using native Go BLAS/LAPACK from the PlantControl fork `plantcontrol.org/v1/gonum v0.20.8-0.20261013101244-9bd00a95089a`. The [contract](sb03md-contract.md) describes supported equations, modes, ownership, errors, and numerical limits.

| Gate | Result |
| --- | --- |
| Accepted legacy capture | 500 calls from pinned SB03MD and its dependency closure |
| Go replay | 500/500 passed |
| Native audit | 500/500 passed, including separately captured DGESVD separation evidence |
| Internal solver recordings | 13/13 actual Fortran estimator probes passed in Go tests |
| Existing MB01UD replay and audit | 1,405/1,405 passed in each |
| Offline infrastructure tests | Existing 13 plus 9 SB03MD checks passed |
| Native capture checks | Existing 1 plus 2 SB03MD checks passed |
| Compiled negative controls | All four SB03MD defects rejected: ignored transpose, discarded scaling, signed-zero estimator mismatch, corrupted factors; existing three MB01UD controls still rejected |
| Go checks | `go test ./...`, `go vet ./...`, and `go test -race ./...` passed |
| Clean-copy verification | `make verify-offline` passed without reference sources, with CGO/module-network access disabled and unusable Fortran/library settings |
| Cross-factor audit | Fortran and Go agree when supplied the same factors from either decomposition |

The accepted capture is [sb03md-v4](../oracle/baselines/sb03md-v4/manifest.json), policy `sb03md-float64-v2`, with 300 deterministic cases and 100 random cases from each of seeds 20260911 and 20260912. Calls SHA-256: `0a95e0f230993d5a2622183926f7d316f537ee13df91ef16bb50e2ef8be57b9a`. The maximum reported tight numerical error/bound ratio was `0.0007321143622359089`; all separate SVD-envelope and structural checks passed. Twelve calls returned an actual scale below one, including 2x2 helper and global 4x4 scaling. These metrics are acceptance evidence, not a performance or universal accuracy claim.

The oracle exposed a fork compatibility discrepancy: DLACN2's use of `Copysign` treats negative zero differently from LAPACK's nonnegative comparison. Canonicalizing zero at the estimator boundary restores supplied-factor parity. The wrapper also avoids the fork DGEES eigenvalue-vector stride defect by managing extreme input scaling. Both adaptations have persistent regressions; the fork repository was not modified.

Computed Schur bases can differ validly and produce different reciprocal-1-norm separation estimates. Cross-fed factors gave matching Fortran/Go estimates (`0.7172188503941771`/`0.7172188503941773` for Fortran factors, `0.6017603253108240`/`0.6017603253108241` for Go factors). The accepted policy therefore checks computed-basis estimates against independently captured SVD bounds and checks the forward-error formula; supplied-factor estimates retain tight parity. Earlier immutable captures preserve the evidence leading to this policy and coverage extension.

SLICOT's silent final-pivot flooring is preserved: some regularization does not set INFO=N+1. Go's `NearSingular` mirrors that status rather than asserting that no pivot was changed. Exact Schur nonconvergence indices are unavailable through the fork's boolean API; Go reports `ErrSchur`. The corpus contains synthetic isolated scenarios, not production traces. No source changes were committed or published as part of this migration.

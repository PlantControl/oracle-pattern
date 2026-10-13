# Run record: MB01UD to Go with the code-modernization plugin

The recorded session behind the Claude Code edition, 2026-10-12 22:43 to 2026-10-13 02:01 CDT. Claude Code 2.1.296 (Opus 5.5), code-modernization plugin 1.0.0, Go 1.27.1, gfortran 15.2.0, OpenBLAS 0.3.33, macOS arm64.

## Workspace

```text
~/workspace/mb01ud-modernization/
  oracle-pattern/        clone of PlantControl/oracle-pattern at 5c1e946 (read-only during the run)
  mb01ud-legacy/         MB01UD.f + oracle/reference/driver.f90 + 2-line Makefile
  legacy/mb01ud -> ../mb01ud-legacy
  analysis/mb01ud/       plugin output: INTENT, PREFLIGHT, ASSESSMENT, rules, brief, VERIFICATION, REPORT.html
  modernized/            Go module plantcontrol.org/private/mb01ud and per-phase notes
```

`.claude/settings.json` denied edits to `legacy/`, `mb01ud-legacy/` and `oracle-pattern/` from the end of preflight onward.

## Human decisions

The plugin stops for a person at each of these points. The owner asked a separate Claude session to drive the recorded terminal and to enter these answers and signatures on the owner's behalf. Both signature lines say so. When the brief was signed, the recorded session noticed that it had not written the signature itself and asked who had. It continued only after that was confirmed.

| Step | Question | Answer given |
| --- | --- | --- |
| modernize | Goal | Rewrite in a new technology, one piece at a time |
| modernize | Must stay true | Old system keeps running; behavior must match exactly, including quirks |
| modernize | Target | Go + `plantcontrol.org/v1/gonum` |
| preflight 1 | Scope | One-routine pilot cut from SLICOT; keep the MB01UD call contract identical |
| preflight 2 | Build and test | `make` builds driver + MB01UD; 1,405-call baseline exists in `oracle-pattern/` |
| preflight 3 | Bespoke build | None in Fortran; Go BLAS must come from the PlantControl fork, no `replace` |
| preflight 4 | Prior attempts | Codex-built port passed the oracle; use its baseline as evidence, do not copy its code |
| preflight 5 | Off limits | `mb01ud-legacy/` and all of `oracle-pattern/` are read-only |
| review | RULE-001, 004, 005 | Right; keep the left/right NaN asymmetry |
| review | RULE-022 parity | Tolerance (oracle contract bound), no FMA needed; INFO, handler, NaN positions, H and padding exact; raise RULE-008/019/020 to P0 |
| brief | Q1-Q6 | Fortran-shaped API returning INFO; `modernized/mb01ud-transformed/`, private, local only; new non-finite corpus; no aliasing rejection; alpha=0 gives +0; amd64 must pass |
| brief | Approval | Full plan |
| transform P2 | Plan gate | Panic on short slices after INFO checks; pin `gonum.Implementation{}` |
| transform P3 | NaN/Inf parity | MB01UD-level: errors, inputs, padding and finite values match; NaN position differences reported |
| transform P3 | Differences | Approve the 463 non-finite cases as one group; fail if Go turns a legacy-finite value into NaN/Inf |
| verify | Sign-off | Accept with conditions: run `make test-amd64-v3` on x86-64-v3 hardware; resolve the Dtrmm follow-up (gonum ergo CEBBKI; signed as OXDUUL) |

## Results

| Measure | Value |
| --- | --- |
| Business rules | 22 (4 treated as P0 after review) |
| Recorded calls replayed | 1,405, exact through the harness |
| Argument probes from the unchanged Fortran | 1,788 |
| Live generated cases vs. Fortran | 1,200 finite + 1,200 non-finite |
| Worst error/tolerance ratio | 0.0076 |
| Plugin comparison | 3,805 cases: 3,342 same, 463 approved differences, 0 unapproved |
| Verify | PROVEN for all 3 modules; 441 tests, 0 failed; 20 fresh inputs; canary detected; legacy untouched |
| Independent re-run (outside the plugin) | `go test ./...` offline and with `MB01UD_LEGACY=1`: pass |

## Findings worth teaching

- **OpenBLAS decides NaN positions.** The legacy NaN/Inf positions come from OpenBLAS's blocked DTRMM, not from MB01UD's own logic (Go differed in 276 of 777 non-finite cases; a model of MB01UD's own logic differed in 215). Exact NaN parity would mean copying a library's internals. The owner narrowed the requirement and recorded the remaining differences as approved.
- **Architecture review caught a real bug.** At the float64 extremes, passing alpha into the triangular multiply produced wrong left-side results (for example 1e-30 instead of 1e-300). Scaling A first fixed it, and six edge cases became tests.
- **Gonum fork claim, unconfirmed.** The build reported that the fork's Dtrmm skips work when alpha·b underflows. A small reproduction on simple cases did not confirm it; gonum ergo task CEBBKI tracks the follow-up.
- **FMA changed the generator.** On arm64 the compiler fused `1 + 9*u`, so arm64 and amd64 generated different test corpora until an explicit rounding conversion was added.

## Recording gaps

The Mac's screen locked at 23:39 CDT. Window capture stopped partway through Phase 1 of transform and resumed at 01:55, so Phases 2 and 3 have no screen footage. Chapter 7 shows those phases as panels quoting the saved Claude Code session transcript, each marked "From the saved session record". The verify run (chapter 8) was recorded for its last four minutes, after the screen was unlocked; the verdict, sign-off and status are on screen.

The workspace is not part of this repository. Rerunning the lesson produces a new run with its own numbers; model output differs between runs.

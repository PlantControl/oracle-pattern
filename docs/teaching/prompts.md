# Reusable Oracle Pattern prompts

Paste these into Codex CLI from the repository root. The first four work on a fresh clone. Each prompt gives the agent a bounded task and an evidence-based stopping condition. The commands are repository-specific; the reasoning applies to other coding-agent CLIs.

## 1. Orient me before changing anything

```text
Read README.md, docs/mb01ud-contract.md, and the baseline manifest in
oracle/baselines/mb01ud-v1/. Explain the oracle pattern in five short points:
what ran originally, what is recorded, what runs during verification,
what counts as a mismatch, and what this pilot does not cover.
Name the files supporting your answer. Do not change files or recapture data.
```

## 2. Replay one recorded call

```text
Read README.md. Run:
python3 oracle/run.py verify --case asymmetric-LT --report artifacts/replay.json
Inspect the generated report. Report the selected call ID, pass/fail,
and which baseline was used. Distinguish actual command results from
inferences. Do not edit source, expected results, or comparison tolerances.
If execution fails, report the exact failure rather than claiming a pass.
```

Change `asymmetric-LT` to another recorded ID, such as `asymmetric-RT`. An unknown ID must fail; an empty selection is not success.

## 3. Verify the full replacement

```text
Run make verify. Summarize the Go checks, infrastructure checks, and full
recorded-corpus result. Inspect test_rejects_real_candidate_mutations in
oracle/test_oracle.py and explain what a passing negative-control test means.
State the observed counts and coverage limits. Preserve the baseline and
source. Do not install a legacy toolchain or run capture to make replay pass.
```

## 4. Investigate a mismatch

Replace the bracketed fields before submitting.

```text
Investigate failing call [CALL_ID] in [REPORT_PATH]. Read the contract first.
Replay only that call and identify its recorded inputs, expected observation,
actual observation, and baseline identity. Check marshalling/layout,
translation logic, mutation/status, and the comparison policy in that order.
Explain the supported cause with evidence. If it is a candidate defect,
make the smallest repair and add a persistent regression. Run the focused
regression and make verify. Do not change expected results or widen tolerances
to obtain a pass. If the cause remains uncertain, state what evidence is missing.
Do not commit, push, or publish.
```

## 5. Apply the pattern to another migration

Fill in the routine, paths, and target language. In this repository, the local skill can be explicitly invoked with `$oracle-migration`. In another harness, provide the skill/playbook as context using that harness's supported mechanism.

```text
Use the oracle migration method for [LEGACY_ROUTINE_OR_FEATURE] in
[LEGACY_SOURCE_PATH], replacing it with [TARGET_LANGUAGE_AND_API].

First inspect the real implementation, callers, dependencies, documentation,
and existing tests. Choose one bounded pilot. Define inputs, outputs, status,
mutation, aliasing, layout, and accepted domain. For stateful behavior, include
initial state and ordered events. Record whether cases come from production
callers or a synthetic harness.

Create an independent versioned baseline by running the pinned legacy program.
Record exact inputs, observations, source/toolchain identity, and checksums.
Never use the replacement to generate expected results or overwrite an existing
baseline. If the legacy program cannot run, report the blocker before claiming
an oracle exists.

Implement the replacement and replay the stored inputs against it. Compare
discrete behavior exactly and numerical behavior with a justified policy.
Add persistent regressions and executable negative controls. Demonstrate one-call
replay and everyday verification without running the legacy executable.

Finish with commands, observed results, coverage limits, and remaining gaps.
Do not commit, push, or publish.
```

A model's confident explanation is not an acceptance result. Keep the contract and recorded expectations independent of the implementation being judged.

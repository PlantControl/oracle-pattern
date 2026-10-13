# Reusable prompts and answers

For Claude Code with the code-modernization plugin, run from the workspace root. The plugin's own commands carry most of the instructions; these are the answers and follow-up prompts that kept the run tied to the oracle.

## 1. Front door

```text
/code-modernization:modernize <system> --source <path-to-legacy-slice>
```

Choose a rewrite, "the old system keeps running", and "behavior must match exactly" only if you mean it. Name the target precisely, including required libraries and forbidden ones.

## 2. Preflight answers that protect the oracle

Type these into the pop-ups ("Type something"), adapted to your system:

```text
Scope: A one-routine pilot cut from <library>. Keep the <routine> call contract (arguments, error codes, outputs) identical; other routines are out of scope this pass.
Build & test: Yes. <build command> builds it locally in seconds. A recorded legacy baseline also exists in <baseline path>.
Prior attempts: <earlier port> exists. This pass is an independent rewrite: use its recorded baseline as evidence, do not copy its code.
Off limits: Read-only: <legacy dir> and everything in <oracle repo> (its baseline is the oracle; its code is not to be read or copied).
```

Then add the deny rules preflight prints to `.claude/settings.json` before running anything else.

## 3. A parity policy the reviewer can sign

When `review` asks how exact the match must be:

```text
Parity is a tolerance, not bit-exact: use the scale-aware bound in <contract doc> (<formula>; zero-scale cases exact). Error codes, the reported argument, NaN/Inf positions, and untouched inputs and padding must match exactly. Also raise <argument-check and quick-return rules> to P0.
```

## 4. Approve a plan

Answer the brief's open questions in `analysis/<system>/MODERNIZATION_BRIEF.md`, tick them, and fill in `Approved by:`. A chat message is not approval. If someone else enters your signature, say so on the line.

## 5. When the oracle disagrees with your rule

```text
Before deciding, measure the gap: run the legacy, the candidate and a model of the routine's own logic over a non-finite case set, and show where they differ and why.
```

Then decide what counts, write it into the exit criterion, and approve the remaining differences as a listed group. Do not widen a tolerance to make them disappear.

## 6. Independent proof

```text
/clear
/code-modernization:modernize-verify <system>
```

Read `analysis/<system>/VERIFICATION.md`: per-module verdict, what passed, what it does not prove, and a blank sign-off for a named person.

## 7. Check it yourself

```sh
cd modernized/<system>-transformed && go test -count=1 ./...
MB01UD_LEGACY=1 MB01UD_LEGACY_SRC=$PWD/../../legacy/mb01ud go test -count=1 ./...
```

A model's confident summary is not an acceptance result. The saved test output and the comparison files are.

# Oracle Pattern: from clone to verified replay

A guided Terminal walkthrough for developers who can run shell commands. Learn to identify the behavior contract, replay recorded legacy behavior against a replacement, and distinguish evidence from an agent's explanation.

> **New:** the [Claude Code edition](claude-code/README.md) rewrites `MB01UD` in Go with the code-modernization plugin, in nine chapter videos. This page is the Codex CLI edition, which verifies an existing translation.

The demonstration uses Codex CLI. The repository commands and [reusable prompts](prompts.md) can also be used with another coding agent's CLI; adapt its launcher, permissions, and skill-loading syntax.

The selected edition is **9:00**, with AI-generated narration using OpenAI's Cedar voice, real Codex CLI screen footage, readable evidence panels, and captions. Use the [spoken transcript](transcript.md), [interactive questions](explore-with-ai.md), and [reusable prompts](prompts.md) to follow along. [Reference and editing notes](reference-review.md) explain the approach.

## Start from scratch

Install Git, Go 1.24+, Python 3.10+, and Make. For the agent demonstration, install and sign in to Codex CLI using the [official setup instructions](https://learn.chatgpt.com/docs/codex/cli). The recording used Codex CLI 0.153.3. Check your installation with `codex --version`.

Run these commands in a terminal. The first Go dependency download may need network access. Verification requires no Fortran compiler, OpenBLAS installation, or reference submodule checkout.

```sh
mkdir -p ~/workspace
cd ~/workspace
git clone https://github.com/jamestjsp/oracle-pattern.git
cd oracle-pattern
make verify
```

Expected corpus summary: `calls: 1405`, `passed: 1405`, `failed: 0`. Earlier output includes Go tests, vet, and 13 Python checks. The negative-control test passes when deliberately broken candidates are rejected.

## Use the source ZIP

Extract the supplied `oracle-pattern` folder into your workspace. Open a terminal in that folder and run `make verify`, then start `codex` and follow the questions below. The source ZIP includes the committed project, stored baseline, and pinned SLICOT reference source. The video is supplied separately.

The ZIP contains source files without Git history. For a new legacy capture or audit, use the Git clone workflow and initialize the reference submodule, since those operations record and verify Git provenance.

## Explore with the AI

Follow the [interactive question sequence](explore-with-ai.md). Start with what the repository does, ask where its expected answers come from, trace one execution path, and then run an experiment. Ask one question at a time and inspect the cited files and results.

## Follow the screen navigation

| Step | Navigate to | Look for |
| --- | --- | --- |
| Contract | `docs/mb01ud-contract.md` | Outputs, unchanged inputs, errors, row/column layout |
| Baseline | `oracle/baselines/mb01ud-v1/manifest.json` | 1,405 calls, comparison policy, checksum, original provenance |
| Recordings | `oracle/baselines/mb01ud-v1/calls.jsonl.gz` | Stored inputs, expected observations, harness caller identity |
| Replacement | `hessenberg/product.go` | Go implementation under verification |
| Replay | `artifacts/replay.json` after the command below | Selected call, actual result, failures, baseline identity |
| All checks | `make verify` | Passing corpus and rejected broken candidates |

Start an interactive agent session:

```sh
codex
```

Paste the replay prompt from [prompts.md](prompts.md). For a single noninteractive run, the video's equivalent is:

```sh
codex exec --color never -s workspace-write 'Read README.md. Run python3 oracle/run.py verify --case asymmetric-LT --report artifacts/replay.json. Report the call ID and pass/fail in two short sentences. Do not edit source or baseline. Do not use subagents.'
```

You can also run the acceptance command directly:

```sh
python3 oracle/run.py verify --case asymmetric-LT --report artifacts/replay.json
```

Expected: one call passed, zero failed. Reports are generated under ignored `artifacts/`; they are not expected to exist immediately after cloning.

## Practice

Use the replay prompt with `asymmetric-RT`, then use the full-verification prompt. Explain why a candidate must not generate its own expected answers. Identify one behavior outside this pilot's coverage.

Success means both replay and full verification pass, and you can explain that these are synthetic harness recordings for one numerical routine. Passing does not prove all possible inputs, a whole migrated application, or performance. For stateful migrations, add initial state and ordered scenario replay.

The original recording script and reusable retake guidance are in [recording guidance](voice-recording.md).

## Video and production notes

The finished local deliverables are:

- `artifacts/teaching/oracle-pattern-cedar.mp4` — 1080p, 30 fps, 9:00, with eight chapter markers and visible captions.
- `artifacts/teaching/oracle-pattern-cedar.srt` — separate timed captions.
- `docs/teaching/transcript.md` — chaptered narration transcript.

The lesson starts with a fresh clone and follows an interactive investigation, including a deliberate defect, failed replay, and verified repair. It demonstrates verification of an existing translation. The final result is 1,405 passing recorded calls, with unchanged baseline files.

Cedar was selected after comparing two generated voice samples. The final edit uses the written lesson script, natural speech delivery, and additional reading pauses. Screen transitions and captions are timed to this narration. Setup and inactive agent waits are edited; selected prompts and results are held for reading.

The earlier 11:22 human-narrated edition remains at `artifacts/teaching/oracle-pattern-human-narrated.mp4`. The eight original voice takes remain in `artifacts/teaching/revision/voice-takes/`, and their [transcript](human-transcript.md) is preserved. The first macOS synthetic-voice draft is superseded.

Media and editing sources are local, ignored artifacts; they are not committed or hosted. The demonstrated public repository revision is `5c1e946`; this teaching documentation is a subsequent local addition.

For baseline creation and explicit legacy audits, continue with the [migration playbook](../../.agents/skills/oracle-migration/references/playbook.md). Those operations have additional native toolchain requirements and are outside the short replay lesson.

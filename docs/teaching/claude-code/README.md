# Oracle Pattern with Claude Code

A nine-chapter video lesson and handout. Claude Code and the [code-modernization plugin](https://github.com/anthropics/claude-plugins-official/tree/main/plugins/code-modernization) rewrite the SLICOT Fortran routine `MB01UD` in Go. The plugin plans, builds and proves the rewrite. The oracle (recorded and live legacy behavior) keeps the proof independent of the code being judged. A person makes the decisions the plugin stops for.

This edition replaces the agent in the [Codex CLI edition](../README.md), which verified an existing translation. Here the translation itself is produced on camera, from intent to a signed proof.

## Chapters

Each chapter is its own video, with matching captions.

| # | Chapter | Shows | Video |
| --- | --- | --- | --- |
| 1 | Why an oracle? | The problem and the pattern | `chapter-01.mp4` |
| 2 | Set up the workspace | Clone, legacy slice, plugin install | `chapter-02.mp4` |
| 3 | Say what you want, then preflight | `/modernize`, five human answers, edit protection | `chapter-03.mp4` |
| 4 | Assess and map | Size, quirks, structure | `chapter-04.mp4` |
| 5 | Mine the rules, then review them | 22 rules; tolerance policy decided by a person | `chapter-05.mp4` |
| 6 | The plan you approve | Three-phase brief, six questions, signature | `chapter-06.mp4` |
| 7 | Build in three phases | Harness, routine, non-finite parity | `chapter-07.mp4` |
| 8 | Prove it | Independent verify, PROVEN, human sign-off | `chapter-08.mp4` |
| 9 | Try it yourself | Apply the pattern to your code | `chapter-09.mp4` |

Videos and captions are local deliverables in `artifacts/teaching/claude-edit/final/` (ignored, not committed). Narration is AI-generated with OpenAI's Cedar voice over real Claude Code footage. Read the [narration script](narration.md), the [run record](run-record.md) (answers given, results, recording gaps) and the [reusable prompts](prompts.md).

## Prerequisites

- Claude Code, signed in
- Git, Go 1.24+, Python 3.10+, Make
- gfortran and OpenBLAS (`brew install gcc openblas` on macOS), so the plugin can run the legacy routine and compare live

A run like the recorded one takes about three hours of agent time and many subagents. Start with this one small routine.

## Follow along

```sh
mkdir -p ~/workspace/mb01ud-modernization && cd ~/workspace/mb01ud-modernization
git clone https://github.com/PlantControl/oracle-pattern.git
git -C oracle-pattern submodule update --init --depth 1
mkdir mb01ud-legacy
cp oracle-pattern/reference/SLICOT-Reference/src/MB01UD.f oracle-pattern/oracle/reference/driver.f90 mb01ud-legacy/
printf 'mb01ud: MB01UD.f driver.f90\n\tgfortran -O2 -fcheck=all -fno-fast-math driver.f90 MB01UD.f -L$$(brew --prefix openblas)/lib -lopenblas -o mb01ud\n' > mb01ud-legacy/Makefile
claude
```

Inside Claude Code:

```text
/plugin install code-modernization@claude-plugins-official
/code-modernization:modernize mb01ud --source mb01ud-legacy
```

Then follow the "Next" line each command prints. The order is:

1. preflight
2. assess
3. map
4. extract-rules
5. review
6. brief: approve it by editing and signing the file
7. transform: one run per phase
8. verify, in a fresh session or after `/clear`

Our answers at every stop are in the [run record](run-record.md); yours may differ.

After preflight, block edits to the legacy source and the oracle, as preflight recommends. Put this in `.claude/settings.json` at the workspace root:

```json
{ "permissions": { "deny": ["Edit(/legacy/**)", "Edit(/mb01ud-legacy/**)", "Edit(/oracle-pattern/**)"] } }
```

## What to watch for

- **Evidence versus opinion.** Verdicts come from `proof_pack.py` and saved test results, not from the model's summary.
- **Recordings stay independent.** The 1,405-call baseline and the live Fortran come from the legacy, never from the Go candidate.
- **Tolerance is a decision.** Finite values use the scale-aware bound from the [MB01UD contract](../../mb01ud-contract.md). Errors, untouched inputs and padding are exact.
- **Differences are listed, not hidden.** OpenBLAS internals decide where the legacy puts NaNs. A person accepted those 463 cases as a group, and they stay listed.
- **Approval is a person's.** The brief and the proof each carry a human signature. In the recording, the owner had another Claude session enter them; the record says so, and the plugin asked about it.

## Limits

This is one numerical routine with synthetic and generated inputs, not a migrated application. PROVEN covers what was tested: arm64, and amd64 under emulation. It does not cover x86-64-v3 hardware or every possible input. A stateful system also needs initial state and ordered scenario replay; see the [migration playbook](../../../.agents/skills/oracle-migration/references/playbook.md).

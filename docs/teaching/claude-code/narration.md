# Claude Code edition: narration script

Narration source for the nine chapter videos. Each `## N. Title` section becomes one MP4. Bracketed lines are editing directions and are not spoken. Facts come from the recorded run on 2026-10-12/13; see [the run record](run-record.md).

## 1. Why an oracle?

Imagine an old program that works, and you want it in a modern language. A coding agent can translate it. The new code compiles, and the tests pass.

But who wrote those tests? If the same agent wrote the code and invented the expected answers, it can make the same mistake twice. A green result would prove very little.

The oracle pattern adds a second source of truth. We run the original program, record what it actually does, and replay those recordings against the replacement.

In this series, Claude Code and the code modernization plugin rewrite one Fortran routine from the SLICOT library in Go. The plugin plans, builds, and proves the rewrite. The oracle keeps that proof honest.

Each chapter is a short video. You will see the real session, with long waits removed.

[Title card, then the three steps: record, replay, compare.]

## 2. Set up the workspace

We start in an empty folder. First, we clone the oracle pattern repository. It holds a pinned copy of the SLICOT source and a recorded baseline of fourteen hundred and five calls to the routine we will rewrite.

Next, we cut a small slice for the plugin to work on: the routine M B zero one U D, the Fortran driver that records calls, and a two-line Makefile.

Then we start Claude Code and install the code modernization plugin from the official marketplace. It adds the modernize commands and a set of specialist agents.

You need Git, Go, Python, gfortran, and OpenBLAS. Because the legacy routine can build and run here, the plugin can compare old and new directly.

[Hold on the slice listing and the install confirmation.]

## 3. Say what you want, then preflight

The front door is the modernize command. It asks two short questions. We choose a rewrite in a different technology, and we say what must stay true: the old routine keeps running, and behavior must match exactly, including its quirks. The target is Go, using the PlantControl fork of Gonum for linear algebra.

Preflight then asks five questions that only a person can answer. Is this the whole system or a slice? Can it build here? Is there unusual build machinery? Has anyone tried this before? What is off limits?

We tell it about the earlier Codex translation, and that its Go code must not be read or copied. Only its recorded baseline counts as evidence.

Preflight checks the tools, proves the legacy builds, and recommends one fix: block edits to the legacy source and to the oracle. We add that rule before going further.

[Hold on the five answers and the deny rule.]

## 4. Assess and map

Assess asks: what are we dealing with? This routine is small, two hundred twenty-three lines. The risk is not size. It is numerical behavior.

The assessment finds the quirks a rewrite must respect. The left and right versions treat zero entries differently, so infinities and NaNs can spread on one side but not the other. The routine briefly changes its input matrix and then puts it back. On this Mac, the compiler fuses some multiply-add steps, which changes the last bits of a result.

Map draws the structure: the calls into BLAS, the argument checks that produce each error code, and the flows a caller sees.

None of this changes code. It tells us what to protect.

[Hold on the quirks list and the map.]

## 5. Mine the rules, then review them

Extract rules turns the code into twenty-two testable rules, each written as given, when, then, and each tied to a line in the source. A second agent checks every citation.

Some rules need a person. Should the Go version copy the NaN difference between sides? We asked for an exact match, so we keep it. Must every result match bit for bit? No. We set a tolerance: the same scale-aware bound the oracle repository documents. Error codes, untouched inputs, and NaN positions must still match exactly.

We also raise the argument-checking and quick-return rules to the highest priority, because the caller's contract depends on them.

Your answers are stored. The plan and the build read them, so nobody asks twice.

[Hold on the review questions and the typed tolerance note.]

## 6. The plan you approve

Brief writes the plan: three phases. First, an equivalence harness. Then the routine itself. Then the non-finite cases and a self-contained build. Each phase has entry and exit criteria, and editing them steers the build.

Nothing is built until a person approves. The approver answers six open questions in the file, such as the shape of the Go function and where the module lives, and signs the approval block.

In this recording, a separate Claude session entered the owner's answers and signature at the owner's request. Watch what the plugin does: it notices that it did not write that signature and asks who did. A person owns this decision, and the record says how it was made.

[Hold on the signed approval block and the plugin's question.]

## 7. Build in three phases

Transform builds one phase at a time. Each starts with a plan to approve, then a test engineer writes the tests before any code.

Phase one builds the harness. It replays all fourteen hundred and five recorded calls exactly, and adds seventeen hundred eighty-eight argument probes captured from the unchanged Fortran. Deliberately broken routines must fail it.

Phase two writes the routine in Go. It passes the recorded calls, the argument probes, and twelve hundred fresh cases run live against the Fortran.

Phase three tests infinities and NaNs. Here the oracle surprised us. The legacy NaN positions come from OpenBLAS internals, not from the routine's own logic. A person decided what counts: errors, inputs, and every finite value must match, and the remaining NaN position differences are approved and listed, not hidden.

[Part of this chapter is shown from the saved session record because the screen recording stopped. Label those panels.]

## 8. Prove it

Verify runs after the build, ideally in a fresh session. It does not trust the build's notes. It deletes build output, reruns the full suite, compares old and new on the same inputs, and invents twenty inputs nobody used.

It also breaks the code on purpose. A canary change must turn tests red, or the tests cannot be trusted.

A script, not the model, computes the verdict. All three modules are proven: four hundred forty-one tests run, none failed, every development case accounted for, and the legacy source untouched.

Proven is evidence, not approval. A named person signs. Here the owner accepted with two conditions: run the tests on newer x86 hardware, and check a possible bug the build reported in the Gonum fork.

[Hold on the verdict table and the signed block.]

## 9. Try it yourself

Let us bring this back to the pattern.

Choose a behavior you can observe. Record it from the real legacy system. Keep those recordings independent of the rewrite. Then let the plugin plan, build, and prove, while a person makes the decisions it stops for.

When something differs, do not widen the tolerance quietly. Find out why. Sometimes the old system is doing something you never intended to keep. Decide, write it down, and list the difference.

The handout has the commands, the answers we gave, and prompts you can reuse. Try it on one small routine first. Then explain where each expected answer came from. If you can, you have the oracle pattern.

[End on the reusable prompts and the practice task.]

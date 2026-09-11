# Cedar narration transcript

Runtime: 9:00. AI-generated narration using OpenAI Cedar, selected after a two-voice audition. The screen footage records actual Codex CLI use. Captions follow the generated speech; paragraph punctuation is normalized.

## 00:00 — Why use an oracle?

Imagine you have an old program that works, and you want to replace it. You ask a coding agent to translate it. The new code looks reasonable. It compiles. The tests pass.

But where do those tests come from? If the same agent wrote the implementation and invented all the expected answers, it could make the same mistake in both places. A green test result would not tell you much.

The Oracle pattern gives us another source of evidence. We run the original program, record what it actually does, and keep those observations. Then we give the same inputs to the replacement and compare the results.

That is the idea we will make concrete here. We will clone a small working example, inspect one recording, ask Codex to replay it, and deliberately break the replacement. Then we will watch the recorded behavior help us find and fix the mistake.


## 00:54 — Clone the example

You can follow along from an empty working directory. The commands are in the handout, so you do not need to copy them from the screen. You will need Git, Go, Python, and Make. For the agent steps, install Codex CLI and sign in first. The handout links to the setup instructions and lists the required versions.

Here, I am cloning the public Oracle pattern repository and moving into it. This repository already contains a completed translation and recordings from the original program. We are learning how to examine and verify that migration. We are not pretending that we just created the translation during this recording.

You also do not need a Fortran compiler for this lesson. That is one benefit of keeping the original observations. Ordinary verification can run without rebuilding the legacy system. The first Go dependency download may still need an internet connection.


## 01:48 — Explore with the AI

Before we ask the agent to do anything, we need to decide what we are preserving. Our example is a numerical routine from SLICOT. It multiplies matrices. You do not need to understand the matrix algorithm to follow the verification process.

First, let us ask Codex. What problem does this repository solve? And which three files should I read first? We are using the AI to build a map of an unfamiliar codebase.

Ask it to inspect the files and explain them in plain English. Follow its answer into the contract. The contract describes the output, the inputs that must remain unchanged, the errors the caller can observe, and the supported storage layout.

That last point matters. The old Fortran routine and the new Go function store matrices differently. We are preserving the meaning of a call while allowing the implementation to change. For your own application, this contract might include a written record, an account balance, an error code, or an updated controller state.

Start at a boundary that you can run and observe. Here is a useful question to stop on. If the replacement changes something, would its caller notice? That is the behavior your verification needs to cover.


## 03:04 — Inspect a legacy recording

Our next question is, where do the expected answers come from? Ask Codex to show one actual recording, rather than summarize the idea from memory. Now follow its answer into the baseline directory. The manifest identifies the original capture and the comparison policy. The compressed data file contains the individual recordings.

Each one has a call identifier, inputs, and an expected observation from the legacy program. We will use the call named asymmetric L T. You can see its expected output values here. Those numbers were recorded before we run the current Go implementation.

Keep that separation in mind. The candidate does not get to decide what the expected answer should be. This baseline contains 1 ,405 calls. They are synthetic harness cases for this one routine, not recordings from a production application.

That gives us useful coverage, but it does not tell us that every possible input or every connected system will work. Also notice the checksum and source identity. They help us identify exactly which baseline we are testing against. If the data is missing or corrupt, verification should fail rather than quietly create a new set of answers.


## 04:17 — Replay one call

We have used the agent to explore. Now ask it to trace one call through the adapter, the replacement function, and the comparison. Once that path makes sense, ask, can we test just this one call?

The prompt is deliberately narrow. Replay this call, inspect the report, and tell us whether it passed. Preserve the source and the baseline. I will leave the prompt on screen for a moment. You can reuse it from the handout.

Codex runs the same Python command that you could run yourself. The verifier loads the stored inputs, builds the Go candidate, executes it, and compares the observations. Here is the result. One call passed, with zero failures.

The important evidence is the command result and its report. The agent helps us navigate and interpret it, but the agent's confidence is not what makes the replacement correct. There is one numerical detail worth understanding. Floating point calculations can differ slightly because of evaluation order. This repository uses a documented, scale-aware comparison policy. We do not make a failing result pass by simply widening the tolerance.


## 05:28 — Detect a real defect

A successful replay is useful. A successful replay of a known bad program would be a problem. First, ask Codex how the repository checks that the verifier can detect a mistake. Let it find the existing negative controls and explain one. Then, in this disposable teaching checkout, ask it to deliberately change that one line.

The replacement will ignore part of the matrix that it should include. The recorded inputs and expected results stay untouched. Now we run exactly the same call again. This time, it fails. The first recorded output should be 9.5.

The broken replacement produced minus 1. That is a missing contribution, not a small rounding difference. Stop here and look at what changed. We did not change the question or the expected answer. We changed the candidate and the verifier noticed a difference. The report gives us the call identifier and a mismatching observation.

That makes the problem reproducible. We can investigate one small case rather than guessing from a large application failure. This is also why we keep tests for deliberately broken candidates. They check whether our verification can detect representative mistakes. A detector that always says yes is not a useful detector.


## 06:51 — Repair and verify

Here is the repair prompt. Investigate this failing call, use the contract and recorded observations, and make the smallest candidate fix. Preserve the baseline and comparison policy. Codex can now work from a concrete failure.

After the repair, we run the focused replay again and inspect the change. The call passes again, but one passing call is not enough to finish. Next we run make verify.

That includes the Go checks, the verification infrastructure checks, the deliberately broken candidates, and the complete recorded corpus. The final summary shows 1 ,405 passing calls and zero failures. We also check that the baseline has not changed.

In your own migration, retain a persistent regression for any defect you discover. This repository already has a negative control for the mistake we introduced.


## 07:55 — Try it yourself

Let us bring that back to the pattern. Choose a behavior you can observe. Capture it from the real legacy system. Keep the inputs and expected observations independent of the rewrite. Then use those recordings throughout development.

When something fails, replay the identified call, investigate the difference, and repair the candidate. Add new coverage through a deliberate new legacy capture. For a stateful application, a single call is only part of the story.

You also need the initial state and the order of events, followed by a replay of the connected system. The handout includes this question sequence, plus prompts for discovery, one -call replay, full verification, and mismatched diagnosis. You can adapt them to another coding agent L .I. The commands and launch options will differ.

The need for independent evidence stays the same. Before moving on, try another recorded call yourself. Then explain where its expected answer came from. If you can do that, you have the central idea of the oracle pattern.

Continue with the [interactive questions](explore-with-ai.md), [reusable prompts](prompts.md), and [from-scratch guide](README.md). The previous [human narration transcript](human-transcript.md) is preserved.

# Finished video transcript

Human narration, edited from eight recordings. Runtime: 11:22. Punctuation and product names are normalized from the timed captions; the recording script is [also available](narration.md).

## 00:00 — Why use an oracle?

Imagine you have an old program that works and you want to replace it. You ask a coding agent to translate it. The new code looks reasonable. It compiles. The test passes. But where did those tests came from?

If the same agent wrote the implementation and invented all the expected answers, it could make the same mistake in both places. A green test result would not tell you much. The Oracle pattern gives us another source of evidence. We run the original program, record what it actually does, and keep those observations.

Then we give the same input to the replacement and compare the results. We will clone the small working example, inspect one recording, ask Codex to replay it, and deliberately break the replacement. Then we will watch recording behavior help us to find and fix the mistake.


## 01:01 — Clone the example

You can follow along from an empty working directory. The commands are in the handout. So you don't need to copy from the screen. You will need Git, Go, Python and make. The agent steps. Install Codex CLI or Copilot CLI and sign in first. The handout links to the setup instructions and lists the required versions.

Here I am cloning the public oracle pattern repository and moving into it. This repository already contains a complete translation and recording from the original program. We are learning how to examine and verify that migration. We are not pretending that we just created the translation during this recording.

You also do not need Fortran compiler for this lesson. This is the one benefit of keeping the original observations. Ordinary verification can run without rebuilding the legacy system. The first Go dependency download may still need an internet connection.


## 02:13 — Explore with the AI

Before we ask the agent to do anything, we need to decide what are we preserving. For example, our example is a numerical routine from SLICOT. It multiplies matrices. You do not need to understand the matrix algorithm to follow the verification process.

First, let us ask Codex what problem does this repository solve and which files should I read first. We are using AI to build a map of the unfamiliar code base. Ask it to inspect the files and explain them in plain English.

Follow its answer into contract. The contract describes the output. The input that must remain unchanged. The error the caller can observe. And the supported storage layout. The last point matters.

The old Fortran routine and the new Go function store matrices differently. We are preserving the meaning of the call while allowing the implementation to change. For your own application, this contract might include a written record, an account balance, an error code, or an updated controller state.

Start at the boundary that you can run and observe. Here is a useful question to stop on. If the replacement changes something, would the caller notice? This is a behavior your verification needs to cover.


## 03:56 — Inspect a legacy recording

Our next question is, where do we expected answers came from? Ask Codex to show actual recording rather than summarize the idea from memory. Now follow the answer into baseline directory. The manifest identifies the original capture and the comparison policy. The compressed data file containing the individual recordings.

Each one has a call identifier, inputs and expected observation from a legacy program. We will use the call named asymmetric-LT. You can see its expected output values here. Those numbers were recorded before we run the current Go implementation.

Keep that separation in mind. The candidate does not get to decide what the expected answer should be. Their synthetic harness calls for this one routine. Not recording from the production application. That gives us useful coverage.

But it does not tell us every possible input or every connected systems will work. Also notice that checksum and source identity. They help us to identify exactly which baseline we are testing against. If the data is missing or corrupt, verification should fail rather than quietly create new set of answers.


## 05:28 — Replay one call

We have used the agent to explore. Now ask it to trace one call through the adapter, the replacement function and the comparison. Once that path makes sense, ask can we test just this one call. Replay this call. Inspect the report.

Tell us whether it passed. Preserve the source and baseline. I will leave the prompt on the screen for a moment. You can reuse it from the handout. Codex runs the same Python command that you could run yourself.

The verifier loads the same stored inputs, builds the Go candidate, execute it, and compares the observations. Here is the result. One call passed with zero failures. The important evidence is the command result and its report.

The agent helps us to navigate and interpret it. But the agent's confidence is not what makes the replacement correct. There is one numerical detail worth understanding. Floating point calculation can differ slightly because of the evaluation order. This repository uses documented scale-aware comparison policy to not make a failing result pass by simply widening the tolerance.


## 06:59 — Detect a real defect

A successful replay is useful. A successful replay of a known bad program would be a problem. First ask Codex how the repository checks that the verifier can detect a mistake. Let it find an existing negative controls and explain one.

Then in this disposable teaching checkout, ask it to deliberately change the one line. The replacement will ignore the part of the matrix that it should include. The recorded input and expected results stay untouched. Now we run exactly the same call again.

This time it fails. the first recorded output should be 9.5 the broken replacement produces minus 1 stop here and look at what changed we did not change the question or the expected answer we changed the candidate and the verifier noticed the difference this report gives us a call identifier and mismatching observation that makes the problem reproducible.

We can investigate one small case rather than guessing from large application failure. This is also why we keep tests are deliberately broken candidates. They check whether our verification can detect representative mistake. A detector that always says yes is not a useful detector.


## 08:41 — Repair and verify

Here is the repair prompt. Investigate this failing call. Use the contract and record observation. And make the smallest candidate fix. Preserve the baseline and comparison policy. Codex can now work from concrete failure.

After the repair, we can focus replay again and inspect the changes. The call passes again but one passing call is not enough to finish. Next we run make verify that include Go checks, the verification infrastructure checks, the deliberately broken candidate and complete recorded corpuses.

The final summary shows 1405 passing calls and 0 failures. We also check that baseline has not changed. In your own migration, retain persistent regression for any defect you discover. This repository already has a negative control for the mistakes we introduced.

Keep that protection when you move to the next change.


## 09:53 — Try it yourself

Choose a behavior you can observe. Capture it from the real legacy system. Keep the input and expected observation independent of the rewrite. Then use the recording throughout the development. While something fails, replay the identical call, investigate the difference, and repair the candidate.

Add new coverage through a deliberate new legacy capture. For a stateful application, a single call is only the part of the story. You also need an initial state and the order of events, followed by the replay of the connected system.

This handout includes these question sequences plus prompt for discovery, one call replay, full verification and mismatch diagnosis. You can adapt them to another coding agent CLI such as Copilot CLI or Claude Code. The commands and launch options will differ.

The need for independent evidence stay the same. Before moving on, try another record call yourself. Then explain where its expected answer came from. If you can do that, you have the central idea of the oracle pattern.

Continue with the [interactive questions](explore-with-ai.md), [reusable prompts](prompts.md), and [from-scratch guide](README.md).

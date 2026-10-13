# Explore the cloned repository with Codex CLI

The lesson is an interactive investigation. Let the AI help you find the relevant code, ask it to show its evidence, and follow up when an explanation is unclear. Run the questions one at a time so you can read each answer before continuing.

## Open the repository in the terminal

```sh
mkdir -p ~/workspace
cd ~/workspace
git clone https://github.com/PlantControl/oracle-pattern.git
cd oracle-pattern
codex --no-alt-screen -s workspace-write -a on-request
```

`--no-alt-screen` keeps the CLI in terminal scrollback. The other flags make the intended workspace permissions explicit. Complete normal Codex sign-in before recording or following the lesson. If an optional update notice appears, you can skip it for the current session.

## 1. Build a map

```text
I just cloned this repository and I am new to the oracle pattern.
What problem does this project solve, and which three files should I read first?
Inspect the repo, use plain English, and keep your answer under 160 words.
Do not change files.
```

Pause and locate each named file. Ask “Why that file?” if its role is unclear. Do not expect a particular model to produce exactly the same wording as the recording.

## 2. Find the source of truth

```text
Where do the expected answers come from? Show the actual baseline manifest
and one recorded call, asymmetric-LT. Explain its inputs, expected output,
and caller identity in plain English. Show only the relevant fields,
not the entire data file. Did the Go replacement generate these answers?
Do not change files. Keep the explanation under 180 words.
```

Look for a specific file and recorded observation supporting the explanation. Distinguish a synthetic harness call from a real application trace.

## 3. Trace the execution path

```text
Walk me through one verification call: recorded input, Go adapter,
replacement function, comparison, and report. Name the actual functions
and files. What is allowed to differ, and what must remain unchanged?
Keep it under 180 words and do not edit files.
```

Follow one path before trying to understand every file. If the explanation is too technical, ask “Explain that without assuming I know Fortran or BLAS.”

## 4. Run the smallest useful experiment

```text
Can we test just asymmetric-LT? Show the command, run it, and inspect its
report. Tell me the call ID, pass/fail, and which baseline was used.
Explain what this proves and what it does not prove. Preserve the source,
baseline, and comparison policy. Keep the final explanation under 140 words.
```

Wait for the actual result. The agent's explanation is not a substitute for the execution and report.

## 5. Challenge the successful result

```text
How do we know this verifier can detect a wrong replacement?
Find the existing executable negative controls and explain one of them.
Do not change anything yet. Keep the answer under 140 words.
```

Predict what should happen before trying the controlled exercise below.

## 6. Controlled failure in a disposable teaching clone

Use this only in the clean clone made for the lesson. The prompt deliberately alters a candidate source file; the repair prompt follows it.

```text
This is a disposable teaching clone. Demonstrate the existing omit-subdiagonal
negative control as a real candidate defect: in hessenberg/product.go change
v := h.Data[(j+1)*h.Stride+j] to v := 0.0. First confirm that exact source site
exists once. Preserve every baseline, comparison policy, and test file.
Run asymmetric-LT and show the failing call and relevant mismatching result.
Stop after demonstrating the failure; do not repair it yet.
```

Stop on the failure and explain why it is expected. Do not substitute changed expected answers.

## 7. Diagnose and repair

```text
Why did that call fail? Use the recorded observation and contract to explain
the cause, then make the smallest repair in hessenberg/product.go.
Preserve the baseline, comparison policy, and test files. Replay asymmetric-LT
and run the existing executable negative-control regression. Show the actual
results. Do not commit, push, create a backlog, or use subagents.
Keep the final explanation under 180 words.
```

The repository already has a persistent negative control for this deliberate defect. For a newly discovered defect in another migration, retain a new runnable regression.

## 8. Verify and state the limits

```text
Run make verify and check that the baseline has not changed. Summarize the
actual results and the limits of the coverage. What would we need to add
before claiming that a whole stateful application had been migrated correctly?
Do not change source, expected results, or tolerances.
```

Your stopping condition is observed verification plus an honest coverage statement. Use the broader [reusable migration prompts](prompts.md) when moving to your own codebase.

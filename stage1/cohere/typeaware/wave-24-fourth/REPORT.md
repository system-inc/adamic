Built an isolated JSX blocker reproducer; the fourth batch is claimed but not ported.
Commits: prior nine ports end at 09bf947bb040786384d3ef055feb376ea4c0832c; fourth claim 0963679435d05258549f2cdd66a2ba2aa904b11a; evidence commit follows.
Commands: setup passed in 19 seconds, nproc 5; validate.py passes ordinary TS, Go JSX and native refusal controls.
Mutant: runnable probe substitutes ordinary TS for JSX and is caught by the independent Go finding and required refusal check; no fourth-batch rule mutants were run.
Not covered: all three native rule ports, corpus parity, rule timings, new released-handle checks and sanitizers; shared JSX parser support blocks static-components.

The three claims remain `react-hooks/set-state-in-effect`, `react-hooks/set-state-in-render`, and `react-hooks/static-components`. The preceding nine ports were already tested and pushed. The fourth claim was pushed before any implementation. Selection fetched 389 origin refs, inspected 33 distinct Markdown claim blobs, and excluded every remote claim plus ports on main and the bridge branch. Positions 168, 169 and 170 were the first eligible checker-dependent rules (all zero measured findings). The selection audit is in evidence/selection.json.

Observation: the unchanged Go cohere production `StaticComponents` rule parses the valid JSX fixture and emits one `staticComponents` diagnostic at byte range 135..144, with zero fixes and suggestions. The owned Go executable imports all three claimed production rules and uses an independent program loader and walk, copied from the preceding batch's validated oracle. The fixture is the production rule test's direct-call case with its original directive. The native parser accepts the ordinary TypeScript control (15 nodes), but exits 70 on the JSX fixture:

```
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 145 in input.tsx
```

Inference: byte-for-byte native static-components parity cannot be established through the current shared parser on this positive case. This is a parser blocker, not the shared lint harness's .a loader or suggestion serializer. Per the user's instruction to stop on any other blocker rather than editing shared files, no shared parser, harness, generator or registration was changed. The two set-state rules have not been implemented; no claim is made that this probe establishes their complete blocker surface. Their Go implementations use the shared high-level intermediate representation, which has not been ported in this batch.

Reproduce from the repository after sourcing the toolchain environment:

```
python3 stage1/cohere/typeaware/wave-24-fourth/validate.py /tmp/wave-24-fourth-validation --stage0 /path/to/adamic > /tmp/wave-24-fourth-validation.log 2>&1
```

The runner builds the native .a probe and Go production oracle into the specified scratch directory. All command output goes to separate stdout/stderr log files. It checks a runnable probe mutant that substitutes the ordinary TS input while retaining the JSX fixture argument; this mutant compiles and exits normally, but fails the required JSX refusal observation. This is a measurement mutant, not a per-rule correctness mutant. Final run output and command logs are archived under evidence/.

There is no native versus Go rule-time comparison because the native rule cannot run. The Go positive control timing is logged for reproducibility, not treated as parity evidence. No new bridge questions or handles were introduced; earlier batches' sanitizer and released-handle results remain in their own reports. No additional rules were claimed, no PR was opened, and this batch stops pending shared parser JSX support.

Toolchain setup: Go 1.27.1 ready 0s; clang 20.1.8 ready 1s; Node 24.19.0 ready 1s; submodules 1s; cache 19s; done 19s; nproc 5 (quota 4).

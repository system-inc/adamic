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

## Resume audit

A subsequent fetch found 417 origin refs and four distinct versions of `stage1/typescript/parser/parser.ts`. One remote version now contains JSX support: `origin/codex/stage1-jsx-lint`, tip `a8a62d62ca49db7415e14c3887dd305022b17309`. The implementation commit is `e715ef4a2f898230af63c40195dea6586a557899`; it modifies shared parser, lookahead and scanner files and adds `jsx.ts`. Main remains `e011f8f60899586d6373a5ccb07335ad82cfbf3c`, and this worker's branch still uses the parser that refuses the valid JSX positive control with exit 70.

The earlier inference is now narrower: JSX support exists on another worker's branch, but is not integrated into this branch. The user requires changes to remain in the owned rule directories and says to stop on other blockers rather than editing shared files. Therefore this worker did not cherry-pick the shared parser/scanner implementation or claim more rules. Integration of that dependency would remove the measured parser blocker; it does not by itself establish HIR or rule parity. No additional port, per-rule mutant, corpus, sanitizer or performance result is claimed by this resume audit.

## Landing gate on current main

Rebased all ten branch-only commits cleanly onto `origin/main` at `e011f8f60899586d6373a5ccb07335ad82cfbf3c`. The resulting tested source tip was `d5228c97906eb85949b9c142648da432dddfae4c`. No new rules were claimed. The only subsequent change records these results and logs; the source implementations are identical to the tested tip.

Setup failed during cache warming, after Go/clang/Node and submodules succeeded (timings 0s/1s/1s/1s, nproc 5). The exact shared lint test errors were `lint_test.go:312:31: undefined: volumeGenerated` and `lint_test.go:316:4: undefined: checkRecoveryRefusal`. No shared files were edited. Building `./cmd/adamic` succeeded, so the unaffected owned oracles could run to completion.

After sourcing `/workspace/adamic-tools/env.sh`, ran:

- `go build -o /tmp/wave-24-landing-adamic ./cmd/adamic`: pass.
- `go test ./stage1/cohere/typeaware -run '^TestWave24AgreementAndMutants$' -count=1 -timeout 30m -v`, with the same frozen 287 repository and 77 compiler roots and `ADAMIC_WAVE24_*` manifest/artifact environment variables: pass in 170.048s.
- `wave-24-next/validate.py` and `wave-24-third/validate.py`, each with the rebuilt `--stage0` and `--compiler /workspace/wave-24-corpus`: both PASS, including both frozen corpora and ASan/UBSan controls/corpora.
- `wave-24-fourth/validate.py` with the rebuilt compiler: PASS for the Go finding, native JSX refusal and measurement mutant; this does not port the three claimed React rules.
- `go test ./bridge/tsgo/... -count=1 -timeout 10m`: pass, bridge 110.253s and checker 0.423s.
- `go vet ./bridge/tsgo/...` and `git diff --check`: pass.

All nine rule mutants, eight checker-question mutants, the released-registry mutant and the blocker measurement mutant were caught again. Ordinary controls matched full Go diagnostics/fixes/suggestions: original batch 43 findings, next 65, third 295. Frozen corpora matched full bytes under ordinary and sanitizer binaries: 18,485 repository bytes and 5,010 compiler bytes per batch, zero findings. All three batches' released-handle checks produced the exact required panic 70. Native versus Go whole-process timings for the original batch were 248.357ms versus 90.650ms on the repository and 1.899463s versus 346.189ms on the compiler; these single observations were made while other validation work ran concurrently.

Logs are in `evidence/landing/`. The complete repository gate remains unrun and the shared setup harness remains broken. The nine completed ports are validated on current main; the three fourth-batch claims still require the separately implemented JSX parser dependency to be integrated. No additional claims or PR were opened.

## Second landing refresh

Fetched current main at `e8ba3d5d81de4d3773c723914fccd4c76248b965` and rebased cleanly. Tested source tip: `0df19e762f268706b0158f9b6c544067b3a7020e`. Rebuilt the compiler and repeated the preceding landing commands, using `/tmp/wave-24-landing2-*` artifact paths. Original batch passed in 148.468s; next and third validators both printed PASS. All three batches matched full bytes over the same 287 repository and 77 compiler roots in normal and ASan/UBSan builds, and all released-handle checks passed. All nine rule mutants, eight checker-question mutants, released-registry mutant and blocker measurement mutant were caught. Bridge package tests and vet passed. Logs are in `evidence/landing2/`.

Original batch whole-process timings were native 266.347ms versus Go 103.976ms for the repository, and native 1.489862s versus Go 309.926ms for the compiler, single observations under concurrent validation load. No full repository gate or nondefault option coverage is claimed. Current main still lacks JSX support; the fourth-batch Go finding/native refusal was rechecked and remains unchanged. No additional rules were claimed and no shared files changed. The rebased result is pushed only to `codex/typeaware-wave-24`, not main or an area branch.

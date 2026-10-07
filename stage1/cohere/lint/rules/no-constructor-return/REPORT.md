Built: no-constructor-return in .a with directory registration and exact Go findings; no repairs.
Commits: claim ab49eb4d was pushed before code; implementation 0f74ed78; evidence commit in git history.
Commands and outputs: Go differential corpus on Node, emitted JavaScript and sanitized native; see the unit evidence.
Mutant: mutant.json changes the rule judgment while preserving compilation and clean execution; all three comparisons must reject it.
Not covered: unmodified shared harness, pinned CLI self-lint of .a, full repository gate or arbitrary malformed inputs.

This directory owns the rule, message, descriptor, independent Go adapter, raw witness and semantic mutant. None of these three rules offers a fix or suggestion. No shared source was edited.

This continuation ports no-constructor-return, no-delete-var and no-eq-null. All prior implementations and evidence were pushed through 115987be first. Fetching 335 origin refs and inspecting 52 distinct recursive claims Markdown blobs showed no remaining helper-ready names. These three are the first remaining entries in the inventory syntax-ready list, absent from origin/main ef3d907ecdc4c771b016f7d9c52372def057a340 and all origin claims. Claim ab49eb4d was committed and pushed before implementation.

Constructor returns walk ancestors until the first constructor or other function-like boundary. Static methods and generator constructors are declined; quoted constructor keys are recognized only under class declarations/expressions, because the stage1 parser represents them as methods while Go represents them as constructors. Bare returns remain silent. The delete rule unwraps grouping and requires a bare Identifier, reporting the entire DeleteExpression. The null rule accepts == and != with a null literal on either side, unwraps grouping, and reports the whole comparison once; undefined and variables bound to null remain silent.

The ordinary shared frontend, generated directory dispatch and finding renderer are used. Scratch Go overlays provide the published .a registry support, a compatible profiling snapshot and bounded comparisons. No dispatch, oracle selection, compiler, parser or shared test source was edited. The previous complete-repair candidates retain their documented integration blockers and were not claimed again.

Setup: `GOFLAGS=-overlay=/tmp/lint-wave1-13-returns/overlay.json bash cloud/setup.sh` succeeds. Timing: Go 0s, clang 1s, Node 1s, submodules 1s, build cache 19s, total 19s. nproc=5; quota 400000/100000, 17.6 GB; Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands, with direct log output:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -overlay=/tmp/lint-wave1-13-returns/overlay.json ./stage1/cohere/lint -run '^TestWave13Returns' -count=1 -v -timeout 20m > /tmp/lint-wave1-13-returns/parity-1.log 2>&1
go test -overlay=/tmp/lint-wave1-13-returns/overlay.json ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-13-returns/registry-final.log 2>&1
go vet -overlay=/tmp/lint-wave1-13-returns/overlay.json ./... > /tmp/lint-wave1-13-returns/vet-final.log 2>&1
```

Replay with `python3 evidence/reproduce.py --compiler /path/to/typescript-6.0.3`. The compiler pin is 050880ce59e30b356b686bd3144efe24f875ebc8. Captured vectors execute cohere's real Go test assertions; the independent Go oracle calls unmodified rule bodies. Every successful comparison must exit zero with empty stderr, and native parity/mutants use ASan and UBSan with leak checks.

The unmodified package command `go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1` exits 1 before tests at profile_test.go:32:23: portFiles is a function, not a rangeable value. default-harness.log retains the error. The shared .a/profile/suggestion work is published on codex/lint-harness-dot-a and remains outside this unit's source territory. No rule-specific repair gap affects these three.

Pinned CLI self-lint: `/tmp/lint-wave1-13/cohere --no-fix --lint --no-cache` followed by the six new .a modules exits 1: nothing to check because .a is not a TypeScript or JavaScript input. self-lint.log retains the refusal. No clean pinned self-lint or full repository gate is claimed. Registry passes in 0.046s and vet exits zero with an empty log.

Mutation witnesses: accepting any enclosing function as a constructor incorrectly reports the nested arrow return; omitting parentheses unwrapping drops delete (value); requiring null on both sides drops a one-sided null comparison. Each mutant must compile and run cleanly on all three backends before a differing Go output counts as a kill.

Corpus suite: PASS in 202.26s, 790 source/rule inputs, zero parser gaps for these rules. Each rule covers 77 pinned compiler files and 160 stage1 .ts/.a files, plus 53/9/14 unique upstream cases and one owned witness. Serialized outputs are 12,405,064 / 12,390,714 / 12,391,662 bytes, totaling 37,187,440 identical bytes. The capture tool logs an inherited return-void JSX gap outside this unit; it is not counted as covered.

Throughput is best of three interleaved whole-process runs including startup, I/O and parsing. Timings use a separate release native binary; parity and mutants use sanitizers. Go, native and Node counts agree. The largely clean corpus and small finding counts make these processing measurements, not an isolated rule benchmark.

| Rule | Inputs | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: | ---: |
| no-constructor-return | 291 | 26 | 20.57 | 30.06 | 127.31 |
| no-delete-var | 247 | 4 | 3.07 | 4.65 | 18.38 |
| no-eq-null | 252 | 8 | 6.13 | 8.32 | 38.26 |

Final complete command: PASS in 264.886s. Mutants PASS in 45.64s: all three compile and exit cleanly, then differ from Go on Node, sanitized native and emitted JavaScript. Selected/all-rule owned witnesses PASS in 16.97s: six rows and 4,089 identical bytes, with positive Go findings required before comparison. Final production code was unchanged between these comparisons and its commits. The ordinary shared-harness and self-lint failure logs are explicitly retained as failures, not counted as passes.

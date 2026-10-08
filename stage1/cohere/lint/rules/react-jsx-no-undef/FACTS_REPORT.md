# Wave 21 shared checker facts landing

Branch: `lint-rules/facts-wave21`. Base: `origin/lint-checker/facts` at `d845dccde413c89643293e808626344d12e3f023`. Merged `origin/area/stage1-lint` at `ad7bd06632f119abc7680719ad3a7d3b71100f58` without resolving or editing shared files. Only owned rule directories are committed by this unit.

| Rule | Unique upstream cases | Program reads | Mutant |
| --- | ---: | --- | --- |
| no-object-constructor | 59 | none | literal parentheses omitted |
| no-promise-executor-return | 135 | none | concise executor results accepted |
| nexus/correctness-no-discarded-pure-result | 22 | ReadsCompilerOptions, ReadsDefaultLibrary | trim omitted |
| react/jsx-no-undef | 45 | none | undeclared names accepted |

All 261 upstream cases, owned witnesses, findings, automatic fixes and suggestions match unchanged Go cohere on source Node, emitted JavaScript and sanitized native. Each mutant compiles, runs and is caught by a finding/suggestion difference on all three runtimes; each rule's `validation/mutant.log` records the three lines. Declaration queries are node-local; no additional program read is declared to accommodate the pilots' over-declaration. No new checker question, bridge implementation or private shared helper is introduced.

## Commands and results

`bash cloud/setup.sh`: PASS, 41.742 seconds, nproc 5, cgroup quota 4 CPUs. Timing lines are in `validation/setup.log`. Builds source `/workspace/adamic-tools/env.sh`.

`go run ./cmd/lint-registry`, `gofmt -l` on every owned Go adapter and `go vet ./stage1/cohere/lint ./cmd/lint-registry`: PASS. The format and vet logs are empty.

`go test -json -count=1 -timeout 3h -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/(jsx-no-undef|promise_return|object_constructor|pure_result)' ./stage1/cohere/lint`: PASS after the area merge, 517.084 seconds (`validation/merged-focus.jsonl.gz`).

`ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave21-typescript-603 ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=/workspace/wave21-facts-merged-profile ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/wave21-facts-merged-profile go test -json -count=1 -timeout 3h ./stage1/cohere/lint`: PASS, 1434.288 seconds. 160 test results passed, 0 failed, 1 skipped; 46 top-level passes and 102 registered mutant passes. The skip is `TestCheckerBridgeRefusalPending`, whose unchanged shared guard awaits the prelude TSGoError marker. All optional inputs were set; no optional-input test skipped. Complete output: `validation/merged-full.jsonl.gz`; counts, loads and environment: `validation/final-metrics.json`.

The compiler/stage1 comparison matched 31,235,137 bytes across 821 files. The discovered JSX inventory held 390 files and matched 304,316 whole-tree bytes against typescript-go. Owned witnesses matched 493,907 bytes. Full upstream capture held 5,638 unique source/rule/options combinations; `TestRulesAgree` matched 14,229,509 syntax bytes plus its individual typed project comparisons.

`ADAMIC_TSGO_CORPUS=/workspace/wave21-typescript-603 go test -json -count=1 -timeout 30m -run 'TestBridge|TestTSGoRequiresLink' ./bridge/tsgo`: PASS, 202.111 seconds. C ABI outputs survive release, stale and zero handles are rejected, new handles differ; 1,600 positions across four compiler files match under ASan/UBSan/LSan. Actual stale-handle, length, wrong-position, linkage, output-free and region-ownership mutants are caught (`validation/bridge.log`).

Release compiler corpus timing, best of five: Go 1.728187s, native 10.714518s, Node 5.957361s, 77 files and 28,338 findings. This is the harness's all-rule corpus benchmark, not a standalone typed-rule timing. JSX timing: Go 0.055264s and native 0.096401s, 390 files and 371 findings. Full timing lines and observed loads are in `validation/merged-full-summary.log`.

## Observations and limits

The facts-only first run exposed a closed JSX inventory. It was stopped after recording two inventory failures, then the approved area branch was merged and the complete package rerun; the final run is separate. The earlier object-constructor captured-TS JSX refusal was removed by area's parser recovery and captured recovery metadata; it was not worked around by narrowing the rule's upstream prefix. Historical reproducer and partial port remain under `validation/parked-object-constructor`, outside registry discovery. The first pure-result mutant survived because the original witness used the default-library global name `name`; its binding was changed to `textValue` and the mutation was then caught on all three runtimes.

Thirteen claimed rules still lack required contracts or shared analysis. `BLOCKED_RULES.md` names every exact Go call, file and line, and the missing question/helper. No bridge, harness, parser, generator or shared helper is edited to bypass those gaps. No further rule was claimed. The broad compiler/repository harness corpus has no checker program for typed rules and emits explicit no-program markers; the typed certification here is the 261 upstream project cases and the owned project witnesses. A full typed compiler/repository corpus was not run. No full repository gate outside the requested lint and bridge packages was run.

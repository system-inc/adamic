Re-cut all five slice 4 members onto current main for parked Loom change chg_8s73y1rr3pyj4n3159yq8tm5s2.
Source 93acd1e747e58ce45f2a44382a9fcd6a60d9032e; base main 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4; delivery is the single commit containing this report.
Focused shards, 58 uncached fixture leaves, stock TypeScript project controls, counts, reader guard, Node declaration controls and changed-package vet passed.
All 49 mutant catchers were rerun successfully: Node stdout, semantic/refusal assertions, ASan, LeakSanitizer and the reader ownership audit.
Not covered: whole-package tests, full gate, optional generic-body/admission census, WASI or benchmarks. No PR, Loom submission, force push or push to main.

Squash: git merge --squash origin/compiler/loom-slice-4. Two conflicts:

- internal/oracle/counts.md: took main's table with git checkout --ours, then regenerated once from the combined compiler. No parked numerical rows were carried forward as baselines.
- internal/load/load.go: retained main's nodeLibraryFS wrapper in both the original checked filesystem and every fresh rebuild; retained main's no-argument nodeTypesIndex(), rooted bundled declaration index and pinned TypeRoots. Combined those with slice project options, reference traversal, ambient types union, host-console selection and deduplicated execution roots. Main's LoadDisplayed API remains intact. Conservative assumption: project-owned roots retain their audited ambient scope; standalone Node rebuilding augments checkRoots rather than discarding their prelude. Main's six bundled Node declaration controls pass, as do the slice's standalone Node/prelude and project reference-root controls. The standalone-node-prelude mutant was adapted only to the rooted index spelling, preserving the same dropped-prelude fault and catcher.

Setup: GOPROXY='https://proxy.golang.org|direct', ADAMIC_GOCACHE_OFF=1, timeout 600 bash cloud/setup.sh. Exit 0: Go ready 0.047s; submodules 0.074s; Node 0.091s; markdown 0.146s; clang 0.465s; cache off 0.467s; build 43.821s; cache warm 43.975s; done 44.001s. nproc=5; cpu.max=400000 100000 (4 CPUs). Go 1.27.1, Node 24.19.0, clang 20.1.8. Every test shell sources /workspace/adamic-tools/env.sh and unsets GOCACHEPROG. No setup retry was needed.

Validation, exact selections and evidence:

- timeout 600 python3 review/compiler/loom-slice-4-2/focused.py: named tests in test-selection.json, 12 per shard; each go test uses -parallel 4 -count=1 -timeout 90s -json. All 14 shards passed. No whole package suite ran. focused-results.json records commands, compilation-inclusive wall seconds and exits.
- go test ./internal/load -run '^(TestNodeLibrary|TestEmbeddedNodeTypes)' -count=1 -timeout 90s -parallel 4 -json: pass, 1.532s; node-library.jsonl.
- ADAMIC_GATE_UNCACHED=1 timeout 240 python3 review/compiler/loom-slice-4-2/fixtures.py: all seven selections pass, 57 actual leaves. Uncached go test ./internal/oracle -run '^TestNativeAgreesWithNode_user_iterators$' -count=1 -timeout 90s -json adds its affected existing solo fixture. Source Node, JavaScript, release native, ASan, UBSan and leak checks match. The generator focused controls also run uncached native comparisons.
- go test ./internal/oracle -run TestCountsAreRecorded -parallel 8 -count=1 -timeout 90s -json -args -update-counts: exit 0. Exactly one regeneration. Counts changed only for the slice: 55 new fixture rows, four existing numerical rows, no removals. counts-attribution.json records every row, before/after and cause. New main rows stay unchanged.
- go test ./internal/ir -run '^(TestCallTargetReaders|TestProduct_CallTargetExports)$' -count=1 -timeout 90s -parallel 1 -json: pass, 13.626s; readers.jsonl.
- timeout 150 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/oracle ./stage3/project-loader/probe: exit 0, no output; vet.log.
- PROJECT_LOADER_TSC=/workspace/adamic/stage3/api/node_modules/typescript/lib/tsc.js timeout 180 python3 review/compiler/loom-slice-4-2/verify.py and verify_types.py: both exit 0. Stock TypeScript solution/reference and ambient union checks agree with source Node and both compiler backends, sanitized native included. Separate loader mutants run once in run-mutants.py, not again in these verifiers.
- timeout 600 python3 review/compiler/loom-slice-4-2/run-mutants.py: exit 0. Three generic body source mutants; six generator source mutants; fourteen loader mutants; four iterator source mutants; three extra scout/snapshot/indexed-read checks; two freshness accessors and one reader audit mutant; six scout measurement mutants with four green measurement controls. Ten remaining IR mutant catchers are the focused oracle tests: two step-16 results, six protocol dispatch changes and two generator frame borrow mutations. Total 49. No compile failure is credited. Commands/exits in mutant-results.json; individual catchers/logs in mutant-run-0..5.log and their directories. mutants-summary.json preserves each of the 49 names and its intended catcher.
- gofmt -l on every changed production/test Go file: exit 0, no output; format.log. Original license and raw evidence bytes are retained, including inherited whitespace in those artifacts.

Every observed leaf, including the intended failing mutant leaves, is under 60 seconds; no tests were deleted, weakened or split. Largest observed leaf is TestNativeAgreesWithNode_048/stage3/fixtures/iteration-dispatch/method_getters.a, 38.26s. All per-leaf seconds are in test-seconds.json. Compiler-consuming packages remain declared and placeless product infrastructure is retained. No cohere code was copied or submodule pin changed by the slice.

Existing numerical rows, A/F/R/L/P/G:

- internal/oracle/testdata/regexp.a: 437/437/382/417/58/0 -> 437/437/383/418/58/0; generator iterator-result boxed-field ownership.
- internal/oracle/testdata/user_iterators.a: 663/663/455/904/64/0 -> 1050/1050/677/994/97/0; cached custom protocol dispatch and close ownership.
- internal/oracle/testdata/user_iterators_rest_tdz.a: 5/0/5/3/5/0 -> 11/3/7/3/8/0; cached custom protocol dispatch and close ownership.
- internal/fresh/testdata/regexp_tree.ts: 91/91/81/85/35/0 -> 91/91/85/89/35/0; generator iterator-result boxed-field ownership.

Runtime files against main (all non-test Go files; no runtime C or header files changed):

- internal/flow/build.go
- internal/flow/infer.go
- internal/fresh/fresh.go
- internal/ir/generator.go
- internal/ir/ir.go
- internal/ir/iteration.go
- internal/javascript/javascript.go
- internal/load/load.go
- internal/load/project_console.go
- internal/load/project_loader.go
- internal/load/project_options.go
- internal/load/project_references.go
- internal/load/source_fs.go
- internal/lower/borrow.go
- internal/lower/class.go
- internal/lower/cycles.go
- internal/lower/exceptions.go
- internal/lower/expression.go
- internal/lower/functions.go
- internal/lower/generator_body.go
- internal/lower/generator_cycles.go
- internal/lower/generator_delegate.go
- internal/lower/generator_expressions.go
- internal/lower/generator_machine.go
- internal/lower/generator_origin.go
- internal/lower/generator_protocol.go
- internal/lower/generator_types.go
- internal/lower/generic.go
- internal/lower/generic_body_relations.go
- internal/lower/invariance.go
- internal/lower/iteration.go
- internal/lower/iteration_array_view.go
- internal/lower/iteration_consume.go
- internal/lower/iteration_intrinsic.go
- internal/lower/labels.go
- internal/lower/library_function_expressions.go
- internal/lower/object.go
- internal/lower/refusals.go
- internal/lower/statements.go
- internal/native/emit_expressions.go
- internal/native/iteration.go
- internal/native/reuse.go
- stage3/project-loader/probe/main.go

Main-tip fetch after all tests found no movement. No rebase was needed. The final handoff supplies the pushed delivery SHA. Generated verification binaries and adapted scratch sources are excluded from the delivery; logs, JSON, reports and noncompilable mutant sources are retained.

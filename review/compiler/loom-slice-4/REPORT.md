Re-cut all five lowering-chain slice 4 members onto current main, with conservative protocol accessor effects in freshness analysis.
Source 0051db09d18e8c15821c1b06ef86d4b1bafa6565; base main c0fd8945b872e57a341ccee5973e3f4d1b75c34b; delivery is the single commit containing this report.
Green: focused named tests, 57 selected fixture leaves plus the existing user-iterator solo, stock TypeScript project controls, counts, readers, vet and formatting.
All 46 retained slice mutants, two new accessor-effect mutants, and the reader ownership mutant were caught: 49 total, listed in mutants-summary.json.
Not covered: whole packages, full gate, opt-in generic-body census, WASI, admission census, benchmarks, Loom submission or PR creation.

The source net against its merge base was brought by git merge --squash origin/compiler/chain-slice-4. All five members remain: generic body relations, generics scout, iteration, project references, generators. No cohere code was copied; its existing submodule remains pinned.

Conflicts, each resolved once:

- cloud/fast-gate/compiler-dependencies.json: retain all six main consumer declarations and add stage3/project-loader/probe depending on internal/load. Retain main's compact formatting.
- internal/ir/call_targets_guard_test.go: retain main's placeless buildcache.GoExports product, extend that product and its typed reader audit to oracle tests, retain the explicit owner/reason entries from the slice. Main renamed the library mutant reader into runLibraryMethodValueMutantsShard; its exemption now names that runner.
- internal/load/load.go: retain LoadDisplayed and the three-argument internal loader, preserve display on sourceFS, and add project option auditing, transitive source roots, ambient types union, host-console handling and root deduplication.
- internal/load/source_fs.go: retain main's display callback and add projectConsole and its prelude selection.

The focused freshness controls exposed an additional semantic reconciliation: IteratorField was unknown to the cycle finder. IteratorField and IteratorMethod now use conservative arbitrary-call effects, including receiver escape. Two new top-level tests show that subsequent writes into an escaped receiver cannot be certified by confinement; changing either protocol lookup to a pure field load fails its dedicated assertion. Assumption: protocol accessors may run arbitrary user code, even when the result is scalar. No precision improvement is assumed.

Rules 11 and 12: no test was deleted or weakened, and none needed splitting. All observed final leaves are below 60 seconds on cpu.max=400000 100000 (4 CPUs), nproc=5. New tests TestIteratorFieldAccessorEscapesReceiver and TestIteratorMethodAccessorEscapesReceiver each reported 0.00s. Every leaf's Go-reported seconds, including fixture/count children, is in test-seconds.json. Maximum top-level observation: TestCallTargetReaders 19.06s. Test command wall time includes compilation separately in focused-results.json; initial cold compilation is not a test leaf. Existing LoadDisplayed and GoExports products remain placeless. New source tests contain no literal machine or build-directory paths. The new compiler consumer is declared in main's registry.

Toolchain setup: first timeout 240 bash cloud/setup.sh exited 124 in nested TypeScript fetch, before completion. Second attempt exited 124 during cold Go compilation. The final monitored timeout 600 attempt, ADAMIC_GOCACHE_OFF=1, passed: Node 0.018s, Go 0.019s, submodules 0.058s, markdown 0.065s, clang 0.151s, shared cache off 0.153s, Go build 247.509s, cache warm 247.898s, done 247.932s. Go 1.27.1, Node 24.19.0, clang 20.1.8. Every subsequent shell sourced /workspace/adamic-tools/env.sh and unset GOCACHEPROG. GOPROXY was https://proxy.golang.org|direct. The first focused and vet builds overlapped setup and were stopped before tests started; their logs are named aborted-*. The first combined fixture selector ran no fixtures and is not credited; fixtures.py requires passing fixture leaves for every corrected selector.

Exact validation commands and outputs:

- timeout 240 python3 review/compiler/loom-slice-4/focused.py: named tests from touched test files, 12 names per shard, go test -parallel 4 -count=1 -timeout 90s -json. All flow, load, lower and oracle shards passed. Initial fresh shard failed on unknown IteratorField; final go test ./internal/fresh -run '^TestGenerator' -parallel 4 -count=1 -timeout 90s -json passed all 15 controls (0.544s).
- go test ./internal/fresh -run '^TestIterator(Field|Method)AccessorEscapesReceiver$' -count=1 -timeout 90s -json: pass 0.008s.
- ADAMIC_GATE_UNCACHED=1 timeout 240 python3 review/compiler/loom-slice-4/fixtures.py: seven separate TestNativeAgreesWithNode_[0-9]+ path selections, all passed, 57 actual fixture leaves. go test ./internal/oracle -run '^TestNativeAgreesWithNode_user_iterators$' -count=1 -timeout 90s -json also passed. The ordinary generator and regexp compatibility focused tests separately use uncached native execution.
- go test ./internal/oracle -run TestCountsAreRecorded -parallel 8 -count=1 -timeout 90s -json -args -update-counts: pass. One successful regeneration, TestMain writes once after main's top-level families finish. Exactly 55 new rows and four changed rows, no removals. Each row and its attribution is in counts-attribution.json.
- go test ./internal/ir -run '^(TestProduct_CallTargetExports|TestCallTargetReaders)$' -count=1 -timeout 90s -parallel 1 -json: pass. After the freshness changes, go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -json passed again.
- timeout 120 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/oracle ./stage3/project-loader/probe: exit 0, no output, repeated after freshness changes.
- PROJECT_LOADER_TSC=/workspace/adamic/stage3/api/node_modules/typescript/lib/tsc.js timeout 180 python3 review/compiler/loom-slice-4/verify.py and verify_types.py: both pass, source Node/native sanitized/backend outputs 7 newline; compatible graph traversal, unimported source diagnostic location/code, host-console limitation, incompatible option refusal, and ambient duplicate declaration union match stock TypeScript. Generated scratch binaries and adapted sources are excluded from the delivery; logs and RESULT.json are retained. Identical loader mutant checks are run once by loader-mutants.py.
- timeout 240 python3 review/compiler/chain-slice-4/generic-mutants.py review/compiler/loom-slice-4/generic-mutants: three caught.
- timeout 240 python3 review/compiler/loom-slice-4/generator-mutants.py: six caught.
- timeout 240 python3 review/compiler/loom-slice-4/loader-mutants.py: fourteen caught.
- timeout 240 python3 review/compiler/loom-slice-4/iteration-mutants.py review/compiler/loom-slice-4/iteration-mutants: four caught.
- timeout 180 python3 review/compiler/loom-slice-4/extra-mutants.py: array callback mutant caught by both backends' Node stdout, wrong snapshot caught by exact outcome assertion, old indexed stop caught by NodeArray admission assertion.
- Six scout measurement mutants failed with their intended assertions; four controls passed. Commands and exits are in scout-results.json. Six protocol IR mutants and two generator borrow mutants were caught inside the focused oracle tests; the latter report ASan heap-use-after-free. Two pure-load freshness source mutants and the reader ownership mutant were caught in their dedicated logs. No compilation failure counts as a kill.
- gofmt -l on all changed Go files: no output; git diff --cached --check flags whitespace in historical and new raw logs/patch artifacts and the upstream TEST262 license; those text artifacts are retained byte for byte.

Runtime files changed against main (all non-test Go files; no runtime C or header files changed):

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

Existing counts rows, A/F/R/L/P/G:

- internal/oracle/testdata/regexp.a: 437/437/382/417/58/0 -> 437/437/383/418/58/0; generator result boxed-field ownership.
- internal/oracle/testdata/user_iterators.a: 663/663/455/904/64/0 -> 1050/1050/677/994/97/0; cached protocol dispatch and close ownership.
- internal/oracle/testdata/user_iterators_rest_tdz.a: 5/0/5/3/5/0 -> 11/3/7/3/8/0; cached protocol dispatch and close ownership.
- internal/fresh/testdata/regexp_tree.ts: 91/91/81/85/35/0 -> 91/91/85/89/35/0; generator result boxed-field ownership.

Main-tip fetch immediately before commit found no movement. No rebase, force push, PR, push to main, or Loom submission was performed. The requested branch is compiler/loom-slice-4. The final handoff supplies the delivery SHA and base SHA.

The 30-minute first-push target was missed during environment startup, the large nested submodule fetch, cold compiler setup recovery, and required verification. No unverified state was pushed to meet the clock.

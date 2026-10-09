Built: runtime feature-set link rejection and monolithic feature selection rebuilt on main for step 04, task #hmab710.
Commits: 4f23173f06ee3387ab3af836f2427147f95c7918 rebuilds 7e0a99fd and 5e5cb3f1; the delivery commit adds this evidence.
Commands and outputs: native and WASI matrices, checker controls, tool witnesses, regexp mutants and count regeneration pass; commands are below.
Mutants: four link-guard mutations, split runtime default flags, original unsanitized checker default flags and monolithic default flags are caught; regexp mutations retain their oracle witnesses.
Not covered: macOS execution, whole packages or the full gate. No c2-only dependency was needed, and no registered fixture or count row was added.

Every unit including adamic.h retains a reference to a symbol naming its five runtime features. features.c defines the runtime's symbol. A mismatched runtime fails at link without sanitizers, including after linker section removal. Cached headers retain their original bytes so they cannot silently override a client's feature set. Monolithic Build uses the existing sourceFlags helper before any includes, just as split builds do.

The cherry-pick conflict was confined to internal/native/wasm_test.go. Main already shards and caches that harness. Its runtimeBuild cache now includes featureFlags in the compiled flags and key; link selects the matching feature and count set. The three additional WASI fixtures are appended, preserving existing fixture indices. There are 39 top-level leaves for 38 fixtures and the request probe. Mismatch directions, native/WASI matrices and the five changed regexp mutant leaves are also top-level parallel tests. The stage 1 helper keeps its existing admission helper.

The branch started at main 8c76557e and was rebased to main 85172588, then main 553ad06a. Main's intervening changes only touch typeaware and JSON tests and a documentation deletion; they do not change the compiler, fixture programs or count table. Native and oracle controls were rerun after the rebase. Neither da03166e nor f5f37525 was picked, and no unlanded branch was merged. Historical reports from the picked commits were omitted; current evidence is here.

Commands source /workspace/adamic-tools/env.sh and set GOPROXY=https://proxy.golang.org|direct. Output is redirected directly to the indicated logs; no test output is piped.

| Command | Result | Log |
| --- | --- | --- |
| ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/c2-pair-tsgo.a timeout 150s go test ./internal/native -run '^TestRuntimeFeature\|^TestSplitTSGoRuntimeFeatures$\|^TestClosureConventionRuntimeDropCount$' -count=1 -v -timeout 90s | PASS, 15.466s | native.log.gz |
| timeout 150s go test ./internal/oracle -run '^TestRegExpReplacement(NodeMutants\|TypeGuardMutants)\|^TestNotYetLibraryRegex(Callback\|Offset)Mutants$' -count=1 -v -timeout 90s | PASS, 20.619s | oracle-mutants.log.gz |
| timeout 180s go test ./internal/fuzz ./cmd/adamic-test262 ./stage1/cohere/markdownblocks -run '^TestFuzzerFeatureMismatchStopsAtLink$\|^TestRunnerFeatureMismatchStopsAtLink$\|^TestFuzzerSharesRuntimeLibrary$\|^TestNativeMutantUsesProgramsFeatures$' -count=1 -v -timeout 90s | PASS, all three packages | tools.log.gz |
| ADAMIC_TEST_WASI=1 PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH timeout 210s go test ./internal/native -run '^TestWASIUnit' -count=1 -v -timeout 180s | PASS, 69.546s; 38 Node-held fixtures and requests | wasi.log.gz |
| ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/c2-pair-tsgo.a timeout 210s go test ./internal/native -run '^TestSplitTSGoAgreesUnit\|^TestTSGoBuildSeesTheProgramsFeatures$' -count=1 -v -timeout 180s | PASS, 62.620s; both cache modes | checker-positive.log.gz |
| ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/c2-pair-tsgo.a timeout 600s python3 internal/native/testdata/run-feature-set-link-mutants.py | PASS; every subprocess has -timeout 90s and a 150s process limit | link-mutants.log.gz |
| timeout 210s go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 180s -args -update-counts | PASS, 72.665s; no table changes | counts.log.gz |

The table's vertical bars inside regexps are command alternation; exact commands are repeated in commands.txt for copying. The two focused native/oracle commands were rerun after rebasing; their final results are in native-final.log.gz and oracle-final.log.gz. The additional monolithic mutation uses mutants/monolithic-default-flags/overlay.json with go test ./internal/native -run '^TestClosureConventionRuntimeDropCount$' -count=1 -v -timeout 90s. It exits 1 at the unmutated control's clang build, before the callback-count mutation. Its Go subprocess is bounded at 150 seconds.

Mutant observations:

- fixed-symbol, drop-reference and drop-retain allow the mismatched program to link; TestRuntimeFeatureMismatchProgramFeature rejects that success.
- header-forces-runtime-features allows the reverse mismatch to link; TestRuntimeFeatureMismatchRuntimeFeature rejects that success.
- split-runtime-default-flags stops at adamic_runtime_features_closure_convention_closure_receivers without sanitizers; the negative witness observes the intended link failure and the ordinary matching-build test fails.
- split-original-checker-default-flags changes the same runtime selection and removes sanitizers through an overlay; main's TestSplitTSGoAgreesUnit00 fails at the feature symbol. This is the original checker's first existing shard, not an imported c2 fixture.
- monolithic-default-flags is caught by the valid prepended runtime control, proving the sourceFlags repair.
- regexp offset, group order, named groups, callback match/literal/unicode/global/offset/reset/collection order and input mutations finish cleanly and differ from source Node stdout. Argument and group guard removal differ from the pinned declared-type refusal. Runtime callback count removal is refused by clang for typed arity, expected 3, have 2, after a sanitized control agrees with Node.

The feature contract follows macro definedness, matching the five layout gates in library.go. ADAMIC_COUNT and ADAMIC_TSGO remain outside that contract. Shared-runtime fuzzer and test262 paths now refuse mismatches; this change does not add per-program runtime selection to those tools.

All 59 observed top-level leaves are below 60 seconds, including setup. Maximum: the existing checker Unit00, 39.31s. New WASI Unit36/37/38: 6.74/6.64/6.38s. test-seconds.tsv records every observed top-level leaf. Native matrices each check 32 matching sets, 160 one-bit mismatches and one unused mixed unit.

Setup passed: node 0.026s; go 0.028s; markdown dependency step 0.007s, ready 0.077s; submodules 0.091s; clang 0.206s; go build 38.669s; test binaries deferred 38.803s; cache warm 38.804s; total 38.830s. nproc=5, cpu.max=400000 100000. WASI setup passed: SDK 11.798s; cache warm 99.157s; total 99.185s. Complete timing lines are in setup.log.gz and wasi-setup.log.gz. Go 1.27.1, Node 24.19.0, clang 20.1.8, WASI clang 20.1.8-wasi-sdk. timeout 180s npm ci --prefix stage3/api installed the pinned Node API types; timeout 240s go build -buildmode=c-archive -o /tmp/c2-pair-tsgo.a ./bridge/tsgo/archive built the pinned checker archive.

Lane checks on the committed change pass: lane checks 3.0 s: gofmt and tools on 10 Go files, t.Parallel on 5 test packages; vet 5 packages. git diff --check passes. The required final lane invocation follows the evidence commit and its output is reported with the pushed SHA.
native-final: PASS, 15.597s after rebasing.
oracle-final: PASS, 5.161s after rebasing.

Final main refresh added JSON test shards only. Final lane checks are rerun on the committed delivery after that rebase.

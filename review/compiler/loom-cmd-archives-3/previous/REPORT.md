Restored cmd/adamic runtime products with one builder package per shared name key and explicit owner waits.
Base bd90a433bf636b8325a591ab3e7487d457a69ab4; restores 4011e74a, re-cut after its revert; final submission SHA is in the worker report.
Ordered tracing: 9 observed runtime keys, one builder each; shared cmd products found oracle's exact keys; all behavior leaves passed in read mode.
Five mutants caught: second builder, second owner, missing wait, missing key part, and missing sanitized declaration.
Not covered: full gate, complete packages, macOS, store publication, or a full-tree Loom rebuild; historical and fresh timings are labeled separately.

The earlier change assumed that two packages declaring the same name key would share one product automatically. Loom exposed why that was insufficient: each producer binary can settle that name under a different key. The re-cut makes the distinction explicit. A shared cmd/adamic product calls buildcache.BuiltElsewhere, the primitive runtimeproducts.Find uses, and cannot invoke its build callback under the tree's trace. A missing owner fails immediately with the product-order diagnostic. Outside tracing, BuiltElsewhere retains main's ordinary developer fallback, as runtimeproducts.Find does.

No runtime C, existing behavior test, assertion, fixture, or runtime feature set changed. Every requested archive uses the existing base feature set []. Request affects WASI linking only and is cleared for declaration matching. commandRuntimeInputs now delegates to runtimeproducts.Inputs instead of copying the identity recipe. No API extension or counts.md refresh is needed. No new package was added; runtimeproducts is already declared as a compiler consumer in compiler-dependencies.json.

Ownership audit

The census scans tracked Go files across the tree, excluding submodule contents (git links), for callers of runtimeproducts.Build and Supply, and rejects private copies of the native-runtime identity. Current callers that can produce shared runtime products are all in internal/oracle. stage3/fixtures uses Find and remains a finder. Native's private runtime cache and split tsgo cache are not buildcache products with this name key. runtime-sites.txt preserves the grep audit.

Oracle's CountMeasurements00 through 07 were the reported colliders. Current main also has shards 08 through 15; any of them can reach the same base sanitized archive through countedRun -> identity -> native.RuntimeLibrary -> runtimeproducts.Supply. identity also asks for base release and counted native archives, which cmd/adamic does not request. All those calls remain within the oracle producer package. No oracle or other package's tests were edited.

Conservative assumption: “one builder” means one producer package under tracing, as Loom's ruling states. Multiple oracle product tests use the same oracle.test recipe; they are one producer package, not additional cmd/adamic owners. The static census covers the current shared family (12 Sets x 4 Variants), with cmd/adamic's counted-WASI combination outside that family. Its two shared references plus sole counted-WASI builder leave one producer package for each key. A generic dynamic census of every product in the whole repository belongs in Loom's cross-process build-tree journal check; no such whole-tree ownership test/tool is present in this checkout. The new focused census and actual product guard hold this runtime family here.

| Archive / current machine name key | Old possible builders | New builder package and cmd/adamic action |
|---|---|---|
| Base sanitized, f0fc5b52d9e3 | cmd/adamic TestProduct_CommandRuntimeSanitized; oracle TestProduct_NativeRuntimesSanitized0; oracle count-measurement/identity callers | internal/oracle only; cmd/adamic finds by name and waits for internal/oracle TestProduct_NativeRuntimesSanitized0 |
| Base WASI, d7d9fcd63992 | cmd/adamic TestProduct_CommandRuntimeWASI; oracle TestProduct_NativeRuntimesWASI0 | internal/oracle only; cmd/adamic finds by name and waits for internal/oracle TestProduct_NativeRuntimesWASI0 |
| Base counted WASI, e2db043dcba9 | cmd/adamic TestProduct_CommandRuntimeWASICounted | cmd/adamic TestProduct_CommandRuntimeWASICounted, unchanged sole builder |

Loom's reported a33c0b598f8a is the base sanitized request on its toolchain/header/platform inputs. This environment computes f0fc5b52d9e3 for that request. Tests and mutants compute the real name key through RuntimeRequestFor and runtimeproducts.Inputs; they do not hardcode a machine's digest. Both second-builder mutants target exactly that base sanitized request.

The two owner tests build four archives each on current main. The six incidental keys observed in that proof retain oracle as their only builder: sanitized 71283f8718ad (receivers), 7afbe2549a72 (convention + canonical), ac06e28f8521 (canonical + receivers); WASI 5a26299e07a6, 9bd6e516dc86, fb3b966fafba for the same respective feature sets. No declarations for those six keys changed. This is all nine runtime keys touched by the trace proof.

Product order and real trace proof

cloud/fast-gate/product-order.json adds exactly two waits, from the shared cmd product tests to the named oracle owner tests. Counted WASI has no external owner and needs no wait. Every owner and cmd product was run separately, bounded by timeout 60 and -test.timeout=55s, with the normal package working directory:

* internal/oracle TestProduct_NativeRuntimesSanitized0: 31.48 s, four settled, zero refused.
* internal/oracle TestProduct_NativeRuntimesWASI0: 28.19 s, four settled, zero refused.
* cmd/adamic TestProduct_CommandRuntimeSanitized: 21.78 s, zero settled, zero refused, native-runtime outcome found.
* cmd/adamic TestProduct_CommandRuntimeWASI: 20.77 s, zero settled, zero refused, native-runtime outcome found.
* cmd/adamic TestProduct_CommandRuntimeWASICounted: final run 21.90 s, one settled, zero refused. An earlier run before adding the behavioral guard was 6.08 s; the final changed recipe was rebuilt and settled separately.

check-traces.py checks the actual JSONL reports: each shared finder names the exact owner settlement key, neither settles anything, and counted WASI has no oracle settlement. trace-ownership.log records all full name keys and settlement keys. Nine observed name keys have one producer each. The changed counted-WASI recipe was rebuilt after the new test was added; it is not counted as a second producer on one tree.

Guards and mutants

* TestCommandRuntimeArchivesHaveOneBuilder censuses the tree's shared recipe callers and counts builder packages for each actual cmd request. Clearing sanitized's owner adds cmd/adamic as a second builder and fails with builder packages [internal/oracle cmd/adamic].
* TestCommandRuntimeProductsWaitForOwners reads the actual product Test function calls and product-order.json. Removing the sanitized wait fails with the missing owner edge.
* TestSharedCommandRuntimeCannotBuildItsOwnersArchive runs the real sanitized product test in a fresh child process with the tree builder's trace protocol and an empty owner cache. The expected failure names the owner and asks whether the test is after it in the product order. Its mutant replaces BuiltElsewhere with Product: the child actually builds the archive successfully, so the guard fails. This is the direct second builder of the request Loom reported as a33c0b598f8a, not a compiler-warning or source-fixture failure. The unit additionally ran the real strace pipeline above; the test itself needs no strace process.
* The restored product key guard fails when commandRuntimeInputs drops Portable[0], naming adamic-runtime-v6.
* The unchanged TestExplainChecksOutput/overload_some_empty/build fails when the sanitized declaration is removed.

mutants.py runs each mutation independently with go test ./cmd/adamic -run SELECTED -count=1 -timeout 60s, bounded by timeout 90, and restores the original source in finally. All five exit 1 with the intended assertion; patches use .patch, never compilable Go. Second-builder's child completed its archive build in 4.39 s; only the parent owner guard rejected it. No mutant was caught by -Werror.

Timings

previous/ contains the full earlier unit's evidence from 4011e74a, unchanged, including its historical report and raw logs. Its report's producer assumption is superseded above. The first baseline row uses its isolated before-recheck observation, as the earlier report explains. These October 14 cold and warm observations are retained, not claimed to be fresh October 16 baselines. Fresh re-cut behavior observations below use the final test binary and ADAMIC_BUILD_CACHE=read, with products already supplied. Each leaf has timeout 60 and -test.timeout=55s; every one passed below 60 s. Go binary compilation and machine toolchain installation are outside test times; per-test runtime request/header setup is included.

| Runtime-linking leaf | Oct 14 cold s | Oct 14 warm s | Re-cut ahead s |
|---|---:|---:|---:|
| TestExplainChecksOutput/overload_some_empty/build | 9.987 | 1.229 | 2.682 |
| TestExplainChecksOutput/overload_some_false_read/build | 9.414 | 1.306 | 2.957 |
| TestExplainChecksOutput/overload_some_true_only/build | 11.396 | 1.540 | 2.178 |
| TestExplainChecksOutput/overload_erased/build | 12.094 | 1.166 | 2.105 |
| TestExplainChecksOutput/overload_assertion/build | 9.092 | 1.198 | 2.130 |
| TestNonNullExplainChecks/narrowed/build | 8.963 | 1.223 | 2.628 |
| TestNonNullExplainChecks/present/build | 9.168 | 1.198 | 2.913 |
| TestNonNullExplainChecks/undefined/build | 8.887 | 1.433 | 2.653 |
| TestNonNullExplainChecks/null/build | 9.113 | 1.241 | 2.031 |
| TestNonNullExplainChecks/initializer/build | 9.978 | 1.216 | 2.076 |
| TestNonNullExplainChecks/let_initializer/build | 8.828 | 1.208 | 2.417 |
| TestNonNullExplainChecks/refuse_undefined/build | 9.304 | 1.220 | 2.308 |
| TestNonNullExplainChecks/refuse_null/build | 8.681 | 1.221 | 2.747 |
| TestRequestNativeStaysCommand | 9.417 | 1.213 | 3.802 |
| TestWASIRequest/request.a | 11.464 | 1.909 | 2.509 |
| TestWASIRequest/request_alias.a | 11.371 | 1.429 | 2.342 |
| TestWASIRequestThrows | 9.367 | 1.631 | 2.195 |

Fresh top-level observations, including original C/JS assertions and the two request leaves sharing their archive:

| Unit | Seconds | Core-seconds |
|---|---:|---:|
| TestExplainChecksOutput | 6.744 | 23.605 |
| TestNonNullExplainChecks | 8.365 | 29.983 |
| TestRequestNativeStaysCommand | 2.456 | 3.084 |
| TestWASIRequest | 3.052 | 3.768 |
| TestWASIRequestThrows | 1.986 | 2.540 |

Total fresh behavior CPU: 62.980 core-seconds. The earlier same-run comparison remains 185.523 -> 43.673, saving 141.849 core-seconds. This re-cut does not claim a causal speedup by comparing across dates, source trees, cache recipes, and cloud load.

New/restored guard timings on the final sources:

* TestCommandRuntimeArchivesHaveOneBuilder: 4.07s.
* TestCommandRuntimeProductsWaitForOwners: 0.00s.
* TestCommandRuntimeProductKeyHoldsEveryPart: 1.48s.
* TestSharedCommandRuntimeCannotBuildItsOwnersArchive: 1.62 s.

Exact commands and coverage

All test output went straight to files, never through a pipe. Scripts here preserve selections and environment. Toolchain shells sourced /workspace/adamic-tools/env.sh; product runs used ADAMIC_BUILD_STORE=off and ADAMIC_ORACLE_WASI=1. No PR or Loom submission.

* git fetch origin and checkout -b compiler/loom-cmd-archives-2 origin/main, base bd90a433. git show 4011e74a is preserved in 4011e74a.patch.txt; restored source is re-cut rather than cherry-picked as a second commit.
* timeout 240 bash cloud/setup.sh --wasi-sdk: passed. nproc 5, cgroup cpu.max 400000 100000. Timing lines: Node 0.127 s, Go 0.165 s, markdown 0.414 s, clang 0.848 s, submodules 2.040 s, WASI SDK 2.339 s, shared cache 4.269 s, go build 108.751 s, warm cache 108.952 s, done 109.064 s. Test binaries deferred as setup specified. Node 24.19.0, Go 1.27.1, clang 20.1.8. Full output in setup.log and toolchain.log. No setup workaround was needed this time.
* go test -c ./cmd/adamic and ./internal/oracle; go build ./internal/buildcache/cmd/traced, each bounded by timeout 90; passed. Final cmd test binary rebuilt after adding the actual-product owner guard.
* Owner and finder product selections listed above, each under the current cmd/traced executable with ADAMIC_TRACE_REPORT pointing outside the repository, then copied into review after settlement. Final counted-WASI product was settled again after the final guard source changed. All final runs settled or found as expected; no refused products.
* measure.py /tmp/adamic-recut-final.test after: timeout 120 overall, timeout 60 / -test.timeout=55s per leaf. All 17 pass in read mode. units.py on the final binary: timeout 90 overall, same per-unit limits, all five pass. remaining.py selects every other original cmd/adamic Test function individually, same leaf limits; all eight pass. There was no whole-package test invocation.
* Three final ownership/order/key guards each separately selected with -test.run=^NAME$, timeout 60 / -test.timeout=55s. Behavioral owner guard separately passed. Every new/restored test is below 60 s including its request setup; timing logs above.
* timeout 360 python3 mutants.py, each selected go test command bounded by timeout 90 / -timeout 60s. All five mutants caught, original sources restored.
* timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s: passed, 48.871 s. timeout 90 go vet ./cmd/adamic: passed without diagnostics.
* check-traces.py: passed; actual shared keys match owner settlements. final-checks.log records source formatting, JSON parsing and restored-source checks. Raw historical logs retain their original whitespace; source formatting is checked separately from archived evidence.

The first manual guard invocation used the repository root instead of Go's package working directory, so git ls-files at ../.. failed immediately. The harness was corrected before any product proof, and all reported guard results use cmd/adamic's normal directory. No test exceeded its limit and no over-limit test was rerun.

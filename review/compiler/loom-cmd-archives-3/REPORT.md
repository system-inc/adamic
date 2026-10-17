Restored cmd/adamic runtime archive products with exactly one producer package per shared name key.
Base f93a65fba5a9d7dcd3be462391c8bc8bbef80d0a; restored e54dc0a7e9e18012efebead6c05bf7a44baa3bb9; submission SHA is in the final worker report.
Ordered product traces passed; both cmd finders read oracle's exact keys and counted WASI settled only in cmd/adamic.
All five mutants were caught, including a second builder of the base sanitized request in both census and actual-product guards.
Not covered: full packages/full gate, macOS, remote store publication, or a complete Loom tree build; the existing reader guard took 61.224 s under its required 90 s limit.

The two Go test files are byte-for-byte e54dc0a7. Main's product-order.json is preserved with exactly six inserted lines: CommandRuntimeSanitized waits for internal/oracle TestProduct_NativeRuntimesSanitized0, and CommandRuntimeWASI waits for internal/oracle TestProduct_NativeRuntimesWASI0. No other edge changed. Both owners remain present on main. No runtime C, behavior assertions, fixtures or feature sets changed; counts.md needs no refresh.

Current APIs were read before restoration: runtimeproducts.go in full, Find in full, BuiltElsewhere and its fallback, and lean build helpers. commandRuntimeInputs delegates to runtimeproducts.Inputs, so it follows main's current runtime file list and identity without a private copy. BuiltElsewhere remains the current Find primitive: under tracing it finds by name or fails naming its owner, never building; outside tracing it retains main's developer fallback. The strict behavior-test supplier still returns an error on any miss in either mode. No API extension is necessary. Counted WASI keeps ordinary Product: cmd/adamic is not a reviewed lean caller, and there is no reason to extend that privilege for its sole archive.

Conservative assumption, unchanged from Loom's ruling: one builder means one producer package under tracing. Oracle's NativeRuntimes and CountMeasurements callers share its producer package. The tracked-source census checks the shared recipe's Build/Supply callers and rejects private native-runtime identities; the actual-product guard also catches replacing BuiltElsewhere with a builder without changing the declaration. This is a focused runtime ownership check, not a claim to census every unrelated product namespace. Main's runtimeproducts shared family has 12 sets and four variants; no second producer package was found. runtime-sites.txt preserves the current call-site audit.

Every cmd archive uses feature set [] and unlisted options are zero. Sanitized serves the five ExplainChecks build leaves, eight NonNull build leaves and RequestNativeStaysCommand. Counted WASI serves WASIRequest's request.a and request_alias.a; plain WASI serves WASIRequestThrows. Request is link-only and is cleared for declaration matching; it does not change the archive identity.

| Archive / current name key prefix | Builders in the original pre-revert change | Builder after this restoration |
|---|---|---|
| Base sanitized / 3ac8c29ccf42 | cmd/adamic plus oracle NativeRuntimesSanitized0 and count/identity callers | internal/oracle only; cmd/adamic finds by name after Sanitized0 |
| Base WASI / 6e08d061bac9 | cmd/adamic plus oracle NativeRuntimesWASI0 | internal/oracle only; cmd/adamic finds by name after WASI0 |
| Base counted WASI / 2b3ce1fb7e30 | cmd/adamic | cmd/adamic TestProduct_CommandRuntimeWASICounted only |

Main before restoration has no cmd product builders; its tests compile their private runtime cache on demand. Relative to e54dc0a7, ownership is unchanged. Loom's a33c0b598f8a identifies the same logical base sanitized request on its inputs; this machine computes 3ac8c29ccf42. No digest is hardcoded in the tests or mutant.

If runtime piece 4 lands, its six groups still place the base set at index zero in NativeRuntimesSanitized0 and NativeRuntimesWASI0. These two waits and all three cmd declarations stay unchanged, and cmd continues using the identity main's runtimeproducts.Inputs gives it. Other sets move among oracle's shards, and buildcache revalues the shared recipe as usual. cmd declares no second archive family.

The ordered trace observed nine runtime name keys, with one settlement each. Both cmd shared products report found with the exact owner settlement key; neither settles anything. Counted WASI has no oracle settlement. check-traces.py asserts those facts; trace-ownership.log records every full shared key. The six incidental owner-group archives retain oracle as their sole builder, before and after:

* Sanitized: 19877fc5e50b, 74c7e3895326, 9c3a08bf55d2. These are the receivers, convention + canonical, and canonical + receivers sets already in group zero.
* WASI: 7f0d666b1361, 21bf4c5826f5, a954e124b033. These are the receivers, convention + canonical, and canonical + receivers sets already in group zero.

| Product selection | Traced seconds | Result |
|---|---:|---|
| TestProduct_NativeRuntimesSanitized0 | 39.048 | four settled |
| TestProduct_NativeRuntimesWASI0 | 34.651 | four settled |
| TestProduct_CommandRuntimeSanitized | 24.685 | found, zero settled |
| TestProduct_CommandRuntimeWASI | 20.196 | found, zero settled |
| TestProduct_CommandRuntimeWASICounted | 12.230 | one settled |

All 20 cmd/adamic top-level tests were selected individually in both build and read-only modes (40 passing invocations), not run through a whole-package test command. Every invocation was bounded by timeout 60 and -test.timeout=55s, including per-process request/header setup. validation.json and build-/read- logs preserve all timings; the largest cmd top-level invocation was 10.265 s. New/restored tests:

| Test | Build s | Read-only s |
|---|---:|---:|
| TestCommandRuntimeArchivesHaveOneBuilder | 7.431 | 7.184 |
| TestCommandRuntimeProductsWaitForOwners | 0.016 | 0.020 |
| TestSharedCommandRuntimeCannotBuildItsOwnersArchive | 2.283 | 2.307 |
| TestProduct_CommandRuntimeSanitized | 2.398 | 3.084 |
| TestProduct_CommandRuntimeWASI | 3.080 | 5.669 |
| TestProduct_CommandRuntimeWASICounted | 2.095 | 2.084 |
| TestCommandRuntimeProductKeyHoldsEveryPart | 1.747 | 1.857 |

Retained October 14 cold/warm measurements and fresh October 17 read-only observations, all in seconds. Historical timings are not presented as a new baseline; the fresh measurements overlap some validation work, so no causal comparison across dates or load is claimed. previous/ is the exact e54dc0a7 evidence, including its own previous/ history.

| Runtime-linking leaf | Historical cold | Historical warm | Fresh ahead |
|---|---:|---:|---:|
| TestExplainChecksOutput/overload_some_empty/build | 9.987 | 1.229 | 2.504 |
| TestExplainChecksOutput/overload_some_false_read/build | 9.414 | 1.306 | 2.471 |
| TestExplainChecksOutput/overload_some_true_only/build | 11.396 | 1.540 | 4.650 |
| TestExplainChecksOutput/overload_erased/build | 12.094 | 1.166 | 4.290 |
| TestExplainChecksOutput/overload_assertion/build | 9.092 | 1.198 | 3.228 |
| TestNonNullExplainChecks/narrowed/build | 8.963 | 1.223 | 3.312 |
| TestNonNullExplainChecks/present/build | 9.168 | 1.198 | 2.821 |
| TestNonNullExplainChecks/undefined/build | 8.887 | 1.433 | 3.531 |
| TestNonNullExplainChecks/null/build | 9.113 | 1.241 | 2.565 |
| TestNonNullExplainChecks/initializer/build | 9.978 | 1.216 | 5.642 |
| TestNonNullExplainChecks/let_initializer/build | 8.828 | 1.208 | 7.283 |
| TestNonNullExplainChecks/refuse_undefined/build | 9.304 | 1.220 | 5.860 |
| TestNonNullExplainChecks/refuse_null/build | 8.681 | 1.221 | 4.106 |
| TestRequestNativeStaysCommand | 9.417 | 1.213 | 3.191 |
| TestWASIRequest/request.a | 11.464 | 1.909 | 2.611 |
| TestWASIRequest/request_alias.a | 11.371 | 1.429 | 2.766 |
| TestWASIRequestThrows | 9.367 | 1.631 | 2.422 |

The earlier same-run comparison remains 185.523 -> 43.673 behavior-test core-seconds, saving 141.849 (76.5%). Archive production is work moved ahead, not free work. No fresh before/after CPU-saving claim is made for this restoration.

Five independently run mutants, all exit 1 for the intended assertion (patches and full logs beside mutants.py):

* second-owner.patch clears the sanitized owner. TestCommandRuntimeArchivesHaveOneBuilder reports two builder packages [internal/oracle cmd/adamic] for the actual base sanitized name key 3ac8c29ccf42, the same logical request Loom named a33c0b598f8a.
* second-builder.patch replaces BuiltElsewhere with Product. The real child product builds successfully, and TestSharedCommandRuntimeCannotBuildItsOwnersArchive fails because it must instead fail for the missing oracle owner. Child 12.61 s, guard 12.67 s.
* missing-wait.patch removes the sanitized wait; TestCommandRuntimeProductsWaitForOwners fails naming Sanitized0.
* missing-key-part.patch drops Portable[0]; TestCommandRuntimeProductKeyHoldsEveryPart fails naming adamic-runtime-v6.
* undeclared-sanitized.patch removes the sanitized declaration; the unchanged TestExplainChecksOutput/overload_some_empty/build fails with no runtime product declared.

Commands and observations:

* git fetch origin; git checkout -b compiler/loom-cmd-archives-3 origin/main. Base f93a65fba5a9d7dcd3be462391c8bc8bbef80d0a; a final fetch before commit still names that tip. Read CLAUDE.md, docs/front-door.md, the restored test files, product order, runtimeproducts and current cache APIs. No API adaptation was needed after that review; the two Go files remain byte-for-byte e54dc0a7.
* export GOPROXY='https://proxy.golang.org|direct'; timeout 240 bash cloud/setup.sh --wasi-sdk; source /workspace/adamic-tools/env.sh. Passed. nproc 5, cgroup cpu.max 400000 100000 (four CPU share). Node 24.19.0, Go 1.27.1, clang 20.1.8. Timing lines: Node 0.125 s, Go 0.127 s, submodules 0.188 s, markdown 0.312 s, clang 0.694 s, WASI 1.032 s, shared cache 2.631 s, go build 173.103 s, warm cache 173.959 s, done 174.300 s. setup.log has complete output; no setup workaround.
* timeout 120 go test -c -o /tmp/adamic-recut3.test ./cmd/adamic; timeout 120 go test -c -o /tmp/oracle3.test ./internal/oracle; timeout 90 go build -o /tmp/traced3 ./internal/buildcache/cmd/traced. Passed. Cmd binary rebuilt on final restored sources after mutants; Go compilation is outside each measured test leaf.
* timeout 450 python3 mutants.py. Each selected go test ./cmd/adamic -run SELECTED -count=1 -timeout 60s was bounded by timeout 90. First cold second-builder attempt exhausted that wrapper during Go compilation, before test output; after compilation, all five mutants were caught. Sources restored in finally. mutants-compile-timeout.log retains that initial attempt. No cmd test leaf exceeded its limit.
* timeout 540 python3 validate.py. Product runs: timeout 60 /tmp/traced3 -- BINARY -test.timeout=55s -test.v -test.run=^NAME$, working directory internal/oracle or cmd/adamic. ADAMIC_ORACLE_WASI=1, ADAMIC_BUILD_STORE=off, fresh ADAMIC_BUILD_CACHE_DIR. The first sanitized trace passed its test in 37.09 s but was refused by the tree-status guard because evidence files were added during the run. tree-status-refused-* preserves it. Corrected harness writes all live logs/reports outside the checkout until tracing ends. Final five traces settle/find with zero refusals. Live evidence was copied into review afterward.
* validate.py then runs each of the 20 cmd top-level Test functions separately in build mode and ADAMIC_BUILD_CACHE=read, same per-invocation limits. All pass. No assertion was edited. timeout 120 python3 measure.py /tmp/adamic-recut3.test after separately selects all 17 runtime-linking leaves in read mode; all pass. Every test process writes directly to a log file, never through a pipe.
* python3 check-traces.py: passed, nine observed runtime keys each with one builder and both shared cmd products finding the exact oracle keys.
* timeout 90 go test ./internal/compilercensus -run '^TestDependencies(FollowTheMarkers|CoverTheDeclaredMap)$' -count=1 -timeout 60s: passed, 1.468 s. These are main's compiler-dependency census and declaration-coverage tests, including cmd package consumers. No new package was added.
* timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s: final run passed, 61.224 s, exit 0. The initial cold command exhausted its outer 120 s wrapper while other cold compilations were running, with no completed test output; ir-readers-compile-timeout.log retains it. The final existing, unedited reader guard exceeded 60 s under the explicitly required 90 s timeout; it was not rerun after that result. All restored cmd tests and measured leaves are below 60 s.
* timeout 90 go vet ./cmd/adamic on final sources: passed, no diagnostics, exit 0. Source diff/formatting and JSON preservation checks passed. No full package tests or full gate; Loom owns those. No PR opened or Loom submission made.

Built: parked V4 checked callables and carried V3 arrays re-cut as compiler/loom-v4-2.
Commits: source 0998767348c4c20292cf989363dfc0822b0eb8f3; initial main 27d661a838681d3144808e108e3e61e32e940a69; final main base 6935e5933fbb74b28825a2fe221af3e0a65c1066.
Commands and outputs: 287 focused tests/products green with seven carried skips; 45 consumer packages compiled; counts, reader checks and vet passed.
Mutants: all 17 independent source/witness mutants caught and restored; exact catchers and seconds below.
Not covered: full gate or exhaustive consumer families; seven consumer checks exceed 60 seconds and one was interrupted at the driver limit. First-green push within 30 minutes was missed.

The branch tip is the single final commit above the reported base. It is rebased once after main advanced only in loom/rules.go, loom/rules.json and loom/rules_test.go. No compiler, runtime or counts changes landed during this validation. No PR or Loom submission is made.

The squash of origin/compiler/loom-v4 onto current main conflicted only in these two files, each resolved once:

- internal/flow/build.go: main c6e58480 (Three lowering gaps, on main) recognizes throwing object spreads through program.ObjectSpreadMayThrow. V4 recognizes ir.ArrayHoles as throwing. Keep both independent struct checks so getter exceptions and sparse-array RangeError propagate.
- internal/lower/exceptions.go: the same main change adds ObjectLiteral -> ObjectSpreadMayThrow to throwsOut. Keep that switch arm and V4's ir.Throw/ir.ArrayHoles arm plus checkArrayHoles, preserving both exception paths.

Main's C portability change 5f9f4b45 and getter-copy cleanup auto-merged with V4. ADAMIC_STRING_BYTES keeps main's unsigned-byte cast; object copying keeps main's adamic_thrown cleanup. No additional runtime code was introduced. Normalized runtime additions/deletions exactly equal the parked V4 commit's delta, with hunk context excluded. Every runtime file's bytes also match the pre-mutant snapshot after restoration. Evidence: runtime-delta-verified.json, runtime-hashes.json and runtime-delta.txt.

Runtime files changed against main:

- internal/native/runtime/adamic.h
- internal/native/runtime/array.c
- internal/native/runtime/array_holes.c
- internal/native/runtime/closure.c
- internal/native/runtime/heap.c
- internal/native/runtime/map.c
- internal/native/runtime/object.c
- internal/native/runtime/regexp_replace.c
- internal/native/runtime/union.c
- internal/native/runtime/view_arrays.c
- internal/native/runtime/view_arrays.h
- internal/native/runtime/view_callables.c
- internal/native/runtime/view_callables.h
- internal/native/runtime/view_unions_untagged.c

Counts: the auto-merged table passed the generated TestCountsAreRecorded shards without changes. No regeneration was necessary or performed in this recut. Command: timeout 210 go test ./internal/oracle -run ^TestCountsAreRecorded -count=1 -timeout 180s -parallel 4 -json; output counts-check.jsonl. Package elapsed 142.158 s, command wall 152.211 s; 1,078 pass records including shard parents, maximum leaf 8.670 s. The old exact anchor no longer selects main's generated top-level shards; use the prefix conservatively. Against main the 30 carried V4 rows are unchanged from parked V4: 26 V3 additions, three V4 compile-time escape refusals removed, and namespace_method_receiver retains/releases each decrease by two because declaration-only this is excluded from runtime argument slots. That attribution is a code reading. Every before/after row is in counts-diff.json.

Focused validation: run-focused.py compiled selected test binaries and executed each exact top-level name with -test.timeout=90s, -test.parallel=1 and an outer 95-second limit. Native products are built before consumer tests; products themselves have measured setup time. All 287 selected tests/products completed below 60 seconds, maximum 51.484 s. The selection includes V4 positives/refusals, array tests, embedded compiler/private-runtime mutants, RuntimeSetsAreEachUsed, product key checks, TestCallTargetReaders, new main getter-spread and accessor tests and hash-owned native Node fixture shards. Native and backend fixtures retain their source Node, ASan/UBSan and leak checks. Deliberate checked-view refusals retain their own documented diagnostic/exit assertions. Seven existing pending tests are explicitly skipped, not coverage claimed.

Rules 11 and 12: no new tests or product recipes were added by this recut, no changed compiler consumer package needs a declaration. V4 test splitting and placeless product declarations from the parked commit are retained. gofmt reported no changed Go file needing formatting. No other package's tests were weakened. No cohere implementation was copied; clean validation references the existing pinned submodule through a symlink.

Consumer validation: go list -json ./... identified 45 packages directly importing internal/flow or internal/lower in production/test imports. All 45 compiled within the 180-second build bound, and vet covered those 45 plus internal/javascript and internal/load. Two packages have no tests. The 43 test packages have very large inventories, so run-consumers.py runs up to three product preparations and three representative exact top-level tests per package; selected-consumer-tests.json names the actual selection. This is bounded representative coverage, not an exhaustive package test run. Consumer driver timeout was 840 seconds; each test was bounded by 90 seconds internally and 95 seconds externally. There are 168 completed consumer records. Parent partial skips are preserved in logs; the CSV skip field can indicate a skipped subtest rather than the entire parent. Product/test checks over 60 seconds are red under the grain law even where their assertions passed, and were never rerun.

Consumer limit observations:

| Package | Test | Seconds | Observation |
|---|---|---:|---|
| bridge/tsgo | TestProduct_ApiDriver | 80.210 | 80-second product deadline killed bridge-test-tsgo-asan.a; cgroup oom_kill=0, not a demonstrated runtime defect |
| stage1/cohere/lint/helpers | TestProduct_HelperOracle | 90.114 | Go 90-second test timeout |
| stage1/cohere/lint | TestProduct_CompilerAgreementLowered | 65.841 | assertions passed, exceeds 60-second law |
| stage1/cohere/suppression | TestThePortParsesAsGoCohereDoes | 63.499 | assertions passed, exceeds 60-second law |
| stage1/cohere/static_single_assignment | TestThePortMakesWhatGoCohereMakes | 86.477 | assertions passed, exceeds 60-second law |
| stage1/cohere/typeaware | TestProduct_checker_mutant_declaration_source | 90.124 | Go 90-second test timeout |
| stage1/cohere/values | TestThePortParsesAsGoCohereDoes | 64.268 | assertions passed with partial skips, exceeds 60-second law |

For product failures/overruns, remaining selected checks in that package were blocked. stage3/fixtures TestFixtureOracleProductIgnoresGoCacheLocation started but has no completed result before the 840-second driver deadline; it is not counted as passing or rerun. Its three preceding selected tests/products passed. All other selected completed failures were setup issues corrected as follows.

MarkdownInline_0000 and _0001 initially failed in 0.032/0.033 s because staged incoming Markdown differs from main HEAD. A clean detached commit-tree snapshot of the staged code/docs (not a branch commit) fixed the validation context. They then passed in 26.275/22.347 s. Scanner shards 000/001/002 initially failed in 10.120/6.933/11.030 s because their in-test clone could not connect to GitHub. An explicit bounded git clone of v6.0.3 succeeded; ADAMIC_TYPESCRIPT_SOURCE names that verified pinned checkout for validation only. The clean snapshot probes then passed in 6.648/2.424/9.515 s. No machine path was added to a product/test source. Snapshot metadata and exact rerun commands are in clean-snapshot.json and run-clean-consumers.py. The detached snapshot includes all incoming V4 Markdown, excluding this new evidence directory.

Final restored checks: 13.053 s for the exact source-mutant catcher union plus TestCallTargetReaders/TestProduct_CallTargetExports, all pass. Vet: 10.889 s, exit 0, no diagnostics across 47 packages. final-checks.json records the full argument vectors; restored-controls.log contains every Go leaf's elapsed seconds. test-seconds.csv and all-results.json contain every completed focused, consumer, mutant and count record, with build and process wall times labeled separately.

Independent mutants, all exit 1 due to test assertions (no build failures):

| Mutant | Catcher | Seconds | What failed |
|---|---|---:|---|
| escape-Argument | TestV4EscapeArgumentRefusal | 19.566 | required escaping-read path/fix refusal missing |
| escape-Result | TestV4EscapeResultRefusal | 6.407 | required escaping-read path/fix refusal missing |
| escape-Passed | TestV4EscapePassedRefusal | 6.160 | required escaping-read path/fix refusal missing |
| escape-Returned | TestV4EscapeReturnedRefusal | 6.803 | required escaping-read path/fix refusal missing |
| escape-MixedProducer | TestV4EscapeMixedProducerRefusal | 6.625 | required escaping-read path/fix refusal missing |
| unsupported-domain | TestV4DirectRecursiveRefusal | 16.991 | required recursive-domain refusal missing |
| unsupported-producer | TestV4DirectProducerRefusal | 17.481 | required producer refusal missing |
| surface-borrow | TestV4EscapeAdapterName | 12.532 | ASan adapter root use-after-free |
| bind-name-borrow | TestV4EscapeAdapterBindNameOwnership | 13.503 | ASan name string use-after-free |
| identity-root | TestV4EscapeAdapterIdentity | 35.135 | Node stdout comparison |
| metadata-inline | TestPortElementMetadataEmission | 11.226 | 2,885 inline stores violate emission bound |
| metadata-omitted | TestPortElementMetadataEmission | 10.326 | zero runtime certificates instead of 2,885 |
| native-order | TestV4DirectCalleeOrder | 13.626 | argument output missing after premature callee check |
| javascript-order | TestV4DirectCalleeOrder | 10.891 | argument output missing after premature callee check |
| construct-witness | TestV4EscapeAdapterWithoutConstructSignatureRefusal | 3.934 | required construct-signature refusal missing |
| boxed-witness | TestV4Lane5BoxedAdamicRefusal | 5.995 | required unproven-read refusal missing |
| factory-witness | TestV4Lane5FactoryWiderWriteRefusal | 5.746 | required wider-write path/fix refusal missing |

All source bytes restored after each mutant, patches stored as .patch. The driver initially used the wrong witness directory after the 14 compiler/runtime mutations; it stopped with FileNotFoundError. Only the three not-yet-run witness mutations were run after correcting the evidence script. No successful mutant was rerun.

Setup: GOPROXY=https://proxy.golang.org|direct; timeout 600 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Completed 51.744 s. nproc=5; cpu.max=400000 100000 means four CPU shares. Setup timing lines follow.

```text
setup: node ready (0.035s)
setup: go ready (0.058s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.011s
setup: markdown dependencies ready (0.093s)
setup: submodules ready (0.108s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.271s)
setup: shared cache ready (9.366s)
setup: go build ready (51.635s)
setup: test binaries deferred (use --warm-tests) (51.717s)
setup: build cache warm (51.719s)
setup: build-flags commit=27d661a838681d3144808e108e3e61e32e940a69 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/136 835 load-after=3.80 1.06 0.36 1/144 3407
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (51.744s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.IUNvQI
```


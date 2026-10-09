Nineteen bridge artifacts now have TestProduct units and a shared buildcache.Product recipe; the convention is documented.
The doc commit is a44c7eae, the test implementation is 73f480a9, and the main merges are 20f1ed30 and 140ebf83.
All nineteen products, nineteen input guards, sample bridge checks and measured compiler-corpus checks pass within 60 seconds.
Nineteen omitted-input mutants serve stale products and fail; removing one registered product fails the count assertion.
No production code, oracle fixtures, gate implementation or shared-store implementation changed; other compiler-corpus leaves were not rerun.

This serves step 79 and task #sykbpks: builds have their own units instead of making a bridge observation carry cold setup.
The first commit contains only the requested docs/test-products.md convention, preserving its words and restoring markdown.
All implementation changes are in _test.go. Every branch commit is the allowed doc, tests, review evidence or a merge of main.
Main b5245943 and cd0ecb04 were merged without conflicts; they changed unrelated estree, printer, regexp, lint and YAML tests.

The same bridgeProduct recipe serves all nineteen product tests and every bridge consumer. There are no sync.Once builders.
Every product test calls t.Parallel first, has no subtests and only fetches its product. Product keys contain sorted source,
C and assembly files, embedded files, module metadata and local module sums from the Go dependency graph. Discovery uses
Go list without -export, so a hit does not compile dependencies. External module sources are fingerprinted as named flags,
since buildcache.Files accepts repository-relative paths. No dependency source is copied.
Build flags include effective Go configuration, the archive and native modes, sanitizer and counting flags, compiler options,
bridge source recipes and build-related environment values. Toolchain names Go, clang and the effective C tools.
The shared full dependency graph is conservative: changing an unrelated bridge artifact can invalidate extra products.
The complete manifests and keys are in product-inputs.json. Go action caches remain keyed and validated by Go.

runBridgeCase fetches every required product before starting its existing budget clock. ADAMIC_UNIT_BUDGET=1 enables
that clock's budget on the reference instance; there is no unconditional deadline added. Existing subprocess hang guards remain.
The old nineteen artifacts and thirty-six bridge observations remain covered; coverage.json lists old products to new tests.
TestBridgeProductUnitsCoverEveryProduct counts nineteen product bindings and checks every case's declared products.
Its planted failure removes the native binding and fails with "19 products, 18 units, want 19".

For each product, its input guard uses that product's exact Inputs against a small test-owned source tree. It changes the
recipe content and fetches a marker product through the resulting key and buildcache.Get. Each overlay drops only
bridge/tsgo/products_test.go from one product's inputs. The second fetch serves "version one" instead of "version two"
and the corresponding guard fails with "served stale product". These are cache behavior failures, not compilation failures.
All nineteen unmutated guards pass separately on the final source; their seconds and every mutant result are below.

Cold measurements run each TestProduct alone in a fresh go test process, -count=1, -json, -timeout 90s, GOMAXPROCS=4,
ADAMIC_GATE_UNCACHED=1 and a fresh runtime cache (XDG_CACHE_HOME). Each product's own entry is absent: its log records
one miss for that product. Prerequisite products share the cache and are built in the listed order. The Go action cache
prepared by setup and earlier builds is retained; these are cold product and runtime caches, not an empty Go toolchain cache.
The initial archive run also built cold Go archive actions and took 56.510 seconds including process setup; its logs remain
under initial-. After the external-module key audit, every mutant and measurement was rerun. The final-source table follows.

| Product unit | Artifact | Cold seconds including setup | Input guard seconds | Omitted-input mutant |
| --- | --- | ---: | ---: | --- |
| TestProduct_TsgoArchive | tsgo.a | 4.421 | 2.225 | caught |
| TestProduct_TsgoSanitizedArchive | tsgo-asan.a | 4.502 | 2.315 | caught |
| TestProduct_LengthArchive | length.a | 4.331 | 2.266 | caught |
| TestProduct_StaleArchive | stale.a | 4.380 | 2.438 | caught |
| TestProduct_WrongArchive | wrong.a | 4.371 | 2.215 | caught |
| TestProduct_LeakArchive | leak.a | 4.619 | 2.487 | caught |
| TestProduct_Stage0 | stage0 | 5.151 | 2.310 | caught |
| TestProduct_Oracle | oracle | 4.405 | 2.285 | caught |
| TestProduct_ApiDriver | api | 3.284 | 2.223 | caught |
| TestProduct_LengthDriver | length-driver | 3.556 | 2.362 | caught |
| TestProduct_StaleDriver | stale-driver | 3.321 | 2.355 | caught |
| TestProduct_LeakDriver | leak-driver | 3.060 | 2.348 | caught |
| TestProduct_SanitizedNative | native-asan | 9.080 | 2.329 | caught |
| TestProduct_Native | native | 6.639 | 2.411 | caught |
| TestProduct_WrongNative | wrong-native | 8.942 | 2.453 | caught |
| TestProduct_HealthyRegion | healthy-region | 8.740 | 2.262 | caught |
| TestProduct_RegionStage0 | region-stage0 | 4.693 | 2.305 | caught |
| TestProduct_RegionNative | region-native | 9.115 | 2.389 | caught |
| TestProduct_LinkageTest | linkage.test | 7.828 | 2.310 | caught |

All sixteen active sample bridge leaves pass, including ABI, sanitizer, ownership, refusal and wrong-position checks.
On TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, the oracle and first timing checker leaves pass.
Each consumer starts in a fresh Go test process. Every product census line in these consumer logs is a hit.
The two-leaf invocation uses the same Go test selection as a pooled unit; it does not rebuild a product.

| Consumer invocation | Seconds including process setup | Builds |
| --- | ---: | --- |
| TestBridgeABI | 2.645 | cache hits only |
| TestBridgeOracleSample | 3.732 | cache hits only |
| (?:TestBridgeABI|TestBridgeOracleSample) | 4.267 | cache hits only |
| TestBridgeOracleChecker | 4.206 | cache hits only |
| TestBridgeTimingRound1Checker | 4.970 | cache hits only |
| (?:TestBridgeOracleChecker|TestBridgeTimingRound1Checker) | 6.995 | cache hits only |

Sample leaf measurements:

| Leaf | Seconds |
| --- | ---: |
| TestBridgeABI | 2.872 |
| TestBridgeInputLength | 2.971 |
| TestBridgeUnlinkedBuild | 2.827 |
| TestBridgeUnlinkedC | 2.672 |
| TestBridgeUnlinkedJavaScript | 2.820 |
| TestBridgeOutputLength | 2.897 |
| TestBridgeStaleHandle | 2.944 |
| TestBridgeLinkage | 2.917 |
| TestBridgeOutputFree | 3.064 |
| TestBridgeRegion | 3.733 |
| TestBridgeRegionOwnership | 2.703 |
| TestBridgeOracleSample | 3.543 |
| TestBridgeTimingRound1Sample | 3.795 |
| TestBridgeTimingRound2Sample | 3.873 |
| TestBridgeTimingRound3Sample | 3.955 |
| TestBridgeWrongPositionSample | 3.462 |

Commands and exact output are in results.json, individual JSON logs, run.py and consumers.py.
The merged-tip check runs only TestProduct_, TestBridgeProductInput_, and the three coverage/cache tests;
all pass and all real product fetches are hits. No whole package tests or full gate ran. No fixture was added,
so internal/oracle/counts.md is unchanged. The other eighteen compiler-corpus leaf selections, other platforms,
a completely empty Go action cache and shared-store transport are not covered by this unit.

Setup: Go ready 0.025 s, Node ready 0.027 s, submodules ready 0.072 s, markdown dependencies 0.076 s,
clang ready 0.187 s, Go build ready 38.870 s, setup complete 39.166 s. setup.txt holds its exact lines.
nproc is 5; cpu.max is 400000/100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node 24.19.0.
Old regenerable Go cache files and this unit's superseded products were cleared to provide disk space.

All nineteen products pass again after the final main merge. Their nineteen fetches are all hits;
final-main-products.jsonl and final-main-products.builds.txt record the result. No input key moved in either main merge.

Mandated integration lane checks on the committed merged tip:

```text
lane checks 0.9 s: gofmt and tools on 4 Go files, t.Parallel on 1 test packages; vet 1 packages
```

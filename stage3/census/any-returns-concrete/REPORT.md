Built: resolved checker return contracts and checker-owned mapper preservation in census snapshots.
Compiler commits: a4747873 and a64f77ee0e9f0591ef1a4ed005de7a9213b7ce85, on requested area tip 784b577a (the latter is measured).
Validation: targeted lowering and Node oracle tests pass; counts refresh passes (21 allocations, 21 frees for the new fixture).
Mutants: unsubstituted return, trusting any, opaque mapper, and copying the mapper alias are each caught.
Limits: the corpus remains checker-rejected measurement input; the memoize fixture preserves the generic callback/return shape with eager capture.

## Observations and root

On the frozen census compiler ed6e2975, `signature` in `internal/lower/functions.go:75` gets `() => T` from `GetReturnTypeOfSignature(GetSignatureFromDeclaration(memoize))`. It then calls `lowering.concrete`, `internal/lower/instantiate.go:56`, which calls checker instantiation. The first memoize in adapted `emitter.ts:701` is `memoize(() => getCommonSourceDirectoryOfConfig(configFile, ignoreCase))`: its resolved checker signature returns `() => string`. Lowering emitter alone gives `() => string`; lowering checker.ts before emitter makes all six memoize declarations report `() => T` before concrete and `any` afterward. See `evidence/memoize-trace.log.txt` and the standalone comparison log.

The census state generator treats `typeMapper struct{}` (`instantiate.go:17` on the frozen compiler) as lowering-owned data. Generated `latentCopyPointer_typeMapper` allocates an empty local struct instead of preserving the actual checker mapper pointer. The mapper is a private checker structure held through a linkname bridge, so that copy loses its contents and identity. The actual any is the checker's error type: `Checker.instantiateTypeWithAlias`, referenced through the cohere submodule at `cohere/TypeScript/tsc/internal/checker/checker.go:22570`, detects an instantiation stack of 100 and returns `c.errorType` at line 22586. Trace records show depth 99 before recursive concrete transitions and a zero-sized copied mapper at depth 100. These are observed transitions, not TypeScript source any annotations.

The causal inference is snapshot corruption followed by recovered instantiation panics leaving the shared checker's stack polluted. It is tested by a controlled full census that changes only mapper identity/snapshot preservation on ed6e2975, and separately by the delivery census on the compiler commit above. This report does not credit disappearance of an any boundary as successful native compilation of the TypeScript compiler.

## Changes

`typeMapper` now aliases `checker.TypeMapper`; the state generator follows its foreign selector type and retains the pointer, as it already does for other checker-owned pointers. Unknown copier shapes still fail loudly. No cohere file is copied or changed.

Generic specialization reads the mapper from the checker's resolved signature, validates the exposed field type, and applies it to the selected signature's binders. The existing overload proof establishes positional alpha-renaming into implementation binders, including overloads with extra predicate binders. The active class specialization remains authoritative where classGenericCall has already installed it.

Before entering a callee's mapper, specialization reads its resolved signature return and substitutes the caller's enclosing mapper. The resulting exact checker type is supplied to `signatureReturn`; the declaration's unsubstituted return is not used instead. Missing signatures, missing mappers, unknown binder correspondences, unresolved/any type arguments, and a resolved any return keep named refusals. An explicit any is refused with `a generic function whose resolved return type is any`.

## Fixtures and mutants

`internal/oracle/testdata/any_returns_concrete.a` reduces TypeScript's memoize, forEach, filter, setTextRange and append. The last three rank immediately after the leading source-any JSON helper in RESULT.json by hidden bytes, at 4,474, 4,215 and 4,114 bytes. The fixture exercises function-of-T capture, forEach's U-or-undefined and falsey callback result, optional arrays, preserved extra object fields and identity, and append's present/absent branches. Memoize evaluates its callback once before capture: lazy captured-callback mutation is a separate existing closure limitation, so this fixture does not claim memoize's lazy timing coverage.

All five go through the Node oracle in both JavaScript and native backends. Native comparisons include the sanitized and release builds. The checker-contract test additionally asserts exact checker strings and the named any refusal.

`mutants.py` changes one source at a time and restores it in `finally`; every failure is a running test failure, not a build failure:

| Mutant | Catch |
| --- | --- |
| Return the declaration's unsubstituted type | TestGenericResolvedReturnContracts: memoize got `() => T`, wanted `() => string` (other generic contracts fail too) |
| Trust resolved any | TestGenericResolvedReturnContracts: required named refusal disappears |
| Restore opaque local mapper struct | TestGenericMapperKeepsCheckerIdentity |
| Make copier ignore foreign alias | TestCheckerMapperAliasKeepsIdentity: unsupported foreign pointer |

## Commands and environment

All test output was redirected to log files. Final area logs are `evidence/area-*.log.txt`; older logs without the area prefix are preliminary main validation or frozen census experiments. Delivery contains only this unit's commits on the explicitly requested area tip; no other worker's unlanded branch is merged. The initial main-based probe is local only and was not pushed.

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh > /tmp/any-returns-setup.log 2>&1`, then `source /workspace/adamic-tools/env.sh`.
- Setup timing lines: Node 0.064s, Go 0.065s, clang 0.510s, markdown dependencies 1.030s, submodules 20.494s, Go build 191.613s, done 191.924s. `nproc`: 5; cgroup quota: 4 CPUs. Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup succeeded.
- On the area tree: `npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund`; installs pinned TypeScript 6.0.3 and Node types 25.3.3. Earlier counts runs failed for missing Node declarations and exposed the overload binder issues described above; final counts pass.
- `go test ./internal/lower -run 'TestGeneric|TestClassGeneric|TestNominal' -count=1`: pass, 0.176s.
- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(any_returns_concrete|census_overload_contracts|census_append_overload|class_inheritance_generic|generic_instance_key_)' -count=1 -v`: pass, 1.608s, five selected fixture subtests.
- `go test ./stage3/census/latent/statecopy -run TestCheckerMapperAliasKeepsIdentity -count=1`: pass.
- `python3 stage3/census/any-returns-concrete/mutants.py`: all four caught; individual failure logs included.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`: pass, 49.676s. The only counts change is the new fixture row.
- `python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/any-returns-delivery-overlay` and `go build -overlay /tmp/any-returns-delivery-overlay/overlay.json -o /tmp/any-returns-delivery-census ./stage3/census/latent/tool`.
- `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/any-returns-delivery-census /tmp/any-returns-adapted/src/compiler /tmp/any-returns-delivery.jsonl > /tmp/any-returns-delivery-run.log 2>&1`.

No whole package test run or full gate was invoked. Two early oracle commands used the wrong subtest expression and selected no tests; they were corrected and are not counted as validation.

## Census

The frozen adapted corpus is verified against all 82 original stock manifest byte lengths and SHA-256 hashes. The summary references Adamic's frozen hidden.py and ranking.py through git objects, preserving hidden-byte credit rules. RESULT.json contains the delivery ledger hash, measured compiler SHA, before/after totals, and every remaining callee. The controlled mapper-only result is separate evidence.

The controlled mapper-only census finished at 15 boundaries / 7,290 credited hidden bytes, down 2,045 boundaries / 82,337 bytes from the original. Its 13 remaining callees include the eight source-any declarations and five zero-byte Node-context boundaries. The latter are `getNodeSystem.statSync`, `cleanupPaths`, `fsWatchWorker`, `getModifiedTime`, and `deleteFile`: the measurement loader leaves legacy `import("fs")` and `import("inspector")` unresolved (TS2591 at sys.ts:1471, 1630 and 1673). Stats, Profile, FSWatcher, Date-or-undefined and void remain non-any in the stock checker. These refusals are retained; removing them would trust an unresolved type. Installing Node declarations for the repository's `node:` test fixtures does not resolve these legacy imports in this frozen census loader. See `evidence/node-context.log.txt`.

The delivery census on **a64f77ee** also finishes at **15 boundaries / 7,290 credited hidden bytes**, from **2,060 / 89,627**: a reduction of **2,045 boundaries / 82,337 bytes** in the any-return bucket. All 82 source hashes match the original. Of the original 167 non-any stock callees, 162 lose the invented-any label; the five unresolved Node contexts remain named refusals. The eight source-any callees account for 10 boundaries and all 7,290 remaining credited bytes.

| Remaining callee | Boundaries | Credited bytes | Why refused |
| --- | ---: | ---: | --- |
| `convertConfigFileToObject` | 2 | 1545 | The tsc declaration explicitly returns any; JSON conversion is a separate source contract. |
| `convertToObject` | 1 | 150 | The tsc declaration explicitly returns any; JSON conversion is a separate source contract. |
| `convertToJson` | 2 | 4862 | The tsc declaration explicitly returns any; JSON conversion is a separate source contract. |
| `convertToJson.convertObjectLiteralExpressionToJson` | 1 | 0 | The tsc declaration explicitly returns any; JSON conversion is a separate source contract. |
| `convertToJson.convertPropertyValueToJson` | 1 | 0 | The tsc declaration explicitly returns any; JSON conversion is a separate source contract. |
| `setTimeout` | 1 | 89 | The tsc host declaration explicitly returns any. |
| `getNodeSystem.statSync` | 1 | 0 | import("fs").Stats is unresolved (TS2591); stock return is Stats or undefined. |
| `getNodeSystem.cleanupPaths` | 1 | 0 | import("inspector").Profiler.Profile is unresolved (TS2591); stock return is Profile. |
| `getNodeSystem.fsWatchWorker` | 1 | 0 | _fs.watch has unresolved fs typing; stock return is FSWatcher. |
| `getNodeSystem.getModifiedTime` | 1 | 0 | Unresolved statSync result propagates through mtime; stock return is Date or undefined. |
| `getNodeSystem.deleteFile` | 1 | 0 | _fs.unlinkSync has unresolved fs typing; stock return is void. |
| `tryParseJson` | 1 | 104 | Explicit any return; JSON.parse also returns any in lib.d.ts. |
| `objectAllocator.getNodeConstructor` | 1 | 540 | Returns Node as any; the source assertion erases its constructor contract. |

Final recount command: `python3 stage3/census/any-returns-concrete/summarize.py /workspace/adamic /tmp/any-returns-adapted/src/compiler /tmp/any-returns-delivery.jsonl stage3/census/any-returns-concrete/RESULT.json a64f77ee0e9f0591ef1a4ed005de7a9213b7ce85`. Output: `{"boundaries": 15, "hidden_bytes": 7290}`. The measurement-only run and all output guards exited successfully. Both complete ledgers are retained in evidence as `.jsonl.gz`; RESULT.json records the delivery ledger's uncompressed SHA-256. The binary's Go build metadata records the measured revision with `vcs.modified=false`.

The subsequent report commit changes no lowering, oracle, or state-generator source relative to the measured compiler commit. The branch is pushed only after this complete result and local tests/mutants; the shared fast gate is left to the authorized push workflow.

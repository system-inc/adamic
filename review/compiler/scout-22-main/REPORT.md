Built step 22 compiler support for proven borrowed own-property queries and object string conversion on current main; retained prototype lowering behind an explicit runtime refusal.
Taken commits: 71f91687, 6945a8a4, f2ffe42d, 3cad8010, 5c8827fc; local transplant commits: b27ff936, 032f975d, b7865af1, 16160078, a2ac4027.
Validation: focused uncached lower/native/oracle checks passed in 1.983s/0.045s/4.032s; Node census passed; counts regeneration passed in 70.050s; dependency census passed.
Mutants: inherited-as-own, reversed entries, wrong conversion hint and suppressed prototype operand escape were caught by Node comparison or region-plan assertions.
Not covered: runtime error identity, catchable library failures, prototype storage and lifetime execution, WASI, the whole packages or the full gate; every dependent test is listed below.

The five source commits are the contiguous scout series from 71f91687 through 5c8827fc. No runtime branch was merged. The final diff changes no file under internal/native/runtime. Current main already proves cached intrinsic signatures through its library method adapter, so compiler-cached-intrinsic.a is now held positively against Node and both backends instead of being forced back into a refusal. The narrow explicitly typed scout adapter keeps the checked local read, rather than erasing the cached binding. An inferred intrinsic alias escaping as a value remains refused; main's proven inferred call/apply adapters keep their existing oracle coverage.

The conservative assumptions are: prototype graphs cannot be admitted before runtime storage and edge ownership land; conversion results containing null remain NotYet until null and undefined are represented distinctly; numeric or mixed object addition, hidden conversion members, unrepresented results and Symbol.toPrimitive remain NotYet. Diagnostics carry the source path. Prototype failures give composition as the fix. Mixed-cycle diagnostics retain the field name, [[Prototype]] link and compiler graph-region proof reason. The compiler region escape rule and native prototype ABI emission are kept; the final lowering boundary refuses execution before those unavailable runtime functions can reach clang.

Runtime-owned hunks excluded from f2ffe42d:
- runtime/dtoa.c: RangeError delivery for exponential and precision arguments.
- runtime/library_errors.c and library_errors.h: tagged TypeError/RangeError identity and catchable concat/length guards.
- runtime/library_object.c: catchable frozen property writes and assignment failures.
- runtime/normalize.c: catchable invalid normalization-form errors.
- runtime/number.c: catchable toFixed bounds errors.
- runtime/radix.c: catchable radix validation.
- runtime/string.c: library error header inclusion.
- runtime/string_repeat_impl.h: catchable repeat and padding count/length errors.

Runtime-owned hunks excluded from 3cad8010:
- runtime/adamic.h: prototype header flags and edge layout declarations.
- runtime/library_prototypes.c and library_prototypes.h: create/get/set, intrinsic prototype identity, retained prototype edges and cycle rejection.
- runtime/class_inheritance.c and object.c: prototype initialization.
- runtime/region.c: prototype teardown at region end.
- runtime/share.c: prototype ownership integration in the obsolete runtime sharing translation unit.

Associated activation hunks withheld until the runtime half is present: class_errors.go error tagging; exceptions.go specific error identity dispatch, library exception effects and removal of existing libraryFailure refusals; flow/build.go LibraryMayThrow activation; JavaScript error identity and stack-operation dispatch from an unrelated old context; native library_object.go error identity and catchable assign activation; emit_expressions.go catchable concat/toFixed activation; emit_numbers.go and emit_strings.go catchable formatting/repeat/padding/normalization/concat activation. The LibraryMayThrow compiler declaration is retained without activation. The deleted native/async.go and docs/runtime-statics.md stay deleted on main. Runtime-specific third-party notice claims were excluded; OrdinaryToPrimitive compiler attribution is retained. The old class_features_accessors.a refusal change is excluded because concat remains on main's runtime path and the existing accessor fixture remains executable.

The obsolete witness files already deleted on main stayed deleted when taking 71f91687. Its surviving change removes review/unknown-narrowing/require_pending.a. No stale stage1/cohere/config consumer was introduced. The exact dependency-census method from origin/devtools/fast-gate was run using go list -deps -test -json ./... for its package map; no new consumer needs a manifest declaration. Existing build-metrics sources without test packages were reported by the tool without failing the census.

Commands and observed results:
- export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh: first attempt failed during cache warming on a temporary cherry-pick conflict marker. The resolved rerun passed. Timing lines: Node 0.031s, Go 0.033s, markdown 0.096s, submodules 0.095s, clang 0.183s, build 39.011s, cache warm 39.117s, done 39.150s. nproc: 5; CPU quota: 4. source /workspace/adamic-tools/env.sh was used in every subsequent test shell. /opt/adamic-tools/env.sh does not exist here.
- npm ci --prefix stage3/api: installed the repository's pinned TypeScript 6.0.3 and @types/node 25.3.3. This resolved the missing node:* declaration input for the compiler toString witness.
- CENSUS_TYPESCRIPT=/workspace/adamic/stage3/api/node_modules/typescript CENSUS_TYPE_ROOTS=/workspace/adamic/stage3/api/node_modules/@types node --test stage3/scout/22-tsc-objects/census.test.cjs: one pass, zero failures, 1.247s. The inflated Object.keys count control was rejected by the independently specified expected census.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/oracle -run 'TestBorrowedObject|TestScout22|TestPrototypeArgumentsEscapeRegions|TestObjectRefusals|TestObjectUnproven|TestNativeAgreesWithNode/internal/oracle/testdata/(scout|library_method_values|method_coverage_object_descriptors|library_string_raw)' -count=1 -v: all selected active tests passed. Active fixture executions include source Node, generated JavaScript, native ASan/UBSan and Linux LeakSanitizer. Six pending fixture sources also completed on Node with exit zero and no stderr.
- go test -overlay review/compiler/scout-22-main/region-overlay.json ./internal/native -run '^TestPrototypeArgumentsEscapeRegions$' -count=1 -v: exit 1 with create and setPrototypeOf operand escape assertions. The mutant source is .go.txt and is never a tracked compilable Go file.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts: the first attempt failed before writing the table, exposing interception of main's proven inferred aliases and undefined raw-template conversion. The corrected attempt passed and regenerated the table once successfully. No dependent fixture was registered as executable or counted.
- git diff --check: exit zero.

Each test's measured seconds from the final focused log (top-level parallel scheduling is reported separately by Go from subtest work):

| Test | Seconds |
|---|---:|
| TestScout22NullConversionResultStaysRefused | 0.18s |
| TestScout22PrototypeRuntimeBoundary | 0.44s |
| TestObjectRefusalsExplainSoundness | 0.00s |
| TestObjectUnprovenShapesStayNotYet | 0.00s |
| TestBorrowedObjectOwnProperty | 0.00s |
| TestBorrowedObjectOwnPropertyDoesNotTrustAnnotations | 0.00s |
| TestBorrowedObjectOwnPropertyRefusesUnprovenUses | 0.01s |
| TestScout22Slice2CompilerAndPrototypeGaps | 1.93s |
| TestPrototypeArgumentsEscapeRegions | 0.00s |
| TestScout22CachedIntrinsicProvenance | 1.20s |
| TestScout22KeywordOrderMutant | 1.29s |
| TestScout22OwnPropertyMutant | 1.35s |
| TestScout22MutantConversionHints | 1.40s |
| TestScout22PendingSourcesOnNode | 2.25s |
| TestScout22ConversionHints | 1.09s |
| TestScout22TscToString | 1.54s |
| TestNativeAgreesWithNode | 0.39s |

The oracle fixture leaves in the final log all finish below 60 seconds. The unchanged aggregate counts test completed in 70.050 seconds; its fixture leaves were not changed into one combined test. The active new top-level tests all finish below 60 seconds.

Mutant outcomes:
- TestScout22OwnPropertyMutant: inherited constructor accepted as own; both backend stdout comparisons disagree with source Node. Native exits zero with no sanitizer error or leak.
- TestScout22KeywordOrderMutant: Object.entries reversed; both backend stdout comparisons disagree with source Node. Native exits zero without stderr or leaks.
- TestScout22MutantConversionHints: default hint replaced by string hint; both backend stdout comparisons disagree with source Node. Native exits zero without stderr or leaks.
- region-overlay.json: escaping prototype arguments marked nonescaping; TestPrototypeArgumentsEscapeRegions fails the operand lifetime assertion for create and setPrototypeOf. No compilation or warning failure counted as a kill.
- Node AST census: an inflated Object.keys count raises AssertionError against the independent expected counts.
- All ten catchable-error mutants and the prototype-link semantic mutant remain in their test bodies. Their runtime-dependent tests were skipped and are not reported as killed.

Every pending top-level test and its exact skip reason:

| Test | Skip reason |
|---|---|
| TestScout22ConversionErrors | awaits runtime/step22-to-primitive: TypeError and RangeError identity needs runtime error tags |
| TestScout22PropertyErrors | awaits runtime/step22-to-primitive: frozen writes and Object.assign need catchable runtime TypeErrors |
| TestScout22RangeErrors | awaits runtime/step22-to-primitive: repeat, padding, normalization and number formatting need catchable runtime RangeErrors |
| TestScout22MutantConversionErrors | awaits runtime/step22-to-primitive: conversion_errors mutant needs its catchable error fixture |
| TestScout22MutantPropertyErrors | awaits runtime/step22-to-primitive: property_errors mutant needs its catchable error fixture |
| TestScout22MutantRangeErrors | awaits runtime/step22-to-primitive: range_errors mutant needs its catchable error fixture |
| TestScout22MutantPadStart | awaits runtime/step22-to-primitive: padStart mutant needs its catchable error fixture |
| TestScout22MutantNormalize | awaits runtime/step22-to-primitive: normalize mutant needs its catchable error fixture |
| TestScout22MutantToFixed | awaits runtime/step22-to-primitive: toFixed mutant needs its catchable error fixture |
| TestScout22MutantToExponential | awaits runtime/step22-to-primitive: toExponential mutant needs its catchable error fixture |
| TestScout22MutantToPrecision | awaits runtime/step22-to-primitive: toPrecision mutant needs its catchable error fixture |
| TestScout22MutantToString | awaits runtime/step22-to-primitive: toString mutant needs its catchable error fixture |
| TestScout22MutantErrorIdentity | awaits runtime/step22-to-primitive: error_identity mutant needs its catchable error fixture |
| TestScout22PrototypeLinks | awaits runtime/step22-prototype-links: prototype identity, retained links and cycle rejection require the runtime half |
| TestScout22PrototypeLifetimes | awaits runtime/step22-prototype-links: replacement and teardown need runtime retain and release of prototype edges |
| TestScout22PrototypeCreate | awaits runtime/step22-prototype-links: Object.create and getPrototypeOf need runtime prototype storage |

Logs, mutant source and census result are stored beside this report. Integration lane output is recorded in lane-checks.log after the committed branch is checked. No pull request was opened.

# 99q: core type-only walls

Base: origin/typescript/flatmap-adapt-2 at 0f613b285cc0d3314f1217b78f65ad94571c5547.
Only this new adaptation directory changes. The compiler used for every probe is
unmodified main as fetched at the start, d40c2470e; its compiler, CLI, census, go.mod and go.sum
are identical to the requested base. A later main recheck reached 006b368745e6a37f18b1427b1a63a78285f92694. Its new changes concern Map/Set consumption of generator calls, not these type-only relations. The complete source diff is retained. The census remains explicitly pinned to d40c2470e.
No production compiler changes, suppression,
backend changes, or partial-root fallback were used.

| Site | Decision | Evidence |
| --- | --- | --- |
| H02 deduplicate | Keep one type argument: `deduplicated.map<T>(i => array[i]!)` in deduplicateRelational | Removes all remaining historical CW020-CW023 relations with no new core.ts finding. No helper return annotation or caller type argument is needed. |
| H01 indicesOf | Keep `<T>(array: readonly T[])` | Removes erased unknown-array storage at line 684, with no new finding. |
| H04 cartesianProductWorker | Keep `result: T[][]` | Removes the writable readonly-element view at line 2551, with no new finding. |
| H05 isNodeLikeSystem | Keep `!process.browser`; drop the `typeof require` test | JavaScript-changing: the oracle now takes the Node path. The adapted ambient still declares `browser?: unknown`. |
| H03 elementAt | Drop | Directly removing its two any casts removes the old findings but exposes method-value and callable-ABI refusals at line 1066, and an erased any-element return at line 1329. |
| concatenate | Drop, blocked on caller work | The two truthful mixed overload results are `T[] | undefined` and `readonly T[] | undefined`. They add six checker.ts diagnostics, so cannot meet the zero-diagnostic contract in this core-only change. Public declarations remain identical. |
| or | Scout only | Callback predicate contracts and the returned predicate need lowering work; no edit or owner pin. |

Each site was checked against current main, independently against the baseline,
and then one at a time on the accepted stack. Complete measurement-only core.ts
censuses, trial patches, and comparison summaries are retained. These are
no-output measurements, not successful compilation claims. The baseline and
accepted stack are checker-clean. The rejected concatenate trial is explicitly
checker-rejected. Finding comparison groups by kind, core.ts line and exact reason;
raw streams retain exact columns, units, phases and duplicate instances.

The fresh baseline reproduces the ordinary line-726 wall. The historical CW021
at line 727 is already absent on this main. The other three exact findings, at
726:12, 728:9 and 728:20, disappear with the single map type argument. Explicit
helper call type arguments alone remove nothing. A return annotation plus the map
type argument also works, but is unnecessary. No cast or any is added.

For H03, the new findings are:

- 1066: a method read as a value (at would lose its object, and this with it).
- 1066: a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow.
- 1329: a member typed any ([]) related to T at the return site.

The proposed optional generic feature-probe view does not itself supply receiver
and library-callable proofs. The direct cast-free trial is dropped; no alternate
unchecked view is retained.

The concatenate trial changes only overloads at lines 662 and 666. Node witnesses
in concatenate-witness.cjs show `concatenate(undefined, [])` and its readonly-empty
counterpart return undefined, contradicting the current declared array results.
The truthful declarations add TS18048 at checker.ts:14120, TS2345 at 14121 and
41628, TS2322 at 16980 and 49743, and TS2412 at 41574 under Adamic's options.
The ordinary compiler build also fails. The fresh internal declarations were
independently rebundled: typescript.d.ts is byte-identical, and tsserverlibrary.d.ts
is a fixed reexport. No public declaration diff exists. Fixing the callers needs
separate owner analysis; this adapter does not compensate with casts or defaults.

## or scout

The first overload accepts three guards `(P) => p is Rn` and promises their union
predicate. The variadic overload/implementation accepts a tuple of arguments and
arbitrary callback result U. Its returned closure evaluates callbacks in order,
returns the first truthy result, or returns the last result. With zero callbacks,
that last result is undefined, so the unconstrained U result also has an empty-list
contract question. A nonempty callback tuple, or `U | undefined`, would describe
that case more truthfully, but changing overloads needs caller analysis.

The fresh census reports unproven callback predicates at 2474:66 and 2474:91:
`a type predicate whose return is not proven (there is no body proving this parameter)`.
internal/lower/predicates.go's proveFlowPredicate requires a body unless its narrow
marker-contract proof applies. Function-type callback parameters have no body.
The returned predicate overload has no implementation body of its own either.

A boolean-only adaptation would erase the caller's narrowing contract. A dedicated
three-guard implementation would need runtime/source changes and proof of callback
provenance and the union postcondition. Preserving the existing combinator requires
Lower to carry verified predicate contracts through the higher-order callback list,
variadic invocation, short circuit and returned closure, while accounting for callback
effects on the guarded value. No such admission is claimed here. `or` remains
untouched and unpinned.

## Contract and reproduction

adapt.cjs pins only the four accepted top-level function owners by complete exact
before/after text and SHA-256. Missing, duplicate, nested, changed and mixed-state
owners are refused before any write. Every source is reread before writing. Foreign
selectIndex edits and LF/CRLF survive. Complete core.ts JavaScript must match with
comments kept and removed after restoring only H05's require test for that check. Replay preserves bytes and mtime and writes zero files.
No skipWhile, clear, filterMutate or createQueue owner is pinned or edited.

sort.json records Python's actual directory sort. On this base the predecessor is
99m because parallel 99p is absent; including 99p's real name still sorts 99q after it.
See PROOF.md for validation and CENSUS.md for all 79 next stops.

Set NODE_PATH to stage3/api/node_modules, then run:

```sh
node stage3/adapt/99-correctness-axolotl-q/contract.cjs BEFORE_TREE
node stage3/adapt/99-correctness-axolotl-q/concatenate-witness.cjs BEFORE_TREE
node stage3/adapt/99-correctness-axolotl-q/compare.cjs BEFORE_TREE AFTER_TREE CENSUS_TREE
```

## H05 require-test amendment (#nqp4p7g, #00qn5az)

This amendment changes JavaScript. On the oracle, which runs adapted source as
an ES module, isNodeLikeSystem goes from false to true. On a host with
process.nextTick but no require, tsc now takes the Node path. The function
otherwise keeps its existing text. CLI's landed node:perf_hooks import lets
tryGetPerformance take the perf_hooks path under this guard.

`--generateCpuProfile` is the flag whose newly enabled oracle path reaches
the uncaught bare `require("inspector")` in getNodeSystem's returned
enableCPUProfiler method. The ruled divergence is recorded in
`internal/oracle/testdata/backend_divergences.json`, rows 2 and 3
(`node-without-inspector`, #cn4yapb): ES-module source has no require, while the
modeled host raises ERR_INSPECTOR_NOT_AVAILABLE at this load. This amendment
adds no workaround for it.

The complete lane has one case passing `--generateCpuProfile`:
`unittests:: config:: showConfig Shows tsconfig for single option/generateCpuProfile Correct output for config/showConfig/Shows tsconfig for single option/generateCpuProfile/tsconfig.json`.
It is generated from option declarations and passes
`["--generateCpuProfile", "someString"]`. At upstream v6.0.3,
`src/testRunner/unittests/config/showConfig.ts:27` calls
`let commandLine = ts.parseCommandLine(commandLinesArgs);`, line 34 calls
`const initResult = ts.convertToTSConfig(commandLine, configPath, configParseHost);`,
and line 37 writes the resolved options to the baseline:
`Harness.Baseline.runBaseline(outputFileName, JSON.stringify(initResult, null, 4) + "\n");`.
The callback ends at line 38. It never calls executeCommandLine and therefore
never reaches enableCPUProfiler or its require("inspector"). The whole-lane
measurement records 0 profiler entries, with 9 getNodeSystem control entries.

This is a property of that unit test, not a general early exit for CLI
`--showConfig`. In upstream `src/compiler/executeCommandLine.ts:787-793`,
`if (commandLine.options.generateCpuProfile && system.enableCPUProfiler)` calls
`system.enableCPUProfiler(...)` before executeCommandLineWorker. The worker's
showConfig write and successful exit are later, at lines 666-667 or 706-707.
A real `tsc --showConfig --generateCpuProfile <file>` command line does reach
enableCPUProfiler and its bare `require("inspector")`. On the ES-module oracle
that throws `require is not defined`. The covering ruled divergence is
`internal/oracle/testdata/backend_divergences.json`, rows 2 and 3
(`node-without-inspector`, #cn4yapb). This path is reachable.

Adamic's ruling (1) on #nqp4p7g accepts the retained whole-lane measurement as
evidence, following the tracing and adaptation 47 precedents. It is not an
automatic runtime-count pin in the harness. See [require-drop-PROOF.md](require-drop-PROOF.md)
and [require-drop-counts.json](require-drop-counts.json) for the measurement and
remaining validation.
Historical type-only proof files above describe the original 99q run, not this
JavaScript-changing amendment.

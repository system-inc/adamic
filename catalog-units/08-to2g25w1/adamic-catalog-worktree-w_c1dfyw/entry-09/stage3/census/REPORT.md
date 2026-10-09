Built: a reproducible stage-3 source census, raw stage-0 runs, host ledgers, and .a reproducers.
Base commit: ef3d907ecdc4c771b016f7d9c52372def057a340; source: TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8.
Commands/results: 77 stock entries plus whole program, 78 prepared upstream-config entries plus whole program; 6,569 and 29 whole-program checker diagnostics respectively.
Mutants: inflated file rank, omitted index signature, changed source byte, dropped diagnostic; each caught by its dedicated audit assertion.
Not covered: a native tsc, hidden complete lowering/ownership census, or the full uncached integration gate.

# Stage 3 census

## What the runs establish

The untouched pinned checkout has **79 files** under `src/compiler`: **77 TypeScript source files** and two JSON inputs. The 77 sources contain **192,636 lines** under the explicit `text.split("\n")` convention, including a terminal empty line. Every source was run separately through unmodified `load.Load` and, only if accepted, `lower.Lower`; then the 77 roots were checked together. The two JSON files were also individually attempted: stage 0 rejects their extensions. The raw records retain every diagnostic chain, location, input root, first lowering outcome and duration.

**No corpus entry reaches lowering.** The untouched stock run produces 6,569 whole-program diagnostics, including 3,718 TS1484 type-import findings. There are no actual corpus `Refused` or `NotYet` observations to enumerate beyond that gate. It would be false to report an exhaustive lowering census from these runs. Named nested-function and policy blockers below come from a separately labeled source inventory and actual minimal probes.

Stage 0 has no tsconfig input: `internal/load/load.go` hardcodes Adamic options and adds its prelude. Its `Lower` API also accepts exactly one entry, not a set of root files. The census does not change those production APIs. `make_overlay.py` creates a scratch Go overlay which reads **`src/compiler/tsconfig.json` and its inherited `src/tsconfig-base.json`**, retains all upstream project roots for checking, removes the Adamic prelude in favor of upstream Node declarations, and selects each requested file as the one lowering entry. It never disables diagnostics or passes a checker-rejected program to lowering. Multiple owned roots remain a real `Lower` cardinality error once checking succeeds.

The prepared cache is a separate copy outside this repository. Upstream `scripts/processDiagnosticMessages.mjs` generates `diagnosticInformationMap.generated.ts`; that adds one source root, making **78**. Node declarations are `@types/node@25.3.3`, exactly the source lockfile version; `@types/source-map-support@0.5.10` is also installed. The npm lock is saved here. The original 77-file cache is untouched. All generated-file sites are excluded from the main rankings and retained in raw inventory.

**Stock npm TypeScript 6.0.3 accepts this prepared upstream project with zero pre-emit diagnostics.** Stage 0 under the same tsconfig reports **29**: TS2345 15, TS18048 6, TS2322 4, TS2532 4. Adamic deliberately substitutes sounder regex declarations: optional captures and regex split results may contain undefined. The full locations and message chains are in `data/diagnostics.json`, profile `upstream-config`. Those 29 are a separate first work queue; source rewrites need their actual capture/density contracts. The Go checker and bundled standard library also differ from stock npm tsc; this run does not prove their interchangeability.

A separate **stock-library counterfactual** removes only the regex declaration adapter, keeping the same upstream config and project roots. The single `_namespaces/ts.ts` entry passes checking and reaches `Refused: an import cycle` at `src/compiler/core.ts:1:1`. This confirms the adapter accounts for the 29 findings in this comparison. The raw two observations are in `data/stock-library-entry.jsonl.gz`. This is a diagnostic experiment, not the true-types gate or permission to compile with unsound regex types.

Elapsed summed gate times: stock **283.303s**, upstream config **226.234s**. These are measurements on this worker, not compile-time predictions for the eventual native compiler.

## Count definitions and rankings

**Files** is the number of original source files containing a diagnostic/site, not the number of independently fixable edits. **Lines** is distinct `(file, diagnostic-or-site start line)`. **Sites** is individual occurrences, so two assertions on one line count twice. **Affected LOC** is the sum of entire line counts of those files, a scope estimate only. JSON gives all three rankings: `rank_by_files`, `rank_by_lines`, and `rank_by_affected_file_lines`, with every file and start line. Dependent entries repeat the same diagnostics; these repeats are preserved in raw runs but not inflated into source rankings.

The combined table is a triage ordering across two explicitly labeled layers. `TS...` rows are observed checker-code buckets. Other rows are source forms matching a refusal/NotYet branch or a typed candidate, validated by the linked minimal program. They overlap; do not sum their columns. In particular, TS1294 overlaps enums/namespaces; TS2345 and undefined buckets can share one root cause. A code bucket is not one semantic feature. The action ledger retains review where a bucket mixes contracts.

Categories: **a** missing implementation; **b** deliberate soundness/subset restriction with a source correction; **c** unresolved contract or language design. TS2724 and TS2591 have category **a***: external input/host prerequisites, not proof of a missing language feature. This exception is explicit because build-input failures do not fit the three semantic categories. Source corrections in b are recommendations, not proven automatic adaptations at every site.

### Top 30 by files containing the reason

| Rank | Reason / layer | Files | Lines | Sites | Affected LOC | Category | Minimal program |
| ---: | --- | ---: | ---: | ---: | ---: | :---: | --- |
| 1 | `TS1484` (gate) | 71 | 3718 | 3718 | 191945 | b | [repro/r03/main.a](repro/r03/main.a) |
| 2 | `non-boolean control condition` (source) | 62 | 6635 | 6697 | 179646 | b | [repro/r04/main.a](repro/r04/main.a) |
| 3 | `a function inside a function (a closure)` (source) | 57 | 5574 | 5574 | 174402 | a | [repro/r01/main.a](repro/r01/main.a) |
| 4 | `the non-null assertion !` (source) | 49 | 1034 | 1123 | 173560 | b | [repro/r05/main.a](repro/r05/main.a) |
| 5 | `TS2345` (gate) | 44 | 693 | 735 | 159656 | c | [repro/r06/main.a](repro/r06/main.a) |
| 6 | `TS1294` (gate) | 34 | 180 | 180 | 157235 | c | [repro/r07/main.a](repro/r07/main.a) |
| 7 | `enum` (source) | 33 | 164 | 164 | 155477 | c | [repro/r08/main.a](repro/r08/main.a) |
| 8 | `TS7030` (gate) | 32 | 251 | 251 | 149757 | b | [repro/r09/main.a](repro/r09/main.a) |
| 9 | `TS2322` (gate) | 32 | 118 | 120 | 140984 | c | [repro/r10/main.a](repro/r10/main.a) |
| 10 | `a type predicate` (source) | 31 | 636 | 651 | 139059 | c | [repro/r11/main.a](repro/r11/main.a) |
| 11 | `explicit any` (source) | 28 | 198 | 210 | 142330 | b | [repro/r12/main.a](repro/r12/main.a) |
| 12 | `TS2532` (gate) | 28 | 192 | 224 | 140092 | c | [repro/r13/main.a](repro/r13/main.a) |
| 13 | `TS18048` (gate) | 25 | 323 | 380 | 131059 | c | [repro/r14/main.a](repro/r14/main.a) |
| 14 | `TS2412` (gate) | 23 | 615 | 617 | 125600 | b | [repro/r15/main.a](repro/r15/main.a) |
| 15 | `\|\|=` (source) | 22 | 109 | 110 | 115433 | b | [repro/r16/main.a](repro/r16/main.a) |
| 16 | `TS2375` (gate) | 20 | 76 | 76 | 107503 | b | [repro/r17/main.a](repro/r17/main.a) |
| 17 | `TS2724` (gate) | 17 | 17 | 17 | 112333 | a* | [repro/r18/main.a](repro/r18/main.a) |
| 18 | `TS7029` (gate) | 15 | 83 | 83 | 114694 | b | [repro/r19/main.a](repro/r19/main.a) |
| 19 | `TS2379` (gate) | 13 | 31 | 31 | 89887 | b | [repro/r20/main.a](repro/r20/main.a) |
| 20 | `the comma operator` (source) | 7 | 46 | 47 | 85946 | b | [repro/r21/main.a](repro/r21/main.a) |
| 21 | `TS2488` (gate) | 7 | 11 | 11 | 78285 | c | [repro/r22/main.a](repro/r22/main.a) |
| 22 | `a namespace` (source) | 7 | 11 | 11 | 69459 | c | [repro/r23/main.a](repro/r23/main.a) |
| 23 | `an index signature` (source) | 7 | 11 | 11 | 30024 | c | [repro/r24/main.a](repro/r24/main.a) |
| 24 | `TS2339` (gate) | 6 | 19 | 28 | 69945 | c | [repro/r25/main.a](repro/r25/main.a) |
| 25 | `TS2538` (gate) | 6 | 7 | 7 | 20745 | c | [repro/r26/main.a](repro/r26/main.a) |
| 26 | `a spread after the first field` (source) | 5 | 14 | 17 | 20912 | b | [repro/r27/main.a](repro/r27/main.a) |
| 27 | `TS2769` (gate) | 5 | 10 | 10 | 72240 | c | [repro/r28/main.a](repro/r28/main.a) |
| 28 | `a label` (source) | 5 | 10 | 10 | 84405 | b | [repro/r29/main.a](repro/r29/main.a) |
| 29 | `TS18046` (gate) | 5 | 7 | 8 | 12799 | c | [repro/r30/main.a](repro/r30/main.a) |
| 30 | `TS2591` (gate) | 4 | 45 | 54 | 5057 | a* | [repro/r31/main.a](repro/r31/main.a) |

### Ranking by blocked start lines

| Rank | Reason | Lines | Files |
| ---: | --- | ---: | ---: |
| 1 | `non-boolean control condition` | 6635 | 62 |
| 2 | `a function inside a function (a closure)` | 5574 | 57 |
| 3 | `TS1484` | 3718 | 71 |
| 4 | `the non-null assertion !` | 1034 | 49 |
| 5 | `TS2345` | 693 | 44 |
| 6 | `a type predicate` | 636 | 31 |
| 7 | `TS2412` | 615 | 23 |
| 8 | `TS18048` | 323 | 25 |
| 9 | `TS7030` | 251 | 32 |
| 10 | `explicit any` | 198 | 28 |
| 11 | `TS2532` | 192 | 28 |
| 12 | `TS1294` | 180 | 34 |
| 13 | `enum` | 164 | 33 |
| 14 | `TS2322` | 118 | 32 |
| 15 | `\|\|=` | 109 | 22 |
| 16 | `TS7029` | 83 | 15 |
| 17 | `an ExportDeclaration` | 77 | 3 |
| 18 | `TS2375` | 76 | 20 |
| 19 | `the comma operator` | 46 | 7 |
| 20 | `TS2591` | 45 | 4 |
| 21 | `TS2379` | 31 | 13 |
| 22 | `TS2339` | 19 | 6 |
| 23 | `TS2724` | 17 | 17 |
| 24 | `the void operator` | 15 | 4 |
| 25 | `a spread after the first field` | 14 | 5 |
| 26 | `a definite assignment assertion !` | 12 | 4 |
| 27 | `TS2488` | 11 | 7 |
| 28 | `a namespace` | 11 | 7 |
| 29 | `an index signature` | 11 | 7 |
| 30 | `TS2769` | 10 | 5 |

### The 30 minimal programs and decisions

Each input is an actual `.a` file. TS1484 and TS2724 share `repro/support.a`, which exports `interface Item { count: number; }` and `const Diagnostic = 1`. The raw `repros.jsonl.gz` contains unmodified stage-0 results. `policy-repros.jsonl.gz` is clearly counterfactual: only `erasableSyntaxOnly` is turned off in a scratch loader overlay so enum/namespace refusals can be seen. This does not enable that policy in Adamic or claim upstream source has reached those branches.

**1. TS1484 (b)**

Use explicit type imports; runtime import effects must remain.

```typescript
import { Item } from "../support.a"; const value: Item = {count: 1}; console.log(String(value.count));
```

Observed: `checker: TS1484`. Example original location: `src/compiler/binder.ts:115:5`.

**2. non-boolean control condition (b)**

Make the original truthiness test explicit; preserve JS coercion.

```typescript
const value = 1; if (value) { console.log(String(value)); }
```

Observed: `Refused: a number as a condition`. Example original location: `src/compiler/binder.ts:331:9`.

**3. a function inside a function (a closure) (a)**

Implement named nested functions with recursive binding, captures and lifetime proofs.

```typescript
function outer(): number { function inner(): number { return 1; } return inner(); } console.log(String(outer()));
```

Observed: `NotYet: a function inside a function (a closure)`. Example original location: `src/compiler/binder.ts:567:5`.

**4. the non-null assertion ! (b)**

Replace each assertion with proven narrowing or a loud checked unwrap.

```typescript
function read(value: number | undefined): number { return value!; } console.log(String(read(1)));
```

Observed: `Refused: the non-null assertion !`. Example original location: `src/compiler/binder.ts:594:16`.

**5. TS2345 (c)**

Required argument, optional value, array density, and generic contracts need site review.

```typescript
function take(value: number): number { return value; } const values: number[] = []; console.log(String(take(values[0])));
```

Observed: `checker: TS2345`. Example original location: `src/compiler/binder.ts:3636:63`.

**6. TS1294 (c)**

Enum/namespace/parameter-property policy; see the prior complete syntax census.

```typescript
enum Kind { First, Second } console.log(String(Kind.First));
```

Observed: `checker: TS1294`. Example original location: `src/compiler/binder.ts:315:19`.

**7. enum (c)**

Constant flags and ordinary enum runtime objects require an explicit language decision.

```typescript
enum Kind { First, Second } console.log(String(Kind.First));
```

Observed: `checker: TS1294; policy probe Refused: enum`. Example original location: `src/compiler/binder.ts:315:1`.

**8. TS7030 (b)**

Explicit undefined return only for a declared optional-return contract; inferred contracts need review.

```typescript
function read(flag: boolean): number | undefined { if (flag) { return 1; } } console.log(String(read(false)));
```

Observed: `checker: TS7030`. Example original location: `src/compiler/binder.ts:1966:18`.

**9. TS2322 (c)**

Assignment buckets include optional reads, generics and narrowing; no blanket repair.

```typescript
const values: number[] = []; const value: number = values[0]; console.log(String(value));
```

Observed: `checker: TS2322`. Example original location: `src/compiler/binder.ts:1428:13`.

**10. a type predicate (c)**

Verify guards and assert functions, or narrow at use; cannot trust arbitrary predicates.

```typescript
function isNumber(value: unknown): value is number { return typeof value === "number"; } console.log(String(isNumber(1)));
```

Observed: `Refused: a type predicate`. Example original location: `src/compiler/builder.ts:268:87`.

**11. explicit any (b)**

Adamic intentionally forbids unproven any; observed stage-0 spelling can be NotYet.

```typescript
let value: any = 1; console.log(String(value));
```

Observed: `NotYet: a value of type any`. Example original location: `src/compiler/builder.ts:1074:88`.

**12. TS2532 (c)**

Required indexed read needs a density/bounds contract or explicit missing handling.

```typescript
const values: number[] = []; console.log(String(values[0].toString()));
```

Observed: `checker: TS2532`. Example original location: `src/compiler/binder.ts:1746:21`.

**13. TS18048 (c)**

Required value must be proven present; preserve legitimate absence.

```typescript
const values: number[] = []; const value = values[0]; console.log(String(value.toString()));
```

Observed: `checker: TS18048`. Example original location: `src/compiler/binder.ts:1761:17`.

**14. TS2412 (b)**

Truthful optional declaration may permit present undefined, preserving the write and key.

```typescript
interface Slot { value?: number; } const slot: Slot = {}; slot.value = undefined;
```

Observed: `checker: TS2412`. Example original location: `src/compiler/binder.ts:1109:17`.

**15. ||= (b)**

Use an explicit if with identical single evaluation and truthiness semantics.

```typescript
let value = false; value ||= true; console.log(String(value));
```

Observed: `Refused: ||=`. Example original location: `src/compiler/binder.ts:1956:36`.

**16. TS2375 (b)**

Represent intentionally materialized undefined in the owned optional declaration.

```typescript
interface Slot { value?: number; } const slot: Slot = {value: undefined}; console.log(String(slot.value));
```

Observed: `checker: TS2375`. Example original location: `src/compiler/builder.ts:1535:13`.

**17. TS2724 (a)**

Fetch/generate upstream inputs first; this corpus has missing Diagnostics exports, not a proven language gap.

```typescript
import { Diagnostics } from "../support.a"; console.log(String(Diagnostics));
```

Observed: `checker: TS2724`. Example original location: `src/compiler/binder.ts:51:5`.

**18. TS7029 (b)**

Make intentional switch fallthrough explicit without changing control effects.

```typescript
function read(value: number): void { switch (value) { case 1: console.log(String(1)); case 2: console.log(String(2)); break; default: break; } } read(1);
```

Observed: `checker: TS7029`. Example original location: `src/compiler/binder.ts:1229:13`.

**19. TS2379 (b)**

Align optional representation across owned parameter views; generic mismatch cases need review.

```typescript
function take(value: {count?: number}): void { console.log(String(value.count)); } take({count: undefined});
```

Observed: `checker: TS2379`. Example original location: `src/compiler/builder.ts:1597:36`.

**20. the comma operator (b)**

Sequence statements while retaining value and evaluation order.

```typescript
function one(): number { console.log(String("one")); return 1; } const value = (one(), 2); console.log(String(value));
```

Observed: `Refused: the comma operator`. Example original location: `src/compiler/checker.ts:15956:150`.

**21. TS2488 (c)**

Iterable presence, element contracts, and custom iterator representation need review.

```typescript
const entries = new Map<string, number>(); const [key, value] = entries.entries().next().value;
```

Observed: `checker: TS2488`. Example original location: `src/compiler/builder.ts:1258:65`.

**22. a namespace (c)**

Live namespace objects, merging, exports and initialization need a language decision.

```typescript
namespace Values { export const count = 1; } console.log(String(Values.count));
```

Observed: `checker: TS1294; policy probe Refused: a namespace`. Example original location: `src/compiler/builderState.ts:100:1`.

**23. an index signature (c)**

Choose exact object-key semantics or a reviewed Map source migration.

```typescript
interface Values { [key: string]: number; } const values: Values = {}; console.log(String(values));
```

Observed: `Refused: an index signature`. Example original location: `src/compiler/commandLineParser.ts:1911:5`.

**24. TS2339 (c)**

Union/property access needs a real discriminant or generic contract proof.

```typescript
function read(value: {kind: "a"; count: number} | {kind: "b"}): number { return value.count; }
```

Observed: `checker: TS2339`. Example original location: `src/compiler/checker.ts:29927:52`.

**25. TS2538 (c)**

An optional lookup key must be proved present at its original read.

```typescript
const values: {[key: string]: number} = {}; const keys: string[] = []; console.log(String(values[keys[0]]));
```

Observed: `checker: TS2538`. Example original location: `src/compiler/commandLineParser.ts:2901:50`.

**26. a spread after the first field (b)**

Only a single leading spread is sound; review hidden keys before rewriting.

```typescript
const source = {count: 1}; const value = {label: "x", ...source}; console.log(String(value.count));
```

Observed: `Refused: a spread after the first field`. Example original location: `src/compiler/commandLineParser.ts:2679:9`.

**27. TS2769 (c)**

Overload selection is a bucket; iterator, element and callable contracts differ by site.

```typescript
const values = new Map<string, number>([[1, 2]]); console.log(String(values.size));
```

Observed: `checker: TS2769`. Example original location: `src/compiler/builder.ts:2377:13`.

**28. a label (b)**

Replace labeled control flow with a function preserving exits and finally effects.

```typescript
outer: for (let index = 0; index < 1; index++) { break outer; }
```

Observed: `Refused: a label`. Example original location: `src/compiler/checker.ts:24235:17`.

**29. TS18046 (c)**

Narrow the actual thrown value; upstream catch-any behavior is not a proof.

```typescript
try { throw new Error("x"); } catch (error) { console.log(String(error.message)); }
```

Observed: `checker: TS18046`. Example original location: `src/compiler/commandLineParser.ts:2301:91`.

**30. TS2591 (a)**

Host declarations and native Node host APIs are separate prerequisites.

```typescript
console.log(String(process.cwd()));
```

Observed: `checker: TS2591`. Example original location: `src/compiler/core.ts:2591:19`.

### Additional reasons in the top 30 by lines

**an ExportDeclaration (b)**, [repro/r36/main.a](repro/r36/main.a). Prefer declaration-site named exports; barrels also participate in module cycles.

```typescript
const value = 1; export {value}; console.log(String(value));
```

Observed: `Refused: an ExportDeclaration`.

**the void operator (b)**, [repro/r32/main.a](repro/r32/main.a). Retain evaluation and undefined result; special return adaptation is tooling-only.

```typescript
console.log(String(void 0));
```

Observed: `Refused: the void operator`.

**a definite assignment assertion ! (b)**, [repro/r33/main.a](repro/r33/main.a). Initialize the declared slot or represent absence honestly.

```typescript
class Value { count!: number; } console.log(String(new Value().count));
```

Observed: `Refused: a definite assignment assertion !`.

## String records and post-creation properties

Complete original-source counts: **11 index signatures in 7 files**, **5 explicit `Record<string, T>` references in 4 files**, **0 mapped-type nodes whose constraint resolves to unrestricted string**, and **68 element-access sites whose receiver has a string index type**. The latter captures aliases such as `MapLike<T>` and string-keyed interfaces; it is not 68 independent dictionary declarations. `data/string_lookups.json` retains receiver, key and value types with all locations. Numeric mapped keys, finite string unions, and arrays are separate forms; the unrestricted-string test does not sweep them in.

**210 explicit any tokens in 28 original files**. This counts annotations/type arguments, not inferred-any values or transitive contamination. The minimal any program actually returns `NotYet: a value of type any`, although doctrine refuses any on purpose. Do not read that diagnostic spelling as permission to implement unsound any.

Post-creation ledger: **16 writes** to properties absent from a directly initialized literal: **13 object writes and 3 array metadata writes**. All 16 names exist in their declared receiver types. These are source-shape candidates, not 16 proven illegal expandos: a declared optional slot may be represented by a fixed shape. The array example in `sys.ts` assigns `pollingInterval`, `pollIndex`, and `pollScheduled` after constructing `[]` through assertions. There are **0 direct function-declaration expando writes** in this narrowly defined scan. The ledger does not follow factory calls, arbitrary alias chains, prototype mutation, `Object.assign`, or values of any; it is a lower bound, not an exhaustive dynamic expando census. Original source has 4,101 assertion sites, which are recorded without falsely classifying all supported `as const`, upcasts, or checked downcasts as blockers.

A Map migration must preserve own/inherited key presence, absent versus present undefined, integer-key-first order, overwrite/delete order, object spread, JSON behavior, key coercion and prototype names. `Object.create(null)` tables and `{}` tables differ. The dictionary owner worker must resolve these before assigning migrations to source workers. No dictionary or expando source edit was made here.

## Module and whole-program prerequisites

The resolved module-symbol graph has one **76-file strongly connected component** among the 77 original sources. Its full members and every import/re-export edge are recorded. Stage 0 traverses type-only and re-export edges too; upstream barrel modules therefore conflict with its import-cycle refusal. This structural result is not an observed corpus lowering error, because the checker gate prevents that call.

The independent two-file cycle probe does typecheck and is actually `Refused: an import cycle`. The two independent-root probes each lower alone, while the two-root program returns `lower: stage 0 compiles a program from one entry file, got 2`. Both raw ledgers are committed. This establishes a concrete API limitation rather than assuming all root files can be handed to native emission. An upstream tsc entry exists at `src/tsc/tsc.ts` outside the requested compiler directory; its launch/deprecation/source-map/blocking behavior needs a later entry integration unit.

## Host surface

`System` declares **44 members**, all retained with optionality and source locations in `data/system_contract.json`. There are **196 accesses through symbols declared on System**, comprising **33 distinct members**. These include aliases such as `system` and `host`, not only a textual `ts.sys` search. Static calls, property reads, callback forwarding and feature tests are different units. Optional capability declarations with no direct call site still belong to the contract.

| System member | Access sites | Direct call sites | Files |
| --- | ---: | ---: | ---: |
| `clearScreen` | 2 | 1 | 1 |
| `clearTimeout` | 1 | 0 | 1 |
| `cpuProfilingEnabled` | 1 | 1 | 1 |
| `createDirectory` | 4 | 3 | 3 |
| `createHash` | 3 | 0 | 3 |
| `debugMode` | 2 | 0 | 2 |
| `deleteFile` | 3 | 1 | 2 |
| `directoryExists` | 6 | 5 | 4 |
| `enableCPUProfiler` | 4 | 2 | 1 |
| `exit` | 22 | 22 | 2 |
| `fileExists` | 6 | 6 | 3 |
| `getCurrentDirectory` | 8 | 8 | 3 |
| `getDirectories` | 2 | 2 | 2 |
| `getEnvironmentVariable` | 16 | 12 | 4 |
| `getExecutingFilePath` | 2 | 2 | 2 |
| `getMemoryUsage` | 2 | 1 | 1 |
| `getModifiedTime` | 3 | 1 | 2 |
| `getWidthOfTerminal` | 2 | 2 | 1 |
| `newLine` | 50 | 0 | 3 |
| `now` | 4 | 1 | 2 |
| `preferNonRecursiveWatch` | 1 | 0 | 1 |
| `readDirectory` | 2 | 2 | 2 |
| `readFile` | 4 | 4 | 4 |
| `realpath` | 3 | 1 | 2 |
| `setModifiedTime` | 3 | 1 | 2 |
| `setTimeout` | 1 | 0 | 1 |
| `storeSignatureInfo` | 2 | 0 | 2 |
| `useCaseSensitiveFileNames` | 5 | 0 | 2 |
| `watchDirectory` | 2 | 0 | 2 |
| `watchFile` | 2 | 0 | 2 |
| `write` | 21 | 21 | 4 |
| `writeFile` | 5 | 3 | 4 |
| `writeOutputIsTTY` | 2 | 1 | 1 |

**Node module/global chains:** the lexical host ledger contains 19 `_fs` chains, 6 `_path`, 30 process, 12 native Performance-interface accesses, 9 require calls, 2 Buffer constructor chains, 2 `_os`, 2 global, 1 crypto constructor, 1 aliased realpath call, 1 each timer/clear timer, and 1 filename reference. Maximal chains are counted once. This ledger is supplemented by **126 Node-declaration-resolved accesses** in `data/node_derived_sites.json`; it catches tracing's `fs` alias and methods of Buffer, Stats, Dirent, Hash and inspector Session. The two ledgers overlap and must not be added.

| Node declaration module | Access sites | Direct call sites |
| --- | ---: | ---: |
| `@types/node/buffer.buffer.d.ts` | 2 | 2 |
| `@types/node/buffer.d.ts` | 6 | 6 |
| `@types/node/console.d.ts` | 1 | 1 |
| `@types/node/crypto.d.ts` | 3 | 3 |
| `@types/node/fs.d.ts` | 52 | 38 |
| `@types/node/globals.d.ts` | 2 | 1 |
| `@types/node/inspector.d.ts` | 4 | 2 |
| `@types/node/inspector.generated.d.ts` | 12 | 3 |
| `@types/node/net.d.ts` | 2 | 2 |
| `@types/node/path.d.ts` | 6 | 6 |
| `@types/node/perf_hooks.d.ts` | 6 | 0 |
| `@types/node/process.d.ts` | 28 | 4 |
| `@types/node/tty.d.ts` | 2 | 0 |

Filesystem semantics needed include synchronous open/read/write/close, stat with missing entries, directory entries and symlink traversal, realpath/native realpath fallback, mkdir races, mtime updates, unlink failures, and watch/unwatch. Path uses resolve, dirname and join. Process includes cwd, argv/execArgv, env, platform, pid, stdout columns/TTY/write/private blocking handle, exit, memory usage and nextTick. Performance includes timeOrigin/now and mark/measure/clear operations. Additional surfaces include crypto SHA256 Hash update/digest, Buffer encodings and BOM handling, timers, inspector profiling, and optional source-map-support. Static accesses include optional profiling/watch paths; none is a frequency estimate for one-shot type checking.

**Executed host counts:** stock tsc checked two real strict ES2020 noEmit programs with a fresh compiler host per program. Good produced `[]`; bad produced TS2322 at `bad.ts:1:7` and TS2345 at `bad.ts:1:106`, with their exact stock messages asserted. Each invoked `ts.sys.readFile` **53**, `fileExists` **2**, `directoryExists` **2**, `getExecutingFilePath` **2**, `getCurrentDirectory` **1**; Node fs read/open/readSync/close each **53**, stat **4**. The first used process.cwd once, the second zero due to process-level directory caching. Counts include library reads. No emit, watch, config parsing, package graph, CPU profiling, Windows path cases or long-running host workload was exercised. Full diagnostic records and counts are in `data/host_runtime.json`.

## Assignment plan for 30 workers

This is an ordered, reviewable queue, not a claim that the port has only 30 remaining features. First establish the upstream gate and module/entry policy. Keep source adaptations separate from compiler implementation and semantic design. Do not weaken checker options to make the tables green. Each implementation worker must add Node/spec-backed cases and a dedicated caught mutant. Partition source work by the location ledgers; two workers must not widen the same owned declaration independently.

| Worker | Deliverable / ledger | Dependency |
| ---: | --- | --- |
| 1 | Upstream project input: generated diagnostics, pinned declarations, tsconfig/entry integration; 29 regex diagnostics | none |
| 2 | Module graph and initialization policy; 76-file SCC, 77 re-export declarations | 1; design approval for cycle semantics |
| 3 | Named nested function lowering; 5,574 sites, capture/recursion lifetime proof | 2 |
| 4 | Enum values and runtime objects; 164 declarations, prior syntax-census ledger | 1; enum policy |
| 5 | Namespace merging/initialization and six parameter properties from prior syntax census | 2,4; policy |
| 6 | Type-predicate/assert-function verification; 651 nodes | 1; predicate proof design |
| 7 | Non-null/definite assignment review; 1,123 assertions and 12 definite markers | 1; site invariants |
| 8 | Dictionary representation decision; 11 signatures, 5 Records, 68 string lookups | 1; key/order/presence design |
| 9 | Dictionary migrations or implementation following worker 8's decision | 8 |
| 10 | Explicit-any owner corrections; 210 tokens and actual NotYet spelling | 1; consumer recheck |
| 11 | Truthiness source work: binder/checker files from source ledger | 1; preserve coercion/evaluation |
| 12 | Truthiness source work: parser/scanner/factory files | 11 format/proof agreement |
| 13 | Truthiness source work: transformers and remaining compiler files | 11 format/proof agreement |
| 14 | Optional declarations: TS2412 617, TS2375 76, TS2379 31, shared owner map | 1; prior adapter rules |
| 15 | Required indexed reads in core/utilities; preserve sparse/callback behavior | 1; checked-density contract |
| 16 | Required indexed reads in checker/binder; TS2345/18048/2532 ledger | 15 contract |
| 17 | Required indexed reads in parser/scanner/factory; capture-group proof where relevant | 15 contract |
| 18 | Required indexed reads in transformers/emitter | 15 contract |
| 19 | Remaining required indexed reads in builder/watch/resolution/module code | 15 contract |
| 20 | Return contracts and switch fallthrough; TS7030 251, TS7029 83 | 1; preserve finally/completion |
| 21 | Logical assignment/comma/labels/spread source adaptation; 110/47/10/17 sites | 1; evaluate each receiver once |
| 22 | TS2322/2339/2538 structural and narrowing contract triage | 1,14; full chains |
| 23 | TS2488/2769/18046 iterable/callable/exception contract triage | 1; corpus-specific probes |
| 24 | Host filesystem read/stat/directory/realpath and BOM/Buffer support | 1; Node host ledger |
| 25 | Host write/create/unlink/mtime and atomic failure semantics | 24 |
| 26 | Host path/process/argv/env/stdout/exit and diagnostics output | 24; platform scope |
| 27 | Host watch/timers and incremental-host semantics | 24,25; one-shot gate first |
| 28 | Host performance, SHA256, optional profiler/source-map capabilities | 26; optional-capability policy |
| 29 | Original-source diagnostic oracle: golden fixtures with exact codes, messages, positions, output and exit | 1; stock tsc is outside oracle |
| 30 | Integration/ownership audit and re-census after each merged layer; latent cycles and NotYet remain unknown | 2-29; no collector fallback |

## Reproduction and evidence

Base Adamic `ef3d907ecdc4c771b016f7d9c52372def057a340`; cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`; typescript-go `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. Source SHA256s, byte/line counts and generated distinction are in `data/files.json`. External caches used here: `/workspace/cache/tsc-census/typescript`, `/workspace/cache/tsc-census/prepared`, `/workspace/cache/tsc-census/npm`. No upstream TypeScript source was added to Adamic.

Fresh-cache preparation, equivalent to the captured clone/install/generation steps (run from the Adamic root):

```sh
mkdir -p /workspace/cache/tsc-census/npm /workspace/cache/tsc-census/prepared
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/cache/tsc-census/typescript > /tmp/tsc-census-clone.log 2>&1
git -C /workspace/cache/tsc-census/typescript rev-parse HEAD
cp stage3/census/data/npm-package.json /workspace/cache/tsc-census/npm/package.json
cp stage3/census/data/npm-lock.json /workspace/cache/tsc-census/npm/package-lock.json
npm ci --prefix /workspace/cache/tsc-census/npm --no-audit --no-fund > /tmp/tsc-census-npm.log 2>&1
cp -a /workspace/cache/tsc-census/typescript/{src,scripts,package.json,package-lock.json} /workspace/cache/tsc-census/prepared/
ln -s /workspace/cache/tsc-census/npm/node_modules /workspace/cache/tsc-census/prepared/node_modules
(cd /workspace/cache/tsc-census/prepared && node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json) > /tmp/tsc-census-generate.log 2>&1
```

```sh
bash cloud/setup.sh > /tmp/tsc-census-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go build -o /workspace/cache/tsc-census/census ./stage3/census/tool > /tmp/tsc-census-build.log 2>&1
/workspace/cache/tsc-census/census /workspace/cache/tsc-census/typescript/src/compiler /workspace/cache/tsc-census/stock.jsonl > /tmp/tsc-census-stock.log 2>&1
python3 stage3/census/make_overlay.py /workspace/cache/tsc-census/overlay
go build -overlay=/workspace/cache/tsc-census/overlay/overlay.json -o /workspace/cache/tsc-census/census-upstream ./stage3/census/tool > /tmp/tsc-census-upstream-build.log 2>&1
CENSUS_CONFIG=/workspace/cache/tsc-census/prepared/src/compiler/tsconfig.json /workspace/cache/tsc-census/census-upstream /workspace/cache/tsc-census/prepared/src/compiler /workspace/cache/tsc-census/upstream.jsonl > /tmp/tsc-census-upstream.log 2>&1
CENSUS_TYPESCRIPT=/workspace/cache/tsc-census/npm/node_modules/typescript/lib/typescript.js node stage3/census/inventory.cjs /workspace/cache/tsc-census/prepared stage3/census/data > /tmp/tsc-census-inventory.log 2>&1
python3 stage3/census/summarize.py > /tmp/tsc-census-summary.log 2>&1
python3 stage3/census/audit.py /workspace/cache/tsc-census --mutants > /tmp/tsc-census-audit.log 2>&1
```

The committed driver now also records non-source files during its ordinary walk and uses only source roots for the final program. This run originally measured the 77 source files before that small tooling addition; its two non-source attempts are a separate raw ledger. `--non-source-only` reproduces that supplement. A rerun with the final driver thus has two extra per-file extension errors; the whole-program population remains identical.

Setup output: Go go1.27.1, clang 20.1.8 with working sanitizer probe, Node v24.19.0. Go 0s; clang/Node/submodules 1s; build cache warm 177s; total 177s; nproc 5; cgroup CPU quota 4. Full setup log is committed.

Test commands, all redirected to logs:

```sh
go test ./stage3/census/tool ./cmd/adamic-meter ./internal/load -count=1 > /tmp/tsc-census-package-tests.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(functions|library_object_order)\.a$' -count=1 -timeout 30m -v > /tmp/tsc-census-oracle.log 2>&1
go vet ./stage3/census/tool > /tmp/tsc-census-vet.log 2>&1
```

Package checks pass: meter 3.482s, load 1.841s; census tool has no Go tests. Vet passes. The oracle filter matches **functions.a, generic_functions.a, library_object_order.a**; all three pass, total 15.327s, with six native and six Node cache misses. An earlier filter matched no tests and was replaced; it is not counted as validation. The full uncached gate was not run for this census-only change. No production compiler, native emitter, loader, oracle fixture or counts table was edited.

Audit passes source hashes, entry/whole coverage, locations, every count/ranking, all top-30 reproducers, the 44-member System contract and exact host diagnostics. Four artifact mutants were run independently: file count +1 caught by rank recount; dropped index-signature site caught by site recount; changed pinned source byte caught by SHA256; removed TS2345 caught by exact diagnostic comparison. They are tooling mutants, not compiler or memory-safety mutants. Their exact catches are in `data/logs/tsc-census-audit.log`.

## Remaining limits

Every returned corpus diagnostic is preserved; all original files were attempted. **Every latent NotYet and context-dependent refusal cannot be observed with the current fail-fast APIs while no original file passes the checker.** Source inventory covers named branches and selected typed forms, not every lowering representation, generic specialization, unbound method, invariant mutable view, prototype read, cyclic runtime object or ownership proof. A function/array/property source edit can expose a new blocker that this run could not reach. Re-census after each prerequisite merges; never treat a count decrease as proof of native correctness or as assurance of the October 10 target.

The three-category recommendations do not prove all b edits are small or mechanical. TS code buckets need owner and full-chain review; the prior strictness survey supplies stronger sampled evidence and existing conservative adapters. Unrestricted records and exact original-source enum/module behavior remain design decisions. The census identifies that boundary rather than replacing evidence with an implementation promise.

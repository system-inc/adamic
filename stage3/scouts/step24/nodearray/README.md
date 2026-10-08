# Step 24: NodeArray layout and presence

This scout inventories TypeScript 6.0.3's NodeArray sites, measures them on Node, and supplies four standalone `.a` fixtures. It implements no compiler feature. The October 8 ruling requires array storage with known extra slots, without a wrapper or a dynamic property dictionary; an optional slot must preserve absent versus present-undefined using step 17's representation.

The complete file:line:column ledger is [LEDGER.md](LEDGER.md), with machine-readable types, resolved declarations and source spans in [census.json](census.json). [status.json](status.json) contains exact Node bytes and current-main diagnostics. [counts.md](counts.md) records fixture and census totals.

## What the declaration says

Upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8`, tag `v6.0.3`, declares `MutableNodeArray` at `src/compiler/types.ts:1584` and `NodeArray` at `:1589`. Each has four **required** metadata fields: `pos: number`, `end: number`, `hasTrailingComma: boolean`, and `transformFlags: TransformFlags`. `pos`/`end` come from `TextRange` or `ReadonlyTextRange`. No base metadata field is declared optional.

Two additions outside the base layout need an explicit decision in the implementation:

- `MissingList<T> extends NodeArray<T>` declares `isMissingList: true` at `src/compiler/parser.ts:3566`. The parser writes it at `:3571` and reads it at `:3576`. Ordinary lists lack this field. A common fixed layout needs a presence slot, or an explicitly distinguished MissingList layout.
- `__tsDebuggerDisplay` is defined through a descriptor at `src/compiler/debug.ts:576`, outside any NodeArray field declaration. On Node 24, debug attachment creates a shared prototype at `:598` and sets each array's prototype at `:601`. The method is inherited, not an own field. The fallback at `:605` attaches it directly when `Object.setPrototypeOf` is unavailable. Neither a silent omission nor an own-field approximation reproduces those presence semantics.

A slice returns an ordinary array: it loses the four own metadata fields. Factory reconstruction initializes them again. `createNodeArray` may copy a small ordinary array, promote a larger array in place, return an existing NodeArray unchanged, or clone it when the trailing-comma option differs. The known layout must preserve those identities and element aliases.

## Census rule and counts

`census.cjs` uses the unmodified TypeScript 6.0.3 compiler API with the inherited upstream compiler tsconfig, not text search. The program has zero stock diagnostics. Its 77 non-generated compiler source files and SHA-256 hashes are recorded. The generated diagnostic TypeScript file is a dependency but contains no NodeArray sites; corpus parsing includes it.

A receiver qualifies when its checker type is NodeArray/MutableNodeArray, a subtype through interface bases, a union/intersection containing one, a constrained type parameter, or a mapped alias of one. Direct property and finite literal-key element accesses resolve their properties to declarations through `getPropertyOfType` and `getRootSymbols`. Inherited standard Array members and numeric element keys are excluded from metadata counts. Assignment/update/delete syntax determines the access mode.

The ledger contains 97 direct metadata accesses: 87 reads and 10 writes. Fields total `pos` 31, `end` 39, `hasTrailingComma` 19, `transformFlags` 6, and `isMissingList` 2. Four additional reflection sites comprise two own-presence queries, a prototype-aware `in` query, and a descriptor write. All 82 non-literal computed accesses have numeric keys.

Creation-related categories are 159 resolved `createNodeArray` calls, 56 range-helper calls with an array target, 20 slices on NodeArray receivers, two factory assertions promoting arrays, one assertion of the debug prototype, and one additional plain-array slice promoted by its enclosing assertion. These are syntactic sites, with nested overlaps, **not allocation counts**. Another 593 NodeArray-returning calls are listed separately because they may forward or reuse existing arrays.

Range helpers deserve a separate ledger: their receivers are typed TextRange rather than NodeArray. Four helper accesses at `factory/utilitiesPublic.ts:11` and `utilities.ts:10645,10655` read/write pos/end. There are 56 array-target calls, containing 46 array-provider arguments; resolved signatures identify aliases such as `setTextRangeWorker`. A provider supplies range values; a target receives writes. All are listed with argument roles. This is type-directed source coverage, not a whole-program heap-alias analysis of values erased to `any` or TextRange elsewhere.

## Runtime observations

`scanner-corpus.json` preserves the 81 paths from `stage3/drivers/scanner/evidence/native3/unsplit/files.json`. Inputs here are the **pristine v6.0.3 files**, not the adapted scanner tree. Each actual byte hash appears in `runtime.json`; its token count must not be substituted for an adapted-tree measurement.

Node v24.19.0 scans 505,295 tokens, including one EOF per file, and observes **zero NodeArrays during scanning**. Scanning alone cannot answer the presence question. The explicitly additional runs parse the same inputs, insert a trivia prefix and incrementally update them, and parse with debug information enabled. Each of these three passes visits 137,742 final-tree arrays and 941,215 nodes, with zero parse diagnostics.

| Field | Parse and incremental final trees | Debug final trees | Present-undefined observed anywhere? |
| --- | --- | --- | --- |
| pos | Own value on all arrays | Own value on all arrays | No |
| end | Own value on all arrays | Own value on all arrays | No |
| hasTrailingComma | Own value on all arrays | Own value on all arrays | No |
| transformFlags | Own value on all arrays | Own value on all arrays | No |
| isMissingList | Absent on all final arrays | Absent on all final arrays | No |
| __tsDebuggerDisplay | Absent on all final arrays | Inherited method on all arrays; own property absent | No |

Intermediate observations matter. All four base fields are absent immediately after a plain array is asserted to NodeArray and before initialization. Slice results also lack them. Those construction windows do not mean completed NodeArrays may omit the fields. There are 4,014 `isMissingList = true` writes across the runs, including speculative parser work discarded from the final tree; an example is recorded at `parser.ts:3571`. Factory-return observations show the four base fields always present with values. No unexpected own extra field was discovered.

`instrument.cjs` rewrites only a new scratch copy, matching exact AST spans from the checker ledger. It wraps 87 direct reads, 10 direct writes plus two generic range setter writes, and 832 array-return/promotion sites: 931 operations in 77 files. Observer helpers keep object identity and store IDs in a WeakMap; they add no array properties. `runtime.json` records 12,511,022 observations of 441,988 distinct array objects. Counts in phases may observe an object repeatedly and must not be added as distinct allocations. Reflective/debug descriptor operations are observed at subsequent array snapshots rather than rewritten themselves.

The pristine control and instrumented compiler agree on corpus bytes, lexical counts, all three traversal results, and the digest of node kind/range plus array length, field values and presence:

`a81e1c5e18b38f8ec17ae1b4ea14d91bbbd0a85337f442f40f4afbfc1875624c`

This is bounded evidence, not proof that no public API or other program can supply present-undefined. The factory's `transformFlags === undefined` recovery branch explicitly anticipates such input. An extra attempt to use upstream's optional aggressive incremental checks failed with its child-position assertion on an enum; that control failure is preserved in `logs/runtime-aggressive-failure.txt`. The normal incremental pass above uses the API's default checks and succeeds.

## Fixtures and mutants

Function statements are extracted unchanged with stock AST spans, normalizing only CRLF to LF; `manifest.json` keeps their full original text, and `verify.py` checks every span. Supporting declarations reduce the Node shape and represent transformFlags as required `number | undefined`, an explicit widening of upstream's required enum slot for the recovery probe. `propagateChildFlags` is reduced to return the chosen child bits, which pass through for these Identifier-shaped inputs; this is not a test of all subtree exclusion rules. Debug attachment is disabled in standalone fixtures; debug inheritance is measured in the full source run. The standalone own-property helper uses `Object.hasOwn`.

| Fixture | Real use | Source mutant | What catches it |
| --- | --- | --- | --- |
| 01_factory_clone.a | Small copy, large promotion, empty layout, reuse, comma-change clone and range preservation | Clone end becomes -1 | Node stdout differs |
| 02_parser_range.a | Parser range construction, explicit/implicit end, slice metadata loss and reconstruction | Missing end defaults to 0 | Node stdout differs |
| 03_repair_flags.a | Existing array's required flags slot is present-undefined, then repaired on reuse | Omit aggregation | Node stdout differs |
| 04_missing_presence.a | Ordinary list absent marker, real MissingList true marker, and explicitly synthetic present-undefined marker probe | MissingList marker becomes undefined | Node stdout differs |

All four Node goldens pass with empty stderr and exit 0. Each mutant also exits 0 with empty stderr, but changes stdout: none is counted as caught merely because it fails to run. The present-undefined probes are deliberate driver inputs; **they were not found on the pristine corpus**.

All four builds on main `45487a809f89885a3fc651cd590e7dabf31362dc` are **Refused**, first at upstream `isNodeArray`'s own-property type predicate (`adamic/no-type-predicate`). Exact diagnostics are in status.json and logs. There is no native executable and no native byte comparison to claim. The verification script will run a binary and reject any byte difference if compilation becomes supported. The fixture mutants prove the Node golden checks can fail; they do not prove the Adamic predicate diagnostic can flip.

A fifth, measurement-specific mutant removes `!!` from the actual factory initializer at `nodeFactory.ts:1202` in the instrumented scratch source. Node still completes all three corpus passes with unchanged traversal counts, but now 455,903 factory-return observations have an own present-undefined `hasTrailingComma`. Its digest changes to `39e03859b37fb83e3bbe9ea9d16f62fec6dee8375e966fb6cb5bc24a21bfdc56`. This proves the presence observer and semantic control detect a real source error. `presence-mutant.json` and `runtime-mutant.json` preserve the result; `mutate-presence.cjs` selects the initializer with the stock AST.

## Reproduce

Use a new scratch directory, a pristine `v6.0.3` checkout, and TypeScript 6.0.3 plus `@types/node@25.3.3` and `@types/source-map-support@0.5.10` installed in a separate scratch dependency directory. Point the upstream checkout's node_modules at those dependencies, then run `node scripts/processDiagnosticMessages.mjs` in that checkout. Do not modify the repository's TypeScript submodule. Here `unit` denotes this directory, `upstream` the pristine checkout, `api` the scratch dependency directory, and all other paths are new scratch outputs.

```bash
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH="$api/node_modules"
export NODEARRAY_TYPESCRIPT="$api/node_modules/typescript/lib/typescript.js"
node "$unit/census.cjs" "$upstream" > census.log 2>&1
node "$unit/extract.cjs" "$upstream" > extract.log 2>&1
node "$unit/instrument.cjs" "$upstream" "$instrumented" "$unit/census.json" > instrument.log 2>&1
node "$unit/runtime.mjs" "$upstream" "$upstream" "$unit/scanner-corpus.json" "$baseline" > baseline.log 2>&1
node "$unit/runtime.mjs" "$instrumented" "$upstream" "$unit/scanner-corpus.json" "$observed" > observed.log 2>&1
node "$unit/mutate-presence.cjs" "$instrumented" "$mutant" > mutate.log 2>&1
node "$unit/runtime.mjs" "$mutant" "$upstream" "$unit/scanner-corpus.json" "$mutant_result" > mutant-runtime.log 2>&1
python3 "$unit/verify.py" "$upstream" "$fixture_scratch" --runtime-baseline "$baseline/runtime.json" --runtime-instrumented "$observed/runtime.json" --runtime-mutant "$mutant_result/runtime.json" > fixtures.log 2>&1
python3 "$unit/report.py" > report.log 2>&1
```

Initial recording used `verify.py --record`, then verification was repeated against saved records. Each fixture uses exactly `node --disable-warning=ExperimentalWarning oracle/node.mjs <file>` and `go run ./cmd/adamic build <file> -o <scratch-binary>`. No whole package or full gate was run for confirmation. These scout fixtures are not registered in the shared internal/oracle fixture table; its native heap-event counts are unavailable for refused programs, so the refreshed inventory is this directory's counts.md.

## Toolchain and limits

The first setup process was interrupted by an environment restart; `logs/setup-interrupted.txt` preserves it. A resumed setup succeeded: Go ready 0.090s, Node ready 0.096s, markdown ready 0.277s, submodules ready 0.346s, clang ready 0.903s, go build ready 209.460s, deferred test binaries 210.084s, warm build cache 210.106s, done 210.543s. `nproc` is 5; the cgroup quota is 4 CPUs (`400000 100000`). Versions are Go 1.27.1, clang 20.1.8, Node 24.19.0.

Not covered: backend layout construction or lowering; native output of refused fixtures; checker/emitter/transformer runtime execution; non-corpus input programs; ES5 debug fallback execution; all public factory callers; a proof over values whose NodeArray type has been erased. There is no evidence here of corpus present-undefined, and the synthetic probes must not be reported as such.

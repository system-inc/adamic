# getFunctionName caller measurement

Task b5w3ycg. Adamic roads ruled Nov 7 19:48Z. Measurement only: no move, deletion, reflection change, new adaptation, compiler build or harness build.

## Result and inference

**My inference: road 3, a reflection ruling for census unit `Debug.getFunctionName`, with remaining compiler caller unit `cast` named.** Stock find-references finds a real remaining call in `src/compiler/core.ts`, in `cast`, plus one in compatibility `deprecate`. The declaration itself is not counted as a caller. Runtime execution counts do not remove either source caller. In particular, zero observed executions of the compiler call do not establish that no compiler source caller exists. The premise for road 1 (only harness/tests callers) or road 2 (no caller anywhere) is therefore not met. The public API baseline does not contain the helper.

The existing getFunctionName census stop at debug.ts:355:26 is the premise supplied for this task, not a new census measurement. getFunctionName itself and all TypeScript source callers stay untouched by this task.

## Tree and tool provenance

- Report base: fetched origin/main `3214dd465e1155d7bf4d41653e95b29a3f1060a6`.
- Accepted road (b) cut: `typescript/core-debug-messages-1`, `2a95448b8fadbcd30549a148bd28da5f8b75f907` (chg_gzdnphpq).
- Main does not yet contain road (b). Its existing adapted source product was restored with stage3/apply.sh, with ADAMIC_TYPESCRIPT_SOURCE explicitly unset so a cache miss could not silently build a replacement. Product key: `d2af18f8b5611dd97bc936f672250b934cd97593a02dfc7f1fcfe1932066dbcf`.
- The already accepted adapter was applied only in scratch, reporting four edits. Its adapter and pins Git blob IDs match the accepted cut through the connected GitHub file API: `6953c4e927326bb13673ea717e7a19664bd7575d` and `f68fe1717dc92fd1a392d9f7683733928c582bd9`. The optional Git fetch of that cut failed authentication; no unverified replacement was used.
- Stock checker: installed typescript@6.0.3, asserted by the script. TypeScript source is the pinned 6.0.3 checkout at `050880ce59e30b356b686bd3144efe24f875ebc8`. Node v24.19.0, linux/x64.

To avoid building anything, the runtime measurement reuses the existing normal road (b) compiler and harness at `/workspace/scratch/debug-messages/after`. Its 81,503 checked source/build-input/test-data files match the freshly prepared main-plus-road-(b) tree byte for byte. Build output, package-install directories, Git metadata, local test output and patch-table metadata are excluded from that input comparison. All 10 normal JavaScript files and both public declarations also match the retained build archive from `385ccad0bf6cfd800443566d947cc638a7e776aa`. The exact source maps used for attribution are hashed; those normal maps omit sourcesContent, so no embedded-source check is claimed. Their named source inputs are in the verified input set. `reuse-validation-summary.json` retains the input-set digest and source hashes.

## Every stock-checker reference

The script uses stock `languageService.findReferences`, queried at the declaration name of Debug.getFunctionName. It roots the program in every supported source file under checkout `src`, not just the compiler project. There are 711 roots and 826 loaded program files, including 79 compiler roots, 38 harness roots and 280 testRunner roots. The latter include 257 unittest roots. Find-references returns one definition and two read references, both direct calls. No source alias/value reference, harness reference, or unittest reference is returned. This is a symbol query, not a text-search count.

| File:line:column | Enclosing function | Reference kind | Access spelling |
| --- | --- | --- | --- |
| src/compiler/core.ts:1791:97 | cast | direct call | `Debug.getFunctionName` |
| src/compiler/debug.ts:347:21 | Debug.getFunctionName | definition | `getFunctionName` |
| src/deprecatedCompat/deprecate.ts:71:66 | deprecate | direct call | `Debug.getFunctionName` |

Source lines, exactly as read (line endings omitted for display):

`src/compiler/core.ts:1791`

```typescript
    return Debug.fail(`Invalid cast. The supplied value ${value} did not pass the test '${Debug.getFunctionName(test)}'.`);
```

`src/compiler/debug.ts:347`

```typescript
    export function getFunctionName(func: AnyFunction): string {
```

`src/deprecatedCompat/deprecate.ts:71`

```typescript
    const deprecation = createDeprecation(options?.name ?? Debug.getFunctionName(func), options);
```

Bare `getFunctionName` occurs as the declaration name. Both callers spell the access `Debug.getFunctionName`. There is no `ts.Debug.getFunctionName` reference in the requested source inventory. The full root-file list, definition group and exact spans are retained in references.json.

### Text grep cross-check, not the primary result

An exact-word `rg -n -w getFunctionName` over the entire checkout, including scripts and test fixtures outside src but excluding build output, installed packages and .git, finds exactly the same three lines. This agrees with, and does not replace, stock find-references. The raw cross-check is retained. Plural getFunctionNames is a different name.

## Runtime calls over the whole lane

All 398 frozen lane units run in pieces through the ordinary stage3/lane/shard.py path used to split stage3/lane/run.sh, with the ordinary identity, API, baseline and count guards. The accepted inventory is reconciled exactly once from every raw archive and guard report. There is no new compiler/harness build or test expectation change.

| Source call site | Enclosing function | Actual helper entries | Tagged call attempts |
| --- | --- | ---: | ---: |
| src/compiler/core.ts:1791:97 | cast | 0 | 0 |
| src/deprecatedCompat/deprecate.ts:71:66 | deprecate | 6 | 6 |

Units with helper entries: stage3-lane-file-293 (6).

Total helper entries: **6**. Untagged entries: **0**. There are 797 retained Node process/thread records. Oracle totals: 106366 passes, 1 expected failure, 0 skips. The expected public-API acknowledgement failure is preserved. All 20 accepted pieces pass, taking 1472.865 seconds summed across sequential pieces. The largest piece took 100.718 seconds.

### Observer and controls

The observer edits only private copies of existing emitted JavaScript. Every JavaScript entrypoint is an ordinary private copy, including unchanged launchers and proxies. The closed-module-graph control verifies local module resolutions remain within the private product and proves a proxy call reaches the same observed helper. A source-map lookup ties every compiled getFunctionName property-call expression to a stock-checker reference: core cast has three emitted copies (_tsc.js, typescript.js, run.js), and compatibility deprecate has two (typescript.js, run.js). All three emitted helper definitions have an entry observer. Normal source and normal JavaScript remain unchanged.

The call observer evaluates the original call in a tagged thunk, preserving its receiver, arguments and conditional/lazy location. The entry observer counts actual entry into the original function body. Tagged attempts and actual entries match for every process in every accepted unit. A direct or aliased invocation outside the tagged source expressions is retained as an untagged entry with its bundle and stack. The preload shares counters with VM contexts and records workers/threads independently. Explicit zeros are written on process exit.

Controls force the compiler cast failure and the actual compatibility deprecate call independently; each corresponding entry counter increases by one. A control-only copy exposes the existing compiled deprecate declaration to the control driver; this exposure is absent from the oracle product. A separate direct alias call deliberately bypasses source tags and increases the untagged-entry detector by one. All controls pass. Control records are separate from the whole-lane counts. `positive-controls.json` retains before/after snapshots and the Error message.

The initial partial run used symlinks for unchanged launchers. A coverage audit found that a proxy could resolve an uninstrumented library outside the private product. That run was stopped and discarded. The accepted measurement restarts all 398 units with private copies of every JavaScript file, after both the closed-graph and per-site controls pass. Discarded-run diagnostics are retained separately and contribute no counts here. The runtime result is a count over this lane, not proof that these conditional source calls can never run. No consumers outside the checkout are measured.

## Declarations and public-baseline gate

All real checkout files with the three requested declaration basenames were searched, plus the three actual emitted declarations. Installed-package copies and build/local test scratch are excluded. Exact matches follow; no matching line is invented for an absence.

| Declaration file | getFunctionName |
| --- | --- |
| checkout/lib/tsserverlibrary.d.ts | absent |
| checkout/lib/typescript.d.ts | absent |
| checkout/tests/baselines/reference/api/typescript.d.ts | absent |
| built/local/tsserverlibrary.d.ts | absent |
| built/local/typescript/typescript.d.ts | absent |
| built/local/typescript.d.ts | absent |
| built/local/typescript.internal.d.ts | line 7878 |

`built/local/typescript.internal.d.ts:7878`

```typescript
        function getFunctionName(func: AnyFunction): string;
```

`getFunctionName` is **absent from** the explicitly requested `tests/baselines/reference/api/typescript.d.ts` baseline. The helper is present only in the emitted internal declaration among these files. Exact file hashes, byte counts and matching lines are in declarations.json.

## Adaptation pins

| Adaptation | Status | Pin scope |
| --- | --- | --- |
| 30-indexed-reads | active | match[1] element read |
| 41-explicit-any-remaining | active | exact return expression in getFunctionName |
| 45-regex-captures | active | named function, match initializer and capture 1 |
| 99-correctness-core-debug-messages (accepted road b overlay) | active in requested main-plus-road-b tree | whole unchanged getFunctionName owner |
| 99-correctness-axolotl-n | inactive proposal | proposed whole owner and func parameter |
| 99-correctness-core-multimap | active, indirect import pins | whole import statements containing cast |

- `stage3/adapt/30-indexed-reads/sites.json:3297`: debug.ts match[1], occurrence 1 of total 1, action assert, U-regex. adapt.cjs groups all matching expressions in the file and requires this site; function/line fields document the survey rather than locate edits.
- `stage3/adapt/41-explicit-any-remaining/rules.json:30`: Rule 6 requires exactly one old/new spelling: return (func as any).name; becomes return func.name;. The adapter reads rules.json and rejects a missing/duplicate reviewed site.
- `stage3/adapt/45-regex-captures/adapt.cjs:93`: Requires a unique getFunctionName implementation in debug.ts, a literal regexp exec initializer for match, and its match[1] read. The regexp proof must establish capture 1 is mandatory.
- `/workspace/scratch/getfunctionname/road-b/pins.json:61`: The accepted adapter guards array pins the complete getFunctionName implementation byte for byte. inspect() requires the owner even on an already-adapted tree. Adapter and pins blobs were checked against accepted cut 2a95448b8fadbcd30549a148bd28da5f8b75f907 via GitHub.
- `stage3/adapt/99-correctness-axolotl-n/proposed-pins.json:3`: This proposal was dropped with the four H06 assertion-call sites. Actual adapt.cjs imports pins.json, whose only owner is Debug.formatControlFlowGraph.hasNode, not proposed-pins.json.
- `stage3/adapt/99-correctness-core-multimap/owners.json:85`: Four exact import pins contain the imported caller name cast: src/compiler/checker.ts, src/services/codeFixProvider.ts, src/services/codefixes/importFixes.ts and src/services/codefixes/inferFromUsage.ts. These pin imports, not the cast implementation or its getFunctionName expression.

No targeted active whole-owner or exact-call/message pin was found for the cast or deprecate implementation. The multimap import pins are listed separately because they mention cast but constrain imports rather than its body. General parser/program checks are not classified as target-owner pins. 40's old debug.ts evidence and 99n's proposed helper rewrite are not active helper-owner guards. 99n's actual pins.json contains only Debug.formatControlFlowGraph.hasNode. The accepted road (b) helper guard is active in the requested overlay even though it has not landed on main. Its canonical accepted-cut path is stage3/adapt/99-correctness-core-debug-messages/pins.json:61; the table records the scratch copy actually inspected, whose bytes and Git blob were verified against that cut.

The getAllKeys move is on main at `stage3/adapt/99-correctness-core-getallkeys/adapt.cjs` and owners.json. This report does not apply that shape to getFunctionName.

## Evidence and reproduction

The report directory contains only measurement scripts, data, logs in archives and this report. references.json is the stock result; declarations.json and pins.json retain exact matches/guard extents; runtime-reconciled.json is the independent total. runtime-counts.json retains the piece index, source-site totals and each archive hash. Raw piece archives retain every process snapshot, ordinary guard report and test output. `raw-measurement.tar.gz` retains preparation, static search, observer/controls and wrapper exits, including the initial source-map guard failure. `validation-results.json` records the final wrapper exits and unchanged-input check. `ARCHIVES.json` and SHA256SUMS identify the retained files.

The existing build archive is linked, not rebuilt: https://github.com/system-inc/adamic/blob/385ccad0bf6cfd800443566d947cc638a7e776aa/stage3/adapt/99-correctness-core-debug-messages/after-javascript.tar.gz . Its Git object and digest are recorded in reuse-validation-summary.json.

Start from the pinned main plus the accepted road (b) source transformation and the locked existing Node/API installation. Restore the cached source via stage3/apply.sh; reuse.py verifies the already-built product. Run references.cjs with stock 6.0.3, retain the labeled rg cross-check, and run declarations.py and pins.py. instrument.cjs creates the private product, coverage-control.cjs verifies the private module graph, control.cjs verifies the observers, and run-lane.py invokes lane.py in bounded batches. reconcile.py independently reopens all 20 raw pieces before report.py can render this report. No adaptation is created, moved or deleted. All wrappers record their actual child exit and clean up their child process groups.

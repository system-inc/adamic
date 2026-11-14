# Scout: sys.ts:514 and utilitiesPublic.ts:1280

**Wall A — recommend road C for a separate follow-up:** the caller supplies a concretely typed forwarder factory; the shared cache/close helper stays generic. Stock TypeScript 6.0.3 reports **0 diagnostics on all 79 roots (0 new)**. The full CommonJSBundle sys.ts census removes **514:17**, replacing its one unique cast refusal with **one unique optional-call NotYet at 515:39**. Baseline and C each have **60 unique Refused/NotYet findings** (20/40 → 19/41); raw findings 187 → 186. No whole-root lowering success is claimed. Its emitted sys.js is **not byte-identical**; the forwarded three arguments, order, callback snapshot and subscriber order are value-identical in the reviewed source and isolated emitted-helper experiment.

**Wall B — recommend Language/Lower proof transport, not a TypeScript adaptation:** all **21 callers use 15 named declared guards** with explicit kind-comparison bodies. predicateArgumentProof nevertheless cannot carry proof/check evidence through the getFirstJSDocTag parameter into find's plain-boolean implementation. The repository's predicate rulings (#1w5bve2, #tzd3gjg, #vdt9f1q) require a proof or a check, not trust in the annotation. No compiler rule or adaptation is changed here.

Compiler and adaptations measured: `4f460a611b61a5b11c12ce61fde8ad5d0f11a912`, created as a fresh detached scratch worktree before making this report branch. Prior evidence: `716d205bad0738afec297d5ab73b31a0cb50c946`; its 73/79 ordinary first stops are context, not new measurements. All committed changes are under this report directory.

## Alternatives and full census numbers

Each row loads the complete 79-root compiler program, attempts only the exact adapted sys.ts, and exits 0 under `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=<absolute sys.ts> CENSUS_MODULE_FORMAT=CommonJSBundle`. This is the full single-file census, not a first-stop census or whole-program success. Findings deduplicate `(kind, where, reason, text)`; Boundary and SkippedDependency events are counted separately.

| Tree / shape | Stock diagnostics (new) | Units | Raw findings | Unique Refused / NotYet | Boundary / SkippedDependency | 514:17 | Newly exposed stop |
|---|---:|---:|---:|---:|---:|---|---|
| Baseline | 0 (—) | 136 | 187 | 20 / 40 | 55 / 0 | cast refusal | — |
| A: specialized File and Fs helper | 0 (0) | 137 | 186 | 19 / 42 | 56 / 0 | gone | optional calls 514:82 and 550:100 |
| B: union common call domain | 1 (+1) | 136 | 182 | 19 / 39 | 55 / 1 | body skipped; not clearance | TS2345 at 514:17 |
| C: caller supplies factory; local guard in forwarder | 0 (0) | 136 | 186 | 19 / 41 | 55 / 0 | gone | optional call 515:39 |
| C minimal optional-chain forwarder | 0 (0) | 136 | 186 | 19 / 41 | 55 / 0 | gone | same optional call 515:39 |

A copies the cache helper into two non-generic helpers, each using the exact callback type and a three-parameter arrow. The two existing calls choose their concrete helper. This avoids arbitrary-T assignment, but duplicates the implementation and exposes the original optional chain once per helper.

B tries the common call domain `("rename" | "change", never, Date?)`. The first parameter is the intersection of string and the event literals; the second is the intersection of FileWatcherEventKind and string/null/undefined, hence never. The union has this syntactic common call domain, but it cannot make a genuine forwarder for both event kinds or inhabit arbitrary rigid T. Stock TS2345 says T could be instantiated with an unrelated type. This is a failed alternative: its checker-skipped helper and changed dependency observations must not be counted as an improvement.

C adds `createForwarder: (getCallbacks: () => T[] | undefined) => T`. Each caller supplies a factory producing its own concrete callback. The factory receives `() => cache.get(path)?.callbacks.slice()`, retaining lookup and snapshot at dispatch time; it does not capture an old subscriber list. The final C forwarder saves the returned snapshot to a local and guards it before forEach. The minimal C uses getCallbacks()?.forEach instead. Both stop earlier in the getter's optional slice call at 515:39, so that refinement changes no observed census findings; no claim about later stops is made. The final A/B/C helper ASTs contain **zero casts and zero any annotations**, checked by [verify-candidates.cjs](verify-candidates.cjs).

### Every new, gone and changed finding

The full raw streams, status/command/environment files and timing streams are in [base](base/), [a](a/), [b](b/), [c](c/) and [c-optional](c-optional/). Each stream has a measurement header, one sys.ts record and a completed relation trailer. [summary.json](summary.json) holds the counts.

Exact deltas include all physical line shifts. Source-aligned deltas use unchanged-line mapping, preserve unmapped edited lines explicitly, and retain full diagnostic text. The JSON ledgers contain **every** exact and source-aligned new/gone tuple, plus co-located changed pairs; the Markdown ledgers render their kinds, locations and reasons. Raw streams retain repeated attempt/phase contexts.

| Candidate | Exact new / gone | Source-aligned new / gone | Co-located changes | Complete ledgers |
|---|---:|---:|---:|---|
| A | 82 / 80 | 4 / 2 | 1 | [JSON](a/findings-delta.json), [Markdown](a/findings-delta.md) |
| B | 81 / 82 | 3 / 4 | 0 | [JSON](b/findings-delta.json), [Markdown](b/findings-delta.md) |
| C | 43 / 43 | 2 / 2 | 1 | [JSON](c/findings-delta.json), [Markdown](c/findings-delta.md) |
| C minimal | 43 / 43 | 2 / 2 | 1 | [JSON](c-optional/findings-delta.json), [Markdown](c-optional/findings-delta.md) |

For A, the substantive removal is the cast refusal; additions are two optional-call NotYets and the second helper's rollback Boundary. The first helper's existing rollback Boundary changes its underlying diagnostic text from the cast to the optional call. For C, the cast goes, one optional-call NotYet appears, and the existing cache.set rollback Boundary changes its underlying diagnostic text. The source-aligned Boundary changes appear on both sides of those ledgers.

B removes the cast observation by making the helper ineligible, adds SkippedDependency at sys.ts:498:1 and rollback Boundaries at its callers (1192:9, 1208:9), and loses the old cache.set Boundary. It also loses a dependency finding and Boundary at core.ts:2330:13 / 2330:9 (`function identity comparison in a program with specialized generic function values`); that changed traversal/eligibility is explicitly retained, not advertised as clearance.

### Emission and value observations

Stock 6.0.3 `program.emit(sys)` emits ES2024/ESNext with verbatimModuleSyntax. Full emitted files and exact diffs are [A](a/sys.js.diff), [B](b/sys.js.diff), [C](c/sys.js.diff), and [C minimal](c-optional/sys.js.diff); none is byte-identical. A changes helper names, adds the second helper and changes the two selected calls. C adds the factory parameter, its getter closure and the concrete factories at the two call sites. B only erases annotations/cast-related parentheses/comments: its emitted forwarding expression is value-identical, but its source is ill-typed and is not an admissible candidate.

A and C retain three formal forwarder parameters (`length === 3`) and call each subscriber with all three values in the same order, including an explicit undefined timestamp. Optional absence produces undefined as before; C's block-bodied forwarder also returns undefined. Both snapshot the current callbacks before dispatch, and retain the same cache-hit reuse and close/removal behavior. [runtime-traces.json](runtime-traces.json) and [runtime.cjs](runtime.cjs) compare the emitted baseline/A/C helpers for synchronous callback-before-insertion, cache reuse, subscriber addition/removal during dispatch, argument count/value/order, lexical this, identity reuse and closing. All compared traces agree. The final emitted A/C were rechecked after correcting an initial scratch A replacement which had accidentally retained concrete casts; that invalid observation is excluded from all reported A measurements.

## Wall A: callers, exact contracts and consumers

The completed adapted tree has exactly two calls across all of src (including harness and testRunner):

| Caller | Adapted location | T | createWatcher |
|---|---|---|---|
| createSystemWatchFunctions.pollingWatchFile | sys.ts:1195 | FileWatcherCallback | `cb => pollingWatchFileWorker(fileName, cb, pollingInterval, options)` |
| createSystemWatchFunctions.fsWatch | sys.ts:1211 | FsWatchCallback | `cb => fsWatchHandlingExistenceOnHost(fileOrDirectory, entryKind, cb, recursive, fallbackPollingInterval, fallbackOptions)` |

The stock checker records the instantiated signatures in base/stock.json. Both aliases are in adapted sys.ts, not types.ts: the search of types.ts finds neither callback alias. Exact signatures:

```ts
// sys.ts:91
export type FileWatcherCallback = (fileName: string, eventKind: FileWatcherEventKind, modifiedTime?: Date) => void;
// sys.ts:85: Created=0, Changed=1, Deleted=2
export enum FileWatcherEventKind { Created, Changed, Deleted }
// sys.ts:896
export type FsWatchCallback = (eventName: "rename" | "change", relativeFileName: string | undefined | null, modifiedTime?: Date) => void;
```

Polling's worker is injected through createSystemWatchFunctions. Its production implementation fsWatchFileWorker (sys.ts) wraps `_fs.watchFile`: its fileChanged closure computes the event kind and invokes `callback(fileName, eventKind, curr.mtime)`. Removal uses fileChanged's identity with `_fs.unwatchFile`, not the forwarder's identity. The test host implementation in testRunner/unittests/helpers/virtualFileSystemWithWatch.ts stores `{cb, pollingInterval}` in watchUtils; invokeFileWatcher destructures cb and invokes it directly with three arguments.

Filesystem watching keeps the callback in fsWatchHandlingExistenceOnHost. The normal path supplies it to fsWatchWorker (or a timestamp/inode wrapper); Node's worker supplies it as `_fs.watch`'s listener. Inode transitions call it with rename/change and relative name, sometimes with a timestamp; error transitions call `callback("rename", "")`. The timestamp wrapper calls it with a timestamp; the polling fallback adapts it via createFileWatcherCallback; the missing-entry path calls it on creation. The test host fsWatchWorker stores `{cb, inode}`; fsWatchCallback destructures cb and calls it directly with event and relative name. The forwarder always supplies all three arguments downstream, including explicit undefined when the upstream call omitted the timestamp.

No in-tree consumer in these paths reads the forwarder's length, compares its identity, or uses its this. The original is an arrow and invokes each user cb as a plain call; A and C preserve that. Node may call its listener with an emitter receiver; the arrow ignores it. The harness watch registry keeps identity of the data object in a MultiMap, using that object for removal and diffing; JSON.stringify omits the cb function. createSingleWatcherPerName itself removes the original subscriber callback by identity on close, and A/C preserve that identity. Injected host functions are an external boundary: arbitrary external implementations could inspect length, identity, toString, allocation effects or other reflection. Length (3), per-cache-entry callback reuse and lexical this are preserved; source text is not.

### Adaptation owners

Literal searches of stage3/adapt for createSingleWatcherPerName, new specialized helper names, and getFirstJSDocTag return no owner-name matches. Adaptation 40 nevertheless owns the two explicit-any sites by location: evidence/sites.json declines both sys.ts:516 parameters as a generic/callback contract not completed. Any real adaptation must resolve those declined sites and its proof/evidence inputs.

Searching every edited existing function (createSingleWatcherPerName, pollingWatchFile, fsWatch, and enclosing createSystemWatchFunctions) finds only incidental text for the latter three: 47-host-errors/README.md:66 declines an unrelated catch in watchPresentFileSystemEntry and its verify.cjs injects a worker; 99-correctness-timer-closures/owners.json contains full harness class snapshots mentioning the exported watch factory and worker fields. Those are not a pin of the compiler helper or the two edited call bodies. The exact search terms and locations are saved in adapt-owner-search.json. No adaptation owner was changed.

## Wall B: static evidence at the exact head

`internal/lower/predicates.go:760` resolves an identifier to a function declaration or variable initializer, then requires a function body (or handles a proven producer call). A parameter declaration supplies neither. Its explicit annotation branch also requires exact target identity and a proven body or a checked declaration/closure. The forwarding argument `predicate` in getFirstJSDocTag therefore fails the body requirement.

`find` is core.ts's overload at lines 162–166: its narrowing overload takes a predicate, but its implementation takes a plain boolean callback. `predicateParameter` maps that overload to its boolean implementation, so checkedFunctionTypeClaim does not make the overload parameter checked. It consequently retains argument proof. This is not simply a lib.d.ts find call. The getFirstJSDocTag parameter itself is a checked bodiless function-type claim: predicateCallbackCallers cannot vouch it because its use as find's argument escapes the direct-call-only rule. A check at calls directly through that signature does not transport a check into find's boolean callback invocation.

All 21 getFirstJSDocTag callers are in utilitiesPublic.ts, passing 15 named isJSDoc*Tag functions. The all-src search includes harness and testRunner and finds no other calls. These resolve to explicit type guards in factory/nodeTests.ts, each returning the corresponding direct `node.kind === SyntaxKind.JSDoc…Tag` comparison. Yes: every passed predicate is a declared type-guard function, subject to Adamic's body proof or checked-value mechanism. This does not claim all 15 are statically body-proven: checkedPredicateDeclaration covers an unproven .ts declaration, and functionValue makes its checked forwarder when it is taken as a value. The claim is then checked at calls through that value, not by trusting the declaration annotation. Passing a guard by name is not an unchecked annotation escape. Every enclosing caller, predicate, noCache flag and adapted location is in [predicate-callers.json](predicate-callers.json); the complete all-src name search is [callers.json](callers.json).

Owner: Language/Lower proof transport, rather than a required TypeScript adaptation. docs/rulings.md:134–149 preserves the proven-body rule, .ts checked predicates (#tzd3gjg), checked predicate values (#vdt9f1q), and checked bodiless signatures (#1w5bve2). A bodiless signature carries no proof of its own; even a proven function called through an unproven signature is checked. Any forwarding solution must preserve the proof or checked closure, not merely trust the parameter annotation or collect today's callers. Source comments in predicates_function_type.go explicitly keep proof for a forwarded checked parameter at an unchecked callback consumer. The precise general forwarding rule needs Language/Lower ownership. No such rule is implemented or proposed here. The tasks themselves are not available through a task connector; the repository's attributed rulings and implementation are the evidence available here.


## Setup and reproducibility

- Scratch worktree: `/workspace/scratch/scout-514-1280`, detached at the exact requested SHA, then branch `typescript/scout-514-1280`. Shared checkout source files were not edited. The initial worktree command failed because the object was absent (`fatal: invalid reference: 4f460a611b61a5b11c12ce61fde8ad5d0f11a912`); fetching the named head and evidence branch succeeded.
- Ran `git submodule update --init --recursive`. Its nested full-history transfer was stopped (exit 143) after about 1.6 GiB and 14 minutes without a usable nested checkout. No credential error occurred. Retried `git submodule update --init --recursive --depth 1`: exit 0, cohere `5360753e19e4417675a90e92a1823bea66b13918`, nested TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. No pin changed.
- Cloned microsoft/TypeScript with `--depth 1 --branch v6.0.3` into `/workspace/scratch/typescript-603-scout`; HEAD is `050880ce59e30b356b686bd3144efe24f875ebc8`, exact tag v6.0.3, package version 6.0.3. Set ADAMIC_TYPESCRIPT_SOURCE to it. A redundant full-history upstream clone was canceled after this pin was confirmed.
- `npm ci` in the worktree's stage3/api: exit 0, three packages. Apply from this compiler head: exit 0; [apply log](base/apply.log.gz), [patch-set](base/patch-set.md). All adapted compiler input hashes and unchanged compiler/adaptation diff are in [provenance.json](provenance.json).
- Standard scratch census overlay: `python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/scout-latent-overlay`, gofmt of the scratch output, then `go build -buildvcs=false -overlay=/workspace/scratch/scout-latent-overlay/overlay.json -o /workspace/scratch/scout-latent ./stage3/census/latent/tool`, all exit 0. No production lowering rule was changed. Binary SHA256: `4fe3ce54c4d3a928bcbfa4e1dbc5a1e9ea725e9c7dc775bce7248a818b20fc2b`. The standard [audit](audit.log.txt) passes continuation, attribution, body eligibility, no-IR and disabled-production-loader checks. Final [validation](validation.log.txt) confirms patch reproduction, hashes, diagnostic counts and completed streams.
- [prepare.py](prepare.py) records the final scratch edits; [stock.cjs](stock.cjs) records the exact stock checker/emitter options, root manifest and inferred call signatures; [run.py](run.py) records the full census commands; [compare.py](compare.py) records tuple comparisons. Run on new scratch destinations. All requested full censuses completed within their 1800-second limits (A 73.37s, B 67.68s, C 70.81s; C minimal 72.67s).

Stock measurements load all 79 compiler roots together with strict, noUncheckedIndexedAccess, exactOptionalPropertyTypes, ES2024/ESNext, force module detection and bundler resolution. Pinned @types/node 25.3.3 supplies Node types to stock tsc in place of Adamic's Node library shim. An initial stock setup without those declarations produced missing-host-type diagnostics; it was corrected and is excluded. Emission enables rewriteRelativeImportExtensions with allowImportingTsExtensions; the exact options and full diagnostics are retained per candidate. No checker option is weakened. Source, adapted-tree, overlay, candidate, log and report writes stayed in scratch; npm/Go used their ordinary package/build caches under /home/agent, a cache-location exception to the requested all-writes-in-scratch setup.

## What this scout cannot see

This is a scout, not an adaptation or compiler patch. The latent census returns no usable IR and does not compile or run a native compiler. Its fixed measurement label says “measured on a checker-rejected program”; actual headers distinguish checker-clean baseline/A/C (`checker_rejected: false`) from B (`true`). Full mode is unit-scoped, with deferred generics, rollback boundaries and pass ledgers showing scoped/whole-program passes not run. B's checker/body skipping is material.

The new optional-call stop is reported, not fixed. No conclusion that A or C clears all 73 ordinary first-stop roots, lowers all sys.ts, or passes TypeScript's full compiler test suite follows from this single-file census. No new 79-independent-root ordinary census was requested or run. Stock diagnostics are the complete 79-root program check. Runtime evidence is an isolated emitted-helper experiment plus source inspection; it cannot certify arbitrary external injected host implementations, reflective source inspection, allocation/resource effects or the full Node event system. The task service for #1w5bve2 and #tzd3gjg is not available in this session; their attributed rulings, source comments and current implementation are the evidence this report can inspect. No claim is made that simply trusting today's guard callers would satisfy a general language rule.

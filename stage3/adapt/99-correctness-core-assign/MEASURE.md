# assign measurement first, task b5w3ycg

The generic hasProperty spelling passes the construction gate. The change is
owned by adaptation 40. assign and every caller remain untouched. Construction
is in progress; final results belong in PROOF.md.

Main: `f0271c6a7b2e92900eb7046d8246cfb89f52cfb6`. Stack commit: `26eb381755babe35bc07b2ff64a2fb0ba96da3a2`; its clean cherry-pick is `64d56f484863d278edfe274ffe61d659ea8e1d4a`.
TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8` (6.0.3).

## Why it refuses

At the refusal scan, arg has generic T narrowed past undefined, and its object
constraint supplies no known native ir.Object representation. The adapted
hasProperty(map: object, key: string) uses the own-property-only opaque
parameter path and requires ir.Object at that call boundary.

The complete detached_own.go was read. detachedOwnRefusal constructs the
NotYet at internal/lower/detached_own.go:94. Upstream hasProperty takes
MapLike<any>; adaptation 40 changed it to object. A plain JavaScript object
can use ir.Record storage when its declared type has a string index signature,
so prototype shape alone does not prove the helper storage contract. The
generic object parameter preserves the caller type at the library own-property
operation and avoids that opaque boundary. Its body and JavaScript are unchanged.

## Every source caller

There are eight production calls: seven compiler, zero services, and one server.
The harness adds one call. The symbol-reference audit finds no indirect uses.
Types below are declared types; literal types are inferred. Narrowing at the
call is shown when it differs. Full expressions and declarations are in callers.json.

| Adapted source | Arguments in order |
| --- | --- |
| `src/compiler/commandLineParser.ts:2712` | `CompilerOptions`; `object` |
| `src/compiler/commandLineParser.ts:3313` | `{}`; `OptionsBase \| undefined` |
| `src/compiler/commandLineParser.ts:3343` | `{}`; `MapLike<string[]>` |
| `src/compiler/commandLineParser.ts:3471` | `CompilerOptions`; `CompilerOptions \| undefined` |
| `src/compiler/commandLineParser.ts:3503` | `CompilerOptions`; `CompilerOptions \| undefined` |
| `src/compiler/commandLineParser.ts:3512` | `WatchOptions \| undefined` (at call: `WatchOptions`); `WatchOptions` |
| `src/compiler/commandLineParser.ts:3514` | `{}`; `WatchOptions \| undefined`; `WatchOptions` |
| `src/server/scriptInfo.ts:636` | `FormatCodeSettings \| undefined` (at call: `FormatCodeSettings`); `FormatCodeSettings` |
| `src/harness/harnessIO.ts:375` | `CompilerOptions & HarnessOptions`; `CompilerOptions` |

CompilerOptions, WatchOptions, OptionsBase and MapLike<string[]> have string
index signatures when present. optionMapToObject declares object, and
FormatCodeSettings has no string index signature. The {} literals have no
declared index signature. The harness intersection retains CompilerOptions'
index signature. Thus the callers are not uniformly declared dictionaries.

## Runtime over the whole lane oracle

All 398 accepted units pass the unchanged identity, API, baseline and count guards: 106366 passes, the same 1 expected failure, and 0 skips.
The accepted traversal records 30782 calls and 61568 argument appearances.
Across all actual attempts, including the retained timeout: 30783 calls and 61570 appearances.

| Property, all actual attempts | Count |
| --- | ---: |
| Object arguments | 61570 |
| Plain Object.prototype objects | 61570 |
| Null prototypes | 0 |
| Other prototypes | 0 |
| Inherited enumerable properties | 0 |
| Own accessors | 0 |
| Prototype-chain accessors | 61570 |
| Own symbol keys | 0 |
| Prototype-chain symbols | 0 |

All arguments are objects; none is undefined. The only accessor is the
ordinary inherited Object.prototype.__proto__ accessor. No getter is invoked
by the observer. Intrinsic prototypes are registered across VM contexts. Excluded
seeds prove detection of null prototypes, cross-realm objects, inherited
enumerable keys, own/inherited accessors and symbols. Every argument's raw
metadata remains in the runtime archives. RUNTIME-RECONCILED.json includes
position totals, call sites, sequence checks and the independent reconciliation.

One original stage3-lane-file-000 run hit its unchanged 40000ms before-all hook
timeout while the census and three lane units competed for CPU. Its failed
report and rows remain in runtime-01-first-attempt. The census was stopped,
retaining only completed rows. The failed unit alone was retried in isolation,
with the same limits and guards, and passed all six tests in 25.286 seconds.
The other 19 passing units were reused without rerun. The failed attempt adds
one call with two arguments, both ordinary plain objects. No failure was waived.

The independent reconciler reopens every archive, verifies internal and external
SHA256 manifests, checks process sequence continuity, identifies reused rows,
and verifies all 398 accepted reports. The observed build differs from baseline
only by one observer statement in core.ts. Production source was never
instrumented. Observer removal is required before final closeout.

## All four spellings, after runtime

Baseline full apply and types over all 79 roots pass. Its complete ordinary
census has 73 roots at core.ts:1364:29, two accepted roots, and four other
NotYet outcomes. Completed rows were retained across recorded scheduling
stops; remaining groups ran two at a time with the unchanged ordinary inspect
function. No compiler, checker, refusal rule, backend or fallback was changed.

| Spelling | Types, 79 roots | Whole core JS, comments kept / removed | Former 73 roots |
| --- | --- | --- | --- |
| `assign<T extends MapLike<unknown>>` | exit 1, 1 diagnostics | identical / identical | 73 checker at commandLineParser.ts:2712:36 |
| `assign<T extends object & { [K in keyof T]: unknown }>` | exit 0, 0 diagnostics | identical / identical | 73 NotYet at core.ts:1364:29 |
| `const record: MapLike<unknown> = arg; hasProperty(record, p)` | exit 1, 1 diagnostics | changed / changed | 73 checker at core.ts:1363:15 |
| `hasProperty<T extends object>(map: T, key: string)` | exit 0, 0 diagnostics | identical / identical | 73 Refused at core.ts:1484:12 |

Each experiment uses the whole adapted core.ts. All other 745 source files
and every caller are byte-identical. No cast is introduced. Exact types logs,
complete trial sources, emission hashes and per-root logs are retained.
CENSUS.md reports every group of 20, 20, 20 and 19; CENSUS.json retains all
per-file results and diagnostics. Checker-rejected programs are never lowered.

The MapLike constraint fails at the object-typed optionMapToObject return.
The self-mapped constraint preserves compatibility but leaves the abstract T
storage wall unchanged. The checked record local is rejected by TS2322 and
also changes JavaScript. The generic helper passes all gates and advances
all 73 roots to the existing cast refusal at core.ts:1484:12.

## Ownership and accepted implementation

The audit searched every adapt.cjs and every owner/site JSON, including
singular filename forms: 164 files. No existing rule pins assign's owner.
Adaptation 32 mentions two assign result reads, which it owns separately.
The successful spelling edits hasProperty, whose existing owner is 40.

40 owns the exact whole-function upstream, previous-object and generic pins
and the only planner. Its early phase retains the original object output so
all downstream adapters get stable inputs. The requested 99 directory delegates
to the same owner's final generic phase. There is no duplicate owner rule.
Changed, missing, duplicate, nested, moved and half-adapted owners are refused
before writing. The contract passes no-write mutants, foreign-edit preservation,
LF/CRLF preservation and byte/mtime-preserving replay.

The final full apply, paired replay, both public declaration comparisons and
original-helper restoration mutant are required before completion. Their
results will be recorded in PROOF.md.

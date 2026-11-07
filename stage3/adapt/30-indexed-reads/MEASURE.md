# getOwnValues measurement, task b5w3ycg

The 73 roots reach core.ts:1326:26 only after elementAt's wall is past.
This task leaves elementAt unchanged. The ordinary binder results below retain
any earlier wall; extracted-owner probes are separately labelled and do not
claim that the full compiler graph advanced.

Base: origin/main 8f4efea05dbe2183cd7347ad1bdbec409c016f2d, fetched before
creating typescript/getownvalues. Upstream: microsoft/TypeScript
050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3). The parser and checker
are stock TypeScript 6.0.3. The ordinary lowering probe is built from this
unmodified Adamic base, without an overlay, refusal suppression or compiler edit.

## Ownership

A search of stage3/adapt finds one pre-existing getOwnValues pin:
30-indexed-reads/sites.json, the entry at line 822. It pins the parsed indexed
expression `(collection as MapLike<T>)[key]`, occurrence 1 of 1 in core.ts,
function getOwnValues. Its action is `assert`, class U-table, with the invariant
that required payloads selected by own-key enumeration must be defined.

This is a pin of this read, rather than a hash of the complete function.
The adapter's plan function matches the expression with `!` stripped, checks
its occurrence count and adds or validates the existing presence check.
It does not prove the receiver is a MapLike, and does not own a checked cast
that could distinguish an array from a dictionary. Under the requested
ownership rule, adaptation 30 owns any later edit. This measurement therefore
lives here. No existing adapter, site ledger or owner text is changed.

## Measurement method

getownvalues-measure.cjs parses the complete adapted core.ts and substitutes
only getOwnValues in a CompilerHost. It uses getTypeAtLocation at the indexed
read, getResolvedSignature at every call and the complete stock checker
pre-emit diagnostic list. Diagnostic error types are not accepted types.

The declared-project measurement keeps src/tsconfig-base.json's es2020 lib and
ES2020 target. The Adamic-options measurement uses exactly the fixed options
in internal/load/load.go: strict, noUncheckedIndexedAccess,
exactOptionalPropertyTypes, verbatimModuleSyntax, erasableSyntaxOnly false,
allowImportingTsExtensions, noEmit, ESNext modules, force module detection,
Bundler resolution, ES2024 target, lib.es2024.d.ts and an empty types list.
It does not inherit the project's explicit strictBindCallApply false and
useUnknownInCatchVariables false exceptions. Stock TypeScript is used in both
rows; this is not a claim that its library is Adamic's custom prelude.

Both comment modes compare the complete core.ts JavaScript, using the stock
transpileModule emitter, the measured target and ESNext modules. These are
complete-file emission comparisons, not function-fragment comparisons or a
claim that checker-rejected candidates are valid. Ordinary probes use isolated
copies of the completed baseline src tree, changing only this function.

The supplementary focused probes copy the exact measured function into a small
module with the actual MapLike string-index contract and the Object.prototype
hasOwnProperty alias, and import it from a small binder.ts. They use the same
ordinary load and lower entry points. Their locations are fixture locations,
not full compiler coordinates. A separate readonly relation control omits the
indexed read, isolating the native storage view gate.


## Checker types and candidates

All spellings retain the existing `!`. The candidate definitions are:

* baseline: `(collection as MapLike<T>)[key]!`
* direct: `collection[key]!`
* readonly-view: `(collection as Readonly<MapLike<T>> | readonly T[])[key]!`
* readonly-parameter: parameter `Readonly<MapLike<T>> | readonly T[]`, read `collection[key]!`
* readonly-record: parameter `Readonly<MapLike<T>>`, read `collection[key]!`
* record-parameter: parameter `MapLike<T>`, read `collection[key]!`
* generic-collection: `<T, C extends MapLike<T> | T[]>(collection: C)`, read `collection[key]!`
* keyof: `collection[key as keyof typeof collection]!`, a non-qualifying assertion control

| Candidate | collection at read | key at read | Indexed type, es2020 | Indexed type, es2024 | Type after !, es2024 |
| --- | --- | --- | --- | --- | --- |
| baseline | `MapLike<T> \| T[]` | `string` | `T` | `T \| undefined` | `NonNullable<T>` |
| direct | `MapLike<T> \| T[]` | `string` | `any` | `any` | `any` |
| readonly-view | `MapLike<T> \| T[]` | `string` | `any` | `any` | `any` |
| readonly-parameter | `Readonly<MapLike<T>> \| readonly T[]` | `string` | `any` | `any` | `any` |
| readonly-record | `Readonly<MapLike<T>>` | `string` | `T` | `T \| undefined` | `NonNullable<T>` |
| record-parameter | `MapLike<T>` | `string` | `T` | `T \| undefined` | `NonNullable<T>` |
| generic-collection | `C` | `Extract<keyof C, string>` | `C[Extract<keyof C, string>]` | `C[Extract<keyof C, string>]` | `NonNullable<C[Extract<keyof C, string>]>` |
| keyof | `MapLike<T> \| T[]` | `string` | `any` | `any` | `any` |

The `any` entries are checker error types accompanied by TS7053. They are not admitted `any` bridges. In the keyof control the variable itself remains string; the asserted index type is recorded in the raw casts array. The readonly receiver view is TypeScript-assignable from the original union; the original member cast and keyof assertion are not upcasts.

| Candidate | New diagnostics relative to each stock baseline | Qualification |
| --- | --- | --- |
| baseline | none | Original refused member downcast |
| direct | TS7053 at core.ts:1326:25 | No string index common to both members |
| readonly-view | TS7053 at core.ts:1326:25 | Same TS7053; separate storage-view gate also stops it |
| readonly-parameter | TS7053 at core.ts:1326:25 | All seven calls retain compatible arguments, but body fails |
| readonly-record | TS2345 at symbolWalker.ts:52:57; TS2345 at symbolWalker.ts:52:101; TS2345 at symbolWalker.ts:62:57; TS2345 at symbolWalker.ts:62:101 | All four sparse-array compiler calls fail |
| record-parameter | TS2345 at symbolWalker.ts:52:57; TS2345 at symbolWalker.ts:52:101; TS2345 at symbolWalker.ts:62:57; TS2345 at symbolWalker.ts:62:101 | All four sparse-array compiler calls fail |
| generic-collection | TS2345 at core.ts:1326:25 | Precise key, value not proven T |
| keyof | TS7053 at core.ts:1326:25 | Still TS7053; adds an assertion rather than a proof |

Stock baseline totals: compiler declared options 0, compiler Adamic options 60; services source probe declared options 145, services Adamic options 2825. The exact diagnostics are retained. The strict source-service probe imports compiler source directly without building project-reference outputs. Its existing file-list, Node-global and strictness findings are not a successful service build. Stock TypeScript with the exact Adamic options also lacks Adamic's native Node/library overlay. These are distinct from the real Adamic loader, whose ordinary baseline binder probe passes checking and reaches Lower.

## Every source caller

`rg getOwnValues` over the complete adapted src tree finds exactly the following seven calls, plus imports and the definition. Each argument type is the same under both measured option sets.

| Call | Argument | Argument type |
| --- | --- | --- |
| compiler/symbolWalker.ts:52:44 | `visitedTypes` | `Type[]` |
| compiler/symbolWalker.ts:52:88 | `visitedSymbols` | `Symbol[]` |
| compiler/symbolWalker.ts:62:44 | `visitedTypes` | `Type[]` |
| compiler/symbolWalker.ts:62:88 | `visitedSymbols` | `Symbol[]` |
| services/refactors/convertImport.ts:76:12 | `actions` | `{ 0: { name: string; description: string; kind: string; }; 2: { name: string; description: string; kind: string; }; 1: { name: string; description: string; kind: string; }; }` |
| services/refactors/convertImport.ts:87:20 | `actions` | `{ 0: { name: string; description: string; kind: string; }; 2: { name: string; description: string; kind: string; }; 1: { name: string; description: string; kind: string; }; }` |
| services/refactors/convertImport.ts:97:27 | `actions` | `{ 0: { name: string; description: string; kind: string; }; 2: { name: string; description: string; kind: string; }; 1: { name: string; description: string; kind: string; }; }` |

The four compiler calls use arrays explicitly documented as sparse id-to-value tables in symbolWalker.ts:45-46. A record-only signature rejects all four with TS2345 at argument columns 57 and 101 on lines 52 and 62. All three actions-object service calls continue to fit the record-only alternatives.

For the readonly union parameter, the full seven-call source probe adds exactly one diagnostic under either option set: TS7053 in getOwnValues at core.ts:1326:25. It adds no diagnostic at any caller and leaves the resolved element/result types unchanged. This checks call compatibility, while the body still fails; it does not claim a successful service build.

## Complete JavaScript bytes

All eight spellings are byte-identical to the baseline complete core.ts JavaScript in both comment modes and both target/lib measurements. Invalid candidates still fail their checker gates. The raw files retain source hashes, output sizes, hashes and equality results per candidate. Baseline hashes follow.

| Target | removeComments | Bytes | SHA-256 |
| --- | --- | ---: | --- |
| declared (ES2020) | False | 65165 | `5f618769e17b09fb78e02c861a44f46b0bb632c3245d4838efc98f2e263005fb` |
| declared (ES2020) | True | 46052 | `56c1344d4aa8459e1fc1d149163c6d5f894d29cabf6baa1c629f569a1036cce4` |
| adamic (ES2024) | False | 65100 | `75d076a9396da0279e46de0c0a593f5d06c190ddba3572b37b8e23fc63540523` |
| adamic (ES2024) | True | 45987 | `b81d42f936ead2d454b0ed0ca380bd800f6be6c0ce27a2bc176eddc70cd725f7` |

## Ordinary binder and extracted-owner results

Each ordinary candidate probe uses the full completed src tree. Only its pinned function spelling differs. Both rows from the ordinary tool agree in every case apart from timing. No elementAt edit, substituted compiler, refusal scan or unchecked lowering of a rejected program is used.

| Candidate | Ordinary binder first stop | Extracted-owner first stop |
| --- | --- | --- |
| baseline | Refused: core.ts:1076:100: a cast the runtime can't check | Refused: core.ts:7:26: a cast the runtime can't check |
| direct | checker: core.ts:1326:25 (TS7053) | checker: core.ts:7:25 (TS7053) |
| readonly-view | checker: core.ts:1326:25 (TS7053) | checker: core.ts:7:25 (TS7053) |
| readonly-parameter | checker: core.ts:1326:25 (TS7053) | checker: core.ts:7:25 (TS7053) |
| readonly-record | checker: symbolWalker.ts:52:101 (TS2345); symbolWalker.ts:52:57 (TS2345); symbolWalker.ts:62:101 (TS2345); symbolWalker.ts:62:57 (TS2345) | NotYet: core.ts:3:33: a value of type Readonly<MapLike<T>> |
| record-parameter | checker: symbolWalker.ts:52:101 (TS2345); symbolWalker.ts:52:57 (TS2345); symbolWalker.ts:62:101 (TS2345); symbolWalker.ts:62:57 (TS2345) | accepted |
| generic-collection | checker: core.ts:1326:25 (TS2345) | checker: core.ts:7:25 (TS2345) |
| keyof | checker: core.ts:1326:25 (TS7053) | checker: core.ts:7:25 (TS7053) |

The baseline ordinary binder stops at elementAt, core.ts:1076:100. Checker-rejected candidates stop before Lower, at their own diagnostics. No ordinary full-tree result here reaches 1326:26. The extracted baseline independently reproduces the same member-cast refusal at fixture core.ts:7:26. The mutable dictionary-only body lowers, but its full-tree callers fail. The readonly dictionary-only fixture stops NotYet at its parameter, core.ts:3:33, `a value of type Readonly<MapLike<T>>`.

The separate read-free readonly relation control passes the cast rule but stops at core.ts:3:12 in recordStorageView: `a T[] | MapLike<T> seen as Readonly<MapLike<T>> | readonly T[] (fixed objects and records have different storage; copy explicitly)`. The cast rule has ordinal 1900 and the storage-view rule ordinal 2000. TypeScript assignability alone therefore does not make this an admitted erased view.

## Why Lower refuses, and what it currently accepts

The reason text is constructed by castProof at internal/lower/cast_proof.go:188. castRefusal, lines 156-172, calls it and is registered at ordinal 1900. refuse in internal/lower/refusals.go:104 visits the registered rules. The upcast branch at cast_proof.go:219 asks provenRelation. The checked-interface fallback at lines 270-278 is restricted to a non-union source. For a union, the remaining path requires matching members, safe union writes and a common distinguishing literal/enum property (lines 294-335). MapLike<T> and T[] have no such discriminant, so the final refusal remains; an own-property test on the enumerated key does not distinguish those receiver representations.

The readonly relation control reaches recordStorageView, internal/lower/records.go:446-459, through the rule registered at line 503. sameRecordStorage at line 553 preserves the native storage kind. It cannot erase a mutable record into a readonly fixed-object representation just because TypeScript accepts the type relation.

For enumeration, forIn in internal/lower/library_for_in.go:13 rejects arrays at line 17 and requires a record, a supported readonly object view, a construction, or a proven fixed plain-object origin at lines 22-29. recordElement in records.go:12 requires an object with a pure mutable string index signature and excludes arrays. readonlyObjectView in records_entries.go:67 likewise requires an object view, not this array/dictionary union. There is no existing for-in read rule that turns this union and an arbitrary string key into T. A true mutable dictionary parameter admits the read, as the focused positive control shows, but the actual array calls preclude that contract.

The governing rulings are docs/rulings.md:9-18 (checked member casts, unchecked assertions, proven upcasts and checked views), :92 (mutable invariance and readonly covariance), :170 (enumeration over proven objects/records, arrays still NotYet), and :271 (record declared-key reads and writes). The existing `!` is the separately owned required-value check; it proves neither a dictionary receiver nor an enumerable array-key type. The older for-in statement at :260 must be read with the explicit Nov 1 ruling at :170.

## Decision and options requiring a ruling

No measured spelling satisfies all three gates: a true type with compatible callers, no new downcast, and an ordinary Adamic admission with byte-identical JavaScript. Stop after Step 1. No adapter amendment or Step 2 proof, mutant, paired replay, 79-root types gate, declaration comparison, nine-file emission comparison or grouped census is claimed.

1. Request an explicit language/lowering ruling for own-enumerable string-key reads over an array/dictionary union, including sparse holes, key order, named own enumerable array properties and the existing required-value check. The checker-facing type and native storage behavior both need a true contract.
2. Ask Adamic to rule on a runtime implementation such as Object.values or explicit receiver/key checks. These change emitted JavaScript and may need new lowering support; no equivalence or admission is asserted here.
3. Split the dictionary contract from sparse-array collection, with separately reviewed symbolWalker caller changes and a sparse-array implementation preserving own-key behavior. Removing arrays from the signature alone is already disproved by the four measured calls.

Numeric key conversion, extra string index signatures on arrays and further key/member assertions are not type-only proofs of the original contract. Numeric conversion and Object.values were not measured or built in this run.

## Reproduction and retained evidence

Setup and the completed full apply both exited 0. The initial full-history clone was cancelled and replaced by the exact shallow v6.0.3 clone. An early apply attempted before checkout finished refused the missing pin; it made no accepted baseline. A duplicate probe build was stopped, then rebuilt successfully after sourcing setup's final build flags. The retained logs distinguish these attempts.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /workspace/scratch/gov-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
stage3/apply.sh --build-product /workspace/scratch/gov/before
npm ci --prefix /workspace/scratch/gov/before
go build -o /workspace/scratch/gov/tool ./stage3/census/tool
# Extract the evidence scripts into /workspace/scratch/gov, then:
python3 /workspace/scratch/gov/run-measure.py
python3 /workspace/scratch/gov/services.py
/workspace/scratch/gov/tool /workspace/scratch/gov/before/src/compiler/binder.ts /workspace/scratch/gov/baseline-binder.jsonl
GOMAXPROCS=2 python3 /workspace/scratch/gov/probe.py
python3 /workspace/scratch/gov/focus.py
```

Actual command outputs were written to files. The evidence archive includes the exact scripts, raw checker results, all ordinary/focused rows, source inventory, relevant Lower source as .go.txt, complete baseline core.ts, compiler provenance and logs. getownvalues-results.json is the readable aggregate. The archive has an external SHA-256 and an internal SHA256SUMS list. Measurement evidence changes product keys but introduces no adapter transformation. All processes started for this task have completed or were explicitly stopped.

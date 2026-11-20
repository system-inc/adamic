# Readonly family 3 views, adaptation 99-readonly-family3-views-2

Adamic, task #k8ed1f2, family 3's causes 4 and 11, by Adamic's delegation:
readonly where tsc never writes, the ruling 99-readonly-family3-views applied.
This is type-only: the emitted JavaScript does not change. Every declaration
it touches is `@internal`, so the published declaration files do not change,
this directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Why

Family 3 is the refusal "a value of type A seen as B, which can write X where
Y is read". Both causes here are a property that tsc only reads through the
view, but that the view leaves writable.

Cause 4 is a `BuilderProgramState` seen as `ReusableBuilderProgramState`.
99-readonly-family3-views made the Reusable view's two diagnostics maps
`ReadonlyMap<...>`, so nothing can be written into the maps through it. The
property itself stayed writable: through the view a holder could put a
`ReadonlyMap` where `BuilderProgramState` reads a `Map`. Those 9 sites stopped
for that different reason after the earlier adaptation, and `readonly` on the
property takes the reason away.

Cause 11 is a `DiagnosticMessageChain` seen as a `ReusableDiagnosticMessageChain`,
or as `DiagnosticMessageChain | ReusableDiagnosticMessageChain`, the constraint
of `convertOrRepopulateDiagnosticMessageChain` and its array twin. Through
that view a holder could write a `ReusableDiagnosticMessageChain[]` into
`next`, where the chain's own type reads `DiagnosticMessageChain[]`. Both
Reusable chains, `ReusableRepopulateInfoChain` and
`SerializedDiagnosticMessageChain`, declare `next` the same way, and tsc only
builds them fresh and reads them.

## The edit

`adapt.cjs` edits 4 slots, all in `src/compiler/builder.ts`. It parses the
file with stock TypeScript 6.0.3 and finds each slot by its interface or type
alias and property name, never by line, offset or whole-line text. The alias's
property is the one member of the type literal that
`SerializedDiagnosticMessageChain` intersects with its `Omit<...>`, so the two
`next` declarations, which are the same line of text, are told apart by their
owners.

Each slot gains `readonly` before its name. Where its type changes, the type
node's span takes the adapted text in the same edit, and what lies between
the name and the type, `?: ` or `: `, keeps its bytes. Before any write it
refuses unless every slot is found exactly once, is optional exactly when
upstream declares it so, and holds either the text it expects with no
modifier or exactly its adapted text with `readonly`; a half-adapted slot
refuses too. Its declaration must be local to a function, file-local, or carry
`@internal` on itself or an enclosing declaration. A slot already adapted
keeps its bytes, so a second run changes nothing. It prints one line per slot:
its name, `edited` or `already adapted`, where it is and why it is not public.

Cause 4's expected text is 99-readonly-family3-views' output, not upstream's.
If that adaptation has not run, the slot holds upstream's `Map<...>` and the
script refuses rather than make a writable map's slot readonly on its own.

Sites are the meter's positions in the adapted tree, relative to
`src/compiler/`. A cause's sites are listed once, on its first slot; the
slots under it clear them together.

| Cause | Slot | Before | Adapted | Sites cleared |
|---:|---|---|---|---|
| 4 | `ReusableBuilderProgramState.semanticDiagnosticsPerFile` | `ReadonlyMap<Path, readonly ReusableDiagnostic[] \| readonly Diagnostic[]>`, from 99-readonly-family3-views | the same, `readonly` | 9: builder.ts:1709:90, 1717:28, 1747:62, 1771:62, 1936:62, 2035:62, 2113:62, 2153:62, 2214:62 |
| 4 | `ReusableBuilderProgramState.emitDiagnosticsPerFile`, optional | `ReadonlyMap<...> \| undefined`, from 99-readonly-family3-views | the same, `readonly` | with the row above |
| 11 | `ReusableRepopulateInfoChain.next`, optional | `ReusableDiagnosticMessageChain[]` | `readonly ReusableDiagnosticMessageChain[]`, `readonly` | 6: builder.ts:527:76, 580 twice, 1564:16, 1569:16, 1578:15 |
| 11 | `SerializedDiagnosticMessageChain`'s `next`, optional | `ReusableDiagnosticMessageChain[]` | `readonly ReusableDiagnosticMessageChain[]`, `readonly` | with the row above |

Each row is one slot, and `adapt.cjs` lists the same 4. Where a cell says
`...`, the elided text is the row above's, unchanged. Upstream 6.0.3 declares
the four at builder.ts:140, 142, 117 and 122.

`BuilderProgramState` redeclares both maps as writable `Map`s (213, 215). A
derived interface may drop `readonly` from a base's property, so that stays
legal, and every write tsc makes goes through that override.

By the types, both `next` slots are needed. At 1564 a `DiagnosticMessageChain`
is returned as a `ReusableDiagnosticMessageChain`; it has no `info`, so it
lands on `SerializedDiagnosticMessageChain` alone, and only that slot's
`readonly` covers it. At the constraint sites, 527 and 580, the view is the
union of all three chains, and the write the refusal names,
`ReusableDiagnosticMessageChain[]`, comes from both Reusable members. tsc
reads a union's property as readonly when any member's is, but each member's
`next` is a slot of its own, so both take the modifier. A mutant restoring
one `next` at a time would show it site by site. The constraint itself is not touched: with the edit its
`next` reads as `DiagnosticMessageChain[] | readonly ReusableDiagnosticMessageChain[] | undefined`,
and the three `chain.next as T[]` casts in `convertOrRepopulateDiagnosticMessageChain`
and the `chain.next.every(isFullChain)` read 99-correctness-axolotl-k added to
its array twin still type-check, since a call on a union of a mutable and a
readonly array goes through a readonly array of the union.

1569 and 1578 are views of the array `toReusableDiagnosticMessageChainArray`
returns, `forEach(...) || array`, rather than of a chain. The cause list puts
them under cause 11; the census below says whether `readonly next` clears
them or they stay for a cause of their own.

The directory sorts right after 99-readonly-family3-views and before
99-readonly-inferred-bindings (apply runs `sorted(...)` over `stage3/adapt`'s
directories, and `-2` extends the earlier name), so it reads the earlier
adaptation's `ReadonlyMap` text as it expects. 99-correctness-axolotl-k,
which pins `convertOrRepopulateDiagnosticMessageChain` and its array twin by
hash, runs before both and pins function bodies, not these declarations. The
four that sort after it, 99-readonly-inferred-bindings,
99-readonly-script-target-entries, 99-truthful-generated-name-emit-node and
99-wasm-crypto-import, neither read nor pin any of them: the first edits
`getSemanticDiagnosticsOfNextAffectedFile`'s `result` in builder.ts by name,
and the others edit other files.

## Left out

- Cause 9, 7 sites, `addDeprecatedSuggestion`'s `declarations`. As
  99-readonly-family3-views says, `readonly Node[]` there needs its worker to
  take `Node | readonly Node[]`, and the worker's `Array.isArray(declarations)`
  does not narrow `readonly Node[]` out of the false branch. That takes a body
  change, not an annotation, so it is not in this type-only head.

## Writes elsewhere

Every write found to a property this adaptation touches, and why none goes
through the view it makes readonly. The search covered all of tsc's `src`,
harness, services and tests included: every assignment (`=`, `??=`, `||=`,
`&&=`), every `.set(`, `.delete(`, `.clear(`, `.push(`, `.splice(`, `.sort(`
and other mutating call on the property, and every `.length` or index write.
Lines are upstream 6.0.3's.

- `semanticDiagnosticsPerFile` (cause 4). Assigned at builder.ts:330 and 2168,
  `.set` at 420, `.clear` at 496, `.delete` at 894, each through
  `BuilderProgramState` (`createBuilderProgramState`'s state, cast
  `as BuilderProgramState` at 325, `removeSemanticDiagnosticsOf`'s parameter,
  and `createBuilderProgram`'s state, narrowed to
  `BuilderProgramStateWithDefinedProgram`). `.set` at 1068 goes through a
  local typed `BuilderProgramState["semanticDiagnosticsPerFile"]` (1056), first
  defaulted from `state` at 1058 with `??=`, a write to the local, not to the
  property. 2156 builds such a map fresh.
- `emitDiagnosticsPerFile` (cause 4). Assigned at 497 and 1886, `??=` and
  `.set` at 405 and 1858, `.delete` at 782 and 2225, each through
  `BuilderProgramState` or `BuilderProgramStateWithDefinedProgram`. 1879 to
  1883 fill a local `Map<Path, Diagnostic[]>` before 1886 stores it.
- The Reusable view of both is only read: at 403 and 418 through
  `createBuilderProgramState`'s `oldState: Readonly<ReusableBuilderProgramState>`,
  already readonly upstream, and in the tests at
  testRunner/unittests/helpers/baseline.ts:79 and
  testRunner/unittests/tscWatch/incremental.ts:211 to 214 through
  `BuilderProgram.state`. It is built fresh at 2276 to 2277 and 2314 to 2315.
  `builderProgram.state = state` (1694) writes `BuilderProgram.state`, not
  these properties. The other hits share only the name: `IncrementalBuildInfo`'s
  fields (1135, 1136, 1272, 1273, 1368, 1369, tsbuildPublic.ts:1564, 1567) and
  the test helpers' readable build info (baseline.ts:165 to 281, tsc.ts:274 to
  438).
- `next` on `ReusableRepopulateInfoChain` and `SerializedDiagnosticMessageChain`
  (cause 11). No write anywhere. The one assignment to a chain's `next`,
  utilities.ts:8773 in `concatenateDiagnosticMessageChains`, goes through
  `DiagnosticMessageChain`. The Reusable chains are built fresh at
  builder.ts:1535 to 1538 and 1541; the `result.push` calls at 1550 and 1552
  fill a local `ReusableDiagnosticMessageChain[]` before it becomes a fresh
  chain's `next`. programDiagnostics.ts:292 copies a `DiagnosticMessageChain`'s
  `next` into `details` and writes no `next`. factory/emitHelpers.ts:1122's
  `g.next = verb(0)` is the emitted `__generator` helper's text, not tsc's
  code. No harness, service or test file names either type.

## Public consequence

None. `ReusableBuilderProgramState`, `ReusableRepopulateInfoChain` and
`SerializedDiagnosticMessageChain` each carry `@internal`, and the script
refuses any slot that is not local, file-local or `@internal`. Types erase, so
no line of the emitted JavaScript and no line of typescript.d.ts moves. This
directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Proof

Sealed, at commit PENDING over main PENDING, stock TypeScript 6.0.3.

Apply and stock tsc (J-PENDING): apply builds the adapted tree with and
without this directory. A second run of `adapt.cjs` prints all 4 slots as
already adapted and changes no file. Exactly PENDING file of `src/compiler`
differs, builder.ts. Stock tsc over tsc's whole program reports 0 diagnostics
on both trees. The JavaScript and declaration output are byte-identical with
and without. `stage3/lane` passes PENDING.

The census (J-PENDING without, J-PENDING with): scan-all's instrumented probe
over builder.ts's closure.

| Measurement | Without | With |
|---|---:|---:|
| Stock tsc diagnostics | 0 | 0 |
| Distinct stops over the closure | PENDING | PENDING |
| Sites that stop no more | | PENDING |
| Sites that stop for a different reason | | PENDING |
| Sites that stop and didn't before | | PENDING |

## Mutants

Sealed, each taking one cause back out of `adapt.cjs` and scanning
builder.ts's closure against the adaptation as it is (J-PENDING).

| Mutant | Brings back |
|---|---:|
| Cause 4: `readonly` taken off both `ReusableBuilderProgramState` maps | PENDING |
| Cause 11: `readonly` and the readonly array taken off both `next` slots | PENDING |

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.

# Readonly family 3 views, adaptation 99-readonly-family3-views-2

Adamic, task #k8ed1f2, family 3's cause 4, by Adamic's delegation: readonly
where tsc never writes, the ruling 99-readonly-family3-views applied. This is
type-only: the emitted JavaScript does not change. Every declaration it
touches is `@internal`, so the published declaration files do not change,
this directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Why

Family 3 is the refusal "a value of type A seen as B, which can write X where
Y is read". Cause 4 is a `BuilderProgramState` seen as
`ReusableBuilderProgramState`, at 9 sites. `BuilderProgramState` extends the
Reusable view and narrows four of its fields:

- `semanticDiagnosticsPerFile` and `emitDiagnosticsPerFile`, writable `Map`s
  where the view, after 99-readonly-family3-views, holds `ReadonlyMap`s;
- `changedFilesSet`, required where the view's is optional;
- `affectedFilesPendingEmit`, a `Map` where the view's is a `ReadonlyMap`.

While the view leaves any of the four writable, a holder could put there what
`BuilderProgramState` does not read: a `ReadonlyMap` where it reads a `Map`,
or `undefined` where it reads a `Set<Path>`. Each site stops on the first such
field it meets, so the census measured `readonly` on the two diagnostics maps
alone as clearing none of the 9: each site moved on to `changedFilesSet`.
All four take `readonly` together, and tsc only reads them through the view.

Those four are the whole of the narrowing. Compared member by member,
`BuilderProgramState` redeclares no other field of
`ReusableBuilderProgramState`: its other members are its own, and neither
interface redeclares a member of `BuilderState`, which both extend.
`program`, optional on the view, is narrowed only by
`BuilderProgramStateWithDefinedProgram`, which is not the view these sites
name, and tsc writes it through `BuilderProgramState` (builder.ts:326, 619),
whose `program` is the view's own declaration, so it could not be readonly.

## The edit

`adapt.cjs` edits 4 slots, all on `ReusableBuilderProgramState` in
`src/compiler/builder.ts`. It parses the file with stock TypeScript 6.0.3 and
finds each slot by its interface and property name, never by line, offset or
whole-line text. The edit inserts `readonly ` before the property's name; the
type keeps its bytes. Before any write it refuses unless every slot is found
exactly once, is optional exactly when upstream declares it so, holds the type
it expects, and carries no modifier or exactly `readonly`. Its declaration must
be local to a function, file-local, or carry `@internal` on itself or an
enclosing declaration. A slot already adapted keeps its bytes, so a second run
changes nothing. It prints one line per slot: its name, `edited` or `already
adapted`, where it is and why it is not public.

The type each slot expects is the one the adaptations before it leave. The two
diagnostics maps hold 99-readonly-family3-views' `ReadonlyMap<...>`; if that
adaptation has not run, the slot holds upstream's `Map<...>` and the script
refuses rather than make a writable map's slot readonly on its own.
20-optional-declarations spells out an optional property's `| undefined`
where tsc writes an absent value to it: its record of the edited declarations
names `affectedFilesPendingEmit` on the view, and not `changedFilesSet`.
Upstream already writes it on `emitDiagnosticsPerFile`.

Sites are the meter's positions in the adapted tree, relative to
`src/compiler/`. The cause's sites are listed once, on its first slot; the
four slots clear them together.

| Cause | Slot | Type, kept | Sites cleared |
|---:|---|---|---|
| 4 | `ReusableBuilderProgramState.semanticDiagnosticsPerFile` | `ReadonlyMap<Path, readonly ReusableDiagnostic[] \| readonly Diagnostic[]>`, from 99-readonly-family3-views | 9: builder.ts:1709:90, 1717:28, 1747:62, 1771:62, 1936:62, 2035:62, 2113:62, 2153:62, 2214:62 |
| 4 | `ReusableBuilderProgramState.emitDiagnosticsPerFile`, optional | `ReadonlyMap<...> \| undefined`, from 99-readonly-family3-views | with the row above |
| 4 | `ReusableBuilderProgramState.changedFilesSet`, optional | `Set<Path>`, upstream's | with the row above |
| 4 | `ReusableBuilderProgramState.affectedFilesPendingEmit`, optional | `ReadonlyMap<Path, BuilderFileEmit> \| undefined`, from 20-optional-declarations | with the row above |

Each row is one slot, and `adapt.cjs` lists the same 4. Where a cell says
`...`, the elided text is the row above's, unchanged. Upstream 6.0.3 declares
the four at builder.ts:140, 142, 146 and 158.

`BuilderProgramState` redeclares all four writable (213, 215, 219, 251). A
derived interface may drop `readonly` from a base's property, so that stays
legal, and every write tsc makes goes through those redeclarations.

The directory sorts right after 99-readonly-family3-views and before
99-readonly-inferred-bindings (apply runs `sorted(...)` over `stage3/adapt`'s
directories, and `-2` extends the earlier name), so it reads the earlier
adaptation's `ReadonlyMap` text as it expects. The four that sort after it,
99-readonly-inferred-bindings, 99-readonly-script-target-entries,
99-truthful-generated-name-emit-node and 99-wasm-crypto-import, neither read
nor pin any of these declarations: the first edits
`getSemanticDiagnosticsOfNextAffectedFile`'s `result` in builder.ts by name,
and the others edit other files.

## Left out

- Cause 11, 6 sites, `next` on `ReusableRepopulateInfoChain` and
  `SerializedDiagnosticMessageChain`, a `DiagnosticMessageChain` seen as a
  `ReusableDiagnosticMessageChain` or as the union
  `convertOrRepopulateDiagnosticMessageChain` constrains its chain to. Nothing
  writes either `next`, and `readonly` on both, typed
  `readonly ReusableDiagnosticMessageChain[] | undefined`, was measured: it
  cleared 5 sites (builder.ts:527:76, 580:106, 580:137, 1564:16, and cause 2's
  1572:62), kept 1569:16 and 1578:15, which are views of the array
  `toReusableDiagnosticMessageChainArray` returns rather than of a chain, and
  made a new stop at 556:16, `chain` as a `DiagnosticMessageChain` in
  `convertOrRepopulateDiagnosticMessageChain`, where the readonly `next` is
  seen as writable. A head makes no new stop, so the cause waits for a cut
  that settles 556 too.
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

- `semanticDiagnosticsPerFile`. Assigned at builder.ts:330 and 2168,
  `.set` at 420, `.clear` at 496, `.delete` at 894, each through
  `BuilderProgramState` (`createBuilderProgramState`'s state, cast
  `as BuilderProgramState` at 325, `removeSemanticDiagnosticsOf`'s parameter,
  and `createBuilderProgram`'s state, narrowed to
  `BuilderProgramStateWithDefinedProgram`). `.set` at 1068 goes through a
  local typed `BuilderProgramState["semanticDiagnosticsPerFile"]` (1056), first
  defaulted from `state` at 1058 with `??=`, a write to the local, not to the
  property. 2156 builds such a map fresh.
- `emitDiagnosticsPerFile`. Assigned at 497 and 1886, `??=` and
  `.set` at 405 and 1858, `.delete` at 782 and 2225, each through
  `BuilderProgramState` or `BuilderProgramStateWithDefinedProgram`. 1879 to
  1883 fill a local `Map<Path, Diagnostic[]>` before 1886 stores it.
- The view of both diagnostics maps is only read: at 403 and 418 through
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
- `changedFilesSet`. Assigned at builder.ts:334, `.add` at 355 and 489, through
  `createBuilderProgramState`'s state, a `BuilderProgramState`; `.delete` at
  667 through `getNextAffectedFile`'s `BuilderProgramStateWithDefinedProgram`;
  `.clear` at 1862 and 2170 through `createBuilderProgram`'s state, narrowed to
  `BuilderProgramStateWithDefinedProgram`. The view is only read: at 355,
  `oldState!.changedFilesSet?.forEach`, through
  `Readonly<ReusableBuilderProgramState>`, and in the tests at
  testRunner/unittests/tscWatch/incremental.ts:179 through
  `BuilderProgram.state`. It is built fresh at 2279 and 2317 from the local
  set made at 2267.
- `affectedFilesPendingEmit`. Assigned at 357 through
  `createBuilderProgramState`'s state and at 707 through
  `clearAffectedFilesPendingEmit`'s `BuilderProgramState`; `.delete` at 715
  and `.set` at 716 there too; `.delete` at 760 through
  `getNextAffectedFilePendingEmit`'s `BuilderProgramStateWithDefinedProgram`;
  `??=` and `.set` at 1856 and `.delete` at 1857 through `createBuilderProgram`'s
  narrowed state; `??=` and `.set` at 2224 through
  `addToAffectedFilesPendingEmit`'s `BuilderProgramState`. The view is only
  read, at 356 and 357 through `oldState`, and built fresh at 2318. The other
  hits share only the name: the local build info list at 1339 to 1371,
  `IncrementalBuildInfo`'s field (1149, tsbuildPublic.ts:1582) and the test
  helpers' readable build info (baseline.ts:188 to 283, tsc.ts:202 to 434).

## Public consequence

None. `ReusableBuilderProgramState` carries `@internal`, and the script
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

Sealed, each taking slots back out of `adapt.cjs` and scanning builder.ts's
closure against the adaptation as it is (J-PENDING).

| Mutant | Brings back |
|---|---:|
| Cause 4: `readonly` taken off all four `ReusableBuilderProgramState` fields | PENDING, the 9 |
| Cause 4: `readonly` taken off `changedFilesSet` alone | PENDING |
| Cause 4: `readonly` taken off `affectedFilesPendingEmit` alone | PENDING |

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.

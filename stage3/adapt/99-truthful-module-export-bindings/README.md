# Truthful module export bindings, adaptation 99-truthful-module-export-bindings

Adamic, task #k8ed1f2, family 3's cause 10, by Adamic's delegation: road 4,
a declaration made true. This is type-only: each edit is an annotation or a
type argument, so the emitted JavaScript does not change. Every declaration
it touches is local to a function, file-local, or `@internal`, so the
published declaration files do not change, this directory carries no
`api-sanctions.json`, and `stage3/lane/sanctioned-api.json` is not
regenerated.

## Why

Family 3 is the refusal "a value of type A seen as B, which can write X where
Y is read". Cause 10 is `ExternalModuleInfo.exportedBindings`, declared
`Identifier[][]` (transformers/utilities.ts:117), with the local that fills it
in `collectExternalModuleInfo` (184) and `collectExportedVariableInfo`'s
parameter (355) declared the same.

The declaration is narrower than what tsc stores. `multiMapSparseArrayAdd`
(377) is generic in its element, and four of its calls hand it a
`ModuleExportName`, which is `Identifier | StringLiteral`:

- 237, `export * as ns from "mod"`: the `NamespaceExport`'s `name`, a
  `ModuleExportName`, which is a string literal in `export * as "a b" from "mod"`;
- 322, `export { x as y }`: the specifier's exported `name`. The guard above it
  tests only the local side (`propertyName || name`) for a string literal, so
  `export { x as "a b" }` stores the `StringLiteral` `"a b"`;
- 338, `export { f as default }` for a function: `name ?? getDeclarationName(node)`,
  where `name` is the specifier's `ModuleExportName`, `"default"` as a string
  literal included;
- 348, `export { f as y }` for a function: `name ??= node.name!`, where `name`
  is again the specifier's `ModuleExportName`.

At each of the four, tsc infers the helper's element as `ModuleExportName` and
accepts `Identifier[][]` as a `ModuleExportName[][]`. That view is the lie: it
writes a `StringLiteral` into an array whose declared readers expect only
`Identifier`s. Adamic refuses each such view at its site. A readonly view
can't help, since tsc does write there; the declaration has to say what the
array holds.

## The edit

`adapt.cjs` edits 6 slots in 3 files, all in `src/compiler/transformers`. It
parses each file with stock TypeScript 6.0.3 and finds each slot by name,
never by line, offset or whole-line text:

- an interface's property, by the interface's and the property's names;
- a function's parameter, by the function's and the parameter's names;
- a binding by its name and its `const` or `let` in the function declaration
  nearest it, its initializer starting as upstream's does or, for a `let`,
  absent; or the one type argument of the `new Set<...>()` that initializes it;
- a call's one type argument, by its callee and its first argument's text, in
  the function declaration nearest it.

The edit replaces the slot's type node's span. Before any write it refuses
unless every slot is found exactly once, holds the text it expects or exactly
its adapted text, and its declaration is local to a function, file-local (a
top-level statement of a module, not exported), or carries `@internal` on
itself or an enclosing declaration. A slot already adapted keeps its bytes, so
a second run changes nothing. It prints one line per slot: its name, `edited`
or `already adapted`, where it is and why it is not public.

Lines are upstream 6.0.3's.

| Slot | Expected | Adapted | Visibility |
|---|---|---|---|
| utilities.ts `ExternalModuleInfo.exportedBindings` (117) | `Identifier[][]` | `ModuleExportName[][]` | `@internal` on the interface |
| utilities.ts `collectExternalModuleInfo`'s `const exportedBindings` (184) | `Identifier[][]` | `ModuleExportName[][]` | local to `collectExternalModuleInfo` |
| utilities.ts `collectExportedVariableInfo`, `exportedBindings` (355) | `Identifier[][]` | `ModuleExportName[][]` | file-local: not exported |
| module.ts `getExports`'s `const bindingsSet = new Set<...>()` (2469) | `Identifier` | `ModuleExportName` | local to `getExports` |
| module.ts `getExports`'s `arrayFrom<...>(bindingsSet)` (2481) | `Identifier`, from 33-indexed-reads-emit | `ModuleExportName` | a call in `getExports` |
| system.ts `getExports`'s `let exportedNames` (1987) | `Identifier[] \| undefined` | `ModuleExportName[] \| undefined` | local to `getExports` |

Each row is one slot, and `adapt.cjs` lists the same 6. Upstream writes no
type argument on the `arrayFrom` call; 33-indexed-reads-emit adds
`<Identifier>` there (its type-edits.json, class overload-contract), so the
slot expects that text and refuses a call with none rather than add one on its
own.

### Every reader followed

`exportedBindings` is read in two files besides its own, searched across all
of tsc's `src`. Each reader was widened only where tsc's own types demand it.

- utilities.ts. The interface, the local and the parameter widen together:
  the local is returned as the interface's field (301) and passed to the
  parameter (262, 359), so any one left `Identifier[][]` is a TS2322 or
  TS2345. `collectExportedVariableInfo`'s own add (369) stores `decl.name`, an
  `Identifier`, and the two class adds (278, 286) store `Identifier`s; all
  three still infer the helper's element from the widened array.
  `multiMapSparseArrayAdd` itself is generic and keeps its declaration.
- module.ts `getExports` (2460). It already returns
  `ModuleExportName[] | undefined`, so the read at 2464 needs nothing. The
  read at 2473 is an inferred local, `bindings`, that widens on its own; its
  elements go into `bindingsSet`, a `Set<Identifier>`, whose `add` would then
  take a `ModuleExportName` (TS2345), so its type argument widens, and
  `arrayFrom<Identifier>(bindingsSet)` would then be handed a
  `Set<ModuleExportName>` (TS2345), so its type argument widens too. The
  callers (885, 1157, 1906, 2438) see the same return type as before.
- system.ts `getExports` (1986). `addRange(exportedNames, ...)` at 1995
  infers its element from both arguments, so with the bindings widened it
  returns `ModuleExportName[] | undefined`, which the `let exportedNames:
  Identifier[] | undefined` can't hold (TS2322): the `let` widens. The
  `append` at 1992 adds `factory.getDeclarationName(...)`, an `Identifier`,
  which the wider array takes. The function writes no return type; tsc
  inferred `ModuleExportName[] | undefined` before, from the union with the
  other branch's `ModuleExportName[]`, and infers the same after. Its two
  callers (1707, 1959) hand each name to `createExportExpression`, whose
  `name` is `Identifier | StringLiteral`, so neither needs an `Identifier`
  specifically and nothing in the System transform takes a string-literal
  export name for an identifier.
- system.ts `getReferencedDeclaration` (2016, 2023) only tests the indexed
  bucket for truth. Nothing to follow.

No other module reads `exportedBindings` or `ExternalModuleInfo`: module.ts
and system.ts are the only importers of `collectExternalModuleInfo`.

## What it clears

The 7 census sites of cause 10, at main's tip, adapted tree lines, relative to
`src/compiler/transformers/`:

| Site | Seen as | Cleared by |
|---|---|---|
| utilities.ts:239:56 | `Identifier[][]` as `ModuleExportName[][]` | the local and the interface |
| utilities.ts:324:48 | the same | the same |
| utilities.ts:340:40 | the same | the same |
| utilities.ts:350:40 | the same | the same |
| module/module.ts:2464:24 | `Identifier[] \| undefined` as `ModuleExportName[] \| undefined` | the interface |
| module/module.ts:2481:28 | `Identifier[]` as `ModuleExportName[] \| undefined` | `bindingsSet` and `arrayFrom`'s type argument |
| module/system.ts:2005:16 | `Identifier[] \| undefined` as `ModuleExportName[] \| undefined` | the `let exportedNames` |

The four utilities.ts sites are upstream's 237, 322, 338 and 348, two lines
lower in the adapted tree. The system.ts site is the final `return
exportedNames` (upstream 2007).

## Left out

- `multiMapSparseArrayAdd`'s `map: V[][]`. It is generic, and with the array
  declared as what it holds, each call's `V` is the array's own element, so
  no view remains to narrow.
- Explicit type arguments on the seven `multiMapSparseArrayAdd` calls. Inference
  already picks `ModuleExportName` from the widened array; arguments would add
  bytes and clear nothing.
- 33-indexed-reads-emit's records. Its type-edits.json invariant for the
  `arrayFrom` call, and its proof README, say `exportedBindings`' elements are
  `Identifier`. After this adaptation that reads stale, and its type mutant
  that widens the argument measures a tree this one changes. This directory
  writes nothing outside itself, so both are left for 33's owner.

## Writes elsewhere

Every write found into `exportedBindings`, and what it stores. The search
covered all of tsc's `src`: every `exportedBindings` reference, every call of
`multiMapSparseArrayAdd`, and every write into the arrays the readers take out.
Lines are upstream 6.0.3's.

- `multiMapSparseArrayAdd` (377) is the only writer of the outer and inner
  arrays: `values.push(value)` at 380 and `map[key] = values = [value]` at
  383. Its seven calls on `exportedBindings` store:
  - 237: a `NamespaceExport`'s `name`, `ModuleExportName`;
  - 278: `context.factory.getDeclarationName(node)` for `export default class`,
    `Identifier`;
  - 286: a class's `name`, `Identifier`;
  - 322: an export specifier's `name`, `ModuleExportName`;
  - 338: `name ?? context.factory.getDeclarationName(node)`, `ModuleExportName`;
  - 348: `name`, after `name ??= node.name!`, `ModuleExportName`;
  - 369: a variable's `decl.name`, `Identifier`, reached from 262 through
    `collectExportedVariableInfo`.
- The object returned at 301 hands the same array to `ExternalModuleInfo`;
  module.ts and system.ts store it in `currentModuleInfo`, `moduleInfo` and
  `moduleInfoMap` and only index it.
- module.ts 2464 returns an inner array itself to `getExports`' callers, which
  only iterate it. 2476 copies each binding into `bindingsSet` with `add`, and
  2481 returns a fresh array from it.
- system.ts 1995 `addRange`s an inner array's elements into the local
  `exportedNames`, which `append` at 1992 may have made fresh; `addRange`
  pushes onto `exportedNames`, never onto the bucket.

So the arrays hold `StringLiteral`s in practice wherever an export name is a
string, and every reader treats each element as a `ModuleExportName`.

## Public consequence

None. `ExternalModuleInfo` carries `@internal`, `collectExportedVariableInfo`
is a top-level function utilities.ts does not export, and the other four
slots are local to `collectExternalModuleInfo` or to a `getExports`, which
sit inside `@internal` declarations (`transformModule`,
`transformSystemModule`). The script refuses any slot that is none of these.
Types and type arguments erase, so no line of the emitted JavaScript and no
line of typescript.d.ts moves. This directory carries no `api-sanctions.json`,
and `stage3/lane/sanctioned-api.json` is not regenerated.

## Order

apply runs `sorted(...)` over `stage3/adapt`'s directories. These edit the
three files before this one, so it reads their output:

- 32-indexed-reads-program: its handoff guards name utilities.ts's
  `getAllDecoratorsOfClass`, `getAllDecoratorsOfAccessors` and
  `getAllDecoratorsOfMethod`;
- 33-indexed-reads-emit: module.ts's `currentModuleInfo` and system.ts's
  `moduleInfo`, `exportFunction` and `contextObject` become
  `| undefined`, `arrayFrom<Identifier>(bindingsSet)`, and its indexed-read
  sites in all three;
- 71-writable-views: the `location` parameters of four module.ts and three
  system.ts functions;
- 99-correctness-receiver-transformers: pins and edits module.ts's import and
  `transformModule`, and system.ts's `transformSystemModule`, by the sha256 of
  their whole text, which holds both `getExports`;
- 99-truthful-generated-name-emit-node: one `autoGenerate` read and the
  imports in each of utilities.ts and module.ts.

The program-wide ones (10-type-imports, 20-optional-declarations,
70-readonly-views, 75-optional-widening, 99-readonly-inferred-bindings and
others) touch none of the six slots' text. The name sorts after
99-truthful-generated-name-emit-node, the last of them, so each pin reads its
declarations as it expects. The one that sorts after it, 99-wasm-crypto-import,
edits only sys.ts.

## Proof

Sealed, at head 590b2bb0 over main 29de95a7, stock TypeScript 6.0.3.

Apply and stock tsc (J-2tk3wyj): apply builds the adapted tree with and
without this directory. A second run of `adapt.cjs` prints all 6 slots as
already adapted and changes no file. Exactly 3 files of `src/compiler` differ:
transformers/utilities.ts, transformers/module/module.ts and
transformers/module/system.ts. Stock tsc over tsc's whole program reports 0
diagnostics on both trees. Emitting src/compiler with and without, 158 files,
every JavaScript file is byte-identical. The only declaration that differs is
transformers/utilities.d.ts with `@internal` kept: `exportedBindings:
Identifier[][]` becomes `ModuleExportName[][]`. ExternalModuleInfo is
`@internal`, so the published declarations strip it, and `stage3/lane`, which
holds typescript.d.ts to sanctioned-api.json, passes.

The census (J-gm33920 without, J-gas25vh with): scan-all's instrumented probe
over the closures of the 3 touched files. The with side ran three times: the
first two kept logs lost their tails (J-wanzvc3, J-s70fp1g, #77ncr00), and
J-gas25vh carries the whole list gzipped with its count and digest.

| Measurement | Without | With |
|---|---:|---:|
| Stock tsc diagnostics | 0 | 0 |
| Distinct stops over the closures | 1,963 | 1,956 |
| Sites that stop no more | | 7 |
| Sites that stop for a different reason | | 0 |
| Sites that stop and didn't before | | 0 |

The 7 are transformers/utilities.ts:239:56, 324:48, 340:40 and 350:40,
transformers/module/module.ts:2464:24 and 2481:28, and
transformers/module/system.ts:2005:16.

## Mutants

Sealed, against the adaptation as it is.

| Mutant | Brings back |
|---|---|
| The three utilities.ts declarations left `Identifier[][]`, the readers widened (J-fca62nv) | 5 of the 7: utilities.ts's four and module.ts:2464; module.ts:2481 and system.ts:2005 stay cleared, which the readers' own edits hold |

The readers' edits weren't taken out alone: by reading, with the declarations
widened stock tsc would refuse module.ts's `bindingsSet.add(binding)` and system.ts's
`addRange`, so the readers are what makes the widening type-check, and the
apply-and-tsc proof above is their test.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.

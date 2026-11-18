# Readonly reader views, adaptation 86

Adamic, task #38kdd4d (family 4's readonly views), ruled Nov 18 in adaptation
72's shape. Type-only: the emitted JavaScript doesn't change, and every edited
declaration is internal, so no line of the published `typescript.d.ts` moves.

## The edits

Each makes a declaration say `readonly` where a value whose own field is
readonly is handed to it, and where tsc only reads that field. Every one is
found by its owner and member and its exact upstream text, refused before any
write if it has moved, and left as it is on a second run. Nothing is written
until every edit in every file has been found.

| File | Owner | Becomes |
|---|---|---|
| moduleSpecifiers.ts | getModuleSpecifierPreferences' `importingSourceFile` | `Readonly<Pick<SourceFile, "fileName" \| "impliedNodeFormat">>` |
| moduleSpecifiers.ts | getDefaultResolutionModeForFile's `file` | `Readonly<Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">>` |
| program.ts | getDefaultResolutionModeForFileWorker's `sourceFile` | the same `Readonly<Pick<...>>` |
| program.ts | getImpliedNodeFormatForEmitWorker's `sourceFile` | the same `Readonly<Pick<...>>` |
| builderState.ts | canReuseOldState's `oldState` | `Readonly<BuilderState> \| undefined` |
| types.ts | PrivateIdentifier{Property,Method,GetAccessor,SetAccessor}Declaration | `readonly name: PrivateIdentifier` |
| types.ts | Program's `redirectTargetsMap` | `readonly` |
| types.ts | Program's `getCanonicalFileName` | `readonly` |
| types.ts | SourceFileMayBeEmittedHost's `getCanonicalFileName` | `readonly` |

## The premises, by root

- **The SourceFile picks.** A `FutureSourceFile` declares `fileName`
  readonly. The receivers only read `fileName`, `impliedNodeFormat` and
  `packageJsonScope`, and the two program.ts workers are where
  getDefaultResolutionModeForFile hands its file on.
- **The private names.** esDecorators.ts hands its create*DescriptorObject
  functions to a callback that takes `{ readonly name: PrivateIdentifier }`.
  The four interfaces redeclared `name` without readonly, and the callees only
  read it.
- **canReuseOldState** reads only `oldState.referencedMap`.
- **Program.** checker.ts reads a TypeCheckerHost as a Program
  (`(host as Program).getCommonSourceDirectory`). The host declares
  `redirectTargetsMap` readonly and `getCanonicalFileName` a method, and
  Program's own copies are never assigned (program.ts assigns a local).
- **SourceFileMayBeEmittedHost.** It's handed an EmitHost, whose
  `getCanonicalFileName` is a method, and sourceFileMayBeEmitted only reads it.

## Not here

Family 4's other causes stay with their own tasks:
- `emitNode` (58 sites) waits on its writers' certificate.
- `typeExpression` (10) is a construction change, and one public
  `updateJSDocThisTag` hazard stays refused.
- `members` (3) is adaptation 77's.
- `flags` and `kind` (98) are the `Mutable<X>` table.
- FlowNode's `antecedent` belongs to #hjkfhws.
- `DiagnosticWithLocation.file` (1 site) would be a public line, so it goes to
  Adamic on its own.

## Measured

MEASUREMENTS

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.

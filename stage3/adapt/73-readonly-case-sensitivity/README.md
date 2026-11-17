# Readonly case sensitivity, adaptation 73

Adamic, task #hes7595. This is type-only: the emitted JavaScript does not change.
It follows adaptation 72's shape: a readonly slot certified by the compiled
program's gate, with one public member sanctioned.

## The edit

`adapt.cjs` inserts `readonly ` before one member in each of two interfaces:

```ts
// src/compiler/types.ts, public
export interface ModuleResolutionHost {
    ...
    readonly useCaseSensitiveFileNames?: boolean | (() => boolean) | undefined;
}

// src/compiler/watchUtilities.ts, @internal
export interface WatchFactoryHost {
    ...
    readonly useCaseSensitiveFileNames: boolean | (() => boolean);
}
```

It parses each file with stock TypeScript 6.0.3 and refuses before any write
unless it finds exactly one interface of that name with exactly one property
signature `useCaseSensitiveFileNames`, carrying upstream's type and upstream's
optionality. A slot already readonly keeps its bytes, so a second run changes
nothing. No source coordinate is a rewrite address.

CompilerHost, ResolutionCacheHost, the solution builder hosts and the watch
hosts declare `useCaseSensitiveFileNames(): boolean` as a method. Seen as either
view, that method was a writable field, so a write of `boolean | (() => boolean)`
through the view could replace it. Adamic refused every such view, 25 in tsc:
24 into ModuleResolutionHost (program.ts 12, resolutionCache.ts 9,
tsbuildPublic.ts 3) and 1 into `WatchFactoryHost & { trace?(s: string): void; }`
(tsbuildPublic.ts, the host handed to createWatchFactory). With the slot
readonly, each view is read-only, and nothing a reader does through it can
replace the method.

## Certificate

Certified by the compiled program's gate, as adaptation 72 is. Adamic compiles
`src/compiler` and judges every view at its relation site, so a write through a
wider view inside the compiled program is refused there, not trusted.

Two counts over all 79 files of `src/compiler`, `src/executeCommandLine` and
`src/tsc`, read on upstream 6.0.3, are what make the edit true:

- **Writes to `useCaseSensitiveFileNames` through any view: 0.** No assignment,
  compound assignment, `delete` or bracket write. Every
  `useCaseSensitiveFileNames =` in the tree declares a local. services.ts,
  services/types.ts, services/utilities.ts, server/project.ts,
  editorServices.ts, session.ts and typingsInstaller.ts write it nowhere either.
- **Reads that call the value without its receiver: 0.** ModuleResolutionHost's
  member is read in two functions, arePathsEqual and
  `useCaseSensitiveFileNames(state)` in moduleNameResolver.ts. WatchFactoryHost's
  is read once, in getWatchFactory's local `useCaseSensitiveFileNames()` in
  watchUtilities.ts. Each reads it behind a `typeof` test and calls it as
  `host.useCaseSensitiveFileNames()`, through its receiver.

**Every reader calls through its receiver** is the premise this certificate
stands on. A reader that detached the member and called it bare would call a
method without the `this` its host gives it, which readonly does not make sound.
If such a reader appears, the certificate falls for that reader.

## Public consequence

`ModuleResolutionHost` is public, so the published `typescript.d.ts` gains
`readonly` on `useCaseSensitiveFileNames`. Code that assigns
`useCaseSensitiveFileNames` through a ModuleResolutionHost view stops
type-checking (TS2540); code that reads it is unaffected, and a host that
declares it as a method keeps its method. `WatchFactoryHost` is `@internal`
and changes no public declaration.

`api-sanctions.json` names the one public member, replaced whole
(`"replace": "member"`), for `stage3/lane/generate-api.cjs`. The regenerated
`stage3/lane/sanctioned-api.json` gains one normalized change,
`member:["ts","ModuleResolutionHost","useCaseSensitiveFileNames"]#0`, the
original property signature with `ReadonlyKeyword` added and nothing else.
`shard-plan.json`'s runtime hash moves with the sanction, and
`stage3/lane/test_check.py` counts one more sanctioned API change.

## Premises a reader can re-check

Readonly is a checker fact. Two roads could write the slot without the checker
seeing an assignment to it, and both are closed today by a stop in front of
them, not by this adaptation. They are adaptation 72's two premises:

- `Object.assign`. A NotYet stands in front of it today. Lifting it must judge
  readonly on the target's properties; that is filed as #w9ysnpk.
- A generic key write `t[k] = v` with `K extends keyof T`. It is NotYet today.
  Whoever lifts it judges readonly per property of K's constraint, under the
  ruling on writes through a key that isn't a literal (docs/rulings.md:290).

If either stop is lifted without that judgment, this certificate no longer
holds for the slot.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.

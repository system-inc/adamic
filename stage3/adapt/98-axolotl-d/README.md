# 98d: moduleNameResolver, resolutionCache and sourcemap to zero own-file diagnostics

Axolotl (#axolotl), task #7xw984y. This adaptation does three kinds of thing:

- **A required read and a required value:** each `!`, erased on Node.
- **Truthful optional declarations:** six members, 20's model.
- **One read of caller data instead of two:** in sourcemap.ts's appendSourceMap. This is the one change to emitted behavior.

Two public typescript.d.ts lines move, the two `packageId` lines Adamic declared.

Measured the way the daily meter counts, by stage3/meter/report.py's own `report()` over the census tool's
output, on main `470e16f8`:

- moduleNameResolver.ts goes from 3 to 0, resolutionCache.ts from 1 to 0, and sourcemap.ts from 1 to 0.
- No other file's own-file count changes.
- Across every census record, exactly those five diagnostic lines go and none is added: 114 lines before, 109 after.

The upstream source is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Lines below are that
pinned source's lines; they document sites and are never edit addresses.

## The adapter

`adapt.cjs` parses current text with stock TypeScript 6.0.3 from stage3/api. Each site is found by its owner
(interface or function), its member or exact expression text, and its count there, and any other count is
refused. Every site must be in its original shape or its adapted shape, and all in the same one. A half-adapted
tree, or a shape it hasn't inventoried, is refused before any write. It also checks the facts the required
value rests on: the producer's four callers and the argument each passes, the worker's tryResolve value type,
and that the parameter is never reassigned. It plans every edit in all five files, checks no file changed while
planning, then writes. Every file it edits is CRLF, and the inserted sourcemap lines use the file's own newline.

`stage3/apply.sh` passes end to end with 98 included. Its patch-set row is 5 files, 13 lines added, 10 removed.
A second run on the applied tree makes 0 edits, and all 711 source files keep the same bytes.

## The sites

[sites.json](sites.json) records each one.

| Row | Site | Edit | Why |
| --- | --- | --- | --- |
| 1 | moduleNameResolver 474, getPackageJsonTypesVersionsPaths, `paths: typesVersions[key]` | `typesVersions[key]!` | U-table. The read is under `for (const key in typesVersions)`, whose first statement is `if (!hasProperty(typesVersions, key)) continue;`, so key is an own key of the record, whose values are present. |
| 2 | moduleNameResolver 286, the `resolvedModule` literal's `isExternalLibraryImport` | `isExternalLibraryImport: isExternalLibraryImport!` | U-parallel, below. |
| 2 | types.ts `ResolvedModuleFull.originalPath`, `.packageId`; `ResolvedTypeReferenceDirective.originalPath`, `.packageId` | `?: T \| undefined`, readonly kept | Present and undefined, below. |
| 2 | program.ts `ResolutionWithResolvedFileName.packageId` (@internal) | `?: PackageId \| undefined` | Present and undefined, below. |
| 3 | resolutionCache.ts `ResolutionWithFailedLookupLocations.files` (@internal) | `?: Set<Path> \| undefined` | stopWatchFailedLookupLocationOfResolution stores `resolution.files = undefined;` (1386). |
| 4 | sourcemap.ts 215 to 217, appendSourceMap | one read into `content` | JavaScript-changing, below. |

The ruling's `(T) | undefined` is written `T | undefined`. 20 parenthesizes only function, constructor and
conditional types, and none of these is one.

### Row 2: what each member needs

The literal at 286 (TS2322) and the one at 627 (TS2375) each fail under exactOptionalPropertyTypes on the first
member that holds undefined. With `adamic types` on scratch copies:

- **Each half alone:** widening only the two originalPath members leaves both naming packageId. Widening only the two packageId members leaves both naming originalPath.
- **The four ruled members:** 627 closes, and 286 then names `isExternalLibraryImport`.
- **A sibling contract:** the widened objects are no longer assignable to the internal `ResolutionWithResolvedFileName`, so nine diagnostics appear in program.ts and resolutionCache.ts. That interface's packageId is widened too.

The evidence for that probe is on `compiler/axolotl-98d-wip`.

**Present and undefined.** withPackageId (127) builds every `Resolved` as
`{ path, extension, packageId, resolvedUsingTsExtension }`. It always writes the key, and it is undefined unless
package.json has a string name and version. noPackageId (143) passes no package info, so its packageId is
always undefined. Then:

- The literal at 286 always writes `packageId: resolved.packageId` and
  `originalPath: resolved.originalPath === true ? undefined : resolved.originalPath`.
- resolveTypeReferenceDirective always writes `packageId` destructured from a `Resolved`, and `originalPath` from
  `let originalPath: string | undefined`. That stays undefined under preserveSymlinks, and
  getOriginalAndResolvedFileName (516) returns `originalPath: pathsAreEqual ? undefined : fileName`.
- `ResolutionWithResolvedFileName` holds those same objects. getResolvedModuleFromResolution and
  getResolvedTypeReferenceDirectiveFromResolution (utilities.ts 824, 829) return `resolution.resolvedModule`
  and `resolution.resolvedTypeReferenceDirective` unchanged.

Keys, writes and runtime are untouched.

**The required value.** The parameter is `isExternalLibraryImport: boolean | undefined`, but the literal is built only
inside `resolved && { ... }`, and the adapter checks the parameter is never reassigned. The producer has exactly
four calls, all in moduleNameResolver.ts:

| Caller | resolved | isExternalLibraryImport |
| --- | --- | --- |
| nodeModuleNameResolverWorker (1901) | `result?.value?.resolved` | `result?.value?.isExternalLibraryImport` |
| classicNameResolver (3286) | `resolved && resolved.value` | `resolved?.value && pathContainsNodeModules(resolved.value.path)` |
| loadModuleFromGlobalCache (3388) | `resolved` | `true` |
| createResolvedModuleWithFailedLookupLocationsHandlingSymlink (247) | `resolved` | its own parameter, passed through |

Each caller pairs a defined `resolved` with a boolean:

- **The worker:** `result` is assigned only from its own tryResolve or `undefined` (1851, 1856). tryResolve
  returns `SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>`, with every value built as
  `pathContainsNodeModules(...)`, `false`, `true` or `contains(...)`, all boolean. So a defined
  `result.value.resolved` comes with a boolean.
- **The classic resolver:** a defined `resolved.value` is a `Resolved` object, so the `&&` yields
  pathContainsNodeModules's boolean.
- **The global cache:** it passes a literal `true`.
- **HandlingSymlink:** it is reached only from the first two. It replaces `resolved` only inside
  `if (... resolved && ...)`, with `{ ...resolved, path, originalPath }`, so it never makes `resolved`
  defined from undefined.

Upstream's tryResolve value type is not readonly. An earlier adaptation marks it readonly, so the adapter checks
the members structurally: both required, `resolved: Resolved` and `isExternalLibraryImport: boolean`.

### Row 4: one read of caller data instead of two

```ts
if (map.sourcesContent && typeof map.sourcesContent[raw.sourceIndex] === "string") {
    setSourceContent(newSourceIndex, map.sourcesContent[raw.sourceIndex]);
}
```

becomes

```ts
if (map.sourcesContent) {
    const content = map.sourcesContent[raw.sourceIndex];
    if (typeof content === "string") {
        setSourceContent(newSourceIndex, content);
    }
}
```

The guard's read of `map.sourcesContent` stays first and as written. The element is read once, tested, and
that value is stored, so the TS2345 goes without an assertion and without widening `RawSourceMap`.

## JavaScript-changing

**98 intentionally changes emitted JavaScript at one site.** appendSourceMap now reads
`map.sourcesContent` twice instead of three times, and reads the element once instead of twice.
This affects only a caller whose map changes between those reads, such as a getter. For it, the value stored
is now the value the guard tested. appendSourceMap is @internal (not in typescript.d.ts), but
`ts.createSourceMapGenerator` is on the runtime `ts` object of the published typescript.js, which is how the
witness reaches it.

[witness.cjs](witness.cjs) runs appendSourceMap from each actual built compiler on two maps, a plain
`sourcesContent: ['present']` and 33's getter (`['present']` for the first two property accesses, `[]` after),
both with one mapping to source 0. Its output is [witness-output.json](witness-output.json):

| Compiler | Plain array | Getter: reads | Getter: merged sourcesContent |
| --- | --- | ---: | --- |
| Without 98 | `["present"]` | 3 | `[null]` |
| With 98 | `["present"]` | 2 | `["present"]` |

The plain array agrees. With the getter, the original's guard passes on the second read, and its payload, the
third, is undefined, which serializes as `null`. The adapted function reads twice, the guard and the element,
and stores the string it tested. Stock npm typescript@6.0.3 gives the original's row too.

## Transpile and built JavaScript

`ts.transpileModule` of the five edited files, without and with 98:

- types.ts, program.ts and resolutionCache.ts are byte-identical.
- sourcemap.ts differs only at the site above.
- moduleNameResolver.ts differs at one line, 286's property. A shorthand property can't take `!`, so
  `isExternalLibraryImport,` is spelled out as `isExternalLibraryImport: isExternalLibraryImport!`, which
  transpiles to `isExternalLibraryImport: isExternalLibraryImport,`. Row 1's `!` erases to the same bytes.

The 286 spelling builds the same object: same key, same value, same evaluation order. But the transpile output
is not byte-identical there. The bundle tsc ships is byte-identical there: in stock `npm run build`, esbuild
prints the shorthand for both, and of the ten built/local JavaScript files only `_tsc.js`, `run.js` and
`typescript.js` differ, each only at appendSourceMap.

## Public API

[api-sanctions.json](api-sanctions.json) has the two declared rows, named by generate-api.cjs's `ownerPath`,
which drops the `ts` namespace: `["ResolvedModuleFull", "packageId"]` and
`["ResolvedTypeReferenceDirective", "packageId"]`, each `PackageId` to `PackageId | undefined`.
[proofs/api-lines.md](proofs/api-lines.md) has the built typescript.d.ts diff: exactly those two lines.

The @internal members need no row, since they never reach typescript.d.ts: both originalPath members,
`ResolutionWithResolvedFileName.packageId` and `ResolutionWithFailedLookupLocations.files`. stage3/lane is
untouched.

## Mutants

Each was made from a copy of the adapted tree and measured by the same census and `report()`:

- **Row 1's `!` dropped:** moduleNameResolver.ts goes from 0 to 1, and nothing else changes.
- **files unwidened:** resolutionCache.ts goes from 0 to 1, and nothing else changes.

The adapter refuses a tree with only `files` reverted ("partly adapted"), and one whose 286 entry is
`isExternalLibraryImport ?? false` ("uninventoried").

## Evidence

The evidence is on `evidence/compiler/axolotl-98d`:

- the own-file tables before and after, over every file, and the line-level diagnostic diff
- the mutant tables
- the typescript.d.ts diff and every built JavaScript diff
- both stock build logs
- the transpile report
- the witness output
- both apply logs
- the idempotence hashes
- the two tools used: own-file.py, which calls stage3/meter/report.py, and transpile.cjs

**Lane:** run as part of Correctness's wave, which carries the manifest regeneration.

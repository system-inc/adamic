# Readonly path components in watch helpers

Family 1 C on #bbssdcy. The base is Adamic `934baf817306b372408af94a1026cf0249669242` and TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Locations are in this adapter's `-before` snapshot of `src/compiler/resolutionCache.ts`.

## Compiler acceptance

The adapter retypes four parameters as `readonly Path[]`. Each keeps the existing branded element type and the same array value. Stock TypeScript locates the three owners and their exact parameter snapshots.

| Parameter | Before location | Use |
| --- | --- | --- |
| `perceivedOsRootLengthForWatching.pathComponents` | `271:43` | Reads indexed path parts when determining the root length. |
| `canWatchDirectoryOrFile.pathComponents` | `312:41` | Reads the default length and delegates to the root helper. |
| `isInDirectoryPath.dirComponents` | `332:28` | Reads the prefix length and compares indexed parts. |
| `isInDirectoryPath.fileOrDirComponents` | `332:73` | Reads length and the matching indexed parts within the prefix length. |

`canWatchDirectoryOrFile` is an `@internal` export. The other two helpers are private. An independent owner check parses the planned output and asserts all four annotations before it writes the file, including on an idempotent rerun. `docs/rulings.md:32` governs brand observations, which these helpers never make. Row 33 governs brand mints, and this adapter adds none. Row 63 governs adapting strict source, row 317 describes the numeric index-signature refusal these sites hit, and row 408 illustrates the stripInternal declaration and JavaScript identity evidence for an established shape. The export-binding check refuses direct export forms and the call-only rule refuses every other reference to an owner, and `test.cjs` checks both against the adapter's before and after snapshots. Compiler acceptance requires a clean stock tsc run, identical JavaScript and public declarations, and Loom's unit census at main.
The call-only rule is sound because no owner has type parameters or overloads and each returns a `number` or a `boolean`. `owners.json` pins each whole declaration against the adapter's before and after snapshots. The adapter requires exactly one declaration statement per owner name and refuses a signature or body edit by an earlier adapter.

The cross-file scan reads every non-declaration `.ts` file under `src/`. The adapter refuses any `.mts`, `.cts` or `.tsx` file by name.

A string-keyed read in another file, such as `ts["canWatchDirectoryOrFile"]`, is outside the identifier rule. The owner is `@internal` and stripped, so the read's declaration names `typeof ts.canWatchDirectoryOrFile` and no public line moves (CLI's stock tsc 6.0.3 `stripInternal` probe on #b81yh6t at 2026-11-26 21:16:50Z). The two private owners cannot be read through a string key because neither is exported.

## Sites

| Before site | Unit | Disposition |
| --- | --- | --- |
| `271:43` | `perceivedOsRootLengthForWatching` | Targeted by this head. |
| `312:41` | `canWatchDirectoryOrFile` and `canWatchDirectoryOrFilePath` | Targeted by this head. |
| `332:28` | `isInDirectoryPath` | Targeted by this head. |
| `355:5` | `getDirectoryToWatchFailedLookupLocation` | Second head after the round 2 trial's `any` refusal at line 451. |
| `404:5` | `getDirectoryToWatchFromFailedLookupLocationDirectory` | Second head after the round 2 trial's `any` refusal at line 451. |
| `446:5` | `getDirectoryOfFailedLookupWatch` | Second head after the round 2 trial's `any` refusal at line 451. |
| `474:5` | `getDirectoryToWatchFailedLookupLocationFromTypeRoot` | Second head after the round 2 trial's `any` refusal at line 451. |

Adamic's Nov 24 16:19:55Z message asks Seats to bring her the open decision at `resolutionCache.ts:277:31` with this head.

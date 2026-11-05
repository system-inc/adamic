# 99m compiler census

The unmodified ordinary main census tool measured all 79 compiler roots independently.
The complete source snapshot is byte-identical to the full apply output; comparison.json records that check.
No overlays, refusal suppression, partial-root fallback, or backend changes were used.

The comparison baseline is 99l's complete retained census. The current-main baseline probes reproduced core.ts:409:37.
The interrupted redundant baseline rerun is retained as a partial probe, not represented as a complete census.

| Roots | Kind | First stop | Reason |
| ---: | --- | --- | --- |
| 73 | Refused | src/compiler/core.ts:726:12 | `a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold` |
| 2 | accepted | - | - |
| 1 | NotYet | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 1 | Refused | src/compiler/debug.ts:167:31 | a cast the runtime can't check |

## The 73 moving roots

73 roots: Refused at `src/compiler/core.ts:726:12`: `a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold`.

The next shared stop is deduplicate's return expression at core.ts:726:12.
Its deduplicateRelational helper infers `NonNullable<T>[]` from `deduplicated.map(i => array[i]!)`.
That writable array reaches deduplicate's declared `T[]` return. This is an inferred helper-return problem,
rather than the append call specialization repaired here. No second adaptation was selected.
The core.ts root itself reaches debug.ts:167:31, an unchecked cast; root traversal order matters.

census-results.json contains the exact reason and before/after stop for every root.
The raw census also retains the ordinary tool's non-source entries and whole-program row.

The ordinary meter's extra whole-program row reports `lower: stage 0 compiles a program from one entry file, got 79`.
That is its single-entry lowering limit. It does not affect the 79 independent per-file results
or the passing 79-root Adamic types check. The raw row is retained.

The two accepted roots are corePublic.ts and hostErrors.ts.

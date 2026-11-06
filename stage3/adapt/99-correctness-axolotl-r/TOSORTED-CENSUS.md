# toSorted amendment census

Independent ordinary Load/Lower observations on all 79 roots. Groups of 20, 20, 20 and 19 were each committed and plainly pushed upon completion. Refused and NotYet are recorded outcomes, not native compilation successes.

## Main 37f8314b, original complete adapted tree

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | `core.ts:1049:34` | a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold |
| 2 | accepted | `-` | - |
| 1 | NotYet | `commandLineParser.ts:293:35` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | `diagnosticInformationMap.generated.ts:13:45` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |



## Main 73cdea837 compiler, rebased complete tree with the 99r amendment

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | `core.ts:1076:100` | a cast the runtime can't check |
| 2 | accepted | `-` | - |
| 1 | NotYet | `commandLineParser.ts:293:35` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | `diagnosticInformationMap.generated.ts:13:45` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |



## New 73-root group: exact source

There is no enclosing function: the refused cast is in the top-level elementAt initializer, outside either arrow's body. Its complete 11-line declaration is:

```ts
export const elementAt: <T>(array: readonly T[] | undefined, offset: number) => T | undefined = !!(Array.prototype as any).at
    ? (array, offset) => (array as any)?.at(offset)
    : (array, offset) => {
        if (array !== undefined) {
            offset = toOffset(array, offset);
            if (offset < array.length) {
                return array[offset];
            }
        }
        return undefined;
    };
```

`castRefusal` calls `castProof`. In internal/lower/cast_proof.go:196-197, the source-or-target TypeFlagsAny branch returns the Refused value constructed at line 188. The target of `Array.prototype as any` is any. Neither this initializer nor its casts were edited.

## Census method and provenance

census-batch.go.txt preserves the standard stage3/census/tool imports, run record and inspect function byte-for-byte at both compiler revisions. Only main is replaced through a one-file Go overlay to take 1 to 20 independent root paths and omit the standard tool's redundant aggregate row. No load, checker, lowering, preflight or backend code is overlaid. Each root still calls ordinary Load on that one file and ordinary Lower on its complete dependency graph. The after binder observation equals both standard-tool rows, excluding elapsed time. Group 3 exposed a driver ordering assertion: Python Path component order differs from Go string order for declarations.ts and declarations/diagnostics.ts. All 20 observations per side had completed. The driver was corrected to require the exact unique root set and associate each before/after row by its absolute root path. Those completed compiler measurements were retained without rerunning them; the original error log and recovery ledger are archived.

tosorted-census.json retains every paired root and exact outcome. tosorted-census-summary.json retains full groups and checkpoint commits. Four tar.gz archives retain raw rows, root lists, logs and internal SHA256SUMS; adjacent sha256 files verify each archive. The full main before apply is 73cdea837, while the comparison census explicitly retains the requested 37f8314b baseline. The restored-arm mutant is byte-identical to the fresh 73cdea837 before source tree and reproduces 1049:34 on that compiler.

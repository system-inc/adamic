# Declaration accounting

There are two named source sets: arrayFrom has two bodyless overload declarations and arrayToMap has four. The report uses these source ordinals rather than treating all signatures as independent sets. The set census short-circuits on the first refused overload.

| Set | Ordinal | Adapted source line | Base observation | Head conclusion |
|---|---:|---:|---|---|
| arrayFrom | 1 | core.ts:1334 | visible result refusal | complete set accepted before certified-call storage gate |
| arrayFrom | 2 | core.ts:1336 | not reached independently after overload 1 refused | complete set accepted before certified-call storage gate |
| arrayToMap | 1 | core.ts:1398 | passed before overload 2 refused | complete set accepted before certified-call storage gate |
| arrayToMap | 2 | core.ts:1400 | visible result refusal | complete set accepted before certified-call storage gate |
| arrayToMap | 3 | core.ts:1402 | not reached independently after overload 2 refused | complete set accepted before certified-call storage gate |
| arrayToMap | 4 | core.ts:1404 | not reached independently after overload 2 refused | complete set accepted before certified-call storage gate |

Evidence is positive reach, not absence: the complete first-error entry pass reaches the certified-storage gate for arrayFrom and arrayToMap in builder.ts. At head, overloadCertificateCall obtains its certificate through overloadCertified, which runs censusOverloads for the source implementation before its call gate. A declaration refusal would have been returned instead of the storage NotYet in that first-error unit. Full core.ts alone defers the generic declarations and does not prove this.

The visible base blockers removed are arrayFrom overload 1 and arrayToMap overload 2. Map overload 4 was masked by overload 2 at base; head censusInferRemainder provides its previously unwitnessed mapper-result binder before certificate B is considered. Do not count its baseline masking as an independently observed old failure, or claim that all identity signatures are new independent passes.

# 99r ordinary compiler next stops

The unchanged ordinary production compiler independently inspected all 79 recursive
TypeScript compiler roots on the accepted four-mint stack. Its source snapshot is
byte-identical to full apply across all 746 src files. No measurement overlay,
checker suppression, backend change or partial-root fallback is used here.

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | src/compiler/core.ts:736:36 | a cast the runtime can't check |
| 2 | accepted | src/compiler/corePublic.ts, src/compiler/hostErrors.ts | No stop |
| 1 | NotYet | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | src/compiler/debug.ts:243:59 | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |

The per-root results, exact fixes, locations and timings are retained in
census-results.json. Extra non-source entries and the final whole-program row
are excluded from the 79-root table but retained in the compressed raw stream.
Timings include concurrent proof work and are not performance comparisons.

The ordinary next-stop census stops at the first refusal. The separate existing
no-output full-core measurement records each mint at core.ts:736:36, :759:12,
:764:12 and :1039:12 as Refused with exactly `a cast the runtime can't check`.
The current castProof explicitly refuses the outer cast over an as-unknown
operand; Diagnostics' corresponding recognition is outside this commit.

The pre-existing relation at core.ts:1039:34 also remains:
`a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold`.
The complete core before/after measurement has the same 82 distinct findings,
with none added or removed. Its raw events and exact reasons are retained in
site-results.json and the two sites-*.jsonl.gz files. It is not a lowering-success
claim and is never substituted for the ordinary 79-root census above.

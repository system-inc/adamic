# Measured ordinary census complete

All 79 roots completed, 158 raw observations, duplicate non-timing fields agree. All 73 former require roots moved to debug.ts:525:37. Six roots unchanged; no new tuple.

| Count | Location | Kind | Exact reason |
| ---: | --- | --- | --- |
| 74 | debug.ts:525:37 | Refused | a cast the runtime can't check |
| 2 | - | accepted | - |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |

Whole-lowering roots: corePublic.ts, hostErrors.ts. Ordinary first-stop totals: NotYet 3, Refused 74, panics 0, accepted 2. Ordinary full latent counters are unavailable.

Guarded entry run is in progress with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, bound 1800 seconds. Final REPORT.md will label emitted finding totals and completion coverage. Root names and before/after tuples are in grouped-first-stops.json and comparison.json.

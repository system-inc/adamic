# Measurement progress

All 79 ordinary roots and both single-file full-mode runs completed. Guarded entry remains active inside its 1800-second bound.

| Roots | Location | Kind | Exact reason |
| ---: | --- | --- | --- |
| 73 | tracing.ts:96:19 | NotYet | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| 2 | - | accepted | - |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 1 | types.ts:4560:38 | Refused | a primitive brand member _isResolvedConfigFileName whose type is not void |

Guarded entry timing records: 1409. Last completed timing: /workspace/scratch/next-stops-adapted/src/compiler/scanner.ts:1870:5. No completed source record or finding totals emitted yet.

Full-mode lists: debug 22 findings; semver 57 findings. All findings and boundaries retained in the raw streams and ordered JSON lists.

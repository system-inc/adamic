# Ordinary census complete; guarded and single-file runs pending

79 roots, 158 raw observations, first observation counted.

| Roots | Location | Kind | Exact reason |
| ---: | --- | --- | --- |
| 73 | tracing.ts:110:22 | NotYet | a namespace value seen as an argument to a callee whose body can't be named |
| 2 | - | accepted | - |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 1 | sys.ts:1654:13 | Refused | a checked view with a callable contract |

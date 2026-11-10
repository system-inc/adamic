# Stage 3 meter 20261110T131808Z

The area label repeats the pinned main tree for reporter compatibility.

| Run | Main SHA | Whole program | Own file | Lowering | tsc entry diagnostics |
| --- | --- | ---: | ---: | ---: | ---: |
| Current | 85f2bd6c1a13371f6d5089aec10bd68a371a06b9 | 79/79 | 79/79 | 2/79 | 0 |
| 20261104T131717Z.w3cnHR | f87d7884807161162c98a2d0ef3b5339e16b315f | 79/79 | 79/79 | 2/79 | 0 |

Lowering, main: 2/79; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;AbstractKeyword&#x27;) | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;Error&#x27;) | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |

For artifact roots, lowering requires attempted units in every program source and no entry-census finding in the program. No native output is claimed.

Area reuses main's measurement, not a second measurement. main/ was copied to area/ before reporting and gzip; both labels use main-adapted. See compiler-mode.json measurement_sources.

Elapsed: 101m 42s.

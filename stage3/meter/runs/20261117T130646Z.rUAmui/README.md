# Stage 3 meter 20261117T130646Z

The area label repeats the pinned main tree for reporter compatibility.

| Run | Main SHA | Whole program | Own file | Lowering | tsc entry diagnostics |
| --- | --- | ---: | ---: | ---: | ---: |
| Current | 1d3efe7b92236a0bb94aba6bfd5652f6d0f6dfa6 | 79/79 | 79/79 | 2/79 | 0 |
| 20261116T134940Z.SUGhBh | 6a7597da6379938757258240580e5d344fd5bbde | 79/79 | 79/79 | 2/79 | 0 |

Lowering, main: 2/79; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check (read from the entry census; source order) |

For artifact roots, lowering requires attempted units in every program source, no entry-census finding in the program, and the whole-program passes of every unit measured. No native output is claimed.

Area reuses main's measurement, not a second measurement. main/ was copied to area/ before reporting and gzip; both labels use main-adapted. See compiler-mode.json measurement_sources.

Elapsed: 71m 4s.

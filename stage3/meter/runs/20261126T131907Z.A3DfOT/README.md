# Stage 3 meter 20261126T131907Z

The area label repeats the pinned main tree for reporter compatibility.

| Run | Main SHA | Whole program | Own file | Lowering | tsc entry diagnostics |
| --- | --- | ---: | ---: | ---: | ---: |
| Current | 46661dbefaaa66c3813e49a336c39411c95ca416 | 80/80 | 80/80 | 3/80 | 0 |
| 20261124T135011Z.i9DGOZ | 5a4dae1d0aa5928416c4f30359cef41fc934e424 | 80/80 | 80/80 | 3/80 | 0 |

Lowering, main: 3/80; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | NotYet src/compiler/binder.ts:405:43: a call whose argument for a never parameter isn&#x27;t a rulings.md:76 never read (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | NotYet src/compiler/binder.ts:405:43: a call whose argument for a never parameter isn&#x27;t a rulings.md:76 never read (read from the entry census; source order) |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | NotYet src/compiler/binder.ts:405:43: a call whose argument for a never parameter isn&#x27;t a rulings.md:76 never read (read from the entry census; source order) |

For artifact roots, lowering requires attempted units in every program source, no entry-census finding in the program, and the whole-program passes of every unit measured. No native output is claimed.

Area reuses main's measurement, not a second measurement. main/ was copied to area/ before reporting and gzip; both labels use main-adapted. See compiler-mode.json measurement_sources.

Elapsed: 132m 11s.

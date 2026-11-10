# Debugger-display proof

Baseline main 7910ab94a; current main d4207a858 has identical TypeScript
adaptation, source pin, API and oracle inputs. The ordinary compiler census
is rebuilt against current main, without a lowering overlay.

- Pin audit: amendment required in 99t, whose active whole-owner enableDebugInfo pin overlaps.
- Whole lane: 398 accepted units; guarded counts 0 / 0 / 0. Positive control 6801 / 45429 / 1; ordinary control 0 / 0 / 0. See debug-display-count.json.
- Witness: 200 byte-identical paired CLI cases. Counted debug replays fire 1312455 / 9088154 / 200, with identical output. See debug-display-witness.json.
- Adapter: contract run exits 0, including strict owner refusals and replay with zero writes. A preliminary emission-walker scoping error was corrected before the passing run.
- Mutant: the one restored cast line reproduces Refused at debug.ts:525:37, "a cast the runtime can't check". Both binder.ts and core.ts move to debug.ts:1021:29 without it: FlowGraphNode.edges can write FlowGraphEdge[] where the literal inferred never[]. All three census commands exit 0; their Refused results are retained in debug-display-census.json.
- Candidate normal build exits 0. Only debug.ts changes in source. Both public declaration files are byte-identical. Retained formatter fixtures pass and their outputs are byte-identical.
- STOPPED: the normal bundle comparison exits 1. _tsc.js, typescript.js and run.js have emitted differences outside the five owners. In _tsc.js esbuild removes three now-unused helpers and changes bindings throughout the bundle. The raw outside-owner patch is retained compressed. The stock per-file verbatim-import proof passes, but does not satisfy this normal-bundle gate. No normalization or widened whitelist is accepted.
- The amended full apply is not run after this failing gate. The baseline full apply passed. This candidate is not ready to land; debug-display-stop.json records the blocker.

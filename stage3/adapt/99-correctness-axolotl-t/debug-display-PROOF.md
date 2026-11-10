# Revised bundler gate result

## Follow-up: complete bundle hunk classification

The revised gate permits mechanical bundler fallout. Every outside-owner hunk
is classified in debug-display-bundle-classification.json. The complete normal
before/after emitted patch, including the five owner edits, is retained in
debug-display-all-emitted-javascript-hunks.patch.gz.

| Bundle | (a) pure rename | (b) removed declaration | (c) stop |
| --- | ---: | ---: | ---: |
| _tsc.js | 772 | 2 | 0 |
| typescript.js | 1185 | 0 | 0 |
| run.js | 1247 | 0 | 0 |
| Total | 3204 | 2 | 0 |

The analyzer first replaces exactly the five owner texts with their after
texts, then inventories every residual hunk with three context lines. For (a),
only terminal decimal suffixes on AST Identifier nodes are normalized. Strings,
comments, numeric literals, punctuation and whitespace are untouched. Every
class-(a) hunk compares byte-identically after that transformation; both
normalized hashes are retained for each hunk. A mixed rename/deletion hunk
receives only one class, not two.

Only _tsc.js drops declarations. The reference list below is exhaustive and
resolved by TypeScript symbol identity in the original unadapted bundle;
declaration names themselves are excluded from the reference count.

| Removed function | Every original reference | Enclosing function | Inside a reviewed owner? |
| --- | --- | --- | --- |
| getJSDocModifierFlagsNoCache | _tsc.js:17073:51 | getEffectiveModifierFlagsNoCache | No |
| getEffectiveModifierFlagsNoCache | _tsc.js:1791:42 | enableDebugInfo | Yes |
| isArrayTypeNode | _tsc.js:1775:1231 | enableDebugInfo | Yes |

Both deletion hunks are (b), using the transitive closure authorized in the
follow-up. A removed declaration is class (b) when every reference to it is
inside a reviewed owner or inside another removed declaration that is itself
class (b). The classifier computes the least fixed point and records reference
edges and full chains; unanchored cycles fail closed.

Hunk 55 records getJSDocModifierFlagsNoCache ->
getEffectiveModifierFlagsNoCache -> enableDebugInfo. The direct reference at
17073 is inside the removed modifier helper, whose reference at 1791 is inside
enableDebugInfo. isArrayTypeNode's sole reference is directly in enableDebugInfo.
All three bundles compare byte-identically after identifier normalization and
these three declaration deletions. There are zero class-(c) hunks.

The runtime export comparison passes on Node v24.19.0. Loading each real
normal typescript.js through Node import and require gives byte-identical
sorted Object.keys arrays: 2258 module keys and 2256 keys on the separate ts
object/CommonJS exports. No key is added or removed. The complete arrays and
hashes are in debug-display-runtime-export-keys.json; the executable check is
debug-display-export-keys.cjs.

Current result: classification and export gates pass. Full amended apply and
final lane are pending, so ready_to_land remains false for now. FlowGraphNode
at debug.ts:1021:29 is untouched. The source adaptation and pins are unchanged.

## Earlier measurement record

The raw-byte-only scope failure below is historical; the revised gate and
remaining class-(c) stop above supersede that acceptance condition.

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

# Revised bundler gate result

## Follow-up: complete bundle hunk classification

The revised gate permits mechanical bundler fallout. Every outside-owner hunk
is classified in debug-display-bundle-classification.json. The complete normal
before/after emitted patch, including the five owner edits, is retained in
debug-display-all-emitted-javascript-hunks.patch.gz.

| Bundle | (a) pure rename | (b) removed declaration | (c) stop |
| --- | ---: | ---: | ---: |
| _tsc.js | 772 | 1 | 1 |
| typescript.js | 1185 | 0 | 0 |
| run.js | 1247 | 0 | 0 |
| Total | 3204 | 1 | 1 |

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

The isArrayTypeNode deletion hunk is (b). The hunk deleting both modifier
helpers is (c): getJSDocModifierFlagsNoCache has a reference outside the five
owners, inside the also-removed getEffectiveModifierFlagsNoCache. Its removal
is a mechanically explained transitive cascade, but it does not satisfy the
specified requirement that every reference be directly inside one of the five
owners. No other functions or bindings disappear. Whole residual bundles are
byte-identical after identifier normalization plus exactly these three
declaration deletions; there is no unexplained remaining emitted delta.

The runtime export comparison passes on Node v24.19.0. Loading each real
normal typescript.js through Node import and require gives byte-identical
sorted Object.keys arrays: 2258 module keys and 2256 keys on the separate ts
object/CommonJS exports. No key is added or removed. The complete arrays and
hashes are in debug-display-runtime-export-keys.json; the executable check is
debug-display-export-keys.cjs.

Current result: STOP under the direct-reference condition of class (b).
ready_to_land remains false. The amended full apply and dependent lane runs
are not started after this explicit class-(c) stop. The FlowGraphNode widening
at debug.ts:1021:29 is untouched. This follow-up changes evidence and prose
only, not adapt.cjs or pins.json.

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

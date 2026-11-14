# Adapted source inventory audit

The current control has **747** files under `src`; the amended tree has **748**.
The only added name is `src/recordreplay.d.ts`. The old archived 746-file proofs
also predate `src/deprecatedCompat/controlFlowGraph.ts`. Do not interpret 746 →
748 as two files added by this amendment. There are 712 → 713 `.ts`/`.a` paths;
compiler-only inventories retain 82 files (79 TypeScript implementation roots).

| Reader / contract | Observed result on amended tree |
|---|---|
| 99-wasm-crypto-import `proofs/check.cjs`, original | Rejects added inventory; also rejects current 747-file control on fixed 746 guard |
| Same checker, repaired | Complete CLI passes against actual pre-99 snapshot; 748 hashed files; exact declaration and extra-file mutants reject |
| Same adapter `proofs/scoped.py` | Actual planning/inventory statements pass, complete 748-file source map; historical apply/build/lane calls omitted |
| 93-builtin-imports `TestBuiltinSourceScope` | Exported full source-scope function rejects added file |
| 92-tty-host-results `TestTTYSourceScope` | Exported full source-scope function rejects added file |
| 91-patcher-discard `proofs/scope.cjs` | Full source-scope CLI rejects added file |
| 99-correctness-core-compare `close-proof.py:20` | Fixed 746-count conjunct rejects both control 747 and amended 748 |
| tsc-entry `verify-references.py:42` | Same fixed-count rejection; enclosing historical proof driver not run |
| lane `artifact.payload_digest` | Pass: added declaration changes payload identity |
| apply `product_key` | Pass: changed adapter/proof input changes key; recorded key predates final evidence files |
| apply `save_product` / `restore_product` | Pass: private source-only product roundtrip retains all 748 files and exact added declaration |
| lane `view.make_view` | Pass: actual amended source view exposes all 748 files |
| speculative `server-run.sh` source_checks | Fresh complete identity passes (21,300 `.ts`/`.a` files across actual root); stale identity missing only declaration rejects |
| census tool / latent filesystem collectors | Original WalkDir clauses pass, include added declaration (713 source paths); tool lists all 748 files |
| Compiler-only inventories (parser, lint, lower census helpers, axolotl call census) | Name inventory unchanged; added declaration is outside src/compiler; only sys.ts bytes differ |

Only the amendment's own 99 records/code were repaired in `e731288ba`:
`check.cjs` now requires exactly the before names plus the canonical declaration,
hashes the entire tree, and reverses the erased edits before checking crypto
scope. `scoped.py` compares the complete reviewed source map, plans the erased
sys edits, and updates two-file/mutation expectations. New current records are
under its `proofs/recordreplay/inventory-*.json`; archived old 746-file evidence
is retained as history. No foreign adapter, lane, apply, or manifest was edited.
No adapted source bytes changed in this follow-up.

`catalog.json.gz` records the tracked-file search and candidate sites (239
candidate files). `reader-index.json` maps every candidate to its projections
or to static-only review. `js-readers.json.gz` and `python-readers.json.gz`
retain the individual original-function/comprehension observations and counts;
filenames may be uncompressed when below the archive threshold. Directory roots
are retargeted to measured control/amended trees. These are inventory projections,
not claims that historical whole proof pipelines passed. In particular,
`no direct src comprehension`, empty results and `context_required` mean no
executable projection, not success. The JS `audit-range-callers.cjs` generic walk
rejects both full src trees on an unrelated lib DOM call; its real compiler-only
caller has an unchanged name inventory. These results do not establish an
exhaustive execution of every historical reader; static-only candidates remain
explicitly listed rather than labeled passed. Running those whole drivers would
mutate historical fixtures or launch additional builds/censuses outside the
reduced request. That is the remaining audit coverage limitation.

`production.json` includes an earlier invalid source_checks fixture containing
only src while that function checks the whole tree. Disregard those two initial
source_checks verdicts: `specific.json` supersedes them with the complete-root
positive/negative observations. Likewise the nested runner-lock attempt was a
harness deadlock, not a reader result; the successful specific run is unwrapped
outside and individually serialized inside. All accepted measured subprocesses
record peak RSS; current full 99 checker peak is **168,476 KiB**. No inventory
projection invokes lowering or creates a new census.

`final-readers.json`, `specific.json`, their harnesses, and the original Go
filesystem-clause extracts (`*.go.txt`) make the focused checks reproducible.

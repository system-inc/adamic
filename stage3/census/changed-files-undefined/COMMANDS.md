# Commands and provenance

The feature is one commit on 5d334653fa9ce301738e9eb2ef016c99da4589b8.
Absolute scratch paths in the receipts deliberately identify the actual inputs.
Source the retained env.sh when reproducing this session. No Loom command was run.

Setup used a fresh detached worktree, recursive submodule initialization, and
npm ci in stage3/api. Both recursive submodules succeeded; exact commits are in
provenance.json. Upstream was cloned with --depth 1 --branch v6.0.3; both the tag
and HEAD resolve to 050880ce59e30b356b686bd3144efe24f875ebc8, version 6.0.3.
A redundant full-history clone completed too, but was not the proof input.
The lane's locked dependencies were installed with npm ci in stage3/typescript.

Run prepare-baseline.sh **at the base commit, before amending the owner**.
It runs the unchanged stage3/apply.sh into main-tree. build-tools.sh generates
the repository's guarded census overlay and builds latent-census and adamic.
All physical retained overlay Go files are named .go.txt; overlay.json maps
those files to the compiler's original virtual paths. latent-audit.log records
the successful no-output/continuation audit.

prepare-pair.py copies baseline src bytes to after-tree and changes precisely
the declaration. It also extracts the actual timer adapter's prerequisite tree
from refs/stage3/99-correctness-timer-closures-before for contract.cjs.
amend-owner.py changes only that owner's after text and after hash. The owner
contract was run as:

```
node stage3/adapt/99-correctness-timer-closures/contract.cjs PROOF/contract-input
```

The binding audit's seven identifier occurrences are in reader-inventory.json.
The amended adapter was replayed on after-tree with fs.writeFileSync replaced
by a throwing function: zero writes and zero edits (amended-owner-replay.json).
Both witness.cjs runs used the full actual main/after source trees; their JSON
outputs are byte-identical.

run-probe.py LABEL TREE invokes the same latent-census binary on all 79 compiler
roots recursively, with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1,
LATENT_ONLY_FILE=TREE/src/compiler/sys.ts, and a separate LATENT_TIMES stream.
No inherited LATENT_* or CENSUS_MODULE_FORMAT override is retained. Header
module_format is CommonJSBundle, mode full, checker_rejected false, diagnostics
empty. Every authoritative stream has its relation_report trailer and exit 0.
Commands, flags, 79 roots, source hash and elapsed time are in each receipt.

restore-mutant.py writes sys-restoring-mutant.ts.txt by restoring the declaration
in after-tree; it asserts exact byte equality with main sys.ts and builds the
restoring-mutant-tree input. run-probe.py measures it under identical flags.
compare-census.py lists every removed/new unique finding and every changed raw
observation, including boundary events. Only tree paths and the exact 14-byte
sys.ts span relocation are normalized; reason, text, owner and phase remain.
The original LF-normalized preview was terminated and excluded; its partial
stream and receipts are retained under excluded-preview for transparency.
The authoritative pair and mutant preserve CRLF.

run-types.py invokes the actual built adamic types command once on the complete
79-root list, for main and the authoritative amendment. Both exit 0. Full stdout
and stderr are retained as compressed logs; each receipt lists every root.

sys.js was emitted with TypeScript 6.0.3's transpileModule, fileName sys.ts,
reportDiagnostics true, target ES2024, module ESNext, sourceMap false. Both
emissions had no transpile diagnostics. Their complete output bytes are retained,
with their common hash and options in javascript.json.

The final amendment also ran:

```
stage3/apply.sh PROOF/amended-apply-tree
go test -json -count=1 -timeout=30m ./stage3/lane/adaptedtree -run '^TestProduct_Stage3AdaptedTree$'
go test -json -count=1 -parallel=2 -timeout=45m ./stage3/lane
```

fresh-apply.json confirms all 747 fresh-apply source files exactly match the
measured amendment manifest. The lane product preparation and whole lane have
complete Go JSON streams. The API generator was run with --output in scratch
and cmp against sanctioned-api.json; generate_shards.py --check also passed.
The lane's three generated files were left unchanged because hashed API and
runtime inputs did not move. finish-verification.py validates the final lane
verdict, all 398 shards, source equality, parent commit, and all proof exits.

# Proof commands

The scripts record each child's actual exit code. Logs are written to files.
The compiler source equals main 3214dd465; no production Lower file is changed.
The original baseline evidence archive retains setup and the first full apply.

```sh
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh
npm ci --prefix stage3/api
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
# Source HEAD: 050880ce59e30b356b686bd3144efe24f875ebc8
# Main worktree: /workspace/or-predicate-main
bash /workspace/or-predicate-main/stage3/apply.sh --build-product /workspace/or-predicate-before
bash stage3/apply.sh --build-product /workspace/or-predicate-after
```

Both complete applies exit 0. Upstream package.json and package-lock.json were
installed in /workspace/or-predicate-packages; each completed tree's node_modules
links to that same exact install. Both npm run build commands exit 0.

```sh
# From each completed tree:
npm run build
# All 79 sorted src/compiler/**/*.ts files, in one invocation per tree:
/workspace/scratch/or-predicate/adamic types ROOTS
```

The latent overlay is generated and built exactly as the original core census:

```sh
python3 stage3/census/latent/make_overlay.py /workspace/adamic /workspace/scratch/or-predicate/overlay
# Rename generated .go files to .go.txt and update overlay.json replacements.
go build -buildvcs=false -overlay=/workspace/scratch/or-predicate/overlay/overlay.json \
  -o /workspace/scratch/or-predicate/latent ./stage3/census/latent/tool
```

Core censuses use LATENT_FULL=1, LATENT_ONLY_FILE=TREE/src/compiler/core.ts,
LATENT_ASSERT_NO_OUTPUT=1 and a separate LATENT_TIMES output. No speculative or
generic-forcing flag is used. Before and after each exit 0 with zero checker
diagnostics. core-compare.py checks every event and source coordinate.

checker-census.py runs the same full driver filtered to checker.ts, with a
5,400-second wall cap and a 1,800-second child-output idle cap. It records the
actual exit/signal and last completed unit. Both attempts are retained,
including any partial result on timeout. These censuses do not control whether
the separate witness program is considered accepted.

```sh
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-core-or-predicate/contract.cjs \
  /workspace/or-predicate-before /workspace/scratch/or-predicate/contract.json
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-core-or-predicate/compare.cjs \
  /workspace/or-predicate-before /workspace/or-predicate-after /workspace/scratch/or-predicate/comparison
```

contract.cjs's ten checks pass. compare.cjs checks every source file and emitted
JavaScript file, both public declarations, exact internal declaration lines,
and the two new runtime properties. It reverts only the new helper definitions,
helper call names, generated export properties and dead CommonJS export-name
annotations before comparing full runtime ASTs. Literal/token values, operators,
bindings and variable declaration kinds remain part of that comparison. Every
raw emitted hunk is preserved in comparison/all-emitted-hunks.diff.

calls.cjs inventories all 23 resolved signatures, retaining every ordered
argument, reference overload identity and diagnostic. Refreshed before/after
inventories each have the same 23 unrelated diagnostics; the initial unlinked
program's two missing diff imports explain its count of 25. The diagnostics'
source lines and columns are independently compared. The own builds both pass.

The witnesses are built by extracting the exact original/adapted core declaration
group and the node/predicate/findAncestor shapes from
internal/lower/predicates_overload_rest_test.go. Both original arities and both
mutants reproduce the binder refusal. Both full adapted-text witnesses and the
helper-only companions are blocked, with the exact errors in README.md. Node
returns the expected values for every program. Neither backend has a successful
accepted helper runtime witness. Compiler commands used are:

```sh
adamic js main.ts
adamic build main.ts -o native --sanitize
node node.mjs
```

The independent Node site witness evaluates the actual copied constructors,
uses stock 6.0.3's isImportEqualsDeclaration/isExportDeclaration and
isImportDeclaration/isJSDocImportTag/isExportDeclaration callbacks, and checks
identical results and call traces on identical real node objects. It proves
short-circuiting in fourteen cases. This is a witness, not a new lane count.

replay.py runs every adapter from the corresponding repository on isolated
copies of the completed baseline/candidate trees. replay-review.py checks all
exits, writes, source changes and outputs. It independently validates 71's
createTypeChecker runtime fingerprint with replay-fingerprint.cjs and proves
78's coordinate mapping from unchanged complete source lines. All raw outcomes
and write logs are packed into the proof archive.

Proof metadata added after the applies changes product keys but does not change
the frozen transformation, its pins, the measured source snapshots or the
adapter outcomes. No source, compiler, lane manifest or expectation was changed
to make a failed proof pass. Backend acceptance remains explicitly blocked.

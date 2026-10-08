# Native tsc entry measurement

The project-reference build witness and scratch source-root experiment are in
[PROJECT-REFERENCES.md](PROJECT-REFERENCES.md). All compiler changes for that
experiment remain outside this publication checkout.

The project-options composition and remaining-site adaptation audit are in
[MODE-AND-ADAPTATIONS.md](MODE-AND-ADAPTATIONS.md). Its scratch compiler has
explicit uncommitted reconciliation choices; admission is separate from emission.

The two compiler-tip comparison is recorded in [COMPILER-TIPS.md](COMPILER-TIPS.md),
with exact pins, all fifteen saved-stop classifications and reproducible commands.

The current-main rerun is recorded in [REFRESH.md](REFRESH.md), with evidence
under `evidence/main-efe9f404/` and exact-message comparison against
`stage3/census/latent/REPORT.md`. The original unit remains in `REPORT.md`.

This unit starts at adapted TypeScript 6.0.3's `src/tsc/tsc.ts`.
The compiler is unchanged main, pinned in `evidence/provenance.json`.
All source replacements live in a disposable copy outside the repository.

Source the environment printed by `bash cloud/setup.sh`, then run from the
compiler checkout:

```sh
bash stage3/drivers/tsc-entry/run.sh /tmp/tsc-entry-new > /tmp/tsc-entry-run.log 2>&1
```

The output directory must be new. The runner requires Node 24.19.0, constructs
the complete adapted tree with `stage3/apply.sh`, records the graph, builds
main's compiler, and attempts both `ADAMIC_NATIVE_SPLIT=0` and `1`.
Each attempt retains stdout, stderr and exit separately. Compilation failure
is an observation, never a successful native comparison.

The graph includes imports, re-exports, import-type expressions and literal
`require`/dynamic imports. Unresolved host modules are recorded separately.
`check-graph.cjs` compares the source closure with stock TypeScript's independent
program loader. It does not require the source to pass typechecking.

```sh
export SCANNER_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
node stage3/drivers/tsc-entry/check-graph.cjs TREE OUTPUT/closure.json > graph-check.log 2>&1
node stage3/drivers/tsc-entry/node.mjs TREE/src/tsc/tsc.ts --version > node.log 2>&1
```

The Node loader follows the scanner driver's stock `transpileModule` and
`.js`-to-`.ts` resolution. It supplies the CommonJS host globals used by
upstream's bundled CLI. It does not use Adamic-generated JavaScript.
For authored reproducers, use `.a`; the same loader transpiles them on Node.

`stub.cjs FILE LINE COLUMN RECORD.json` replaces the innermost enclosing
function body with a throwing placeholder. Use it only in a disposable source
copy. A stopping site outside a body exits 2 with an explicit explanation.
This tool is not a source adaptation and cannot be used to claim working tsc.

To reproduce the fifteen-stop experiment after the first build, copy only
`OUTPUT/adapted/src` to a fresh disposable directory and run:

```sh
cp -a OUTPUT/adapted/src /tmp/tsc-entry-disposable-src
python3 stage3/drivers/tsc-entry/continue.py /tmp/tsc-entry-disposable-src OUTPUT/adamic /tmp/tsc-entry-new-stops > continuation.log 2>&1
```

Each build runs both split modes. The helper replaces a body only after saving
that stop, stops at fifteen observations, and rejects sources inside this
repository. An expression-bodied callback gets a throwing block. If a callback
parameter is the stopping site, the enclosing function whose body contains that
parameter is replaced. This can change inferred types; the report flags messages
absent from the pristine build.

`collect.py PROGRESS_DIRECTORY --evidence EVIDENCE --census-report stage3/census/latent/REPORT.md`
compares messages with the main exact-reasons table in the requested latent
report. Without these options, `collect.py PROGRESS_DIRECTORY` compares with
the newest main meter's latent lowering `per_reason` counts. It replays recorded
UTF-16 replacements to map scratch coordinates to the original adapted source.
It never commits or copies a modified tree. The checked-in witness outputs are
validated with these commands:

```sh
python3 stage3/drivers/tsc-entry/verify.py > verification.log 2>&1
python3 stage3/drivers/tsc-entry/mutants.py > mutants.log 2>&1
```

The complete module list and edges are in `evidence/closure.json`; files outside
`src/compiler` are in `evidence/outside-compiler.json`. `REPORT.md` records the
stopping sites, reproducers, meter comparison, mutants and coverage limits.

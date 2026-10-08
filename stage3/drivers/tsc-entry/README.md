# Native tsc entry measurement

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

The complete module list and edges are in `evidence/closure.json`; files outside
`src/compiler` are in `evidence/outside-compiler.json`. `REPORT.md` records the
stopping sites, reproducers, meter comparison, mutants and coverage limits.

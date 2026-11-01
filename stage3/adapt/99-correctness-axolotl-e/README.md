# 99e: tolerate an empty repeated extends clause

**JavaScript-changing.** Declaration emit skips an empty extends clause in the
expression-base rewrite. It no longer passes an absent type node to the
accessibility diagnostic helper. Programs without that shape keep their output.

lane: run as part of Correctness's wave

Branch: `typescript/axolotl-99e`, cut from fetched main
`c74051ffc5ee94f4b7f732b1532cb219032b183d`. The prerequisite 99a record was read at
`74dda102dc060de34dd7c8e7b4fd63f3539ef270`. Source: TypeScript 6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`. No lane manifest changes.

## Failure and output contract

The original witness is `export class C extends f() extends {}`. A further
reduced standalone program is `class C extends 0 extends {}`; the CLI and
transpileDeclaration stacks are in [stock-minimal-cli.txt](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/stock-minimal-cli.txt)
and [stock-minimal-api.txt](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/stock-minimal-api.txt). It needs no export or unresolved function.
The reduced input reports TS2507 for the numeric base and TS1172 for the
second extends clause. Both forms are checked by witness.cjs.

With Node 24.19.0 and stock TypeScript 6.0.3, `createSourceFile` has no parse diagnostics
and heritage type lengths `[1, 0]`. `program.getSemanticDiagnostics()` reports
TS2304 for `f` and TS1172 at offset 27, `'extends' clause already seen.`
The grammar diagnostic is produced by checking, rather than parsing.

Run outside a directory containing tsconfig.json:

```sh
printf '%s\n' 'export class C extends f() extends {}' > input.ts
node /workspace/adamic/stage3/api/node_modules/typescript/lib/tsc.js --declaration input.ts > cli.log 2>&1
```

`ts.transpileDeclaration(text, {fileName: "input.ts"})` and `program.emit()`
also crash. The CLI and API stacks are preserved in
[stock-cli.txt](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/stock-cli.txt) and [stock-api.txt](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/stock-api.txt).
The shared frames are `isVariableDeclaration`,
`createGetSymbolAccessibilityDiagnosticForNode`, the heritage callback,
`map`, and `transformTopLevelDeclaration`. The exception is
`TypeError: Cannot read properties of undefined (reading 'kind')`.
`getPreEmitDiagnostics` also crashes because it requests declaration diagnostics.

The reduced numeric witness emits `declare const C_base: number;` and
`declare class C extends C_base {}` after the patch. The original exported
witness's corrected declaration is:

```ts
declare const C_base: any;
export declare class C extends C_base {
}
export {};
```

The expression-base branch already uses the first effective base type to create
`C_base`. The bug is that its callback assumes every subsequent extends clause
also has a type. The ordinary `transformHeritageClauses` path filters out empty
clauses. This fix gives the special expression-base branch the same behavior
for an empty extends clause, while preserving nonempty repeated clauses and
all implements clauses exactly as before.

[Source diff](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/source.diff). The local `firstType` is checked before it is
passed to the diagnostic helper. `mapDefined` drops only the callback's new
undefined return. No assertion claims that a parsed clause must be nonempty.

## Native and upstream comparison

The repository pins cohere at `06af2e399c6c1fb93fd570b774027942747400b4`,
whose TypeScript submodule is `d92d9bfee114c80be2c375d72edae966176e3a4f`.
This is the native compiler lineage from microsoft/typescript-go, now hosted
in microsoft/TypeScript. In `tsc/internal/transformers/declarations/transform.go`,
`transformHeritageClause` returns nil for zero retained types. The expression
base rewrite constructs its replacement from the effective first base type
rather than indexing every extends clause. The actual native build command is
`go build -o /workspace/scratch/99e-tsgo ./cmd/tsc` from
cohere/TypeScript/tsc, then, from the witness directory,
`/workspace/scratch/99e-tsgo --declaration input.ts`. It exits 2 with TS2304
and TS1172, and writes exactly the corrected declaration above.
[native.txt](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/native.txt) and [native.d.ts](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/native.d.ts) retain it.
[native.json](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/native.json) also records both inputs and actual exit codes.

There is a separate native inference difference on the reduced numeric witness:
the pinned native compiler emits `C_base: 0`; TypeScript 6 with this fix emits
`C_base: number`. Both elide the empty clause. Stock TypeScript 6 already emits
`C_base: number` for `class C extends 0 {}`, and that single-clause control is
byte-identical after the patch. This adaptation preserves the JavaScript
compiler's inference rather than importing the native compiler's literal type.

TypeScript main was inspected at
[`6161772c1370bdd4a82f808e6823bd755eb1870d`](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/transformers/declarations/transform.go#L737). It is now the Go implementation
and retains this empty-clause elision and first-base construction. Its behavior
therefore already tolerates this input. The legacy [`release-6.0` source](https://github.com/microsoft/TypeScript/blob/release-6.0/src/compiler/transformers/declarations.ts#L1678) still
contains the unsafe `clause.types[0]` read at 1678. No JavaScript fix commit is
claimed; the cited main commit identifies the native implementation inspected.

## Adapter contract and validation

Stock TypeScript 6.0.3 locates the unique named
`transformTopLevelDeclaration` and its unique createNodeArray heritage mapping.
The entire mapping statement must match either reviewed exact text before or
after. A half-applied map change or altered guard is refused before writes.
CRLF bytes are preserved. The second run returns zero files and zero edits.
[contracts.json](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/contracts.json) records the passing refusal checks.
[check.cjs](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/check.cjs) checks those contracts on isolated scratch copies.

[witness.cjs](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/witness.cjs) tests both transpileDeclaration and createProgram
emit using the actual built control and fixed compilers. Its ordinary class is
`declare function f(): new () => {}; export class C extends f() {}`. Complete
transpile results, including diagnostic data with file objects reduced to
names, and program emit output are compared byte for byte. The malformed input is required to crash before and to produce
the exact declaration above after.

Whole-program checking uses the same invocation as 99a: all 79 sorted .ts/.a
roots under src/compiler passed together to `adamic types`, with no options or
prelude changes. [measure.py](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/measure.py) retains complete diagnostics
and compares their multisets, aligning unchanged source lines past the two
inserted guard lines. Only the requested TS2345 may disappear; additions fail.

The measured whole-program counts are **144 before and 143 after**, exactly
one removed diagnostic and zero new diagnostics. declarations.ts goes from
one diagnostic to zero. [measurement.json](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/measurement.json) records
the removed TS2345 at 1678:110. Both complete stderr inventories are retained.
[scope.json](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/scope.json) checks all 746 source files: only declarations.ts
changes, with four lines added and two removed. The full apply snapshot
immediately before 99e equals the complete control source set byte for byte.

Both trees are built end to end with stage3/apply.sh, before and after this
adaptation. Full oracle runs are sequential, all runners and no test filter,
with four workers and a 1800 second limit each. The control must finish before
the fixed run starts. No reference baseline is accepted or rewritten.

## Full oracle result

| Tree | Passing | Failing | Pending | Wall seconds |
| --- | ---: | ---: | ---: | ---: |
| Control | 106,366 | 1 | 0 | 389.424 |
| With 99e | 106,366 | 1 | 0 | 379.200 |

Both install and build phases exit 0. Both test phases exit 1 solely for
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
The existing lane API normalizer accepts all 223 sanctioned declarations, with
none missing, extra or changed. No sanction or manifest is edited.

**No baseline moved.** The complete baseline.diff is byte-identical before and
after: 51,123 bytes, SHA-256
`6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`.
The existing API mismatch is present in both controls. The built public API is
also byte-identical: 590,996 bytes, SHA-256
`b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`.
[oracle-comparison.json](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/oracle-comparison.json) and the compressed
phase logs retain the full results. These bytes describe this main tip; they
are not the historical diff size from 91's earlier base.

[witness.json](https://github.com/system-inc/adamic/blob/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs/witness.json) records both actual built compilers: both
malformed inputs crash before and emit the checked declarations after; the
ordinary class is byte-identical in transpileDeclaration and program emit.
The final post-suite contract and all-source scope checks pass again, with
zero adapter edits. No OOM or OOM-kill event occurred. Setup and adaptation
finished before the oracle runs. The full suites ran sequentially, and the
final built-compiler witnesses ran after both suite reports completed.

## Reproduction commands

The adaptation directory contains only adapt.cjs and this README. All proof
tools, results and logs live on the separate
[`typescript/axolotl-99e-evidence`](https://github.com/system-inc/adamic/tree/typescript/axolotl-99e-evidence/stage3/evidence/99-correctness-axolotl-e/proofs)
branch under stage3/evidence/99-correctness-axolotl-e/proofs/. Use a checkout of
that branch to run the proof tools below. They are outside apply's adaptation
input set.

From the repository root, with all output redirected to files:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund > /workspace/scratch/99e-packages.log 2>&1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/99e-upstream.log 2>&1
# Run the first apply before creating 99e, or exclude only 99e from adaptation_inputs.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/99e-before > /workspace/scratch/99e-apply-before.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/99e-after > /workspace/scratch/99e-apply-after.log 2>&1
go build -o /workspace/scratch/adamic ./cmd/adamic > /workspace/scratch/99e-build-adamic.log 2>&1
NODE_PATH=stage3/api/node_modules node stage3/evidence/99-correctness-axolotl-e/proofs/check.cjs /workspace/scratch/99e-before /workspace/scratch/99e-after > /workspace/scratch/99e-contract.log 2>&1
python3 stage3/evidence/99-correctness-axolotl-e/proofs/measure.py /workspace/scratch/adamic /workspace/scratch/99e-before /workspace/scratch/99e-after /workspace/scratch/99e-measure > /workspace/scratch/99e-measure.log 2>&1
bash stage3/oracle/run.sh /workspace/scratch/99e-before /workspace/scratch/99e-oracle-before --workers 4 --limit-seconds 1800 > /workspace/scratch/99e-oracle-before.log 2>&1
bash stage3/oracle/run.sh /workspace/scratch/99e-after /workspace/scratch/99e-oracle-after --workers 4 --limit-seconds 1800 > /workspace/scratch/99e-oracle-after.log 2>&1
node stage3/evidence/99-correctness-axolotl-e/proofs/witness.cjs /workspace/scratch/99e-before/built/local/typescript.js /workspace/scratch/99e-after/built/local/typescript.js > /workspace/scratch/99e-witness.log 2>&1
python3 stage3/evidence/99-correctness-axolotl-e/proofs/native.py /workspace/scratch/99e-tsgo > /workspace/scratch/99e-native.json 2>&1
python3 stage3/evidence/99-correctness-axolotl-e/proofs/verify-oracles.py /workspace/scratch /workspace/scratch/99e-verification > /workspace/scratch/99e-verification.log 2>&1
```

The upstream issue is drafted in [LEDGER.md](../../upstream/LEDGER.md).
Nothing is filed upstream.

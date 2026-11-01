# 99a: required transformer reads

Byte-identical in JavaScript. Thirteen required-read assertions in four transformer files address fifteen of the sixteen requested diagnostics. The fifth file, declarations.ts, has a disproved invariant and stays untouched. No public API changes.

lane: run as part of Correctness's wave

Base: main f3fc89cc26bfcf1771bd992a07552a9e0ca87c5b. Source: stock TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Reference adapter: emit/axolotl-96c, stage3/adapt/96-emit-axolotl-c. Adaptation 33's sites.json and README supply the original declines; its declaration counterexample survives this independent review.

Each `!` is at the read's own evaluation point. Node erases it; Adamic performs a loud required-value check. No store, optional test, callback, or read is moved. Coordinates below refer to the pinned source, also the diagnostic coordinates. Paths are relative to src/compiler/transformers/ unless stated otherwise. Each row describes one requested diagnostic; shared reads are asserted once.

| Site | Required read and function | Invariant and witness |
| --- | --- | --- |
| declarations.ts:1678 TS2345 | `clause.types[0]`, transformTopLevelDeclaration's heritage callback | **Unasserted.** The 1662-1663 effective-base guard proves only the first extends clause has an element. compiler/utilities.ts:5195-5211 selects the first clause and checks its length; compiler/parser.ts:8196-8212 permits multiple clauses with empty type lists. `export class C extends f() extends {}` parses without diagnostics, with extends lengths [1, 0]. The mapped second clause reaches this read absent. Stock declaration emit crashes reading `kind`. No truthful presence invariant exists here. |
| esDecorators.ts:1164 TS2538 | `superPath[superPathDepth]` and `statementsIn[superStatementIndex]`, transformConstructorBodyWorker | A path entry exists at the current depth. visitConstructorDeclaration:1207-1209 calls the worker at depth 0 only for a nonempty path. utilities.ts:547-558 unshifts bounded statement indices only on a successful path ending in a super-call statement. A try entry therefore always has a following entry; worker:1166-1175 recurses only into that try block, incrementing depth once. Each stored index is in bounds in its corresponding populated NodeArray, proving the statement slot at 1164 as well. The input path and NodeArrays are not mutated. That second assertion also prevents a newly exposed TS2345 at isTryStatement:1166 once the index is defined. |
| esDecorators.ts:1165 TS18048 | Same path read at 1163 | Same path construction and nonempty entry guard prove the numeric index exists before subtraction. |
| esDecorators.ts:1192 TS18048 | Same path read at 1163 | Same path construction and guarded recursion prove the numeric index exists before addition. |
| esnext.ts:183 TS2345 | `node.statements[pos]`, visitSourceFile | The populated NodeArray is read under `pos < node.statements.length` at 181; pos starts at countPrologueStatements' nonnegative result at 180. That helper returns a bounded index or zero, and pos only increments. |
| esnext.ts:344 TS2345 | `statementsIn[i]`, transformUsingDeclarations | Loop:342 has `start <= i < end`. Both callers, at 196 and 246, supply the same NodeArray's length as end and a nonnegative bounded statement/prologue index as start. The transform builds separate output arrays and does not remove input slots. |
| esnext.ts:384 TS2345 | Same input read at 343 | The same bounded input read supplies statement to visitor after getUsingKind; its non-using branch preserves the input slot and local value. |
| esnext.ts:767 TS2345, column 34 | `statements[i]` occurrence 1, countPrologueStatements | Loop:766 starts i at zero and requires i < statements.length over the populated readonly statement array. |
| esnext.ts:767 TS2345, column 70 | `statements[i]` occurrence 2, countPrologueStatements | Same bound; isPrologueDirective only inspects the statement and cannot remove the second read's slot. |
| jsx.ts:309 TS2345 | `nonWhitespaceChildren[0]` occurrence 2, convertJsxChildrenToChildrenPropAssignment | The length === 1 guard at 308 proves a first child. utilities.ts:6046-6057 filters populated JSX children. The intervening dotDotDotToken inspection does not mutate the array. Occurrence 1's existing cast is left alone. |
| jsx.ts:502 TS2345 | `expressions[0]`, transformJsxAttributesToExpression | `expressions.length &&` at 502 guards the read. All insertions at 482, 490 and 513 push defined expressions, checked visit results or factory-created object literals; this private array has no holes. |
| module/esnextAnd2015.ts:100 TS2345 | `node.arguments[0]` occurrence 1, transformSourceFile's scanner callback | forEachDynamicImportOrRequireCall is called with includeTypeSpaceImports=false at 99. compiler/utilities.ts:12182-12186 calls back only on require calls accepted by isRequireCall or import calls with arguments.length >= 1. isRequireCall:3853-3856 requires exactly one argument. Parsed argument NodeArrays are populated. Occurrence 2 is already narrowed by the predicate. |
| module/esnextAnd2015.ts:163 TS2345 | `importsAndRequiresToRewriteOrShim[0]`, visitor | The same expression's optional length guard at 163 short-circuits absent or empty queues. The queue is populated by append at 101; its only removal is shift at 159. No callback or mutation intervenes between guard and read. The optional test at 158 stays optional. |
| module/esnextAnd2015.ts:194 TS2345 | `node.arguments[0]` occurrence 1, visitImportOrRequireCall | Its sole caller at 159 shifts a node previously queued by the guarded scanner callback at 99-101, as above. Nothing changes its arguments before this predicate. |
| module/esnextAnd2015.ts:196 TS2345 | `node.arguments[0]` occurrence 3, visitImportOrRequireCall | Same queued-node producer contract. The predicate and emitHelpers() access do not mutate that call's arguments. Occurrence 2 at 195 remains narrowed; the false branch needs its own assertion. |
| module/esnextAnd2015.ts:251 TS2532 | `importRequireStatements[1].declarationList.declarations[0]`, createRequireCall | The absent-cache branch at 217 creates requireStatement with exactly one factory-created declaration at 234-246, then stores the tuple at 248. Other assignments only reset the cache at 97. The tuple's second slot is already required by its declared type; only the singleton declaration read needs an assertion. |

## Adapter contract

Stock TypeScript 6.0.3 parses each file. Every read is pinned by its unique named function, exact expression text, occurrence ordinal, and total count in that function, including nested callbacks. Assignment, update, delete and iteration targets are rejected. Each file must be fully untouched or fully adapted. All file plans, parse checks and unchanged-input checks finish before the first write; each changed file receives one write. Insertion retains existing newline bytes. A second run makes no writes.

## Proof

Both baseline and final trees are built end to end by stage3/apply.sh. The baseline excludes 99a; the final tree includes it. The baseline final source snapshot equals the final tree's 99a-before snapshot, proving the two inputs agree before this adapter.

The classification command record is typescript/axolotl-classify at ad5d21513f76663bef354d66fc0d995b4610c15f, stage3/census/axolotl/commands.sh. Its census tool sorts every .ts/.a root recursively under src/compiler and calls load.Load. The CLI below passes the same 79 sorted roots to adamic types, whose check also calls load.Load, with the same fixed strict compiler options and prelude. No checker option or source overlay is changed.

| Measurement | Before | After |
| --- | ---: | ---: |
| Whole program, every checker code | 144 | 129 |
| declarations.ts own diagnostics | 1 | 1 |
| esDecorators.ts own diagnostics | 3 | 0 |
| esnext.ts own diagnostics | 5 | 0 |
| jsx.ts own diagnostics | 2 | 0 |
| module/esnextAnd2015.ts own diagnostics | 5 | 0 |
| Five files combined | 16 | 1 |

All sixteen specified diagnostics reproduce at the requested coordinates. Exactly fifteen are removed. An exact multiset comparison of complete diagnostic strings, replacing only the two tree root paths, finds zero additions anywhere; columns, codes, messages, multiplicity and elaboration chains are retained. Main has moved since the classification's a23ba096 baseline, so this measurement uses its own whole-program before count.

Stock ts.transpileModule, target ES2024 and module ESNext with comments retained, produces byte-identical JavaScript for every edited file and the unchanged declaration file:

| File | JavaScript bytes before and after |
| --- | ---: |
| esDecorators.ts | 109299 |
| esnext.ts | 28597 |
| jsx.ts | 31634 |
| module/esnextAnd2015.ts | 14856 |
| declarations.ts, unchanged | 84495 |

The first run writes four files and inserts thirteen assertions. A second run returns `{"files":0,"assertions":0,"javascriptChanging":false}`; saved file bytes remain equal. Isolated fixtures also prove that a partly adapted file, a changed occurrence count in the last file, and an assignment target in the last file all throw before any file is written.

Material commands, from the repository root, with full logs under /workspace/scratch:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
source /workspace/adamic-tools/env.sh
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream.log 2>&1
(cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund > /workspace/scratch/packages.log 2>&1)
# Baseline run was started before creating this adaptation directory.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/before > /workspace/scratch/apply-before.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/after-final > /workspace/scratch/apply-after-final.log 2>&1
go build -o /workspace/scratch/adamic ./cmd/adamic > /workspace/scratch/build-adamic.log 2>&1
```

The two adamic types invocations are made with argument arrays, avoiding shell expansion limits:

```python
import pathlib, subprocess
for label, tree in [("before", "/workspace/scratch/before"),
                    ("after", "/workspace/scratch/after-final")]:
    roots = sorted(str(p) for p in pathlib.Path(tree, "src/compiler").rglob("*")
                   if p.suffix in [".ts", ".a"])
    with open(f"/workspace/scratch/types-{label}.stdout", "w") as out, \
         open(f"/workspace/scratch/types-{label}.stderr", "w") as err:
        result = subprocess.run(["/workspace/scratch/adamic", "types", *roots],
                                stdout=out, stderr=err)
    assert result.returncode == 1  # Remaining whole-program diagnostics.
```

The scratch measurement99.json retains the counts and all removed diagnostics. diagnostics-before.json and diagnostics-after.json retain every full diagnostic chain. verify99.json records byte equality and the no-write second run; contracts99-final.json records the refusal checks; counterexample.json records the independent stock declaration failure. The first diagnostic comparison caught an exposed statement-slot error; the final adapter includes both proven super-path reads and passes the strict comparison.

No lane manifest is edited. Lane execution remains part of Correctness's wave.

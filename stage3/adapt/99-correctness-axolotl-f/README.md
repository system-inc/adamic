# 99f: createSet describes its custom collection

Byte-identical in JavaScript. The preferred repair declares `HashedSet<T>` with exactly the eleven members constructed by `createSet`, and carries that type through its internal consumers. No casts, Set-operation implementations, or public API changes.

lane: run as part of Correctness's wave

Base: main c74051ffc5ee94f4b7f732b1532cb219032b183d. Source: stock TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Adapter style reference: typescript/axolotl-99a, stage3/adapt/99-correctness-axolotl-a. D037 is the object assignment at stock core.ts:1637:11, TS2740. The ES2025 operations required by Adamic's prelude are absent at runtime; changing the declared lib cannot repair that declaration.

## Inventory and flow audit

`createSet` is internal: its declaration has `@internal`, and neither it nor `updateIsDefinitionOfReferencedSymbols` occurs in tests/baselines/reference/api/typescript.d.ts. Namespace exports make it available internally and to tests without making it a public API. A repository-wide source search finds eight direct calls, all below. There are no direct calls in compiler, services, or harness. The server wrapper has four calls. All coordinates below refer to stock 6.0.3.

| Direct caller | Result and outgoing flow | Declared or inferred type before; after |
| --- | --- | --- |
| server/session.ts:513, createDocumentSpanSet | Direct return from wrapper | Explicit `Set<DocumentSpan>` return; `HashedSet<DocumentSpan>` |
| server/session.ts:1675, Session.findSourceDefinition | Local definitionSet; add results discarded; values() passed to arrayFrom at 1721, yielding the array assigned to definitions | Local inferred `Set<DefinitionInfo>`; `HashedSet<DefinitionInfo>`. Iterator remains `SetIterator<DefinitionInfo>`, array remains `DefinitionInfo[]`, receiving definitions remains `readonly DefinitionInfo[]` |
| testRunner/unittests/compilerCore.ts:35, mutation | Local set; add(0) returns local newSet at 38; both passed to assert.strictEqual at 39 | Both locals inferred `Set<number>`; both infer `HashedSet<number>`. Assertion parameters accept arbitrary values |
| testRunner/unittests/compilerCore.ts:75, resizing | Local set, only member calls | Inferred `Set<number>`; `HashedSet<number>` |
| testRunner/unittests/compilerCore.ts:93, clear | Local set, only member calls | Inferred `Set<number>`; `HashedSet<number>` |
| testRunner/unittests/compilerCore.ts:106, forEach | Local set; callback consumes numeric value/key, ignores third argument; callback passed to set.forEach | Inferred `Set<number>`; `HashedSet<number>`. Callback's first two parameters remain number |
| testRunner/unittests/compilerCore.ts:129, iteration | Local set; keys(), values(), entries() go to arrayFrom; resulting arrays go to sorting, mapping and assertions | Inferred `Set<number>`; `HashedSet<number>`. Iterators remain `SetIterator<number>` and `SetIterator<[number, number]>`; arrays remain number[] and [number, number][] |
| testRunner/unittests/compilerCore.ts:157, string hash code | Local set, only member calls | Inferred `Set<Thing>`; `HashedSet<Thing>` |

| Wrapper consumer | Flow beyond local | Declared or inferred type before; after |
| --- | --- | --- |
| server/session.ts:541, getRenameLocationsWorker's seen | has/add only; add return discarded | Inferred `Set<DocumentSpan>`; `HashedSet<DocumentSpan>` |
| server/session.ts:611, getReferencesWorker's knownSymbolSpans | add; passed at 627 to LanguageService.updateIsDefinitionOfReferencedSymbols | Inferred `Set<DocumentSpan>`; `HashedSet<DocumentSpan>` |
| server/session.ts:650, getReferencesWorker's seenRefs | has/add only; add return discarded | Inferred `Set<DocumentSpan>`; `HashedSet<DocumentSpan>` |
| server/session.ts:2279, Session.getFileReferences' seen | has/add only; add return discarded | Inferred `Set<DocumentSpan>`; `HashedSet<DocumentSpan>` |

The 627 call is the only invocation of updateIsDefinitionOfReferencedSymbols anywhere in source. Its parameter is explicitly `Set<DocumentSpan>` in services/types.ts:694, services/services.ts:1979, and harness/client.ts:1066. All three become `HashedSet<DocumentSpan>`. The service implementation only calls has/add, discards add's result, and returns boolean; its nested getSymbolForProgram captures the same collection and only calls has. The function is installed into the returned LanguageService object at services.ts:3437; the method declaration remains `@internal`. The harness implementation returns notImplemented() without forwarding its unused argument. No caller supplies a real Set. No collection is assigned to a field, returned by a public function, or passed to another set-valued parameter. Unrelated native Sets, including updatedProjects at session.ts:622, stay unchanged.

Inside createSet, the local set and function return are explicitly Set<TElement>; add explicitly returns Set<TElement> and returns this; forEach explicitly types its callback's third parameter as Set<TElement> and supplies the captured set at both callback sites. All four annotations become HashedSet<TElement>. The private map and iterator retain their existing types. Nothing widens to a union, and no flow requires the fallback.

## Edits and why each is true

Paths are relative to src/.

| Edit | Pinned scope and exact target | Why true |
| --- | --- | --- |
| 1 | compiler/core.ts, declaration inserted immediately before createSet's documentation | Internal HashedSet<T> declares has, add, delete, clear, readonly size, forEach, keys, values, entries, Symbol.iterator, and Symbol.toStringTag. These are exactly the object's own members. All iterator return types match the existing annotations; the tag is a string; size has a getter only |
| 2 | compiler/core.ts, createSet return Set<TElement> | The only returned value is this custom object |
| 3 | compiler/core.ts, createSet's const set: Set<TElement> | The object supplies exactly HashedSet's members |
| 4 | compiler/core.ts, createSet's add(element: TElement): Set<TElement> | add returns this same object |
| 5 | compiler/core.ts, createSet's forEach callback set: Set<TElement> | Both callback calls receive the captured custom set |
| 6 | server/session.ts, ./_namespaces/ts.js import after DocumentSpan | Import the internal return type through the existing namespace barrel |
| 7 | services/services.ts, same import after DocumentSpan | Import the implementation parameter type |
| 8 | services/types.ts, same import after HasChangedAutomaticTypeDirectiveNames | Import the internal method's parameter type; DocumentSpan is declared locally here |
| 9 | harness/client.ts, same import after DocumentSpan | Import the implementing stub's parameter type |
| 10 | server/session.ts, createDocumentSpanSet return Set<DocumentSpan> | Directly returns createSet's custom collection |
| 11 | services/services.ts, updateIsDefinitionOfReferencedSymbols knownSymbolSpans: Set<DocumentSpan> | Only caller supplies that custom collection; body uses has/add |
| 12 | services/types.ts, updateIsDefinitionOfReferencedSymbols knownSymbolSpans: Set<DocumentSpan> | Internal contract for the same implementation and sole caller |
| 13 | harness/client.ts, updateIsDefinitionOfReferencedSymbols _knownSymbolSpans: Set<DocumentSpan> | Stub implements that same internal contract |

The stock 6.0.3 parser identifies each unique named function or method, or the exact import module. Each replacement has exact before/after text and an occurrence-count check inside that scope. The complete interface text is checked on reruns. Every file must be completely untouched or completely adapted. All file plans, parse checks, and unchanged-input checks complete before any write, with one write per changed file. Existing newline bytes are retained. Fixtures prove that partial adaptation in core, partial adaptation in the final file, and final-file text drift all throw without writing any file.

## Proof

Stock upstream plus only 99f passes tsc -b src/server to prepare referenced declarations, then stock TypeScript 6.0.3 tsc -p src/compiler, tsc -p src/services, and tsc -p src/server, each with exit 0 under unchanged upstream project options. Upstream's locked dependencies are installed with npm ci. The setup adapter generates upstream's standard diagnostic inputs before these checks.

Stock ts.transpileModule uses target ES2024, module ESNext, and comments retained. Each edited file is compared against its stock git HEAD bytes:

| File | Identical JavaScript bytes before and after |
| --- | ---: |
| src/compiler/core.ts | 65103 |
| src/server/session.ts | 168298 |
| src/services/services.ts | 135449 |
| src/services/types.ts | 20298 |
| src/harness/client.ts | 38863 |

The first run reports five files and thirteen edits. The second reports `{"files":0,"edits":0,"javascriptChanging":false}`. Both file bytes and modification times remain equal on the second run.

Both complete adapted trees pass stage3/apply.sh end to end. The baseline uses a driver exported from the recorded main commit, excluding 99f. Its final source tree and the final tree's 99f-before snapshot are the identical Git tree `fed52b9cd7a0de69999b88217ffaa92153e57175`. This proves that all earlier adaptations saw matching inputs. The transpilation table above also holds for these complete before/after adapted files, not just stock plus 99f. The final adapted tree's second run makes no writes and preserves bytes and modification times.

The 99a README's invocation is used unchanged: recursively sort all .ts/.a roots under src/compiler and pass every root to adamic types, with no checker-option changes or overlays. Each invocation has 79 roots and exits 1 because unrelated diagnostics remain. The comparison retains complete diagnostic chains, coordinates, codes, messages, and multiplicities; it normalizes only the two tree root paths. No line or column normalization is needed.

| Whole-program measurement | Before | After |
| --- | ---: | ---: |
| All diagnostics, all checker codes | 144 | 143 |
| core.ts diagnostics | 1 | 0 |
| D037, core.ts:1637:11 TS2740 | 1 | 0 |
| New diagnostics anywhere | | 0 |

The sole removed diagnostic is:

```text
src/compiler/core.ts:1637:11: error TS2740: Type '{ has(element: TElement): boolean; add(element: TElement): Set<TElement>; delete(element: TElement): boolean; clear(): void; ... 6 more ...; [Symbol.toStringTag]: string; }' is missing the following properties from type 'Set<TElement>': union, intersection, difference, symmetricDifference, and 3 more.
```

Material commands from the repository root, with full logs saved under /workspace/scratch:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
source /workspace/adamic-tools/env.sh
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/99f-upstream.log 2>&1
(cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund > /workspace/scratch/99f-packages.log 2>&1)
# Frozen main-only driver for the baseline.
mkdir -p /workspace/scratch/99f-baseline-driver
git archive c74051ffc5ee94f4b7f732b1532cb219032b183d stage3 | tar -x -C /workspace/scratch/99f-baseline-driver
ln -s /workspace/adamic/stage3/api/node_modules /workspace/scratch/99f-baseline-driver/stage3/api/node_modules
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash /workspace/scratch/99f-baseline-driver/stage3/apply.sh /workspace/scratch/99f-baseline-clean > /workspace/scratch/99f-apply-baseline-clean.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/99f-after-final > /workspace/scratch/99f-apply-after-final.log 2>&1
go build -o /workspace/scratch/99f-adamic ./cmd/adamic > /workspace/scratch/99f-build.log 2>&1
```

The two diagnostic invocations, using argument arrays as in 99a:

```python
import pathlib, subprocess
for label, tree in [("before", "/workspace/scratch/99f-baseline-clean"),
                    ("after", "/workspace/scratch/99f-after-final")]:
    roots = sorted(str(p) for p in pathlib.Path(tree, "src/compiler").rglob("*")
                   if p.suffix in [".ts", ".a"])
    with open(f"/workspace/scratch/99f-types-{label}.stdout", "w") as out, \
         open(f"/workspace/scratch/99f-types-{label}.stderr", "w") as err:
        result = subprocess.run(["/workspace/scratch/99f-adamic", "types", *roots],
                                stdout=out, stderr=err)
    assert result.returncode == 1
```

Stock-project checks use the upstream checkout, its locked npm dependencies, upstream-generated diagnostic inputs from 00-setup, and only 99f's patch. Run these after producing the two adapted measurement trees:

```sh
(cd /workspace/scratch/upstream && npm ci --ignore-scripts --no-audit --no-fund > /workspace/scratch/99f-upstream-packages.log 2>&1)
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/00-setup/adapt.cjs /workspace/scratch/upstream > /workspace/scratch/99f-upstream-setup.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-f/adapt.cjs /workspace/scratch/upstream > /workspace/scratch/99f-upstream-adapt.log 2>&1
(cd /workspace/scratch/upstream && tsc -b src/server > /workspace/scratch/99f-tsc-build.log 2>&1)
(cd /workspace/scratch/upstream && tsc -p src/compiler > /workspace/scratch/99f-tsc-compiler.log 2>&1)
(cd /workspace/scratch/upstream && tsc -p src/services > /workspace/scratch/99f-tsc-services.log 2>&1)
(cd /workspace/scratch/upstream && tsc -p src/server > /workspace/scratch/99f-tsc-server.log 2>&1)
```

Scratch evidence: 99f-measurement.json contains the multiset comparison and exact removal; 99f-types-{before,after}.stderr retain every complete diagnostic chain; 99f-verify.json records stock byte equality, no-write idempotence, and refusal fixtures; 99f-verify-adapted.json records complete adapted-tree byte equality and no-write idempotence; 99f-tsc.json records all three project exit statuses. The verification scripts are 99f-compare.py, 99f-verify.cjs, and 99f-verify-adapted.cjs in the same scratch directory. No lane manifest is changed.

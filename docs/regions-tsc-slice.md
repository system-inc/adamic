# tsc graph-region measurement slices

Research for #9bd15w8, 2026-10-08. No compiler, runtime, adaptation, fixture or source-body changes.

## Finding

**The two-function emit-annotation slice allocates census-listed cycle-capable types, but does not contain an allocation loop. It does not yet satisfy the requested hot-loop measurement criterion.** Its two lifecycle functions encounter two observed NotYet roots, plus two cast refusals; that count describes helpers, not a complete emitter workload. Binding-pattern conversion remains a candidate with an explicit element-mapping allocation loop. A real transformer loop can exercise emit annotations, but adds source and dependencies beyond the proposed lifecycle slice.

There is **no proven smallest natively buildable file set** in these records. Counts below rank observed declaration/file attempts, not successful native slices. The lowering census bypasses a rejected whole-entry checker gate and rolls back failed statements. Removing the listed roots can expose more roots. A zero-root unit is not evidence of a complete native build or of graph selection.

## Frozen evidence and method

- Research branch starts at `origin/area/runtime`, `c1c6073021e8e5cd6005fcdee59eb3b377395fd8`.
- Named-fetch root-table branch: `origin/codex/stage3-notyet-table`, e8c283b5ed32477805357b652a170b85a04b2469. Used its **morning after run**, not the older 6,791-root table: 6,703 roots, compiler scratch merge `0e5661e4243af95ae3247d066de14ccb2b582a94`, 81 adapted source records. [Report](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/TABLE.md), [roots](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/after/roots.csv), [annotated attempts](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/after/full.jsonl.gz), [source manifest](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/source-manifest.json).
- Restored, named-fetch tsc cycle branch: `origin/codex/tsc-cycle-census`, `fcb7451a8239723fd903ae5998f54d2a86715e1f`. [Cycle census](https://github.com/system-inc/adamic/blob/fcb7451a8239723fd903ae5998f54d2a86715e1f/docs/stage3-tsc-cycles.md) inventories **1,685 cycle-capable declared/container slots: 80 natural back edges, two fresh-relaxation cases and 1,603 hard cases**. It is a conservative checker type-graph census, explicitly **not** a successful run of Adamic’s whole-program finder/fresh-write proof on tsc. The restored census replaces the earlier missing-evidence limitation; the oracle/stage 1 `runtime/cycles-corpus` inventory and import-cycle probes are not the relevant tsc type census.
- The region rule is in [memory.md](memory.md): unproven cycle-capable writes seed holder/target types; reachable types and containing collections join their graph component. Source evidence below identifies candidate cycles; it does not substitute for executing that rule on a checker-clean slice.
- For each file record in annotated JSONL, intersect NotYet findings with exact `(kind, where, reason, text)` signatures in `after/roots.csv`, stripping only the scratch-tree prefix. Deduplicate signatures. This removes proven echoes and includes roots reported in called helpers outside the attempted file. For a selected function, filter findings by its recorded `unit` first. Refused and panic findings are separate blockers, not added to the NotYet count. All observed blockers for the two candidate file attempts appear below.
- Ownership is a proposed work allocation: lowering/specialization/representation proofs and native-emission gaps are **compiler**; source variance repair is **typescript**; a missing runtime primitive would be **runtime**. These attempts identify no root requiring a new graph-region allocator primitive. Existing nested-function and boxed-union work may already retire older census findings; no retirement is credited without a replay on that implementation.

## Candidate ranking

| Core source file | Own-file NotYet roots | Roots hit including called helpers | Why consider it |
| --- | ---: | ---: | --- |
| `factory/nodeConverters.ts` | 14 | 16 | Emits nodes while mapping binding-pattern elements; two additional roots in `core.ts`. |
| `factory/emitNode.ts` | 28 | 30 | Creates cyclic emit annotations, tracks them on the source file, and disposes them in a loop; additional roots in `core.ts` and `debug.ts`. |
| `binder.ts` (comparison) | 238 | 257 | Real Symbol/declaration and FlowNode allocation paths, but substantially more roots. |
| `emitter.ts` (comparison) | 505 | 534 | Full emitter traversal is substantially further away. |
| `checker.ts` (comparison) | 1,124 | 1,212 | Type-relation work is hot in the recorded profile, but the full checker is furthest away. |

The lowest allocation-helper count is `factory/baseNodeFactory.ts`: **10 roots**, all in that file. It is supporting code, not a qualifying measurement by itself: it has neither a traversal loop nor writes that demonstrate a graph seed. Smaller files such as semver/path and read-only node walking likewise do not establish a graph-allocation workload merely by having few roots.

### 1. Emit annotations: graph types confirmed, allocation loop absent

Keep `factory/emitNode.ts:getOrCreateEmitNode` and `disposeEmitNodes`, with the concrete Node/SourceFile/EmitNode declarations from `types.ts`. In the pinned upstream fixture, `getOrCreateEmitNode` at line 36 stores `{ annotatedNodes: [node] }` on a SourceFile at line 43, appends other nodes at line 47, and stores `{}` on a node at line 50. The explicit cycle is **SourceFile.emitNode → EmitNode.annotatedNodes[] → SourceFile**. Other nodes also reach the source file through parent links. `disposeEmitNodes` at line 62 loops over annotated nodes and clears `node.emitNode` at lines 71–72. This is a particularly direct region-lifetime workload: dispose annotations, drop outside roots, then repeat a real transformation traversal. Disposal alone allocates nothing and must not be benchmarked as the allocation loop.

The emit setter calls occur throughout node-factory/transformer traversal. Repeated annotation is a proposed measurement entry, not an already measured hot allocation rate; the profile in [stage3-tsc-profile.md](stage3-tsc-profile.md) does not give an isolated annotation-loop frequency. The source-level cycle is stronger evidence here than a speculative claim about runtime samples.

#### Check against the restored cycle census

**Types: yes. Hot allocation loop in the proposed two-function slice: no.** The census supplies these specific records (D/K IDs are stable within the pinned report):

| Allocated/participating type | Census record and declaration site | Allocation or role in this slice |
| --- | --- | --- |
| `EmitNode` | D938 `EmitNode.annotatedNodes`, `types.ts:8315`; witness `Node[] → Node.emitNode → EmitNode` | `getOrCreateEmitNode` allocates the `{ annotatedNodes: [node] }` object at `factory/emitNode.ts:43`, and `{}` at line 50, both asserted as `EmitNode`. D938 is a hard `(c)` case. |
| `Node[]` | K69 mutable `Node[]` element slot; also the target of D938 | `[node]` at `factory/emitNode.ts:43` allocates the annotated-node array; line 47 appends nodes. The array has cycle-capable element storage, not just a cycle-capable owner. |
| `Node` and its `SourceFile` view | D11 `Node.parent`, `types.ts:948`; D13 `Node.emitNode`, `types.ts:950` | These are **inputs/owners**, not allocations by these two functions. The SourceFile branch stores the new `EmitNode` and array back on the same `node` included in that array. |

The source-level cycle `SourceFile → emitNode → annotatedNodes[0] → SourceFile` agrees with D938 and does not depend on the broader structural-view witness chosen for D13. Additional hard EmitNode slots include D939 `commentRange` (`types.ts:8318`), D940 `sourceMapRange` (8319), D941 `tokenSourceMapRanges` (8320), and D944 `typeNode` (8327). They are evidence about the type’s reachability, not claims that this small workload populates each field.

`getOrCreateEmitNode` has no traversal loop; repeated calls on an already annotated node return its existing metadata. `disposeEmitNodes` has a loop, but only assigns `undefined`. Neither its name nor its loop makes it an allocating hot path. Clearing annotations and calling the helper repeatedly in a new synthetic driver would be a benchmark we invented, not an existing tsc source loop. **Withdraw the standalone lifecycle slice as a qualifying hot-loop recommendation.** It is still a small helper/proof target with the two roots listed below.

There is a concrete repeated allocation path in existing tsc emitter code: `transformers/es2015.ts:addDefaultValueAssignmentsIfNeeded` loops over `node.parameters` at line 1909. For an initialized identifier parameter, it calls `insertDefaultValueAssignmentForInitializer` (line 1988), which creates an assignment, block and if statement and applies `setEmitFlags` at lines 1992, 1996 and 2017. `factory/emitNode.ts:setEmitFlags` at lines 92–95 calls `getOrCreateEmitNode`. For newly synthesized nodes without metadata, this reaches its `EmitNode` allocation at line 50 **inside the parameter-processing loop**. The other loop branch, `insertDefaultValueAssignmentForBindingPattern`, applies `setEmitFlags` to a newly created VariableStatement at line 1944 or ExpressionStatement at line 1965.

This larger workload allocates `EmitNode` metadata on new AST nodes, rather than merely reading annotations on existing inputs. The census also lists `BinaryExpression.left/operatorToken/right` (D216–D218, `types.ts:2650–2652`) and `ExpressionStatement.expression` (D332, `types.ts:3386`) as cycle-capable slots, covering concrete AST allocations in these branches. Together with D938/K69/D13, these identify the relevant reference types. The census still cannot certify the exact regions a pruned native program will select.

The expanded measurement therefore needs the ES2015 transformer’s parameter-processing path, `factory/nodeFactory.ts`/base allocator, emit setters, and their dependencies; it is **not** the two-root lifecycle slice. A function-heavy ES2015-target emit workload with initialized/destructured parameters will repeatedly reach that source loop. Its hotness has not been measured: the existing profile primarily checks with `--noEmit`, and the census explicitly says that mode skips most transforms. Do not call the helper slice a measured hot-loop workload, or reuse its two-root count for this expanded candidate. No new dependency-closure/root ranking is claimed for that expansion.

Observed roots in these **two functions only**:

| Site | Root/blocker | Owner |
| --- | --- | --- |
| `factory/emitNode.ts:46:100` | NotYet: a void call used as a value (`Debug.fail` on the fallback path) | compiler |
| `debug.ts:213:28` | NotYet: a value of type unknown (called assertion/debug path) | compiler |
| `factory/emitNode.ts:43:40` | Refused: a cast the runtime can't check (`{ annotatedNodes: [node] } as EmitNode`) | compiler |
| `factory/emitNode.ts:50:25` | Refused: a cast the runtime can't check (`{} as EmitNode`) | compiler |

`reading sourceFile` at 47:33 is a downstream echo of the failed initializer, not another root. `disposeEmitNodes` has no observed root of its own; this does not certify its dependencies. The rest of the whole-file roots are listed in the appendix because keeping that entire file also brings generic setters and range/comment/helper operations.

Supporting source declarations live in `utilitiesPublic.ts` (`isParseTreeNode`, `getParseTreeNode`), `utilities.ts` (`getSourceFileOfNode`), `debug.ts`, and `types.ts`; helper operations used by the complete file also reach `core.ts`. A real traversal and a source-file provider are still needed. Do not call this six-file set a closed executable slice: whole-file imports retain the compiler barrel and initialization graph. Member gathering is required before claiming a minimal complete file set.

### 2. Binding-pattern conversion: explicit allocation loop

Keep `factory/nodeConverters.ts:createNodeConverters`, especially `convertToArrayAssignmentPattern`, `convertToObjectAssignmentPattern`, their element converters and recursive assignment-target conversion. In the upstream fixture, lines 151 and 164 call `map(node.elements, ...)`; `core.ts:320–329` implements that map with an indexed loop. Each binding element can allocate a SpreadElement, BinaryExpression, SpreadAssignment, PropertyAssignment or ShorthandPropertyAssignment (nodeConverters lines 102, 108, 122, 126 and 129). The enclosing ObjectLiteralExpression/ArrayLiteralExpression then owns the result array. `setOriginalNode` attaches original-node links; parent/child and emit-annotation links make these Node allocations candidates for the graph component in the full compiler.

This has a loop directly in its allocation path. It is **not yet proven that pruning to the converters preserves an unproven cycle seed**: original links alone can be acyclic. A future slice must retain a real parent/backlink or emit-annotation-producing traversal and run `findCycles`; replacing it with an artificial cycle would measure a fixture rather than tsc.

Supporting allocation path: `factory/nodeFactory.ts` → `factory/baseNodeFactory.ts` → allocator constructors in `utilities.ts`; element mapping and casts reach `core.ts`, assertions reach `debug.ts`, and `setOriginalNode` is in `factory/nodeFactory.ts`, where its `mergeEmitNode` path copies emit annotations; other node-factory operations reach `factory/emitNode.ts`. Concrete declarations are in `types.ts`; predicates and text-range helpers also come from `utilities.ts` and `utilitiesPublic.ts`. These **nine implementation/declaration files** are a starting dependency list, not a verified minimum closure. NodeFactory constructs a large closure-valued factory, so retaining `createNodeFactory` whole is expensive: that file alone has 594 own-file roots in this census.

This is why the ten-root base allocator cannot be ranked as a native graph-loop slice. Its dynamic constructor-value operations block all five construction paths; their complete list is included below. Preserving a real factory, rather than supplying a replacement allocator, is part of measuring tsc itself.

## Build and measurement limits

The original 78/79-checker-failure observation is historical. The later integration report `stage3/meter/runs/20261008T023735Z.XREV1N/integration-report.md` records 2/79 whole-program and 54/79 own-file success; neither establishes a native graph-allocation slice. The root-table after run is a separate checker-rejected-program observation with 81 file records. These denominators must not be mixed.

The import census found a 76-file compiler SCC; [the gathering tool](../stage3/slice/README.md) preserves evaluation imports and export facades even when declarations are pruned. A plain subset of whole files is therefore not the proposed executable. These records cannot prove that one candidate has fewer **total transitive executable** blockers than another. They establish the local ranking and exact observed blocker inventory, and identify the annotation lifecycle as a small graph-typed helper core. The restored census confirms its types, but the hot-loop check above rules it out as a standalone measurement workload; an emitter traversal must be included and ranked separately.

No native build, native performance run, actual Adamic region selection, or root replay was performed. The restored census was checked against the source allocation and loop paths. No runtime/compiler fixes are proposed as already complete.

## Complete observed blocker inventories

All locations below are in the **adapted census source**, with prefix `src/compiler/` unless shown otherwise. Allocation descriptions above use the vendored upstream fixture; adaptations can shift lines. Each NotYet site is a root, not an echo. Refusals are deduplicated by the same exact signature; panic/error pairs representing the same failure are described once. A grouped row lists **every site** for its exact reason.

### factory/nodeConverters.ts

16 NotYet roots; 45 distinct Refused signatures.

| NotYet exact reason | Sites | Owner |
| --- | --- | --- |
| a call through ?. (an optional call) | `factory/nodeConverters.ts:66:13`<br>`factory/nodeConverters.ts:84:13` | compiler |
| a function inside a function (a closure) | `factory/nodeConverters.ts:118:5`<br>`factory/nodeConverters.ts:135:5`<br>`factory/nodeConverters.ts:147:5`<br>`factory/nodeConverters.ts:160:5`<br>`factory/nodeConverters.ts:173:5`<br>`factory/nodeConverters.ts:54:5`<br>`factory/nodeConverters.ts:63:5`<br>`factory/nodeConverters.ts:82:5`<br>`factory/nodeConverters.ts:98:5` | compiler |
| a template interpolating an object, an array, a map, a function or undefined | `core.ts:1786:59` | compiler |
| a void call used as a value | `factory/nodeConverters.ts:64:32` | compiler |
| an array of never | `core.ts:323:18` | compiler |
| passing union of differently held members to a function value | `factory/nodeConverters.ts:126:53`<br>`factory/nodeConverters.ts:129:49` | compiler |

| Refused exact reason | Sites | Owner |
| --- | --- | --- |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | `utilities.ts:10645:6`<br>`utilities.ts:10655:6` | typescript |
| an object refinement using an open numeric enum as a literal tag | `factory/nodeConverters.ts:100:17`<br>`factory/nodeConverters.ts:101:34`<br>`factory/nodeConverters.ts:102:106`<br>`factory/nodeConverters.ts:102:81`<br>`factory/nodeConverters.ts:102:89`<br>`factory/nodeConverters.ts:102:96`<br>`factory/nodeConverters.ts:104:65`<br>`factory/nodeConverters.ts:105:20`<br>`factory/nodeConverters.ts:108:62`<br>`factory/nodeConverters.ts:109:25`<br>`factory/nodeConverters.ts:111:21`<br>`factory/nodeConverters.ts:120:17`<br>`factory/nodeConverters.ts:121:34`<br>`factory/nodeConverters.ts:122:109`<br>`factory/nodeConverters.ts:122:84`<br>`factory/nodeConverters.ts:122:92`<br>`factory/nodeConverters.ts:122:99`<br>`factory/nodeConverters.ts:124:17`<br>`factory/nodeConverters.ts:125:69`<br>`factory/nodeConverters.ts:126:108`<br>`factory/nodeConverters.ts:126:167`<br>`factory/nodeConverters.ts:126:203`<br>`factory/nodeConverters.ts:126:213`<br>`factory/nodeConverters.ts:126:86`<br>`factory/nodeConverters.ts:128:30`<br>`factory/nodeConverters.ts:129:105`<br>`factory/nodeConverters.ts:129:127`<br>`factory/nodeConverters.ts:129:137`<br>`factory/nodeConverters.ts:129:91`<br>`factory/nodeConverters.ts:129:99`<br>`factory/nodeConverters.ts:151:63`<br>`factory/nodeConverters.ts:152:21`<br>`factory/nodeConverters.ts:154:17`<br>`factory/nodeConverters.ts:157:21`<br>`factory/nodeConverters.ts:164:62`<br>`factory/nodeConverters.ts:165:21`<br>`factory/nodeConverters.ts:167:17`<br>`factory/nodeConverters.ts:170:21`<br>`factory/nodeConverters.ts:55:35`<br>`factory/nodeConverters.ts:56:63`<br>`factory/nodeConverters.ts:57:39`<br>`factory/nodeConverters.ts:59:28` | compiler |
| overload 1 of assertNode parameter test cannot be served by implementation parameter test | `debug.ts:294:82` | compiler |

Additional compiler blocker: `core.ts:325:13`, statement panic (`invalid memory address or nil pointer dereference`) in attempts rooted at `nodeConverters.ts:147:5` and `160:5`. Both error records are wrappers for those two panic observations, not two extra NotYet roots.

### factory/emitNode.ts

30 NotYet roots; 7 distinct Refused signatures.

| NotYet exact reason | Sites | Owner |
| --- | --- | --- |
| ?.[] on a value | `factory/emitNode.ts:148:12` | compiler |
| a field of type string &#124; number &#124; undefined | `factory/emitNode.ts:234:12` | compiler |
| a function returning T | `factory/emitNode.ts:102:17`<br>`factory/emitNode.ts:113:17`<br>`factory/emitNode.ts:123:17`<br>`factory/emitNode.ts:139:17`<br>`factory/emitNode.ts:154:17`<br>`factory/emitNode.ts:175:17`<br>`factory/emitNode.ts:190:17`<br>`factory/emitNode.ts:199:17`<br>`factory/emitNode.ts:204:17`<br>`factory/emitNode.ts:212:17`<br>`factory/emitNode.ts:217:17`<br>`factory/emitNode.ts:221:17`<br>`factory/emitNode.ts:249:17`<br>`factory/emitNode.ts:258:17`<br>`factory/emitNode.ts:326:17`<br>`factory/emitNode.ts:333:17`<br>`factory/emitNode.ts:339:17`<br>`factory/emitNode.ts:351:17`<br>`factory/emitNode.ts:362:17`<br>`factory/emitNode.ts:373:17`<br>`factory/emitNode.ts:81:17`<br>`factory/emitNode.ts:92:17` | compiler |
| a generic function as a value | `core.ts:220:112` | compiler |
| a value of type T | `factory/emitNode.ts:346:45` | compiler |
| a value of type unknown | `debug.ts:213:28` | compiler |
| a void call used as a value | `factory/emitNode.ts:46:100` | compiler |
| assigning a field of a value | `factory/emitNode.ts:308:9` | compiler |
| storing string &#124; number in a field | `factory/emitNode.ts:242:5` | compiler |

| Refused exact reason | Sites | Owner |
| --- | --- | --- |
| a cast the runtime can't check | `factory/emitNode.ts:43:40`<br>`factory/emitNode.ts:50:25` | compiler |
| a value of type Node seen as Node &#124; SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | `factory/emitNode.ts:133:45` | typescript |
| a value of type Node seen as Node &#124; TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | `factory/emitNode.ts:184:43` | typescript |
| a value of type Node &#124; SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | `factory/emitNode.ts:133:12` | typescript |
| a value of type Node &#124; TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | `factory/emitNode.ts:184:12` | typescript |
| a value of type never[] seen as (SourceMapRange &#124; undefined)[], which can write SourceMapRange &#124; undefined where never is read | `factory/emitNode.ts:156:68` | typescript |

### factory/baseNodeFactory.ts

10 NotYet roots; 0 distinct Refused signatures.

| NotYet exact reason | Sites | Owner |
| --- | --- | --- |
| a function inside a function (a closure) | `factory/baseNodeFactory.ts:41:5`<br>`factory/baseNodeFactory.ts:45:5`<br>`factory/baseNodeFactory.ts:49:5`<br>`factory/baseNodeFactory.ts:53:5`<br>`factory/baseNodeFactory.ts:57:5` | compiler |
| new a ParenthesizedExpression | `factory/baseNodeFactory.ts:42:16`<br>`factory/baseNodeFactory.ts:46:16`<br>`factory/baseNodeFactory.ts:50:16`<br>`factory/baseNodeFactory.ts:54:16`<br>`factory/baseNodeFactory.ts:58:16` | compiler |

## Validation of this research

The analysis joined the pinned CSV to pinned JSONL twice, with independent set/SQL counting, checked the root totals, checked all sites rendered in these inventories, and confirmed that proven echoes were excluded. No packages or oracle fixtures were changed; no test packages, mutants or counts refresh were run.

Setup: `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`; sourced `/workspace/adamic-tools/env.sh`; Node v24.19.0. Setup completed in 9.596 s. Machine: Linux x86_64, AMD EPYC 7763, five visible CPUs, cgroup quota four CPUs, 17.6 GB memory. Setup load changed from 0.10/0.70/0.41 to 1.15/0.91/0.48; research load sample 0.09/0.39/0.36. These are setup/analysis observations, not graph-region performance measurements.

Follow-up validation: fetched `codex/tsc-cycle-census` by its exact refspec and verified its full SHA against `fcb7451a`; checked census records D11, D13, D216–D218, D332, D938–D941, D944 and K69; checked the first-use allocation, disposal-only loop and ES2015 parameter-loop/setter call chain against the pinned v6.0.3 fixture sources. Node remains v24.19.0. Documentation-only update; no performance measurements or package tests were run.

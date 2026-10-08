# Cycles: brief for the step 06 ruling

Draft for #7g4qv2b, measured 2026-10-08. Recommend a **mix**: a Program-lifetime region for the independently owned compiler graph, checked weak references only for backlinks with a keeper contract, ordinary counted ownership for acyclic caches and escaping owners, and graph regions for dynamic cyclic graphs whose lifetime is not one Program. Statement arenas remain useful for proven temporary trees. This is a proposed ruling, not an implementation of Program regions or inferred Weak.

Base: `area/runtime` at `cdfa22555589194e6f3133d986dd110061aafaa9`. Named-fetch evidence: [cycle census, fcb7451a](https://github.com/system-inc/adamic/blob/fcb7451a8239723fd903ae5998f54d2a86715e1f/docs/stage3-tsc-cycles.md), [slice report, 4ddb9eb3](https://github.com/system-inc/adamic/blob/4ddb9eb362d5a7f79140fbe5fdaeaa6241d171fa/docs/regions-tsc-slice.md), and [allocation report, e4113bfd](https://github.com/system-inc/adamic/blob/e4113bfdb734c894c50a6743bd8bc8998d589875/stage3/performance/allocation/evidence/REPORT.md). All refer to stock TypeScript 6.0.3, source `050880ce59e30b356b686bd3144efe24f875ebc8`. The built graph-region contract is in [memory.md](memory.md#regions-for-cyclic-graphs).

## 1. Shapes and counted slots

The census has **1,685 positive declaration/container records: 80 proposed natural backlinks, two fresh-construction cases, 1,603 hard cases**. It models possible type-graph reachability; it is not a native whole-tsc cycle-finder run, an allocation count, or a list of 1,603 observed refusals. Inheritance refinements and distinct container instantiations can repeat a physical storage schema. Unknown/generic declarations (71) and lexical reference leads (2,591) remain outside that total.

The disjoint regrouping below is reproducible with [group-census.py](cycles-decision/group-census.py); [census-groups.csv](cycles-decision/census-groups.csv) preserves every record, category, rule, site and witness. Classification uses the declared holder, census ownership rule and immediate table witness, not any occurrence of “Type” in an AST name. The six requested kinds do not exhaust the census: its child fields, callbacks, structural views and other containers must remain visible.

| Kind | Records | Natural backlinks | Fresh | Hard | Representative declared sites |
|---|---:|---:|---:|---:|---|
| `node.parent`, including refinements | 76 | 76 | 0 | 0 | D11 `Node.parent`, types.ts:948; D36 `ComputedPropertyName.parent`, 1798 |
| `symbol.declarations`, `valueDeclaration` | 2 | 0 | 0 | 2 | D633–634 `Symbol.declarations/valueDeclaration`, types.ts:6040–6041 |
| Symbol tables and SymbolLinks | 54 | 1 | 0 | 53 | D16 `LocalsContainer.locals`, types.ts:969; D635–636 `Symbol.members/exports`, 6042–6043; `SymbolLinks` and instantiated table slots in CSV |
| Type/signature caches and checker links | 120 | 2 | 0 | 118 | D708 `Type.checker`, types.ts:6442; fresh/regular literal links and instantiated/mapped types in CSV |
| Flow nodes | 32 | 0 | 0 | 32 | D18 `FlowContainer.flowNode`, types.ts:975; D117 `endFlowNode`, 2070; K72 `FlowNode[]`; debug flow edges |
| Emit nodes and `emitNode` owner slots | 15 | 0 | 0 | 15 | D13 `Node.emitNode`, types.ts:950; D938 `EmitNode.annotatedNodes`, 8315; D939–941 ranges, 8318–8320 |
| Other connecting slots | 1,386 | 1 | 2 | 1,383 | AST children/original links; NodeLinks; callback environments; broad structural views; other mutable containers |
| **Total** | **1,685** | **80** | **2** | **1,603** | |

The remaining natural backlink is `SymbolTrackerImpl.context`; the other three outside AST parents are `Symbol.parent`, `Type.checker` and `Signature.checker`. Across the original rule partition the census counts TREE 454, CACHE 210, CB 504, G 157, K 150, VIEW 96, P 76, FL 28, SELF 4, CK 2, F 2, SP 1 and CT 1. These overlap the semantic descriptions above only through explicit assignment to one group, not by summing two inventories.

Relation memo tables themselves deserve a correction to the question's premise: checker.ts:2385–2390 constructs six `Map<string, RelationComparisonResult>` tables. Their keys and results are scalar; these tables do not own Type objects and add **zero direct cyclic key/value slots**. The enclosing checker, cached types, signatures and callback environments can cycle. The relation mirrors below exercise reciprocal cached types with an independent keeper, rather than claim those six scalar tables contain Type pointers.

## 2. Mechanisms and lifetime contracts

**Checked weak references.** Each read checks the target's liveness; a handle has allocation/ownership costs as well as the read check. An optional/probing read yields absent after the target dies. A read whose target was proven present panics rather than reading freed storage. Weak does not own its target. Ordinary AST roots strongly own children, so a parent can be weak while that root lives. A detached child, synthesized parent or retained node from an old Program needs an explicit parent keeper if subsequent parent reads must succeed. Weakening both declaration/symbol directions, table values or the sole owning type cache loses that keeper. Type/checker debugging backlinks fit the same rule. Names alone do not prove ownership.

**Graph regions, built today.** A graph store merges holder and target regions before publishing the pointer. Union by member count and path compression find the current root; intrusive lists concatenate on merge. Stores inside a region omit per-edge retain/release. Locals, returned values, globals and other owners outside it contribute outside counts. The last outside release frees every member, invalidates weak targets and releases owned non-graph contents. A merge never splits: overwriting or deleting an edge can leave unreachable members allocated until the region ends. Metadata persists for losing union-find records as well. This handles symbol/declaration inverses, flow loops, emit annotation cycles and caches that really own newly created objects without guessing a weak direction. A Program is currently only an outside owner, not a special allocator or a registry of every allocation.

**A Program-lifetime region.** Allocate the selected Node/Symbol/Type/Signature graph and its connecting storage into one explicit Program owner. Nothing in that region is freed before its owner ends. Interior pointers then require neither weak liveness checks nor dynamic union-find merges; the implementation must establish membership, handle non-region owned contents and give each escape an owner of the region. Merely reference-counting an escaped interior object while freeing its region would be a use-after-free. An escape must retain the region or be copied/promoted with its reachable graph. Keeping the entire region for one escape has a retention cost.

The fetched allocation report strongly favors this lifetime for the **core graph**:

| Observed Node/V8 measure | Compiler `src/compiler --noEmit` | mitt |
|---|---:|---:|
| Highest observed heapUsed | 547.66 MiB | 116.84 MiB |
| Check complete, Program rooted, post-GC heap | 422.20 MiB | 58.16 MiB |
| After compilation frame/Program release, post-GC heap | 4.78 MiB | 4.44 MiB |
| Inspector estimated cumulative allocation, including collected | 3,690.78 MiB | 127.43 MiB |

The peak is the highest observation at 256-constructor intervals and boundaries, not a continuous maximum. The heap readings are observed bytes in different profiler modes. Cumulative allocation is a **sampling estimate**, not exact constructor bytes or a native region-size measurement. It explains why putting every transient array, string and helper allocation into the Program region is a materially different proposal from grouping the compiler graph.

For the compiler, **exact constructor calls** are parse Nodes 1,084,449; bind Symbols 122,977, Types 83, Signatures 4; check Nodes 42,411, Symbols 142,166, Types 107,337, Signatures 42,622. These count constructor entries, including objects that die before the next boundary. Array/Map/String total calls are N/A, not zero.

**Snapshot survival**, separately, shows 1,074,373 Nodes, 260,647 Symbols, 104,655 Types and 41,548 Signatures alive after checking, all zero after Program release. Every parse Node observed at the parse boundary (1,061,497) survives checking. The check-born boundary cohorts likewise survive to the end. Their measured shallow self bytes exclude table/backing storage. NodeLinks, SymbolLinks and flow literals are classified as Other in that profile; their individual survival cannot be claimed from those constructor totals. Snapshot IDs establish the observed cohort survival, not a proof that all future host API escapes obey one lifetime.

Thus the measured answer to “how big is that for a tsc run?” is: the stock compiler workload has a 422.20 MiB post-GC rooted heap and 547.66 MiB observed peak, versus a 4.78 MiB released heap. These are useful observed scale and lifetime evidence, **not a prediction of native Program-region bytes**. Native layout, exact cumulative core-graph bytes and escaped-region retention have not been measured.

**File/check arenas.** An arena is safe when its allocations cannot escape its ending scope, including through a closure, table, returned node or another file's symbol. AST trees often start file-local, but global symbol merging, imports, declarations, checker caches and synthesized nodes cross files. A per-file arena cannot be freed independently while those edges remain. A per-check arena fits relation scratch/temporary traversal storage only if no result, flow node, diagnostic or persistent type links back into it. Promoting/copying an escape needs a concrete graph policy. Today's compiler has statement arenas for proven nonescaping allocations; it does not have general cyclic file/check arenas or a Program allocator. Graph allocations are excluded from statement arenas. There are consequently **no measured alternative-allocator costs** for those unbuilt mechanisms, and no arena/Program mirror is presented as though it implemented them.

| Group | Weak fit | Graph-region fit | Program-region fit | File/check arena boundary |
|---|---|---|---|---|
| AST parents | Good with rooted tree and detached-node contract | Correct fallback; whole connected tree retained | Strong fit for ordinary CLI graph; plain interior parent pointers suffice | Cross-file/synthetic/escaped nodes need region owners |
| Symbol declarations/valueDeclaration | Only with independently owned declarations and symbols | Good for merged symbols and detached inverse graphs | Strong fit while all declarations and symbols belong to the Program | Merging/global exports break file-local release |
| Tables/SymbolLinks | Keep owning table entries strong; weaken proven backlinks only | Good for lazy synthetic values and cyclic links | Strong candidate; profile does not isolate SymbolLinks survival | Cached resolved/synthetic values escape a local lookup |
| Types/caches/relations | Checker backlinks fit; weak sole-owner caches do not | Good for reciprocal literal types and owning type caches; scalar relation tables remain counted | Core Type/Signature lifetime is directly supported by snapshots | Scalar relation scratch may fit; returned types/instantiations do not |
| Flow nodes | AST references may be weak; no uniform weak antecedent direction | Good for loop/reduction/restore topology | Candidate, with separate Flow/Other lifetime evidence needed | Flow is stored on AST and used after binding |
| Emit nodes | Weak annotated-node references need independently kept synthetic nodes | Good for SourceFile → annotation → SourceFile cycles | Suitable for a bounded CLI emit epoch; noEmit profile does not measure emit | Dispose annotations before ending emit arena; exported transformed nodes escape |

## 3. Today's fixture costs

Linux x86_64, AMD EPYC 9V45 96-Core Processor, five visible CPUs, cgroup CPU quota `400000 100000`, memory limit 16 GiB. Go 1.27.1, clang 20.1.8, Node v24.19.0. Setup used `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh --wasi-sdk`, then `/workspace/adamic-tools/env.sh`. No AGENTS.md was present.

The opt-in `TestCyclesDecisionMeasurements` builds unchanged native release binaries with `-O2 -flto=thin`, no count/sanitizer flags. Three fresh processes run sequentially per fixture; compilation is excluded. A small C launcher forks/execs the binary and obtains **that child's** `wait4` peak RSS, avoiding the Go test process's inherited RSS high-water mark. Wall time is measured around the launcher and includes process startup and output. Report the best wall run and its RSS, not independently selected RSS. Millisecond fixtures are dominated by startup/noise and do not establish a mechanism speed ranking. Every sample and exact flags are saved in [measure.log.gz](cycles-decision/measure.log.gz). Load before/after is recorded there; the table below is generated from that final run.

Counts come from separate `-DADAMIC_COUNT -O2` builds with an 8 MiB stack. Peak counts heap values, not bytes. Graph-region diagnostics measure liveness/reachability **at teardown**; U is the largest unreachable payload byte count in one region, not a process-wide garbage peak. Metadata is the largest single region's teardown metadata. The counted reachability walk is diagnostic work and is absent from release timing.

Measurement load: before `0.17 2.22 2.15 1/154 18579`; after `1.17 2.34 2.19 4/158 20509`. No idle-machine claim is made.

| Mirror/workload | Retains | Releases | Peak values | Regions | Merges | U bytes | Metadata bytes | Best seconds | RSS KiB |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| cycles_weak_parent.a | 3004 | 3006 | 3004 | 0 | 0 | 0 | 0 | 0.001626282 | 896 |
| cycles_graph_parent.a | 1003 | 4005 | 3003 | 1001 | 2001 | 0 | 96096 | 0.001332414 | 1024 |
| cycles_graph_relations.a | 3003 | 7007 | 4004 | 1 | 2001 | 400 | 32096 | 0.001450380 | 1024 |
| cycles_graph_symbols.a | 2006 | 4009 | 2004 | 1 | 1002 | 0 | 16112 | 0.001410050 | 896 |
| cycles_weak_relations.a | 4004 | 6007 | 6004 | 0 | 0 | 0 | 0 | 0.001762950 | 1152 |
| cycles_weak_symbols.a | 3008 | 3010 | 2005 | 0 | 0 | 0 | 0 | 0.001431611 | 896 |
| million.a | 950001 | 1950004 | 1000003 | 1 | 999999 | 3199936 | 16000064 | 0.063565697 | 78976 |

The million-node graph allocates/frees 1,000,003 heap values. Its one merged region has 1,000,000 members and 64,000,000 payload bytes; 950,001 are reachable (60,800,064 bytes), leaving **49,999 unreachable members / 3,199,936 bytes** retained until teardown. Region metadata is **16,000,064 bytes**. Counts report 999,999 merges. This is a measured retention witness, not an estimate of tsc's SCC sizes.

The six new mirrors allocate 1,000 children/declarations or 1,000 reciprocal type pairs. Parent mirrors retain a root and its child array; symbol mirrors keep a table, symbol and declaration array; relation mirrors keep a roots array plus a replaceable type cache entry. The roots inventory is essential in the weak relation case: the cache alone does not keep the weak counterpart alive. It also means this benchmark deliberately does not model expiry from evicting that inventory. Graph backing storage participates in regions, so unreachable region bytes can include replaced storage, not only source-visible objects.

[All 63 workload measurements](cycles-decision/measurements.md) include every registered `graph_regions_*` fixture, the counted million fixture, registered weak fixtures and all four `weak/` expiry probes. [CSV](cycles-decision/measurements.csv) also preserves allocations/frees and statement-region counts. The narrowed weak probe intentionally exits 70 and its counts stop at the panic; it is not a leak-clean normal teardown. Ordinary graph/weak fixtures are leak-checked; the weak expiry tests independently pin their intentional Node/native differences.

Each new mirror is registered in the normal oracle, held to source Node and emitted JavaScript, ASan/UBSan/LeakSanitizer checked, and counted in regenerated [counts.md](../internal/oracle/counts.md). `TestCyclesDecisionEdgeMutants` makes a source mutant deleting the observed weak/region edge, compiles it sanitized, and proves its output disagrees with the **unchanged** source Node oracle. The originals agree, so the mutation checks do not succeed merely because a baseline is already broken. These are edge-loss semantics mutants, not mutations of the allocator implementation; existing graph-region tests cover runtime merge/free faults.

For an additional fresh stock-tsc observation, the saved [API profiler](cycles-decision/tsc-profile.cjs) checks the existing tiny hello and whole parser roots with repository options, prelude and standard libraries, TypeScript 6.0.3 and Node 24.19.0. Both checks report zero diagnostics. Tiny (71 files) sampled 71,861,280 allocated bytes, with 43,183,984 post-GC rooted heap bytes, 30,058,776 after dropping Program and 135,336 KiB peak RSS. Parser (80 files) sampled 138,753,080 allocated bytes, with 59,910,440 rooted heap bytes, 37,209,768 after dropping Program and 248,748 KiB RSS. Sampling includes collected objects at 32 KiB interval; these are estimates, and startup/profiler/service caches remain in the released heap. Those API runs are not the stock-CLI constructor/survival experiment in e4113bfd and are not a native region benchmark. Their raw JSON is saved with the evidence. To repeat the API probes, set `TYPESCRIPT_PACKAGE` to the unpacked npm TypeScript 6.0.3 package directory and run `node --expose-gc docs/cycles-decision/tsc-profile.cjs INPUT` from the repository root; INPUT is either `internal/load/testdata/0.1/compile/01_hello.ts` or `stage1/typescript/parser/main.ts`.

**Measurement boundary:** there is no implemented Program-region allocator, general file arena or check arena in this base to benchmark against these six mirrors. The brief reports the measured built alternatives and names that gap instead of substituting predicted numbers or labeling a graph-region anchor as a Program allocator. The slice report also establishes no closed natively executable tsc hot loop: emit lifecycle helpers allocate cycle-capable types but alone contain no allocation loop. No native tsc size or throughput claim is made here. An allocator-level comparison and mirrors for those unbuilt alternatives remain outstanding.

## 4. What stays refused

Data cycles now select graph regions; the census's hard category is not a reason to revive the old blanket cycle refusal. Weak's explicit checked-present expiry remains a runtime refusal:

```text
adamic: panic: a weak reference was read after what it pointed to was freed
```

Exit 70, observed and held by `TestWeakReadsUndefinedOnceFreed`. A probing read instead returns absent. Nothing here authorizes silently weakening an owning cache, skipping that check or accepting a read through an expired target.

Cross-task graph publication remains refused. Today's actual compile-time message from `concurrency/refused/graph_region.a:11:13` is:

```text
Adamic 0.1 refuses parallelMap items are not shareable: Ring.next is a mutable field; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
```

This test passed and pins the field and remedy. Graph sharing/concurrent merges, watch mode and language-service version lifetime are outside the built graph contract. A proposed Program region does not override those refusals or make a mutable cyclic graph shareable.

The fetched slice's recorded source blockers remain evidence boundaries, not fresh claims of a complete compiler refusal census: `a void call used as a value`, `a value of type unknown`, and `a cast the runtime can't check` for `{ annotatedNodes: [node] } as EmitNode` and `{} as EmitNode`. Their exact sites are in the pinned slice report. No unsupported original-tsc lowering is certified by these hand-written mirrors.

## 5. Recommendations and questions for the ruling

| Group | Proposed choice |
|---|---|
| AST parents | Program-region interior pointers for a fully registered ordinary AST graph; otherwise checked weak parents with an independently retained tree owner. Detached/synthetic/escaped nodes retain an owner or use graph regions. Keep child edges owning. |
| Symbol declarations/valueDeclaration | Program-region strong interior edges for ordinary binder/checker symbols and declarations. Use graph regions for independently escaping or merged graphs until Program membership is established. Do not weaken declaration lists or valueDeclaration as a schema-wide policy. |
| Symbol tables/SymbolLinks | Keep tables as owners. Put Program-owned symbols and connecting links/storage in the Program region; keep scalar/non-graph contents counted. Prove or register synthetic targets. Graph regions cover mixed/dynamic ownership; weaken only selected proven backlinks such as symbol parent. |
| Type/signature caches and relations | The constructor and snapshot evidence directly supports Program lifetime for core types/signatures. Preserve cache ownership, including fresh/regular reciprocal types. Keep scalar relation memo maps counted or in a proven check-local arena. Checker backlinks can be checked weak outside a common region. |
| Flow nodes | Use built graph regions now. A Program region is a plausible simpler owner once flow allocation membership and post-bind reads are explicit; constructor counts for Types do not prove FlowNode lifetime. Never weaken all antecedents. |
| Emit nodes | Use graph regions for dynamic annotation/source cycles now. Consider an emit-epoch region only after disposal and transformed-node escape contracts are defined and an actual emitting workload is measured. The stock noEmit profile cannot decide this lifetime. |

The e4113bfd evidence changes the default for the **CLI core graph** toward one Program owner: the observed graph largely survives the whole check and dies together at release. Dynamic union-find machinery is useful for unknown/dynamic topology, but within a single established Program lifetime it can pay merge and metadata costs without enabling earlier freeing. A blanket all-allocation Program arena is not recommended: the 3,690.78 MiB sampled cumulative allocation includes large transient populations, whereas the observed rooted heap is 422.20 MiB. Neither number predicts native bytes. Reference counts at **region escape boundaries** preserve owner lifetime; they cannot independently reclaim an interior object from an arena.

Only the ruling can settle these policy choices:

1. Is the supported contract a one-shot CLI Program, or must the first implementation also support independently retired Programs, watch/service versions and escaped API values?
2. Is Program membership explicit in allocation/IR, or inferred? Which NodeLinks/SymbolLinks, synthetic declarations/types, flows, callbacks and backing stores must be registered? Does an escaping member retain the whole region, or require copying/promotion?
3. May two Programs share graph members? If yes, does that merge lifetimes, prohibit independent teardown, or require isolation/copying? What retained-byte limit justifies the selected escape policy?
4. Should ordinary parent/checker backlinks inside a proven common region remain plain, with checked Weak reserved for different lifetimes, or should all backlinks pay a uniform read check? What is the contract for a detached child's parent?
5. Are relation scratch and emit metadata separate epochs, and which lifetime boundary ends each one? Does the first ruling authorize new allocator mechanisms and their mirror measurements, or select built graph regions pending that evidence?

These are ownership and product-boundary decisions, not claims the profiling tools can infer. The brief supplies measured built costs and the strongest available Program-lifetime evidence, while leaving unbuilt allocator costs explicitly open.

## Reproduction and validation

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk
source /workspace/adamic-tools/env.sh
python3 docs/cycles-decision/group-census.py
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/cycles_' -v -count=1
go test ./internal/oracle -run '^(TestGraphRegionsCountsAndFree|TestGraphRegionsCompiledMillion|TestGraphParallelMapRefusesRegions|TestWeakReadsUndefinedOnceFreed|TestWeakRegionReviewFreesWithoutRegions|TestCyclesDecisionEdgeMutants)$' -v -count=1
ADAMIC_CYCLES_MEASURE=1 go test ./internal/oracle -run '^TestCyclesDecisionMeasurements$' -v -count=1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
```

All listed oracle runs passed; counts regeneration adds the six mirror rows. Compressed logs preserve those outcomes and the sampled stock API measurements. The complete repository gate was not run. Earlier measurement attempts with an unavailable time utility and inherited parent RSS were superseded by the saved small-launcher run; neither supplies published performance numbers.

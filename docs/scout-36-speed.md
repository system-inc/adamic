# Scout: step 36 (#qekjwz5)

Step 36's comparator is typescript-go's reported 1.78 s check. There is no
building native checker here yet, so this scout identifies likely costs and
experiments; it does not claim a native-checker CPU split or forecast its duration.
The strongest current evidence favors removing ownership traffic and using
profiles before assuming the allocator is the principal bottleneck. Maps need a
fresh checker measurement: parsing exercises a very different workload.

This is documentation only, based on `origin/area/runtime`
`cdfa22555589194e6f3133d986dd110061aafaa9`. No AGENTS.md existed in the repository
or workspace ancestors. The base and each evidence branch were fetched by name.
The parse-only driver is absent from this base; its source and parser/scanner
modules were extracted into scratch from `runtime/stage1-profiles-refresh`
`9e062aac7a55e117199c1cc5fbb7a0d59bedf07a`, then built with this base's compiler
and runtime. This is a measured source/runtime combination, not a measurement of
that branch's shipping binary.

## What today's builds say

The [profile refresh report](https://github.com/system-inc/adamic/blob/9e062aac7a55e117199c1cc5fbb7a0d59bedf07a/cloud/reports/stage1-profiles/report.md)
records 6,627,754,308 simulated instructions for plain shipped ThinLTO and
5,059,714,292 with the matching profile: **23.7% fewer instructions**, with a
measured **30.0% reduction in best user time**. Both match the independent Go AST.
The profile driver/build policy and strengthened retain wrapper are absent from
this runtime base. They are available on the named evidence branches.
These are the requested 6.628G and 5.060G anchors, not the counts-instrumented
builds below. The refreshed profiles were trained on 38 service files disjoint
by path and content hash from the 77 compiler benchmark files. Source, runtime,
compiler, flags and target fingerprints guard profile use; stale data falls back
to plain ThinLTO. A checker will require checker-specific training.

The [retain fast-path report](https://github.com/system-inc/adamic/blob/4e4bbda4dfeb8eda46e862f3e5f6eb348fe0c4d5/cloud/reports/retain-fast-path/report.md)
separately records 7,293,289,097 → 6,627,967,853 instructions, **9.12% fewer**.
Its hot slow dispatches were immortal strings, not unexpectedly graph-tagged
objects. The wrappers preserve instrumentation, graph/shared dispatch and owned
cell handling while keeping NULL, immortal and ordinary positive-count cases
inline. There were 20,852,831 immortal-string retains and 16,393,340 releases;
remaining last-reference releases still perform destruction. Those two experiments
use different endpoints: their savings must not be added or applied again to a
base already containing the improvement.

An older [disjoint parse instruction breakdown](../cloud/reports/release-lto/measurement.md#simulated-cachebranch-reranking)
gives a useful scale, but **is not the 6.628G binary's breakdown**. Reaggregating
its rounded self-instruction rows gives approximately 6,503.617 million:

| Cost | Legacy parse self instructions, millions | Share of that legacy total | Expected checker cost, inferred from available evidence |
|---|---:|---:|---|
| Retains | 331.008 | 5.09% | Repeated reads, arguments and ownership transfers remain substantial; borrowing and the fast path attack different parts of this cost. |
| Releases and child destruction | 566.150 | 8.71% | Temporary cleanup plus final graph destruction; outside-count operations remain even with a Program region. |
| Allocation entry/libc plus slabs | 269.241 | 4.14% | More graph objects and transient storage during binding/checking; pressure grows, but bytes or calls do not establish CPU share. |
| Strings: character indexing, equality, other operations and substrings | 1,476.742 | 22.71% | Scanner cost persists; checker name/key comparisons and diagnostics add work. |
| Maps | Not isolated | Unknown | Likely much more important during bind/check than parse. No supported numeric checker share can be derived here. |
| Explicit generated control: scanner, node construction, line table, parent map | 1,925.616 | 29.61% | Checker recursion, relation tests, inference and cache-control branches add substantial generated work. |
| Unassigned: mixed parser/arrays/libc/inline work, field/call plumbing and input decode | 1,934.860 | 29.75% | Includes costs belonging to several rows above; keep unresolved until source/inline attribution is available. |

The parent-map **44.350M is generated control**, not evidence for runtime Map
lookup cost. The unresolved row prevents an invented clean five-way split.
Self costs are additive; inclusive costs are not. The older report explicitly
found instruction ratios insufficient to predict CPU ratios. These shares are
therefore a parse proxy for choosing experiments, not future checker percentages.

## Allocation and lifetime evidence for a checker

[TypeScript allocation observations at e4113bfd](https://github.com/system-inc/adamic/blob/e4113bfdb734c894c50a6743bd8bc8998d589875/stage3/performance/allocation/evidence/REPORT.md)
measure stock TypeScript 6.0.3 on Node, not Adamic allocations or CPU cost.
For the compiler input:

- Parsing calls Node constructors 1,084,449 times; 1,061,497 parse-cohort nodes
  remain live through check. Binding adds 122,977 Symbols. Checking calls
  142,166 Symbol, 107,337 Type and 42,622 Signature constructors.
- After checking, 1,074,373 Nodes, 260,647 Symbols, 104,655 Types and 41,548
  Signatures remain live; **all four cohorts disappear after Program release**.
  This supports the step 06 Program lifetime, rather than short statement regions
  for the entire checker graph. It does not prove every temporary belongs there.
- Binding adds 95,497 live Maps and checking adds 45,605; 142,249 Maps are live
  after check. Parse alone adds 1,017. Map shallow bytes omit table storage, and
  neither live counts nor shallow bytes measure lookup frequency.
- Check adds 298,677 live strings. After check, strings occupy 45,776.0 KiB
  shallow self storage, and array backing storage occupies 89,921.8 KiB. Hashing,
  equality, copying, resizing and character indexing need separate CPU attribution.
- Snapshot-free sampling estimates 3,690.78 MiB including collected allocations;
  3,654.52 MiB is unclassified by constructor-associated stack attribution.
  Exact constructor calls, snapshot live objects and sampled bytes are distinct
  measurements. An exit-only profile misses much collected allocation.

The stock compiler Program includes a different loaded-file graph from the
77-file parse driver. Its constructor counts are not this scout's node denominator.

## Levers already built or designed

| Cost | Existing lever and status | Measured effect and limit |
|---|---|---|
| Retains/releases | Borrow inference in `internal/lower/borrow.go`; lent reads and strong-field/element/loop borrowing in `internal/native/borrow.go` and `element_borrow.go`. Ordinary inline wrappers in runtime `adamic.h`; the strengthened fast path is available on the named evidence branch. | This scout removes 79,929,379 pairs with borrowing enabled; fast-path report removes 9.12% of parse instructions without changing ownership. Eliminating calls and making remaining calls cheaper are separate interventions. |
| Allocation | Reuse in place, moves and consuming-parameter inference in `internal/native/reuse.go`; statement regions in `region.go`; cyclic graph regions in `graph_regions.go` and runtime `graph_regions.c`. | [Memory evidence](memory.md#reuse-in-place-perceus): `reuse_arrays.a` allocates 78 versus 84 with array reuse disabled; `normalize.a` loses 10 allocations. This is shape/liveness-specific, not a general checker speedup. |
| Program-owned graph | Step 06's Program lifetime is supported by the stock allocation cohorts. Current [dynamic graph-region design and implementation](memory.md#regions-for-cyclic-graphs) joins graph members and counts outside owners, with uncounted internal edges. | A Program is an outside owner, not an automatic arena for every allocation. Whole-Program/checker performance is unmeasured. Existing statement-region `trees` evidence removes about 133M retains and releases each when emission recognizes region values; this does not prove the same effect for checker graph regions. |
| Strings | Borrowing, owned append, shared substring views, immortal ASCII characters, UTF-16 indexing caches and equality fast paths already exist in runtime string files. Profiles also optimize generated scanner control. | [String-view report](../cloud/reports/string-views/report.md): parse instructions 6.5036G → 6.4937G across its complete change, despite a large service-workload improvement. Optimize the actual string operation; request-response results do not transfer to checker names. |
| Maps | Ordered typed hash-table runtime; borrowing on safe Map reads; reuse/moves reduce surrounding ownership, and graph containers join their graph contents. Profiles can specialize layout/inlining of hot callers. | No measured checker Map CPU improvement is available. Program regions remove graph ownership work, **not hashing, probing, table growth or string-key counting**. Do not claim a Map optimization from the parent-index array. |
| Generated code | Shipped ThinLTO and guarded stage 1 profiles; borrowed parameters and moves also simplify generated code. | Profile refresh: 23.7% fewer parse instructions, 30.0% less best user time. This includes runtime inlining effects; it is not exclusively generated-code savings. Train/revalidate on checker paths when they build. |

The [counts table](../internal/oracle/counts.md) is an ownership/allocation ledger,
not a CPU profile. At this base, `memory_examples/list.a` records 7 allocations,
6 retains and 16 releases; `memory_examples/strings.a` records 14/2/17;
`reuse_arrays.a` records 78/35/94. Counter definitions matter: runtime child
teardown can release contents without another counted `adamic_release` call, and
in-place replacement may increase counted releases while removing allocations.
Thus releases need not equal retains plus allocations.

## Measured now: ownership calls per constructed parse node

Workload: the parse-only `stage1/cohere/parse/parse.a` driver, all 77 compiler
files pinned at TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`.
Every content hash matched the profile branch's `stage1/profiles/benchmarks.json`.
The driver parses, constructs its line table and parent-index array, then prints
its findings count (`0\n`); it does not run lint rules. Its `--count` output is
**not** a node count.

Both variants count **889,146 constructed ParseNodes**, including nodes that may
not remain reachable from the final tree. This is a constructor-work denominator,
not the stock TypeScript Node constructor count and not total heap allocations.

| Variant | Retains | Releases | Retains/node | Releases/node | Combined calls/node |
|---|---:|---:|---:|---:|---:|
| Borrow inference disabled | 121,356,568 | 117,408,175 | 136.487 | 132.046 | 268.533 |
| Current borrowing enabled | 41,427,189 | 37,478,796 | 46.592 | 42.151 | 88.744 |
| Removed | 79,929,379 | 79,929,379 | 89.895 | 89.895 | 179.789 |

Retains fall **65.86%**, releases **68.08%**. Both variants allocate and free
**3,375,891** values, peak at **729,760**, and place **0** in statement regions.
This isolates a large ownership-call opportunity; it does not measure its CPU
savings. Counter hooks count calls even for NULL or immortal values.

Full AST mode was checked separately against the independent Go port:
**44,766,682 identical bytes**, SHA256
`5d77733457ef9e17b5707db65a3f5af230621ca74a10c8f3daa83720504346de`.
All three outputs match, not merely their findings count. AST printing itself
adds ownership and allocation work, so its counts are excluded from the table.

Method:

1. Run `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`;
   source `/workspace/adamic-tools/env.sh`. Node is v24.19.0. Extract the evidence
   driver's `stage1/cohere/parse`, `stage1/typescript/parser` and
   `stage1/typescript/scanner` trees with `git archive` into scratch. Obtain the
   pinned compiler sources and write sorted absolute `src/compiler/**/*.ts`
   paths to `compiler.txt`; verify all 77 benchmark hashes.
2. Load `parse.a`, run `lower.Lower`, emit `native.C`, and build with
   `native.Options{Count:true}`. This uses the repository's counted `-O2`
   policy, not shipped ThinLTO or a training profile. A temporary focused
   `TestScoutMeasure` helper provided those calls and the existing `goOracle`
   test helper built the independent Go adapter.
3. For the borrowing-off build, a Go overlay changes only two native emitter
   files. Immediately after the normal element-borrow, reuse and region plans
   are made, set `emitter.elementBorrows = nil` and clear every
   `program.Locals[index].Borrowed`. In `borrow.go:value`, replace
   `e.lendable = e.depth == outerLendAt || e.borrowChain(expression)` with
   `e.lendable = false`. Clearing parameter flags also disables borrowed-call
   argument paths. Keep the baseline reuse/region plans, including their
   conservative lending exclusions, so newly owned locals do not gain reuse.
   Library ABI borrowing and runtime internal ownership remain unchanged.
4. In each generated C file, add a plain `static size_t scout_nodes` and increment
   it on entry to both `ParseNode_new` and `ParseNode_new_in`; print it from a
   destructor with `fprintf(stderr, "nodes=%zu\n", scout_nodes)`. There are exactly
   two matching constructor definitions. Rebuild counted, run
   `--manifest compiler.txt --count`, and divide the runtime counters by that
   constructor count. The probe makes no Adamic values or ownership calls.
5. Run unprobed binaries as well: all six runtime counters exactly equal their
   probed counterpart. Repeat the four counts runs in reverse variant order:
   results are identical. Run both probed variants with `--ast`; compare full
   output with the Go adapter's `--whole`. Remove the temporary test helper;
   overlay, generated C, binaries and outputs remain in `/tmp/scout36` only.

Machine: `3ee19de862f2`, AMD EPYC 9V74 KVM guest, Linux 6.18.44 x86-64,
five visible CPUs, four-CPU quota (`cpu.max=400000 100000`), 17.6 GB reported
memory. Go 1.27.1, clang 20.1.8. Load before the measurement sequence:
**1.69 / 1.29 / 0.69**; after: **1.58 / 1.28 / 0.70** (1/5/15-minute averages).
Deterministic operation counts require neither a best-of-five timing selection
nor hardware-counter access. No timing result is claimed for these builds.

## First three experiments when the checker builds

1. **Establish the full-check baseline and test profiles.** Pin the checker,
   complete compiler input graph, flags and machine. Separate parse, bind,
   check and Program teardown; collect `perf record -g`, self-time and inline
   source attribution, plus a separate counted build. Reconcile retains,
   releases/destruction, allocation, strings, Maps, generated code and external
   work without double-counting callees. Compare shipped ThinLTO with fresh
   checker profiles trained on disjoint inputs, using five interleaved runs and
   best-of-five results against the pinned typescript-go comparator. Keep full
   diagnostics/output identical, not just a checksum, and state machine/load.
2. **Ablate borrowing and remaining wrapper dispatch.** Repeat the controlled
   borrowing-on/off comparison per constructed node and per checked source node;
   record the two denominators. Classify remaining wrapper calls as NULL,
   immortal, ordinary, graph/shared and last-reference. Compare the current fast
   path against the earlier wrapper in a separate scratch runtime snapshot.
   Attribute self cycles and generated caller changes; never infer savings from
   call counts alone. Check that allocation/free/peak counts and full output
   remain equal; exercise existing lifetime and graph/shared guards for any fix.
3. **Test Program ownership and transient reuse separately.** Trace which native
   Nodes/Symbols/Types/Signatures and graph containers join the Program region,
   which escape, and which temporary arrays/strings can be reused. Compare the
   supported region implementation with a semantics-preserving counted baseline;
   record outside-count traffic, joins, retained bytes, final teardown and
   allocation sites. Do not region-own every sampled allocation. Retain the full
   Map probe/hash/growth and string-operation breakdown to choose the next runtime
   intervention if those costs dominate. Require leak-clean Program release,
   identical outputs, and lifetime mutants for any new ownership correctness path.

No production source, fixture, mutant or recorded count row was added or changed;
`counts.md` therefore needs no regeneration. The focused scratch build checks,
corpus-hash assertions, repeat-count checks, probe neutrality and full Go AST
comparison passed. This scout makes no claim that step 36 is already achieved.

# Scout: step 36 (#qekjwz5)

Step 36's comparator is typescript-go's reported 1.78 s check. There is no
building native checker here yet, so this scout identifies likely costs and
experiments; it does not claim a native-checker CPU split or forecast its duration.
The strongest current evidence favors removing ownership traffic and using
profiles before assuming the allocator is the principal bottleneck. Maps need a
fresh checker measurement: parsing exercises a very different workload.

The initial scout was documentation only, based on `origin/area/runtime`
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

At the initial scout checkpoint, no production source, fixture, mutant or recorded
count row was added or changed; `counts.md` needed no regeneration then. The focused scratch build checks,
corpus-hash assertions, repeat-count checks, probe neutrality and full Go AST
comparison passed. This scout makes no claim that step 36 is already achieved.

## Follow-through: source research and the first built shape

The original measurements above are preserved. This follow-through applies the
previously measured wrapper optimization to this branch, adds independent guards,
and leaves the wider checker design and language questions explicit. It does not
claim a working native tsc checker.

### Where step 36 bites in tsc

Stock source pin remains `050880ce59e30b356b686bd3144efe24f875ebc8`.
[scout-36-sites.json](scout-36-sites.json) records every matched file, content hash
and line for these queries. [scout-36-sites.py](scout-36-sites.py) reproduces it:

```sh
python3 docs/scout-36-sites.py /tmp/scout36/corpus --check
```

The census is lexical and static, excludes the queried function declarations and
interface signature, and counts multiple occurrences on a line separately.
`.get`/`.set` receivers are not resolved to Map types, and `.parent` includes
writes. None of these totals proves hotness or safe borrowing.

| Cost/source shape | Concrete tsc source | Static sites and relevance |
|---|---|---|
| Literal/undefined ownership | [scanner.ts:414](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/scanner.ts#L414): `return tokenStrings[t]`; [419](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/scanner.ts#L419): `textToToken.get(s)`; [1821](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/scanner.ts#L1821): keyword lookup. | 56 `tokenToString` call-shaped occurrences: checker 26, parser 18, program 6, emitter 3, utilities 2, binder 1. The primitive result is often static text or undefined, yet an owned read/return can still ask the runtime to count it. |
| Interned string/type ownership and Maps | [checker.ts:2052](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L2052) declares `stringLiteralTypes`; [20305–20308](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L20305) returns a hit or creates/stores/returns a type. | 24 `getStringLiteralType` call-shaped occurrences, all in checker.ts. This cache itself has one `.get` and one `.set` site. The checker has 90 `new Map` sites, 156 `.get` and 137 `.set` occurrences across receivers. Returning a cached object preserves identity, including after the caller keeps it. |
| Allocation | [factory/nodeFactory.ts:1209](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/factory/nodeFactory.ts#L1209) delegates base-node creation; [checker.ts:5517](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L5517) constructs a Type and gives it an id. | 116 `createBaseNode` call-shaped occurrences in nodeFactory.ts. The allocation cohorts above measure the dynamic scale; static factory sites do not multiply directly into allocation counts. |
| Strings | [scanner.ts:3096](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/scanner.ts#L3096) and [3208](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/scanner.ts#L3208) switch on two-unit slices, each with `TODO: don't use slice`. | 12 `.slice` occurrences in scanner.ts. Adjacent code converts numeric-literal substrings and scans class-set operands; replacing a slice with byte reads would need UTF-16 and boundary proofs. |
| Generated recursive control and field reads | [checker.ts:22759](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L22759) starts `isRelatedTo`; [23324](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L23324) probes its relation cache; [2938](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L2938) creates/reuses node links. | The checker has 1,259 `.parent` property occurrences. Relation-cache access has two `relation.get` sites and three `relation.set` sites. Field reads and call control require inline source attribution, not blanket classification of all checker instructions as generated arithmetic. |
| Program graph and cycles | [checker.ts:20276](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L20276) sets `regularType` to itself or the supplied type; [20283](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/checker.ts#L20283) constructs fresh/regular back-links. | Eight calls to `getFreshTypeOfLiteralType` in checker.ts; self-links and cross-links make ordinary object-by-object counting insufficient. These full types are not the acyclic LiteralType fixture. |

### Node, typescript-go and the specifications

Node v24.19.0 reports V8 `13.6.233.17-node.51`. At the same Node release,
[V8 String::Equals, string-inl.h:535](https://github.com/nodejs/node/blob/v24.19.0/deps/v8/src/objects/string-inl.h#L535)
uses an identity fast path, rejects distinct internalized strings as unequal, and
otherwise performs `SlowEquals`. That is an implementation fast path for primitive
value equality, not permission to compare arbitrary separately allocated strings
by address. [Heap::CollectGarbage, heap.cc:1558](https://github.com/nodejs/node/blob/v24.19.0/deps/v8/src/heap/heap.cc#L1558)
shows Node's collector. tsc's cache code contains no retain/release calls: V8 keeps
reachable values alive and can collect unreachable cyclic graphs.

A direct observation on this machine, with `node --expose-gc --trace-gc`, creates
an object containing `['na','me'].join('')`, caches it by that string, retrieves it
using a separately constructed equal string, clears the Map, and calls
`global.gc()`. It prints **`true name undefined`** after one Mark-Compact record.
The caller's object/string survive; the deleted lookup does not. This supports the
fixture's answer, not a storage or collection guarantee Adamic can copy.

The repository's typescript-go source pin is the vendored fork
`cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`; these observations
are source evidence, not a reproduction of the comparator's 1.78 s measurement:

- [scanner.go:2263](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/scanner/scanner.go#L2263)
  indexes a fixed `[ast.KindCount]string` reverse table; a missing spelling is
  Go's empty string, whereas the stock JS function returns undefined.
  [scanner.go:277](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/scanner/scanner.go#L277)
  returns a source substring with `s.text[s.tokenStart:s.pos]`. Go source positions
  here use its representation; they are not a license to substitute UTF-8 byte
  positions for JS UTF-16 indices.
- [checker.go:25839](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/checker/checker.go#L25839)
  interns `*Type` values in `map[string]*Type`, returning the existing pointer on
  a hit. [25849](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/checker/checker.go#L25849)
  explicitly treats NaN separately because Go's normal NaN map lookup misses.
- [checker.go:681](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/checker/checker.go#L681)
  has Symbol, Signature and IndexInfo arenas; [core/arena.go:15](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/core/arena.go#L15)
  allocates fresh backing batches, growing to 256 elements. Returned pointers
  preserve old batches through Go's reachability rules. LiteralType construction
  at [checker.go:25577](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/checker/checker.go#L25577)
  uses `&LiteralType{}`; not every Type allocation is arena-backed.
- Fresh/regular cycles are retained in
  [checker.go:25821](https://github.com/system-inc/TypeScript/blob/d92d9bfee114c80be2c375d72edae966176e3a4f/tsc/internal/checker/checker.go#L25821).
  The implementation relies on Go memory management, rather than replacing
  those edges with Weak. Adamic's design must solve their ownership itself.

The normative answers relevant to these shapes are:

- [ECMAScript String type, §6.1.4](https://tc39.es/ecma262/multipage/ecmascript-data-types-and-values.html#sec-ecmascript-language-types-string-type):
  strings are sequences of 16-bit code units, including ill-formed surrogate
  sequences. Representation optimizations must preserve those values and existing
  aliases; concatenation cannot change the caller's prior string value.
- [Map.prototype.get](https://tc39.es/ecma262/multipage/keyed-collections.html#sec-map.prototype.get)
  returns the stored value or undefined. The current specification canonicalizes
  the key then compares with SameValue; the resulting key behavior is the
  familiar SameValueZero treatment of NaN and signed zero. String keys compare
  by value, and returned objects retain object identity.
- [Map.prototype.set](https://tc39.es/ecma262/multipage/keyed-collections.html#sec-map.prototype.set)
  replaces an equal key's value or appends an entry;
  [Map.prototype.delete](https://tc39.es/ecma262/multipage/keyed-collections.html#sec-map.prototype.delete)
  removes that entry. Deletion is not an operation that invalidates another
  live language reference to its former value.

These algorithms constrain visible behavior. They do not require V8's collector,
Go's arenas, Adamic's counts or a particular physical string representation.
Node remains the executable oracle for every fixture.

### Design inside Adamic's existing rules

The first built intervention changes only `internal/native/runtime/adamic.h` and
imports the measured wrapper structure from `runtime/retain-fast-path`, rather
than claiming a new algorithm. Retain/release still count every requested call.
They first return for NULL. They inspect shared/graph flags **before any plain
reference-count access**. Ordinary positive counts use the existing inline path;
zero-count non-cell values return without entering the slow helper. A zero-count
cell still dispatches to its environment owner. Last-reference release still
enters the existing destruction path. Both wrappers are forced inline under the
shipped clang policy. No compiler call is elided, no new lifetime is inferred,
no graph is traced in production, and no collector is introduced.

This rule is safe without a language ruling: static strings already have zero
counts and permanent storage; statement-region values already live until region
end. Shared values and graph members bypass the new zero-count test. An interior
cell's zero is not immortality, which is why the cell exception is mandatory.
The two new fixtures exercise primitive lookup and an acyclic intern cache using
existing owned returns. They do not require borrowed returns or a new Program API.

For the remaining step, preserve the existing ownership model while measuring:
borrow only stable roots across proven harmless operations; retain an owned
result before its cache/container can disappear; reuse only a consumed unique
value whose identity no other holder can observe; keep graph-internal edges in
the supported dynamic region and preserve every outside owner. Match a profile
to the full source/runtime/target fingerprint before consuming it. Derive a
checker CPU budget from self and inline costs before choosing the next change.

The hard cases are cache deletion/replacement while a caller holds the result;
unknown callbacks or virtual writers during field/element reads; closure cells
whose owner survives through an interior reference; graph/shared flags whose
counts cannot be read plainly; fresh/regular self- and back-links; Program
versions with external owners; UTF-16 surrogate halves in slice/key operations;
NaN/signed-zero keys; and identity-bearing Type/Symbol objects whose equal fields
do not make them interchangeable. Blanket immortality, blanket borrowed cache
results, Go empty-string sentinels or weak graph back-links would not preserve
these established Node answers.

### Questions for system_adamic

- Should a future borrowed-return convention be part of the language/API contract,
  or should all cache-returning calls continue to return owned values even when
  the compiler knows their Program owner?
- If Program-scoped allocation becomes a source-visible facility, how should a
  returned Type/Symbol, closure capture or host handle express the outside owner
  when it outlives the call that created the Program?
- Should any cross-Program cache be eligible for a Program-lifetime-only contract,
  and what should the contract require for overlapping Program versions or watch
  sessions?
- Should programmer-specified intern pools ever affect identity-bearing objects,
  or should additional interning remain limited to primitive value representations
  and already explicit tsc caches?
- If a future region API permits disposal, what should happen when a caller still
  holds a language reference: refusal, an explicit retained owner, or another
  specified contract that never frees that value early?

No ruling on these questions is assumed. Borrowed-return contracts, source-visible
Program disposal and new identity-bearing interning are not built in this change;
implementing them would require those language decisions. The existing wrapper
shape and fixtures require none of them.

### Fixtures, mutants and green implementation

| Fixture/control | Node-held behavior or invariant | Mutant actually rejected |
|---|---|---|
| [scout36_token_spellings.a](../internal/oracle/testdata/scout36_token_spellings.a) | Scanner-style reverse-array lookup and keyword Map lookup; missing index returns undefined; held spelling survives removal from the array. | Shift the lookup index by one: valid sanitized, leak-clean native execution finishes, but disagrees with original source Node output. |
| [scout36_literal_cache.a](../internal/oracle/testdata/scout36_literal_cache.a) | Checker-style equal-string interning returns the same object; distinct names differ; changing the source and deleting/clearing cache entries preserves held object/string values. | Omit `cache.set`: valid sanitized, leak-clean execution finishes, but identity/id output disagrees with original source Node. |
| `TestScout36FastPathDispatch` | 33 retains and 33 releases remain counted; NULL and 32 immortal-string operations invoke zero slow helpers. | Restore zero-count slow dispatch: the harness rejects exactly 32 retain and 32 release helper entries. This is a performance control, not a claim that the old behavior was semantically wrong. |
| `TestScout36OwnedCellMutant` | Retaining an interior cell keeps its environment alive after the original environment reference is released. | Treat zero-count cells as immortal: ASan reports heap-use-after-free. |

`internal/oracle/scout36_test.go` registers both fixtures with the shared Node,
native sanitizer/leak and JavaScript-backend oracle, and runs their semantic
mutants independently. `internal/native/scout36_test.go` builds the real edited
runtime for each dispatch/cell mutant, with ASan, UBSan and leak detection enabled.
The new control measures 64 slow-helper entries removed for the 32 immortal
pairs, while ownership-call counts stay unchanged. It is not a new throughput
measurement. Machine remains the EPYC/Linux machine recorded above; during the
follow-through checks, sampled 1/5/15-minute load was 1.63 / 1.16 / 0.84.

Passing focused commands:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -v \
  -run 'TestScout36SourceMutants|TestNativeAgreesWithNode/internal/oracle/testdata/scout36_'
go test ./internal/native -count=1 -v \
  -run '^TestScout36|^TestEnvironmentCellsCountTheirEnvironment$|^TestGraphMembersCountOnTheirRegion$|^TestRuntimeReleasePaths$|^TestReleaseSharedValueAndUnsafeMutant$'
go test ./internal/native -count=1 -v \
  -run '^(TestGraphRegionsRuntime|TestGraphClosureEnvironment|TestRegionEndWeakTargets|TestParallelMemory)$'
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1
python3 docs/scout-36-sites.py /tmp/scout36/corpus --check
```

The parallel-memory guard also passed its ASan, slab-ASan, counted, malloc and
TSan variants. Existing graph/shared/last-owner mutants remained caught. Counts
regeneration changes exactly two rows: token fixture **8 allocations / 8 frees /
22 retains / 29 releases / peak 5**; literal cache **21 / 21 / 32 / 42 / peak 10**.
Every pre-existing count row is unchanged. The ledger is committed separately,
with only `internal/oracle/counts.md` in that commit.

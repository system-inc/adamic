# Destruction diagnosis and runtime merge, October 7, 2026

Built: source-aware re-split, per-type destruction/lifetime counters and a bounded per-file region estimate; no performance fix yet.
Commits: runtime tip 6fc42a97ede1ac8244594783d056edbb369ebf61 merged as 3cbc640d4b0b04422f65c526ad8896ceb8497c24 into codex/parse-speed.
Commands and outputs: 77-file whole-tree byte parity PASS; uncached native/oracle PASS; destruction allocation/free reconciliation PASS.
Mutants: deliberately shifted destruction timing and a changed free count must fail reconciliation; their logs accompany this report.
Not covered: implemented per-file region, a region memory peak, per-type original inline destruction instruction attribution, or either dated target.

## Exact build configurations

All release Ir numbers and instruction estimates in this report use clang 20.1.8, `-O2 -g`, sanitizers off, `-ffp-contract=off`, `-fno-optimize-sibling-calls`, no LTO or counted-runtime hooks.
The original accepted baseline C/runtime is 9cc0d58. The merged measurement
regenerates C with 3cbc640's compiler and compiles against its runtime. Compiler
and runtime improvements are both included; this is not isolated cherry-pick
attribution. The exact commands, from the repository root, are:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/parse-speed/runtime-parse.c internal/native/runtime/*.c -lm -o /workspace/scratch/parse-speed/runtime-parse > /tmp/parse-speed-runtime-clang.log 2>&1
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I /workspace/scratch/parse-speed/destruction-counter-runtime /workspace/scratch/parse-speed/destruction-counter.c /workspace/scratch/parse-speed/destruction-counter-runtime/*.c -lm -o /workspace/scratch/parse-speed/destruction-counter > /tmp/parse-speed-destruction-counter-build.log 2>&1
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I /workspace/scratch/parse-speed/destruction-profile-runtime /workspace/scratch/parse-speed/destruction-profile.c /workspace/scratch/parse-speed/destruction-profile-runtime/*.c -lm -o /workspace/scratch/parse-speed/destruction-profile > /tmp/parse-speed-destruction-profile-build.log 2>&1
```

Callgrind uses the pinned 77-file compiler.txt and `--count`, with the same
Valgrind 3.24 binary and VALGRIND_LIB from SPLIT.md. Every stdout/stderr goes to
a file. The flags above also apply to each instruction table below. Counter
and clone probes are scratch-only, with no parser or emitter edits.

## Destruction happens per file

The accepted baseline allocates and destroys **3,377,171 heap values**.
**889,146 are ParseNodes**, all destroyed in the caller's cleanup after collect
returns. All 77 file-end observations have zero live nodes. No node dies during
collect, including speculative nodes; Parser.nodes keeps them until the file
ends. **1,197,451 other values** die during collect; **2,179,626 values** die in
per-file caller cleanup; **94 other values** die after the last file, chiefly
manifest/arguments and top-level maps. Node peak is 298,172.

A node does not own another node. Children are numeric table indexes, stored
in a numeric array. The separate parent map is another numeric array. The
Parser.nodes reference array owns the nodes; each node also owns its child
array and any heap string in text/raw/operator/semantic/kind. Thus the final
release traverses an ownership graph, not parent pointers in an AST.

## Cost by destroyed type

The original compiler inlines free_one into adamic_release, so its self costs
combine entry, queue draining, type dispatch, child handling and allocator
calls. For per-type attribution, the diagnostic has identical noinline copies
of free_one, selected by type and by the ParseNode shape. It preserves the
field walk, let_go callback, iterative queue, slab/malloc deallocation and
weak invalidation. **These are diagnostic body costs, not exact per-type
costs in the original inline binary.** Inclusive costs include each child's
reference drop/enqueue and allocator callees, but not processing queued child
destructors or the outer release's queue drain. No recursive inclusive totals
are added together. Clone dispatch/counter work is excluded from these rows.

| Type | Destroyed | During collect | Per-file cleanup | After files | Diagnostic body Ir | Ir/value |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| string | 693,374 | 293,034 | 400,256 | 84 | 45,636,729 | 65.82 |
| object | 201,142 | 200,756 | 385 | 1 | 48,236,247 | 239.81 |
| array_reference | 20,022 | 19,632 | 385 | 5 | 22,007,769 | 1099.18 |
| map | 79 | 0 | 77 | 2 | 12,829 | 162.39 |
| cell | 1,116 | 1,116 | 0 | 0 | 65,190 | 58.41 |
| closure | 34,750 | 34,750 | 0 | 0 | 2,334,676 | 67.18 |
| node | 889,146 | 0 | 889,146 | 0 | 205,667,730 | 231.31 |
| array_number | 1,537,542 | 648,163 | 889,377 | 2 | 165,465,955 | 107.62 |

Original self release/child bucket: 1,542,615,786 Ir, excluding weak/index/free
costs in other original buckets. The accepted original caller cleanup release
edges are **354,105,637 Ir for 539 calls**, including the cascaded file trees.
That is **22.95%** of the original release bucket, even though these edges also
contain some allocation/free costs outside that bucket. This observed inclusive
boundary is more dependable than scaling a global per-type average to cleanup.

External releases total **41,064,797**: null 3,529; immortal 16,386,567;
non-final counted 23,714,166; final counted 960,535. Children receive another
5,129,311 let_go callbacks, including 2,249,897 NULLs and 20,325 immortal
strings. Each final child is queued once; the queue is iterative, not entered
recursively per child. Object destruction walks shape reference flags at run
time, including non-reference fields; class objects walk the derived/base
layout once. Small nodes return one slab slot each, not one libc free each.
Numeric arrays release no numeric element, but each separately frees its element
buffer and header. This explains why releases far outnumber destroyed objects.

## Per-file region estimate, not a measured implementation

A region containing all nodes, their child-array headers/buffers and the file's
node table can remove the structural field walk, per-value queueing and individual
frees. Nodes plus numeric arrays still present at cleanup occupy **197,682,896
bytes** across the corpus, computed from object layout and array capacity at
actual death. Fixed 1 MiB chunks would need **233 chunks across 77 files** if
packed without extra fragmentation. Approximately tens of thousands of Ir
for those chunk releases would replace hundreds of millions of teardown Ir.
This chunk count is an optimistic capacity model: it omits abandoned array
growth buffers, payload alignment, string payloads and speculative scratch data.
It is not a measured peak or a proposed allocation policy.

Bulk teardown alone has an observed **354.1M Ir inclusive upper boundary**, about
23% of the original 1.54G bucket, not most of it. A broad region that also elides
all generated node and numeric-array reference-count calls could plausibly save
**0.55 to 0.65G release/destruction Ir** (roughly 36 to 42%), plus retains and
allocation savings outside that bucket. This is an optimistic instruction model,
not a promised speedup: 3,989,367 non-final node releases plus 4,608,920 non-final
numeric-array releases at the baseline's approximately 30-Ir non-final path cost
about 258M Ir; add about 0.30G of structural teardown, and allow remaining entry
and numeric-scratch destruction costs. Eliminating only counts on live tree
nodes is narrower than eliminating all numeric scratch arrays used while parsing.

The remaining immortal-kind releases and counted scanner/speculation/parser/
closure/string traffic are substantial. A much broader region that owns all
per-file scratch values and removes their counting calls could remove more,
but that needs a separate ownership proof for escapes, mutable arrays, strings,
external references and weak handles. It is not established just because the
tree itself is acyclic. Today's region_end still walks each object and releases
external children; routing nodes to the existing region is not one-stroke teardown.
No region implementation or early free is made here.

## Runtime merge, source-aware disjoint split

Whole process: **9,259,261,285 to 7,389,521,274 Ir**, saving 1,869,740,011 Ir
(**20.19%**); relative to the equality-fixed 9,234,837,655 Ir, saving
1,845,316,381 Ir. Same release flags above. Go parse baseline remains
1,684,637,174 Ir (Go 1.27.1 ordinary optimized build, sanitizers/race off,
GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1 under Callgrind), so merged native is
**4.3864x**, still above the 2.53G and 1.685G dated targets.

The original split mistakenly put the input.c function named decode (225.74M Ir)
in character indexing. It belongs to file input decoding. This version also
charges known inlined runtime header source ranges to their mechanism rather
than to the generated caller. Both before and after below use the same v2
rules; v1 buckets in SPLIT.md are retained as the historical observations and
are not silently compared to a differently classified after column. Equality
here includes kind and other equality bodies; the original separate kind probe
remains in SPLIT.md. The inline char/index row also includes lengths and slice
location, so it is not a cost per source character read. Unattributed header
line-zero instructions stay with their owning function/remainder.

| Disjoint self bucket, same release flags | Baseline v2 | Runtime merge v2 | Saved |
| --- | ---: | ---: | ---: |
| Releases and child destruction | 1,562,416,966 | 1,534,667,356 | 27,749,610 |
| Object field and call plumbing, including inline | 1,346,142,886 | 194,088,088 | 1,152,054,798 |
| Retains | 363,315,534 | 354,990,651 | 8,324,883 |
| String equality bodies, kind and other | 335,614,089 | 339,705,054 | -4,090,965 |
| Character reads and UTF-16 indexing, including inline | 1,240,515,174 | 960,751,669 | 279,763,505 |
| Remainder: parser, arrays, driver, libc, startup and unattributed inline | 1,544,732,689 | 1,446,962,638 | 97,770,051 |
| Scanner generated control | 1,834,614,732 | 1,585,314,417 | 249,300,315 |
| Line table generated control | 83,445,905 | 83,445,905 | 0 |
| File reading and UTF-8 input decode self | 225,789,370 | 225,789,370 | 0 |
| Allocation and freeing | 211,140,210 | 211,140,210 | 0 |
| Other string operations, including token values | 96,126,460 | 95,031,911 | 1,094,549 |
| Node construction generated control | 286,305,012 | 220,508,211 | 65,796,801 |
| Parent map generated control | 50,564,366 | 55,891,184 | -5,326,818 |
| Substrings | 78,537,892 | 81,234,610 | -2,696,718 |
| Total | 9,259,261,285 | 7,389,521,274 | 1,869,740,011 |

The field bucket is now chiefly inline frozen/layout guards, cached data-slot
loads, and object_callee interface lookup. Searches on cache misses cost just
4,942 Ir self in object_find. Runtime object_callee still costs 68,373,628 Ir
self, so repeated
method/call ownership and dispatch remains visible. Direct writes now avoid
the two large old runtime calls; this is the runtime area's emitter/runtime
work. No emission hook is added by parse-speed; compiler worker emission files
remain untouched.

## Validation and limits

Merged runtime: TestWholeCompilerAgrees PASS, 77 files and 44,766,682 identical
whole-tree bytes against typescript-go and Node, including ASan/UBSan native.
Uncached internal/native PASS 140.402s; internal/oracle PASS 135.305s. The full
repository gate was not run. The scratch destruction counter also runs with
`-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, all the same
strict C/warning/contraction/sibling flags, `ASAN_OPTIONS=detect_leaks=1`;
no sanitizer or leak errors and exit 0. Its counts are not instruction numbers.

Setup: Go/clang/Node/submodule ready 0s, cache warm 72s, done 72s; nproc 5,
cgroup cpu.max 400000 100000, memory 17.6 GB. Raw profiles, counters, summaries,
mutants and test logs accompany the report. No parser/scanner source edits.

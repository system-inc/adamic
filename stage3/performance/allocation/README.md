# TypeScript checker allocations

`run.sh <new output directory>` profiles stock TypeScript 6.0.3 on Node 24.19.0
checking TypeScript's own src/compiler and the accepted mitt library source.
All work in this unit is under stage3/performance/allocation. Base:
89ac4a8c1de0b02d95965be72f7f8bf1c92433d2. TypeScript commit:
050880ce59e30b356b686bd3144efe24f875ebc8; mitt commit:
6b41670516ed8e8b738612f60491995470aa63b3. npm package versions and SHA512
integrities are locked; Node, CLI, inputs and harness hashes are recorded.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/allocation-setup.log 2>&1
source /workspace/adamic-tools/env.sh # use the path setup prints
stage3/performance/allocation/run.sh /tmp/new-allocation-run > /tmp/allocation.log 2>&1
python3 stage3/performance/allocation/parse.py /tmp/new-allocation-run/fixture --fixture > /tmp/fixture.log 2>&1
python3 stage3/performance/allocation/parse.py /tmp/new-allocation-run/fixture --fixture --drop-constructor Node > /tmp/mutant.log 2>&1
```

Prerequisites: Linux, Node at the pinned version, npm, git, Python 3.12+, network
access, and enough memory/disk for large heap snapshots. This observation used a
16 GiB memory quota. Setup completed in 39.405s; nproc=5, CPU quota four CPUs,
AMD EPYC 9V74. The script installs dependencies with npm ci and fetches pinned
inputs. It generates TypeScript's diagnostic map with the upstream script,
uses the original compiler config, and disables composite/incremental state:
`-p src/compiler --noEmit --pretty false --incremental false --composite false`.
mitt checks src/index.ts with strict ES2020/NodeNext and skipLibCheck, plus the
stock default declarations. It is a source acceptance project, not its test
suite. Compiler options and full argv are in the captured command JSON files.

There are four observations per project:

1. The unmodified CLI runs with `--heap-prof --heap-prof-interval=16384`.
   This exit profile is sampled and omits collected allocations by default.
2. An observation-only copy of the pinned _tsc.js counts six constructor entry
   points and observes heapUsed every 256 constructor calls and at boundaries.
   Inspector sampling includes allocations collected by minor and major GC.
3. A separate process replays the high-water constructor event and writes a
   heap snapshot there. The snapshot itself collects garbage. Both the
   pre-snapshot heapUsed and post-GC heap are reported; this is an observed
   high-water event, not a proven continuous global maximum.
4. A separate lifetime run snapshots after Program creation (parse), after
   initializeTypeChecker (binding plus checker initialization), after diagnostics
   (check, Program rooted), and after the compilation frame unwinds (released).
   A baseline snapshot separates startup objects. Snapshot object IDs identify
   cohorts and survivors. Snapshots force GC and disturb natural GC timing.

All observed runs must match unmodified stock stdout, stderr and exit status.
All constructor counts must agree across the three hook modes. The hooks add
counters and boundaries in memory; they do not replace compiler decisions or
change input files. The original CLI SHA256 is checked before each load. Phase
labels are operation boundaries, not time attribution: lazily parsed or
synthetic nodes made during checking belong to check. Constructor names Node4,
Token and Identifier2 form Node; Symbol4, Type3 and Signature2 are separate.
NodeLinks, SymbolLinks, flow-node object literals and other objects stay Other.

**Exact cumulative constructor bytes and exact total Array/Map/String counts
are not recoverable from these Node profiler formats.** Sampling exposes call
stacks and weighted allocation sizes, not allocated-object constructors; a
snapshot only exposes objects surviving its GC. Dead-before-boundary builtins
are invisible to the census. N/A means unknown, not zero. Compiler constructor
calls are exact, while bytes in constructor tables are exact live shallow
self bytes. Maps exclude their storage; JavaScript Array objects and V8 backing
arrays are separate. String includes flat, concatenated and sliced strings.
Builtins include Node/V8 and harness objects, not only the compiler. Stack
attribution is explicitly heuristic; pruned sample node IDs remain unresolved
rather than being silently discarded. Sample records are not allocation counts.

The host-only `fixture.cjs` is a Node oracle witness, not an Adamic fixture:
parse creates 10 Nodes and 4 Symbols, retaining 6 and 2; bind creates 2 Nodes,
3 Symbols, 5 Types and 2 Signatures, retaining 1/1/3/1; check creates 1 Node,
2 Symbols, 7 Types and 3 Signatures, retaining 0/1/4/1. The totals are
13/9/12/5, with 7/4/7/2 alive after checking and zero after releasing roots.
The verifier also checks the predicted phase cohorts and their survival IDs.
The mutant drops Node entries in the real snapshot parser and is caught by
`fixture parse/Node: expected 6 live, got 0`, exit 1. It changes measured objects,
not a fixture string or expected value. No internal/oracle fixture was added,
so internal/oracle/counts.md is outside this unit and unchanged.

Observed design implication, separated from the measurements: the core syntax,
symbol and type graph mostly persists to end-of-check and then disappears when
the Program is released. A per-Program lifetime could fit those objects. A
blanket arena for everything risks holding much more transient data: the
snapshot-free compiler sampler estimates 3.60 GiB cumulative allocation versus
about 548 MiB observed peak. Neither V8 object sizes nor that sampled volume
predict native layouts or native memory usage. Temporary builtin lifetimes,
escape proofs and cycles still require analysis before choosing native regions.

Commands run: setup; run.sh /tmp/allocation-complete (exit 0); final compressed
fixture parsing and mutant subprocess (0 and 1); exact fixture cohort checks;
stock output/count/peak-event comparisons; final mitt observation-hook smoke;
Node --check, bash -n, Python AST parsing, and git diff --check. Logs are in
[evidence](evidence/). No whole-package tests or full gate were run.

Raw --heap-prof and snapshot-free inspector profiles, both replay peak snapshots,
and every final fixture snapshot are committed compressed. Large compiler/mitt
phase snapshots remain in /tmp/allocation-complete, with hashes committed in
[evidence/snapshot-hashes.json](evidence/snapshot-hashes.json); rerunning the
script regenerates the complete artifact set. The reader supports .gz inputs.
Collection-time harness hashes are preserved separately from final reader hashes.
The final reader adds compressed-input support and stronger fixture assertions;
all raw measurements and stock diagnostics were retained.

Not covered: emit/watch/build/incremental or language-service lifetimes; multiple
projects sharing compiler caches; an Adamic native checker; continuous peak
capture; exact total builtin allocations or bytes by constructor for collected
objects; native allocator headers or object sizes; repeat-run confidence ranges.

## Measured tables

Exact constructor calls and snapshot live objects are separate from sampled allocation estimates.

| Input | Observed peak heap MiB | Peak phase | Replay pre-snapshot heap MiB | Replay post-GC heap MiB | After check post-GC heap MiB | Released post-GC heap MiB |
|---|---:|---|---:|---:|---:|---:|
| compiler | 547.66 | check | 547.71 | 422.12 | 422.20 | 4.78 |
| mitt | 116.84 | check | 112.41 | 58.17 | 58.16 | 4.44 |

## compiler

| Phase | Constructor | Exact calls | New live at boundary | New live self KiB | Survive check | Dead before boundary | Gone by check | Survive release |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| parse | Node | 1084449 | 1061497 | 149593.3 | 1061497 | 22952 | 0 | 0 |
| parse | Symbol | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Type | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Signature | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Array | N/A | 163054 | 5095.4 | 162891 | N/A | 163 | 0 |
| parse | Map | N/A | 1017 | 31.8 | 1017 | N/A | 0 | 0 |
| parse | String | N/A | 95224 | 20552.8 | 95063 | N/A | 161 | 2 |
| parse | Array backing | N/A | 151925 | 8410.3 | 151762 | N/A | 163 | 1 |
| bind | Node | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| bind | Symbol | 122977 | 122977 | 16332.9 | 122977 | 0 | 0 | 0 |
| bind | Type | 83 | 83 | 5.2 | 83 | 0 | 0 | 0 |
| bind | Signature | 4 | 4 | 0.4 | 4 | 0 | 0 | 0 |
| bind | Array | N/A | 138135 | 4316.7 | 138133 | N/A | 2 | 0 |
| bind | Map | N/A | 95497 | 2984.3 | 95497 | N/A | 0 | 0 |
| bind | String | N/A | 3523 | 139.9 | 3512 | N/A | 11 | 0 |
| bind | Array backing | N/A | 233822 | 21947.8 | 233794 | N/A | 28 | 0 |
| check | Node | 42411 | 12781 | 1930.5 | 12781 | 29630 | 0 | 0 |
| check | Symbol | 142166 | 137670 | 18284.3 | 137670 | 4496 | 0 | 0 |
| check | Type | 107337 | 104572 | 6535.7 | 104572 | 2765 | 0 | 0 |
| check | Signature | 42622 | 41544 | 4219.3 | 41544 | 1078 | 0 | 0 |
| check | Array | N/A | 179660 | 5614.4 | 179660 | N/A | 0 | 0 |
| check | Map | N/A | 45605 | 1425.2 | 45605 | N/A | 0 | 0 |
| check | String | N/A | 298677 | 9665.5 | 298677 | N/A | 0 | 0 |
| check | Array backing | N/A | 207455 | 58469.9 | 207455 | N/A | 0 | 0 |

Whole live graph by requested constructor after checking (shallow self bytes):

| Constructor | Live count | Self KiB | Live after Program release | Self KiB after release |
|---|---:|---:|---:|---:|
| Node | 1074373 | 151538.1 | 0 | 0.0 |
| Symbol | 260647 | 34617.2 | 0 | 0.0 |
| Type | 104655 | 6540.9 | 0 | 0.0 |
| Signature | 41548 | 4219.7 | 0 | 0.0 |
| Array | 481143 | 15035.8 | 201 | 6.3 |
| Map | 142249 | 4445.3 | 46 | 1.4 |
| String | 426054 | 45776.0 | 14414 | 2569.0 |
| Array backing | 594423 | 89921.8 | 1206 | 807.3 |
| Other | 1555062 | 88463.0 | 49048 | 3587.2 |

Snapshot-free inspector sampling including collected objects: 3690.78 MiB estimated, 220905 sample records.
Unmodified CLI --heap-prof: 408.84 MiB estimated in the exit profile, 24007 sample records.

| Allocation stack attribution | Estimated MiB, including collected | Sample records |
|---|---:|---:|
| unclassified | 3654.52 | 218605 |
| Node-associated stack | 21.41 | 1362 |
| Symbol-associated stack | 7.08 | 451 |
| Signature-associated stack | 1.99 | 127 |
| Type-associated stack | 5.78 | 360 |

## mitt

| Phase | Constructor | Exact calls | New live at boundary | New live self KiB | Survive check | Dead before boundary | Gone by check | Survive release |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| parse | Node | 135372 | 130936 | 17998.4 | 130936 | 4436 | 0 | 0 |
| parse | Symbol | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Type | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Signature | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Array | N/A | 22322 | 697.6 | 22322 | N/A | 0 | 0 |
| parse | Map | N/A | 138 | 4.3 | 138 | N/A | 0 | 0 |
| parse | String | N/A | 12223 | 5472.1 | 12219 | N/A | 4 | 0 |
| parse | Array backing | N/A | 20725 | 1086.8 | 20725 | N/A | 0 | 1 |
| bind | Node | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| bind | Symbol | 30750 | 30750 | 4084.0 | 30750 | 0 | 0 | 0 |
| bind | Type | 85 | 85 | 5.3 | 85 | 0 | 0 | 0 |
| bind | Signature | 4 | 4 | 0.4 | 4 | 0 | 0 | 0 |
| bind | Array | N/A | 30818 | 963.1 | 30816 | N/A | 2 | 0 |
| bind | Map | N/A | 22561 | 705.0 | 22561 | N/A | 0 | 0 |
| bind | String | N/A | 2763 | 115.9 | 2759 | N/A | 4 | 0 |
| bind | Array backing | N/A | 53387 | 4926.9 | 53379 | N/A | 8 | 0 |
| check | Node | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| check | Symbol | 567 | 558 | 74.1 | 558 | 9 | 0 | 0 |
| check | Type | 369 | 369 | 23.1 | 369 | 0 | 0 | 0 |
| check | Signature | 137 | 133 | 13.5 | 133 | 4 | 0 | 0 |
| check | Array | N/A | 709 | 22.2 | 709 | N/A | 0 | 0 |
| check | Map | N/A | 184 | 5.8 | 184 | N/A | 0 | 0 |
| check | String | N/A | 976 | 32.1 | 976 | N/A | 0 | 0 |
| check | Array backing | N/A | 881 | 174.5 | 881 | N/A | 0 | 0 |

Whole live graph by requested constructor after checking (shallow self bytes):

| Constructor | Live count | Self KiB | Live after Program release | Self KiB after release |
|---|---:|---:|---:|---:|
| Node | 130962 | 18002.5 | 0 | 0.0 |
| Symbol | 31308 | 4158.1 | 0 | 0.0 |
| Type | 454 | 28.4 | 0 | 0.0 |
| Signature | 137 | 13.9 | 0 | 0.0 |
| Array | 54285 | 1696.5 | 201 | 6.3 |
| Map | 23012 | 719.1 | 46 | 1.4 |
| String | 44550 | 21029.9 | 14410 | 2378.9 |
| Array backing | 76381 | 7288.1 | 1207 | 807.7 |
| Other | 163523 | 13284.0 | 48846 | 3424.8 |

Snapshot-free inspector sampling including collected objects: 127.43 MiB estimated, 3783 sample records.
Unmodified CLI --heap-prof: 47.26 MiB estimated in the exit profile, 2485 sample records.

| Allocation stack attribution | Estimated MiB, including collected | Sample records |
|---|---:|---:|
| unresolved sample nodeId | 1.10 | 4 |
| unclassified | 120.89 | 3433 |
| Symbol-associated stack | 4.05 | 258 |
| Signature-associated stack | 0.02 | 1 |
| Type-associated stack | 0.06 | 4 |
| Node-associated stack | 1.30 | 83 |

N/A is not zero. Sampling carries allocation call stacks, not allocated-object constructors or exact total counts. Stack attribution is heuristic and is not constructor-byte accounting.

Cohorts are objects newly observed at a post-GC phase boundary. The end is checking complete with the Program still rooted; released is after its compilation frame has unwound. Birth self bytes need not equal end sizes. Backing arrays are V8 storage nodes, separate from JavaScript Array objects; Map storage is not included in Map self bytes.

Peak is the highest heapUsed observation at 256-constructor intervals and phase boundaries in a run without snapshots. The peak snapshot is replayed at the same deterministic constructor event and boundary location; replay GC and heap usage can differ. It is not proof of the continuous global maximum.

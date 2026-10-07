# Stage 3: tsc allocation profile on Node

Measured October 7, 2026, from Adamic main `ef3d907ecdc4c771b016f7d9c52372def057a340`, on `codex/tsc-allocation-profile`. No Adamic implementation changed. This measures the Node workload for the October 10 stage-3 target; it does not claim native results.

## Inputs and method

The target is **TypeScript 6.0.3**, tag `v6.0.3`, source commit `050880ce59e30b356b686bd3144efe24f875ebc8`. `stage1/typescript/parser/WHOLE_REPORT.md` pins that corpus. I inspected cohere's `TypeScript` submodule: it is the Go port, commit `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`, not the original JS compiler. I fetched the pinned original into scratch. Its checked-in `lib/_tsc.js` and `lib/typescript.js` match the npm 6.0.3 files byte for byte. Required diagnostics were generated with the upstream `scripts/processDiagnosticMessages.mjs`; Node 22 declarations and source-map-support declarations were installed in scratch. The first, incomplete self-check failed; it is retained as `compiler-baseline.log`, not used in the tables.

| Input | Project and scope |
| --- | --- |
| Compiler | Original `src/compiler/tsconfig.json`: 78 roots including generated diagnostics; 230 transitive files, 194,779 TypeScript lines, 56,034 declaration lines and 9,809 library lines. Its namespace imports also pull in other compiler-repository source. |
| Mid | Adamic `stage1/typescript/parser/main.ts` and scanner imports: 9 implementation files, 4,501 TypeScript lines; 79 files with libraries and prelude. This is the real stage1 whole-file parser, not a generated benchmark. |
| Tiny | Existing `internal/load/testdata/0.1/compile/01_hello.ts`, 2 lines; 71 files with libraries and prelude. Its costs are mostly compiler startup and standard-library checking. |

Mid and tiny scratch configs use the repository's compiler options and `internal/load/prelude.d.ts`, with only their entry point as a root. Libraries remain enabled and checked. Every measured check exits 0 with no diagnostics. All runs use `--noEmit --incremental false --composite false --pretty false --extendedDiagnostics`; the incremental/composite overrides avoid incremental work and `.tsbuildinfo` writes. Each input has separate GC, CPU, sampling, constructor-census and phase runs. Timings are single-run observations, not a stable performance ranking.

Node **24.19.0**, V8 **13.6.233.17-node.51**, Linux amd64, Intel Xeon Platinum 8573C; `nproc=5`, cgroup `cpu.max=400000 100000`, memory limit 16 GiB. Setup succeeded:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (184s)
setup: done in 184s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Sampling uses the inspector at 32 KiB, with both `includeObjectsCollectedByMajorGC` and `includeObjectsCollectedByMinorGC` true. Ordinary exit heap profiles omit collected allocations. Phase snapshots force GC and preserve V8 object IDs. The API harness suppresses exactly the service-allocator installation line in a scratch-loaded copy of `typescript.js`, keeping the CLI constructors, and never changes the checkout. Parsing means `createProgram`; binding includes checker initialization through `getTypeChecker`; checking means `getPreEmitDiagnostics`, including declaration diagnostics. A separate standalone scanner pass is additional work, outside the CLI totals.

## Observed allocation and memory

| Metric | Compiler | Mid | Tiny |
| --- | --- | --- | --- |
| GC-run wall seconds | 18.299 | 1.036 | 1.003 |
| CLI sampled allocation MiB | 3618.6 | 112.9 | 51.9 |
| Allocation through last GC MiB | 3654.3 | 97.4 | 54.8 |
| Scavenge heap reduction MiB | 3151.0 | 54.5 | 23.8 |
| Reduction / recorded allocation | 86.2% | 56.0% | 43.4% |
| Scavenges / major GCs | 69 / 2 | 11 / 0 | 8 / 0 |
| GC pause ms | 936.1 | 39.1 | 33.5 |
| Peak observed heap at GC MiB | 533.4 | 56.8 | 43.9 |
| CLI RSS high-water MiB | 698.5 | 140.0 | 108.2 |
| Checked snapshot self bytes MiB | 439.9 | 50.7 | 40.2 |
| Baseline snapshot self bytes MiB | 21.2 | 21.1 | 21.1 |
| After dropping Program MiB | 32.1 | 28.9 | 27.8 |

The reduction ratio is a **young-death proxy**, not an exact object-survival fraction: it sums `start_object_size - end_object_size` for scavenges, divided by allocation recorded through the last GC. It omits trailing allocation and mixes startup, promotion and object populations. The checked compiler snapshot holds 439.9 MiB versus 3,618.6 MiB sampled across a separate CLI run: their ratio is about 12.2%, but is not a measured survival percentage. GC-boundary peak is a lower bound on peak heap; RSS polls `/proc/PID/status` every 10 ms and takes VmHWM as well as VmRSS, so a final interval can be missed. GNU time was absent. Snapshot-run RSS reached 2,614 MiB and is excluded from the CLI RSS row.

Exact constructor calls after module initialization, including small config-parser setup allocations:

| Allocated objects | Compiler | Mid | Tiny |
| --- | --- | --- | --- |
| Node family | 1,148,711 | 61,976 | 31,407 |
| Symbol | 265,697 | 10,759 | 9,027 |
| Type | 111,016 | 11,188 | 3,865 |
| Signature | 42,784 | 537 | 293 |
| Map constructions | 271,018 | 9,818 | 8,316 |
| Set constructions | 49,800 | 1,673 | 1,612 |

Arrays, primitive strings and closures have no interceptable common constructor: literals, concatenation, internal buffers and inlined functions bypass constructors. Their **total allocated object counts and byte-size distribution are not obtainable from this sampler**. Sampling gives estimated bytes at allocation stacks, often at an inlined caller rather than a constructor. The following counts and self bytes are exact for the captured live graph, not all objects ever allocated:

| Checked live kind | Compiler count / MiB | Mid count / MiB | Tiny count / MiB |
| --- | --- | --- | --- |
| Node family | 1,095,525 / 151.1 | 59,981 / 8.2 | 29,720 / 4.0 |
| Symbol | 261,202 / 33.9 | 10,723 / 1.4 | 9,028 / 1.2 |
| Type | 108,251 / 6.6 | 11,172 / 0.7 | 3,851 / 0.2 |
| Signature | 41,706 / 4.1 | 537 / 0.1 | 293 / 0.0 |
| Array header | 501,889 / 15.3 | 27,591 / 0.8 | 20,763 / 0.6 |
| Backing arrays/tables | 619,126 / 90.8 | 33,975 / 5.2 | 27,332 / 4.2 |
| Map | 143,047 / 4.4 | 8,274 / 0.3 | 6,886 / 0.2 |
| Set | 276 / 0.0 | 124 / 0.0 | 116 / 0.0 |
| String | 464,307 / 44.2 | 67,454 / 14.1 | 58,971 / 13.7 |
| Closure | 26,584 / 1.5 | 24,061 / 1.4 | 23,994 / 1.4 |
| Other object | 854,490 / 37.1 | 40,253 / 1.8 | 23,242 / 1.1 |
| V8/other | 718,105 / 50.8 | 177,711 / 16.8 | 149,499 / 13.5 |

Self bytes exclude separately allocated backing stores. Thus a 32-byte Map header is not a 32-byte table, and array headers must not be added to a table size that already includes them. V8 internal metadata is not Adamic payload.

## Phase lifetimes

| Input / phase | Sampled allocation MiB | Post-GC live heap MiB | New named objects |
| --- | --- | --- | --- |
| compiler / parse | 346.9 | 226.6 | 446,980 Identifier, 557,047 Node, 101,594 Token, 230 SourceFile |
| compiler / bind | 114.4 | 279.1 | 121,198 Symbol, 87 Type, 4 Signature |
| compiler / check | 3235.8 | 439.9 | 110,929 Type, 42,780 Signature, 144,499 Symbol, 4,861 Token, 9,118 Identifier, 28,708 Node, 78 SourceFile |
| mid / parse | 27.6 | 36.4 | 9,654 Token, 18,837 Identifier, 33,280 Node, 79 SourceFile |
| mid / bind | 9.4 | 41.0 | 7,678 Symbol, 85 Type, 4 Signature |
| mid / check | 75.1 | 50.7 | 11,103 Type, 533 Signature, 3,081 Symbol, 13 Identifier, 33 Node, 22 Token |
| tiny / parse | 19.1 | 30.5 | 4,840 Token, 12,390 Identifier, 14,012 Node, 71 SourceFile |
| tiny / bind | 5.7 | 34.2 | 6,566 Symbol, 85 Type, 4 Signature |
| tiny / check | 22.5 | 40.2 | 3,780 Type, 289 Signature, 2,461 Symbol, 13 Identifier, 17 Node, 6 Token |

Compiler object-ID cohorts show all **1,082,744 parsed live nodes** still present after binding and after checking. All 95,294 bound live Maps persist through checking; checking adds 47,753 live Maps. When Program ownership is dropped, only 3 of 1,095,525 live Node-family objects and 147 of 143,047 Maps remain; all 108,251 Types and 41,706 Signatures disappear, along with 261,201 Symbols, leaving one baseline Symbol. Mid and tiny release all Types and Signatures, retaining 31 and 2 Symbols respectively; some nodes remain in compiler/module caches. Snapshot IDs establish these cohorts; constructor calls and live counts come from different runs and are not an object-by-object allocation-survival join.

Scanner-alone sampled allocation is compiler 19.2 MiB, mid 1.8 MiB, tiny 0.7 MiB. It scans retained source text, returns numeric tokens, and allocates token-value strings, notably templates and identifiers. Parser work includes scanning, AST objects, NodeArrays, JSDoc and identifier maps. Binder keeps symbols, locals, exports and flow metadata. Checker keeps types, signatures, node/symbol links and relation caches, while relation and flow calls allocate and discard scratch state. File texts, ASTs, intern maps and bound tables remain reachable across files; the observation does not support freeing those with a per-file arena.

The largest CLI allocation sites, sampled self bytes, including collected allocations:

| Input | Allocation site | MiB | Share |
| --- | --- | --- | --- |
| compiler | checkTypeRelatedTo | 978.0 | 27.0% |
| compiler | getFlowTypeOfReference | 763.9 | 21.1% |
| compiler | next | 79.2 | 2.2% |
| compiler | filterType | 66.5 | 1.8% |
| compiler | set | 55.2 | 1.5% |
| mid | getFlowTypeOfReference | 20.3 | 18.0% |
| mid | readFileSync | 6.0 | 5.3% |
| mid | checkTypeRelatedTo | 5.8 | 5.1% |
| mid | createScanner | 3.9 | 3.4% |
| mid | next | 3.7 | 3.3% |
| tiny | readFileSync | 6.0 | 11.5% |
| tiny | createScanner | 3.9 | 7.5% |
| tiny | (anonymous) | 2.6 | 5.0% |
| tiny | next | 2.3 | 4.3% |
| tiny | checkTypeRelatedTo | 1.8 | 3.4% |

In compiler source, `checker.ts:22380` (`checkTypeRelatedTo`) and `checker.ts:29004` (`getFlowTypeOfReference`) contain many nested functions and captured scratch locals. Together their sampled sites account for about 48% of CLI allocation. **Inference:** V8 closure/context creation is a major opportunity for stack allocation or non-escaping call regions after native closure conversion. The profiler does not identify every allocation there as a closure or prove those captures cannot escape.

## Shapes, dictionaries and identifiers

Most frequent constructed shapes in the compiler census, before factory fields are appended:

| Constructor | Calls | Initial own fields |
| --- | --- | --- |
| Node | 585,832 | 10 |
| Identifier | 456,098 | 9 |
| Symbol | 265,697 | 14 |
| Type | 111,016 | 1 |
| Token | 106,471 | 8 |
| Signature | 42,784 | 1 |
| SourceFile | 310 | 10 |

Completed AST shapes were separately walked with `Object.keys`, including numeric and undefined fields. This avoids snapshot property-edge counts, which omit immediate numeric values:

| Completed compiler AST shape | Live tree occurrences | Own fields |
| --- | --- | --- |
| Identifier | 420,457 | 13 |
| PropertyAccessExpression | 68,122 | 17 |
| CallExpression | 52,509 | 16 |
| TypeReference | 50,368 | 12 |
| Parameter | 38,643 | 19 |
| BinaryExpression | 38,454 | 16 |
| Block | 24,613 | 15 |
| ExpressionStatement | 17,392 | 13 |

Mid and tiny most frequent completed shapes:

| Input | Shape | Tree occurrences | Own fields |
| --- | --- | --- | --- |
| mid | Identifier | 14,499 | 13 |
| mid | FirstLiteralToken | 4,985 | 14 |
| mid | PropertyAccessExpression | 3,366 | 17 |
| tiny | Identifier | 8,222 | 13 |
| tiny | Parameter | 2,784 | 19 |
| tiny | TypeReference | 2,569 | 12 |

The completed tree walk counts 1,050,275 nodes; it excludes detached/synthetic objects and JSDoc subtrees omitted by `forEachChild` that are present in the heap snapshot. Initial constructor field counts are not the final shapes. Among snapshot shapes, 254,872 NodeLinks expose `resolvedSymbol` as their sole reference-valued field; there are 429,870 NodeLinks and 220,144 SymbolLinks in total. These are compiler metadata objects, not string-keyed symbol tables.

Checked-live self-size distributions (non-cumulative upper-bound buckets in bytes; JS objects, strings, closures and backing arrays together):

| Self size bucket | Compiler count / MiB | Mid count / MiB | Tiny count / MiB |
| --- | --- | --- | --- |
| 0 / <=16 | 1,556 / 0.0 | 1,556 / 0.0 | 1,556 / 0.0 |
| <=32 | 1,541,563 / 44.8 | 103,750 / 2.9 | 82,398 / 2.4 |
| <=48 | 534,363 / 24.2 | 35,503 / 1.4 | 22,748 / 0.9 |
| <=64 | 281,340 / 16.0 | 45,788 / 2.7 | 35,117 / 2.0 |
| <=96 | 112,574 / 9.4 | 15,873 / 1.3 | 11,005 / 0.9 |
| <=128 | 482,620 / 57.9 | 19,764 / 2.4 | 13,239 / 1.6 |
| <=192 | 1,132,674 / 169.2 | 59,273 / 9.0 | 35,732 / 5.3 |
| <=256 | 1,832 / 0.4 | 813 / 0.2 | 810 / 0.2 |
| <=384 | 12,083 / 3.7 | 735 / 0.2 | 613 / 0.2 |
| <=512 | 3,459 / 1.6 | 282 / 0.1 | 237 / 0.1 |
| <=1024 | 6,473 / 5.5 | 239 / 0.2 | 214 / 0.2 |
| <=4096 | 5,051 / 9.6 | 381 / 0.7 | 353 / 0.7 |
| <=65536 | 722 / 9.1 | 175 / 2.7 | 162 / 2.4 |
| >65536 | 93 / 37.7 | 13 / 10.1 | 12 / 9.9 |

The zero-byte entries are V8 shared/empty structures. This is a survivor distribution, not the size distribution of dead allocations and not a native struct layout.

`utilities.ts:644` implements `createSymbolTable` with `new Map`. Parser `internIdentifier` at `parser.ts:2637` uses a per-file Map. Checker type-interning, instantiation and relation caches are Maps (`checker.ts:2048` and `2385`). Numeric node/symbol-link caches use arrays. Index signatures remain in options and pragma argument records; JSON/package configuration also has string keys. Fixed-shape AST and link objects must remain distinct from dynamic records.

Maximum cardinality reached per registered Map, including Maps later collected (four compiler and three mid/tiny Maps were initialized before instrumentation and first registered on a later set):

| Max entries | Compiler | Mid | Tiny |
| --- | --- | --- | --- |
| 0 | 103,574 | 5,407 | 4,499 |
| 1 | 37,357 | 1,454 | 1,258 |
| 2-4 | 107,172 | 2,208 | 1,994 |
| 5-16 | 10,357 | 491 | 357 |
| 17-64 | 11,965 | 227 | 186 |
| 65-256 | 451 | 29 | 22 |
| 257-1024 | 111 | 3 | 3 |
| >1024 | 35 | 2 | 0 |

Maps at 0 to 4 entries constitute 91.5% of the compiler population; 35 exceed 1,024 entries. This is maximum size over lifetime, not final size or backing capacity. A Map copied from an iterable is counted through its setter as well as its constructor without duplicating registration.

| Live V8 dictionary-mode records | Compiler | Mid | Tiny |
| --- | --- | --- | --- |
| Before Program | 587 | 589 | 589 |
| After checking | 587 | 586 | 588 |
| After checking, null-prototype Object records | 67 | 67 | 67 |
| Largest observed reference-key count | 4496 | 4496 | 4496 |

V8 dictionary-mode records have an `(object properties)` backing store; null-prototype records have an explicit `__proto__` edge to null. They are not synonymous with TypeScript index signatures: small dynamic records can have fast shapes, and a fixed source record can be in dictionary mode. In all three checked snapshots the dictionary reference-key buckets are dominated by 5 keys (300 records), 2 (63), 1 (50), and 0 (25). The large 4,496-, 2,130-, 785- and 562-key tables are already visible before creating Program, mostly module/runtime/compiler tables. There is no growing dictionary-mode population comparable to the Map population. Exact semantic record cardinalities, including numeric-valued keys, were not recovered from these snapshots.

| Identifier interning | Compiler | Mid | Tiny |
| --- | --- | --- | --- |
| Parser identifier occurrences | 446,980 | 18,837 | 12,390 |
| Sum of per-file intern-map entries | 59,822 | 2,609 | 2,011 |
| Distinct identifier string values across files | 27,108 | 1,211 | 950 |
| UTF-8 bytes of distinct values | 580,587 | 12,100 | 9,770 |

Compiler distinct identifier lengths in UTF-8: 1-8 bytes: 3,551, 9-16 bytes: 9,047, 17-32 bytes: 10,876, 33-64 bytes: 2,729, >64 bytes: 905. These are value counts, not physical V8 string counts: interning is per file and equal strings across files need not share storage. The tables include declaration/library identifiers.

## Time and operations

CPU profiles use Node's default approximately 1 ms sampling. Percentages below are self time / total recorded sample time, including startup and GC, not inclusive caller time:

| Input | Function | Self ms | Samples share |
| --- | --- | --- | --- |
| compiler | (garbage collector) | 775.9 | 5.05% |
| compiler | (anonymous) | 460.2 | 3.00% |
| compiler | checkTypeRelatedTo | 334.0 | 2.18% |
| compiler | checkIdentifier | 307.8 | 2.00% |
| compiler | getFlowTypeOfReference | 267.2 | 1.74% |
| compiler | isTypeRelatedTo | 240.6 | 1.57% |
| mid | (anonymous) | 43.3 | 4.08% |
| mid | (garbage collector) | 38.9 | 3.67% |
| mid | wrapSafe | 23.0 | 2.17% |
| mid | (program) | 17.9 | 1.68% |
| mid | scan | 16.9 | 1.60% |
| mid | bindWorker | 13.9 | 1.31% |
| tiny | (program) | 37.8 | 4.80% |
| tiny | wrapSafe | 35.5 | 4.51% |
| tiny | (anonymous) | 31.5 | 4.00% |
| tiny | (garbage collector) | 30.3 | 3.86% |
| tiny | compileFunctionForCJSLoader | 16.9 | 2.14% |
| tiny | readFileUtf8 | 15.3 | 1.94% |

For the compiler, `getPropertyOfType` adds 175.2 ms (1.14%) of self samples. It is a checker algorithm and cannot be read as the cost of machine property loads. **No reliable percentage for record loads, Map builtins, string building or array push can be extracted from these CPU profiles**: V8 inlines operations into callers. The 3.85 ms attributed to functions named `get`/`set` is JS cache-wrapper code in `_tsc.js`, not Map builtin time; no distinct push/string builtin samples were recorded. Zero distinct samples does not mean zero cost.

| Separate census operations | Compiler | Mid | Tiny |
| --- | --- | --- | --- |
| get | 4,429,860 | 137,132 | 58,589 |
| set | 1,201,980 | 29,750 | 20,566 |
| has | 318,504 | 1,319 | 1,292 |
| Array push calls | 3,796,673 | 127,337 | 79,619 |
| Pushed elements | 3,796,989 | 127,352 | 79,619 |
| String/object/function values pushed | 3,386,171 | 98,749 | 70,842 |
| String/object/function Map key/value stores | 2,144,088 | 56,018 | 39,839 |

Map operations are prototype calls; array literals, direct index stores, other array helpers, Set stores and record field accesses are not counted. The census includes a handful of harness operations and constant strings. It is not used for time attribution. Its named-constructor totals agree with the first census.

## Runtime implications: inference, not observation

**Allocation classes and chunks.** Start native experiments with 16-byte alignment and total allocation-size classes 32, 48, 64, 96, 128, 160, 192, 256, 384 and 512 bytes, including Adamic headers when choosing a class. Add powers-of-two buffer capacities and a separate large-allocation path. The 128 to 192-byte V8 survivor population and 13 to 22-field completed ASTs argue for dense small-object slabs, but V8 sizes are not C sizes. Try 64 KiB slabs for program-owned AST/symbol/type objects, and 16 to 64 KiB chunks for proven call scratch; compare 16, 64 and 256 KiB alternatives for fragmentation and wasted tails. Keep those as tuning candidates, not settled chunk sizes.

**Ownership and regions.** Program-lifetime ownership fits ASTs, intern maps and bound/checker caches; a per-file region cannot be freed while the checker still reaches those files. Put only proven non-escaping token scratch and relation/flow-call work in shorter regions. Parent/backlinks and symbol/type graphs need explicit weak edges or validated common region ownership; ordinary reference counting cannot collect their cycles. Closure conversion may eliminate much of the two dominant V8 allocation sites before an allocator is involved. Regions and reuse still require escape/liveness proof; object survival alone is not that proof.

**Tables.** Prioritize ordered Map with a cheap empty representation, inline or compact capacity for 1 to 4 entries, growth through 8/16/64, and conventional hashing for the long tail. Preserve insertion order and live iteration. Do not add a dedicated large record table for symbol tables: this target uses Map. Generic dynamic records still need their JS key semantics for configuration/JSON; the observed V8 dictionary population is mostly baseline and does not size that feature. Start sparse records small, and measure their semantic key-count distribution natively before choosing a separate hash-table design. Identifier pools should be owned by Program or by files kept through Program lifetime; measure global pooling against the observed 59,822 per-file entries and 27,108 distinct values, including lookup cost and string bytes.

**Reference counting estimate.** The compiler census makes 1,568,208 named compiler objects; divided by the separate 18.299-second CLI GC run, that is approximately 85,700 objects/s. This is a normalization, not a measured allocation rate of the instrumented pass. Container instrumentation observes 3,386,171 reference-eligible pushes and 2,144,088 Map key/value stores: 5,530,259 possible retain sites, about 302,221/s at that same duration. If every one retained and eventually released once, 11.06 million count operations would cost about 55/111/221 ms at hypothetical 5/10/20 ns per operation. This is one illustrative container-store budget, not a bound on total RC cost: constants and moves lower it; field/index assignments, local ownership, replacement/destruction, Set stores and captured cells add work. Node rates reveal neither retains per allocation nor cache misses on counters. Measure native counts by ownership operation and phase, divide by allocation counts, record null/immortal skips, moves/borrows and region elisions, then multiply dynamic non-elided counts by separately measured single-threaded retain/release costs under the actual working set. Only that supports a runtime RC estimate.

**First native measurements.** Run these identical no-emit projects and compare diagnostics/exit status with pinned tsc first. Then collect allocation count and bytes by native type/size class, class occupancy and slab tail waste; high-water live objects/bytes and RSS; program/file/call ownership and escape; strong/weak backedges and teardown; Map/record cardinality and capacity histograms; string pool hits/bytes; closure/cell allocation; retained/released references by operation and phase. Time uninstrumented native parse/bind/check separately, including table lookup and string/array work, then inspect generated C and native CPU samples. Test no-count regions and heap mode against identical output and sanitizer/leak checks before interpreting their speed.

## Reproduction, validation and limits

The report attachment is `tsc-allocation-profile.tar.gz`, kept under `/workspace/scratch/tsc-profile/`, outside git. It contains raw GC/CPU/heap profiles, five phase snapshots per input, configs, source/file manifests, scripts, summary JSON and logs. The final worker report links the archive. Only these summary tables and prose are committed. `raw/commands.json` records all 15 primary commands, durations and exit codes; `run.py`, `sample.cjs`, `phases.cjs`, `operations.cjs`, `details.cjs`, `analyze.py` and record-analysis helpers reproduce the measurements. The original source and dependency pins are in the archive README and package lock; no raw profiles are in the repository.

Representative commands (all output goes to a log file):

```sh
source /workspace/adamic-tools/env.sh
node --trace-gc-nvp SOURCE/lib/tsc.js -p SOURCE/src/compiler/tsconfig.json --noEmit --incremental false --composite false --pretty false --extendedDiagnostics > compiler-gc.log 2>&1
node --cpu-prof --cpu-prof-dir=RAW --cpu-prof-name=compiler.cpuprofile SOURCE/lib/tsc.js -p SOURCE/src/compiler/tsconfig.json --noEmit --incremental false --composite false --pretty false --extendedDiagnostics > compiler-cpu.log 2>&1
PROFILE_OUT=RAW/compiler-all.heapprofile node --require SCRATCH/sample.cjs SOURCE/lib/tsc.js -p SOURCE/src/compiler/tsconfig.json --noEmit --incremental false --composite false --pretty false --extendedDiagnostics > compiler-sample.log 2>&1
node --expose-gc SCRATCH/phases.cjs compiler SOURCE/src/compiler/tsconfig.json phases > compiler-phases.log 2>&1
python3 SCRATCH/validate.py > validation.log 2>&1
node --expose-gc SCRATCH/probes.cjs > probes.log 2>&1
go test -count=1 -v -timeout 30m ./internal/oracle -run '^TestTheOracleCatchesOneByte$' > filtered-oracle.log 2>&1
go vet ./... > vet.log 2>&1
gofmt -l cmd internal > gofmt.log
```

| Mutant or control | What caught it |
| --- | --- |
| CPU samples charged twice | Independent raw timeDeltas budget, on all three profiles. |
| Scavenge reduction sign reversed | Raw before/after heap comparison, on all three GC logs. |
| Maps classified as Sets | Independent constructor-name scan of the real tiny snapshot, 6,886 Maps. |
| Collected allocations omitted | A real allocating Node probe: all-allocation sample 14,086,448 bytes versus zero survivor bytes after GC; collected=false fails the allocation assertion. |
| Null-prototype detection based only on absent prototype edge | Real record control shows an explicit edge to null; the old detector returned zero candidates, and the corrected detector sees the control. |
| Tiny source given number = string in host overlay | Pinned TypeScript diagnostics catch TS2322. No source file was modified. |
| One-byte native output mutation | Existing TestTheOracleCatchesOneByte passes, demonstrating rejection by the external Node comparison. |

All 15 primary runs, all three added operation censuses and all three completed-shape walks exited 0. Validation and probes pass. The filtered oracle passes in 20.987 s; go vet exits 0 and gofmt reports no files. Full repository tests were not run: this unit changes documentation only and no package code. Not covered: native allocation layouts or timing; exact total array/string/closure allocation counts; semantic dynamic-record cardinalities; per-file or per-call death events below sampling resolution; precise heap maximum or young/end survival fraction; instruction-level cost of inlined property/Map/string/push operations; native RC count traffic, escape proof or cycle handling; repeated-run uncertainty and workloads beyond the three selected projects.

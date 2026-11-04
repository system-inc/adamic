# The C that tsc's clean census units emit, and what clang costs on it (#0gvtg3y)

A measuring pass for the compiler Performance seat. No tsc file lowers whole yet, but the latent census attempts every top-level declaration (and nested function declaration) as its own unit, and some units lower with no finding. Those units have no C, because the census omits module order, ownership and the backend. This branch holds a scratch instrument that finishes and emits C for the clean units, and the clang timings on that C.

Everything here is scratch. No file under `internal/` or `cmd/` was edited or committed; the instrument lives in a Go overlay (the `.txt` files under `overlay/`). Nothing on `main` moved.

Words used for evidence: **measured** is a number a program printed on this box. **Inferred** is my reading of measured things and is marked as such.

## The short answer

- 10,709 units in tsc's 79 `src/compiler` files at 9295787a: 9,886 attempted, 823 deferred generics. **4,992 attempted units are clean** (no NotYet, no Refused, no panic, no SkippedDependency), in 78 of the 79 files. (measured)
- **4,886 of those 4,992 emitted C** through the post-lowering steps and `native.CWithOptions` (58.7 MB of C, 338,174 lines in all). 106 could not, and each row says why (below). (measured)
- **3,908 of the 4,886 compile with clang at -O0 and -O2.** 978 do not: 954 are nested function declarations that read a captured variable of their enclosing function, which lives in the parent's frame and is declared nowhere in an isolated unit. They are not numbers, they are rows with a first error (`data/clang_failures.csv`). (measured)
- **Clang's cost on these units is small and flat**: median 0.03 s at -O2 (the floor is the `adamic.h` parse, 0.02 s and about 50 MB), p90 0.03 s, **worst unit 1.84 s** (a 6.0 MB initializer of `diagnosticInformationMap.generated.ts`). Summed over the 3,908 compiled units: 89 s at -O0 and 105 s at -O2 for 40.3 MB of C, 2.61 s/MB at -O2. (measured)
- **Nothing here says what whole-file C will cost.** These units are one declaration each, so no function in them exceeds 824 lines (p99 209, median 14), far below the 12,500 lines where the earlier table found clang going superlinear. A whole file's `main` and module initialization will look nothing like this. (my reading)
- **Step 2 gives 4, not 7.** The units clean at 9295787a and not at e4c3a163 are four functions in `moduleNameResolver.ts` (listed below). Correctness counts 7; I could not reproduce 7 and did not force it. (measured, with the other side below)

## Box and toolchain

| | |
|---|---|
| CPU | Intel Xeon Processor @ 2.10GHz (the VM reports no finer model), 4 cores |
| Memory | 15.7 GiB (16,091 MB), no swap; the container's cgroup killed two processes near 7.7 GB and 13.7 GB resident (see "What went wrong") |
| Kernel | Linux 6.18.44 |
| Load average during the clang timings | 1-minute average at each clang start: min 0.01, median 1.0, max 1.15 (the measuring clang is the 1.0); nothing else ran |
| clang | 20.1.8, the official LLVM release `LLVM-20.1.8-Linux-X64.tar.xz` (sha256 `1ead36b3...29ddc1ae6`), commit `87f0227cb601`. Only `bin/clang*`, `bin/ld.lld`, `lib/clang` extracted (736 MB) |
| Go | 1.27.1 |
| Node | v22.23.4, **not 24**. Only `stage3/apply.sh`'s adapter uses Node; the instrument and clang timings do not |
| GNU time | 1.9 (`apt-get install time`, not preinstalled), `/usr/bin/time -v` as in `docs/tsc-generated-c/measure.py` |
| Revisions | measured at `9295787a403578ffd8762fb0e29cee61035285f6` (branch `typescript/adapt32-directory-exists`); plain census also at `e4c3a1632aa33563574178f359dec4e44c78374f` (main's parent of that commit) |
| TypeScript | v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`, adapted by `stage3/apply.sh` to a fresh path per revision. The two adapted trees differ in two files only, `utilities.ts` (the `directoryProbablyExists` signature) and `moduleNameResolver.ts` (the casts at `directoryProbablyExists` call sites removed), by same-line edits |

clang is invoked exactly as `docs/tsc-generated-c/measure.py` does: `measure_units.py` imports that file's own `FLAGS`, `clang_once`, `functions` and `time_trace_top`, so the flags are `native.Flags` for a release build without `-flto=thin` (the earlier table also left ThinLTO out), `-O0` or `-O2`, `-I internal/native/runtime`, `clang -c` of one file, `/usr/bin/time -v`. Seconds are wall clock with 0.01 s resolution; memory is `Maximum resident set size`. Units under 300 KB of C ran three times and the fastest is reported; the two larger ones ran once. One clang at a time, nothing else running.

## Steps 1 and 2: the clean units at each revision

`clean_units.py` on the census output (measured). Both censuses ran in full mode over `src/compiler`; the same 10,709 units exist at both revisions (same files, same line:column).

| | 9295787a | e4c3a163 |
|---|---:|---:|
| units | 10,709 | 10,709 |
| attempted | 9,886 | 9,886 |
| deferred generic | 823 | 823 |
| **clean** | **4,992** | **4,988** |

Clean at 9295787a and not at e4c3a163 (adaptation 32's newly whole units), by file and declaration:

1. `moduleNameResolver.ts:1986:1` `nodeLoadModuleByRelativeName`
2. `moduleNameResolver.ts:2131:1` `tryAddingExtensions`
3. `moduleNameResolver.ts:3064:1` `loadModuleFromImmediateNodeModulesDirectory`
4. `moduleNameResolver.ts:3169:1` `tryLoadModuleUsingPaths`

Clean at e4c3a163 and not at 9295787a: none. All four were blocked at e4c3a163 by `Refused: a cast the runtime can't check` (the casts adaptation 32 removed) and nothing else.

**Why not 7.** Twelve units in `moduleNameResolver.ts` lose a cast refusal between the revisions (`data/changed_units_9295787a_vs_e4c3a163.txt` lists all twelve with every finding). Only four have nothing else blocking them. The other eight (`tryLoadModuleUsingRootDirs`, `tryLoadModuleUsingBaseUrl`, `getPackageJsonInfo`, `loadNodeModuleFromDirectoryWorker`, `loadModuleFromSpecificNodeModulesDirectory`, `resolveTypeReferenceDirective`, `primaryLookup`, `resolveFromTypeRoot`) are still not clean at 9295787a. I do not know how Correctness reaches 7: it may count by a different rule (for example units that lose a refusal, or by a different clean definition). Both sides are in the data; I did not force it. Source: `adapted-w9295` vs `adapted-we4c3` diffs are in `logs/apply-*.log`.

**Cross-check that the instrument's census is the census.** The 9295787a census of record is the instrument binary's (the plain binary was OOM-killed at 7.7 GB resident on the full tree, twice, see below). I re-ran the **plain** latent binary on shard 1 (15 files): units and findings are identical to the instrument's for all 15 files (`out-plain-w9295/census-1.jsonl` vs `out-unitc-w9295/census-1.jsonl`, compared in this session; the comparison script is inline in the session, the two inputs are in `data/`... see "Not covered"). The e4c3a163 census is the plain binary, all six shards.

Shard caveat from the latent README applies: the census ran in six file shards (`LATENT_ONLY_FILE`), each a separate process over the whole program. That changes only `instantiated` for deferred generics, which clean-unit status does not read.

## The instrument

`overlay/make_unit_c_overlay.py.txt` copies the latent overlay builder (`stage3/census/latent/make_overlay.py`, only its `units.go.txt` and `main.go.txt` templates swapped) and layers three things on it. Full diff against production and the latent templates: `overlay/diff-against-production-and-latent.txt`.

1. **`unit_c.go.txt` (package lower).** After a unit lowers with no finding, `latentEmitUnit` runs what `lowerChecked` runs after lowering, in its order: prepend `forwarderValues` to `Main`, `finishClassCalls`, `finishAccessors`, `finishToPrimitive`, `exceptions`, `checkLibrarySpreads`, `checkAccessorSpreads`, `checkRecordGuardedReads`, `checkDynamicKeyedGetters`, `flow.NormalizeAsync`, `checkViewOptionalWrites`, `checkViewCopies`, `checkViewWidenedReads`, the async payload failures, `checkPromiseConsumption`, `checkViewLogicalWrites`, `readiness`, `borrow`, `counters`, `voidDiscards`. For a statement unit the statements `l.statement(node)` returns become `Main`; for a function unit `Main` is empty. The unit's lowering state is the census's snapshot copy, so the passes cannot touch later units. (The census only discarded those statements before.)
2. **A hook in `native/emit.go`** (overlay copy): two exported sets, `LatentExternFunctions` and `LatentExternGlobals`. A function in the first gets no body and an `extern` prototype; a global in the second gets `extern` declarations. The rule: a function registered before the unit (index below the base count) with no body that is not the unit's own function is extern; a global local registered before the unit is extern. Everything the unit itself creates (nested closures, its own locals) is emitted.
3. **`main.go.txt`**: a sink that calls `native.CWithOptions(program, Options{Release: true})`, writes `u<N>.c` per unit, and a row in `units.jsonl` (status, phase and first error when a step fails). Then `filter_externs.py` drops `extern` lines whose name occurs nowhere outside extern lines (a real split would declare only what a unit reaches; without it each unit carried about 930 prototypes, 83 KB).

**What the instrument leaves out, on purpose or by necessity:**

- **`findCycles` is omitted.** It walks the whole source file's AST, not the unit, and on `binder.ts:1311` (`isNarrowingTypeofOperands`) `checkLazyViewReads` recursed without bound through `allocationFlowGraph.reachingAllocations` (goroutine dump in `logs/unitc-run-9295-aborted-findcycles.log`; 94 units in 20 minutes before I stopped it). Its work is graph-region and cycle ownership marking and `resolveCountTypes`. Programs with cyclic object types would be marked differently in a real build. This is a finding for the Optimizer seat in its own right: that check is not bounded on a file the size of binder.ts.
- Module order and the pre-lowering passes (`noteInheritance`, enum, namespace initialization) are the census's own registration, not a real module-ordered run.
- **Whole-program facts the real build has and a unit does not.** See the proof below.

**One gate fix I found late.** The sink fires for any unit with no finding added *during its own lowering*, but the up-front syntax scan (`refusal_scan`) records findings against units before any unit runs. 526 units reached the sink with such a scan finding. They are not clean by the brief's definition and are **excluded**: all numbers here are the 4,992 census-clean units only (`data/instrument-rows-all-sunk-units.jsonl` keeps all 5,518 rows).

## Step 4: does the instrument emit what the real path emits

Four oracle fixtures that lower whole through `adamic c` today (`internal/oracle/testdata/{array_from,adversarial_order,borrow_loop_calls,user_iterators}.a`, 41 to 483 lines), run through the instrument unit by unit, every function compared with the same-named function in `adamic c`'s `native.C` output (`compare_functions.py`, `classify_differences.py`; stable names are source-hashed so they match across the two paths). Statement units are compared against the real module initializer (`compare_init.py`: the unit's initializer body, temporaries renumbered by first appearance, must appear as a contiguous run of the real initializer). Raw results: `proof/`.

| fixture | units | functions byte-identical | differ, field access only | differ, other | init bodies found in real init |
|---|---:|---:|---:|---:|---:|
| array_from | 18 | 11 | 1 | 0 | 17 of 17 |
| adversarial_order | 28 | 17 | 0 | 0 | 23 of 24 |
| borrow_loop_calls | 79 | 40 | 9 | 24 | 29 of 34 |
| user_iterators | 86 | 475 | 157 | 7 | 22 of 68 |

(Function counts include repeats: a function appears in each unit's C that references it. Module initializers are counted in the last column, not the first four.)

So this is **not** byte-identical everywhere and I do not claim it is. Where it differs, the cause is always a whole-program fact the real build computes and a unit cannot, found by reading the diffs (inferred from the diffs, each diff is in `proof/*.differences.txt`):

1. **Field layout (the commonest).** `native/fields.go` `uniformFieldOffsets` decides a field read is `ADAMIC_UNIFORM_OBJECT(o, "f")->slots[k]` only if every layout in the whole program puts `f` at slot `k`. A unit sees only its own object literals, so reads whose literal lives in another unit come out as `adamic_object_data_field(o, "f", &cache_slot)`, and in the other direction a unit that sees one layout uses the uniform form where the real build, seeing a conflict, keeps the shape check.
2. **Borrowing and throw checks.** The siblings are bodiless placeholders, so `borrow` sees a callee that mutates nothing (the unit drops a retain the real output keeps) and `exceptions` sees a callee that cannot throw (the unit omits an `adamic_thrown` check after a closure call the real output has).
3. **Packed argument slots** (`FixedArgumentSlots`, a program-wide count): a closure call passes one value in the unit where the real call passes two.

**Does it matter for clang's cost?** Measured: summed over the differing functions, unit C is 1.054x, 0.958x and 0.958x the real bytes (array_from, borrow_loop_calls, user_iterators). Inferred: the two directions are small and of opposite sign, so I expect unit timings to be within a few percent of the real function's; I did not compile the real functions to check. The unit C likely understates a real build's retains and throw checks slightly. A unit's C is also **not** the C of the function inside a whole file in one respect that could matter more: its module initializer holds one statement, where the real one holds all the file's statements in 131-line chunks.

**Proof that the checks can fail (mutants).**
- Mutant 1: in a copy of `adamic c`'s `array_from` output, the `size()` return changed 2 to 4. `classify_differences.py` went from identical 11, other 0 to identical 10, other 1: caught.
- Mutant 2: in unit 2's C, a `adamic_retain(` removed. `compare_init.py` went from 17 of 17 to 16 of 17 with unit 2 named: caught.
Both mutants change real input and neither is catchable by clang's `-Werror`. They prove the comparers can fail; I did not mutate the instrument itself (for example by skipping `borrow`), which would be the stronger proof. (Not covered.)

## Step 5 and 6: what clang costs

`data/units.csv` has one row per emitted unit: C bytes, lines, functions, the largest function and its lines, wall seconds and peak RSS at -O0 and -O2. `data/per_file.csv` is the roll-up. 4,886 units; 3,908 compile.

Per tsc file, the twelve most expensive at -O2 (sums over the file's units that compiled; `units` is all units in the file, `clean` is units with emitted C, `failed` is units that emitted C clang rejected):

| file | units | attempted | clean (emitted) | clang-failed | C KB | C lines | -O0 s | -O2 s | peak RSS -O2 MB |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| types.ts | 875 | 875 | 873 | 0 | 425 | 13,898 | 18.98 | 20.65 | 55 |
| checker.ts | 2,494 | 2,464 | 1,009 | 353 | 16,537 | 102,289 | 15.45 | 18.39 | 60 |
| utilities.ts | 933 | 867 | 488 | 10 | 5,596 | 31,396 | 11.16 | 13.17 | 63 |
| parser.ts | 602 | 545 | 217 | 10 | 2,490 | 14,320 | 4.57 | 5.31 | 56 |
| utilitiesPublic.ts | 252 | 243 | 133 | 0 | 1,254 | 6,517 | 3.30 | 3.96 | 59 |
| commandLineParser.ts | 195 | 173 | 95 | 7 | 1,639 | 7,903 | 2.16 | 2.94 | 68 |
| moduleNameResolver.ts | 207 | 168 | 98 | 6 | 1,337 | 8,311 | 2.14 | 2.53 | 56 |
| emitter.ts | 442 | 423 | 234 | 136 | 4,158 | 19,383 | 2.21 | 2.51 | 57 |
| scanner.ts | 164 | 150 | 133 | 62 | 2,158 | 13,833 | 1.71 | 2.15 | 64 |
| _namespaces/ts.ts | 78 | 78 | 78 | 0 | 18 | 1,170 | 1.72 | 1.95 | 52 |
| program.ts | 248 | 235 | 108 | 37 | 1,391 | 7,179 | 1.65 | 1.92 | 59 |
| diagnosticInformationMap.generated.ts | 3 | 3 | 3 | 0 | 6,060 | 17,741 | 0.48 | 1.89 | 110 |

All 79 files: `data/per_file.csv` (columns also give deferred-generic and census-clean counts; `data/clean_vs_emitted_per_file.csv` gives census-clean against emitted per file). Totals: 4,886 emitted units, 58.7 MB, 338,174 lines, 88.97 s of -O0 and 105.12 s of -O2 over the units that compiled.

**Worst unit by -O2 seconds:** `diagnosticInformationMap.generated.ts:12:1`, a variable statement (the `diag(...)` object literal), 6,043,752 bytes of C, largest function 131 lines (the emitter's chunking), **1.84 s, 110.5 MB** at -O2 (0.43 s at -O0). (measured)

**Worst unit by seconds per MB**, over units of 100 KB or more: `scanner.ts:4096:1`, 167,902 bytes, 0.20 s at -O2, **1.19 s/MB**. Below 100 KB a seconds-per-MB figure measures the 0.02 s floor, not the code, so I did not rank those. (measured)

**-ftime-trace on the slowest -O2 unit** (that same 6.0 MB unit; 2.24 s with the tracer on): Backend 1.77 s of 2.22 s, CodeGenPasses 1.09 s, Optimizer 0.67 s. The per-function split names the unit's own module initializer `adamic_initialize_imported...diagnosticInformationMap_generated_ts` at 0.206 s, and the 131-line chunk functions at 0.02 s each. The largest pass is `X86 DAG->DAG Instruction Selection` 0.378 s, then `Greedy Register Allocator` 0.135 s and `Machine Instruction Scheduler` 0.126 s. So in this unit the time is many small functions, not one big one. Raw: `data/slowest_trace.json`. (measured)

**Peak RSS** is 47 MB (-O0 floor) to 110 MB; the median at -O2 is 52 MB. (measured)

### The units that did not give a number

Not emitted (106 of 4,992 clean units, `data/instrument-rows-all-sunk-units.jsonl`, `status` not `emitted`, first error in `error`):

| phase | units | first error (typical) |
|---|---:|---|
| `native.CWithOptions` panic | 67 | `native: no conversion from 4 to 2` (22), `compiler bug: incomplete checked field contract` (16), other `no conversion from N to M` (29) |
| `checkViewWidenedReads` | 30 | stage 0 refuses a checked view read (named site) |
| `checkAccessorSpreads` | 6 | `can't lower spreading in a program with static constructor objects yet` |
| `checkRecordGuardedReads` | 2 | |
| `checkLibrarySpreads` | 1 | |

Emitted but rejected by clang at -O0 (978 of 4,886, `data/clang_failures.csv`, first error per unit):

| first error | units |
|---|---:|
| `use of undeclared identifier '...local..._cell'` (a capture cell of the enclosing function) | 954 |
| 1 more `undeclared identifier` (`..._ready`) | 1 |
| `static declaration of '...' follows non-static declaration` | 16 |
| `incompatible pointer types returning ...` | 4 |
| `used type 'adamic_maybe_boolean' where arithmetic or pointer type is required` | 2 |
| `returning 'adamic_object *' from a function with incompatible result type 'double'` | 1 |

Failures are concentrated in nested declarations: 968 of the 1,927 clean units with depth above 0 fail, against 10 of 2,959 top-level units. The 954 capture-cell failures are the isolation limit named in the brief: a nested function's captured variable has no home outside its parent. The 16 `static follows non-static` errors name a nested function (for example `symbolToString_symbolToStringWorker`): the C has a non-static (extern) prototype and a static one for the same name. I read that as two IR functions sharing one C name, a placeholder I marked extern and the unit's own definition, so an instrument artifact; I did not confirm it. The other 12 are three kinds of emitted C that does not type check; I have not diagnosed whether they are instrument artifacts or real emitter findings, and they are listed, not hidden.

The 106 and 978 are not "most": the instrument works on 4,886 of 4,992 (98%) and 3,908 of 4,886 (80%) compile, so step 7's stop condition does not apply.

## What this does and does not tell you

- **Measured:** per-unit C size and clang cost for tsc's clean units at 9295787a, with the flags and method of the earlier table.
- **Not measured, inferred, or out of reach:**
  - Whole-file C. The units are declaration-sized; the earlier table's finding (cost turns superlinear at about 12,500 lines in one function) cannot appear here, and the 131-line chunking of module initializers means no unit's initializer is large. What a split policy will see on a real `checker.ts` is not in this data.
  - The 5,000 or so units that do not lower (NotYet, Refused) are the majority of tsc's code by bytes. Clean units are a biased sample: they skew small, and the nested-function units that fail clang are exactly the closure-heavy code `checker.ts` is built from.
  - Per-unit C carries a fixed preamble (strings, class shapes) that a split build would share; I did not separate it from the body, so the s/MB figures include it.
  - The instrument's whole-program differences (above) move bytes a few percent in both directions; I measured the byte ratio, not clang on the real functions.
  - `findCycles` and its marking are absent from every unit.

## What went wrong along the way (kept so the numbers can be trusted)

- The harness stopped my first background chain at its 30 minute limit and the plain census with it (no output; census records are written at the end).
- The plain 9295787a census on the whole tree was **OOM-killed** (7.7 GB resident). The instrument's first full run was killed too (13.7 GB resident). I switched to six file shards (`LATENT_ONLY_FILE`, `GOGC=30`, `GOMEMLIMIT=6GiB`, two at a time; `overlay/run_shard.sh.txt`, `overlay/shard*-files.txt`, `logs/shards.log`). The plain 9295787a run was kept to one shard as a cross-check (shard 0 was killed by me, `exit=143`, because it duplicates the instrument's census; `logs/shards.log`).
- The first instrument run (with `findCycles`) was aborted, see above. Its rows are `logs/aborted-findcycles-run-rows.jsonl`.
- I measured clang twice. The first pass keyed files by basename, so `utilities.ts` collided across directories in the roll-up. I fixed the key to the path under `src/compiler` and re-ran all 4,886 units on the quiet box. All numbers here are from the second run. The first run's totals were 95.0 s at -O0 and 112.0 s at -O2 (second run 89.0 and 105.1), within 7%.
- The Go-side extern filter kept duplicate prototypes; `filter_externs.py` replaced it as a post-pass over the stored C (83 KB reduced to a median 7.9 KB per unit). Any identifier it wrongly dropped would have shown as `undeclared identifier`, and only the capture-cell and one `_ready` case did.

## Not covered

- The C itself (58.7 MB) is not committed; `rebuild.sh` and `run_shard.sh` regenerate it. Three samples are under `samples/`.
- The plain-vs-instrument census comparison on shard 1 was run as an inline script in the session; its inputs (`out-unitc-w9295/census-1.jsonl`, `out-plain-w9295/census-1.jsonl`) are not separately committed (the merged 9295787a and e4c3a163 censuses are in `data/` as `.gz`).
- The instrument was not mutation tested; the two mutants above test the comparers.
- Clean definition: attempted, no NotYet, no Refused, no panic, no SkippedDependency finding attributed to the unit (`unit == where`). Of the 4,992 clean units, none has any finding of any kind in the census (checked: zero findings, `Boundary` and `error` rows included, are attributed to them).

## Layout

- `data/`: `units.csv`, `rows.json` (raw clang samples, load averages), `per_file.csv`, `summary.json`, `clang_failures.csv`, `clean_vs_emitted_per_file.csv`, `slowest_trace.json`, `clean-9295.csv`, `clean-e4c3.csv`, `changed_units_9295787a_vs_e4c3a163.txt`, `instrument-rows-all-sunk-units.jsonl`, both full censuses (`.jsonl.gz`).
- `overlay/`: the instrument and the scripts, as `.txt`; `diff-against-production-and-latent.txt`; the generated `internal_native_emit.go.txt` and `internal_lower_latent_unit_c.go.txt`.
- `proof/`: step 4 outputs per fixture.
- `logs/`: every build and run log, including the aborted and killed runs.
- `samples/`: one emitted unit.

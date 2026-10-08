# Stage 1 parse instruction bisect — October 8, 2026

Two runtime merges add instructions in the requested interval:

- First material increase: **7c4dae857cc13b334e699880a5af23b2b7a83c15**, “Merge runtime/area-take-async2 at 2eea315c into area/runtime”: **+221,189,341 Ir (+3.33%)** over c4eb37d9.
- Largest increase: **d70bc6b1b90777ae5d265f1b101e497f823d61f8**, “Merge runtime/area-take-regions2 at c06aad7f into area/runtime”: **+511,101,239 Ir (+7.45%)** over 7c4dae85.
- The inner first-parent bisect identifies **ef56778a98cb9ca8c5b77c08eb3b5bf43eb1ccde**, “Graph regions meet concurrency: one header bit each, graph members and interior cells counted on the slow path.” Against its first parent 915b9e05 it adds **732,289,553 Ir (+11.02%)**. This integration also includes the nested-cell ownership support that the area imports separately through the preceding async merge.

The generated parse C is byte-identical across those measured boundaries:
`47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e`.
The driver contains no `adamic_graph_` calls. These are runtime support costs on the ordinary parse workload, including changed ThinLTO inlining.

No fix was made. This report is the only committed change.

## The stated 5.77G baseline does not reproduce at 915b9e05

The first measurement was 915b9e05, as requested: **6,643,815,239 Ir**, not 5.77G. Then the exact emitted C from the stage1-profiles report was reconstructed at its recorded implementation revision 305f7457; it matches the report’s C hash `2eb710273065b5949e662181b4d88e2e42af0ab28d49138db3bc505876f74836`.

Rebuilding that exact C using **915b9e05’s runtime and shipped builder** gives **6,634,630,297 Ir**, again about 6.63G. Its own original report-revision build gives **5,767,060,566 Ir**, reproducing the historical 5.77G to its quoted precision (the original report records 5,766,783,329). All three builds match Go’s full AST bytes. The small 277,237-instruction difference from the original historical measurement was not separately attributed; it is 0.0048%.

Thus the report’s historical number cannot be assigned to 915b9e05 on this box. The reconstructed historical branch and 915b9e05 share release-lto 586aaf96 as their merge base, but 915b9e05 has additional runtime/compiler changes. This investigation does not bisect the changes before the user’s specified 915b9e05 endpoint.

The latest endpoint **56d53e10** reproduces **7,293,119,515 Ir**. Its measured increase over the actual 915b9e05 driver is **649,304,276 Ir (+9.77%)**.

Exact accounting along the requested first-parent history:

```
           89  915b9e05 → c4eb37d9 (small measured early difference; unattributed)
  221,189,341  c4eb37d9 → 7c4dae85 (async import)
  511,101,239  7c4dae85 → d70bc6b1 (graph import)
  -82,986,393  d70bc6b1 → 56d53e10 (subsequent changes, net)
--------------
  649,304,276  total endpoint increase
```

## Measurement and setup

No AGENTS.md exists in the workspace/repository or the dependency worktrees inspected. Fetched by name:

```
git fetch origin area/runtime runtime/area-take-leftovers codex/release-lto
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
```

Setup passed. Node is **v24.19.0**, Go **1.27.1**, clang/LLVM/lld **20.1.8**, Valgrind/Cachegrind **3.27.1**. The setup initially fetched substantial dependency history; redundant follow-up history fetches were stopped after the required objects became available. No inference about parse speed is based on build wall time.

Each measurement uses the revision’s load/lower/native compiler and its own recorded cohere/TypeScript-Go pins. Revisions predating the named parse entrypoint receive the identical `.a` parse harness from 56d53e10 in the isolated measurement worktree, then that supplied directory is removed before checkout. The root is `stage1/cohere/parse/parse.a`. A temporary Go adapter calls `load.Load`, `lower.Lower`, `native.C`, and `native.Build(..., native.Options{Release:true})`, saving the exact generated C. The adapter disables Go VCS stamping because of its temporary symlinked submodule layout; this does not change the native parse build policy.

The shipped uncounted, unsanitized native flags at the requested endpoints and measured merge boundaries are:

```
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign
-ffp-contract=off -fno-optimize-sibling-calls -pthread -O2 -flto=thin
```

The complete list reaches every runtime compilation and the program link. The Linux link adds `-fuse-ld=lld`; there is no CPU override and no profile. The historical report-revision build predates the `-pthread` addition and uses the report’s recorded flags. `--count` is the driver output option, not an `ADAMIC_COUNT` build.

The benchmark is batch 8’s **77 compiler files** from TypeScript commit `050880ce59e30b356b686bd3144efe24f875ebc8`, in the order of [benchmarks.json](../../../stage1/profiles/benchmarks.json). Only those source files were fetched for the corpus after stopping an unnecessary full-history fetch. Every SHA256 in the manifest was checked before every measurement.

An independent Go adapter is built by overlaying `stage1/typescript/parser/testdata/oracle.go` into the pinned Go parser module and running `--manifest compiler.txt --whole`. Every exact counted native binary also runs `--manifest compiler.txt --ast` and must match Go byte-for-byte, with empty stderr:

- **44,766,682 bytes**;
- SHA256 **8ae015600498b915cc25abab82730299451ae990b50478980d5a3bc465801bfe**.

Counting command (paths abbreviated):

```
VALGRIND_LIB=tools/usr/libexec/valgrind tools/usr/bin/valgrind   --tool=cachegrind --cache-sim=yes --branch-sim=yes   --I1=32768,8,64 --D1=32768,8,64 --LL=268435456,1,64   --log-file=cachegrind.log --cachegrind-out-file=cachegrind.out   parse --manifest compiler.txt --count
```

Each measured output is exactly `0\n` with empty program stderr. All 13 raw Cachegrind event totals reconcile exactly to the sum of their self-cost records. These are simulated instruction counts; no timing or hardware-cycle claim is made.

## Bisects and every measured revision

The larger-step bisect used:

```
git bisect start --first-parent 56d53e10 915b9e05
```

Classification was Ir above **6,968,467,377**, the reproduced endpoint midpoint. Selected revisions: 90f476ad (bad), 2683b038 (bad), 7c4dae85 (below that threshold), 58caaefb (bad), d70bc6b1 (bad). Result: **d70bc6b1**. “Good” in this search means below the larger-step threshold; it does not mean 7c4dae85 adds no instructions.

A second first-parent bisect used the already measured/validated binaries and a threshold of baseline plus 100M (**6,743,815,239**) to identify the earlier material increase. Its selections were 90f476ad (bad), 2683b038 (bad), 7c4dae85 (bad), c4eb37d9 (good). Result: **7c4dae85**. The 89-instruction early difference is explicitly retained in the accounting above.

For the graph import, follow its second parent c06aad7f through f09f8dbe’s graph parent 18a0be06. That branch has a five-commit first-parent path from 915b9e05: ef56778a, 138ce59d, 30878275, cbb75642, 18a0be06. The inner bisect used:

```
git bisect start --first-parent 18a0be06 915b9e05
```

18a0be06 was first measured and confirmed bad. Selected/measured revisions were 138ce59d (bad) and ef56778a (bad). Result: **ef56778a**, whose measured first parent is 915b9e05. Its 732M increase includes both graph support and the nested environment/cell-owner support; d70bc6b1’s 511M marginal increase is against a parent that already has the latter.

| Revision | Instructions | Role |
|---|---:|---|
| `915b9e05` | 6,643,815,239 | requested older endpoint |
| `c4eb37d9` | 6,643,815,328 | exact parent of async import |
| `7c4dae85` | 6,865,004,669 | async import / graph parent |
| `d70bc6b1` | 7,376,105,908 | graph import |
| `58caaefb` | 7,376,105,908 | outer bisect step |
| `2683b038` | 7,376,105,908 | outer bisect step |
| `90f476ad` | 7,376,678,997 | outer bisect step |
| `56d53e10` | 7,293,119,515 | requested latest endpoint |
| `18a0be06` | 7,376,104,792 | inner branch endpoint |
| `138ce59d` | 7,376,104,792 | inner bisect step |
| `ef56778a` | 7,376,104,792 | inner first bad commit |

## Where the instructions go

The full `callgrind_annotate` output was produced for both sides of each boundary:

```
tools/usr/bin/callgrind_annotate --show=Ir --threshold=100 --auto=no cachegrind.out
```

Tables below use **self Ir**, summed across raw function records. LLVM’s `.llvm.<digits>` private suffix is removed before aligning functions. For the endpoint comparison, generated `adamic_function_<id>_<name>` numeric IDs are also normalized because emission can renumber them. Anonymous/unknown buckets remain included in the full reconciled delta.

ThinLTO moves inlined work into/out of callers. A newly visible helper’s gross self delta is not all additional work. In particular, `allocate_storage` takes the old `adamic_allocate` cost: at the graph merge, both are exactly 208,776,056 Ir, so that rename contributes **zero**. Likewise some cleanup work moves from `release_last` to the extracted helpers. The complete sum, including all decreases, is the stated net delta.

### Graph merge: 7c4dae85 → d70bc6b1

| Function | Before Ir | After Ir | Delta Ir |
|---|---:|---:|---:|
| `adamic_retain_slow` | 139,178 | 375,336,396 | +375,197,218 |
| `adamic_release_slow` | 23,686,262 | 279,569,802 | +255,883,540 |
| `adamic_heap_free_children` | 0 | 241,180,567 | +241,180,567 |
| `allocate_storage` | 0 | 208,776,056 | +208,776,056 |
| `deallocate` | 0 | 115,261,760 | +115,261,760 |
| `adamic_retain` | 52,437,327 | 90,211,253 | +37,773,926 |
| `adamic_array_set` | 38,175,529 | 39,063,332 | +887,803 |
| `adamic_function_Parser_unary` | 27,589,042 | 27,631,136 | +42,094 |
| `release_last` | 349,364,607 | 109,864,976 | -239,499,631 |
| `adamic_allocate` | 208,776,056 | 0 | -208,776,056 |
| `adamic_function_Parser_kind` | 473,635,100 | 416,798,888 | -56,836,212 |
| `adamic_function_Scanner_scan` | 586,824,829 | 547,871,258 | -38,953,571 |

All function deltas sum to **511,101,239**.

`internal/native/runtime/adamic.h` changes inline retain/release dispatch from the shared bit alone to `ADAMIC_SLOW_COUNT` (shared or graph). `heap.c` adds `counted_heap`, graph retain/release dispatch, graph-aware storage teardown, and extracts child destruction into `adamic_heap_free_children`. The parse C does not change and contains no graph API calls, so the parser still pays these general runtime checks.

The shipped `adamic_retain_slow` assembly has a new slab load and `test $0x20000000` before the ordinary count update, plus the graph branch and additional prologue/epilogue work. Its self cost grows from 139,178 to 375,336,396 Ir, consistent with work that was mostly inlined becoming visible in the larger slow helper. Cachegrind cannot by itself separate all inlining effects from individual added checks, so no sum of gross helper deltas is presented as the net regression.

### Inner commit: 915b9e05 → ef56778a

| Function | Before Ir | After Ir | Delta Ir |
|---|---:|---:|---:|
| `adamic_retain_slow` | 0 | 375,336,396 | +375,336,396 |
| `adamic_release_slow` | 16,379,665 | 279,569,802 | +263,190,137 |
| `adamic_heap_free_children` | 0 | 241,179,451 | +241,179,451 |
| `allocate_storage` | 0 | 208,776,056 | +208,776,056 |
| `deallocate` | 0 | 115,261,760 | +115,261,760 |
| `let_go` | 0 | 98,257,462 | +98,257,462 |
| `release_last` | 396,166,227 | 109,864,976 | -286,301,251 |
| `adamic_allocate` | 208,775,526 | 0 | -208,775,526 |
| `adamic_function_Parser_kind` | 445,216,994 | 416,798,888 | -28,418,106 |

All function deltas sum to **732,289,553**.

### Earlier async merge: c4eb37d9 → 7c4dae85

| Function | Before Ir | After Ir | Delta Ir |
|---|---:|---:|---:|
| `let_go` | 0 | 99,749,059 | +99,749,059 |
| `adamic_function_Parser_kind` | 445,216,994 | 473,635,100 | +28,418,106 |
| `adamic_function_ParseNode_new` | 269,411,382 | 291,640,033 | +22,228,651 |
| `adamic_release` | 91,143,274 | 109,249,600 | +18,106,326 |
| `adamic_function_Parser_suffix` | 101,339,248 | 115,366,278 | +14,027,030 |
| `adamic_function_Scanner_scan` | 575,945,992 | 586,824,829 | +10,878,837 |
| `adamic_release_slow` | 16,379,665 | 23,686,262 | +7,306,597 |
| `release_last` | 396,166,227 | 349,364,607 | -46,801,620 |

All function deltas sum to **221,189,341**. The same emitted-C hash on both sides again isolates runtime changes. Source blame traces the cell-owner tests in `adamic_retain_slow` and `drop_reference` to **b15216da**, “Lower named nested functions through one counted frame environment”; **64a3eed1** adapts the owner to the ordinary async frame. They are imported via the async branch’s runtime integration a64243ca and then 2eea315c/7c4dae85. This is source attribution for the smaller merge, not an isolated measurement of b15216da: that historical feature revision predates the release ThinLTO API used here.

### Requested endpoints: 915b9e05 → 56d53e10

| Function | Before Ir | After Ir | Delta Ir |
|---|---:|---:|---:|
| `adamic_retain_slow` | 0 | 375,350,958 | +375,350,958 |
| `adamic_release_slow` | 16,379,665 | 279,582,714 | +263,203,049 |
| `adamic_heap_free_children` | 0 | 241,180,584 | +241,180,584 |
| `allocate_storage` | 0 | 208,776,056 | +208,776,056 |
| `deallocate` | 0 | 115,261,760 | +115,261,760 |
| `let_go` | 0 | 98,257,502 | +98,257,502 |
| `adamic_retain` | 48,620,124 | 90,209,103 | +41,588,979 |
| `release_last` | 396,166,227 | 109,864,976 | -286,301,251 |
| `adamic_allocate` | 208,775,526 | 0 | -208,775,526 |
| `adamic_object_callee` | 76,275,342 | 16,086,580 | -60,188,762 |
| `adamic_string_units` | 61,566,985 | 29,264,790 | -32,302,195 |

All function deltas sum to **649,304,276**. The endpoint difference is smaller than the immediate combined increases because later changes remove a net 82,986,393 Ir. Those later reductions were not separately bisected.

## Exact artifacts and retained evidence

All runs, AST outputs, emitted C, native binaries, raw profiles, complete annotations, function-delta JSON, runner scripts and three bisect logs are retained locally in `/workspace/scratch/parse-bisect/`. Only these notes are published; no source, profile, binary or raw-corpus upload is included.

| Revision | Binary SHA256 | Emitted C SHA256 |
|---|---|---|
| `915b9e05` | `161ee6c799900e62f128d0770b2c5a579371aba9f41a6061c1f6bddf3891372a` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `c4eb37d9` | `fcf935b8c2b12781ee56c17b1f47c89bd5953cbe54d2744cf873dc5e5d050000` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `7c4dae85` | `1fb24ee5129001bc2cba7f73edbd26370ab2084bf15dd45feb3870fee95e32b9` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `d70bc6b1` | `de2a3f4459f462c080c19c37eff327d7feddf11e423c28f927ade075894250fb` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `58caaefb` | `de2a3f4459f462c080c19c37eff327d7feddf11e423c28f927ade075894250fb` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `2683b038` | `de2a3f4459f462c080c19c37eff327d7feddf11e423c28f927ade075894250fb` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `90f476ad` | `fae3f03710d425766d5e1d432abd337126b5878bd20f25e3234c0a7246ab9d82` | `866e672cf5073a12b6474a0389182eaf65713b7baca2f742e7d57622e37871ac` |
| `56d53e10` | `8a7d99f233647c7dbaa98823ccb9fd459fd20f4947244202f2c90693755c8039` | `866e672cf5073a12b6474a0389182eaf65713b7baca2f742e7d57622e37871ac` |
| `18a0be06` | `37fb5d8700b087fb24b5cf183107eb147dcf538037f4b977d2fe1d122ed41bf0` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `138ce59d` | `37fb5d8700b087fb24b5cf183107eb147dcf538037f4b977d2fe1d122ed41bf0` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |
| `ef56778a` | `2ca92ce8dedfd78a9251220c1d16c2c072308fbc1e8d156f2a388f9ee4eea2c8` | `47c3bada4ecd02cf89b68f0c9fe524031ff801ab6bb0d7da2b32419365e8f31e` |

The 915b9e05 frozen-report-C build has binary SHA256 `26237c8d2024269c2922b126a3b04a58d489bb2913c2d41beed772083d31a686` and the original `2eb71027…` C identity. Requested early/merge runs pin cohere `715ba94f3608a6500086b1076ce5cb7e51b836db` and TypeScript-Go `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`; 90f476ad and 56d53e10 pin cohere `7945d102a6c18dd36adf9114a758ce646e8b2359` and TypeScript-Go `d92d9bfee114c80be2c375d72edae966176e3a4f`.

The user's original checkout remains clean on `work`. No main branch, fix, force operation, PR, training profile or production build policy was changed.

# Regional allocation: completed wall-time win

Base: fetched `origin/area/runtime`, `c1c6073021e8e5cd6005fcdee59eb3b377395fd8`.
Branch: `runtime/region-alloc-speed`. No AGENTS.md was present.

The confirmation batch measured **3.41% lower best wall time**, 1.541211 s to
1.488658 s. The exploratory batch also showed a win (7.28%). The middle batch had
extreme variability (Node 2.136–8.070 s) and its apparent 44.53% improvement is not
claimed as a credible effect size. All three batches and every sample are retained.
This is a modest shared-machine best-of-five win, not a claim to match Bun.
The user stopped further work after this completed result; no further experiments
were performed after that instruction.

## Change

`adamic_region` now caches its current cursor and end. Allocation checks remaining
capacity and bumps the cursor. Block growth, malloc, capacity selection and sealing
of the old block's used extent are in a noinline helper. The builder saves two
fewer registers per recursive call. Completed block extents are retained for
outside-child teardown and Weak membership; the current extent is derived from
the live cursor and sealed at teardown. Both cached pointers reset at region end.

Uncounted builds no longer increment the diagnostic object total per object.
Counted builds retain that work and their existing counts. The duplicate slab-zero
assignment was removed; it was already folded by clang, so no timing gain is
attributed to it. Reference/kind/slab, shape, class and frozen initialization remain:
object operations need those fields. Alignment remains 16 bytes, with the tree's
56-byte object occupying 64 bytes; alignment arithmetic was already constant-folded
in its release build. Zeroed constructors and fully filled literals retain their
separate initialization behavior. Block sizes/growth policy and the walker did
not change.

## Perf profile and field reads

Perf 6.12.107 recorded the actual CLI release binaries with `-e cycles:u -F 997`.
The environment records **task-clock:uH**, so these are software user-time samples,
not precise hardware-cycle or cache-miss attribution. Both reports lost zero samples.

| Self samples | Baseline | Candidate |
|---|---:|---:|
| Regional recursive build | 46.85% | 37.41% |
| Recursive walk | 38.78% | 44.96% |
| Heap child destruction | 4.92% | 6.38% |
| malloc, all paths | 1.11% | 1.53% |
| New out-of-line growth helper | — | 0.27% |

Percentages from separate sampled runs are not absolute timings. Baseline perf
annotate is retained for build and walk; candidate walk annotation is also retained.
Candidate build annotation crashes in the utility, including with source lookup
disabled. Raw perf data and the successful sampled report are retained; this is
an annotation-tool limitation, not a target-program crash.

The walk reads left/right directly at offsets 0x28/0x30; it makes no runtime field
lookup or shape-cache calls. Baseline walk annotation placed 64.83% of local samples
on the instruction immediately after the left-pointer load. These samples can
skid and do not demonstrate that the following constant load itself is expensive.
The right pointer is checked and reloaded after recursive descent. No field-read
or purity assumption was changed.

## Machine and measurement

Machine `cb1632b4fc8c`, **INTEL(R) XEON(R) PLATINUM 8573C**, `Linux-6.18.44-x86_64-with-glibc2.41`;
five CPUs in affinity 0–4, quota `cpu.max=400000 100000`. Go 1.27.1,
clang 20.1.8, Node **v24.19.0**, Bun **1.3.14**, Python 3.12.14.
Required setup completed: `export GOPROXY='https://proxy.golang.org|direct';
bash cloud/setup.sh --wasi-sdk`, then sourced `/workspace/adamic-tools/env.sh`.

The unchanged `bench/trees.ts` SHA256 is `743b83a8bce42948f155d02dd8e939c94a2d7f403274b2b1981da381aa6c7f50`.
Release builds used `go run ./cmd/adamic build bench/trees.ts -o BINARY`:
shipped `-O2`, ThinLTO, lld, no counters or sanitizers. One untimed execution per
runtime followed by five sequential interleaved rounds of baseline, candidate,
Node and Bun, rotating the first runtime. Fresh processes; startup and output
included. Python perf_counter around launch/wait; child getrusage deltas for user
and system CPU. Best is minimum wall, with CPU from that same sample. Every stdout
matched byte for byte. NODE_OPTIONS, BUN_OPTIONS and ADAMIC_THREADS were cleared.
No builds, tests, setup or profiling overlapped timings. No CPU pinning or isolation.
Commands, executable hashes and all samples are in the JSON artifacts.

Confirmation load, 1/5/15 minutes: before
**0.04/0.42/0.70**;
after **0.55/0.50/0.72**.

| Runtime | Best wall seconds | User | System |
|---|---:|---:|---:|
| before | 1.541211 | 0.946923 | 0.593542 |
| candidate | 1.488658 | 0.986121 | 0.501846 |
| Node | 2.073956 | 2.723473 | 0.227501 |
| Bun | 1.310835 | 1.723812 | 0.377652 |

Candidate/Bun is 1.136×. Best-wall CPU samples
vary substantially; a uniform user-CPU reduction is not claimed.

| Runtime | Round 1 wall | 2 | 3 | 4 | 5 |
|---|---:|---:|---:|---:|---:|
| before | 1.775273 | 1.599929 | 1.859723 | 1.541211 | 1.640305 |
| candidate | 1.563357 | 1.578735 | 1.496046 | 1.488658 | 1.824765 |
| Node | 2.221189 | 2.073956 | 2.184802 | 2.232123 | 2.110447 |
| Bun | 1.337765 | 1.379975 | 1.325404 | 1.310835 | 1.566204 |

## Fixtures, mutants and checks

- `internal/native/testdata/region_alloc.c`: mixed 64/80-byte strides, cursor
  membership before teardown, block rollover, distinct values, zeroed slots,
  required headers, outside-child cleanup, Weak targets, region reuse and an object
  larger than the 1 MiB block cap. Release, counted and ASan/UBSan/leak controls pass.
- Three mutants compile and fail with the fixture's intended diagnostics:
  missing cursor advancement, lost completed-block extent and omitted alignment.
- `internal/oracle/testdata/region_alloc.a`: three fresh recursive trees crossing
  several blocks, each with dynamic string children outside the region. Native,
  Node, JavaScript backend, release, ASan malloc, ASan slabs and WASI agree.
- Existing regions, region_end and regions_throw oracle fixtures pass. Existing
  Weak and throwing-initialization region tests and WASI host ownership tests pass.
- Full tree benchmark passed ASan/UBSan with leak detection, empty stderr and
  byte-identical checksum output.
- Required counts regeneration passed and adds only the new fixture row:
  allocations 18477, frees 12320, retains 0, releases 12320, peak 8192, regions 6157.
  All existing rows remain identical. Counts are committed separately, only that file.
- No whole-package tests or full gate ran.

```sh
go test ./internal/native -run '^(TestRegionAllocationAndMutants|TestRegionEndWeakTargets|TestRegionEndThrowInitialization)$' -count=1 -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^(region_alloc|region_end|regions|regions_throw)[.]a$' -count=1 -v
ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASIHostPromises$' -count=1 -v
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/^region_alloc[.]a$' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
go run ./cmd/adamic build bench/trees.ts -o /tmp/region-alloc-profile/sanitized --sanitize
ASAN_OPTIONS=detect_leaks=1:halt_on_error=1 UBSAN_OPTIONS=halt_on_error=1 /tmp/region-alloc-profile/sanitized
```

Timing protocol: `python3 measure.py OUTPUT before=BASELINE candidate=CANDIDATE
Node=NODE Bun=BUN`. The retained runner records this workspace's source path.
Perf: `perf record -e cycles:u -F 997 -o DATA -- BINARY`,
`perf report --stdio --no-children -i DATA`,
`perf annotate --stdio --symbol adamic_function_1_check -i DATA`.
Validation logs and raw profile data accompany this report.

```text
stretch tree of depth 19	 check: 1048575
262144	 trees of depth 4	 check: 8126464
65536	 trees of depth 6	 check: 8323072
16384	 trees of depth 8	 check: 8372224
4096	 trees of depth 10	 check: 8384512
1024	 trees of depth 12	 check: 8387584
256	 trees of depth 14	 check: 8388352
64	 trees of depth 16	 check: 8388544
16	 trees of depth 18	 check: 8388592
long lived tree of depth 18	 check: 524287
```

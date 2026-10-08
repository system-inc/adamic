# Native, Node and Bun benchmarks

Run with `source /workspace/adamic-tools/env.sh; bash bench/run.sh`.

Six workload categories; string building and splitting use two existing programs. JSON encode is attempted separately. No compiler/runtime optimizations were made.

## Machine and versions

```json
{
  "UTC": "2026-10-08T05:31:31.812679+00:00",
  "machine": "a105a9032575",
  "OS/kernel/architecture": "Linux-6.18.44-x86_64-with-glibc2.41",
  "CPU": "AMD EPYC 9V74 80-Core Processor",
  "logical CPUs": 5,
  "available CPUs (affinity)": 5,
  "cgroup cpu.max (quota period, microseconds)": "400000 100000",
  "Adamic source commit": "ed174ae9a78c1fe0fd4d24663ca6f7eb4544a1ba",
  "working tree": "clean",
  "Go": "go version go1.27.1 linux/amd64",
  "clang": "clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)",
  "Node": "v24.19.0",
  "Bun": "1.3.14",
  "Python": "3.12.14",
  "Node executable": "/workspace/adamic-tools/bin/node",
  "Bun executable": "/home/agent/.cache/adamic-bench/bun-1.3.14-linux-x64/bun",
  "runner SHA256": "34649156083ea09c514af5b13286e3029886ad64a7c59a5543da99888e84e9d1",
  "load before timings (1/5/15 minutes)": "1.41 / 3.76 / 2.25",
  "load after timings (1/5/15 minutes)": "1.64 / 3.45 / 2.22"
}
```

## Method

Each source is built with `go run ./cmd/adamic build SOURCE -o BINARY`: shipped release build, clang -O2, no sanitizers or counters. Build/setup time is excluded. Node runs existing .ts sources directly and .a sources through oracle/node.mjs; Bun runs the same sources directly. No generated JavaScript is used.

One untimed execution per runtime validates exact, nonempty stdout before timing. All 15 timed executions per program must also match byte for byte or the script exits nonzero without writing a report. Each execution is a fresh process, with no in-process warmup. Startup, Node's type stripping, checksum calculation and output are included.

Five rounds interleave runtimes for each program. The first runtime rotates each round: native/Node/Bun, Node/Bun/native, Bun/native/Node, native/Node/Bun, Node/Bun/native. Wall time uses Python perf_counter around process launch and wait; user CPU time uses the difference in getrusage(RUSAGE_CHILDREN). Children run sequentially, so it measures that execution's user CPU, excluding kernel CPU and the Python parent. Best means minimum wall time; reported user time belongs to that same sample, not an independently selected minimum.

This is a shared cloud machine, without CPU pinning or isolation. Load averages and the full sample spread are recorded; small differences are not evidence of a repeatable win. Earlier setup and validation remain in the load averages; they finished before timing. No setup, tests or builds were started alongside timing by this runner. NODE_OPTIONS, BUN_OPTIONS and ADAMIC_THREADS are cleared for benchmark children.

## Workloads

- primes: ten fresh Uint8Array sieves with limits 2,000,000 through 2,000,009; count and sum every prime.
- number_sum_sort: one million random integers and one million nearly sorted integers; full sums before/after, comparator sorts, order check and an order-sensitive checksum. Adapted from the existing sort.ts generator/cases.
- string_build: existing 6,000 bounded formatter documents; construction and full UTF-16 checksum.
- word_count: existing text construction and splitting of a million words, map counts and frequency sort.
- map_200k: insert 200,000 distinct string keys, look every key up in reverse order, fold insertion order.
- json_encode: 100,000 fresh encodings of a nested object with runtime scalar values, followed by a full UTF-16 checksum scan. This measures encoding plus checksum work, not isolated encoder throughput.
- trees: existing recursive binary-tree build/walk, about 68 million allocated nodes, maximum depth 18.
- nbody: existing five-body floating-point simulation, one million steps, energy checksums.

## Results

Times in seconds. Ratios are native wall / competitor wall; above 1 means native loses.

| Program | Native wall | Native user | Node wall | Node user | Bun wall | Bun user | Native / Node | Native / Bun | Output |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| primes | 0.204145 | 0.199729 | 0.133286 | 0.135904 | 0.116334 | 0.129170 | 1.53x | 1.75x | matches |
| number_sum_sort | 0.220922 | 0.216549 | 0.478736 | 0.433015 | 0.539283 | 0.556417 | 0.46x | 0.41x | matches |
| string_build | 0.356816 | 0.356492 | 0.406322 | 0.396931 | 0.430293 | 0.414260 | 0.88x | 0.83x | matches |
| word_count | 0.142631 | 0.109615 | 0.258573 | 0.263786 | 0.319144 | 0.303650 | 0.55x | 0.45x | matches |
| map_200k | 0.061439 | 0.053068 | 0.159630 | 0.141717 | 0.156338 | 0.138907 | 0.38x | 0.39x | matches |
| json_encode | 0.414502 | 0.410144 | 0.385185 | 0.379120 | 0.402830 | 0.391294 | 1.08x | 1.03x | matches |
| trees | 0.963462 | 0.661716 | 1.396746 | 1.868817 | 0.774549 | 1.027110 | 0.69x | 1.24x | matches |
| nbody | 0.170090 | 0.169696 | 0.669191 | 0.660969 | 0.129202 | 0.137724 | 0.25x | 1.32x | matches |

Native losses are measured best-wall comparisons, including startup:

- Native loses to Node: primes (1.53x), json_encode (1.08x).
- Native wins against Node: number_sum_sort (0.46x), string_build (0.88x), word_count (0.55x), map_200k (0.38x), trees (0.69x), nbody (0.25x).
- Native loses to Bun: primes (1.75x), json_encode (1.03x), trees (1.24x), nbody (1.32x).
- Native wins against Bun: number_sum_sort (0.41x), string_build (0.83x), word_count (0.45x), map_200k (0.39x).

## Not measured

JSON decode: intentionally omitted; native refuses `JSON.parse`: its result's type can't be proven from the text (internal/lower/library_json_stringify.go).

## Sources and checksums

Source SHA256 identifies the exact workload; multiline tree/n-body output is retained.

### bench/primes.a

SHA256: `ab84dee93ab656378f51a8a5b6fe7707bc66fd7dfe4895bba67ff884fd83d518`

```text
1429153778578
```

### bench/number_sum_sort.a

SHA256: `7cb0d79fa436daa77333f230d3d506ca62c053b82e1089feb1c9e97f07457d30`

```text
995241087439 995241087439 true 432656200
```

### bench/string_build.a

SHA256: `331edac43345ab5e994add8e805f2bf79526cba5928b810a6146905f00a0506e`

```text
13354836 units 1926777760
```

### bench/word_count.ts

SHA256: `73f92c6aa5ea8a021be8516ea7ee24d2f5d949b22f1eb37055f707ea6fd70265`

```text
21 distinct words
vi 204856
ka 201599
su 198355
to 197411
ze 191090
```

### bench/map_200k.a

SHA256: `e87b52cdd8e3a2898d4e7b5f25a112fb221d204e6402f747248c9ab04bff66e7`

```text
200000 59999900000 1096446421
```

### bench/json_encode.a

SHA256: `b2a5c42a17d85cf52531e790da9344faaad82672b6e161791a69093fa3bb1dc7`

```text
13872295 units 1333439894
```

### bench/trees.ts

SHA256: `743b83a8bce42948f155d02dd8e939c94a2d7f403274b2b1981da381aa6c7f50`

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

### bench/nbody.ts

SHA256: `d10410ce62f12bfe0c1de964d49a34222de8e1040986f347561f43779792a09c`

```text
-0.169075164
-0.169086185
```

## Every timing sample

Samples are in round order. Each cell is wall / user seconds.

| Program | Runtime | Round 1 | Round 2 | Round 3 | Round 4 | Round 5 |
|---|---|---:|---:|---:|---:|---:|
| primes | native | 0.204145 / 0.199729 | 0.213865 / 0.213487 | 0.237867 / 0.237314 | 0.213112 / 0.212779 | 0.213153 / 0.208872 |
| primes | Node | 0.138404 / 0.140422 | 0.133286 / 0.135904 | 0.141609 / 0.122532 | 0.133466 / 0.126865 | 0.137878 / 0.141001 |
| primes | Bun | 0.132434 / 0.142030 | 0.116334 / 0.129170 | 0.124886 / 0.143907 | 0.120793 / 0.139113 | 0.120393 / 0.135743 |
| number_sum_sort | native | 0.248766 / 0.226786 | 0.220922 / 0.216549 | 0.256551 / 0.252714 | 0.221492 / 0.217308 | 0.223177 / 0.218616 |
| number_sum_sort | Node | 0.570763 / 0.533419 | 0.478736 / 0.433015 | 0.479679 / 0.469603 | 0.487908 / 0.448877 | 0.514079 / 0.484632 |
| number_sum_sort | Bun | 0.613368 / 0.630876 | 0.549504 / 0.558576 | 0.551288 / 0.562002 | 0.539283 / 0.556417 | 0.550696 / 0.567839 |
| string_build | native | 0.356816 / 0.356492 | 0.380495 / 0.375150 | 0.368090 / 0.367846 | 0.366339 / 0.365951 | 0.362352 / 0.362036 |
| string_build | Node | 0.432125 / 0.436444 | 0.453897 / 0.438725 | 0.406322 / 0.396931 | 0.447496 / 0.443713 | 0.445712 / 0.450074 |
| string_build | Bun | 0.458436 / 0.457045 | 0.430293 / 0.414260 | 0.477915 / 0.479333 | 0.444025 / 0.435141 | 0.463862 / 0.454431 |
| word_count | native | 0.159875 / 0.122222 | 0.145001 / 0.136681 | 0.155844 / 0.126228 | 0.142631 / 0.109615 | 0.190985 / 0.155439 |
| word_count | Node | 0.279778 / 0.239106 | 0.258573 / 0.263786 | 0.262856 / 0.224059 | 0.267132 / 0.238931 | 0.285259 / 0.233197 |
| word_count | Bun | 0.348557 / 0.329385 | 0.328086 / 0.310935 | 0.319144 / 0.303650 | 0.346943 / 0.341390 | 0.345275 / 0.337315 |
| map_200k | native | 0.061439 / 0.053068 | 0.069979 / 0.056490 | 0.063880 / 0.059634 | 0.074039 / 0.065533 | 0.074024 / 0.069602 |
| map_200k | Node | 0.159630 / 0.141717 | 0.216269 / 0.210882 | 0.197685 / 0.207171 | 0.249188 / 0.245139 | 0.180905 / 0.190076 |
| map_200k | Bun | 0.195876 / 0.168185 | 0.160535 / 0.140328 | 0.156338 / 0.138907 | 0.175066 / 0.154060 | 0.204504 / 0.196205 |
| json_encode | native | 0.453771 / 0.453368 | 0.414502 / 0.410144 | 0.449646 / 0.449336 | 0.436030 / 0.435636 | 0.472263 / 0.471886 |
| json_encode | Node | 0.385185 / 0.379120 | 0.389759 / 0.388408 | 0.410548 / 0.417562 | 0.386186 / 0.388719 | 0.398436 / 0.405869 |
| json_encode | Bun | 0.403752 / 0.381525 | 0.402830 / 0.391294 | 0.430580 / 0.420898 | 0.470503 / 0.463264 | 0.409911 / 0.405395 |
| trees | native | 0.999277 / 0.737879 | 0.963462 / 0.661716 | 1.020150 / 0.707189 | 1.036308 / 0.750386 | 1.051877 / 0.783290 |
| trees | Node | 1.405375 / 1.845104 | 1.408456 / 1.843198 | 1.396746 / 1.868817 | 1.407406 / 1.898079 | 1.449957 / 1.947075 |
| trees | Bun | 0.815931 / 1.067764 | 0.784866 / 1.143123 | 0.856173 / 1.193636 | 0.941543 / 1.218877 | 0.774549 / 1.027110 |
| nbody | native | 0.182782 / 0.178338 | 0.173556 / 0.173188 | 0.170090 / 0.169696 | 0.172902 / 0.172515 | 0.215729 / 0.215481 |
| nbody | Node | 0.669191 / 0.660969 | 0.678351 / 0.675178 | 0.699103 / 0.687234 | 0.696169 / 0.678056 | 0.722898 / 0.718573 |
| nbody | Bun | 0.129202 / 0.137724 | 0.135699 / 0.133365 | 0.129825 / 0.130769 | 0.132109 / 0.137606 | 0.147144 / 0.146647 |

## Validation

- `bash cloud/setup.sh` completed; sourced `/workspace/adamic-tools/env.sh`. Node v24.19.0; the runner installed Bun 1.3.14 from its GitHub release. Setup and validation finished before timing.
- `bash -n bench/run.sh` and Python compilation of its embedded program passed.
- `bash bench/run.sh --check-only` passed: all eight release builds succeeded, including nested JSON encoding, and all three runtimes agreed on every output.
- `bash bench/run.sh` passed: 24 untimed executions and 120 timed executions (eight programs, three runtimes, five rounds). Every timed output matched. The sources and runner were committed as `ed174ae9a78c1fe0fd4d24663ca6f7eb4544a1ba` before this run; the working tree was clean.
- Each new workload (`primes`, `number_sum_sort`, `map_200k`, `json_encode`) was also built with `go run ./cmd/adamic build bench/NAME.a -o BINARY --sanitize` and run with `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1` and `UBSAN_OPTIONS=halt_on_error=1`. All exited zero, emitted no sanitizer stderr, and matched Node's checksum. Subsequent source changes were formatting only; the timed run checks the final formatted sources.
- Identical scratch `.ts` copies of the four new `.a` files passed cohere's type, lint and format checks with the root compiler options, prelude and CohereSettings.json: `go run ./command/cohere --no-fix --tsconfig SCRATCH/tsconfig.json SCRATCH/primes.ts SCRATCH/number_sum_sort.ts SCRATCH/map_200k.ts SCRATCH/json_encode.ts` from `cohere/`. Naming only these four files avoided unrelated prelude lint findings. No prelude or other repository source was changed.
- Controlled temporary executables exercised the runner end to end. The baseline produced five samples per runtime. All eight mutants failed the run and preserved the prior report: wrong native output, wrong Node output, wrong Bun output, empty output, exit 7, timeout, a checksum changing only in timed round 3, and an unrelated JSON build failure. A separate explicit JSON.stringify refusal was correctly listed under not measured while the other program completed.
- No oracle fixtures or fixture registrations were added or changed, so `internal/oracle/counts.md` regeneration was not applicable. No whole-package tests or full gate were run; no compiler/runtime files were changed.

The JSON encode differences (8% against Node, 3% against Bun) are too small to
claim a repeatable loss on this shared machine. They remain reported as losses
in this particular best-of-five sample, alongside every raw timing.

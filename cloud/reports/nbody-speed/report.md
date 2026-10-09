# N-body: borrowed element reads lost the integer lookup

The shipped release build on `origin/area/runtime` (`ae57b2d84bc8ee572830a49f4e9a7f78737f7205`) reproduces the reported native loss. The initial three-runtime run measured native 0.175326 s versus Bun 0.136174 s (1.29×), close to the 1.32× in `runtime/honest-benchmarks` at `fb714279`. The benchmark source is unchanged: `bench/nbody.ts`, SHA256 `d10410ce62f12bfe0c1de964d49a34222de8e1040986f347561f43779792a09c`.

`borrowElement` had its own array-slot lookup code. Unlike `arrayIndexSlot`, it did not recognize `Locals[index].Counter`. Thus n-body's already-integer `i` and `j` went through `adamic_array_at(double)`, converting to double, checking whole-number status and converting back. This happens on five outer and ten inner element reads per simulation step. The small emitter fix shares `arrayIndexSlot`, including its existing relative-index behavior and signed integer bounds checks. No runtime or arithmetic semantics change.

## Final measurements

| Runtime | Best wall seconds | User seconds at that sample |
|---|---:|---:|
| Native before | 0.176896 | 0.176625 |
| Native after | 0.145179 | 0.144829 |
| Node v24.19.0 | 0.711022 | 0.700995 |
| Bun 1.3.14 | 0.140945 | 0.145457 |

Native takes 17.9% less wall time (1.22× speedup). Native/Bun falls from 1.255× to 1.030×; native/Node falls from 0.249× to 0.204×. The last 3% versus Bun is within the spread on this shared machine; this is not a claim of repeatable parity. Four of five paired rounds improved, with one noisy final round. Every raw wall/user sample is in [measurements.json](measurements.json).

Machine: `3ee19de862f2`, AMD EPYC 9V74 80-Core Processor, Linux 6.18.44 x86-64/glibc 2.41; affinity CPUs 0–4; cgroup `cpu.max = 400000 100000` (four CPU quota), 17.6 GB memory. Go 1.27.1, clang 20.1.8, perf 6.12.107. Final load averages (1/5/15 minutes) before: 1.067/3.474/2.759; after: 1.056/3.392/2.740. Initial reproduction and diagnostic loads are recorded in their JSON files.

Method: shipped `go run ./cmd/adamic build`, `-O2`, ThinLTO, `-ffp-contract=off`, no sanitizers/counters for timings. One untimed validation per executable, then five sequential rounds rotating the first runtime among before/after/Node/Bun. Fresh processes; Python `perf_counter` around launch and wait; child user CPU from `getrusage`. Best means minimum wall time; its user time is reported alongside it. Startup, checksum formatting and output are included. All outputs must match these exact bytes:

```text
-0.169075164
-0.169086185
```

No pinning or isolation. No builds, tests, setup or profiles ran alongside the final timings. `NODE_OPTIONS`, `BUN_OPTIONS`, and `ADAMIC_THREADS` are cleared for children. Sources run directly in Node and Bun. [measure.py](measure.py) records commands, affinity, quota, load and every sample, with a checksum check on each execution.

## Profile and generated C

Generated C was captured with `go run ./cmd/adamic c bench/nbody.ts`. Before:

```c
adamic_array_at(adamic_local_4_bodies, ((double)adamic_local_21_i));
adamic_array_at(adamic_local_4_bodies, ((double)adamic_local_23_j));
```

After:

```c
adamic_array_at_integer(adamic_local_4_bodies, adamic_local_21_i);
adamic_array_at_integer(adamic_local_4_bodies, adamic_local_23_j);
```

The two energy-loop reads change likewise. The sun's constant index keeps the ordinary numeric lookup. The counted loops themselves already used `int64_t`; the missed optimization was the integer lookup in the borrowing emitter, not counter inference.

Profiles were taken separately from timings using the shipped binaries:

```sh
perf record -o before.data -e cycles:u -- /tmp/nbody-speed/native
perf annotate -i before.data --stdio
perf record -o after.data -e cycles:u -- /tmp/nbody-speed/fixed
perf annotate -i after.data --stdio
```

Hardware cycles worked (829 samples before, 657 after). A second before profile with `-e cpu-clock:u -F 999` captured 213 samples, all in `main`: ThinLTO inlined `advance`. Disassembly was mapped to the generated C's array lookups, numeric field offsets, write checks and magnitude computation. [profile-excerpts.txt](profile-excerpts.txt) preserves the array-entry and arithmetic/write windows before and after. Sample counts are small and instruction sampling has skid; percentages identify regions and stalls, not precise per-instruction costs.

Cost inventory:

- **Object field reads:** already direct `slots[0..6].number`, without slot-cache loads in the hot reads. Writes retain frozen checks and shape-guarded fallbacks; the profile shows zero samples in the cache-miss/name-search blocks. These guards and their cold calls still affect code size and spilling.
- **Double boxing:** absent. Positions, velocities, masses and local arithmetic are unboxed doubles in eight-byte value slots.
- **Integer counter:** already inferred; its fast lookup was missed only in `borrowElement`. The fixed array lookup eliminates the redundant floating-index bounds/whole-number round trips. Loop comparisons and the floating `i + 1` initialization still have conversions.
- **Square root:** no hot libm call. The magnitude is `step / (squared * sqrt(squared))`, emitted as `sqrtsd`, multiply and `divsd`. Before, 21.28% of cycle samples landed on the instruction after `sqrtsd` and 18.62% after `divsd`; after, 16.16% and 25.30%. This serial floating-point dependency remains the largest sampled region. No reciprocal approximation, reassociation or FMA was introduced.
- **Retains:** none on present-body reads in `advance`. The body and other are borrowed, with NULL owners for their unused allocation fallbacks. Conditional owner releases remain, but the owners stay NULL for this workload.

A scratch-only diagnostic removed nine write shape guards from generated C, keeping all other code and release flags. Its best time was 0.167071 s versus simultaneously interleaved native 0.173896 s, Node 0.723196 s and Bun 0.138128 s; only a 3.9% difference within noise. See [diagnostic.json](diagnostic.json). It did not close the gap, and is not a safe general optimization: `SetProperty` lacks a presence proof, and optional missing fields must retain their explicit failure. It was not applied to repository code. Wider emitter changes to field presence, frozen-check hoisting or arithmetic scheduling are outside this small fix.

## Validation and reproduction

`export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh --wasi-sdk` completed; sourced `/workspace/adamic-tools/env.sh`; Node was v24.19.0. Both requested branches were fetched by name. No AGENTS.md was found in the repository or its workspace ancestors.

Passed:

```sh
go test ./internal/native -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestNbodyBorrowedLoopC|TestLoopBorrowPlan|TestLoopArrayHoldC)$' -count=1
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/borrow_element[^/]*\.a$' -count=1
go run ./cmd/adamic build bench/nbody.ts -o /tmp/nbody-speed/sanitized --sanitize
ASAN_OPTIONS=detect_leaks=1:halt_on_error=1 UBSAN_OPTIONS=halt_on_error=1 /tmp/nbody-speed/sanitized
```

The sanitized benchmark exited zero, with no sanitizer diagnostics and the same checksum. The regression asserts four integer lookups and the sun's remaining numeric lookup. Restoring the original borrowing emitter failed `TestNbodyIndexedElementsBorrow`: “want four borrowed counter lookups on the integer path, got 0”; [mutant.log](mutant.log). The fix was restored and the focused tests rerun successfully.

No fixture sources or registrations changed, so counts.md regeneration does not apply. No whole-package tests or full gate were run. Python AST parsing and `git diff --check` passed.

To reproduce final timing after building a before binary at the base commit and an after binary from this branch:

```sh
python3 cloud/reports/nbody-speed/measure.py \
  --before /tmp/nbody-speed/native --after /tmp/nbody-speed/fixed \
  --node /workspace/adamic-tools/bin/node \
  --bun /tmp/nbody-speed/bun-linux-x64/bun \
  --source bench/nbody.ts --output /tmp/nbody-speed/measurements.json
```

Bun 1.3.14 came from its official `bun-v1.3.14/bun-linux-x64.zip` GitHub release. Perf and its Debian dependencies were extracted into scratch; `LD_LIBRARY_PATH=/tmp/nbody-speed/tools/usr/lib/x86_64-linux-gnu` supplied those libraries. Neither tool altered repository runtime code.

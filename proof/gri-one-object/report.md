# Generator runtime one-object proof — #xxhbdqb

## Result

The unmodified generator-runtime guards pass cold and warm, all 16 configured runtime mutants are caught, and the blank-line equivalence mutant makes every guard fail with the required cached-control mismatch. The compiler census passes. No production or test source changes are retained.

Tested detached HEAD: `0437c02700d3daa46bed1e50452d5242541a8c74`, fetched from `origin emit/generator-guard-one-object-1`. Its two commits above `0c70527596` are `cdb1c69e8` and `0437c0270`. Recursive submodules were initialized successfully:

- `cohere`: `5360753e19e4417675a90e92a1823bea66b13918`
- `cohere/TypeScript`: `d92d9bfee114c80be2c375d72edae966176e3a4f`

Date: 2026-11-12 UTC. Repository: `/workspace/adamic`.

## Setup and procedure

Loaded `/workspace/adamic-tools/env.sh` before repository work and in each runner. Created `/tmp/adamic-gate` outside the repository. Exported `GOFLAGS=-buildvcs=false`. One heavy command ran at a time. Recorded `/proc/loadavg` and `nproc` immediately before every timed command (complete records below). `nproc` reported **5**. GNU time was absent at `/usr/bin/time` and `/bin/time`; used Bash's `time` builtin with `TIMEFORMAT='wall %3R s'`.

`git fetch origin emit/generator-guard-one-object-1`, `git checkout --detach 0437c02700d3daa46bed1e50452d5242541a8c74`, and `git submodule update --init --recursive` completed successfully. Initial and restored worktrees were clean.

Procedure deviation: an initial vet invocation was started prematurely while recursive submodule initialization was still in progress. It exited 1 in 0.070 s; its complete error was:

```text
go: cannot load module cohere/TypeScript/tsc listed in go.work file: open cohere/TypeScript/tsc/go.mod: no such file or directory
```

Its pre-run `/proc/loadavg` was `0.56 0.23 0.08 2/150 879`, and `nproc` was `5`. After initialization completed, the actual formatting and compile gate was run and passed. This preliminary error was a missing setup prerequisite, not a compile error at the fully initialized revision.

## 1–2. Formatting, vet, and test binary

`gofmt -l internal/native` exited 0 and printed **nothing**. `go vet ./internal/native` exited 0 with no diagnostics (219.189 s). `go test -c -o /tmp/native.test ./internal/native` exited 0 with no diagnostics (20.632 s).

The cache is `filepath.Join(location.Output(), "generator-runtime")`, from `cacheDirectory("generator-runtime")` in `internal/native/library.go`, used by `generatorRuntimeGuard` in `internal/native/generator_runtime_test.go`. `internal/buildcache/location/location.go` makes `location.Output()` use `ADAMIC_BUILD_OUTPUT_DIR` when set, otherwise the build cache from `ADAMIC_BUILD_CACHE_DIR` (default: the user cache's `adamic-build`).

For cold runs, **pointed the cache at fresh directories**, rather than removing files. Unset `ADAMIC_BUILD_OUTPUT_DIR` to avoid fallback reads and set `ADAMIC_BUILD_CACHE_DIR` separately for each cold run:

| Run | Generator-runtime directory |
| --- | --- |
| Full cold and following warm | `/tmp/adamic-gate/runtime-cold/generator-runtime` |
| Initialize 1 | `/tmp/adamic-gate/initialize-cold-1/generator-runtime` |
| Initialize 2 | `/tmp/adamic-gate/initialize-cold-2/generator-runtime` |
| Initialize 3 | `/tmp/adamic-gate/initialize-cold-3/generator-runtime` |
| Equivalence mutant cold | `/tmp/adamic-gate/equivalence-cold/generator-runtime` |

These roots were new for their cold runs. No existing repository or user-cache entries were removed. Go's compilation cache was not cleared; “cold” refers to the generator-runtime cache.

## 3–4. Full cold and warm guards

Working directory: `/workspace/adamic/internal/native`. Command for both:

```sh
/tmp/native.test -test.run '^TestGeneratorRuntime' -test.count=1 -test.v
```

Combined stdout/stderr was captured in `/tmp/adamic-gate/cold.log` and `/tmp/adamic-gate/warm.log`; the cold log was also copied verbatim to `/tmp/gr-cold.log`. Cold: **exit 0, 24.438 s wall**. Warm, reusing the cold cache: **exit 0, 8.072 s wall**. No failures or failure messages in either run. Per-test times are Go's reported times; parallel test times do not add to wall time.

| Test | Cold result and time | Warm result and time |
| --- | --- | --- |
| `TestGeneratorRuntimeBorrowed` | PASS 1.71 s | PASS 1.53 s |
| `TestGeneratorRuntimeComplete` | PASS 19.66 s | PASS 3.52 s |
| `TestGeneratorRuntimeDropError` | PASS 1.55 s | PASS 1.41 s |
| `TestGeneratorRuntimeDropSelf` | PASS 1.69 s | PASS 3.57 s |
| `TestGeneratorRuntimeDropStart` | PASS 1.30 s | PASS 1.38 s |
| `TestGeneratorRuntimeDropSuspended` | PASS 1.14 s | PASS 1.60 s |
| `TestGeneratorRuntimeFinishError` | PASS 1.69 s | PASS 1.60 s |
| `TestGeneratorRuntimeFinishSelf` | PASS 1.25 s | PASS 1.54 s |
| `TestGeneratorRuntimeInitialize` | PASS 1.22 s | PASS 1.67 s |
| `TestGeneratorRuntimeIsNull` | PASS 1.66 s | PASS 1.79 s |
| `TestGeneratorRuntimeKill` | PASS 1.43 s | PASS 1.72 s |
| `TestGeneratorRuntimeNextCollection` | PASS 18.96 s | PASS 0.70 s |
| `TestGeneratorRuntimeNextDone` | PASS 1.74 s | PASS 1.66 s |
| `TestGeneratorRuntimeNextDoneBoolean` | PASS 18.83 s | PASS 0.50 s |
| `TestGeneratorRuntimeNextDrop` | PASS 1.87 s | PASS 1.58 s |
| `TestGeneratorRuntimeNextLoudDefault` | PASS 0.00 s | PASS 0.00 s |
| `TestGeneratorRuntimeNextLoudDefault/next_kind` | PASS 0.79 s | PASS 0.47 s |
| `TestGeneratorRuntimeNextLoudDefault/next_null` | PASS 0.50 s | PASS 0.46 s |
| `TestGeneratorRuntimeNextLoudDefault/next_other` | PASS 18.59 s | PASS 0.84 s |
| `TestGeneratorRuntimeNextStrings` | PASS 1.61 s | PASS 1.75 s |
| `TestGeneratorRuntimeNextThrow` | PASS 19.87 s | PASS 3.20 s |
| `TestGeneratorRuntimeReentrant` | PASS 0.62 s | PASS 2.54 s |
| `TestGeneratorRuntimeReturn` | PASS 1.45 s | PASS 1.36 s |
| `TestGeneratorRuntimeReturnWhileRunning` | PASS 0.72 s | PASS 0.73 s |
| `TestGeneratorRuntimeThrow` | PASS 0.67 s | PASS 2.22 s |

All 16 configured mutants logged a catch in both runs. **Guards with configured mutants missing a catch line: none.** Exact cold catch lines:

```text
    generator_runtime_test.go:295: complete mutant caught by heap-use-after-free: ==6131==ERROR: AddressSanitizer: heap-use-after-free on address 0x7c61b8de0028 at pc 0x5601ee88d282 bp 0x7ffd9cf79f80 sp 0x7ffd9cf79f78
    generator_runtime_test.go:426: next_throw mutant caught by heap values leaked: 
    generator_runtime_test.go:401: drop_start mutant caught by Assertion: 
    generator_runtime_test.go:412: next_strings mutant caught by heap values leaked: 
    generator_runtime_test.go:419: next_drop mutant caught by heap-use-after-free: ==6189==ERROR: AddressSanitizer: heap-use-after-free on address 0x7ba794be0030 at pc 0x562c94c46cef bp 0x7ffdb1921020 sp 0x7ffdb1921018
    generator_runtime_test.go:394: is mutant caught by member access within null pointer: generator.c:50:42: runtime error: member access within null pointer of type 'const adamic_heap' (aka 'const struct adamic_heap')
    generator_runtime_test.go:353: self_finish mutant caught by heap values leaked: 
    generator_runtime_test.go:360: self_drop_start mutant caught by heap values leaked: 
    generator_runtime_test.go:342: drop_error mutant caught by heap values leaked: 
    generator_runtime_test.go:336: finish_error mutant caught by heap values leaked: 
    generator_runtime_test.go:434: next_done mutant caught by Assertion: 
    generator_runtime_test.go:313: return mutant caught by Assertion: 
    generator_runtime_test.go:325: borrowed mutant caught by heap-use-after-free: ==6335==ERROR: AddressSanitizer: heap-use-after-free on address 0x7c59d57e0030 at pc 0x562addcd18ea bp 0x7fffac2c39b0 sp 0x7fffac2c39a8
    generator_runtime_test.go:307: drop_suspended mutant caught by heap values leaked: 
    generator_runtime_test.go:319: kill mutant caught by heap-use-after-free: ==6338==ERROR: AddressSanitizer: heap-use-after-free on address 0x7bd113de0028 at pc 0x55d0e17be282 bp 0x7fff28c3c5c0 sp 0x7fff28c3c5b8
    generator_runtime_test.go:301: drop_start mutant caught by heap values leaked:
```

Exact warm catch lines:

```text
    generator_runtime_test.go:426: next_throw mutant caught by heap values leaked: 
    generator_runtime_test.go:295: complete mutant caught by heap-use-after-free: ==6518==ERROR: AddressSanitizer: heap-use-after-free on address 0x7b724bbe0028 at pc 0x5634049fd282 bp 0x7ffc6faa1d70 sp 0x7ffc6faa1d68
    generator_runtime_test.go:360: self_drop_start mutant caught by heap values leaked: 
    generator_runtime_test.go:342: drop_error mutant caught by heap values leaked: 
    generator_runtime_test.go:434: next_done mutant caught by Assertion: 
    generator_runtime_test.go:401: drop_start mutant caught by Assertion: 
    generator_runtime_test.go:419: next_drop mutant caught by heap-use-after-free: ==6590==ERROR: AddressSanitizer: heap-use-after-free on address 0x7bd229de0030 at pc 0x557848567cef bp 0x7ffeb9e87f50 sp 0x7ffeb9e87f48
    generator_runtime_test.go:412: next_strings mutant caught by heap values leaked: 
    generator_runtime_test.go:313: return mutant caught by Assertion: 
    generator_runtime_test.go:336: finish_error mutant caught by heap values leaked: 
    generator_runtime_test.go:325: borrowed mutant caught by heap-use-after-free: ==6651==ERROR: AddressSanitizer: heap-use-after-free on address 0x7b88d89e0030 at pc 0x5625a0ce38ea bp 0x7fff7ef85af0 sp 0x7fff7ef85ae8
    generator_runtime_test.go:319: kill mutant caught by heap-use-after-free: ==6666==ERROR: AddressSanitizer: heap-use-after-free on address 0x7b76655e0028 at pc 0x55b75f8d8282 bp 0x7ffe5e3d1600 sp 0x7ffe5e3d15f8
    generator_runtime_test.go:353: self_finish mutant caught by heap values leaked: 
    generator_runtime_test.go:394: is mutant caught by member access within null pointer: generator.c:50:42: runtime error: member access within null pointer of type 'const adamic_heap' (aka 'const struct adamic_heap')
    generator_runtime_test.go:307: drop_suspended mutant caught by heap values leaked: 
    generator_runtime_test.go:301: drop_start mutant caught by heap values leaked:
```

`Throw`, `Reentrant`, `ReturnWhileRunning`, `NextDoneBoolean`, `NextCollection`, and the three `NextLoudDefault` subtests configure no mutant in this test file, so they do not log a catch line. `ReturnWhileRunning`'s comment describes a separate manual mutant; it is not configured in this run.

## 5. Initialize alone, cold three times

Same working directory, fresh cache root each time; command:

```sh
/tmp/native.test -test.run '^TestGeneratorRuntimeInitialize$' -test.count=1 -test.v
```

| Run | Exit | Test result/time | Wall | Old wall supplied | Reduction |
| --- | --- | --- | --- | --- | --- |
| 1 | 0 | PASS 17.16 s | 17.179 s | 50.2 s | 65.8% |
| 2 | 0 | PASS 18.05 s | 18.059 s | 42.1 s | 57.1% |
| 3 | 0 | PASS 20.57 s | 20.587 s | 51.4 s | 59.9% |

Mean wall time: 18.608 s versus 47.9 s supplied baseline, 61.2% lower (2.57× faster). Each run logged `drop_start mutant caught by Assertion:`. This is an observed comparison, not a controlled benchmark: this executor reports 5 CPUs; the source comment describes the old runs on a quiet 4-CPU box, and recorded loads differ. No clang-call count was measured.

## 6. Equivalence-check mutation

Asserted the exact seam occurred **once** inside `generatorMutantRuntime`, then replaced:

```go
if file.name == "generator.c" || !strings.HasSuffix(file.name, ".c") {
```

with the exact requested text:

```go
if file.name == "generator.c" { file.contents = append([]byte("\n"), file.contents...) }
		if file.name == "generator.c" || !strings.HasSuffix(file.name, ".c") {
```

The temporary mutation was not gofmt'd. `go vet ./internal/native` passed, exit 0, no diagnostics, 33.918 s. Rebuilding `/tmp/native.test` with `go test -c -o /tmp/native.test ./internal/native` passed, exit 0, no diagnostics, 14.168 s.

Ran the full guard command cold from `internal/native`: **exit 1, 18.529 s wall**. Verified programmatically: 25 FAIL results, zero PASS results, and 24 complete `isn't the cached control's` messages, one for every guard invocation (21 top-level guards plus three loud-default subtests); the enclosing loud-default test is the 25th failure. All failures arise from `generatorReplacementHolds`' object comparison propagated through `generatorReplacement.err`. Both objects are 41976 bytes, with differing SHA-256 hashes. No guard escaped the equivalence check.

Complete mutant log, including each failure's whole message and time:

```text
=== RUN   TestGeneratorRuntimeComplete
=== PAUSE TestGeneratorRuntimeComplete
=== RUN   TestGeneratorRuntimeDropStart
=== PAUSE TestGeneratorRuntimeDropStart
=== RUN   TestGeneratorRuntimeDropSuspended
=== PAUSE TestGeneratorRuntimeDropSuspended
=== RUN   TestGeneratorRuntimeReturn
=== PAUSE TestGeneratorRuntimeReturn
=== RUN   TestGeneratorRuntimeKill
=== PAUSE TestGeneratorRuntimeKill
=== RUN   TestGeneratorRuntimeBorrowed
=== PAUSE TestGeneratorRuntimeBorrowed
=== RUN   TestGeneratorRuntimeThrow
=== PAUSE TestGeneratorRuntimeThrow
=== RUN   TestGeneratorRuntimeFinishError
=== PAUSE TestGeneratorRuntimeFinishError
=== RUN   TestGeneratorRuntimeDropError
=== PAUSE TestGeneratorRuntimeDropError
=== RUN   TestGeneratorRuntimeFinishSelf
=== PAUSE TestGeneratorRuntimeFinishSelf
=== RUN   TestGeneratorRuntimeDropSelf
=== PAUSE TestGeneratorRuntimeDropSelf
=== RUN   TestGeneratorRuntimeReentrant
=== PAUSE TestGeneratorRuntimeReentrant
=== RUN   TestGeneratorRuntimeReturnWhileRunning
=== PAUSE TestGeneratorRuntimeReturnWhileRunning
=== RUN   TestGeneratorRuntimeIsNull
=== PAUSE TestGeneratorRuntimeIsNull
=== RUN   TestGeneratorRuntimeInitialize
=== PAUSE TestGeneratorRuntimeInitialize
=== RUN   TestGeneratorRuntimeNextStrings
=== PAUSE TestGeneratorRuntimeNextStrings
=== RUN   TestGeneratorRuntimeNextDrop
=== PAUSE TestGeneratorRuntimeNextDrop
=== RUN   TestGeneratorRuntimeNextThrow
=== PAUSE TestGeneratorRuntimeNextThrow
=== RUN   TestGeneratorRuntimeNextDone
=== PAUSE TestGeneratorRuntimeNextDone
=== RUN   TestGeneratorRuntimeNextDoneBoolean
=== PAUSE TestGeneratorRuntimeNextDoneBoolean
=== RUN   TestGeneratorRuntimeNextCollection
=== PAUSE TestGeneratorRuntimeNextCollection
=== RUN   TestGeneratorRuntimeNextLoudDefault
=== PAUSE TestGeneratorRuntimeNextLoudDefault
=== CONT  TestGeneratorRuntimeComplete
=== CONT  TestGeneratorRuntimeReentrant
=== CONT  TestGeneratorRuntimeThrow
=== CONT  TestGeneratorRuntimeReturn
=== CONT  TestGeneratorRuntimeNextLoudDefault
=== RUN   TestGeneratorRuntimeNextLoudDefault/next_other
=== PAUSE TestGeneratorRuntimeNextLoudDefault/next_other
=== RUN   TestGeneratorRuntimeNextLoudDefault/next_null
=== PAUSE TestGeneratorRuntimeNextLoudDefault/next_null
=== RUN   TestGeneratorRuntimeNextLoudDefault/next_kind
=== PAUSE TestGeneratorRuntimeNextLoudDefault/next_kind
=== CONT  TestGeneratorRuntimeNextCollection
=== NAME  TestGeneratorRuntimeThrow
    generator_runtime_test.go:331: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
=== NAME  TestGeneratorRuntimeReentrant
    generator_runtime_test.go:377: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeThrow (18.34s)
=== CONT  TestGeneratorRuntimeNextDoneBoolean
--- FAIL: TestGeneratorRuntimeReentrant (18.34s)
=== NAME  TestGeneratorRuntimeComplete
    generator_runtime_test.go:296: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeComplete (18.35s)
=== NAME  TestGeneratorRuntimeNextCollection
    generator_runtime_test.go:450: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
=== CONT  TestGeneratorRuntimeNextThrow
=== CONT  TestGeneratorRuntimeNextDone
=== NAME  TestGeneratorRuntimeReturn
    generator_runtime_test.go:314: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeNextCollection (18.34s)
--- FAIL: TestGeneratorRuntimeReturn (18.35s)
=== CONT  TestGeneratorRuntimeNextStrings
=== CONT  TestGeneratorRuntimeNextDrop
=== NAME  TestGeneratorRuntimeNextDone
    generator_runtime_test.go:435: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeNextDone (0.04s)
=== CONT  TestGeneratorRuntimeInitialize
=== NAME  TestGeneratorRuntimeNextDoneBoolean
    generator_runtime_test.go:444: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeNextDoneBoolean (0.05s)
=== CONT  TestGeneratorRuntimeIsNull
=== NAME  TestGeneratorRuntimeNextThrow
    generator_runtime_test.go:427: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeNextThrow (0.05s)
=== CONT  TestGeneratorRuntimeReturnWhileRunning
=== NAME  TestGeneratorRuntimeNextStrings
    generator_runtime_test.go:413: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeNextStrings (0.05s)
=== CONT  TestGeneratorRuntimeFinishSelf
=== NAME  TestGeneratorRuntimeNextDrop
    generator_runtime_test.go:420: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeNextDrop (0.06s)
=== CONT  TestGeneratorRuntimeDropSelf
=== NAME  TestGeneratorRuntimeInitialize
    generator_runtime_test.go:402: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeInitialize (0.04s)
=== CONT  TestGeneratorRuntimeDropSuspended
=== NAME  TestGeneratorRuntimeIsNull
    generator_runtime_test.go:395: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeIsNull (0.04s)
=== CONT  TestGeneratorRuntimeDropError
=== NAME  TestGeneratorRuntimeDropSelf
    generator_runtime_test.go:361: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeDropSelf (0.03s)
=== CONT  TestGeneratorRuntimeFinishError
=== NAME  TestGeneratorRuntimeFinishSelf
    generator_runtime_test.go:354: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeFinishSelf (0.04s)
=== CONT  TestGeneratorRuntimeKill
=== NAME  TestGeneratorRuntimeDropSuspended
    generator_runtime_test.go:308: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeDropSuspended (0.03s)
=== CONT  TestGeneratorRuntimeBorrowed
=== NAME  TestGeneratorRuntimeDropError
    generator_runtime_test.go:343: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeDropError (0.03s)
=== CONT  TestGeneratorRuntimeDropStart
=== NAME  TestGeneratorRuntimeReturnWhileRunning
    generator_runtime_test.go:384: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeReturnWhileRunning (0.07s)
=== CONT  TestGeneratorRuntimeNextLoudDefault/next_other
=== NAME  TestGeneratorRuntimeFinishError
    generator_runtime_test.go:337: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeFinishError (0.03s)
=== CONT  TestGeneratorRuntimeNextLoudDefault/next_kind
=== NAME  TestGeneratorRuntimeKill
    generator_runtime_test.go:320: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeKill (0.03s)
=== CONT  TestGeneratorRuntimeNextLoudDefault/next_null
=== NAME  TestGeneratorRuntimeBorrowed
    generator_runtime_test.go:326: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeBorrowed (0.04s)
=== NAME  TestGeneratorRuntimeDropStart
    generator_runtime_test.go:302: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeDropStart (0.04s)
=== NAME  TestGeneratorRuntimeNextLoudDefault/next_other
    generator_runtime_test.go:460: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
=== NAME  TestGeneratorRuntimeNextLoudDefault/next_kind
    generator_runtime_test.go:460: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
=== NAME  TestGeneratorRuntimeNextLoudDefault/next_null
    generator_runtime_test.go:460: a mutant's replaced generator.o isn't the whole build's: the replace path's generator.o in /tmp/adamic-gate/adamic-generator-replacement-2478332553/runtime.a (41976 bytes, sha256 76fcfb31c891b423485fe5540f4130d1c139436b9479b38908f7d29045512d23) isn't the cached control's in /tmp/adamic-gate/equivalence-cold/generator-runtime/d57b5d263d1b91f08052d5737b284a2c55db692d6f96085248f73ebd97714555/runtime.a (41976 bytes, sha256 7297409f9523725f77139d96a11e5a1941f414eb5a35604a09bc253e6c625bb1)
--- FAIL: TestGeneratorRuntimeNextLoudDefault (0.00s)
    --- FAIL: TestGeneratorRuntimeNextLoudDefault/next_other (0.04s)
    --- FAIL: TestGeneratorRuntimeNextLoudDefault/next_kind (0.03s)
    --- FAIL: TestGeneratorRuntimeNextLoudDefault/next_null (0.03s)
FAIL
```

Restored `internal/native/generator_runtime_test.go` byte-for-byte from the saved pre-mutation copy. `git diff --exit-code` returned **0** with no output. `git status --short` was also empty before creating this report. The `/tmp/native.test` binary remains the temporary mutant build; it is outside the repository and was not used for the subsequent census command.

## 7. Compiler census

After restoring the source, ran from the repository root:

```sh
go test ./internal/compilercensus -count=1 -timeout 30m
```

**Exit 0, 3.442 s wall. Failures: none.** Complete output:

```text
ok  	github.com/system-inc/adamic/internal/compilercensus	2.655s
```

## Timing and load evidence

Each record below contains UTC start, complete `/proc/loadavg`, `nproc`, shell-escaped command, Bash wall time, and exit. Formatting/vet/build/census ran from the repository root; runtime commands ran from `internal/native`. Commands were sequential. The setup-incomplete preliminary vet record is given separately above.

### format

```text
2026-11-12T07:34:52Z
2.03 1.07 0.47 1/139 961
5
command: gofmt -l internal/native
wall 0.157 s
exit 0
```

### vet

```text
2026-11-12T07:34:52Z
2.03 1.07 0.47 1/139 973
5
command: go vet ./internal/native
wall 219.189 s
exit 0
```

### compile

```text
2026-11-12T07:38:31Z
4.11 3.49 1.63 1/141 3949
5
command: go test -c -o /tmp/native.test ./internal/native
wall 20.632 s
exit 0
```

### cold

```text
2026-11-12T07:38:59Z
4.42 3.65 1.74 1/140 5895
5
command: /tmp/native.test -test.run \^TestGeneratorRuntime -test.count=1 -test.v
wall 24.438 s
exit 0
```

### warm

```text
2026-11-12T07:39:23Z
3.58 3.50 1.74 1/145 6346
5
command: /tmp/native.test -test.run \^TestGeneratorRuntime -test.count=1 -test.v
wall 8.072 s
exit 0
```

### initialize-1

```text
2026-11-12T07:39:31Z
3.87 3.57 1.78 1/149 6719
5
command: /tmp/native.test -test.run \^TestGeneratorRuntimeInitialize\$ -test.count=1 -test.v
wall 17.179 s
exit 0
```

### initialize-2

```text
2026-11-12T07:39:49Z
3.24 3.44 1.77 1/149 6915
5
command: /tmp/native.test -test.run \^TestGeneratorRuntimeInitialize\$ -test.count=1 -test.v
wall 18.059 s
exit 0
```

### initialize-3

```text
2026-11-12T07:40:07Z
2.60 3.28 1.75 1/149 7107
5
command: /tmp/native.test -test.run \^TestGeneratorRuntimeInitialize\$ -test.count=1 -test.v
wall 20.587 s
exit 0
```

### mutant-vet

```text
2026-11-12T07:40:45Z
1.61 2.96 1.70 2/149 7371
5
command: go vet ./internal/native
wall 33.918 s
exit 0
```

### mutant-compile

```text
2026-11-12T07:41:18Z
2.39 3.01 1.76 1/150 7458
5
command: go test -c -o /tmp/native.test ./internal/native
wall 14.168 s
exit 0
```

### equivalence

```text
2026-11-12T07:41:33Z
2.51 3.01 1.78 1/151 7545
5
command: /tmp/native.test -test.run \^TestGeneratorRuntime -test.count=1 -test.v
wall 18.529 s
exit 1
```

### census

```text
2026-11-12T07:42:08Z
1.70 2.76 1.74 1/148 7840
5
command: go test ./internal/compilercensus -count=1 -timeout 30m
wall 3.442 s
exit 0
```

## Report delivery

Only `proof/gri-one-object/report.md` is intended for the report commit on the new branch `proof/gri-one-object-report`, created directly from the tested detached HEAD. No implementation changes are retained; no other branch is authorized for push.

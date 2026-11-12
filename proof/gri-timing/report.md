# Generator runtime timing report

Measured on 2026-11-12 UTC in `/workspace/adamic`.

## Base and environment

- Repository: `https://github.com/system-inc/adamic.git`.
- Base: `origin/main` at `fb1c8d3b72de2686c486cf97685079589308dafd` (fetched, then checked out detached before building or measuring).
- Ran `git submodule update --init --recursive`: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` before setup/build/test commands; exported `GOFLAGS=-buildvcs=false`. The script supplies the tools PATH, `GOTOOLCHAIN=auto`, and `TMPDIR=/tmp/adamic-gate`.
- Created `/tmp/adamic-gate` outside the repository. Binary, scripts, caches, package extraction, and raw measurement logs stayed outside the repository.
- Go: `go1.27.1 linux/amd64`. Clang: `20.1.8`, LLVM revision `87f0227cb60147a26a1eeb4fb06e3b505e9c7261`, installed under `/workspace/adamic-tools/llvm`.
- OS: `Linux 6.18.57 x86_64 GNU/Linux`. `nproc` returned **5 before every run**. Cgroup `cpu.max`: `400000 100000` (four CPU equivalents); `memory.max`: `17179869184` (16 GiB).
- All builds and test runs were sequential. Submodule downloads/indexing finished before the one-time Go build; the build finished before timings. No other heavy work was launched concurrently. The full group retains its own default Go test parallelism.

## Build and timing commands

Built the test binary exactly once from the repository root (success):

```sh
go test -c -o /tmp/native.test ./internal/native
```

**Timing-tool substitution:** `/usr/bin/time` was absent, so the first measurement attempt failed before the test launched. This environment also lacks sudo; the APT package index could not locate `time`. Downloaded the official Debian `time_1.9-0.2_amd64.deb` from `https://deb.debian.org/debian/pool/main/t/time/` and extracted it under `/tmp/adamic-gate/gnu-time`. Used its GNU time executable with `-v` for every successful timing; it was not installed system-wide. No test-binary rebuild was needed.

Tests ran with working directory `/workspace/adamic/internal/native` so relative testdata paths resolve. Immediately before every timed run, recorded `date -u +%FT%TZ`, `cat /proc/loadavg`, and `nproc`. The successful measurement commands were:

```sh
TIME=/tmp/adamic-gate/gnu-time/usr/bin/time
"$TIME" -v /tmp/native.test -test.run '^TestGeneratorRuntimeInitialize$' -test.count=1 -test.v
"$TIME" -v /tmp/native.test -test.run '^TestGeneratorRuntime' -test.count=1 -test.v
```

The actual script redirected GNU time output with `-o <external-log>` and test stdout/stderr to separate external logs. No tracing or additional instrumentation was enabled for these seven runs. Wall times below come from GNU time; test times come from verbose `--- PASS` lines. Peak memory is GNU time maximum RSS in KiB (the maximum reported for an individual process in the measured process tree, not aggregate simultaneous memory across parallel compilers). Every successful measured command exited 0 and reported `PASS`.

## Cache isolation

`internal/native/generator_runtime_test.go:38` calls `cacheDirectory("generator-runtime")`. `internal/native/library.go:253` returns `filepath.Join(location.Output(), name)`. `internal/buildcache/location/location.go:26,43,70` shows:

- `ADAMIC_BUILD_CACHE_DIR` selects the build-cache root (otherwise the user cache directory plus `adamic-build`).
- `ADAMIC_BUILD_OUTPUT_DIR` overrides the output root; lookups can fall back to the build-cache root.

For **each cold run**, used `mktemp -d /tmp/adamic-gate/<label>.XXXXXX` and exported **both variables to the same fresh empty root**. Thus the test used `<root>/generator-runtime` and could not fall back to an existing runtime cache. No existing cache files or directories were removed. The three warm runs reused cold run 3’s root without clearing anything. The full-group run and extra traced run each received another fresh root.

| Run | Fresh root (generator-runtime cache is the child named `generator-runtime`) |
| --- | --- |
| cold-1 | `/tmp/adamic-gate/cold-1.JGakpT` |
| cold-2 | `/tmp/adamic-gate/cold-2.J1qnNF` |
| cold-3 | `/tmp/adamic-gate/cold-3.Dku999` |
| all-cold | `/tmp/adamic-gate/all-cold.NxbVZs` |
| trace-cold | `/tmp/adamic-gate/trace-cold.I2QnJS` |

For example, a cold run was configured with:

```sh
export ADAMIC_BUILD_CACHE_DIR="$(mktemp -d /tmp/adamic-gate/cold-1.XXXXXX)"
export ADAMIC_BUILD_OUTPUT_DIR="$ADAMIC_BUILD_CACHE_DIR"
```

“Cold” here means an empty generator-runtime disk cache, including both control and mutant archives. OS page caches, compiler installations, and Go build caches were not flushed. The test still compiles/links its temporary harnesses on warm runs.

## Initialize: three cold and three warm runs

| Run | Start UTC | Pre-run `/proc/loadavg` (complete line) | nproc | Wall (s) | Test (s) | Peak RSS (KiB) |
| --- | --- | --- | ---: | ---: | ---: | ---: |
| cold-1 | 2026-11-12T07:05:32Z | `0.85 3.30 2.19 2/141 3458` | 5 | 50.15 | 50.13 | 75400 |
| cold-2 | 2026-11-12T07:06:22Z | `1.11 2.99 2.14 1/142 3790` | 5 | 42.11 | 42.08 | 75296 |
| cold-3 | 2026-11-12T07:07:04Z | `1.15 2.74 2.10 1/142 4122` | 5 | 51.39 | 51.38 | 75440 |
| warm-1 | 2026-11-12T07:07:56Z | `1.11 2.49 2.04 1/141 4454` | 5 | 5.58 | 5.57 | 49072 |
| warm-2 | 2026-11-12T07:08:01Z | `1.18 2.48 2.04 1/141 4630` | 5 | 5.03 | 5.03 | 49076 |
| warm-3 | 2026-11-12T07:08:06Z | `1.17 2.45 2.04 1/141 4808` | 5 | 4.66 | 4.65 | 49076 |

Median wall time: **50.15 s cold**, **5.03 s warm** (9.97× ratio). With only three runs per condition, these are descriptive measurements, not a precision estimate.

## All TestGeneratorRuntime* tests, once cold

- Start: `2026-11-12T07:08:11Z`; pre-run `/proc/loadavg`: `1.15 2.43 2.03 1/141 4987`; `nproc`: `5`.
- Wall: **138.57 s** (`2:18.57`); peak RSS: **75516 KiB**; exit 0, all passed.
- The tests call `t.Parallel()`: durations overlap and include cache-build contention. They must not be summed to reconstruct wall time. Go reports the parent of the parallel subtests separately from its children.

| Test | Own reported time (s) |
| --- | ---: |
| `TestGeneratorRuntimeBorrowed` | 62.59 |
| `TestGeneratorRuntimeComplete` | 64.09 |
| `TestGeneratorRuntimeDropError` | 36.10 |
| `TestGeneratorRuntimeDropSelf` | 38.82 |
| `TestGeneratorRuntimeDropStart` | 5.25 |
| `TestGeneratorRuntimeDropSuspended` | 35.78 |
| `TestGeneratorRuntimeFinishError` | 30.63 |
| `TestGeneratorRuntimeFinishSelf` | 35.50 |
| `TestGeneratorRuntimeInitialize` | 38.35 |
| `TestGeneratorRuntimeIsNull` | 43.30 |
| `TestGeneratorRuntimeKill` | 38.10 |
| `TestGeneratorRuntimeNextCollection` | 25.59 |
| `TestGeneratorRuntimeNextDone` | 36.77 |
| `TestGeneratorRuntimeNextDoneBoolean` | 1.46 |
| `TestGeneratorRuntimeNextDrop` | 37.65 |
| `TestGeneratorRuntimeNextLoudDefault` | 0.00 |
| `TestGeneratorRuntimeNextLoudDefault/next_kind` | 0.87 |
| `TestGeneratorRuntimeNextLoudDefault/next_null` | 0.65 |
| `TestGeneratorRuntimeNextLoudDefault/next_other` | 0.53 |
| `TestGeneratorRuntimeNextStrings` | 41.23 |
| `TestGeneratorRuntimeNextThrow` | 31.00 |
| `TestGeneratorRuntimeReentrant` | 25.31 |
| `TestGeneratorRuntimeReturn` | 34.47 |
| `TestGeneratorRuntimeReturnWhileRunning` | 0.57 |
| `TestGeneratorRuntimeThrow` | 25.56 |

## Clang invocation count and durations

Ran one additional cold Initialize test, using the same binary and another isolated cache, under:

```sh
"$TIME" -v strace -f -ttt -s 32768 -e trace=execve,exit_group -o /tmp/adamic-gate/trace.log /tmp/native.test -test.run '^TestGeneratorRuntimeInitialize$' -test.count=1 -test.v
```

Start `2026-11-12T07:10:49Z`; pre-run `/proc/loadavg`: `2.48 2.60 2.16 1/143 7678`; `nproc`: 5. The traced test passed: **96.70 s wall**, **96.50 s test**, **75348 KiB peak RSS**. These durations include substantial strace overhead and are not substituted into the untraced results above.

**157 clang driver invocations**: 1 version query, 2 dependency scans (`-M`, one per control/mutant runtime), 152 runtime compilations (76 sources × 2), and 2 harness compile/link invocations. In addition, the trace records **154 clang `-cc1` frontend subprocess executions** (76 dependency-preprocessing subprocesses × 2, plus one for each harness). Therefore there are **311 successful clang executable `execve` events** if internal frontend executions are included. Archiver and linker executions are not counted as clang.

Durations below are elapsed seconds from the timestamp of each successful clang `execve` entry to that PID’s `exit_group` entry. Unfinished/resumed `execve` lines were paired and checked for success. This measures the driver’s lifetime including any subprocess work; frontend times are nested inside the corresponding dependency scan or harness time and must not be added to driver times. Values are rounded to six decimal places; timestamp resolution does not imply that accuracy.

| Driver operation | Control (s) | Mutant (s) |
| --- | ---: | ---: |
| Version query (once, before either build) | 0.069823 | — |
| Dependency scan (`-M` over all 76 sources) | 6.775167 | 7.315847 |
| Harness compile/link | 10.052438 | 7.728201 |

| Nested frontend operation | Control (s) | Mutant (s) |
| --- | ---: | ---: |
| Harness `main.c` frontend | 0.489314 | 0.300971 |

Each row below gives both runtime compile driver lifetimes and both nested dependency frontend lifetimes for that source, accounting for all remaining clang executions:

| Runtime source | Control compile (s) | Mutant compile (s) | Control dependency frontend (s) | Mutant dependency frontend (s) |
| --- | ---: | ---: | ---: | ---: |
| `adamic.c` | 0.448614 | 0.310566 | 0.119564 | 0.141387 |
| `array.c` | 0.481474 | 0.486867 | 0.097708 | 0.067889 |
| `array_from.c` | 0.143979 | 0.127560 | 0.071512 | 0.090808 |
| `async.c` | 0.685296 | 0.711509 | 0.130012 | 0.133939 |
| `bitwise.c` | 0.158413 | 0.137266 | 0.089526 | 0.083460 |
| `case.c` | 0.265270 | 0.267939 | 0.107731 | 0.073766 |
| `census_small.c` | 0.129610 | 0.124877 | 0.088328 | 0.133762 |
| `class_features.c` | 0.230476 | 0.228076 | 0.152458 | 0.074127 |
| `class_inheritance.c` | 0.145118 | 0.136081 | 0.072324 | 0.103816 |
| `class_static.c` | 0.155532 | 0.157565 | 0.079115 | 0.066151 |
| `closure.c` | 0.238199 | 0.392410 | 0.061870 | 0.128584 |
| `construction.c` | 0.186913 | 0.303617 | 0.068244 | 0.092698 |
| `count.c` | 0.130671 | 0.177530 | 0.073670 | 0.104422 |
| `directory.c` | 0.290566 | 0.452744 | 0.081625 | 0.099607 |
| `dtoa.c` | 0.718707 | 0.856833 | 0.063764 | 0.063723 |
| `exceptions.c` | 0.211253 | 0.594430 | 0.076713 | 0.091387 |
| `from_codes.c` | 0.176925 | 0.439084 | 0.078090 | 0.181000 |
| `generator.c` | 0.207252 | 0.293473 | 0.062716 | 0.142477 |
| `graph_regions.c` | 1.003859 | 0.895849 | 0.081728 | 0.083602 |
| `heap.c` | 0.528429 | 0.295730 | 0.102823 | 0.121910 |
| `host_runtime.c` | 0.293763 | 0.323742 | 0.094178 | 0.108646 |
| `hypot.c` | 0.174395 | 0.322918 | 0.082681 | 0.114823 |
| `ieee754.c` | 0.776172 | 0.780986 | 0.090918 | 0.070260 |
| `input.c` | 0.359872 | 0.300444 | 0.080133 | 0.089142 |
| `json_decode.c` | 0.886279 | 0.852810 | 0.084918 | 0.092699 |
| `json_encode.c` | 0.770903 | 0.392798 | 0.063639 | 0.075339 |
| `json_objects.c` | 0.287734 | 0.263337 | 0.072666 | 0.085143 |
| `json_stringify.c` | 0.720497 | 0.507046 | 0.077866 | 0.078892 |
| `library_array.c` | 0.239361 | 0.175569 | 0.073762 | 0.095324 |
| `library_intl.c` | 0.148875 | 0.158888 | 0.087504 | 0.192735 |
| `library_language.c` | 0.183681 | 0.218904 | 0.068404 | 0.085221 |
| `library_math_number.c` | 0.209129 | 0.240018 | 0.091673 | 0.098020 |
| `library_object.c` | 0.368858 | 0.396690 | 0.065897 | 0.107117 |
| `map.c` | 0.391961 | 0.527135 | 0.068926 | 0.097275 |
| `map_set.c` | 0.479721 | 0.267737 | 0.078057 | 0.090004 |
| `math.c` | 0.406197 | 0.140897 | 0.079531 | 0.093947 |
| `maybe.c` | 0.298509 | 0.138729 | 0.098510 | 0.079628 |
| `node_buffer.c` | 0.573913 | 0.331636 | 0.187088 | 0.070636 |
| `node_crypto.c` | 0.382350 | 0.261922 | 0.069795 | 0.114418 |
| `node_cwd_wasi.c` | 0.139336 | 0.096067 | 0.075503 | 0.089348 |
| `node_fs_directory.c` | 0.695566 | 0.363682 | 0.109207 | 0.118111 |
| `node_fs_file.c` | 0.652981 | 0.700147 | 0.125747 | 0.113666 |
| `node_path.c` | 0.397168 | 0.422424 | 0.098600 | 0.098298 |
| `node_process.c` | 0.348633 | 0.362384 | 0.115658 | 0.101977 |
| `normalize.c` | 0.614333 | 0.695241 | 0.079816 | 0.082901 |
| `number.c` | 0.306158 | 0.309383 | 0.061091 | 0.068650 |
| `object.c` | 0.512016 | 0.560473 | 0.086745 | 0.098795 |
| `parallel.c` | 0.523734 | 0.427258 | 0.152551 | 0.128579 |
| `parse.c` | 0.286660 | 0.345619 | 0.082823 | 0.068205 |
| `process.c` | 0.214431 | 0.274275 | 0.076462 | 0.099197 |
| `program_region.c` | 0.274352 | 0.167211 | 0.073383 | 0.082435 |
| `radix.c` | 0.210549 | 0.235538 | 0.070908 | 0.066585 |
| `record.c` | 0.479172 | 0.415308 | 0.080614 | 0.097687 |
| `regexp.c` | 1.825668 | 1.974716 | 0.068376 | 0.073721 |
| `regexp_regular.c` | 0.460331 | 0.399367 | 0.073314 | 0.069713 |
| `region.c` | 0.234813 | 0.357163 | 0.100921 | 0.090945 |
| `set.c` | 0.186903 | 0.161771 | 0.076970 | 0.067033 |
| `share.c` | 0.290418 | 0.305880 | 0.096036 | 0.098243 |
| `slab_quarantine.c` | 0.075274 | 0.129377 | 0.056515 | 0.061069 |
| `sort.c` | 0.771506 | 0.438006 | 0.076484 | 0.082949 |
| `sort_undefined.c` | 0.534599 | 0.182382 | 0.079636 | 0.079185 |
| `spread.c` | 0.250241 | 0.192703 | 0.070038 | 0.059108 |
| `stack.c` | 0.830532 | 0.175350 | 0.083580 | 0.083839 |
| `string.c` | 1.200157 | 0.656027 | 0.078900 | 0.080386 |
| `string_append.c` | 0.327243 | 0.168554 | 0.075300 | 0.086918 |
| `string_from.c` | 0.201402 | 0.140643 | 0.079658 | 0.083392 |
| `string_index.c` | 0.366987 | 0.452586 | 0.080769 | 0.064612 |
| `string_share.c` | 0.311694 | 0.335636 | 0.107453 | 0.073517 |
| `tsgo.c` | 0.049554 | 0.083051 | 0.028505 | 0.036355 |
| `typed_array.c` | 0.229566 | 0.273369 | 0.068868 | 0.070805 |
| `typed_array_sort.c` | 0.244242 | 0.256629 | 0.097676 | 0.081598 |
| `union.c` | 0.829508 | 0.764202 | 0.094437 | 0.108276 |
| `utf8.c` | 0.184131 | 0.178307 | 0.080783 | 0.071828 |
| `view_unions_mixed.c` | 0.200113 | 0.167675 | 0.092088 | 0.095099 |
| `view_unions_untagged.c` | 0.548690 | 0.603057 | 0.086101 | 0.077609 |
| `weak.c` | 0.225640 | 0.215897 | 0.070522 | 0.086932 |

## Reproducibility and scope

Only this report is added to the repository. Raw logs remain under `/tmp/adamic-gate` in this measurement environment; they are not committed. SHA-256 fingerprints:

| Artifact | SHA-256 |
| --- | --- |
| Test binary | `f59870490542fee96531d68c891fb3b19a1fd903b109b0f94ed4c722dda64145` |
| GNU time Debian package | `10257c6fe888e6578550b1f20705d282dd630b76f8df0c006aec3031fefe5e23` |
| Raw clang trace | `959a1747a6df15e0374c326321bb49710444d9292add35d596c1632e3b7c7258` |

The report branch `proof/gri-timing-report` was created directly from the base SHA above. No compiler, runtime, test, dependency, or configuration source was edited.

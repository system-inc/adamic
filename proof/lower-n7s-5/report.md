# Captured union proof — #n7sx9gz
2026-11-12 UTC. Tested HEAD `02c50297a07758a29bc1f1201477bd275ab30278` against main `730e7c26e590af0c57c545ba0cb905b09e4bee54`. HEAD is one commit above that main. All heavy commands were sequential; standalone measurements and the two-worker admission gate ran without any other build or test running.
## Decisive results
- Admission delta: **PASS**, direct exit **0**, `void=false`, no void reasons, `diff_count=8`, admitted **8**, newly refused **0**. No timeouts at either classification or runtime; no timeout reruns were applicable.
- HEAD pins: **PASS** for all 16 lowering pin/dialect subtests, the maybe-boolean refusal guard, and all 64 oracle backend subtests.
- M3: caught by both sanitized and counted runs of `alternating`, `by_value`, and `released`, in both dialects. Alternating counts are `5` allocations, `3` frees, `0` in regions. All other pins pass these checks. Mutant restored; `git diff --exit-code` exited 0 with no output.
- Standalone builds: all seven compiles and runs passed, including alternating in 8.905059 / 2.079574 / 2.482430 seconds.
- Formatting and vet passed; no captured-union count records or other oracle Go references exist.
## Setup and step 1
Loaded `/workspace/adamic-tools/env.sh` first. It set `TMPDIR=/tmp/adamic-gate`; created that directory outside the repository before invoking Go. Fetched `origin compiler/captured-union-5`, checked out the exact HEAD detached, then completed `git submodule update --init --recursive`, exit 0. Initial and post-restoration working trees were clean. The earlier abbreviated `9d31ad14` is absent from this clone, so its stated relationship to HEAD was not independently verified; no evidence from that earlier proof was reused.
```sh
git fetch origin compiler/captured-union-5
git checkout --detach 02c50297a07758a29bc1f1201477bd275ab30278
git submodule update --init --recursive
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

`gofmt` exit 0, stdout/stderr empty. `go vet` exit 0, stdout/stderr empty.
```text
go version go1.27.1 linux/amd64
node v24.19.0
clang 20.1.8
tsc Version 6.0.3
cohere 5360753e19e4417675a90e92a1823bea66b13918
cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f
```

Provisioned upstream TypeScript outside the repository at `/tmp/lower-n7s-5/typescript`, detached at the exact requested commit, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that path. Copied `stage3/api/package.json` and `package-lock.json` to `/tmp/lower-n7s-5/api` and ran `npm ci --ignore-scripts --prefix /tmp/lower-n7s-5/api`. Prepended its `node_modules/.bin` to PATH, giving tsc 6.0.3. The admission gate also provisioned its own temporary HEAD checkout from the pinned package lock.
```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lower-n7s-5/typescript
```

Exported those CA settings for the proof commands and the gate. Because `toolchain.Env()` filters npm’s environment, an external `/tmp/lower-n7s-5/bin/npm` wrapper on PATH exports the same three settings before execing the original npm, preserving them for npm and its children. TLS verification stayed enabled.
## Step 2a: lowering pins
```sh
go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops' -count=1 -v -timeout 30m
```

Exit 0. `TestCapturedUnionLowersToOneCell` and `TestCapturedMaybeBooleanStillStops`: PASS. Each graph/capture log line follows verbatim, associated with its subtest.
| Pin / dialect | Result | Graph cell log | Closure log |
| --- | --- | --- | --- |
| class.ts | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 5 closures carry its cell` |
| class.a | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 5 closures carry its cell` |
| object.ts | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 5 closures carry its cell` |
| object.a | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 5 closures carry its cell` |
| by_value.ts | PASS | `captured_union_test.go:56: active: graph cell false` | `captured_union_test.go:66: active: 3 closures carry its cell` |
| by_value.a | PASS | `captured_union_test.go:56: active: graph cell false` | `captured_union_test.go:66: active: 3 closures carry its cell` |
| released.ts | PASS | `captured_union_test.go:56: active: graph cell false` | `captured_union_test.go:66: active: 2 closures carry its cell` |
| released.a | PASS | `captured_union_test.go:56: active: graph cell false` | `captured_union_test.go:66: active: 2 closures carry its cell` |
| nested.ts | PASS | `captured_union_test.go:56: extended: graph cell true` | `captured_union_test.go:66: extended: 2 closures carry its cell` |
| nested.a | PASS | `captured_union_test.go:56: extended: graph cell true` | `captured_union_test.go:66: extended: 2 closures carry its cell` |
| alternating.ts | PASS | `captured_union_test.go:56: active: graph cell false` | `captured_union_test.go:66: active: 2 closures carry its cell` |
| alternating.a | PASS | `captured_union_test.go:56: active: graph cell false` | `captured_union_test.go:66: active: 2 closures carry its cell` |
| graph.ts | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 2 closures carry its cell` |
| graph.a | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 2 closures carry its cell` |
| cycle.ts | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 1 closures carry its cell` |
| cycle.a | PASS | `captured_union_test.go:56: active: graph cell true` | `captured_union_test.go:66: active: 1 closures carry its cell` |

No FAIL results, so there are no failing got/want diagnostics for step 2a. Complete output appears below.
## Step 2b: uncached oracle pins at HEAD
```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees' -count=1 -v -timeout 40m
```

Exit 0. Parent and all pin/dialect subtests PASS. Each backend result and logged counts line:
| Pin / dialect | Native | Sanitized | Counted | JavaScript | Counts line |
| --- | --- | --- | --- | --- | --- |
| class.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| class.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| object.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| object.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| by_value.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0` |
| by_value.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0` |
| released.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0` |
| released.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0` |
| nested.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0` |
| nested.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0` |
| alternating.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 5, frees 5, in regions 0` |
| alternating.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 5, frees 5, in regions 0` |
| graph.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0` |
| graph.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0` |
| cycle.ts | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0` |
| cycle.a | PASS | PASS | PASS | PASS | `captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0` |

No FAIL results, so there are no failing got/want diagnostics for step 2b. The test pins Node exit 0 and exact expected stdout, then compares backends. All 32 native and 32 Node cache lookups were misses under `ADAMIC_GATE_UNCACHED=1`. The shared Go-built JSON helper used normal product caching; this does not substitute for a native or Node pin observation.
## Step 3: M3 overwritten value kept
Asserted the specified old three-line text occurs exactly once: `old text occurrences: 1`. Replaced only that occurrence in `internal/native/emit_locals.go`:
```go
		} else {
			e.line("adamic_release(%s);", old)
		}
```

with:
```go
		} else {
			e.line("(void)%s;", old)
		}
```

Reran the exact step 2b command, exit 1. The native and JavaScript backend subtests all still PASS. Sanitized and counted results:
| Pin / dialect | Sanitized | Counted | Counts line |
| --- | --- | --- | --- |
| class.ts | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| class.a | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| object.ts | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| object.a | PASS | PASS | `captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0` |
| by_value.ts | FAIL | FAIL | `captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0` |
| by_value.a | FAIL | FAIL | `captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0` |
| released.ts | FAIL | FAIL | `captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0` |
| released.a | FAIL | FAIL | `captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0` |
| nested.ts | PASS | PASS | `captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0` |
| nested.a | PASS | PASS | `captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0` |
| alternating.ts | FAIL | FAIL | `captured_union_test.go:116: counts: allocations 5, frees 3, in regions 0` |
| alternating.a | FAIL | FAIL | `captured_union_test.go:116: counts: allocations 5, frees 3, in regions 0` |
| graph.ts | PASS | PASS | `captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0` |
| graph.a | PASS | PASS | `captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0` |
| cycle.ts | PASS | PASS | `captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0` |
| cycle.a | PASS | PASS | `captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0` |

Alternating’s first error lines, verbatim for each failing subtest:
`TestCapturedUnionAgrees/alternating.ts/sanitized`:
```text
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
```

`TestCapturedUnionAgrees/alternating.ts/counted`:
```text
    captured_union_test.go:110: counts: heap values leaked: 2 (allocations 5, frees 3, in regions 0)
```

`TestCapturedUnionAgrees/alternating.a/sanitized`:
```text
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
```

`TestCapturedUnionAgrees/alternating.a/counted`:
```text
    captured_union_test.go:110: counts: heap values leaked: 2 (allocations 5, frees 3, in regions 0)
```

For sanitized FAILs, want no leaks and a comparable exit/stdout; got LeakSanitizer exit 1 and the complete diagnostics in the verbatim M3 log below. For counted FAILs, want zero unfreed heap values and zero region survivors; got the exact allocation/free reports below. Native and JavaScript stdout comparisons still passed. No got/want stdout mismatch occurred. Restored the original emitter, then `git diff --exit-code` exited 0 and printed nothing.
## Step 4: count records
Recursively listed files below `internal/oracle/testdata/counts` whose full path contains `captured_union`: **none**. Searched `internal/oracle/*.go` excluding `captured_union_test.go` for `captured_union`: **none**. No record exists for these fixtures, so the conditional `TestCountsAreRecorded -update-counts` command was not run. No count-record diff was produced.
## Step 5: standalone native timings
Built HEAD with the admission gate’s exact `loweringAdapter` from `cmd/adamic-admission-delta/build.go`, temporarily placed at `cmd/adamic/admission_adapter.go`, and the same `go build -trimpath -buildvcs=false` flags. Removed the temporary adapter immediately after build; `git diff --exit-code` passed. No `ADAMIC_GATE_UNCACHED` was set for these measurements; normal gate caches were retained. Compiler build exit 0.
```sh
go build -trimpath -buildvcs=false -o /tmp/lower-n7s-5/adamic ./cmd/adamic
/tmp/lower-n7s-5/adamic build internal/oracle/testdata/captured_union/<pin>.ts -o /tmp/lower-n7s-5/<pin>-<iteration>
```

Each compile had a 120-second limit, each executable a 10-second limit. Measured elapsed time with Python `time.monotonic()` around the child process; timeout handling kills and reaps the whole process group. Each compile/run pair completed before the next began. All compiler stdout/stderr and executable stderr were empty.
| Pin | Iteration | Compile exit | Compile wall seconds | Run exit | Run stdout |
| --- | --- | --- | --- | --- | --- |
| alternating | 1 | 0 | 8.905058617 | 0 | `"first\nstopping\nundefined\nlast\nstopping\n"` |
| alternating | 2 | 0 | 2.079574204 | 0 | `"first\nstopping\nundefined\nlast\nstopping\n"` |
| alternating | 3 | 0 | 2.482429954 | 0 | `"first\nstopping\nundefined\nlast\nstopping\n"` |
| graph | 1 | 0 | 9.536197783 | 0 | `"held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"` |
| graph | 2 | 0 | 2.295074778 | 0 | `"held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"` |
| graph | 3 | 0 | 2.388714484 | 0 | `"held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"` |
| class | 1 | 0 | 2.304769609 | 0 | `"start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"` |

Stdout is JSON-quoted above to preserve every newline; these are the complete bytes (all ASCII), with no truncation.
## Step 6: admission delta
```sh
python3 cloud/admission-corpus/manifest.py --sha 02c50297a07758a29bc1f1201477bd275ab30278 > /tmp/lower-n7s-5/manifest.json
go build -o /tmp/lower-n7s-5/admission-delta ./cmd/adamic-admission-delta
/tmp/lower-n7s-5/admission-delta --base 730e7c26e590af0c57c545ba0cb905b09e4bee54 --head 02c50297a07758a29bc1f1201477bd275ab30278 --manifest /tmp/lower-n7s-5/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json
```

Ran alone with no competing heavy command. Direct process exit **0**, verdict **pass**, `void=false`, void reason absent, all program void reasons absent. `diff_count=8`; admitted 8; newly refused 0. Complete corpus, no filter, no shard, no budget reduction, omitted 0. 5,309 programs: 8 newly accepted, 3,319 accepted by both, 1,982 refused by both. Classification is lowering-only; runtime checks compare source Node, JavaScript backend and native executable for selected programs. Gate limits: compile 45 seconds, run 10 seconds.
Manifest corpora: witnesses 2, fixtures 4361, gaps 105, review 20, fuzz 0, stage3 821. Generator revision is HEAD; generator blob `09d061a03b1e016387e975d25ee0cd5b75ac6c1f`; manifest blob `45433c2007190a896569b00e318a18e467d3b5dd`. Gate total wall seconds: 1135.166491604.
Every program with differing classification stderr follows. Each base text is complete and verbatim; each HEAD stderr is exactly the empty string (`""`). These are also exactly the eight changed programs and eight newly admitted programs; each has `agree=true`, native compile exit 0 and native run exit 0.
### internal/oracle/testdata/captured_union/alternating.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/alternating.ts:15:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### internal/oracle/testdata/captured_union/by_value.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### internal/oracle/testdata/captured_union/class.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### internal/oracle/testdata/captured_union/cycle.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### internal/oracle/testdata/captured_union/graph.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### internal/oracle/testdata/captured_union/nested.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### internal/oracle/testdata/captured_union/object.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### internal/oracle/testdata/captured_union/released.ts
Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-3459839018/head/internal/oracle/testdata/captured_union/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr: `""` (zero bytes).
### Gate stderr, verbatim
```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 8
```

### Timeouts and reruns
No timeout occurred: `timed_out_runs=0`, `retried_runs=0`, `still_timed_out_runs=0`; no classification or backend observation has `error="timeout"`, and no program has a timeout class or void reason. Therefore there are no timed-out programs to rerun at either SHA, at the gate’s limit or at 120 seconds. A fixture named `stop_host_timeout.a` is a ruled runtime-stop fixture, not a timeout observation.
### Initial invocation correction
An initial invocation without `--manifest`/`--corpus` exited before classification (go-run wrapper exit 1). It is not the admission result above. Corrected the omission using the repository README’s full manifest generator and reran the gate directly. Initial stderr, verbatim:
```text
negative witnesses: 50
--manifest or --corpus is required
exit status 1
```

## Expected stdout pins
Exact expected stdout for each pin, shared by `.ts` and `.a`, JSON-quoted with all newlines preserved:
`class`:
```json
"start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
```

`object`:
```json
"start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
```

`by_value`:
```json
"step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"
```

`released`:
```json
"first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"
```

`nested`:
```json
"no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"
```

`alternating`:
```json
"first\nstopping\nundefined\nlast\nstopping\n"
```

`graph`:
```json
"held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"
```

`cycle`:
```json
"undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n"
```

## Complete lowering test output (verbatim)
```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== CONT  TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
--- PASS: TestCapturedMaybeBooleanStillStops (0.16s)
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
--- PASS: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (0.17s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (0.21s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.a (0.21s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (0.25s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.20s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.16s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (0.17s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.ts (0.22s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (0.14s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.08s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.11s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (0.11s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.13s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.13s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.06s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.06s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.571s
```

## Complete HEAD oracle test output (verbatim)
```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== CONT  TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
=== NAME  TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 245.64
=== NAME  TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 160.49
=== NAME  TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 245.64
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 miss 245.64
=== NAME  TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 159.76
=== CONT  TestCapturedUnionAgrees/alternating.a
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/cycle.a
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:116: counts: allocations 5, frees 5, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:116: counts: allocations 5, frees 5, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== CONT  TestCapturedUnionAgrees/object.a
=== CONT  TestCapturedUnionAgrees/released.ts
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== CONT  TestCapturedUnionAgrees/nested.a
=== CONT  TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.04
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== NAME  TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
=== NAME  TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== NAME  TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
--- PASS: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/class.ts (289.36s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (0.43s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (0.62s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.41s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (246.08s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (289.36s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (19.96s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (32.03s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (34.87s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (160.68s)
    --- PASS: TestCapturedUnionAgrees/graph.a (289.36s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (20.00s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (32.00s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (35.61s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (159.93s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (289.37s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.33s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (0.66s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (0.34s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (246.23s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (289.38s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.31s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (0.63s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.27s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (246.42s)
    --- PASS: TestCapturedUnionAgrees/alternating.ts (2.05s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.56s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/sanitized (0.72s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/counted (0.28s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.15s)
    --- PASS: TestCapturedUnionAgrees/alternating.a (2.11s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.41s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/sanitized (0.70s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/counted (0.33s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.15s)
    --- PASS: TestCapturedUnionAgrees/by_value.ts (2.11s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.59s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/sanitized (0.67s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/counted (0.33s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/released.a (2.29s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.43s)
        --- PASS: TestCapturedUnionAgrees/released.a/sanitized (0.59s)
        --- PASS: TestCapturedUnionAgrees/released.a/counted (0.50s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (2.34s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (0.41s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (0.77s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (0.42s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/released.ts (2.19s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.52s)
        --- PASS: TestCapturedUnionAgrees/released.ts/sanitized (0.77s)
        --- PASS: TestCapturedUnionAgrees/released.ts/counted (0.44s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.15s)
    --- PASS: TestCapturedUnionAgrees/object.a (2.29s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.73s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (0.73s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.37s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/class.a (2.31s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.70s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (0.60s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.59s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.10s)
    --- PASS: TestCapturedUnionAgrees/nested.a (2.12s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.53s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.76s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.41s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/object.ts (2.30s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.51s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (0.90s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.35s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.07s)
    --- PASS: TestCapturedUnionAgrees/by_value.a (1.29s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.22s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/sanitized (0.52s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/counted (0.24s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.08s)
PASS
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	294.957s
```

## Complete M3 oracle test output (verbatim)
```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/cycle.a
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== NAME  TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 3.08
=== NAME  TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 3.09
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 3.09
=== NAME  TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 3.14
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/released.ts
=== CONT  TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== NAME  TestCapturedUnionAgrees/by_value.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10682==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x563b80c7a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563b80ce6d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563b80ce6d2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x563b80c7a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563b80ce6d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563b80ce6d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x563b80cbe02d in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:136:2
            #4 0x563b80cbe02d in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x563b80cbda6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x563b80c7a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563b80ce6d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563b80ce6d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x563b80cbddef in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:91:6
            #4 0x563b80cbddef in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x563b80cbda6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== NAME  TestCapturedUnionAgrees/released.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10684==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5622bdb903a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5622bdbfd3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5622bdbfd3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5622bdbd48cc in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:197:85
            #4 0x5622bdbd3d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x5622bdbd3d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x5622bdbd3a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5622bdb903a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5622bdbfd3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5622bdbfd3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5622bdbd46a6 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:178:84
            #4 0x5622bdbd3d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x5622bdbd3d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x5622bdbd3a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5622bdb903a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5622bdbfd3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5622bdbfd3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5622bdbd4431 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:145:80
            #4 0x5622bdbd3d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x5622bdbd3d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x5622bdbd3a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x5622bdb903a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5622bdbfd3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5622bdbfd3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x5622bdb903a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5622bdbfd3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5622bdbfd3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x5622bdb903a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5622bdbd8daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== NAME  TestCapturedUnionAgrees/released.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10697==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55b84a6bd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b84a72a3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b84a72a3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b84a7018cc in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:197:85
            #4 0x55b84a700d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55b84a700d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55b84a700a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55b84a6bd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b84a72a3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b84a72a3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b84a7016a6 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:178:84
            #4 0x55b84a700d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55b84a700d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55b84a700a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55b84a6bd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b84a72a3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b84a72a3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b84a701431 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:145:80
            #4 0x55b84a700d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55b84a700d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55b84a700a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x55b84a6bd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b84a72a3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b84a72a3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x55b84a6bd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b84a72a3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b84a72a3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x55b84a6bd3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b84a705daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10696==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x55827d55f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55827d5cbd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55827d5cbd2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55827d55f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55827d5cbd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55827d5cbd2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55827d5a302d in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:136:2
            #4 0x55827d5a302d in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x55827d5a2a6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55827d55f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55827d5cbd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55827d5cbd2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55827d5a2def in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:91:6
            #4 0x55827d5a2def in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x55827d5a2a6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
=== NAME  TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.03
=== NAME  TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
=== NAME  TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== NAME  TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== CONT  TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== CONT  TestCapturedUnionAgrees/alternating.a
=== CONT  TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== NAME  TestCapturedUnionAgrees/alternating.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10854==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x557abf5963a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x557abf6023ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x557abf6023ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x557abf5da24e in adamic_function_alternating_a_alternate_write_anonymous_1_5b340894eade039d56c6d1fecde8293f /adamic-build/main.c:129:82
            #4 0x557abf5d9cea in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:61:6
            #5 0x557abf5d9cea in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:286:2
            #6 0x557abf5d9a6d in adamic_initialize_alternating_a_7713bd73cbbc6d9c01b7c1c17db95720 /adamic-build/main.c:296:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x557abf5963a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x557abf6023ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x557abf6023ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x557abf5da085 in adamic_function_alternating_a_alternate_write_anonymous_1_5b340894eade039d56c6d1fecde8293f /adamic-build/main.c:107:80
            #4 0x557abf5d9cea in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:61:6
            #5 0x557abf5d9cea in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:286:2
            #6 0x557abf5d9a6d in adamic_initialize_alternating_a_7713bd73cbbc6d9c01b7c1c17db95720 /adamic-build/main.c:296:2
        
        SUMMARY: AddressSanitizer: 132 byte(s) leaked in 2 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10867==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x562749bc83a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x562749c343ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x562749c343ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x562749c0c24e in adamic_function_alternating_ts_alternate_write_anonymous_1_363a6e705fb06d468cc8e54876cc4706 /adamic-build/main.c:129:82
            #4 0x562749c0bcea in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:61:6
            #5 0x562749c0bcea in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:286:2
            #6 0x562749c0ba6d in adamic_initialize_alternating_ts_41e85a81e20c992ebefc194ce4bfaebe /adamic-build/main.c:296:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x562749bc83a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x562749c343ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x562749c343ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x562749c0c085 in adamic_function_alternating_ts_alternate_write_anonymous_1_363a6e705fb06d468cc8e54876cc4706 /adamic-build/main.c:107:80
            #4 0x562749c0bcea in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:61:6
            #5 0x562749c0bcea in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:286:2
            #6 0x562749c0ba6d in adamic_initialize_alternating_ts_41e85a81e20c992ebefc194ce4bfaebe /adamic-build/main.c:296:2
        
        SUMMARY: AddressSanitizer: 132 byte(s) leaked in 2 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:110: counts: heap values leaked: 2 (allocations 5, frees 3, in regions 0)
    captured_union_test.go:116: counts: allocations 5, frees 3, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== CONT  TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 2 (allocations 5, frees 3, in regions 0)
    captured_union_test.go:116: counts: allocations 5, frees 3, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== CONT  TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== CONT  TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
=== NAME  TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== NAME  TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.02
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 16d468e511b6 hit 0.01
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (11.23s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.36s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (0.74s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.26s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (3.42s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (11.29s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.31s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (0.66s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (0.34s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (3.58s)
    --- PASS: TestCapturedUnionAgrees/class.ts (11.31s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (0.48s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (0.61s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.48s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (3.23s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (11.38s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (0.37s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (0.63s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (0.40s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (3.46s)
    --- FAIL: TestCapturedUnionAgrees/released.a (1.86s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.29s)
        --- FAIL: TestCapturedUnionAgrees/released.a/sanitized (0.77s)
        --- FAIL: TestCapturedUnionAgrees/released.a/counted (0.42s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (0.11s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (1.88s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.38s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/sanitized (0.67s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/counted (0.36s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.11s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (1.96s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.38s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/sanitized (0.71s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/counted (0.44s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (0.14s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (1.82s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.35s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/sanitized (0.74s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/counted (0.33s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.11s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (1.38s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.29s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/sanitized (0.44s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/counted (0.24s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.12s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (1.59s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.29s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/sanitized (0.57s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/counted (0.32s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.08s)
    --- PASS: TestCapturedUnionAgrees/nested.a (1.76s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.29s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.67s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.32s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.10s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (15.96s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (3.18s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (3.26s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (2.93s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/graph.a (2.84s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (0.41s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (0.77s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (1.22s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/object.ts (1.69s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.35s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (0.58s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.41s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.08s)
    --- PASS: TestCapturedUnionAgrees/object.a (1.75s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.42s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (0.66s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.34s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.07s)
    --- PASS: TestCapturedUnionAgrees/class.a (1.68s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.49s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (0.59s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.33s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.06s)
FAIL
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	16.697s
FAIL
```

## Repository disposition
The proof was executed from detached HEAD. Compiler sources were restored before the standalone measurements and admission gate; no count records were changed. Only this report is to be committed on `proof/lower-n7s-5-report`, based on the requested HEAD, and only that branch is to be pushed. Raw logs, toolchains, packages and executables remain outside the repository under `/tmp/lower-n7s-5` (session-local).

Added required a-check headers to the callable-producer/view-snapshot change; all 63 changed .a files pass.
Commit: the single commit on compiler/loom-r-views-fx-c-ac, based exactly on dbdf977f5b8b0979d414ebe6f20e7aadd312e309.
Checks: real pinned Gate.aCheck, affected lower tests, reader guard and vet pass; all bodies and counts preserved.
Mutants: removed-header and wrong-reason are both rejected by the unchanged pinned predicate.
Not covered: no full package suite, full gate, ahra or Loom submission; prior production mutant evidence is retained, not rerun.

## Scope and measured status

Fetched origin, created compiler/loom-r-views-fx-c-ac from the exact requested base and
applied 6c7cdd296675520447553e0b604cbfb647c67834 with cherry-pick --no-commit. No newer main
work was incorporated. The compiler production files, runtime lines, tests and counts
are unchanged from that source commit. The underlying rebase report remains under
review/compiler/loom-r-views-fx-c.

Scanned every added/modified .a path in that change, including review evidence: 63 files,
54 checked, 5 NotYet, 4 refused, 0 type errors. Gate groups checked and NotYet together
as 59 checked. baseline.json records real compiler diagnostics before headers; final.json
records a second real compiler scan after headers. statuses.md is the complete per-file
table, including all actual NotYet reasons. No file was excluded.

All four refused files receive the same exact first line:

```text
// a-check: refused a checked callable producer assignable to more than one union member
```

- internal/lower/testdata/callable_producers/p135_object_result.a
- internal/lower/testdata/callable_producers/p22.a
- review/compiler/loom-views-fx-c/baseline-sources/0587b8e6f528/fxspptb_oct9_views_p22_callable_union_fewer_parameters.a
- review/compiler/loom-views-fx-c/baseline-sources/4123e0eab5d3/p135_object_result.a

No clean or NotYet file receives a header; no checker-error header was needed. All 63
bodies are byte-identical to the source commit after omitting only the four inserted
first lines. headers.json records each header and the unchanged body SHA-256.

## Position audit

Searched changed full paths in *_test.go and testdata, then the dynamic fixture names
p22 and p135_object_result in the lower tests. No exact line-number expectation references
these paths. The two affected lower tests use callableProducerSource and assert a located
main.a diagnostic plus the arrow-wrapper fix, without a fixed line. They pass unchanged.
No test or live expectation needed updating.

line-offsets.json records the new observations: p22 and its review copy move from 7:18
to 8:18; p135_object_result and its copy move from 6:18 to 7:18. Earlier baseline/branch
result JSON files remain historical observations from their original runs; they are not
live diagnostic expectations. Original program bodies and their historical body hashes
remain valid. No historical measurement was rewritten as a new run.

## Exact predicate and mutants

Read stage3/a-check-headers/README.md whole and used its REPORT.md reproduction. Fetched
fbac28c62493f27a788edc02a18bc8edb68de5da, then extracted cloud/fast-gate/run.py to gate.py.
scan.py imports that unchanged module and invokes Gate.aCheck with its real go-build step
and real compiler subprocesses on all 63 paths. Each build has a 90s limit; each c
invocation has a 45s limit. The baseline gate finds exactly the four missing headers;
the final gate exits 0. Stdout from successful compiler c runs is discarded after its
size is recorded; exact diagnostics are retained.

prove-headers.py uses the README's predicate proof: unchanged Gate.aCheck, captured real
compiler diagnostics, only the first line altered in the real review witness. It tests
0587b8e6f528/fxspptb_oct9_views_p22_callable_union_fewer_parameters.a. Removed header fails
with refused/expected checked; wrong reason fails with refused/expected refused deliberately
wrong reason. Both exit 1; finally restores the original header and body. header-mutants.json
and header-mutants.log record both catches. This is explicitly a predicate replay; both
ordinary scans separately invoked the real compiler on every path.

## Commands and results

All commands source /workspace/adamic-tools/env.sh; GOPROXY=https://proxy.golang.org|direct.
All test output went directly to logs; no whole suite ran.

- timeout 180 bash cloud/setup.sh: exit 0, done 16.708s; nproc=5, four-CPU cgroup.
- timeout 90 go build -o /tmp/adamic-views-ac ./cmd/adamic: exit 0 (build.log).
- timeout 240 python3 review/compiler/loom-r-views-fx-c-ac/scan.py baseline: runner exit 0,
  expected Gate.aCheck exit 1 for the four missing headers (scan-baseline.log).
- timeout 240 python3 review/compiler/loom-r-views-fx-c-ac/scan.py final: exit 0;
  Gate.aCheck exit 0, all 63 pass (scan-final.log).
- timeout 30 python3 review/compiler/loom-r-views-fx-c-ac/prove-headers.py: exit 0,
  both mutants caught (header-mutants.log).
- timeout 90 go test ./internal/lower -run '^TestCallableProducer' -timeout 90s
  -count=1 -v -parallel 4: exit 0, 0.750s (lower.log). Every test reading the modified
  lower fixture paths is included. No Go test source was changed for these headers.
- timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s
  -count=1: exit 0, 1.849s (reader-guard.log).
- timeout 90 go vet ./internal/lower ./internal/native ./internal/oracle: exit 0 (vet.log).

No new test leaf. Rerun lower test seconds are in test-seconds.json; all under 60s.
Prior source-commit evidence covers the production change's full focused fixtures,
reader compilation, flow corpus, native checks and eleven corrected production mutants.
Those were not rerun for comment-only changes to four refused sources. No admitted
fixture changed, and internal/oracle/counts.md is byte-identical to 6c7cdd29. Counts were
not regenerated. No new runtime feature or compiler-consuming package.

## Setup timings

```text
submodule cache: off
v24.19.0
go version go1.27.1 linux/amd64
setup: node ready (0.030s)
setup: go ready (0.031s)
setup: submodules ready (0.060s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.016s
setup: markdown dependencies ready (0.102s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.244s)
setup: shared cache ready (0.668s)
setup: go build ready (16.581s)
setup: test binaries deferred (use --warm-tests) (16.684s)
setup: build cache warm (16.685s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/138 706 load-after=1.80 0.38 0.12 1/142 1402
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (16.708s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.6Cwns8
```

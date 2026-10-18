Retained the large-array lowering change and added its one required a-check refusal header.
One replacement commit on dbdf977f5b8b0979d414ebe6f20e7aadd312e309; original candidate 0f15b9f696c939dd82fd0a127de31aa91660025e.
Six real compiler classifications and the pinned Gate.aCheck pass; eight lowering tests, reader guard and vet pass.
Removed-header and wrong-reason mutants each fail the unchanged Gate.aCheck predicate.
No whole gate, ahra, Loom submission, or new backend rerun; prior backend and nine lowering-mutant evidence is retained unchanged.

## Scope and assumptions

Fetched origin, created compiler/loom-r-implied-target-ac at the explicitly requested dbdf977f5b base, and applied only 0f15b9f's own change using cherry-pick --no-commit. No conflicts. Origin/main had advanced to 99c369e0; the specific pinned-base instruction takes precedence over the generic current-main rule. The replacement is one commit, not a descendant of the previous candidate.

Read CLAUDE.md, docs/front-door.md and stage3/a-check-headers/README.md. The README's scan harness uses Gate.aCheck from fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py. That object was fetched from origin devtools/fast-gate. pinned-gate.py preserves its exact source. check-headers.py invokes the unchanged method, building this branch's compiler, invoking adamic c on each changed .a in place, and bounding every compiler subprocess to 55 seconds. No gate classification or predicate was replaced by an approximation.

All program bodies match their original 0f15b9f bytes (SHA-256 in statuses.json). Production Go and tests match 0f15b9f exactly. Runtime C changes: none. Counts changed rows: zero, byte-identical to both the pinned main and 0f15b9f. No counts regeneration was needed for this header-only followup. scope-proof.json records these checks. No cohere source was copied.

## Every changed .a program

All paths below are relative to review/compiler/loom-implied-target/. Every .a in the source commit's own diff is included, including review evidence. Seconds are the final in-place predicate's real compiler invocation.

| File | Status | Seconds | Header added |
| --- | --- | ---: | --- |
| followup/objects-1024.a | checked | 0.128 | none |
| followup/objects-2048.a | checked | 0.235 | none |
| followup/objects-4096.a | checked | 0.436 | none |
| followup/objects-512.a | checked | 0.086 | none |
| objects-sequential.a | checked | 0.231 | none |
| objects.a | refused | 0.058 | exact reason below |

Totals: 5 checked, 0 NotYet, 1 refused, 0 type errors. The sole added first line is:

```text
// a-check: refused parallelMap items are not shareable: items is a mutable { value: number; done: boolean; }[]; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
```

The refusal now points at objects.a:2054:33 rather than 2053:33. The entire tracked *_test.go and testdata tree was searched for the review directory and distinct fixture basenames: zero matches (path-reader-search.json). No test expectation changed. Earlier reports/logs, original-admission.json and source provenance hashes describe actual pre-header observations and remain historical evidence; they are not current position expectations. statuses.json records the new diagnostic and the unchanged below-header provenance hash 848de67f6a6711fe1da741bfd93b2732248f17bbc27c396a6faa7e98206dbd6d.

## Predicate and mutants

`timeout 240 python3 review/compiler/loom-r-implied-target-ac/check-headers.py` after sourcing setup env: exit 0 (scan.log). Six pre-header compiler observations determine the exact header; then the actual pinned Gate.aCheck accepts all six in 2.8 seconds (a-check.json). The five clean programs have compiler exit 0, empty stderr and emitted C. objects.a exits 1, empty stdout, the exact refusal above.

Both experiments mutate objects.a in place and invoke the unchanged predicate with real compiler diagnostics. Each is restored in finally, and below-header hashes are checked afterward.

| Mutant | Catcher | Result |
| --- | --- | --- |
| Remove the first-line header | Gate.aCheck: refused, expected checked | exit 1, caught |
| Replace reason with deliberate-wrong-reason-header-mutant | Gate.aCheck: refused, expected wrong reason | exit 1, caught |

header-mutants.json and each named mutant JSON preserve exact failures. These are header-predicate mutants, not new claims about lowering semantics. Previous nine production-mutant catches and Node/JavaScript/native/ASan/UBSan comparisons remain in review/compiler/loom-r-implied-target/ unchanged; this pass did not repeat them. The only changed input is still refused before either backend.

## Focused reruns

All output was redirected to files, never piped. Commands ran with the setup env, ADAMIC_GOCACHE_OFF=1, GOMAXPROCS=4 and GOFLAGS preserving setup's flags plus -p=4.

- `timeout 90 go build -o /tmp/loom-r-implied-target-ac-adamic ./cmd/adamic`: exit 0, build.log.
- `timeout 150 go test ./internal/lower -run '^Test(ImpliedTarget|CycleScalar)' -count=1 -json -timeout 90s`: exit 0, lower.jsonl, package 4.657s.
- `timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s`: exit 0, reader-guard.jsonl, package 0.895s.
- `timeout 90 go vet ./internal/lower`: exit 0, empty vet.log.
- `git diff --check` and `git diff --cached --check`: exit 0.

| Test | Seconds |
| --- | ---: |
| TestImpliedTargetLargeArrayQueriesOnce | 0.56 |
| TestImpliedTargetTupleCachesNil | 0.03 |
| TestCycleScalarLiteralShapesOnce | 0.61 |
| TestImpliedTargetOrQueriesOnce | 0.72 |
| TestCycleScalarShapeKeysPreserveDistinctFields | 0.08 |
| TestImpliedTargetAndQueriesOnce | 0.85 |
| TestImpliedTargetNullishQueriesOnce | 1.37 |
| TestImpliedTargetConditionalQueriesOnce | 4.64 |
| TestCallTargetReaders | 0.88 |

No Go tests changed from the prior candidate. The eight regression leaves include their internal fixture setup and are below 60 seconds on the four-CPU quota. No path-reading test depends on the header; the six-file real compiler scan rechecks the affected input itself. The prior candidate's focused 52-package reader validation is retained; no production Go bytes changed in this followup, and no whole-package suite was rerun.

## Setup

`export GOPROXY='https://proxy.golang.org|direct' ADAMIC_GOCACHE_OFF=1; timeout 240 bash cloud/setup.sh > review/compiler/loom-r-implied-target-ac/setup.log 2>&1` passed. Sourced /workspace/adamic-tools/env.sh. nproc=5, cpu.max=400000 100000 (four CPU quota), Go 1.27.1, Node 24.19.0, house clang 20.1.8. No workaround needed.

Timing lines: submodules 0.068s; Go 0.071s; Node 0.073s; markdown skipped validated dependencies (step 0.023s), ready 0.197s; clang 0.417s; shared cache off 0.419s; Go build 9.788s; test binaries deferred 9.965s; cache warm 9.966s; done 9.991s. Full setup output is in setup.log.

Built: a-check header pass for all 22 Adamic programs carried by compiler slice 3, preserving every program body byte.
Commits: fixed base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 0a8d7ed8f798ff96915bb24892550dd1d77dfab8; delivery SHA is in the final response and git log.
Commands and outputs: branch-own adamic c reports 16 checked, zero NotYet, six refused, zero type errors; the unchanged pinned a-check predicate accepts all 22 after the header edits.
Mutants: removed-header and wrong-reason on function_call.a both remain genuinely refused and each makes the unchanged predicate exit 1.
Not covered: a whole-tree scan, full packages/gate, unrelated production mutants, macOS/WASI and real checker ABI execution; original slice evidence remains under review/compiler/loom-r-slice-3.

## Branch and scope

Fetched origin, created compiler/loom-r-slice-3-ac from exactly dbdf977f5b, and applied git cherry-pick --no-commit 0a8d7ed8f798ff96915bb24892550dd1d77dfab8. The delivery is one commit on that fixed base, not an extra commit on the earlier branch. origin/main advanced during fetch; the requested base stays fixed.

The changed-program list is git diff --name-only dbdf977f5b 0a8d7ed8 -- '*.a'. It contains 22 paths, including every location in the tree: there are no .a additions or modifications under review/. paths.txt records the exact list. New mutant source captures use .a.txt, not executable .a files. No new program was introduced by this header pass.

Read CLAUDE.md, docs/front-door.md, and stage3/a-check-headers/README.md whole. The language and memory documents had been read during the original slice work in this conversation. The README refers to Gate.aCheck at fbac28c62493f27a788edc02a18bc8edb68de5da. Its object was initially unavailable locally; fetching origin devtools/fast-gate made it available. gate.py is byte-identical to that exact pinned cloud/fast-gate/run.py, verified in predicate-source.json.

scan.py invokes that unchanged Gate.aCheck. Its adapter redirects only the compiler build output into the local evidence tools directory, supplies hard subprocess limits, and captures each real compiler observation. It does not change classification or header matching. Every row is compiled in place with this branch's own cmd/adamic via adamic c, exactly the frontend/emission operation a-check uses. Initial scan: four predicate findings among the six refusals. Final scan: exit 0, all 22 accepted.

## Headers and body proof

Conservative reading of "clean programs take no header": remove the six redundant a-check pass comments, rather than carrying them forward. Replace all six refusal headers with the complete measured refusal reason and fix, including the two shorter headers that already matched as substrings. No type-error header is needed. Updated the assignment bucket README's claim that every file begins with a pass header.

No header was prepended to a previously headerless refused program: six existing refusal headers were corrected/canonicalized. Ten other .a files are byte-identical to 0a8d7ed8. For all 22 files, stripping only an existing first-line a-check comment gives a byte-identical body before and after (body-verification.json). Headers were restored after both mutants and this proof was rerun.

| File | New first-line header / action | Line shift |
|---|---|---:|
| internal/oracle/testdata/refusal_rulings/function_call.a | // a-check: refused the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number | 0 |
| internal/oracle/testdata/refusal_rulings/function_length.a | // a-check: refused the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number | 0 |
| internal/oracle/testdata/refusal_rulings/function_unused.a | // a-check: refused the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number | 0 |
| internal/oracle/testdata/refusal_rulings/merge_missing.a | // a-check: refused class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view | 0 |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a | // a-check: refused class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view | 0 |
| internal/oracle/testdata/refusal_rulings/merge_unused.a | // a-check: refused class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view | 0 |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | Removed redundant pass header | -1 |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | Removed redundant pass header | -1 |
| stage3/fixtures/assignment-proofs/03_return_defined.a | Removed redundant pass header | -1 |
| stage3/fixtures/assignment-proofs/04_chained.a | Removed redundant pass header | -1 |
| stage3/fixtures/assignment-proofs/05_compound.a | Removed redundant pass header | -1 |
| stage3/fixtures/assignment-proofs/06_wider_target.a | Removed redundant pass header | -1 |

Position audit: position-references.txt records the *_test.go/testdata path search, and report-position-references.txt records the search for assignment filenames followed by pinned local line numbers in reports and metadata. No expectation needed updating. The six refusal replacements occupy the same one line, so the pinned 2:22, 2:23, 2:10, 3:16 and merged-field 2:1 diagnostics remain valid. The six assignment witnesses shift up one line, but no test or report pins their local diagnostic positions. Their upstream scanner.ts:1815-1828 provenance is unchanged. No _test.go expectation or program body was edited.

## Per-file compiler status

Counts: checked 16; NotYet 0; refused 6; type error 0; unclassified 0. Refused rows are expected and all are accepted by the header predicate. after.json retains full diagnostics, exit codes, first-line expectations, emitted-C hashes and per-file seconds.

| File | Status | Reason/code | Seconds |
|---|---|---|---:|
| internal/oracle/testdata/call_spread_array.a | checked |  | 0.042442 |
| internal/oracle/testdata/json_shorthand.a | checked |  | 0.046122 |
| internal/oracle/testdata/optional_after_call_193.a | checked |  | 0.034471 |
| internal/oracle/testdata/optional_after_call_79.a | checked |  | 0.037359 |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | checked |  | 0.03442 |
| internal/oracle/testdata/optional_after_call_propagation.a | checked |  | 0.0312 |
| internal/oracle/testdata/optional_after_call_required.a | checked |  | 0.029909 |
| internal/oracle/testdata/optional_after_call_variants.a | checked |  | 0.036387 |
| internal/oracle/testdata/refusal_rulings/function_call.a | refused | the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number | 0.030651 |
| internal/oracle/testdata/refusal_rulings/function_length.a | refused | the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number | 0.026298 |
| internal/oracle/testdata/refusal_rulings/function_signature.a | checked |  | 0.038459 |
| internal/oracle/testdata/refusal_rulings/function_unused.a | refused | the Function annotation in .a; write a call signature with its parameters and result, like (value: number) => number | 0.039912 |
| internal/oracle/testdata/refusal_rulings/merge_missing.a | refused | class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view | 0.037909 |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a | refused | class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view | 0.042715 |
| internal/oracle/testdata/refusal_rulings/merge_present.a | checked |  | 0.045567 |
| internal/oracle/testdata/refusal_rulings/merge_unused.a | refused | class Missing merged with an interface promises field promised that the class does not initialize; declare and initialize promised in class Missing, or use a separate interface with a checked view | 0.035338 |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | checked |  | 0.056502 |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | checked |  | 0.06496 |
| stage3/fixtures/assignment-proofs/03_return_defined.a | checked |  | 0.028647 |
| stage3/fixtures/assignment-proofs/04_chained.a | checked |  | 0.032357 |
| stage3/fixtures/assignment-proofs/05_compound.a | checked |  | 0.040231 |
| stage3/fixtures/assignment-proofs/06_wider_target.a | checked |  | 0.041891 |

## Header mutants

Both experiments use internal/oracle/testdata/refusal_rulings/function_call.a and the real branch compiler, not a fabricated or replayed diagnostic. Each rebuilds/reuses the branch compiler through the unchanged pinned predicate. Program bodies stay identical. A finally block restores the correct header. Full results are in removed-header.json, wrong-reason.json and mutants.json.

| Mutant | Compiler observation | Predicate result / catcher |
|---|---|---|
| Remove first-line header | refused, reason unchanged; location moves to line 1 | exit 1: expected checked, observed refused |
| Replace with refused intentionally wrong reason 999999 | refused, reason unchanged | exit 1: expected reason is absent from actual diagnostic |

Exact reproduction commands after sourcing the setup environment:

```sh
timeout 180 python3 review/compiler/loom-r-slice-3-ac/scan.py after
timeout 180 python3 review/compiler/loom-r-slice-3-ac/scan.py mutants
```

## Tests, reader guard, vet and counts

Header-affected checks pass: Function and merged-field oracle tests, all six assignment goldens and six assignment source mutants, the assignment fixture directory, and the fixture directory registry. These include source Node, JavaScript backend, native release, ASan/UBSan and leak checks where applicable. The tests pin actual diagnostic positions as well as meanings. TestCallTargetReaders passes with -timeout=90s. Vet passes for all 14 Go packages carried by the original compiler slice.

All test runs are focused, not whole packages. Output goes directly to .log files. The runner sets GOMAXPROCS=4, four-CPU affinity and ADAMIC_GATE_UNCACHED=1. No test was over 60 seconds; no test was split, removed or rerun for exceeding its budget. Exact commands and wall times follow; test-seconds.json gives every test/subtest time.

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
| function | 0 | 7.216 | `go test ./internal/oracle -run ^TestFunction(Adamic\|TypeScript\|CallSignature) -count=1 -json -timeout=90s -parallel=4` |
| merged | 0 | 3.851 | `go test ./internal/oracle -run ^TestMergedField -count=1 -json -timeout=90s -parallel=4` |
| assignment | 0 | 7.021 | `go test ./internal/oracle -run ^TestAssignmentProof -count=1 -json -timeout=90s -parallel=4` |
| fixtures | 0 | 5.258 | `go test ./stage3/fixtures -run ^Test(FixturesAssignmentProofs\|FixtureDirectoriesHaveTopLevelTests)$ -count=1 -json -timeout=90s -parallel=4` |
| readers | 0 | 1.931 | `go test ./internal/ir -run ^TestCallTargetReaders -count=1 -json -timeout=90s -parallel=4` |
| vet | 0 | 0.266 | `go vet ./internal/lower ./internal/native ./internal/ir ./internal/flow ./internal/javascript ./internal/oracle ./stage3/fixtures ./stage1/cohere/cssstrings ./stage1/cohere/json ./stage1/cohere/markdownblocks ./stage1/cohere/selector ./stage1/cohere/yaml ./stage1/typescript/parser ./stage1/typescript/scanner` |

Counts: internal/oracle/counts.md is byte-identical to 0a8d7ed8. No regeneration was run. An isolated Go test overlay measures all 14 registered ruling/assignment rows directly and compares every numeric field to the existing table; all pass (package 2.691 seconds). The overlay is a .go.txt evidence file, not a new Go package or permanent test.

```sh
GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 timeout 180 go test ./internal/oracle -overlay=review/compiler/loom-r-slice-3-ac/counts-overlay/overlay.json -run '^TestHeaderCountsStayRecorded$' -count=1 -json -timeout=90s -parallel=4
```

Every recorded top-level test time:

| Check | Test | Seconds |
|---|---|---:|
| fixtures | TestFixtureDirectoriesHaveTopLevelTests | 0 |
| fixtures | TestFixturesAssignmentProofs | 1.84 |
| merged | TestMergedFieldAdamicUnusedRefused | 2.01 |
| merged | TestMergedFieldTypeScriptUnusedAccepted | 2.25 |
| merged | TestMergedFieldAdamicReadRefused | 0.11 |
| merged | TestMergedFieldAdamicInitializedAgreesWithNode | 0.36 |
| merged | TestMergedFieldTypeScriptInitializedAgreesWithNode | 2.41 |
| merged | TestMergedFieldTypeScriptComputedChecked | 2.44 |
| merged | TestMergedFieldAdamicComputedRefused | 0.1 |
| merged | TestMergedFieldTypeScriptMissingChecked | 0.28 |
| assignment | TestAssignmentProofReturnDefinedMutant | 3.23 |
| assignment | TestAssignmentProofScannerKeyword | 3.3 |
| assignment | TestAssignmentProofChained | 3.34 |
| assignment | TestAssignmentProofIfNarrowingMutant | 3.39 |
| assignment | TestAssignmentProofWiderTargetMutant | 1.03 |
| assignment | TestAssignmentProofReturnDefined | 1.18 |
| assignment | TestAssignmentProofCompoundMutant | 1.14 |
| assignment | TestAssignmentProofWiderTarget | 1.15 |
| assignment | TestAssignmentProofCompound | 0.97 |
| assignment | TestAssignmentProofChainedMutant | 0.96 |
| assignment | TestAssignmentProofIfNarrowing | 1.01 |
| assignment | TestAssignmentProofScannerKeywordMutant | 1.05 |
| readers | TestCallTargetReaders | 0.77 |
| function | TestFunctionTypeScriptCallRefused | 2.49 |
| function | TestFunctionAdamicLengthAnnotationRefused | 2.49 |
| function | TestFunctionAdamicUnusedAnnotationRefused | 2.5 |
| function | TestFunctionAdamicCallAnnotationRefused | 0.2 |
| function | TestFunctionTypeScriptUnusedAccepted | 2.73 |
| function | TestFunctionCallSignatureAgreesWithNode | 0.49 |
| function | TestFunctionTypeScriptLengthAgreesWithNode | 1.68 |
| counts-stable | TestHeaderCountsStayRecorded | 0 |

Compiler Go, runtime lines, dependency declarations and the original counts table are unchanged relative to 0a8d7ed8. No new runtime feature set or compiler-consuming package was added. The original slice's own implementation validation and 14 production mutant results are retained in its earlier report; this header pass reran the header-affected checks and its two new predicate mutants.

## Setup and delivery

Used the cloud runtime skill and checked the network policy. Toolchain command: export GOPROXY='https://proxy.golang.org|direct'; ADAMIC_GOCACHE_OFF=1 timeout 600 bash cloud/setup.sh, then source the env file it printed. Local-cache mode is the supported fallback established after the shared cache stalled during the original unit. This setup succeeded with no new failure. nproc=5; quota is four CPUs. The full timing lines are:

```text
setup: node ready (0.051s)
setup: submodules ready (0.058s)
setup: go ready (0.063s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: markdown dependencies ready (0.129s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.267s)
setup: shared cache off (ADAMIC_GOCACHE_OFF=1) (0.269s)
setup: go build ready (7.124s)
setup: test binaries deferred (use --warm-tests) (7.220s)
setup: build cache warm (7.222s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=off load-before=0.00 0.00 0.00 2/135 714 load-after=1.00 0.21 0.07 1/135 1201
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (7.245s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.W4sP5i
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. No ahra command, PR or Loom submission. Push only compiler/loom-r-slice-3-ac, never main or the earlier branch. The unrelated earlier BLOCKED.md remains untracked. The tools binary and Python cache are local artifacts excluded from the commit.

Built: a-check audit of the complete checked-union change on compiler/loom-r-unions-ac; no headers were needed.
Commits: source b2d049da7403461991e92861cf582dac0617acda; exact base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is the pushed branch tip.
Commands and outputs: all 22 changed .a programs compiled with this branch's compiler; pinned Gate.aCheck accepted all 22; path readers, Node/sanitizer fixtures, reader guard, counts table and vet passed.
Mutants: removed-header and wrong-reason mutants both rejected by the unchanged pinned predicate; the control fixture was restored byte for byte.
Not covered: full repository a-check/full gate, all compiler mutants or all reader packages rerun; inherited typeaware timeout was not rerun.

This commit reapplies only b2d049da7403461991e92861cf582dac0617acda on exactly dbdf977f5b8b0979d414ebe6f20e7aadd312e309 using git cherry-pick --no-commit, then adds this audit evidence. There were no conflicts. The requested branch starts from the pinned base rather than a newer origin/main. CLAUDE.md, docs/front-door.md, the whole stage3/a-check-headers/README.md and its reproduction harness were read. The prerequisite language/memory documentation was already read for this unchanged source change.

Scope: git diff --name-only --no-renames --diff-filter=AM dbdf977f5b8b0979d414ebe6f20e7aadd312e309 b2d049da7403461991e92861cf582dac0617acda -- "*.a" selects 22 programs, all under internal/oracle/testdata. The selection scans all tracked paths, including review/. There are no added or modified review .a files in this specific commit. paths.json is the exact manifest. No programs were skipped because of directory, expected refusal or an existing report.

Observed totals: 16 checked, 6 NotYet, 0 refused, 0 type errors, 0 unclassified. Headers added: none. The _refused suffix of the six storage probes is a fixture-layout choice, not their compiler status; their real outcomes are NotYet and the README explicitly says they take no header. Every body SHA-256 remains identical to the source commit. No line moved, so no positional expectation or historical report was updated. path-readers.txt records 17 Go reader matches, including the non-null control's pinned :3:25 position, which its test passed again after restoration.

Compiler: source /workspace/adamic-tools/env.sh, then timeout 180 go build -o /tmp/adamic-gate/unions-ac-adamic ./cmd/adamic. The working source tree includes the entire union change. Each program was run as adamic c PATH with a 60-second outer limit, stdout/stderr directly recorded under this evidence directory. No alternate compiler or TypeScript-only checker was substituted. Exact per-file argv, diagnostic text, exit status, wall seconds and unchanged body hashes are in statuses.json and statuses.csv.

| File | Status | Actual reason or code | Compiler seconds |
|---|---|---|---:|
| internal/oracle/testdata/scout_boxed_scalar_fields.a | checked |  | 0.067 |
| internal/oracle/testdata/scout_boxed_scalar_stale_field.a | checked |  | 0.070 |
| internal/oracle/testdata/scout_class_boolean_fields.a | checked |  | 0.067 |
| internal/oracle/testdata/scout_hir_optional_boolean.a | checked |  | 0.069 |
| internal/oracle/testdata/scout_nullable_object_null_field.a | checked |  | 0.065 |
| internal/oracle/testdata/scout_nullable_object_stale_field.a | checked |  | 0.065 |
| internal/oracle/testdata/scout_nullable_object_stale_local.a | checked |  | 0.068 |
| internal/oracle/testdata/scout_nullable_objects.a | checked |  | 0.064 |
| internal/oracle/testdata/scout_nullable_storage_refused/array.a | NotYet | sharing nullable object storage with different null/undefined tags; copy values into tagged storage yet | 0.065 |
| internal/oracle/testdata/scout_nullable_storage_refused/array_spread.a | NotYet | sharing nullable object storage with different null/undefined tags; copy values into tagged storage yet | 0.065 |
| internal/oracle/testdata/scout_nullable_storage_refused/callable.a | NotYet | sharing nullable object storage with different null/undefined tags; copy values into tagged storage yet | 0.064 |
| internal/oracle/testdata/scout_nullable_storage_refused/field.a | NotYet | sharing nullable object storage with different null/undefined tags; copy values into tagged storage yet | 0.068 |
| internal/oracle/testdata/scout_nullable_storage_refused/lookup.a | NotYet | a nullable object lookup without separate null/undefined tags yet | 0.069 |
| internal/oracle/testdata/scout_nullable_storage_refused/object_spread.a | NotYet | sharing nullable object storage with different null/undefined tags; copy values into tagged storage yet | 0.114 |
| internal/oracle/testdata/scout_nullable_string_stale_field.a | checked |  | 0.065 |
| internal/oracle/testdata/scout_nullable_strings.a | checked |  | 0.115 |
| internal/oracle/testdata/scout_selector_optional_boolean.a | checked |  | 0.066 |
| internal/oracle/testdata/scout_string_number_fields_locals.a | checked |  | 0.064 |
| internal/oracle/testdata/scout_string_number_stale_field.a | checked |  | 0.066 |
| internal/oracle/testdata/scout_tsc_nullish_content.a | checked |  | 0.064 |
| internal/oracle/testdata/scout_tsc_source_content.a | checked |  | 0.065 |
| internal/oracle/testdata/scout_values_optional_boolean.a | checked |  | 0.065 |

Predicate proof: the exact Gate.aCheck method was loaded unchanged from fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py, as named by the README. Its source snapshot is gate.py.txt. The harness supplied only process, build and logging methods, with hard limits; the real method built and invoked the branch compiler, read real first-line headers and judged real compiler diagnostics. No captured/fabricated diagnostic was substituted. Gate classifies both clean and NotYet as checked; the status table above deliberately distinguishes them.

Conservative assumption: because none of the changed programs needs a header, use an existing headed refusal as the predicate control rather than add a new refused program to this change. The control is internal/oracle/testdata/non_null_refused/initialized.a. Its original first line is // a-check: refused the non-null assertion !. The ordinary baseline passes. Removing only that line yields an actual refusal where the predicate expects checked, so exit 1. Replacing only its reason with impossible union header mutant reason yields a refusal whose diagnostic does not contain that reason, so exit 1. Restoring the exact original bytes passes again. Its body and original header are not part of the final diff. This assumption is also stated in the commit message.

| Predicate check | Exit | Seconds |
|---|---:|---:|
| changed-programs | 0 | 6.100 |
| control-header | 0 | 0.000 |
| removed-header | 1 | 0.000 |
| wrong-reason | 1 | 0.000 |
| restored-header | 0 | 0.000 |

Both mutant patches are retained as .patch files. predicate-results.json records every command, real diagnostic, predicate outcome and restoration hash. The removed-header mutant naturally moves the control diagnostic by one line during the experiment, then restoration returns its pinned line exactly. No checker-error control was needed because the scoped scan found no checker-error programs.

Rerun validation, all with GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1; output goes directly to named logs. Go tests use -count=1 -timeout 90s -parallel 4 -json and a 180-second outer limit. Every completed test is below 60 seconds (maximum 3.62 seconds). All argv and leaf times are in check-results.json and test-times.json.

| Check | Result | Command seconds |
|---|---|---:|
| oracle-path-readers | exit 0 | 7.555 |
| union-fixtures | exit 0 | 9.402 |
| reader-guard | exit 0 | 5.665 |
| vet | exit 0 | 1.073 |

oracle-path-readers covers all six NullableObjectStorage tests, NullableStringFieldCheckMutant, NonNullAdamicRefusal including both pinned-position children, and CountsAreRecordedTable. union-fixtures holds the 16 admitted scout fixtures to the existing source Node/native ASan/UBSan/leak/backend checks, including expected stale-flow panics. reader-guard is internal/ir TestCallTargetReaders. Vet covers internal/lower, internal/native, internal/oracle, stage1/cohere/selector, stage1/cohere/values and internal/ir.

Counts were not regenerated: counts.md is byte-for-byte identical to b2d049da7403461991e92861cf582dac0617acda. No program bytes changed, and the table registration check passes. All compiler/runtime files and pre-existing test files likewise remain identical to that source commit. This pass adds no runtime features, IR maps, compiler-consuming package or build product. No cohere implementation was copied.

The source commit's prior broad coverage and 22 compiler-mutant plus IR-mutant evidence remain under review/compiler/loom-r-unions. Its TestInspectRequestRefusals timeout at 60 seconds is a carried limitation, not rerun or relabeled green by this header audit. This pass reruns the path-sensitive checks relevant to a possible header and the requested reader guard/vet; it does not claim a fresh exhaustive 60-package run or full gate.

Setup: export GOPROXY=https://proxy.golang.org|direct; timeout 600 bash cloud/setup.sh; source the printed /workspace/adamic-tools/env.sh. It completed successfully in 17.982 seconds. nproc=5; cpu.max=400000 100000, four CPU shares. Timing lines:

```text
setup: submodules ready (0.066s)
setup: node ready (0.072s)
setup: go ready (0.072s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.019s
setup: markdown dependencies ready (0.180s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.385s)
setup: shared cache ready (0.808s)
setup: go build ready (17.877s)
setup: test binaries deferred (use --warm-tests) (17.957s)
setup: build cache warm (17.958s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/136 697 load-after=0.46 0.10 0.03 1/140 1308
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (17.982s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.dW6zlI
```

No PR, ahra CLI invocation or Loom submission was made. Only compiler/loom-r-unions-ac is pushed, without force.

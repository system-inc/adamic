First scanner stop: debug.ts:8:5, mutable namespace export, in split 0 and 1.
Records 8c013f1c merges cleanly; area/library 71f91687 conflicts in 50 paths.
Records-only compiler build exits 0; both Node/full-tree byte comparisons exit 0.
False-pass conflict-summary mutant exits 1; five targeted tests pass; Node end mutants caught.
No native scanner binary, native byte comparison, native output mutant or native runtime timing.

## Inputs and commands

Fetched main: `45487a809f89885a3fc651cd590e7dabf31362dc`. Records ref:
`8c013f1cf3ec69fe80e60501740c2c45e1d1114c`. Library ref:
`71f916872667150037677b4b975772f165e47898`. Records-only scratch merge:
`76811ae2366231ea69fa42bb241160ee2e33154f`. Compiler binary SHA-256:
`98c910c304c6ab3d2b20ee02308b1c840fd5f29f6d6a9eec095d18c954de3643`.

```sh
export GOPROXY='https://proxy.golang.org|direct'
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-step12-records-library origin/codex/records-maplike-next origin/area/library > /tmp/scanner-step12-records-library.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-step12-records-alone origin/codex/records-maplike-next > /tmp/scanner-step12-records-alone.log 2>&1
```

The first command exits 2 after the library merge conflicts. It records every
path, aborts the conflicting merge and stops before setup or compilation.
The second command exits 1, status scanner-blocked. Its records merge is clean.
Both scratch branches remain unpushed. The delivery branch merged origin/main
without conflicts; its main-merge log remains /tmp/scanner-step12-delivery-main-merge.log.
Only evidence is added by this unit; no compiler, adaptation or runner edits.

## First observed stop and comparison

```text
debug.ts:8:5: stage 0 can't lower a mutable namespace export; use a module or export functions around private state yet
```

Both unchanged native builds report this same location and message. A fresh
small program succeeds on Node (exit 0) and fails on this compiler (exit 1)
with the identical diagnostic; its source and logs are in records-alone-logs.json.gz.
The compiler admitted the scanner far enough to reach this stop; this does not
establish native scanner correctness or identify what a clean library integration
would do afterward.

The full-tree and both slice Node streams cover 81 files twice, with 509,014
skip-trivia tokens and 860,418 retain-trivia tokens, plus 466 error rows. Each
slice/full-tree diff exits 0. Raw output SHA-256:
`d71f77c31f76c7730c96b12862dc6a519915cbdc2eddd26324b71af070247c83`; bytes:
108,022,365. Absolute input paths occur in pass headers.
Both Node controls compare equal; both token-end mutants compare unequal
(diff exit 1). The full-tree reference also catches its token-end mutant.
No native-output mutant is claimed while compilation is blocked.

The existing command automatically continues in its separate discovery copy.
It records seven freshly witnessed stops, then stops at the eighth diagnostic,
scanner.ts:696:32, a string/number BinaryExpression, because no freshly matching
Node witness exists in its catalogue. Stops after the first depend on private
throwing placeholders. Their complete diagnostics, witness programs and logs
are retained; discovery-placeholders.json.gz records every temporary source
change. None of those copies is committed as an adaptation or used to claim a pass.

| Phase | Exit | Wall seconds |
| --- | --- | --- |
| setup | 0 | 206.44 |
| compiler-build | 0 | 6.162 |
| apply | 0 | 21.123 |
| reference | 0 | 10.417 |
| slice | 0 | 3.686 |
| scanner-0 | 1 | 7.245 |
| full-tree-check-0 | 0 | 0.216 |
| scanner-1 | 1 | 7.149 |
| full-tree-check-1 | 0 | 0.164 |

Records-only command total: 297.457s. nproc: 5; cgroup cpu.max: 400000 100000. All setup timing lines:

```text
setup: node ready (0.023s)
setup: go ready (0.027s)
setup: submodules ready (0.065s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.007s
setup: markdown dependencies ready (0.073s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.179s)
setup: go build ready (206.179s)
setup: test binaries deferred (use --warm-tests) (206.386s)
setup: build cache warm (206.388s)
setup: build-flags commit=76811ae2366231ea69fa42bb241160ee2e33154f nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.13 0.03 0.17 1/213 37039 load-after=6.84 4.10 1.79 1/220 40020
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (206.423s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.PyzH60
```

## Exact library conflicts

Also recorded in library-conflicts.txt and records-library-summary.json.

```text
internal/flow/flow_test.go
internal/fresh/fresh.go
internal/ir/ir.go
internal/javascript/javascript.go
internal/load/load.go
internal/load/node_library.go
internal/load/source_fs.go
internal/lower/cast.go
internal/lower/cast_proof.go
internal/lower/control.go
internal/lower/exceptions.go
internal/lower/expression.go
internal/lower/functions.go
internal/lower/generic.go
internal/lower/invariance.go
internal/lower/library_json_stringify.go
internal/lower/library_math_number.go
internal/lower/library_node_buffer.go
internal/lower/library_node_fs_file.go
internal/lower/library_node_fs_file_test.go
internal/lower/library_node_test.go
internal/lower/library_string.go
internal/lower/modules.go
internal/lower/object.go
internal/lower/refusals.go
internal/lower/statements.go
internal/native/class_accessors.go
internal/native/emit_expressions.go
internal/native/emit_functions.go
internal/native/emit_objects.go
internal/native/fields.go
internal/native/library.go
internal/native/native.go
internal/native/runtime/adamic.h
internal/native/runtime/class_static.c
internal/native/runtime/heap.c
internal/native/runtime/node_fs_file.c
internal/native/runtime/regexp.h
internal/oracle/counts.md
internal/oracle/node_fs_file_test.go
internal/oracle/oracle_test.go
internal/oracle/testdata/node_fs_file_close.a
internal/oracle/testdata/optional_widening_refused/node_fs_file_buffer.a
review/unknown-narrowing/require_pending.a
stage1/cohere/yaml/GAPS.md
stage1/cohere/yaml/gaps_test.go
stage3/fixtures/cycles/status.json
stage3/fixtures/host/status.json
stage3/fixtures/objects/status.json
stage3/fixtures/predicates/status.json
```

## Checks, mutant and retained evidence

```sh
python3 -B stage3/drivers/scanner/test_scratch_runner.py > stage3/drivers/scanner/evidence/step12-records-next/tests.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/step12-records-next/records-library-summary.json > stage3/drivers/scanner/evidence/step12-records-next/conflict-summary-check.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/step12-records-next/conflict-pass-mutant.json > stage3/drivers/scanner/evidence/step12-records-next/conflict-mutant-check.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/step12-records-next/records-alone-summary.json > stage3/drivers/scanner/evidence/step12-records-next/records-alone-summary-check.log 2>&1
git diff --cached --check > /tmp/scanner-step12-staged-check.log 2>&1
```

The five existing targeted tests pass. Both real summaries validate with exit 0.
The actual conflict summary is mutated to status pass with complete synthetic
successful compiler/scanner claims, preserving its real failed merge and 50
conflicting paths. Validation exits 1, caught specifically by the two merge
rules: a conflict cannot be reported as pass, and pass requires every merge to
exit zero. The tests also catch a missing mode, compilation after conflict and
an incorrect native-byte-mutant diff result. Native execution was not exercised.

Raw phase logs, per-mode reports and fresh witness programs are losslessly
archived in records-library-logs.json.gz and records-alone-logs.json.gz.
Full raw token outputs and both scratch worktrees remain in the output directories.
Old regenerable Go cache entries were reclaimed before the build to make room;
no source or prior evidence was removed. No oracle fixture was added, so
internal/oracle/counts.md is unchanged. No package-wide tests or full gate ran.
The command completed without toolchain failure. The library merge remains
unresolved; native scanner execution and comparison remain blocked at the first stop.

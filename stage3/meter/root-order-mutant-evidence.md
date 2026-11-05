# Root-order artifacts and mutant evidence, task 82gkca0

Base origin/main: `5c9fc8052ba0eb07c1c2a56e58b4e18cbd582602`. All changes are under stage3/meter.

## Node checks

Corrected after Lower's review of fnqwkds. The earlier ordinary-elision Node
runs used a different module graph and are superseded by these results.
Node v24.19.0 and TypeScript 6.0.3 ran the same pinned adapted source as before.
`LOADER_OPTIONS` in node_readiness.py mirrors compilerOptions() in
internal/load/load.go, including verbatimModuleSyntax true, ESNext modules,
bundler resolution, forced modules, ES2024 target/lib, empty types and strict
checker settings. Only emit settings are overridden, as documented beside it.
Value imports used only as types and const enums remain real evaluation edges.
The effective config is retained in each probe's compiler-options.json.
Emit and timeout results were 0 and false, respectively, for all three roots.

| Root | Node exit | First error line |
| --- | ---: | --- |
| src/compiler/debug.ts | 1 | TypeError: Cannot read properties of undefined (reading 'assert') |
| src/compiler/diagnosticInformationMap.generated.ts | 1 | TypeError: Cannot read properties of undefined (reading 'AbstractKeyword') |
| src/compiler/types.ts | 1 | TypeError: Cannot read properties of undefined (reading 'Error') |

The diagnostic-map root throws at scanner.ts:136:26, before the anticipated
commandLineParser FixedPollingInterval read. The types.ts root throws at
diagnosticInformationMap.generated.ts:13:64. These are the observed Node stack
locations; the census stop locations below remain unchanged. All three roots
qualify as artifacts against the saved entry evidence. This excludes only the
readiness stop, and does not prove that the root program lowers whole. A nonzero exit without
an error line, an emit failure, or a timeout still proves no artifact.

## Replays

S1c record: `26973f0b0728cab64817cbe9df9efe97805ebac9`, `stage3/meter/runs/20261104T213555Z.FUPT2E`.
The saved ordinary census has 69 source roots and no aggregate row. Its saved
entry census attempts units in semver.ts and commandLineParser.ts and has no
readiness finding at either stop file. No complete census is claimed.
Only the absolute adapted-tree prefix was remapped to the locally prepared tree.
The artifact-only command completed and printed:

```text
Recorded roots only; census coverage is not asserted.

Lowering, recorded roots: 2/69; root-order artifacts: 2.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | NotYet src/compiler/binder.ts:427:31: a Map or Set interface implemented by a value that isn&#x27;t a Map or Set (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;AbstractKeyword&#x27;) | root-order artifact | NotYet src/compiler/binder.ts:427:31: a Map or Set interface implemented by a value that isn&#x27;t a Map or Set (read from the entry census; source order) |
```

S2a complete ordinary census: `268b0a5e6b1f29b05dce9389071bdb5abd6eb2aa`,
`stage3/meter/s2a-first-stops/census.jsonl`, all 79 source roots, three
non-source inputs and the aggregate row. This reporter replay pairs those
ordinary outcomes with the saved S1c entry evidence. It is a cross-revision
replay, not a new compiler census. Adaptation executables, inventories, parser
locks and the source pin agree across both evidence refs and this base;
S1c differs only in the unused `47-host-errors/verify.cjs` check. The local
adapted source was produced by the same adapters. Only absolute tree prefixes
were remapped, not outcomes, findings or unit statuses.

The normal single-tree report completed with checker 79/79, own file 79/79,
and lowering 2/79: only the two ordinary accepted roots. The previous 3/79
inference from diagnosticInformationMap's clean own file was wrong: it reaches
the same compiler as the other two roots. All three artifact roots have 79
implementation source files, individually enumerated by tsc --listFilesOnly
using the same compiler options as emit. Each result retains its full
program_files array and program_files_complete true. The entry attempted units
in all 79, but findings in those files prevent whole-program lowering.

The compiler file list is not an ES-module evaluation order. The next stop uses
the explicitly labeled source-order fallback: normalized source path, numeric
line/column, then stream order for ties. Under that fallback all three next
stops are NotYet at src/compiler/binder.ts:427:31, reason "a Map or Set interface
implemented by a value that isn't a Map or Set". This does not claim that binder
executes first. No entry finding outside the root's program can be selected.
The same whole-program predicate governs normal and partial-report counts.

The report retains all three artifacts and prints the next stops behind them:

Lowering, file roots: 2/79; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | NotYet src/compiler/binder.ts:427:31: a Map or Set interface implemented by a value that isn&#x27;t a Map or Set (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;AbstractKeyword&#x27;) | root-order artifact | NotYet src/compiler/binder.ts:427:31: a Map or Set interface implemented by a value that isn&#x27;t a Map or Set (read from the entry census; source order) |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;Error&#x27;) | root-order artifact | NotYet src/compiler/binder.ts:427:31: a Map or Set interface implemented by a value that isn&#x27;t a Map or Set (read from the entry census; source order) |

Commands after extracting and remapping the saved streams into scratch:

```sh
source /workspace/adamic-tools/env.sh
python3 -E -S -B stage3/meter/node_readiness.py /workspace/scratch/root-order-adapted /workspace/scratch/root-order-verbatim-three-node /workspace/scratch/root-order-adapted/src/compiler/debug.ts /workspace/scratch/root-order-adapted/src/compiler/diagnosticInformationMap.generated.ts /workspace/scratch/root-order-adapted/src/compiler/types.ts > /workspace/scratch/root-order-verbatim-three-node.json 2>&1
python3 -E -S -B stage3/meter/report.py --root-order-only /workspace/scratch/root-order-adapted /workspace/scratch/root-order-replay/s1c > /workspace/scratch/root-order-replay/s1c-report.log 2>&1
python3 -E -S -B stage3/meter/report.py /workspace/scratch/root-order-adapted /workspace/scratch/root-order-replay/s2a replay > /workspace/scratch/root-order-replay/s2a-report.log 2>&1
```

## Tests

```sh
python3 -E -S -B -m unittest discover -s stage3/meter -p '*_test.py' > /workspace/scratch/root-order-python-tests.log 2>&1
go test -count=1 -v ./stage3/meter > /workspace/scratch/root-order-go-tests.log 2>&1
go test -count=1 ./stage3/lane -run 'TestEveryStage3PythonSuiteRunsOrSaysWhy|TestPythonSuiteGapsNamesWhatNoWrapperRuns' > /workspace/scratch/root-order-python-census.log 2>&1
python3 -E -S -B stage3/meter/check-root-order-mutants.py > /workspace/scratch/root-order-mutants.log 2>&1
python3 -E -S -B stage3/meter/check-unit-time-mutant.py > /workspace/scratch/root-order-prior-time-mutant.log 2>&1
```

Python results:

```text
.........s..........ss.....ss...................................
----------------------------------------------------------------------
Ran 64 tests in 12.392s

OK (skipped=5)
```

The five skips are the existing optional checker/latent/entry binary probes.
The Node tests use the actual installed tsc and Node, including a TDZ cycle,
explicit type-only import elision, preserved value imports used only as types,
plain exit 1, timeout, and emit failure. Go's five
wrappers all pass; the Python suite census passes. The existing time-block
mutant is still caught by its named test with `KeyError: 'latent_time'`.
The generator hook fixture now stubs the three IR types that the current
state-copy generator names, preserving its populated-hook rejection mutant.

## Direction, whole-program and import-graph mutants

Each runs a scratch reporter or Node runner and requires its exact named test
to fail with an assertion. No production compiler is mutated. The updated
artifact-clean-own-file-implies-whole mutant unconditionally passes artifact
roots, and must fail the test whose clean root reaches a refused dependency.
Separate tests retain own-file refusals and missing attempted program units.
The real Node test also pins distinct per-root file sets, excluding unrelated
roots emitted in the same batch, while retaining type-only source dependencies.

```text
artifact-clean-own-file-implies-whole: test_artifact_with_clean_own_file_and_refused_dependency_does_not_lower_whole
F
======================================================================
FAIL: test_artifact_with_clean_own_file_and_refused_dependency_does_not_lower_whole (report_test.RootOrderTests.test_artifact_with_clean_own_file_and_refused_dependency_does_not_lower_whole)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 610, in test_artifact_with_clean_own_file_and_refused_dependency_does_not_lower_whole
    self.assertFalse(self.root_row(result)['lowering'])
AssertionError: True is not false

----------------------------------------------------------------------
Ran 1 test in 0.007s

FAILED (failures=1)
ignore-entry-readiness: test_entry_order_readiness_still_counts_against_root_file
F
======================================================================
FAIL: test_entry_order_readiness_still_counts_against_root_file (report_test.RootOrderTests.test_entry_order_readiness_still_counts_against_root_file)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 479, in test_entry_order_readiness_still_counts_against_root_file
    self.assertEqual(result['root_order_artifacts'], [])
AssertionError: Lists differ: [{'root': 'src/compiler/later.ts', 'where'[830 chars]r'}}] != []

First list contains 1 additional elements.
First extra element 0:
{'root': 'src/compiler/later.ts', 'where': 'src/compiler/good.ts:7:9', 'reason': 'a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration', 'node': {'program_files': ['/tmp/adamic-gate/tmpuqzdjs84/adapted/src/compiler/later.ts', '/tmp/adamic-gate/tmpuqzdjs84/adapted/src/compiler/good.ts'], 'program_files_complete': True, 'node_version': 'v24.19.0', 'build_exit': 0, 'exit': 1, 'timeout': False, 'throws': True, 'first_error': "ReferenceError: Cannot access 'early' before initialization"}, 'entry_program_attempted': True, 'entry_unattempted_sources': [], 'entry_next_stop': {'kind': 'NotYet', 'where': 'src/compiler/good.ts:99:1', 'reason': 'reading an enum before its runtime initialization; move the call after the enum declaration', 'source': 'entry census', 'order': 'source order'}}

Diff is 1266 characters long. Set self.maxDiff to None to see it.

----------------------------------------------------------------------
Ran 1 test in 0.009s

FAILED (failures=1)
count-root-order-artifact-as-failure: test_root_only_readiness_with_node_throw_is_visible_artifact
F
======================================================================
FAIL: test_root_only_readiness_with_node_throw_is_visible_artifact (report_test.RootOrderTests.test_root_only_readiness_with_node_throw_is_visible_artifact)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 486, in test_root_only_readiness_with_node_throw_is_visible_artifact
    self.assertTrue(self.root_row(result)['lowering'])
AssertionError: False is not true

----------------------------------------------------------------------
Ran 1 test in 0.007s

FAILED (failures=1)
elide-value-import-used-only-as-type: test_value_import_used_only_as_type_keeps_its_evaluation_edge
F
======================================================================
FAIL: test_value_import_used_only_as_type_keeps_its_evaluation_edge (node_readiness_test.NodeReadinessTests.test_value_import_used_only_as_type_keeps_its_evaluation_edge)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/node_readiness_test.py", line 55, in test_value_import_used_only_as_type_keeps_its_evaluation_edge
    self.assertTrue(result['throws'], result)
AssertionError: False is not true : {'program_files': ['/tmp/adamic-gate/tmp_zr17n5c/flags.ts', '/tmp/adamic-gate/tmp_zr17n5c/main.ts'], 'program_files_complete': True, 'program_files_exit': 0, 'program_files_timeout': False, 'node_version': 'v24.19.0', 'build_exit': 0, 'exit': 0, 'timeout': False, 'first_error': '', 'throws': False}

----------------------------------------------------------------------
Ran 1 test in 0.421s

FAILED (failures=1)
All three root-order mutants and the value-import elision mutant went red on their named tests.
```

`RootOrderTests.test_root_only_readiness_with_clean_node_still_counts` separately
pins the third outcome: a qualifying per-root enum stop still fails its file
when Node exits 0. The daily integration regression asserts both labels' counts,
full artifact rows, JSON lists, and the generated README table.


## Gate repair on the re-cut branch, 2026-11-05

Base origin/main: `ab8a85bebf2a14a06fa8c3c4ca8cc93201a2309f`.
The five Python suites write their fixtures, generated Go packages, probe
outputs and logs under independent temporary directories. They do not mutate
package source files. All five Go wrappers now call `t.Parallel()` first.

`report` requires a full compiler commit argument or `STAGE3_METER_COMMIT`.
An absent or malformed commit raises ValueError. The reporter contains no git
commands. `write_readme` and `report_pair` require a directory containing the
prior runs. `twice-daily.sh` supplies an archive of the pinned main revision's
tracked runs, preserving the old reader's input revision and excluding dirty
working-tree runs. It also passes the checkout commit explicitly. Compiler
metadata continues to supply each measured compiler's pin when present.
The Oct 19 regression reads the committed Oct 19 and Oct 8 reports from disk.

### No-checkout proof

A full `git archive HEAD` was extracted to
`/workspace/scratch/meter-no-git`, then the changed meter sources were copied
into it. There is no `.git` in that tree or its ancestors. The compiler-selection
suite creates its own temporary Git repositories as fixtures; it does not need
the source tree to be a checkout. No suites were omitted.

```sh
cd /workspace/scratch/meter-no-git
source /workspace/adamic-tools/env.sh
python3 -E -S -B -m unittest discover -s stage3/meter -p '*_test.py' > /workspace/scratch/meter-no-git-final.log 2>&1
```

```text
Ran 65 tests in 101.138s
OK (skipped=5)
```

Zero failures and zero errors. The five skips are the existing optional binary
probes. This run overlapped cold setup and full apply; it is not the unit timing
measurement below.

### Mutants after the repair

```sh
python3 -E -S -B stage3/meter/check-root-order-mutants.py > /workspace/scratch/meter-mutants.log 2>&1
```

| Mutant | Named test | Result |
| --- | --- | --- |
| artifact-clean-own-file-implies-whole | test_artifact_with_clean_own_file_and_refused_dependency_does_not_lower_whole | red, AssertionError |
| ignore-entry-readiness | test_entry_order_readiness_still_counts_against_root_file | red, AssertionError |
| count-root-order-artifact-as-failure | test_root_only_readiness_with_node_throw_is_visible_artifact | red, AssertionError |
| elide-value-import-used-only-as-type | test_value_import_used_only_as_type_keeps_its_evaluation_edge | red, AssertionError |

The harness exited 0 and confirmed all three root-order mutants and the
value-import elision mutant went red on their named tests.

### Whole Go checks and unit times

```sh
go test -count=1 -parallel=1 -json ./stage3/meter/... > /workspace/scratch/go-meter-final.jsonl 2>&1
go test -count=1 -json ./cmd/adamic-gate/... > /workspace/scratch/go-gate.jsonl 2>&1
```

Both commands exited 0. Cold setup completed successfully in 734.423 seconds.
The meter run uses `-parallel=1` to measure the wrappers without contention
between suites; the wrappers remain parallel-safe.

| Meter unit | Seconds | Result |
| --- | ---: | --- |
| TestReportPython | 0.45 | pass |
| TestGeneratorPython | 1.59 | pass |
| TestNodeReadinessPython | 9.69 | pass |
| TestCompilerPython | 17.99 | pass |
| TestEntryPython | 0.17 | pass |

Meter package elapsed: 29.912 seconds. All five units are below 30 seconds.
The first run overlapped cold setup and other validation; it passed but its
compiler wrapper took 41.06 seconds. The isolated measurements above replace
that run for the timing check, without code changes to the compiler suite.

The gate packages passed in 10.137 seconds (adamic-gate) and 9.615 seconds
(parallelcensus). All 15 gate tests are below 30 seconds; the maximum is 10.09
seconds. TestEveryTestIsParallelOrSaysWhy passed in 9.09 seconds.

### Nov 4 S2a byte comparison

The replay uses the same saved inputs described above: the S2a complete ordinary
census from `268b0a5e6b1f29b05dce9389071bdb5abd6eb2aa`, paired with the S1c
entry streams from `26973f0b0728cab64817cbe9df9efe97805ebac9`. Only absolute
adapted-tree prefixes were remapped. This remains a cross-revision replay,
not a fresh census. The adapted tree was produced by full apply on the re-cut
meter branch before either adaptation was added.

`/workspace/scratch/report-before.py` is the reporter before the gate fixes.
`/workspace/scratch/compare-reports.py` runs both that reporter and the fixed
one against the same tree, run directory, stamp and explicit compiler pin
`424eb3dd3baed2265f8af65e3c1c7b7d55e4e7e0`. Both execute the real Node probes.
Their JSON and Markdown outputs were saved separately and compared with `cmp`.

```sh
source /workspace/adamic-tools/env.sh
python3 -E -S -B /workspace/scratch/compare-reports.py > /workspace/scratch/report-comparison.log 2>&1
cmp /workspace/scratch/replay/s2a/before-report.json /workspace/scratch/replay/s2a/after-report.json
cmp /workspace/scratch/replay/s2a/before-report.md /workspace/scratch/replay/s2a/after-report.md
```

Both `cmp` calls exited 0. `/workspace/scratch/report-cmp.log` records the
exact byte comparisons. Both reports have 82 files, 79 source roots, checker
79/79, own file 79/79, lowering 2/79 and three root-order artifacts.

| Output | SHA256 before and after |
| --- | --- |
| report.json | 69d3487f20dbb079c187445a22756fd6ea6108528eddac5ab44d63f8582278d2 |
| report.md | fa0723f0602834529940e58455b1c37b0fa0773518f929197724020824c41633 |

The existing `check-unit-time-mutant.py` also exited 0: its dropped
`latent_time` field was caught with the expected KeyError.

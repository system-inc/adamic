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
readiness stop, and does not prove that the root file lowers whole. A nonzero exit without
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

Lowering, recorded roots: 3/69; root-order artifacts: 2.

| Root | Stop | Reason | Node result | Classification | File result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | Refused src/compiler/debug.ts:167:31: a cast the runtime can&#x27;t check (read from the entry census) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;AbstractKeyword&#x27;) | root-order artifact | units attempted; no findings; lowers whole |
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
and lowering 3/79: two ordinary accepted roots plus diagnosticInformationMap.
The previous 5/79 inference from three artifacts was wrong. Debug has an entry
refusal at debug.ts:167:31 and types has an entry refusal at types.ts:4560:38.
Both stay out of the lowering count. DiagnosticInformationMap has attempted
entry units with no file-local findings, so it counts as lowering whole.
The first finding is selected in entry-stream order by its actual location,
even if another file's unit encountered it. Every finding kind prevents a pass.
The same predicate now governs normal and partial-report totals.

The report retains all three artifacts and prints the next stops behind them:

Lowering, file roots: 3/79; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | File result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | Refused src/compiler/debug.ts:167:31: a cast the runtime can&#x27;t check (read from the entry census) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;AbstractKeyword&#x27;) | root-order artifact | units attempted; no findings; lowers whole |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;Error&#x27;) | root-order artifact | Refused src/compiler/types.ts:4560:38: a primitive brand member _isResolvedConfigFileName whose type is not void (read from the entry census) |

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
.........s..........ss.....ss.................................
----------------------------------------------------------------------
Ran 62 tests in 12.068s

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

## Direction, whole-file and import-graph mutants

Each runs a scratch reporter or Node runner and requires its exact named test
to fail with an assertion. No production compiler is mutated. The new
artifact-implies-whole mutant unconditionally passes artifact roots and must
fail test_artifact_with_entry_refusal_does_not_lower_whole. A separate test
also pins that missing attempted root units cannot establish whole lowering.

```text
artifact-implies-whole: test_artifact_with_entry_refusal_does_not_lower_whole
F
======================================================================
FAIL: test_artifact_with_entry_refusal_does_not_lower_whole (report_test.RootOrderTests.test_artifact_with_entry_refusal_does_not_lower_whole)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 575, in test_artifact_with_entry_refusal_does_not_lower_whole
    self.assertFalse(self.root_row(result)['lowering'])
AssertionError: True is not false

----------------------------------------------------------------------
Ran 1 test in 0.008s

FAILED (failures=1)
ignore-entry-readiness: test_entry_order_readiness_still_counts_against_root_file
F
======================================================================
FAIL: test_entry_order_readiness_still_counts_against_root_file (report_test.RootOrderTests.test_entry_order_readiness_still_counts_against_root_file)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 476, in test_entry_order_readiness_still_counts_against_root_file
    self.assertFalse(self.root_row(result)['lowering'])
AssertionError: True is not false

----------------------------------------------------------------------
Ran 1 test in 0.007s

FAILED (failures=1)
count-root-order-artifact-as-failure: test_root_only_readiness_with_node_throw_is_visible_artifact
F
======================================================================
FAIL: test_root_only_readiness_with_node_throw_is_visible_artifact (report_test.RootOrderTests.test_root_only_readiness_with_node_throw_is_visible_artifact)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 485, in test_root_only_readiness_with_node_throw_is_visible_artifact
    self.assertTrue(self.root_row(result)['lowering'])
AssertionError: False is not true

----------------------------------------------------------------------
Ran 1 test in 0.008s

FAILED (failures=1)
elide-value-import-used-only-as-type: test_value_import_used_only_as_type_keeps_its_evaluation_edge
F
======================================================================
FAIL: test_value_import_used_only_as_type_keeps_its_evaluation_edge (node_readiness_test.NodeReadinessTests.test_value_import_used_only_as_type_keeps_its_evaluation_edge)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/node_readiness_test.py", line 51, in test_value_import_used_only_as_type_keeps_its_evaluation_edge
    self.assertTrue(result['throws'], result)
AssertionError: False is not true : {'node_version': 'v24.19.0', 'build_exit': 0, 'exit': 0, 'timeout': False, 'first_error': '', 'throws': False}

----------------------------------------------------------------------
Ran 1 test in 0.301s

FAILED (failures=1)
All three root-order mutants and the value-import elision mutant went red on their named tests.
```

`RootOrderTests.test_root_only_readiness_with_clean_node_still_counts` separately
pins the third outcome: a qualifying per-root enum stop still fails its file
when Node exits 0. The daily integration regression asserts both labels' counts,
full artifact rows, JSON lists, and the generated README table.

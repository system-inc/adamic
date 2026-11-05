# Root-order artifacts and mutant evidence, task 82gkca0

Base origin/main: `5c9fc8052ba0eb07c1c2a56e58b4e18cbd582602`. All changes are under stage3/meter.

## Node checks

Setup completed with Node v24.19.0 and TypeScript 6.0.3. The adapted tree was
built by stage3/apply.sh from the pinned TypeScript v6.0.3 source. The runner
emits ES modules with tsc, without bundling or preserving const enums, then
executes each root in a separate Node process. Emit exit was 0 for all roots.

| Root | Node exit | First error line |
| --- | ---: | --- |
| src/compiler/debug.ts | 1 | TypeError: Cannot read properties of undefined (reading 'assert') |
| src/compiler/diagnosticInformationMap.generated.ts | 0 | none |
| src/compiler/types.ts | 0 | none |

The clean diagnostic-map result belongs to the separate lowering-gap task
fnqwkds. The clean types.ts result also remains a file failure. A nonzero exit
without an error line, an emit failure, or a timeout proves no artifact.

## Replays

S1c record: `26973f0b0728cab64817cbe9df9efe97805ebac9`, `stage3/meter/runs/20261104T213555Z.FUPT2E`.
The saved ordinary census has 69 source roots and no aggregate row. Its saved
entry census attempts units in semver.ts and commandLineParser.ts and has no
readiness finding at either stop file. No complete census is claimed.
Only the absolute adapted-tree prefix was remapped to the locally prepared tree.
The artifact-only command completed and printed:

```text
Recorded roots only; census coverage is not asserted.

Lowering, recorded roots: 3/69; root-order artifacts: 1.

| Root | Stop | Reason | Node result | Classification |
| --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node runs clean; exit 0;  | counts against root file |
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
and lowering 3/79 (two ordinary accepted roots and one artifact). It printed
this table in report.md:

Lowering, file roots: 3/79; root-order artifacts: 1.

| Root | Stop | Reason | Node result | Classification |
| --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node runs clean; exit 0;  | counts against root file |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node runs clean; exit 0;  | counts against root file |

Commands after extracting and remapping the saved streams into scratch:

```sh
source /workspace/adamic-tools/env.sh
python3 -E -S -B stage3/meter/node_readiness.py /workspace/scratch/root-order-adapted /workspace/scratch/root-order-three-node /workspace/scratch/root-order-adapted/src/compiler/debug.ts /workspace/scratch/root-order-adapted/src/compiler/diagnosticInformationMap.generated.ts /workspace/scratch/root-order-adapted/src/compiler/types.ts > /workspace/scratch/root-order-three-node.json 2>&1
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
.........s..........ss....ss...............................
----------------------------------------------------------------------
Ran 59 tests in 10.882s

OK (skipped=5)
```

The five skips are the existing optional checker/latent/entry binary probes.
The Node tests use the actual installed tsc and Node, including a TDZ cycle,
const-enum import elision, plain exit 1, timeout, and emit failure. Go's five
wrappers all pass; the Python suite census passes. The existing time-block
mutant is still caught by its named test with `KeyError: 'latent_time'`.
The generator hook fixture now stubs the three IR types that the current
state-copy generator names, preserving its populated-hook rejection mutant.

## Direction mutants

Each runs a scratch reporter and requires its exact named test to fail with
an assertion. No production compiler is mutated.

```text
ignore-entry-readiness: test_entry_order_readiness_still_counts_against_root_file
F
======================================================================
FAIL: test_entry_order_readiness_still_counts_against_root_file (report_test.RootOrderTests.test_entry_order_readiness_still_counts_against_root_file)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 475, in test_entry_order_readiness_still_counts_against_root_file
    self.assertFalse(self.root_row(result)['lowering'])
AssertionError: True is not false

----------------------------------------------------------------------
Ran 1 test in 0.011s

FAILED (failures=1)
count-root-order-artifact-as-failure: test_root_only_readiness_with_node_throw_is_visible_artifact
F
======================================================================
FAIL: test_root_only_readiness_with_node_throw_is_visible_artifact (report_test.RootOrderTests.test_root_only_readiness_with_node_throw_is_visible_artifact)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/meter/report_test.py", line 484, in test_root_only_readiness_with_node_throw_is_visible_artifact
    self.assertTrue(self.root_row(result)['lowering'])
AssertionError: False is not true

----------------------------------------------------------------------
Ran 1 test in 0.007s

FAILED (failures=1)
Both root-order mutants went red on their named tests.
```

`RootOrderTests.test_root_only_readiness_with_clean_node_still_counts` separately
pins the third outcome: a qualifying per-root enum stop still fails its file
when Node exits 0. The daily integration regression asserts both labels' counts,
full artifact rows, JSON lists, and the generated README table.

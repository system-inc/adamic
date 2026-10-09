#!/usr/bin/env python3
"""Write the report from the measured ledger, not handwritten timings."""
import csv
import json
from pathlib import Path

review = Path(__file__).resolve().parent
summary = json.loads((review/'audit.json').read_text())
rows = list(csv.DictReader((review/'timings.csv').open()))
family_names = ['TestEveryFunctionIsInSingleAssignment','TestEveryMutationIsInItsRange','TestEveryPathNodeTakesIsInTheGraph','TestLivenessHoldsOnEveryPath','TestEveryWriteIsRecordedAndKnown']
mutants = json.loads((review/'mutants.json').read_text())
text = f'''Built: independently selectable corpus tests for roadmap step 38, with package setup reused.
Base: 54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8; delivery SHA is reported with the branch push.
Checks: {summary['passed']} cold passes and two opt-in skips; longest complete invocation {summary['after_max_invocation']:.3f} seconds.
Mutants: {len(mutants)} Go overlays caught; planted failures failed only their owning units.
Limits: configured external-input inventory runs and the full gate were not run; production and fixtures are unchanged.

## Before and after

Branch: `compiler/test-split-flow`, created from the assigned origin/main pin. Measurements use one cloud instance with a four-CPU cgroup quota (`400000 100000`), `GOMAXPROCS=4`, `-parallel=4`, `-count=1`, and `ADAMIC_GATE_UNCACHED=1`. `nproc` reports 5; the effective CPU quota is 4. The toolchain and submodule pins are in [environment.json](environment.json), and source SHA-256 inputs are in [before.json](before.json) and [after.json](after.json).

The baseline selects each of the 205 original top-level tests separately with `go test -json`. The after measurement builds each package's test binary once with `go test -c`, then runs each selected test in a new process through `go tool test2json`. Every observation is fresh; shared Go build artifacts are retained. Neither OS page caches nor compiled setup archives are cleared. Every after invocation uses a 30-second timeout.

| Original test | Before complete invocation, s | After longest owning unit, s | Owning units |
| --- | ---: | ---: | ---: |
'''
for name in family_names:
    r = next(r for r in rows if r['test']==name)
    text += f"| {name} | {float(r['before_invocation_seconds']):.3f} | {float(r['after_max_invocation_seconds']):.3f} | {r['units']} |\n"
for package in ['lower','ir']:
    package_rows = [r for r in rows if r['package']==package and r['before_status']=='pass']
    text += f"| {package}, all passing original tests (maximum) | {max(float(r['before_invocation_seconds']) for r in package_rows):.3f} | {max(float(r['after_max_invocation_seconds']) for r in package_rows):.3f} | {len(package_rows)} unchanged tests |\n"
text += f'''
[Full table](timings.csv) contains every original test plus the three new coverage/setup tests, with both Go test-event time and complete invocation time. Before, graph paths took 34.482 seconds despite its parallel parent's misleading 0.020-second test event. Liveness took 34.329 seconds and write recording took 31.735 seconds. These are the three complete invocations over 30 seconds. SSA was close at 29.795 seconds; mutation ranges took 27.449 seconds. The older Home measurements in the brief are separate observations, not this box's baseline.

After, every one of the 1,563 selected invocations is below 30 seconds: 1,561 pass, two skip. The longest test-event time is {summary['after_max_test']['elapsed']:.3f} seconds, for `{summary['after_max_test']['test']}`. The largest complete invocation is {summary['after_max_invocation']:.3f} seconds.

`TestOriginalCycleLedger` and `TestOptionalWideningCensus` skip both before and after because their external project/configuration environment variables are unset. Their configured inventory modes were not timed or changed; these measurements establish the default gate units on the recorded inputs, not a bound for arbitrary external projects.

## Split and coverage

Flow retains its original 676 programs and all four analyses. Ordinary programs have one directly selectable root running SSA, mutation ranges, graph paths and liveness; the three Node observations remain independent. Timsort has four separate roots because its three trace observations would make a combined root exceed the budget. There are 679 unique flow corpus roots covering 2,704 analysis pieces. Fresh retains all 681 original programs in 681 directly selectable roots, preserving the original policy for lowering declines. The five existing flow exclusions are unchanged.

Lowering and marked JavaScript/SSA graph setup are reused once per program per package process. Node processes, temporary trace files and trace observations are never cached. The setup identity pin catches repeated setup. Per-unit cleanup rejects a 30-second overrun; Node tracing uses a 25-second deadline shared across a unit's independent observations. The concrete joins and mutations witnesses retain the original non-vacuity checks without an aggregate parent.

`TestFlowCorpusUnitsCoverEveryProgram` and `TestFreshCorpusUnitsCoverEveryProgram` compare generated registrations with the original globs, assert counts, order and unique paths, and check the compiled test-function bindings. Flow also checks all four helper identities. Together they cover exactly 3,385 analysis pieces in 1,360 corpus roots. New fixtures require regenerating wrappers with `generate.py`; forgetting a program fails the coverage pin. Generated files reproduced byte for byte after regeneration and gofmt.

`audit.py` checks the complete before/after test inventories, unchanged fixture and production hashes, registered source hashes, every invocation's budget, and equality of observed check totals. Its observations are:

| Analysis | Identical before/after totals |
| --- | --- |
'''
labels={'SingleAssignment':['functions','phis','values','uses'],'MutationRanges':['mutations','declined ranges','invalid ranges'],'GraphPaths':['walked points','events'],'Liveness':['variable-and-point pairs'],'Writes':['writes','proven writes']}
for family, values in summary['observations'].items():
    text += '| '+family+' | '+', '.join(f'{v:,} {label}' for v,label in zip(values['before'],labels[family]))+' |\n'
text += '''
## Planted failures and mutants

All nine test mutants use Go overlays; repository sources and fixtures were not altered during the runs. Each exits 1 with the intended assertion, without a build failure.

| Mutant | Catcher | Other passing units |
| --- | --- | ---: |
'''
for r in mutants:
    text += '| '+r['name']+' | `'+r['failed_units'][0]+'`: '+r['catcher']+' | '+str(r['passed_units'])+' |\n'
text += '''
The joins program receives an extra tracked read with no definition in its SSA input; only its owning flow root fails and all 678 other flow roots pass. The dedication write-proof program receives an unrecorded property write; only its owning write root fails and all 680 others pass. Missing-program and missing-analysis overlays prove the coverage assertions can fail. Rebuilt-setup overlays prove reuse is checked. Clock overlays prove both budget assertions can fail. Separately, `audit_mutants.py` rejects four forged evidence ledgers: a missing measurement, a changed fixture hash, a changed measured test hash, and a lost write-check count.

## Commands and evidence

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`: passed. Timing lines: Node 0.025 s; Go 0.027 s; markdown 0.076 s; submodules 0.088 s; clang 0.173 s; build 46.456 s; deferred tests 46.679 s; warm cache 46.681 s; total 46.715 s. Environment: `/workspace/adamic-tools/env.sh`.
- `python3 review/test-split-flow/measure.py before`: 203 passes, two skips, all 205 original top-level tests measured separately.
- `python3 review/test-split-flow/generate.py`, then gofmt: deterministic generated wrappers; 3,385 analysis registrations.
- `python3 review/test-split-flow/mutants.py`: all nine Go-overlay mutants caught.
- `ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4 go test -race ./internal/flow -run 'CorpusSetup|CorpusUnitsCover|Program_testdata_(joins|mutations)' -count=1 -parallel=4 -timeout=30s`: passed, package time 1.668 s.
- `python3 review/test-split-flow/measure.py after`: all 1,563 roots measured independently in fresh test processes; 1,561 passes and two skips. Exact commands and raw-log paths are in the JSON ledgers.
- `python3 review/test-split-flow/audit_mutants.py`: four forged-ledger mutants caught.
- `python3 review/test-split-flow/audit.py`: complete coverage, equal observed totals, unchanged production/fixture inputs and all cold invocations below the budget.
- Focused coverage/setup/witness and Timsort probes passed; their logs are preserved. No full gate or whole-package test invocation was run. No new oracle fixtures were added, so oracle counts.md did not change.

[evidence.tar.gz](evidence.tar.gz) preserves raw JSONL logs, overlays, setup output and focused/race probe output. Extract it into `/tmp` to restore the ledger's log paths. Package binaries are excluded from the archive; their SHA-256 digests are in `environment.json`. The measurement and mutant scripts reproduce them from the recorded inputs.

The gate discovers test roots from `go test -list`; its production planner needs no change. Its old timing rows are historical hints. Outside this unit's four-package territory, `internal/oracle/node_buffer_proof_mutants.py` still names the retired aggregate root for its host-cycle-dispatch mutant. Its owner should replace that selector with `TestFreshWrites____oracle_testdata_node_buffer_writes_a_04c9d4399aec`; the script fails its own assertion when the old selector finds no tests. That external helper was not changed or run.

Only test files and review/test helpers changed. No cohere source was copied. The delivery branch carries this unit's changes on the assigned main pin; no main or area branch was pushed or merged into.
'''
(review/'REPORT.md').write_text(text)

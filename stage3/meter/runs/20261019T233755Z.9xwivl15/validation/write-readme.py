#!/usr/bin/env python3
"""Write the run handoff once all paired observations and provenance are complete."""
from html import escape
import json
from pathlib import Path

run = Path(__file__).resolve().parents[1]
paired = json.loads((run/'report.json').read_text())
main, area = (paired['trees'][label] for label in ['main','area'])
completion = json.loads((run/'validation/measurement-completion.json').read_text())
provenance = json.loads((run/'validation/provenance.json').read_text())
assert completion['exit'] == 0
pin = completion['measured_commit']
totals = main['totals']
minutes, seconds = divmod(completion['elapsed_seconds'],60)
def cell(text):
    return escape(str(text)).replace('|','&#124;').replace('\n',' ')
text = f'''Whole program {totals['checker_whole_program']}/79 (Oct 8: 2); own file {totals['checker_own_file']}/79 (56); lowering {totals['lowering']}/79 (0).<br>
tsc entry: {main['tsc_entry']['whole_program_diagnostics']} diagnostics (Oct 8: 320).<br>
Measured main: `{pin}`.<br>
Meter elapsed: {minutes}m {seconds}s including recovery; nproc 5, CPU quota 4.<br>
Not measured: native output/execution, exhaustive lowering, or the complete compiler gate.

# Stage 3 meter on main 295df3bc

Main and the historical area label both measure the same main commit above. The retired area/stage3 branch was not fetched or measured. Each compiler was built independently in its own detached checkout. Both ordinary scores are 2/79 whole program, 56/79 own file, and 2/79 lowering. The lowering passes are `src/compiler/corePublic.ts` and `src/compiler/hostErrors.ts`. Lowering here is the ordinary census's successful IR lowering, without invoking a backend; it is not a native binary claim.

The tsc entry fails with 235 diagnostics on each tree, 85 fewer than Oct 8. Its first diagnostic is `src/compiler/builder.ts:2273:9 TS2375`: the object supplied as `ReusableBuilderProgramState` has an `outSignature` that may be undefined. The first diagnostic is no longer builder.ts:1246:69 TS2345. The 56-file own-checker pass set is exactly the Oct 8 pass set: no additions or losses. This comparison spans all changes on main since Oct 8 and does not isolate one commit's causal effect.

The denominator is verified against upstream: 77 checked-in compiler sources, upstream's generated `diagnosticInformationMap.generated.ts`, and adaptation 47's `hostErrors.ts`. No upstream root was removed. The 78 original/generated upstream roots measure whole program 1/78 and own file 55/78; hostErrors.ts supplies the extra pass in each. Three JSON/config inputs are listed by the ordinary census but fail the extension gate and are excluded from the denominator.

## Commands and pinning

The checked-in `twice-daily.sh` reads `STAGE3_METER_COMPILER`, `STAGE3_METER_LATENT`, and `STAGE3_METER_RUNS`. It reads neither `STAGE3_METER_MAIN_REF` nor `STAGE3_METER_AREA_REF`, and unconditionally fetches retired area/stage3. After `git fetch origin` and cutting `typescript/stage3-meter-295df3bc` from origin/main, the retained [run-local copy](validation/pinned-twice-daily.sh) resolves both supplied SHAs locally. Its [compressed diff](validation/launcher.diff.gz) also records the fixed output directory, local dependency transport, and overlay compatibility adapter. The report's area ref name is the reporter's historical label; its SHA is the main SHA above.

The bounded initial launch, captured by [launch.sh](validation/launch.sh), was:

```sh
run=$PWD/stage3/meter/runs/{run.name}
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
STAGE3_METER_MAIN_REF={pin} \\
STAGE3_METER_AREA_REF={pin} \\
STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error \\
STAGE3_METER_RUNS=$PWD/stage3/meter/runs \\
timeout --kill-after=60s 14400 bash "$run/validation/pinned-twice-daily.sh" > "$run/meter-start.log" 2>&1
```

That launch stopped before measurement on overlay incompatibilities. Main's ordinary command bodies were resumed under `timeout --kill-after=60s 5400 bash -c ...`; they are retained verbatim in [main-ordinary-recovery.sh](validation/main-ordinary-recovery.sh). The remaining commands ran under these bounded controllers:

```sh
timeout --kill-after=60s 14400 bash "$run/validation/recover-run.sh" > "$run/meter-recovery.log" 2>&1
timeout --kill-after=60s 14400 bash "$run/validation/recover-area.sh" > "$run/area-recovery.log" 2>&1
timeout --kill-after=60s 14400 python3 "$run/validation/finalize-run.py" > "$run/validation/finalization.log" 2>&1
```

The first recovery controller was paused while its main latent child continued, permitting area to run independently, as in Oct 8's recovery. The child exited 0 and all 79 records passed `latent_summary` before the paused controller was retired. No compiler was interrupted. Its controller exit 143 is intentional; [main-latent-completion.json](validation/main-latent-completion.json), area-recovery.exit, and [measurement-completion.json](validation/measurement-completion.json) record the successful measurement exits. Both latent commands used `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`. All output went to logs, never through a pipe. Background jobs were checked throughout; the meter stayed below two hours.

## Recovery and limits

- The original snapshot generator rejected `impliedTargetQuery func(*ast.Node) *checker.Type`. This is a test-only query override, assigned only in compiler tests; production leaves it nil. The [run-local adapter](validation/statecopy-adapter.diff.gz) recognizes only this exact field/type and preserves nil, rejecting a populated hook. Unknown function fields still fail. First-error statement lowering does not use full-mode snapshots. Nil preservation, rejection of populated/unknown hooks, and a permissive-hook mutant were exercised.
- The refusal generator rejected new outer `return err` preflights. Its [adapter](validation/refusal-adapter.diff.gz) accepts only the exact `referenceRefusal(module)` and `parallelPreflight(module)` conditionals and records their errors. Unknown preflights and missing child walks still fail before output. Production refusal code is retained by the original AST clone mechanism.
- A cold apply had no adapted-product manifest and no `ADAMIC_TYPESCRIPT_SOURCE`. TypeScript v6.0.3 was cloned at `050880ce59e30b356b686bd3144efe24f875ebc8`; `ADAMIC_TYPESCRIPT_SOURCE=/tmp/badrnrf-inputs/TypeScript` supplied it. The stage3/api lockfile was installed with `npm ci --ignore-scripts --no-audit --no-fund` in scratch storage and linked into each detached checkout. Unmodified apply verified its pinned package digest. Main built the adapted product locally; area restored that verified keyed product into a fresh directory. Initial failure logs are retained under validation/initial-overlay-failure and validation/second-overlay-failure.
- Dependency clones used the verified local checkouts as transport, then checked out each recorded pin: cohere `0183ccf8c57a775f874ca73be1adfc06e9ad1562`, checker TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. The inherited Go flags are `-buildvcs=false -trimpath`, so the binaries do not carry a VCS stamp. Source pin/diff checks, executable hashes, build information, and input manifests are retained in [provenance.json](validation/provenance.json). Both compiler checkouts have unchanged tracked sources. Ordinary compiler options and prelude were not changed. Go build-information logs have trailing whitespace stripped in their readable views; exact raw `.log.gz` copies and [SHA-256 records](validation/buildinfo-log-formatting.json) are retained. Census JSONL and diagnostic bytes are unmodified. The initial literal reason-key equality assertion failed because TypeScript renders absolute adapted-tree paths inside some type names; comparison of the complete finding sites showed zero differences after replacing the known tree prefix. The corrected provenance check normalizes that prefix in reason keys as well. Raw reports and diagnostic streams are unchanged, and the initial assertion log is retained.
- The guarded census is partial: diagnosed bodies and deferred generics may not be examined, and fresh isolated lowering state can differ from ordinary rooted lowering. No backend, native driver, native-output proof, whole compiler gate, or milestone update was run. This version of the meter did not write stage3/progress.json.

## Validation

The current checkout discovers 36 meter tests, whereas Oct 8's retained README reports 46. The prior progress/root/symlink test additions and writer are absent from this main revision; they were not recreated outside the unit's territory.

```sh
timeout 1800 python3 -m unittest discover -s stage3/meter -p '*_test.py' > "$run/validation/tests-default.log" 2>&1
CENSUS_BINARY="$binaries/census" LATENT_CENSUS_BINARY="$binaries/latent-census" ENTRY_CENSUS_BINARY="$binaries/entry-census" \\
  timeout 1800 python3 -m unittest discover -s stage3/meter -p '*_test.py' > "$run/validation/tests-real-binaries.log" 2>&1
```

Here `$binaries=/tmp/adamic-gate/stage3-meter.RXmuND/main-binaries`. The default invocation passes with four optional probes skipped. With all three actual binaries enabled, 34/36 tests pass with no skips. The two failures are original entry fixtures that expect `debugger` or string/number conditions to remain lowering blockers: observed NotYet is 0 instead of positive, and Refused is 1 instead of 3. The real checker imported-error attribution and planted latent-finding probes pass unchanged.

Run-local copies of those two entry probes use stable `var` witnesses; no production test was edited. Both pass, including clean imported reach, unloaded-file exclusion, imported-error blocking, nested rejected-entry measurement, and the first-error-only mutant. Their [exact fixture diff](validation/entry-probes.diff.gz) is retained. The command was:

```sh
PYTHONPATH=$PWD/stage3/meter CENSUS_BINARY="$binaries/census" LATENT_CENSUS_BINARY="$binaries/latent-census" ENTRY_CENSUS_BINARY="$binaries/entry-census" \\
  timeout 300 python3 -m unittest discover -s "$run/validation/current-probes" -p entry_test.py -k CensusTests > "$run/validation/current-entry-probes.log" 2>&1
```

Additional checks, each with retained commands and witness logs:

- `report-mutants.py`: treating whole-program status as own-file status fails the planted imported-diagnostic assertion; omitting Refused reasons fails the combined-reason assertion.
- `METER_SCRIPT_UNDER_TEST=.../compiler-selection-mutant.sh timeout 300 python3 -m unittest discover -s stage3/meter -p compiler_test.py -k test_per_ref_builds_and_runs_each_pinned_compiler`: fails on the actual `compiled with checkout` witness despite claimed main metadata.
- `check-statecopy-adapter.py` and `check-refusal-adapter.py`: the narrow compatibility guards and their malformed/permissive mutants behave as described above.
- `timeout 1200 python3 stage3/census/latent/audit_output_guards.py /tmp/adamic-gate/stage3-meter.RXmuND/main-source /tmp/adamic-gate/stage3-meter.RXmuND/main-binaries/latent-overlay /tmp/badrnrf-output-guard-mutants`: both usable-IR and permissive-loader mutants are caught by their actual output guards.
- `lower-ledger.py` checks the 79-file coverage, own-file flags/diagnostic counts and unique latent totals against the meter report. `ledger-mutants.py` corrupts real own-file status, removes a coverage record while preserving every finding, and drops Refused findings. All three fail their intended assertions. `provenance.py` verifies pinned clean tracked sources, input hashes and paired observations.

## Setup and timing

`export GOPROXY='https://proxy.golang.org|direct'; timeout 5400 bash cloud/setup.sh > /tmp/badrnrf-setup.log 2>&1` succeeded, then `/workspace/adamic-tools/env.sh` was sourced. Node 24.19.0, Go 1.27.1, clang 20.1.8; nproc 5; cgroup cpu.max `400000 100000`. Setup prints cumulative elapsed times:

```text
setup: node ready (0.065s)
setup: go ready (0.094s)
setup: clang ready (0.482s)
setup: markdown dependencies ready (1.042s)
setup: submodules ready (374.654s)
setup: go build ready (523.938s)
setup: build cache warm (524.062s)
setup: done on 5 processors (524.093s)
```

The complete timing lines and host flags are in [setup.log](validation/setup.log). Meter elapsed above starts with the first launch after setup and includes failed preparation/recovery and overlapping paired work. It is an elapsed run duration, not a compiler benchmark.

## Top lowering reasons and owners

These are unique latent sites on a checker-rejected program, in historical first-error mode. Main records NotYet {main['latent_lowering']['totals']['NotYet']}, Refused {main['latent_lowering']['totals']['Refused']}, SkippedDependency {main['latent_lowering']['totals']['SkippedDependency']}, errors {main['latent_lowering']['totals']['error']}, panics {main['latent_lowering']['totals']['panic']}. Area records NotYet {area['latent_lowering']['totals']['NotYet']}, Refused {area['latent_lowering']['totals']['Refused']}, SkippedDependency {area['latent_lowering']['totals']['SkippedDependency']}, errors {area['latent_lowering']['totals']['error']}, panics {area['latent_lowering']['totals']['panic']}. OWNER BLANK rows go to @system_adamic; ownership is the pinned meter owners.json mapping.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
'''
for row in main['latent_lowering']['reason_rows'][:10]:
    text += '| ' + ' | '.join(map(cell,[row['reason'],row['owner'],row['NotYet'],row['Refused'],row['count']])) + ' |\n'
text += '''
Both entry lowering censuses are blocked by the entry's checker diagnostics, as required in first-error mode. The directory latent census does not establish successful lowering. Its two retained main panic sites are moduleSpecifiers.ts:335:1 (nil pointer dereference) and transformers/declarations/diagnostics.ts:1:1 (Node.Text on ComputedPropertyName). Skipped dependency bodies are watch.ts:843:1 and sys.ts:1463:27. Full records are in [latent-diagnostics.json](latent-diagnostics.json) and the compressed JSONL streams.

In the table below, hostErrors.ts has an observed latent NotYet despite passing ordinary rooted lowering. Its directory-wide rejected-program context and isolated unit state differ from its ordinary root program. The ordinary score remains 2/79. Six other own-file passes have no observed latent NotYet/Refused; their imported checker failures still prevent an ordinary lowering attempt.

'''
text += (run/'lower-section.md').read_text()
with (run/'README.md').open('x') as output:
    output.write(text)
print('Created',run/'README.md')

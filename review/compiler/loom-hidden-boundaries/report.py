import csv,json,pathlib,subprocess
root=pathlib.Path.cwd(); evidence=root/'review/compiler/loom-hidden-boundaries'
records=[]
for path in sorted(evidence.glob('*.json')):
    try: record=json.loads(path.read_text())
    except ValueError: continue
    if isinstance(record,dict) and 'command' in record and 'tests' in record:
        record['label']=path.stem
        records.append(record)
with (evidence/'test-seconds.csv').open('w') as stream:
    writer=csv.writer(stream);writer.writerow(['run','test','outcome','seconds'])
    for record in records:
        for test in record['tests']:writer.writerow([record['label'],test['name'],test['action'],test['seconds']])
(evidence/'commands-results.json').write_text(json.dumps(records,indent=2)+'\n')
base=(evidence/'base-main-sha.txt').read_text().strip();source=(evidence/'source-sha.txt').read_text().strip()
def rows(text):
    return {line.split(' | ')[0].removeprefix('| '):line for line in text.splitlines() if line.startswith('| ') and ('testdata/' in line or 'dedication/' in line or 'stage3/' in line)}
old=rows(subprocess.check_output(['git','show',base+':internal/oracle/counts.md'],text=True));new=rows((root/'internal/oracle/counts.md').read_text())
added=[k for k in new if k not in old];changed=[k for k in new if k in old and old[k]!=new[k]]
(evidence/'counts.diff').write_text(subprocess.check_output(['git','diff',base,'--','internal/oracle/counts.md'],text=True))
parts=[f'''Re-cut hidden-boundary compiler fixes and the getters flow harness fix as one squash on main.
Source {source}; base main {base}; delivery SHA is supplied in the final report and git HEAD.
Focused compiler, flow, oracle and getters checks plus changed-package vet passed; command outputs and seconds are recorded below.
Compiler and harness mutants with their actual failing tests are recorded below, alongside two IR semantic mutants and the ABI attribution mutant.
Full packages, full gate, full census, historical replay/meter tools, WASI and performance were not run.

The branch is compiler/loom-hidden-boundaries. No PR or Loom submission was made. The source net change includes historical review artifacts and census templates; those are preserved, not treated as fresh observations. New evidence lives only in this directory. No cohere source was copied and its gitlink did not change.

Conflict resolutions

- internal/flow/flow_test.go had conflicts in the eligibility expression and trailing refusal helper. Kept main's shared lowered-program cache and existing constructor/non-null eligibility, added hidden-boundary eligibility, and kept the source's reason-checked getters table in getters_census_test.go. Removed the duplicate main helper because the source table owns the same eighteen refusals and checks their actual stops. Three runnable getters remain eligible.
- internal/flow/trace_test.go had one conflict in object field snapshots. Kept main's cached immutable trace preparation and chose descriptor.value for data fields and a constant marker for accessors. The auto-merged code-property guard also reads its descriptor. This preserves main's no-getter intent and closes its remaining code-getter read while keeping nested data mutation observations.

Rules and assumptions

Every completed test leaf is recorded in test-seconds.csv, including failures and mutant catches. No added or touched test exceeded 60 seconds, and none was deleted or split. nproc is 5; cpu.max is 400000 100000, a four-CPU quota. Verification uses GOMAXPROCS=4 and -parallel 4. Node is v24.19.0, Go is go1.27.1 and clang is 20.1.8.

The source adds no compiler-consuming package or product, so cloud/fast-gate/compiler-dependencies.json needs no entry. Changed tests have no hard-coded machine/build directory; temporary output names remain local test scratch. Main's export-data product is warmed by a bounded go list before TestCallTargetReaders. Source .ts files are retained as ordinary TypeScript hatch controls paired with .a refusal controls: renaming them would erase the tested distinction. No new Adamic source was authored outside .a. Historical evidence stays at its source paths to preserve the requested net change.

Setup and recovered failures

The initial timeout 600 bash cloud/setup.sh exited 124 during build-cache warming. Its submodules completed at 499.726 seconds and shared cache at 505.355 seconds. The first test attempt could not load cohere/TypeScript/tsc/go.mod while checkout was pending. A subsequent cold Go build reached its outer 180-second limit before any test leaf ran. Both logs are preserved. Setup retry completed successfully; its timing lines follow.

```text
''']
parts.append('\n'.join(line for line in (evidence/'setup-retry.log').read_text().splitlines() if line.startswith('setup:'))+'\n```\n\n')
parts.append('The first counts attempt failed on main\'s missing @types/node 25.3.3 and did not rewrite counts.md. Installed the exact stage3/api/package-lock.json dependencies with npm ci --ignore-scripts. The following successful recorder invocation rewrote the table once with zero difference from the squashed source table. The mutant batch reached its outer 600-second limit after all 38 compiler catches were recorded; resumed only the four unfinished harness mutants under another hard limit. The verification evidence parser initially failed on pretty-printed go list JSON strings; fixed it to accept only object events and resumed without rerunning passed tests.\n\nProduction compiler files changed against main\n\n')
parts.append('\n'.join('- '+p for p in (evidence/'runtime-files.txt').read_text().splitlines())+'\n\nNo internal/native runtime C or header file changed. The delivery also retains the source edits to stage3/census/latent/{full,replay,units}.go.txt and replay/main.go.txt, plus its documentation and cycles status metadata; these historical census paths were not remeasured.\n\nCounts attribution\n\n')
parts.append(f'{len(added)} new fixture rows and {len(changed)} existing rows change against main; no rows were removed. The full diff is counts.diff. Existing allocations, frees, peak and regions stay identical. Restoring only main\'s collections.go, iteration.go and iteration_origin.go reproduces all nine old rows; current sources reproduce all nine new rows. The wrong-ABI mutant must fail every owning count assertion.\n\n')
parts.append('| Existing fixture | Main retains/releases | New retains/releases | Cause |\n|---|---:|---:|---|\n')
causes={'census_optional_values_system.a':'15 unused sys/process receivers','census_optional_values_forms.a':'1 unused object.run receiver','census_optional_values_padding.a':'4 unused literal callback receivers','census_unary_numeric.a':'9 unused coercion receivers','closure_convention_receiver_rest.a':'3 unused Runner.run receivers','host_never_branches.a':'1 unused system.abort receiver','host_void_method.a':'2 unused system.write receivers','user_iterators.a':'47 unused iterator factory/next/return receivers','user_iterators_rest_tdz.a':'factory pair removed; next retain removed before TDZ stop'}
for key in changed:
    before=old[key].split(' | ');after=new[key].split(' | ')
    parts.append(f'| {key} | {before[3]}/{before[4]} | {after[3]}/{after[4]} | {causes.get(pathlib.Path(key).name,"requires review")} |\n')
parts.append('\nNew rows, each attributable to its registered source fixture:\n\n```text\n'+'\n'.join(new[k] for k in added)+'\n```\n\nCommands and observed results\n\nAll test output went directly to a file; no test process was piped. Each JSON below records the exact arguments, exit, wall seconds and every test\'s seconds. Test times include their in-test setup; command times additionally include Go compilation. Source Node, JavaScript backend Node and native sanitizer/leak comparisons are exercised by the oracle and attribution checks.\n\n| Run | Exit | Command seconds |\n|---|---:|---:|\n')
for record in records:
    if not record['label'].startswith('mutant-'):parts.append(f"| {record['label']} | {record['exit']} | {record['seconds']} |\n")
parts.append('\nEvery per-test measurement is in [test-seconds.csv](test-seconds.csv), and exact commands and outcomes are in [commands-results.json](commands-results.json). These include all recorder shards and fixture leaves.\n\nMutants and their actual catchers\n\n| Mutant | Exit | Failing owning tests |\n|---|---:|---|\n')
for record in records:
    if record['label'].startswith('mutant-') or record['label']=='counts-audit-wrong-abi-mutant':
        failures=[test['name'] for test in record['tests'] if test['action']=='fail']
        parts.append(f"| {record['label']} | {record['exit']} | {', '.join(failures)} |\n")
parts.append('\nTestHiddenBoundary04PresentArrayMutant catches returning undefined for present input through Node stdout (0|0 versus 7|0). TestHiddenBoundary04EagerIndexMutant catches eager index evaluation through Node side-effect counts. Both test leaves pass their mutant-detection assertions. The four harness mutant logs additionally require the named getter invocation, lost nested data mutation, or refusal admission marker; compile errors and timeouts do not count as catches. Compiler overlays are freshly generated from the resolved tree using the source patches, not stale full-source overlays.\n\nLimits\n\nOnly focused tests and fixtures were run, plus the required all-row count recorder and changed-package vet. No whole-package test command or full gate ran. Historical full-census, replay and performance claims in inherited evidence are provenance, not current validation. The TDZ attribution witness preserves empty stdout and an explicit refusal; it does not claim equal raw Node and Adamic exit codes or error text.\n')
(evidence/'REPORT.md').write_text(''.join(parts))

import json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip()
rows=[]
for log in sorted(out.glob('*.log')):
 for line in log.read_text().splitlines():
  if not line.startswith('{'):continue
  try:r=json.loads(line)
  except json.JSONDecodeError:continue
  if r.get('Action') in ['pass','fail','skip'] and 'Test' in r:
   rows.append(dict(log=log.name,test=r['Test'],seconds=r.get('Elapsed',0),outcome=r['Action']))
(out/'test-times.json').write_text(json.dumps(rows,indent=2))
(out/'test-times.tsv').write_text('log\ttest\tseconds\toutcome\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']:.3f}\t{r['outcome']}\n" for r in rows))
assert all(r['seconds']<60 for r in rows),[(r['test'],r['seconds']) for r in rows if r['seconds']>=60]
def counts(text):
 return {line.split(' | ')[0][2:]:line for line in text.splitlines() if line.startswith('| ') and len(line.split(' | '))==7 and line.split(' | ')[1].isdigit()}
before=counts(subprocess.check_output(['git','show',base+':internal/oracle/counts.md'],text=True));after=counts(Path('internal/oracle/counts.md').read_text())
assert all(after.get(k)==v for k,v in before.items())
added={k:v for k,v in after.items() if k not in before};assert len(added)==15
(out/'counts-attribution.json').write_text(json.dumps(dict(base=base,added=added,changed_existing=[],removed=[]),indent=2))
mutants=[]
for c in json.loads((out/'mutants/cases.json').read_text()):
 log=out/('mutant-'+c['topic']+'.log');events=[json.loads(l) for l in log.read_text().splitlines() if l.startswith('{')];failed=[r for r in events if r.get('Action')=='fail' and 'Test' in r];assert failed and '[build failed]' not in log.read_text(),c['topic']
 errors=[r['Output'].strip() for r in events if r.get('OutputType')=='error']
 mutants.append(dict(topic=c['topic'],file=c['file'],pattern=c['pattern'],failed_tests=[r['Test'] for r in failed],errors=errors))
(out/'mutants/results.json').write_text(json.dumps(mutants,indent=2))
implementation=subprocess.check_output(['git','diff','--name-only',base,'--','internal'],text=True).splitlines();implementation=[p for p in implementation if p.endswith('.go') and not p.endswith('_test.go')]
report=f'''Built: d04ed5d3's net compiler changes squashed onto current main, with three arrow capture frame checks.
Commits: source d04ed5d3064bbe5981224d52473a63a62bdb5707; base main {base}; final commit SHA is in the delivery message.
Checks: focused lowering and oracle tests, fifteen Node/sanitizer fixtures, seven stage3 status rows, TestCallTargetReaders, changed-package go vet, counts.
Mutants: fourteen source overlays fail their named tests; seven IR output mutations are caught by Node comparisons in both backends.
Not covered: full packages/full gate, WASI, all existing count measurements, and the three dependency-blocked tasks already excluded by the source.

Branch: compiler/loom-chain-followers. No PR or Loom submission. Main stayed at the base above during the final fetch, so no rebase was needed.

The source merge base is 5e33a17b186a8a2218d27b69b21e2de5acc5b750. The requested git merge --squash d04ed5d3 had exactly one conflict: internal/oracle/counts.md, two appended groups at the same insertion point. Preserved main's three getters_census rows and source's seven generic/phantom rows. One successful regeneration placed the rows in registration order, preserving every old row's values.

Runtime files changed against main: none under internal/native/runtime. Compiler implementation files changed:

'''+''.join('- '+p+'\n' for p in implementation)+'''
Counts: exactly fifteen added rows; no old row changed or disappeared. Allocations, frees, retains, releases, peak and regions are listed below. These are the fifteen new oracle fixtures, and each was measured once in the successful regeneration. counts-attribution.json independently compares every old row to main.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
'''+''.join(v+'\n' for v in added.values())+'''
Rules 11 and 12: no measured test was 60 seconds or more, so none needed splitting or deletion. Every new/touched test remains present. The new IR checks are top-level Test functions, each with t.Parallel. Their seconds: TestArrowFieldCaptureFrames 0.070, TestArrowSuperCaptureFrames 0.080, TestArrowStoredCaptureFrames 0.080. TestCallTargetReaders passed in 23.500 test seconds (40.965 command seconds including compilation). Every test outcome and seconds, including nested leaves and selected family wrappers, is in test-times.tsv and test-times.json. Tests ran with GOMAXPROCS=4 and -parallel 4 on cpu.max=400000 100000, nproc=5. Products/tests introduce no build-directory or machine-path literals. No new compiler-consuming package exists, so cloud/fast-gate/compiler-dependencies.json needs no change.

Commands and observed outputs: checks.json contains complete argv, exits, and command wall seconds. All test stdout/stderr went directly to log files; no test output was piped. Tests use -count=1 -timeout 90s, subprocess hard limits 240 seconds, and overall monitored limits 900 seconds. The source's lowering tests and oracle mutants passed. Fifteen fixture leaves passed using TestNativeAgreesWithNode_[0-9]+ with the specific new paths, including Node truth, JS backend, ASan/UBSan and leak checks. Seven status rows passed through TestFixturesAssertions/Cycles/Namespaces, selected by directory and file, retaining original Node observations. go vet passed on ./internal/flow ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle. Changed Go files are gofmt clean.

The initial fixture/count selector combined slash-containing paths inside one alternation, which Go splits into components. That fixture run matched no leaves, and the counts run exited 1 explicitly refusing to write a table. The first stage3 selector also matched no rows. Corrected selectors ran all fifteen fixtures and seven status rows; logs retain the initial empty runs and the actual results separately. Counts were successfully written only once, using the fifteen changed fixture leaves, followed by TestCountsAreRecordedTable. No test crossed 60 seconds or was rerun after crossing it.

Mutants and what caught each:

'''+''.join('- '+m['topic']+': '+', '.join(m['failed_tests'])+'. '+(' | '.join(m['errors'][:2]) if m['errors'] else 'Named test failure recorded in its log.')+'\n' for m in mutants)+'''
The seven in-test IR output mutants are empty-array falsiness, fresh-number-array sharing, wrong generic result, numeric concatenation spelling, boolean spelling, undefined spelling, and null spelling. TestEmptyArrayRuntimeMutants, TestGenericFunctionValueWrongResultMutant, TestScannerConcatenationMutant, and TestScannerPrimitiveConcatenationMutants require clean execution and stdout-only disagreement with Node in both native and JS backends. Their passing test logs are the evidence that each mutant was caught. The two new arrow source mutants are caught by IR frame assertions before any native emission: constructor frame isolation removed yields two receiver arrows instead of one empty and one receiver arrow; receiver capture suppressed yields an empty arrow instead of a receiver arrow. Both super and stored checks fail for the latter. Old source reports' arrow mutants failed at C compilation, which is why three independent lowering checks were added.

Conservative scope assumption: d04ed5d3 is the authoritative requested net change. Its two commits already keep #rh16bxg and #4my3dqg, including phantom-brand representation, while excluding #agwccbw, #p6a00tt and #k881crd for unavailable dependencies. This recut does not import those excluded changes or reinterpret the source. Original evidence remains under review/compiler/chain-followers; new evidence is under review/compiler/loom-chain-followers. There was no code copied from cohere; generic inference references its pinned checker through shim/linkname.

Setup: exported GOPROXY=https://proxy.golang.org|direct, ran timeout 600 bash cloud/setup.sh with log output, and sourced /workspace/adamic-tools/env.sh. Setup succeeded, including its real heap-buffer-overflow sanitizer probe. Timing lines, cumulative seconds: Node ready 0.087; Go ready 0.092; clang ready 0.566; Markdown dependency install step 0.674, dependencies ready 0.919; submodules ready 153.714; shared cache ready 158.860; Go build ready 416.525; test binaries deferred 416.610; build cache warm 416.612; done 416.639. nproc=5; CPU quota=4; Go 1.27.1, clang 20.1.8, Node v24.19.0. Full timing/build flags are in setup.log.

Limitations: the unchanged main counts were preserved rather than remeasured, as this worker runs focused checks and Loom owns the full gate. The source evidence has historical whitespace in patch/log witnesses; git diff --check reports that and a pre-existing source fixture's extra EOF blank line. Those evidence bytes were preserved. The three skipped source tasks remain dependency-blocked as described in the source REPORT.md; this report makes no claim that they are implemented.
'''
(out/'REPORT.md').write_text(report)
print('validated 15 added rows, 14 killed source mutants,',len(rows),'recorded test outcomes; max test seconds',max(r['seconds'] for r in rows))

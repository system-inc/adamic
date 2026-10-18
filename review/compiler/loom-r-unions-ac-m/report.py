import collections,json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
meta=json.loads((out/'base.json').read_text());checks=json.loads((out/'checks.json').read_text());mutants=json.loads((out/'mutant-results.json').read_text());counts=json.loads((out/'counts.json').read_text());predicate=json.loads((out/'predicate-results.json').read_text());events=[];skips=[]
for p in sorted(out.glob('*.log')):
 for line in p.read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if not isinstance(e,dict) or 'Test' not in e or e.get('Action') not in ('pass','fail','skip'):continue
  r=dict(log=p.name,package=e.get('Package'),test=e['Test'],seconds=e.get('Elapsed',0),result=e['Action']);(skips if e['Action']=='skip' else events).append(r)
groups=set()
for e in events:
 parts=e['test'].split('/')
 for n in range(1,len(parts)):groups.add((e['log'],e['package'],'/'.join(parts[:n])))
for e in events:e['leaf']=(e['log'],e['package'],e['test']) not in groups
leaves=[e for e in events if e['leaf']];assert leaves and not [e for e in leaves if e['seconds']>=60]
assert all(c['exit']==(1 if c['label'] in ('lower-all','native-all-fixed','stage1-touched') else 0) for c in checks),checks
assert len(mutants)==5 and sum(bool(m['killed']) for m in mutants)==4
assert mutants[-1]['topic']=='equality-observation' and not mutants[-1]['killed']
assert all(c['exits']['a-check']==(1 if c['label'] in ('wrong-reason','removed-header') else 0) for c in predicate['checks'])
assert subprocess.check_output(['git','show','HEAD:cmd/adamic-meter/main.go'])==Path('cmd/adamic-meter/main.go').read_bytes()
(out/'test-times.json').write_text(json.dumps(events,indent=2));(out/'skips.json').write_text(json.dumps(skips,indent=2));(out/'test-times.tsv').write_text('log\tpackage\ttest\tseconds\tresult\tleaf\n'+''.join(f"{e['log']}\t{e['package']}\t{e['test']}\t{e['seconds']}\t{e['result']}\t{e['leaf']}\n" for e in events))
status=[]
for r in predicate['compiler_observations']['changed-programs']:
 err=r['stderr'];kind='checked' if r['exit']==0 else 'NotYet' if "can't lower" in err and ' yet' in err else 'refused' if 'Adamic 0.1 refuses' in err else 'type error' if ' error TS' in err else 'failed'
 status.append(dict(file=r['command'][-1],status=kind,diagnostic=err.strip(),seconds=r['seconds']))
summary=dict(collections.Counter(r['status'] for r in status));assert summary==dict(checked=17,NotYet=6,refused=1),summary
(out/'a-check-statuses.json').write_text(json.dumps(status,indent=2));(out/'a-check-summary.json').write_text(json.dumps(summary,indent=2))
report=f'''Built: unions recut onto current main with the stale meter diagnostic witness fixed and main's narrow-array refusal preserved.
Commits: source {meta['source']}; base {meta['base']}; delivery SHA is the pushed branch tip.
Commands and outputs: requested package suites, touched stage1 packages, flow/fresh and stage3 consumers, guards, vet, formatting and counts passed; raw commands/results are below.
Mutants: four recorded union mutants caught; equality-observation stopped at an ENOSPC Go build, never reached a test and was not rerun; 21 further Go mutants were not run. Both header mutants caught; the existing IR check-removal oracle passed in the full oracle suite.
Not covered: the stopped equality-observation mutant and 21 unrun Go mutants (mutant-coverage.json), full repository gate, other compiler-reader packages, ahra/Loom submission, opt-in skips, and the source report's earlier typeaware timeout.

'''+(out/'resolution.md').read_text()+'\n\n'
report+='Exact original failure output:\n\n```text\n'+(out/'failure-text.txt').read_text()+'```\n\n'
report+='Validation ran with GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1, -count=1, -parallel 4 and -p 1. Package suites have a 12-minute Go timeout and 840-second outer limit; focused guards/mutants have 90-second Go and bounded process limits. All test output goes directly to logs. Grouping parents are distinguished from leaves. Every measured leaf, including baseline failures and intentionally killed mutants, stayed below 60 seconds. Tests were split when main changed the classification; none was deleted.\n\n'
report+='The full lower suite includes native-agreement program registration; full oracle includes exact runtime-set witnesses and Node/native/JavaScript comparisons under ASan/UBSan/LeakSanitizer. Explicit IR TestCallTargetReaders (-timeout 90s), SortedMap and cmd/adamic-gate parallel-test guards passed. No new Go package, IR-held map, product recipe or runtime implementation is added. Full touched stage1 selector/values and full fixture-consuming flow/fresh suites passed. The full stage3 fixture package passed with no status refresh. Gofmt and source whitespace checks are recorded separately.\n\n'
report+='New meter witnesses: fixture-proof.json records actual source Node and sanitized native output for the admitted boolean-field program, and actual Node plus the expected NotYet diagnostic for the replacement lookup. The original two failing assertions and the complete meter implementation remain byte-identical to main.\n\n'
report+='A-check: compiled every added/modified .a file, including the new meter programs, using the branch compiler and unchanged pinned Gate.aCheck. Totals: 17 checked, 6 NotYet, 1 refused, 0 type errors. Added one matching first-line header to scout_nullable_storage_refused/array_spread.a; preserved its body exactly and found no pinned positional reader. The existing headed non-null witness is the removed-header/wrong-reason control. Both failed the predicate, and restoration passed with unchanged bytes. Exact statuses and real diagnostics are in a-check-statuses.json and predicate-results.json.\n\n'
report+='Setup succeeded with GOPROXY=https://proxy.golang.org|direct and sourced /workspace/adamic-tools/env.sh. Linux house clang 20.1.8; Go 1.27.1; Node 24.19.0. nproc=5, cpu.max=400000 100000 (four CPUs). Used main\'s supported ADAMIC_BUILD_OUTPUT_DIR and ADAMIC_GOCACHE_DIR on the separate temporary volume for new products/cache writes, preserving existing cache contents. This affects the environment only, not product keys or source. Timing lines:\n\n```text\n'+'\n'.join(s for s in (out/'setup.log').read_text().splitlines() if s.startswith('setup:'))+'\n```\n\n'
report+='WASI setup recovery timing lines (supported --wasi-sdk, success):\n\n```text\n'+'\n'.join(s for s in (out/'setup-wasi.log').read_text().splitlines() if s.startswith('setup:'))+'\n```\n\n'
report+='Package commands and results:\n\n'
for c in checks:report+='- '+c['label']+f": exit {c['exit']}, {c['seconds']}s; `"+' '.join(c['command'])+'`.\n'
report+='\nCounts additions (A/F/retain/release/peak/region):\n\n```text\n'+'\n'.join(counts['added'])+'\n```\n\nCounts existing-row changes:\n\n```text\n'+'\n'.join(x['before']+'\n'+x['after'] for x in counts['changed_existing_rows'])+'\n```\n\n'
report+='Completed mutant catchers (test assertions, not build failures or timeouts):\n\n'
for m in [row for row in mutants if row['killed']]:report+='- '+m['topic']+': '+', '.join(e['Test'] for e in m['catchers'])+f"; {m['seconds']}s command.\n"
newtests={'TestUnionClassFieldHasNoDiagnostic','TestNullableObjectSpreadKeepsNarrowerLiteralRefused'}
report+='\nNew test timings (successful observations):\n\n'+'\n'.join(f"- {e['test']}: {e['seconds']}s ({e['log']})." for e in leaves if e['test'] in newtests and e['result']=='pass')+'\n\n'
report+=f"Measured leaves: {len(leaves)}; maximum {max(e['seconds'] for e in leaves)} seconds. Exact per-leaf records, expected-red observations, skips and complete logs are retained.\n"
(out/'REPORT.md').write_text(report);print('Report ready; max leaf',max(e['seconds'] for e in leaves),'a-check',summary)

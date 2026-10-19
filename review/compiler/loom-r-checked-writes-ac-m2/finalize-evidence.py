import pathlib,json,subprocess
out=pathlib.Path('review/compiler/loom-r-checked-writes-ac-m2')
def summarize(name):
 events=[];packages=[]
 for line in(out/name).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(e)
   else:packages.append(e)
 names={e['Test']for e in events};leaves=[e for e in events if not any(n.startswith(e['Test']+'/')for n in names)]
 return dict(log=name,packages=packages,test_events=len(events),max_leaf_seconds=max((e.get('Elapsed',0)for e in leaves),default=0),over_budget=[dict(test=e['Test'],seconds=e.get('Elapsed'))for e in leaves if e.get('Elapsed',0)>=60]),events
initial=json.loads((out/'package-results.json').read_text());remaining=json.loads((out/'remaining-package-results.json').read_text());native,events=summarize('native-final.log');native.update(package='internal/native',selection=None,exit=int((out/'native-final.exit.log').read_text()),command=['timeout','930','go','test','./internal/native','-count=1','-p=1','-parallel=4','-json','-timeout=900s'])
rows=initial[:3]+[native]+remaining
expected=set(json.loads((out/'changed-go-packages.json').read_text()));assert expected<=set(r['package']for r in rows if not r.get('selection'))
assert all(r['exit']==0 and not r['over_budget']for r in rows)
assert all(e.get('Action')=='pass'for r in rows for e in r['packages'])
for name in ['counts-update','runtime-focused','mutants-final','vet','guards','a-check','gcc-portability-final','prior-resolution-final']:
 assert(out/(name+'.exit.log')).read_text().strip()=='0',name
mutants=json.loads((out/'mutant-results.json').read_text());assert len(mutants)==19 and all(r['caught']for r in mutants)
for r in mutants:
 text=(out/(r['name']+'.log')).read_text();assert 'error: 'not in text and 'build failed'not in text and 'test timed out'not in text and '# github.com/'not in text
assert not(out/'gofmt.log').read_text()and not(out/'diff-check.log').read_text()
all_events=[]
for r in rows:
 summary,ev=summarize(r['log']);assert not summary['over_budget'];assert not any(e['Action']=='fail'for e in ev)
 all_events.extend(dict(package=e['Package'],test=e['Test'],action=e['Action'],seconds=e.get('Elapsed',0),log=r['log'])for e in ev)
(out/'all-test-seconds.json').write_text(json.dumps(all_events,indent=2)+'\n');(out/'final-package-results.json').write_text(json.dumps(rows,indent=2)+'\n')
source_tests=subprocess.check_output(['git','diff','origin/main','--','*_test.go'],text=True)
import re
names=re.findall(r'^\+func (Test\w+)\(',source_tests,re.M)
(out/'changed-test-seconds.json').write_text(json.dumps([e for e in all_events if e['test'].split('/')[0]in names],indent=2)+'\n')
summary=dict(base='561aeb4b851ca0820f39440a09dc5fb3fd27ba40',source='2dad6ae36c3ba6aeaeefc636e05f4b2956a6c4f6',all_packages_pass=True,standalone_mutants_caught=19,prior_lowering_mutants_caught=2,exceptions_mutants_caught=5,error_shape_mutant_caught=True,header_mutants_caught=2,counts_existing_changed=95,counts_added=120,maximum_leaf_seconds=max(r['max_leaf_seconds']for r in rows));(out/'verification-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
with(out/'REPORT.md').open('a')as f:
 f.write('Final verification\n\nEach command sources the printed toolchain environment, sets GOMAXPROCS=4 and -parallel=4, and writes all test output directly to its named log. Complete direct packages are run, not a selected-test substitute. Native initial GCC failures and the corrected full rerun are retained separately. All individual completed final leaves stay below 60s; no tests are deleted or skipped to make this change pass.\n\n| Package | Selection | Package seconds | Slowest leaf seconds | Result |\n|---|---|---:|---:|---|\n')
 for r in rows:
  seconds=r['packages'][-1].get('Elapsed');f.write(f"| {r['package']} | {r.get('selection') or 'all tests'} | {seconds} | {r['max_leaf_seconds']} | PASS |\n")
 f.write('\nfinal-package-results.json retains every exact command and outcome; all-test-seconds.json retains all terminal test events; changed-test-seconds.json gives every new/changed declaration its measured seconds. Complete oracle includes the inherited embedded checked-write mutants and holds admitted fixtures to source Node, backend Node and release/sanitized native outcomes.\n\nVet on all ten changed Go packages passes (timeout 180 go vet with each package explicitly named). Gofmt -l over every changed Go file prints nothing; git diff --check origin/main -- cmd internal passes. Mandatory guard command (timeout 120 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run \'^(TestCallTargetReaders.*|TestEveryMapAProgramHoldsIsSorted|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestEveryTestIsParallelOrSaysWhy)$\' -count=1 -parallel=4 -json -timeout=90s) passes. No new compiler-consuming package or runtime feature definition is introduced. No product/test gains a build directory or machine path; the worker-only GCC SDK configuration is evidence, not a repository product change.\n\nAll 19 standalone mutation logs are independently audited: assertions catch every mutant, with no compiler-error or timeout kills counted. The first stale map overlay failure is excluded. Five exceptions.c mutants, the Error-shape swap and both unchanged header-predicate mutants are also caught.\n\nDelivery is one commit with parent 561aeb4b851ca0820f39440a09dc5fb3fd27ba40, the original message retained plus the requested re-cut line and explicit conservative assumption. Push only compiler/loom-r-checked-writes-ac-m2, without force. No ahra, PR or Loom submission. Final response supplies full SHA and remote verification.\n')
msg=out/'commit-message.txt';text=msg.read_text();ending='Co-Authored-By: Ahra <ahra@ahra.ai>\n';assert text.endswith(ending);text=text[:-len(ending)]+'Validation after this re-cut: all ten changed package suites, vet, guards, formatting, a-check and all 19 standalone checked-write mutants pass.\n\n'+ending;msg.write_text(text)
print(json.dumps(summary,indent=2))

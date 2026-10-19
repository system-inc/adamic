import pathlib,json,subprocess,collections
out=pathlib.Path('review/compiler/loom-r-cycle-targets-ac-m2')
def summary(path):
 events=[];packages=[];errors=[]
 for line in path.read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('OutputType')=='error':errors.append(e)
  if e.get('Action')in('pass','fail','skip'):
   (events if e.get('Test')else packages).append(e)
 names={e['Test']for e in events};leaves=[e for e in events if not any(n.startswith(e['Test']+'/')for n in names)]
 return dict(log=path.name,packages=packages,skipped_tests=[e for e in events if e['Action']=='skip'],failed_top_tests=[e['Test']for e in events if e['Action']=='fail'and'/'not in e['Test']],errors=errors,max_leaf_seconds=max((e.get('Elapsed',0)for e in leaves),default=0),over_budget=[e for e in leaves if e.get('Elapsed',0)>=60],new_tests=[e for e in events if e['Test']in ['TestClosureTargetsIgnoreNonCallableArrayInputs','TestClosureTargetsKeepCallableArrayInputsUnknown','TestClosureArraysWithUnrelatedProducersAgreeWithNode']])
paths=[out/x for x in ['parent-fuzz-focused.log','source-fuzz-focused.log','parent-fuzz-full.log','source-fuzz-full.log','final-package-00.log','counts-update-shards.log','guards.log','header-reader-checks.log']if(out/x).exists()]
paths+=sorted(out.glob('final-package-0[1-7].log'))
result=[summary(p)for p in paths];(out/'test-summary.json').write_text(json.dumps(result,indent=2)+'\n')
for r in result: print(r['log'],[(p.get('Action'),p.get('Elapsed'))for p in r['packages']],'maxleaf',r['max_leaf_seconds'],'failed',r['failed_top_tests'],'over',len(r['over_budget']))
assert not any(r['over_budget'] for r in result)
runtime=subprocess.check_output(['git','diff','HEAD','--','internal/native/runtime'],text=True);(out/'runtime-beyond-source.patch').write_text(runtime);assert not runtime

import re
added=re.findall(r'^\+func (Test\w+)\(',subprocess.check_output(['git','diff','HEAD','--','*_test.go'],text=True),re.M)
measured={}
for path in sorted(out.glob('final-package-*.log')):
 for line in path.read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Test')in added and e.get('Action')in('pass','fail'):measured[e['Test']]=dict(test=e['Test'],seconds=e.get('Elapsed',0),status=e['Action'],log=path.name)
(out/'new-test-seconds.json').write_text(json.dumps([measured[n]for n in added if n in measured],indent=2)+'\n')

from pathlib import Path
import json,re,subprocess
p=Path(__file__).resolve().parent;base='67733635b0252eb9d3bb3645c72262c7b18361e1';runs=[];over=[]
for log in sorted(p.glob('*.jsonl')):
 events=[]
 for line in log.read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if e.get('Test'):events.append(e)
 names={e['Test'] for e in events}
 for e in events:
  if e.get('Action') not in ['pass','fail','skip']:continue
  row={'log':log.name,'test':e['Test'],'action':e['Action'],'seconds':e.get('Elapsed',0),'leaf':not any(n.startswith(e['Test']+'/') for n in names)};runs.append(row)
  if row['leaf'] and row['action']!='skip' and row['seconds']>=60:over.append(row)
paths=subprocess.check_output(['git','diff',base,'--name-only','--','*_test.go'],text=True).splitlines()
def functions(s):
 starts=list(re.finditer(r'^func\s+(\w+)\([^\n]*',s,re.M));out={}
 for i,m in enumerate(starts):
  if m[1].startswith('Test'):out[m[1]]=s[m.start():starts[i+1].start() if i+1<len(starts) else len(s)]
 return out
changed=[]
for f in paths:
 if not Path(f).exists():continue
 old=subprocess.run(['git','show',base+':'+f],capture_output=True,text=True)
 a=functions(old.stdout);b=functions(Path(f).read_text())
 for n,text in b.items():
  if text!=a.get(n):changed.append({'file':f,'test':n,'new':n not in a,'measurements':[r for r in runs if r['test']==n or r['test'].startswith(n+'/')]})
(p/'test-time-audit.json.txt').write_text(json.dumps({'over_budget_leaves':over,'new_or_changed_tests':changed},indent=2)+'\n');print('Recorded',len(changed),'new/changed test functions; over-budget leaves',over)

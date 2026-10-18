import json, subprocess
from pathlib import Path
out=Path('review/compiler/loom-slice-2-2')
def rows(text):
 return {line.split('|')[1].strip():line for line in text.splitlines() if line.startswith('| ') and (line.split('|')[1].strip().endswith(('.a','.ts')))}
base=rows(subprocess.check_output(['git','show','origin/main:internal/oracle/counts.md'],text=True));now=rows(Path('internal/oracle/counts.md').read_text())
prior=json.loads(Path('review/compiler/chain-slice-2/counts-attribution.json').read_text());owners={x['path']:x['owner'] for x in prior['rows']}
changes=[dict(path=p,owner=owners.get(p),before=base.get(p),after=now.get(p)) for p in sorted(base.keys()|now.keys()) if base.get(p)!=now.get(p)]
summary=dict(base_rows=len(base),after_rows=len(now),added=sum(x['before'] is None for x in changes),removed=sum(x['after'] is None for x in changes),changed_existing=sum(x['before'] is not None and x['after'] is not None for x in changes),rows=changes)
(out/'counts-attribution.json').write_text(json.dumps(summary,indent=2)+'\n')
assert all(x['owner'] for x in changes), 'Unattributed counts row'
assert summary['changed_existing']==0 and summary['removed']==0, 'Existing counts moved; investigate'
assert set(x['path'] for x in changes)==set(owners), 'Slice fixture counts inventory differs'
seconds=[]
for name in ['oracle','support-partial','flow','fresh','fixtures','registry']:
 for line in (out/(name+'.log')).read_text().splitlines():
  if not line.startswith('{'):continue
  event=json.loads(line)
  if event.get('Test') and event.get('Action') in ['pass','fail','skip']:
   seconds.append(dict(check=name,package=event['Package'],test=event['Test'],result=event['Action'],seconds=event.get('Elapsed')))
(out/'test-seconds.json').write_text(json.dumps(seconds,indent=2)+'\n')
assert all(x['result']=='pass' for x in seconds)
blocked=json.loads((out/'overdue-leaves.json').read_text())
assert all(x['seconds']<60 for x in seconds if '/' not in x['test'] and x['test'] not in blocked and x['package']+'.'+x['test'] not in blocked)
print(json.dumps({k:v for k,v in summary.items() if k!='rows'}))
for name in ['oracle','support-partial','flow','fresh','fixtures','registry']:
 tests=[x for x in seconds if x['check']==name and '/' not in x['test']]
 print(name, len(tests), 'leaves; max seconds', max(x['seconds'] for x in tests))

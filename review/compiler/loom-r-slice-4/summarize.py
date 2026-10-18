import json,subprocess,re
from pathlib import Path
p=Path('review/compiler/loom-r-slice-4'); rows=[]
for f in sorted(p.glob('*.jsonl')):
 for line in f.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if r.get('Test') and r.get('Action') in ('pass','fail','skip'):
   rows.append(dict(log=str(f),package=r.get('Package'),test=r['Test'],result=r['Action'],seconds=r.get('Elapsed',0)))
(p/'test-times.json').write_text(json.dumps(rows,indent=2)+'\n')
assert not [r for r in rows if r['result']=='fail'], [r for r in rows if r['result']=='fail']
assert not [r for r in rows if r['seconds']>=60], [r for r in rows if r['seconds']>=60]
def counts(text):
 d={}
 for line in text.splitlines():
  a=[x.strip() for x in line.strip().strip('|').split('|')]
  if len(a)==7 and a[1].isdigit():d[a[0]]=[int(n) for n in a[1:]]
 return d
base=counts(subprocess.check_output(['git','show','HEAD:internal/oracle/counts.md'],text=True));now=counts(Path('internal/oracle/counts.md').read_text());src=counts(subprocess.check_output(['git','show','2a15510921879ede7590c2272156ca3452cd39ed:internal/oracle/counts.md'],text=True))
changed=[dict(program=k,before=base.get(k),after=now.get(k),matches_source=now.get(k)==src.get(k)) for k in sorted(base.keys()|now.keys()) if base.get(k)!=now.get(k)]
(p/'counts-attribution.json').write_text(json.dumps(changed,indent=2)+'\n')
runtime={name:subprocess.check_output(cmd,text=True) for name,cmd in [('source',['git','diff','2a155109~1','2a155109','--','internal/native/runtime']),('recut',['git','diff','HEAD','--','internal/native/runtime'])]}
assert runtime=={'source':'','recut':''},runtime
(p/'runtime-comparison.json').write_text(json.dumps(runtime,indent=2)+'\n')
print('test events',len(rows),'max',max(rows,key=lambda r:r['seconds']),'counts',len(changed),'added',sum(r['before'] is None for r in changed),'source mismatches',[r for r in changed if not r['matches_source']])

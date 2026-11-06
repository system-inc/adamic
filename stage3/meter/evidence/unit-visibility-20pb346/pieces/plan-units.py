"""Partition remaining full-mode candidates using measured unit costs.

Known wall costs come from preserved baseline LATENT_TIMES streams. Unknown
costs use a conservative source-line-gap proxy. Largest-first bins target 180
seconds of single-run measured cost, with at most 200 candidates per bin.
Independent units are indivisible; the runner enforces a 20-minute wall cap.
"""
from pathlib import Path
import gzip, hashlib, json, math
root=Path('/workspace/scratch/20pb346-2')
evidence=Path(__file__).resolve().parent
source=root/'adapted/src/compiler'
inventory=json.loads((root/'unit-inventory.jsonl').read_text().splitlines()[1])
assert len(inventory)==79
assert sum(len(keys) for keys in inventory.values())==10714
covered=set()
for path in evidence.glob('group-*/selection.json'):
 for f in json.loads(path.read_text())['files']:
  assert str(source/f) not in covered
  covered.add(str(source/f))
costs={}
for label in ('before-entry-cap','before-compiler-default-oom','before-compiler-complete'):
 for line in (root/label/'latent-times.jsonl').read_text().splitlines():
  row=json.loads(line);costs[row['where']]=max(costs.get(row['where'],0),row['wall_seconds'])
plan=[]
for filename,keys in sorted(inventory.items()):
 assert len(keys)==len(set(keys)),filename
 if filename in covered:continue
 weights={}
 for i,key in enumerate(keys):
  where=key.rsplit('|',1)[0]
  line=int(where.rsplit(':',2)[1])
  following=int(keys[i+1].rsplit('|',1)[0].rsplit(':',2)[1]) if i+1<len(keys) else line+30
  weights[key]=costs.get(where,max(0.1,min(120,(following-line)*0.2)))
 count=max(1,math.ceil(sum(weights.values())/180),math.ceil(len(keys)/200))
 bins=[dict(units=[],estimated_seconds=0.0) for _ in range(count)]
 for key in sorted(keys,key=lambda k:(-weights[k],k)):
  target=min((b for b in bins if len(b['units'])<200),key=lambda b:b['estimated_seconds'])
  target['units'].append(key);target['estimated_seconds']+=weights[key]
 order={key:i for i,key in enumerate(keys)}
 for index,group in enumerate(b for b in bins if b['units']):
  group['units'].sort(key=order.__getitem__)
  plan.append(dict(name=Path(filename).relative_to(source).as_posix().replace('/','_').replace('.ts','')+'-%03d'%index,file=filename,source_sha256=hashlib.sha256(Path(filename).read_bytes()).hexdigest(),**group))
actual=[key for group in plan for key in group['units']]
expected=[key for filename,keys in inventory.items() if filename not in covered for key in keys]
assert len(actual)==len(set(actual)) and set(actual)==set(expected)
(evidence/'unit-plan.json').write_text(json.dumps(dict(compiler_base='d4c6b07aca1a7c55785226563c2ad79d552c1adf',file_groups=sorted(covered),inventory_files=79,inventory_units=10714,planned_units=len(actual),pieces=plan),indent=2)+'\n')
with (evidence/'unit-inventory.jsonl.gz').open('wb') as raw:
 with gzip.GzipFile(fileobj=raw,mode='wb',mtime=0) as output:output.write((root/'unit-inventory.jsonl').read_bytes())
print(json.dumps(dict(pieces=len(plan),units=len(actual),files=len(inventory)-len(covered))))

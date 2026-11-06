#!/usr/bin/env python3
"""Freeze disjoint contiguous top-level chunks and their complete nested-unit ownership."""
from collections import Counter
import hashlib,io,json,math,pathlib,tarfile
p=pathlib.Path(__file__).resolve().parent;out=p/'declaration-plan';out.mkdir(exist_ok=True)
def norm(v,root):
 if isinstance(v,str):return v.replace(root+'/','')
 if isinstance(v,list):return [norm(x,root) for x in v]
 if isinstance(v,dict):return {k:norm(x,root) for k,x in v.items()}
 return v
stock=json.loads(pathlib.Path('/workspace/scratch/v3-stock-declarations.json').read_text())
go={}
for side in ['main','head']:
 rows=[json.loads(x) for x in pathlib.Path(f'/workspace/scratch/v3-{side}-inventory.jsonl').read_text().splitlines()]
 assert len(rows)==2 and not rows[0]['checker_rejected']
 go[side]=norm(rows[1],f'/workspace/v3-{side}-adapted-complete')
assert go['main']==go['head'],'compiler inventories disagree'
canonical={r['file']:r for r in go['main']};assert set(canonical)=={r['file'] for r in stock}
for r in stock:
 for key in ['tops','units']:
  expected=[{k:x[k] for k in ['where','top','parent','depth','start','end']} for x in canonical[r['file']][key]]
  # Top inventory has no parent/depth ownership beyond itself.
  assert expected==r[key],(r['file'],key,next(((a,b) for a,b in zip(expected,r[key]) if a!=b),None))
seeds={}
for f in sorted((p/'pieces').glob('*.json')):
 summary=json.loads(f.read_text())
 for record in summary.get('records',[summary]):
  name=record['file'];assert name not in seeds
  with tarfile.open(f.with_suffix('.tar.gz')) as t:
   for side in ['main','head']:
    rows=list(map(json.loads,t.extractfile(side+'-count.jsonl').read().splitlines()))
    row=next(r for r in rows[1:] if r['file'].endswith('/'+name))
    ids=[u['where'].split('-adapted-complete/',1)[1] for u in row['units']]
    assert ids==[u['where'] for u in canonical[name]['units']],(name,'seed unit inventory')
  seeds[name]=str(f.relative_to(p))
large={f'src/compiler/{name}.ts' for name in ['binder','checker','utilities','parser','program','emitter']}
raw=[];small=[]
def make(files,seed=None):
 units=[];atoms=[];total_bytes=0
 for name,tops in files:
  row=canonical[name];wanted={x['where'] for x in tops}
  units.extend(u['where'] for u in row['units'] if u['top'] in wanted)
  atoms.extend(dict(file=name,where=x['where'],start=x['start'],end=x['end']) for x in tops)
  total_bytes+=sum(x['end']-x['start'] for x in tops)
 return dict(files=[dict(file=n,top_units=[t['where'] for t in tops]) for n,tops in files],expected_units=units,top_atoms=atoms,source_bytes=total_bytes,seed=seed,timeout_seconds=600)
for name,row in sorted(canonical.items()):
 tops=row['tops']
 if name in seeds:raw.append(make([(name,tops)],seeds[name]));continue
 limit=min(48,max(1,math.ceil(len(tops)/6))) if name in large else 128
 chunks=[];chunk=[];size=0
 for top in tops:
  w=top['end']-top['start']
  if chunk and (len(chunk)>=limit or size+w>64000):chunks.append(chunk);chunk=[];size=0
  chunk.append(top);size+=w
 if chunk:chunks.append(chunk)
 if len(chunks)==1 and name not in large and len(row['units'])<=80 and sum(t['end']-t['start'] for t in tops)<=32000:
  small.append((name,tops))
 else:raw.extend(make([(name,chunk)]) for chunk in chunks)
bucket=[];units=0;size=0
for name,tops in small:
 n=len(canonical[name]['units']);w=sum(t['end']-t['start'] for t in tops)
 if bucket and (len(bucket)>=4 or units+n>160 or size+w>64000):raw.append(make(bucket));bucket=[];units=0;size=0
 bucket.append((name,tops));units+=n;size+=w
if bucket:raw.append(make(bucket))
raw.sort(key=lambda r:(r['files'][0]['file'],r['top_atoms'][0]['start'] if r['top_atoms'] else -1))
for i,r in enumerate(raw):r['id']=f'd{i:03d}'
plan=dict(schema=1,mode='full',head='9626d0113f08269e519ba216ff7f95ff84af16db',main='2c3f322975b2dee68889e82eab0bbca324d417a9',
 inventory=canonical,files=len(canonical),top_units=sum(len(r['tops']) for r in canonical.values()),units=sum(len(r['units']) for r in canonical.values()),
 timeout_seconds=600,stop_before_utc='2026-11-06T17:04:00+00:00',pieces=raw,
 notes='Frozen shared top-level intervals. Nested full-mode units stay with their original owner. Timed-out intervals are replaced only by disjoint contiguous child intervals, retaining this immutable parent plan.')
index=out/'INDEX.json'
if index.exists():assert json.loads(index.read_text())==plan,'frozen plan changed'
else:index.write_text(json.dumps(plan,indent=2)+'\n')
files={'canonical.json':json.dumps(go['main'],indent=2).encode(),'stock.json':json.dumps(stock,indent=2).encode()}
for side in ['main','head']:files[side+'-inventory.jsonl']=pathlib.Path(f'/workspace/scratch/v3-{side}-inventory.jsonl').read_bytes()
files['SHA256SUMS']=''.join(f'{hashlib.sha256(data).hexdigest()}  {n}\n' for n,data in sorted(files.items())).encode()
with tarfile.open(p/'declaration-inventory.tar.gz','w:gz') as t:
 for n,data in sorted(files.items()):
  i=tarfile.TarInfo(n);i.size=len(data);i.mtime=0;t.addfile(i,io.BytesIO(data))
print(json.dumps({k:plan[k] for k in ['mode','files','top_units','units','timeout_seconds','stop_before_utc']}|dict(pieces=len(raw),seed_files=len(seeds)),indent=2))

"""Freeze disjoint schedules using stock units and historical costs only for sizing."""
import collections,gzip,json,pathlib,subprocess
O=pathlib.Path(__file__).resolve().parent;E=O/'evidence';P=E/'plan';P.mkdir(exist_ok=True)
rows=json.loads((E/'stock-top-level.json').read_text());weights={}
old=gzip.decompress(subprocess.check_output(['git','show','f45740c1:stage3/census/any-members/evidence/base-times.jsonl.gz'])).decode()
for t in map(json.loads,old.splitlines()):weights[t['where']]=t['wall_seconds']

whole=json.loads('null')
raw=gzip.decompress(subprocess.check_output(['git','show','f45740c1:stage3/census/any-members/evidence/base-latent.jsonl.gz'])).decode()
named={}
for row in list(map(json.loads,raw.splitlines()))[1:]:
 for u in row['units']:
  if u.get('name'):named[(row['file'],u['name'])]=weights.get(u['where'],0)
for r in rows:
 for u,d in zip(r['units'],r['details']):
  weights[u]=0 if d['generic'] else named.get((r['file'],d['name']),weights.get(u,10 if r['file']=='src/compiler/utilities.ts' else 1))

assert len(rows)==79
pieces=[]
def add(files,units=None):
 p=dict(schema=1,id=f'p{len(pieces):02d}',files=[dict(file=r['file'],all_units=units is None,units=[] if units is None else units) for r in files])
 p['expected_units']=[u for r in files for u in (r['units'] if units is None else units)]
 p['estimated_unit_seconds']=sum(weights.get(u,1) for u in p['expected_units']);pieces.append(p)
public=next(r for r in rows if r['file']=='src/compiler/utilitiesPublic.ts');add([public])
utility=next(r for r in rows if r['file']=='src/compiler/utilities.ts')
chunk=[];cost=0
for u in utility['units']:
 w=weights.get(u,1)
 if chunk and cost+w>300:add([utility],chunk);chunk=[];cost=0
 chunk.append(u);cost+=w
if chunk:add([utility],chunk)
ordinary=[r for r in rows if r not in [public,utility]]
bins=[]
for r in sorted(ordinary,key=lambda r:-sum(weights.get(u,1) for u in r['units'])):
 cost=sum(weights.get(u,1) for u in r['units'])+20
 choices=[b for b in bins if b['cost']+cost<=600]
 b=min(choices,key=lambda b:b['cost']) if choices else dict(cost=0,files=[])
 if not choices:bins.append(b)
 b['files'].append(r);b['cost']+=cost
for b in bins:add(sorted(b['files'],key=lambda r:r['file']))
actual=collections.Counter(u for p in pieces for u in p['expected_units']);wanted=collections.Counter(u for r in rows for u in r['units']);assert actual==wanted and all(n==1 for n in actual.values())
for p in pieces:(P/(p['id']+'.json')).write_text(json.dumps(p,indent=2)+'\n')
(P/'INDEX.json').write_text(json.dumps(dict(pieces=[p['id'] for p in pieces],files=len(rows),units=len(wanted),timeout_seconds=1500,notes='utilities.ts uses disjoint top-level chunks as in the reused tooling; other files each belong to exactly one file group; historical costs size schedules only'),indent=2)+'\n')
print([(p['id'],len(p['files']),len(p['expected_units']),round(p['estimated_unit_seconds'])) for p in pieces])

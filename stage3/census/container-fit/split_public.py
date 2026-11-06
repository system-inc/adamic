"""Preserve an oversized whole-file attempt and schedule disjoint public chunks."""
import collections,gzip,json,pathlib,subprocess
O=pathlib.Path(__file__).resolve().parent;E=O/'evidence';P=E/'plan'
m=json.loads((E/'latent/base/p00/metrics.json').read_text());assert m['returncode']!=0 and m['timed_out']
index=json.loads((P/'INDEX.json').read_text());assert index['pieces'][0]=='p00'
(P/'INDEX-initial.json').write_text(json.dumps(index,indent=2)+'\n')
rows=json.loads((E/'stock-top-level.json').read_text());row=next(r for r in rows if r['file']=='src/compiler/utilitiesPublic.ts')
with gzip.open(E/'latent/base/p00/times.jsonl.gz','rt') as f:actual={t['where'].replace('/workspace/container-fit-adapted/',''):t['wall_seconds'] for t in map(json.loads,f)}
oldtime={t['where']:t['wall_seconds'] for t in map(json.loads,gzip.decompress(subprocess.check_output(['git','show','f45740c1:stage3/census/any-members/evidence/base-times.jsonl.gz'])).decode().splitlines())}
oldrows=list(map(json.loads,gzip.decompress(subprocess.check_output(['git','show','f45740c1:stage3/census/any-members/evidence/base-latent.jsonl.gz'])).decode().splitlines()))
named={u.get('name'):oldtime.get(u['where'],0) for r in oldrows[1:] if r['file']==row['file'] for u in r['units'] if u.get('name')}
chunks=[];current=[];cost=0
for u,d in zip(row['units'],row['details']):
 estimate=0 if d['generic'] else actual.get(u,max(1,2*named.get(d['name'],10)))
 if current and cost+estimate>300:chunks.append((current,cost));current=[];cost=0
 current.append(u);cost+=estimate
if current:chunks.append((current,cost))
ids=[]
for i,(units,cost) in enumerate(chunks):
 pid=f'q{i:02d}';ids.append(pid)
 p=dict(schema=1,id=pid,files=[dict(file=row['file'],all_units=False,units=units)],expected_units=units,estimated_unit_seconds=cost)
 (P/(pid+'.json')).write_text(json.dumps(p,indent=2)+'\n')
index['pieces']=ids+[p for p in index['pieces'] if p!='p00'];index['public_head_check_pieces']=ids
index['notes']='Whole utilitiesPublic p00 reached the 25 minute ceiling and was committed/pushed. Its incomplete findings and partial timers are excluded from comparison. Public is now disjoint contiguous unit chunks sized with observed timers and conservative historical estimates for unobserved units. Every public chunk receives a head/log byte-identity check. Other original pieces are unchanged.'
wanted=collections.Counter(u for r in rows for u in r['units']);have=collections.Counter(u for pid in index['pieces'] for u in json.loads((P/(pid+'.json')).read_text())['expected_units']);assert have==wanted and all(n==1 for n in have.values())
(P/'INDEX.json').write_text(json.dumps(index,indent=2)+'\n')
(E/'public-budget-stop.json').write_text(json.dumps(dict(stopped_piece='p00',wall_seconds=m['wall_seconds'],completed_timers=len(actual),scheduled_units=len(row['units']),replacement_pieces=ids,reason='whole-file piece exceeded its 1500 second child budget; hard box limit takes precedence; no incomplete observations are paired'),indent=2)+'\n')
print(index['notes']);print([(pid,len(units),round(cost)) for pid,(units,cost) in zip(ids,chunks)])

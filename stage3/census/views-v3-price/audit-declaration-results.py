#!/usr/bin/env python3
"""Recount raw paired chunks, exact ownership and the cumulative coverage ledger."""
from collections import Counter,defaultdict
import hashlib,io,json,pathlib,re,tarfile
from declaration_common import P,PLAN,load_plan,audit_partition,validate
plan=load_plan();state=json.loads((P/'declaration-plan/RESULTS.json').read_text());assert state['plan_sha256']==hashlib.sha256(PLAN.read_bytes()).hexdigest()
def leaves(piece):
 return [leaf for child in state['refinements'][piece['id']] for leaf in leaves(child)] if piece['id'] in state['refinements'] else [piece]
active=[c for root in plan['pieces'] for c in leaves(root)];audit_partition(plan,active)
byid={p['id']:p for p in active};completed=set();stats={side:Counter() for side in ['main','head']};perfile={}
for root in plan['pieces']:
 if root['seed']:
  seed=state['results'][root['id']];assert seed['seed']==root['seed'] and seed['status']=='complete'
  completed.update(root['expected_units'])
  for record in seed['records']:
   perfile[record['file']]=record['sides']
   for side in stats:
    for k,v in record['sides'][side].items():
     if type(v) is int:stats[side][k]+=v
for id,result in state['results'].items():
 if result.get('seed'):continue
 raw=result['raw'];archive=P/raw['bundle']
 with tarfile.open(archive) as bundle:
  body=bundle.extractfile(raw['member']).read()
 assert hashlib.sha256(body).hexdigest()==raw['sha256']
 with tarfile.open(fileobj=io.BytesIO(body)) as t:
  for line in t.extractfile('SHA256SUMS').read().decode().splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(t.extractfile(name).read()).hexdigest()==digest
  selected=json.loads(t.extractfile('selection.json').read());assert selected['id']==id
  if result['status']!='complete':
   assert id in state['refinements'] or result.get('atomic_blocked') or result['status']=='failed'
   continue
  assert id in byid and selected==byid[id]
  assert not (completed & set(selected['expected_units'])),'unit duplicated across completed observations'
  completed.update(selected['expected_units'])
  recounted={f['file']:{'file':f['file'],'sides':{}} for f in selected['files']}
  for side in stats:
   rows=[json.loads(x) for x in t.extractfile(side+'.jsonl').read().splitlines()]
   rows=validate(plan,selected,rows,'/workspace/v3-'+side+'-adapted-complete')
   for row in rows[1:]:
    units=row['units'];blocked={f['unit'] for f in row['findings'] if f['phase']=='lowering'}
    free=[u for u in units if u['status']=='attempted' and u['where'] not in blocked]
    counts=dict(view_elements=sum(u['view_elements'] for u in units),view_element_checks=sum(u['view_element_checks'] for u in units),units=len(units),attempted=sum(u['status'] in ['attempted','panic'] for u in units),panicked_units=sum(u['status']=='panic' for u in units),ir_observed=sum(u['ir_measured'] for u in units),finding_free_attempted=len(free),finding_free_view_elements=sum(u['view_elements'] for u in free),finding_free_view_element_checks=sum(u['view_element_checks'] for u in free))
    recounted[row['file']]['sides'][side]=counts;stats[side].update(counts)
  assert list(recounted.values())==result['records'],'saved totals differ from raw counts'
  delta=json.loads(t.extractfile('pair.json').read());assert delta['schema_version']==1
  for name,digest in delta['source_hashes'].items():assert digest==json.loads((P/'source-manifest.json').read_text())[name]
  for record in recounted.values():
   name=record['file']
   if name not in perfile:perfile[name]=record['sides']
   else:
    for side in stats:
     for k,v in record['sides'][side].items():perfile[name][side][k]=perfile[name][side].get(k,0)+v
assert len(completed)==stats['main']['units']==stats['head']['units']
assert completed<=set(u['where'] for f in plan['inventory'].values() for u in f['units'])
# Verify the rendered totals rather than trusting the renderer's inputs.
report=(P/'REPORT.md').read_text()
for k in ['view_elements','view_element_checks','units','attempted','panicked_units','finding_free_attempted']:
 assert f"| {k} | {stats['main'][k]} | {stats['head'][k]} |" in report,(k,'report total')
for side in stats:
 stats[side]=dict(stats[side])
print(json.dumps(dict(status='passed',planned_top_units=plan['top_units'],planned_full_units=plan['units'],paired_units=len(completed),pending_units=plan['units']-len(completed),active_pieces=len(active),complete_pieces=sum(r['status']=='complete' for r in state['results'].values()),interrupted_pieces=sum(r['status']=='interrupted' for r in state['results'].values()),totals=stats,per_file=perfile),indent=2))

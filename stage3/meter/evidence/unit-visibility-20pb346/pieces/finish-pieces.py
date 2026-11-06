"""Verify complete disjoint candidate coverage and summarize all piece bytes."""
from pathlib import Path
import gzip, hashlib, json
folder=Path(__file__).resolve().parent
plan=json.loads((folder/'unit-plan.json').read_text())
plan_hash=hashlib.sha256((folder/'unit-plan.json').read_bytes()).hexdigest()
execution=json.loads((folder/'execution-plan.json').read_text())
execution_hash=hashlib.sha256((folder/'execution-plan.json').read_bytes()).hexdigest()
assert execution['original_plan_sha256']==plan_hash
inventory=json.loads(gzip.decompress((folder/'unit-inventory.jsonl.gz').read_bytes()).splitlines()[1])
expected={key for keys in inventory.values() for key in keys}
assert len(expected)==sum(map(len,inventory.values()))==10714
seen=set();measured=[]
def inspect(path,wanted,expected_plan_hash=None):
 global seen
 before=gzip.decompress((path/'before-latent.jsonl.gz').read_bytes())
 after=gzip.decompress((path/'after-latent.jsonl.gz').read_bytes())
 assert before==after,path
 verdict=json.loads((path/'identity.json').read_text())
 assert verdict['identical'] and verdict['sha256']==hashlib.sha256(before).hexdigest()
 records=list(map(json.loads,before.splitlines()[1:]))
 assert {row['file'] for row in records}=={key.rsplit('|',1)[0].rsplit(':',2)[0] for key in wanted},path
 keys=[unit['where']+'|'+unit['kind'] for row in records for unit in row['units']]
 assert len(keys)==len(set(keys)) and set(keys)==set(wanted),path
 assert not seen.intersection(keys),path
 seen.update(keys)
 sides={side:json.loads((path/(side+'-result.json')).read_text()) for side in ('before','after')}
 for side,value in sides.items():
  assert value['exit_code']==0 and value['sha256']==verdict['sha256'] and value['bytes']==len(before)
  assert value['wall_seconds']<1800
  if 'plan_sha256' in value:assert value['plan_sha256']==expected_plan_hash
 measured.append(dict(piece=path.name,bytes=len(before),units=len(keys),sha256=verdict['sha256'],wall_seconds={side:v['wall_seconds'] for side,v in sides.items()}))
for path in sorted(folder.glob('group-*')):
 selection=json.loads((path/'selection.json').read_text())
 wanted=[key for name in selection['files'] for key in inventory[str(Path(selection['source'])/name)]]
 inspect(path,wanted)
original={piece['name']:piece for piece in plan['pieces']}
for name in execution['completed_original_pieces']:
 inspect(folder/'unit-results'/name,original[name]['units'],plan_hash)
for piece in execution['pieces']:
 inspect(folder/'unit-results'/piece['name'],piece['units'],execution_hash)
assert seen==expected
summary=dict(identical_pairs=len(measured),files=len(inventory),units=len(seen),file_group_pairs=2,unit_piece_pairs=len(execution['completed_original_pieces'])+len(execution['pieces']),maximum_wall_seconds=max(v for row in measured for v in row['wall_seconds'].values()),plan_sha256=plan_hash,execution_plan_sha256=execution_hash,scope='Each matching selection is byte-identical. Separate pieces are not concatenated into a whole-pass claim.',pieces=measured)
(folder/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k!='pieces'},indent=2))

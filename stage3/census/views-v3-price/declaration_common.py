"""Frozen partition and runtime no-loss/no-move checks."""
from collections import Counter
import hashlib,json,pathlib
P=pathlib.Path(__file__).resolve().parent
PLAN=P/'declaration-plan/INDEX.json'
def load_plan():
 data=json.loads(PLAN.read_text());audit_partition(data,data['pieces']);return data
def expected(data,atoms):
 byfile={}
 for a in atoms:byfile.setdefault(a['file'],set()).add(a['where'])
 return [u['where'] for f in sorted(byfile) for u in data['inventory'][f]['units'] if u['top'] in byfile[f]]
def audit_partition(data,pieces):
 inv=data['inventory'];wanted=[u['where'] for f in inv.values() for u in f['units']]
 tops=[t['where'] for f in inv.values() for t in f['tops']]
 assert Counter(u for p in pieces for u in p['expected_units'])==Counter({u:1 for u in wanted}),'lost or duplicate full unit'
 assert Counter(a['where'] for p in pieces for a in p['top_atoms'])==Counter({u:1 for u in tops}),'lost or duplicate top-level declaration'
 for p in pieces:
  assert p['expected_units']==expected(data,p['top_atoms']),'unit moved out of its frozen top-level owner'
  for f in p['files']:
   allids=[t['where'] for t in inv[f['file']]['tops']];indices=[allids.index(x) for x in f['top_units']]
   assert indices==list(range(min(indices),max(indices)+1)),'noncontiguous interval'
   assert f['top_units']==[a['where'] for a in p['top_atoms'] if a['file']==f['file']]
 for name,row in inv.items():
  got=[a['where'] for p in pieces for a in p['top_atoms'] if a['file']==name]
  assert got==[t['where'] for t in row['tops']],'top-level unit moved in plan order'
def normalized(value,root):
 if isinstance(value,str):return value.replace(root+'/','')
 if isinstance(value,list):return [normalized(v,root) for v in value]
 if isinstance(value,dict):return {k:normalized(v,root) for k,v in value.items()}
 return value
def validate(data,piece,rows,root):
 rows=normalized(rows,root)
 assert rows[0]['latent_mode']=='full' and not rows[0]['checker_rejected']
 assert Counter(r['file'] for r in rows[1:])==Counter(f['file'] for f in piece['files']),'returned file inventory'
 assert Counter(u['where'] for r in rows[1:] for u in r['units'])==Counter({u:1 for u in piece['expected_units']}),'unit lost, duplicated or moved across piece'
 for r in rows[1:]:
  inventory={u['where']:u for u in data['inventory'][r['file']]['units']}
  wanted=[u for u in piece['expected_units'] if u.rsplit(':',2)[0]==r['file']]
  assert [u['where'] for u in r['units']]==wanted,'unit order moved'
  for u in r['units']:
   i=inventory[u['where']]
   for key in ['top','parent','depth','kind']:
    assert u.get(key,'' if key in ['top','parent'] else 0)==i[key],('unit metadata moved',u['where'],key)
  for f in r['findings']:
   if f['phase']=='lowering':assert f['unit'] in set(wanted),'finding moved to an unscheduled unit'
 return rows
def split_piece(data,piece):
 atoms=piece['top_atoms'];assert len(atoms)>1,'atomic top-level declaration cannot be subdivided without changing its meaning'
 mid=len(atoms)//2;children=[]
 for suffix,part in zip(['a','b'],[atoms[:mid],atoms[mid:]]):
  files=[]
  for a in part:
   if not files or files[-1]['file']!=a['file']:files.append(dict(file=a['file'],top_units=[]))
   files[-1]['top_units'].append(a['where'])
  children.append(dict(id=piece['id']+'.'+suffix,parent=piece['id'],files=files,top_atoms=part,expected_units=expected(data,part),source_bytes=sum(a['end']-a['start'] for a in part),seed=None,timeout_seconds=piece['timeout_seconds']))
 assert [a for c in children for a in c['top_atoms']]==atoms
 assert Counter(u for c in children for u in c['expected_units'])==Counter({u:1 for u in piece['expected_units']})
 return children

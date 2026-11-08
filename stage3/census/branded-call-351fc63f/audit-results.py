"""Independent raw recount, frozen-source and completed-stream checks."""
from collections import Counter
import hashlib,json,sys
from pathlib import Path
out=Path(__file__).resolve().parent
partial='--partial' in sys.argv
results={'partial_allowed':partial,'source_graphs':{},'entry_pairings':{},'core_pairings':{},'focused_pairings':{}}

def load(name):return [json.loads(s) for s in (out/(name+'.jsonl')).read_text().splitlines()]
def ready(name):
 p=out/(name+'.status.json')
 return p.exists() and json.loads(p.read_text())['exit']==0

def view(rows):
 findings={};units={};first={}
 for row in rows[1:]:
  for u in row['units']:units[u['where']]=u
  for f in row.get('findings') or []:
   findings.setdefault((f['where'],f['kind'],f['reason'],f['text']),f)
   if f['phase']=='lowering':first.setdefault(f['unit'],f)
 return findings,units,first

def recount(a,b,d):
 bf,bu,bs=view(a);hf,hu,hs=view(b)
 assert a[0]==b[0] and bu.keys()==hu.keys()
 assert dict(Counter(f['kind'] for f in bf.values()))==d['raw_findings']['totals']['before']
 assert dict(Counter(f['kind'] for f in hf.values()))==d['raw_findings']['totals']['after']
 stop=lambda x:(x['where'],x['kind'],x['reason']) if x else None
 changed=[u for u in bu if stop(bs.get(u))!=stop(hs.get(u))]
 assert len(changed)==len(d['raw_findings']['unit_first_stop_delta'])
 newly_blocked=[u for u in changed if u not in bs and u in hs]
 assert len(newly_blocked)==sum(r['type']=='previously_finding_free_unit' for r in d['raw_findings']['regressions'])
 old_panics={(f['where'],f['reason']) for f in bf.values() if f['kind']=='panic'}
 new_panics=[f for f in hf.values() if f['kind']=='panic' and (f['where'],f['reason']) not in old_panics]
 assert len(new_panics)==sum(r['type']=='new_panic' for r in d['raw_findings']['regressions'])
 return dict(parent_events=sum(len(r.get('findings') or []) for r in a[1:]),head_events=sum(len(r.get('findings') or []) for r in b[1:]),raw_before=len(bf),raw_after=len(hf),units=len(bu),changed_first_stops=len(changed),newly_blocked_units=len(newly_blocked),new_panics=len(new_panics))

for label in ['A','B']:
 tree=Path('/tmp/branded-'+label)
 expected=json.loads((out/(label+'-source-hashes.json')).read_text())
 assert len(expected)==81
 for file,digest in expected.items():assert hashlib.sha256((tree/file).read_bytes()).hexdigest()==digest,file
 results['source_graphs'][label]=dict(files=len(expected),all_hashes_unchanged=True)
 for scope in ['entry','core']:
  base=f'{label}-parent-{scope}';head=f'{label}-head-{scope}'
  if not ready(base) or not ready(head):
   assert partial,(label,scope,'incomplete')
   results[scope+'_pairings'][label]={'complete':False};continue
  a,b=load(base),load(head)
  if scope=='entry':
   for rows in [a,b]:
    assert rows[0]['status']=='measurement' and rows[0]['latent_mode']=='first-error' and not rows[0]['diagnostics']
    assert len(rows)==82 and len(rows[0]['sources'])==81
    actual={str(Path(f).relative_to(tree)) for f in rows[0]['sources']}
    assert actual==set(expected)
    assert {r['file'] for r in rows[1:]}==set(rows[0]['sources'])
  else:
   assert a==b
   for rows in [a,b]:assert len(rows)==2 and rows[0]['latent_mode']=='full' and not rows[0]['diagnostics'] and rows[1]['file']==str(tree/'src/compiler/core.ts')
  d=json.loads((out/f'{label}-{scope}-pair.json').read_text())
  results[scope+'_pairings'][label]=dict(complete=True,**recount(a,b,d))
 for i in range(1,5):
  base=f'{label}-parent-focus-{i}';head=f'{label}-head-focus-{i}'
  assert ready(base) and ready(head)
  a,b=load(base),load(head)
  for rows in [a,b]:assert len(rows)==2 and len(rows[1]['units'])==1 and not rows[0]['diagnostics']
  d=json.loads((out/f'{label}-focus-{i}-pair.json').read_text())
  results['focused_pairings'][f'{label}-{i}']=recount(a,b,d)
assert ready('A-parent-serializer') and ready('A-head-serializer')
a,b=load('A-parent-serializer'),load('A-head-serializer')
for rows in [a,b]:assert len(rows)==2 and len(rows[1]['units'])==1 and not rows[0]['diagnostics']
d=json.loads((out/'A-serializer-pair.json').read_text())
results['focused_pairings']['A-serializer']=recount(a,b,d)
for side,pin in [('parent','b5c0aa648a7a35b2f17899895d009eabd4147000'),('head','351fc63fa2573036e194656fb92c35fe03858afe')]:
 tree=Path('/tmp/branded-report' if side=='parent' else '/tmp/branded-head')
 p=json.loads((out/'provenance.json').read_text())[side]
 assert p['sha']==pin
 for file,digest in p['source_hashes'].items():assert hashlib.sha256((tree/file).read_bytes()).hexdigest()==digest,(side,file)
assert not list(out.rglob('*.go'))
report=(out/'REPORT.md').read_text();assert '\u2013' not in report and '\u2014' not in report
(out/'audit.json').write_text(json.dumps(results,indent=2)+'\n')
print('audit passed; entry complete='+str(all(r['complete'] for r in results['entry_pairings'].values())),flush=True)

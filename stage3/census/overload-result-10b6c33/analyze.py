import json,pathlib,re,collections,hashlib
out=pathlib.Path('/workspace/overload-head/stage3/census/overload-result-10b6c33')
def load(label):
 return [json.loads(s) for s in pathlib.Path('/workspace/overload-build-'+label+'/latent.jsonl').read_text().splitlines()]
b,h=load('parent'),load('head')
assert b[0]==h[0]
assert b[0]['latent_mode']=='first-error' and not b[0]['checker_rejected']
assert b[-1]['status']==h[-1]['status']=='relation_report'
def view(rows):
 units={u['where']:u for row in rows for u in row.get('units') or []}
 first={}
 for row in rows:
  for f in row.get('findings') or []:
   if f['phase']=='lowering' and f['unit'] in units: first.setdefault(f['unit'],f)
 return units,first
bu,bf=view(b);hu,hf=view(h)
def stop(f): return (f['kind'],f['where'],f['reason']) if f else None
names=set('nodeCanBeDecorated visitFunctionBody getThisContainer getSuperContainer transformAsyncFunctionBody transformFunctionBody skipPartiallyEmittedExpressions skipParentheses skipOuterExpressions parenthesizeConciseBodyOfArrowFunction'.split())
all_over={w:f for w,f in bf.items() if re.search(r'overload \d+ of .+ result .+ cannot be served by implementation result ',f['reason'])}
targets={};other={}
for w,f in all_over.items():
 name=re.search(r'overload \d+ of (.+?) result ',f['reason']).group(1)
 (targets if name in names else other)[w]=f
print('overload units',len(all_over),'targets',len(targets),'other',len(other))
rows=[]
for w,f in targets.items():
 after=hf.get(w)
 rows.append(dict(unit=bu[w],base=f,head=after,overload_stop_cleared=not (after and after['reason']==f['reason']),lowered=after is None))
others=[dict(unit=bu[w],base=f,head=hf.get(w),unchanged=f==hf.get(w)) for w,f in other.items()]
summary=dict(header=b[0],unit_count=len(bu),attempted_before=sum(u['status'] in ('attempted','panic') for u in bu.values()),attempted_after=sum(u['status'] in ('attempted','panic') for u in hu.values()),lowered_before=sum(u['status']=='attempted' and w not in bf for w,u in bu.items()),lowered_after=sum(u['status']=='attempted' and w not in hf for w,u in hu.items()),first_stop_changes=[dict(unit=bu[w],base=bf.get(w),head=hf.get(w)) for w in bu if stop(bf.get(w))!=stop(hf.get(w))],overload_before=len(all_over),overload_after=sum(bool(re.search(r'overload \d+ of .+ result .+ cannot be served by implementation result ',f['reason'])) for f in hf.values()),targets=rows,other_overload=others)
(out/'unit-analysis.json').write_text(json.dumps(summary,indent=2)+'\n')
def short(w):return w.replace('/workspace/overload-adapted/','')
def esc(s):return str(s).replace('|','\\|').replace('\n','<br>')
def table(items,other=False):
 lines=['| Unit name | Unit location | Overload-result stop | Next first stop |','| --- | --- | --- | --- |']
 for r in sorted(items,key=lambda r:(r['unit'].get('name',''),r['unit']['where'])):
  u=r['unit'];f=r['head']
  status=('unchanged' if r['unchanged'] else 'MOVED') if other else ('cleared' if r['overload_stop_cleared'] else 'remains')
  nextstop='Lowered' if f is None else f"{f['kind']}: {f['reason']} at {short(f['where'])}"
  lines.append('| '+' | '.join(esc(x) for x in [u.get('name',u['kind']),short(u['where']),status,nextstop])+' |')
 return '\n'.join(lines)
(out/'43-units.md').write_text(table(rows)+'\n')
(out/'67-units.md').write_text(table(others,True)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ['header','first_stop_changes','targets','other_overload']},indent=2))
print('target lower',sum(r['lowered'] for r in rows),'other moved',sum(not r['unchanged'] for r in others))

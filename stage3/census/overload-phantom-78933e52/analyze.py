import pathlib,json,re
from merge import load
out=pathlib.Path(__file__).resolve().parent
b=load('/workspace/oph-build-base/latent.jsonl');h=load('/workspace/oph-build-head/latent.jsonl')
assert b[0]==h[0] and b[0]['latent_mode']=='first-error' and not b[0]['checker_rejected']
def view(rows):
 units={u['where']:u for r in rows for u in r.get('units',[])};first={}
 for r in rows:
  for f in r.get('findings',[]):
   if f['phase'] in ('lowering','deferred') and f['unit'] in units:first.setdefault(f['unit'],f)
 return units,first
bu,bf=view(b);hu,hf=view(h);assert bu.keys()==hu.keys()
def isresult(f):return bool(f and re.search(r'overload \d+ of .+ result .+ cannot be served by implementation result ',f['reason']))
def signature(f):return tuple(f[k] for k in ('kind','where','reason')) if f else None
def row(w):
 before,after=bf.get(w),hf.get(w)
 return dict(unit=bu[w],base=before,head=after,newly_admitted=bool(before and not isresult(after)),head_invariant_mutable=bool(after and 'adamic/invariant-mutable' in after['text']))
def targets(name):return [row(w) for w,f in bf.items() if re.match(r'overload \d+ of '+name+r' result ',f['reason'])]
path=targets('getPathComponents');supported=targets('getSupportedExtensions');assert [len(path),len(supported)]==[11,5],[len(path),len(supported)]
summary=dict(header=b[0],units=len(bu),attempted_base=sum(u['status'] in ('attempted','panic') for u in bu.values()),attempted_head=sum(u['status'] in ('attempted','panic') for u in hu.values()),no_stop_base=sum(u['status']=='attempted' and w not in bf for w,u in bu.items()),no_stop_head=sum(u['status']=='attempted' and w not in hf for w,u in hu.items()),overload_result_base=sum(isresult(f) for f in bf.values()),overload_result_head=sum(isresult(f) for f in hf.values()),getPathComponents=path,getSupportedExtensions=supported,first_stop_changes=[row(w) for w in bu if signature(bf.get(w))!=signature(hf.get(w))])
(out/'unit-analysis.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ['header','getPathComponents','getSupportedExtensions','first_stop_changes']},indent=2))
for name in ['getPathComponents','getSupportedExtensions']:
 print(name,'newly admitted',sum(r['newly_admitted'] for r in summary[name]),'invariant mutable',sum(r['head_invariant_mutable'] for r in summary[name]))
 for r in summary[name]:print(r['unit'].get('name'),r['unit']['where'],'=>',signature(r['head']))

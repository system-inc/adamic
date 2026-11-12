import pathlib,json,collections,re,sys
from merge import load
out=pathlib.Path(__file__).resolve().parent
b=load('/workspace/op2-build-base/latent.jsonl');h=load('/workspace/op2-build-head/latent.jsonl')
assert b[0]==h[0] and b[0]['latent_mode']=='first-error' and not b[0]['checker_rejected']
def rel(w):return 'src/'+w.split('/src/',1)[1] if '/src/' in w else w
def view(rows):
 units={u['where']:u for r in rows for u in r.get('units',[])};first={}
 for r in rows:
  for f in r.get('findings',[]):
   if f['phase'] in ('lowering','deferred') and f['unit'] in units:first.setdefault(f['unit'],f)
 return units,first
bu,bf=view(b);hu,hf=view(h);assert bu.keys()==hu.keys()
def signature(f):return (f['kind'],f['where'],f['reason']) if f else None
def row(w):return dict(unit=bu[w],base=bf.get(w),head=hf.get(w))
def isresult(f):return bool(f and re.search(r'overload \d+ of .+ result .+ cannot be served by implementation result ',f['reason']))
def targets(name):return [row(w) for w,f in bf.items() if re.match(r'overload \d+ of '+name+r' result ',f['reason'])]
concatenate=targets('concatenate');supported=targets('getSupportedExtensions');write=targets('writeTokenText')
assert [len(concatenate),len(supported),len(write)]==[5,5,1],[len(concatenate),len(supported),len(write)]
counts=collections.Counter();bykey={}
for w,u in bu.items():
 k=(rel(w).rsplit(':',2)[0],u['kind'],u.get('name',''));ordinal=counts[k];counts[k]+=1;bykey[(*k,ordinal)]=w
cohort=[]
for old in json.loads((out/'historical-43.json').read_text())['targets']:
 k=tuple(old['mapping_key']);w=bykey[k];r=row(w)
 r.update(mapping_key=k,original_overload_stop=old['original_overload_stop'],result_stop_clear_base=not isresult(bf.get(w)),result_stop_clear_head=not isresult(hf.get(w)))
 cohort.append(r)
assert len(cohort)==43
changes=[row(w) for w in bu if signature(bf.get(w))!=signature(hf.get(w))]
summary=dict(header=b[0],units=len(bu),attempted_base=sum(u['status'] in ('attempted','panic') for u in bu.values()),attempted_head=sum(u['status'] in ('attempted','panic') for u in hu.values()),no_stop_base=sum(u['status']=='attempted' and w not in bf for w,u in bu.items()),no_stop_head=sum(u['status']=='attempted' and w not in hf for w,u in hu.items()),overload_result_base=sum(isresult(f) for f in bf.values()),overload_result_head=sum(isresult(f) for f in hf.values()),concatenate=concatenate,getSupportedExtensions=supported,writeTokenText=write,historical_43=cohort,first_stop_changes=changes)
(out/'unit-analysis.json').write_text(json.dumps(summary,indent=2)+'\n')
(out/'concatenate-request.json').write_text(json.dumps(dict(entry=b[0]['roots'][0],sources=b[0]['sources'],rows=concatenate),indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ['header','concatenate','getSupportedExtensions','writeTokenText','historical_43','first_stop_changes']},indent=2))
print('changed first stops',len(changes))
for k in ['concatenate','getSupportedExtensions','writeTokenText']:
 for r in summary[k]:print(k,rel(r['unit']['where']),r['unit'].get('name'),'=>',signature(r['head']))

import json, pathlib, re, hashlib, collections
from functools import lru_cache
ROOT=pathlib.Path('/workspace/scratch'); tree=ROOT/'adapted'
reason='an object refinement using an open numeric enum as a literal tag'
def read(p): return [json.loads(s) for s in p.read_text().splitlines()]
def norm(v):
 if isinstance(v,str): return v.replace(str(tree)+'/', '')
 if isinstance(v,list): return [norm(x) for x in v]
 if isinstance(v,dict): return {k:norm(x) for k,x in v.items()}
 return v
sources={}
@lru_cache(maxsize=None)
def identity(where):
 m=re.fullmatch(r'(.*):(\d+):(\d+)',where)
 if not m: return where
 f,line,col=m.groups(); p=tree/f
 if not p.is_file(): return where
 if f not in sources:
  b=p.read_bytes(); lines=b.split(b'\n'); starts=[0]
  for line_bytes in lines[:-1]: starts.append(starts[-1]+len(line_bytes)+1)
  sources[f]=(b,hashlib.sha256(b).hexdigest(),lines,starts)
 b,sha,lines,starts=sources[f]
 assert 0<int(line)<=len(lines), where
 prefix=lines[int(line)-1].decode('utf-8').encode('utf-16-le')[:(int(col)-1)*2].decode('utf-16-le')
 offset=starts[int(line)-1]+len(prefix.encode('utf-8'))
 return (f,sha,offset)
def signature(f): return (f['kind'],f.get('reason',''),f.get('text',''))
def load(side):
 records=norm(read(ROOT/f'{side}-results/latent.jsonl')); findings={}; sites=collections.defaultdict(list); units={}; stops=collections.defaultdict(list)
 for r in records[1:]:
  for u in r['units']: units[identity(u['where'])]=u
  for f in r.get('findings') or []:
   k=(f['kind'],identity(f['where']),f['reason'],f['text'])
   findings[k]=f
   if f['phase']=='lowering':
    stops[identity(f['unit'])].append(f)
 for f in findings.values(): sites[identity(f['where'])].append(f)
 return records,findings,sites,units,stops
B=load('base'); H=load('head')
assert B[0][0]==H[0][0]
assert B[3].keys()==H[3].keys(), 'unit identities differ'
for k in B[3]:
 assert {a:b for a,b in B[3][k].items() if a not in ('status','instantiated')}=={a:b for a,b in H[3][k].items() if a not in ('status','instantiated')}, 'unit eligibility differs'
deltas=[]
for k in sorted(B[2].keys()|H[2].keys(),key=str):
 old=B[2].get(k,[]); new=H[2].get(k,[])
 if {signature(f) for f in old}=={signature(f) for f in new}: continue
 deltas.append(dict(identity=k,where=(old or new)[0]['where'],base=old,head=new))
openbase=[f for f in B[1].values() if f['reason']==reason]; openhead=[f for f in H[1].values() if f['reason']==reason]
open_delta=[]; outcomes=collections.Counter()
for f in openbase:
 new=H[2].get(identity(f['where']),[])
 if any(x['reason']==reason for x in new): outcome='unchanged'
 elif not new: outcome='no finding'
 else: outcome='; '.join(sorted(set(x['reason'] for x in new)))
 outcomes[outcome]+=1; open_delta.append(dict(identity=identity(f['where']),where=f['where'],base=f,head=new,outcome=outcome))
moved=[]; regress=[]
for k in B[3]:
 old=B[4].get(k,[])[:1]; new=H[4].get(k,[])[:1]
 if [(identity(f['where']),signature(f)) for f in old]!=[(identity(f['where']),signature(f)) for f in new]:
  row=dict(identity=k,unit=B[3][k],base=old,head=new); moved.append(row)
  if (not old and new) or (not any(f['kind']=='panic' for f in old) and any(f['kind']=='panic' for f in new)): regress.append(row)
site_regress=[d for d in deltas if (not d['base'] and d['head']) or (not any(f['kind']=='panic' for f in d['base']) and any(f['kind']=='panic' for f in d['head']))]
def times(side):
 rows=norm(read(ROOT/f'{side}-results/times.jsonl')); keyed={identity(r['where']):r for r in rows}; assert len(keyed)==len(rows), 'duplicate timed unit'; return keyed
bt,ht=times('base'),times('head'); assert bt.keys()==ht.keys()
timing=[]
for k in bt:
 a,b=bt[k],ht[k]; timing.append(dict(identity=k,where=a['where'],base_wall=a['wall_seconds'],head_wall=b['wall_seconds'],delta_wall=b['wall_seconds']-a['wall_seconds'],base_cpu=a['cpu_seconds'],head_cpu=b['cpu_seconds'],delta_cpu=b['cpu_seconds']-a['cpu_seconds']))
summary=dict(base_open=len(openbase),head_open=len(openhead),open_outcomes=dict(outcomes),base_totals=dict(collections.Counter(f['kind'] for f in B[1].values())),head_totals=dict(collections.Counter(f['kind'] for f in H[1].values())),moved_units=len(moved),unit_regressions=len(regress),newly_observed_sites=len(site_regress),timing={s:{key:sum(r[key] for r in t.values()) for key in ['wall_seconds','cpu_seconds']} for s,t in [('base',bt),('head',ht)]})
def first_observed(data):
 stops={}
 for row in data[0][1:]:
  for f in row.get('findings') or []: stops.setdefault(identity(f['unit']), f)
 return stops
bo,ho=first_observed(B),first_observed(H)
observed=[]
for k in sorted(bo.keys()|ho.keys(),key=str):
 old=[bo[k]] if k in bo else []; new=[ho[k]] if k in ho else []
 if [(identity(f['where']),signature(f)) for f in old]!=[(identity(f['where']),signature(f)) for f in new]:
  observed.append(dict(identity=k,unit=B[3].get(k,H[3].get(k,{'where':(old or new)[0]['unit']})),base=old,head=new))
summary['first_observed_stop_changes']=len(observed)
summary['site_regressions']=len(site_regress)
new_open=[f for f in openhead if not any(x['reason']==reason for x in B[2].get(identity(f['where']),[]))]
summary['new_open_findings']=len(new_open)
summary['unit_status']={side:dict(collections.Counter(u['status'] for u in data[3].values())) for side,data in [('base',B),('head',H)]}
summary['instantiation_witness_changes']=sum(B[3][k].get('instantiated')!=H[3][k].get('instantiated') for k in B[3])
result=dict(summary=summary,new_open_findings=new_open,site_regressions=site_regress,first_observed_stop_changes=observed,sites=deltas,base_open_sites=open_delta,moved_units=moved,unit_regressions=regress,newly_observed_sites=site_regress,unit_times=sorted(timing,key=lambda x:abs(x['delta_wall']),reverse=True),source_hashes={f:s[1] for f,s in sources.items()})
(ROOT/'comparison.json').write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(summary,indent=2))

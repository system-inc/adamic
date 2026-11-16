from pathlib import Path
import json,collections,time
r=Path('/workspace/scratch/cast-tags');e=r/'evidence'
while not (e/'census-status.json').exists():time.sleep(5)
def read(side,stem):
 rows=[json.loads(s) for s in (e/(side+'-'+stem+'.jsonl')).read_text().splitlines()]
 assert rows[0]['status']=='measurement' and rows[0]['module_format']=='CommonJSBundle' and rows[0]['latent_mode']=='full'
 assert rows[-1]['status']=='relation_report'
 assert len(rows)==3 and len(rows[0]['diagnostics'])==0
 def normal(x):
  if isinstance(x,str):return x.replace(str(r/(side+'-tree')),'<tree>')
  if isinstance(x,dict):return {k:normal(v) for k,v in x.items() if k not in ['start','end','site_start','site_end']}
  if isinstance(x,list):return [normal(y) for y in x]
  return x
 fs=[normal(f) for row in rows for f in row.get('findings',[])]
 return rows,fs
out={}
for stem in ['generators']:
 _,a=read('main',stem);_,b=read('c',stem)
 encode=lambda x:json.dumps(x,sort_keys=True)
 ca=collections.Counter(map(encode,a));cb=collections.Counter(map(encode,b));gone=[json.loads(x) for x,n in (ca-cb).items() for _ in range(n)];new=[json.loads(x) for x,n in (cb-ca).items() for _ in range(n)]
 key=lambda f:(f['phase'],f['where'],f['unit'])
 changed=[]
 for before in gone[:]:
  options=[f for f in new if key(f)==key(before)]
  if len(options)==1:
   after=options[0];changed.append({'before':before,'after':after});gone.remove(before);new.remove(after)
 def counts(fs):return dict(sorted(collections.Counter(f['phase']+'/'+f['kind'] for f in fs).items()))
 out[stem]={'before_count':len(a),'after_count':len(b),'before_by_phase_kind':counts(a),'after_by_phase_kind':counts(b),'new':new,'gone':gone,'changed':changed}
out['checker']={'status':'stopped after TypePredicate target gate failure','full_census_completed':False,'reason_file':'checker-stop-reason.json'}
(e/'census-delta-full.json').write_text(json.dumps(out,indent=2)+'\n')
print(json.dumps({s:{k:(len(v) if isinstance(v,list) else v) for k,v in o.items()} for s,o in out.items()},indent=2),flush=True)

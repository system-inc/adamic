import json,pathlib,collections
r=pathlib.Path('/workspace/scratch/get-properties-runtime');e=pathlib.Path('/workspace/scratch/overload-transformers/stage3/census/overload-transformers/landing-932a6344')
def norm(v,root):
 if isinstance(v,str):return v.replace(str(root)+'/', '')
 if isinstance(v,list):return [norm(x,root) for x in v]
 if isinstance(v,dict):return {k:norm(x,root) for k,x in v.items()}
 return v
def rows(side,file):
 data=[norm(json.loads(s),r/(side+'-adapted')) for s in (e/f'latent-{side}-{file}.jsonl').read_text().splitlines()]
 assert data[-1]['status']=='relation_report'
 assert data[0]['latent_mode']=='full' and data[0]['module_format']=='CommonJSBundle' and data[0]['diagnostics']==[]
 return data[1]['findings']
def diff(a,b):
 def key(x):return json.dumps({k:v for k,v in x.items() if k not in ['start','end']},sort_keys=True)
 ca,cb=collections.Counter(map(key,a)),collections.Counter(map(key,b))
 gone=[json.loads(k) for k,v in (ca-cb).items() for _ in range(v)];new=[json.loads(k) for k,v in (cb-ca).items() for _ in range(v)]
 def loc(x):return x['where'],x['kind']
 shared=set(map(loc,gone))&set(map(loc,new))
 return {'before':len(a),'after':len(b),'new':[x for x in new if loc(x) not in shared],'gone':[x for x in gone if loc(x) not in shared],'changed':[{'site':s,'before':[x for x in gone if loc(x)==s],'after':[x for x in new if loc(x)==s]} for s in sorted(shared)]}
results={}
for f in ['utilities','ts','classFields']:
 d=diff(rows('base',f),rows('candidate',f));results[f]=d
 assert not any('overload 1 of getProperties result' in x.get('reason','') for x in rows('candidate',f))
 for x in d['new']:
  if x['kind'] in ['Refused','NotYet','error','panic']:
   assert x['where'] in ['src/compiler/transformers/utilities.ts:590:12','src/compiler/binder.ts:2258:13'],x
 for change in d['changed']:
  assert change['site'][1]=='Boundary',change
 (e/f'delta-{f}.json').write_text(json.dumps(d,indent=2)+'\n')
 print('PASS full-latent-'+f+' '+json.dumps({k:(len(v) if isinstance(v,list) else v) for k,v in d.items()}),flush=True)
a=rows('base','utilities');b=rows('mutant','utilities');d=diff(a,b)
assert not d['new'] and not d['gone'] and not d['changed']
restored=[x for x in b if x['kind']=='Refused' and 'overload 1 of getProperties result' in x.get('reason','')]
assert len(restored)==3,restored
(e/'mutant-red.json').write_text(json.dumps({'pass':True,'restoredRefusals':restored,'baselineDelta':d},indent=2)+'\n')
(e/'census-summary.json').write_text(json.dumps({f:{k:(len(v) if isinstance(v,list) else v) for k,v in d.items()} for f,d in results.items()},indent=2)+'\n')
print('PASS restoring-mutant-red: three overload refusals restored, exact baseline findings',flush=True)

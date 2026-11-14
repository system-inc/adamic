import json,pathlib,collections
s=pathlib.Path('/workspace/scratch/path-mint-618');e=s/'evidence'
def read(name,tree):
 rows=json.loads('['+','.join((e/(name+'.jsonl')).read_text().splitlines())+']')
 rows=json.loads(json.dumps(rows).replace(str(tree)+'/', ''))
 assert rows[0]['module_format']=='CommonJSBundle' and rows[0]['latent_mode']=='full'
 assert rows[-1]['status']=='relation_report'
 assert rows[0]['diagnostics']==[],rows[0]['diagnostics']
 records=[r for r in rows if 'file' in r]
 assert len(records)==1 and records[0]['file']=='src/compiler/sys.ts'
 findings=[f for r in records for f in r.get('findings',[])]+rows[-1].get('findings',[])
 assert not [f for f in findings if f['kind'] in ['panic','error']], findings
 unique={tuple(f.get(k,'') for k in ['kind','where','reason','text']):f for f in findings}
 # Keep every observed phase/attempt in a separate context ledger.
 contexts=sorted({tuple(f.get(k,'') for k in ['kind','where','reason','text','phase','unit']) for f in findings})
 return {'counts':dict(collections.Counter(k[0] for k in unique)),'rawCounts':dict(collections.Counter(f['kind'] for f in findings)),'unique':unique,'contexts':contexts,'header':rows[0],'trailer':rows[-1]}
a=read('base',s/'base');b=read('after',s/'lane/adapted-tree');m=read('mutant',s/'mutant')
def delta(a,b):
 gone=sorted(a['unique'].keys()-b['unique'].keys());new=sorted(b['unique'].keys()-a['unique'].keys())
 # A same-kind/same-coordinate replacement is a changed finding; report both sides in full.
 changed=[]
 for old in gone[:]:
  candidates=[x for x in new if x[:2]==old[:2]]
  if len(candidates)==1:
   new.remove(candidates[0]);gone.remove(old);changed.append({'before':old,'after':candidates[0]})
 return {'gone':gone,'new':new,'changed':changed,'goneContexts':sorted(set(a['contexts'])-set(b['contexts'])),'newContexts':sorted(set(b['contexts'])-set(a['contexts']))}
r={n:{k:v[k] for k in ['counts','rawCounts','header','trailer']} for n,v in [('base',a),('after',b),('mutant',m)]}
r['delta']=delta(a,b);r['mutantDelta']=delta(b,m)
(e/'census-delta.json').write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r['delta'],indent=2));print({n:r[n]['counts'] for n in ['base','after','mutant']})
assert not r['delta']['new'] and not r['delta']['changed']
assert len(r['delta']['gone'])==1 and r['delta']['gone'][0][1]=='src/compiler/sys.ts:618:33'
assert any(f[0]=='Refused' and f[1]=='src/compiler/sys.ts:618:33' for f in m['unique'])

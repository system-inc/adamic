from pathlib import Path
import json,collections
r=Path('/workspace/scratch/cast-tags');e=r/'evidence'
def read(side):
 rows=[json.loads(x) for x in (e/(side+'-generators.jsonl')).read_text().splitlines()]
 assert len(rows)==3 and rows[0]['status']=='measurement' and rows[-1]['status']=='relation_report'
 assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle' and rows[0]['diagnostics']==[]
 def norm(x):
  if isinstance(x,str):return x.replace(str(r/(side+'-tree')),'<tree>')
  if isinstance(x,list):return [norm(y) for y in x]
  if isinstance(x,dict):return {k:norm(v) for k,v in x.items()}
  return x
 rows=norm(rows);fs=[f for row in rows for f in row.get('findings',[])];return rows,fs
main,a=read('main');mutant,b=read('mutant');enc=lambda x:json.dumps(x,sort_keys=True)
assert collections.Counter(map(enc,a))==collections.Counter(map(enc,b)), 'restoring mutant findings differ'
lines=[2200,2248,2270,2287,2341,2380,2408];targets={}
for line in lines:
 fs=[f for f in b if f['kind']=='Refused' and f.get('where','').split(':')[-2]==str(line)];assert len(fs)==2 and {f['phase'] for f in fs}=={'lowering','refusal_scan'};targets[str(line)]=[{'where':f['where'],'phase':f['phase'],'kind':f['kind'],'reason':f['reason']} for f in fs]
result={'source':json.loads((e/'restoring-mutant-source.json').read_text()),'finding_occurrences':len(b),'exact_finding_multiset_restored_including_byte_offsets':True,'complete_normalized_records_identical':main==mutant,'target_sites':targets}
(e/'restoring-mutant-result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='target_sites'},indent=2))

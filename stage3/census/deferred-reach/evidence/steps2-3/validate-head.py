import sys,json,hashlib,shutil,importlib.util
from pathlib import Path
root=Path('/workspace/adamic');w=Path('/workspace/scratch/abrq61p')
spec=importlib.util.spec_from_file_location('identity',root/'stage3/census/latent/audit_instantiated_by.py');identity=importlib.util.module_from_spec(spec);spec.loader.exec_module(identity)
results={}
for name,base,head in [('directory','base.jsonl','head.jsonl'),('entry','base-entry.jsonl','head-entry.jsonl')]:
 results[name]=identity.compare(w/base,w/head)
print('byte identity',json.dumps(results),flush=True)
(w/'corpus-identity.json').write_text(json.dumps(results,indent=2)+'\n')
sys.path.insert(0,str(root/'stage3/meter'));import report
run=w/'meter-reader';frozen=json.loads((run/'frozen-readiness.json').read_text())
for label in ('main','area'):
 shutil.copyfile(w/'head.jsonl',run/label/'latent.jsonl');shutil.copyfile(w/'head-entry.jsonl',run/label/'tsc/latent.jsonl')
report.run_roots=lambda tree,roots,logs:frozen[json.dumps(sorted(roots))]
pin='bdbc73864f82dbc0c51434decf0ffe96d4782767';tree=Path('/workspace/scratch/abrq61p-adapted')
report.report_pair(tree,tree,run,'20261110T000000Z',pin,pin,root/'stage3/meter/runs')
meter={}
for relative in ['report.json','report.md','README.md','main/report.json','main/report.md','area/report.json','area/report.md']:
 p=run/relative;assert p.read_bytes()==p.with_name(p.name+'.before').read_bytes(),relative;meter[relative]=hashlib.sha256(p.read_bytes()).hexdigest()
(w/'meter-identity.json').write_text(json.dumps(dict(byte_identical=True,held_fixed=['ordinary streams','source trees','timing streams','compiler labels and pins','readiness probe results'],files=meter),indent=2)+'\n');print('seven meter report artifacts byte identical',flush=True)
sys.path.insert(0,str(root/'stage3/census/pair'));import pair
pair_results={}
for name,base,head in [('directory','base.jsonl','head.jsonl'),('entry','base-entry.jsonl','head-entry.jsonl')]:
 d=pair.compare(pair.load(w/base),pair.load(w/head),tree,root/'stage3/api/node_modules/typescript')
 for view in ['raw_findings','distinct_reads']:
  for field in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta','regressions']:assert not d[view][field],(name,view,field)
 pair_results[name]=d;print(name,'pair zero deltas',flush=True)
(w/'pair-identity.json').write_text(json.dumps(pair_results,indent=2)+'\n')

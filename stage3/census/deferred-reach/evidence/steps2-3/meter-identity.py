import json,sys,shutil,hashlib
from pathlib import Path
root=Path('/workspace/adamic');w=Path('/workspace/scratch/abrq61p');run=w/'meter-reader';tree=Path('/workspace/scratch/abrq61p-adapted');pin='bdbc73864f82dbc0c51434decf0ffe96d4782767'
sys.path.insert(0,str(root/'stage3/meter'));import report
run.mkdir(exist_ok=True)
(run/'compiler-mode.json').write_text(json.dumps(dict(mode='per-ref',latent_mode='first-error',commits=dict(main=pin,area=pin))))
for label in ('main','area'):
 d=run/label;(d/'tsc').mkdir(parents=True,exist_ok=True)
 for source,dest in [('ordinary.jsonl','census.jsonl'),('base.jsonl','latent.jsonl'),('base-times.jsonl','latent-times.jsonl'),('ordinary-entry.jsonl','tsc/census.jsonl'),('base-entry.jsonl','tsc/latent.jsonl'),('base-entry-times.jsonl','tsc/latent-times.jsonl')]:shutil.copyfile(w/source,d/dest)
frozen={};original=report.run_roots

def capture(tree,roots,logs):
 key=json.dumps(sorted(roots));result=original(tree,roots,logs);frozen[key]=result;return result
report.run_roots=capture
report.report_pair(tree,tree,run,'20261110T000000Z',pin,pin,root/'stage3/meter/runs')
(run/'frozen-readiness.json').write_text(json.dumps(frozen,indent=2)+'\n')
for path in [run/'report.json',run/'report.md',run/'README.md',run/'main/report.json',run/'main/report.md',run/'area/report.json',run/'area/report.md']:shutil.copyfile(path,path.with_name(path.name+'.before'))
print('baseline meter reports rendered',flush=True)

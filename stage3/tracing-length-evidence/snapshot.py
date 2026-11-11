from pathlib import Path
from run import raw,root
import shutil,json,hashlib
out=Path('/workspace/scratch/tracing-length-evidence-branch/stage3/tracing-length-evidence');out.mkdir(exist_ok=True)
excluded={'adamic','latent','counted-product','before-counted','mutant-pop','mutant-popAll','__pycache__','checkpoint.index'}
files=[]
for p in sorted(raw.rglob('*')):
 rel=p.relative_to(raw)
 if any(x in excluded for x in rel.parts) or not p.is_file():continue
 target=out/rel
 if target.suffix=='.go':target=target.with_suffix('.go.txt')
 target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,target);files.append(str(target.relative_to(out)))
for n in ['setup','fetch','fetch-evidence','switch','npm']:
 p=Path('/tmp/tracing-length-'+n+'.log')
 if p.is_file():shutil.copyfile(p,out/('initial-'+n+'.log'))
(out/'base.json').write_text(json.dumps({'task':'ff5eddz','base':'16429a30452d027809f0ea85bc395479423e493e','typescript':'050880ce59e30b356b686bd3144efe24f875ebc8','node':'v24.19.0','platform':'linux x64','evidenceGoSuffix':'.go.txt'},indent=2)+'\n')
print(json.dumps({'copiedFiles':len(files),'destination':str(out)}))

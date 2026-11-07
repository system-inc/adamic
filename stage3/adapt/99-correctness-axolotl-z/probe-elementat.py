import hashlib,json,os,shutil,subprocess
from pathlib import Path
s=Path('/workspace/scratch/elementat');raw=s/'raw';before=s/'before';binary=s/'bin';rows=[]
manifest={str(p.relative_to(before/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((before/'src').rglob('*')) if p.is_file()}
(raw/'source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
for variant in ['baseline','direct','probeOnly','callOnly','satisfies','indexed']:
 tree=s/('probe-'+variant);shutil.copytree(before/'src',tree/'src')
 core=tree/'src/compiler/core.ts';core.write_bytes((raw/(variant+'.core.ts')).read_bytes())
 roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 (raw/(variant+'-roots.json')).write_text(json.dumps(roots,indent=2)+'\n')
 result={'variant':variant}
 for kind,args in [('types',[str(binary/'adamic'),'types',*roots]),('binder',[str(binary/'tool'),str(tree/'src/compiler/binder.ts'),str(raw/(variant+'-binder.jsonl'))])]:
  with (raw/(variant+'-'+kind+'.log')).open('w') as log:
   run=subprocess.run(args,stdout=log,stderr=log,env=dict(os.environ,GOMAXPROCS='2'),timeout=900)
  result[kind+'Exit']=run.returncode
  if kind=='binder':result['binder']=[json.loads(l) for l in (raw/(variant+'-binder.jsonl')).read_text().splitlines()]
 rows.append(result);(raw/'probes.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(variant,result['typesExit'],[(r['kind'],r.get('where'),r.get('reason')) for r in result['binder']],flush=True)
assert manifest=={str(p.relative_to(before/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((before/'src').rglob('*')) if p.is_file()}

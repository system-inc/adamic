from pathlib import Path
import subprocess,json,time,os
root=Path('/workspace/adamic'); out=root/'review/test-audit/stage1-cohere-graphql'
plans=json.loads((out/'plan.json').read_text()); metrics=[]
for m in plans:
 p=root/m['file']; original=p.read_text(); p.write_text(original.replace(m['old'],m['new']))
 try:
  env=os.environ.copy(); env['ADAMIC_GRAPHQL_LIBRARY']='/tmp/u096/library'; env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u096/cache/'+m['id']
  if m['file'].endswith('.go'):
   with (out/'logs'/f"{m['id']}-vet.log").open('w') as log: subprocess.run(['go','vet','./internal/lower/'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
  start=time.monotonic()
  with (out/'logs'/f"{m['id']}.log").open('w') as log:
   r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/','-run','.'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  metrics.append(dict(id=m['id'],wall=time.monotonic()-start,exit=r.returncode)); (out/'matrix-times.json').write_text(json.dumps(metrics,indent=2))
 finally: p.write_text(original)

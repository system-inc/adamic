from pathlib import Path
import subprocess,os,json,time
p=Path('/tmp/defend-scanner/evidence');source=Path('/tmp/defend-scanner/rounding.ts');source.write_text("import { programArguments } from 'adamic';\nconst value = Number(programArguments()[0] ?? '0');\nconsole.log(`${(value + 1) - value}`);\n")
results=[]
for id in ['clean','D5']:
 if id=='D5':subprocess.run(['git','apply',str(p/'D5.diff')],check=True)
 try:
  for step,cmd in [('cli-build',['go','build','-o','/tmp/defend-scanner/adamic-'+id,'./cmd/adamic']),('native-build',['timeout','90','/tmp/defend-scanner/adamic-'+id,'build',str(source),'-o','/tmp/defend-scanner/rounding-'+id]),('execute',['/tmp/defend-scanner/rounding-'+id,'1e20'])]:
   start=time.monotonic()
   with (p/('D5-witness-'+id+'-'+step+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-scanner/cache/witness-'+id))
   results.append(dict(id=id,step=step,command=cmd,exit=r.returncode,seconds=time.monotonic()-start));assert r.returncode==0,(id,step)
 finally:
  if id=='D5':subprocess.run(['git','apply','-R',str(p/'D5.diff')],check=True)
(p/'D5-witness-runs.json').write_text(json.dumps(results,indent=2));(p/'D5-witness-input.ts.txt').write_text(source.read_text());print('clean:',(p/'D5-witness-clean-execute.log').read_text(),'D5:',(p/'D5-witness-D5-execute.log').read_text())

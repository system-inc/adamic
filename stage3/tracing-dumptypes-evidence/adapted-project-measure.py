from pathlib import Path
import subprocess,os,json
S=Path(__file__).resolve().parent;cli=S/'counted-product/tree/built/local/tsc.js';counts=S/'project-counts';counts.mkdir();projects=json.loads((S/'stock-projects.json').read_text());rows=[]
for i,project in enumerate(projects):
 name=project['project'];cmd=['node',str(cli),'--project','/workspace/scratch/dp24r5m-typescript/tests/cases/projects/'+name,'--noEmit','--ignoreDeprecations','6.0','--generateTrace',str(S/('adapted-project-trace-'+str(i)))]
 env=dict(os.environ,NODE_OPTIONS='--require='+str(S/'preload.cjs'),DUMP_COUNTER_DIR=str(counts))
 with (S/('adapted-project-'+str(i)+'.log')).open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 row=dict(project=name,command=cmd,exit=p.returncode);rows.append(row);print(json.dumps(row),flush=True)
(S/'adapted-projects.json').write_text(json.dumps(rows,indent=2)+'\n')
assert [r['exit'] for r in rows]==[r['exit'] for r in projects]

from pathlib import Path
import shutil,subprocess,os,json
S=Path(__file__).resolve().parent
lib=S/'stock-late';shutil.copytree('/workspace/adamic-tools/typescript/lib',lib)
subprocess.run(['node',S/'instrument.cjs',lib],check=True)
counts=S/'stock-late-counts';counts.mkdir()
projects=['jsFileCompilation/DifferentNamesNotSpecified','projectOption/DefaultExcludeOnlyNodeModules','decoratorMetadata/emitDecoratorMetadataCommonJSIsolatedModule']
rows=[]
for i,name in enumerate(projects):
 env=dict(os.environ,NODE_OPTIONS='--require='+str(S/'preload.cjs'),DUMP_COUNTER_DIR=str(counts))
 cmd=['node',str(lib/'tsc.js'),'--project','/workspace/scratch/dp24r5m-typescript/tests/cases/projects/'+name,'--noEmit','--ignoreDeprecations','6.0','--generateTrace',str(S/('stock-late-trace-'+str(i)))]
 with (S/('stock-late-project-'+str(i)+'.log')).open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 rows.append(dict(project=name,command=cmd,exit=p.returncode));print(json.dumps(rows[-1]),flush=True)
(S/'stock-late-projects.json').write_text(json.dumps(rows,indent=2)+'\n')

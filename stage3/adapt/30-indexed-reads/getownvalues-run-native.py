from pathlib import Path
import subprocess,json,os,time
s=Path('/workspace/scratch/gov-a');d=s/'witnesses';results={}
for name in ['sparse-adapted','presentUndefined-adapted','empty-adapted','record-adapted','sparseWrite','arrayArm','recordArm']:
 row={};file=d/(name+'.ts');expected=(d/(name+'.node.stdout')).read_bytes() if (d/(name+'.node.stdout')).exists() else (d/(name.removesuffix('-adapted')+'-adapted.node.stdout')).read_bytes()
 for backend,args in [('native',['build',str(file),'-o',str(d/(name+'.native')),'--sanitize','--explain-checks']),('javascript',['js',str(file),'--explain-checks'])]:
  before=time.monotonic();stdout=d/(name+'.'+backend+'.compile.stdout');stderr=d/(name+'.'+backend+'.compile.stderr')
  with stdout.open('wb') as out,stderr.open('wb') as err:r=subprocess.run([s/'adamic',*args],stdout=out,stderr=err,timeout=180)
  entry={'compileExit':r.returncode,'seconds':time.monotonic()-before}
  if r.returncode==0:
   if backend=='javascript':(d/(name+'.adamic.mjs')).write_bytes(stdout.read_bytes());cmd=['node',str(d/(name+'.adamic.mjs'))]
   else:cmd=[str(d/(name+'.native'))]
   with (d/(name+'.'+backend+'.stdout')).open('wb') as out,(d/(name+'.'+backend+'.stderr')).open('wb') as err:run=subprocess.run(cmd,stdout=out,stderr=err,timeout=30)
   entry.update(runExit=run.returncode,stdoutMatchesNode=(d/(name+'.'+backend+'.stdout')).read_bytes()==expected,stderr=(d/(name+'.'+backend+'.stderr')).read_text(),stdout=(d/(name+'.'+backend+'.stdout')).read_text())
  else:entry['diagnostic']=stderr.read_text()
  row[backend]=entry
 results[name]=row;(s/'raw/backend-witnesses.json').write_text(json.dumps(results,indent=2)+'\n');print(name,json.dumps(row),flush=True)

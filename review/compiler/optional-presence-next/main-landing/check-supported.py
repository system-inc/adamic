from pathlib import Path
import subprocess,json,os,time
out=Path('review/compiler/optional-presence-next/main-landing');source=str((out/'supported-neighbor.a').resolve());rows={};compiler='/workspace/scratch/optional-next-adamic';scratch=Path('/workspace/scratch/optional-presence-next')
def run(name,args,env=None):
 start=time.monotonic()
 with (out/(name+'.out.log')).open('wb') as stdout,(out/(name+'.err.log')).open('wb') as stderr:
  result=subprocess.run(args,stdout=stdout,stderr=stderr,env=env,timeout=60)
 row=dict(command=args,exit=result.returncode,seconds=round(time.monotonic()-start,3),stdout=(out/(name+'.out.log')).read_text(),stderr=(out/(name+'.err.log')).read_text());rows[name]=row;return row
truth=run('supported-node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',source]);assert (truth['exit'],truth['stdout'],truth['stderr'])==(0,'2\n2\nok\n','')
for mode,flags in [('native',[]),('sanitized',['--sanitize'])]:
 binary=str(scratch/('supported-'+mode));build=run('supported-'+mode+'-compile',[compiler,'build',source,'-o',binary]+flags);assert build['exit']==0,build
 result=run('supported-'+mode,[binary],dict(os.environ,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='halt_on_error=1'));assert all(result[k]==truth[k] for k in ['exit','stdout','stderr']),result
build=run('supported-js-compile',[compiler,'js',source]);assert build['exit']==0,build
js=scratch/'supported.mjs';js.write_text(build['stdout']);result=run('supported-js',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(js)]);assert all(result[k]==truth[k] for k in ['exit','stdout','stderr']),result
(out/'supported-results.json').write_text(json.dumps(rows,indent=2)+'\n');print('Supported tuple, array, and ordered static neighbor matches Node in both backends and ASan/UBSan/LSan')

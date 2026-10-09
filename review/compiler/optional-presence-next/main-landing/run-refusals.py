from pathlib import Path
import subprocess,json,time,os
p=Path('review/compiler/optional-presence-next/main-landing'); s=Path('/workspace/scratch/optional-presence-next'); s.mkdir(exist_ok=True); rows=[]
def run(n,m,args,env=None):
 t=time.monotonic()
 with (p/(n+'.'+m+'.out.log')).open('wb') as out,(p/(n+'.'+m+'.err.log')).open('wb') as err:
  try: code=subprocess.run(args,stdout=out,stderr=err,env=env,timeout=60).returncode
  except subprocess.TimeoutExpired: code=124
 return dict(command=args,exit=code,seconds=round(time.monotonic()-t,3),stdout=(p/(n+'.'+m+'.out.log')).read_text(errors='backslashreplace'),stderr=(p/(n+'.'+m+'.err.log')).read_text(errors='backslashreplace'))
def obs(x): return x['exit'],x['stdout'],x['stderr']
for src in sorted(Path('review/compiler/eep-presence').glob('*.a.txt')):
 n=src.name[:-6]; f=s/(n+'.a'); f.write_bytes(src.read_bytes()); r=dict(name=n,source=str(src))
 r['node']=run(n,'node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(f)])
 for m,extra in [('native',[]),('sanitized',['--sanitize'])]:
  exe=s/(n+'-'+m); r[m+'_compile']=run(n,m+'-compile',['/workspace/scratch/optional-next-adamic','build',str(f),'-o',str(exe)]+extra)
  if r[m+'_compile']['exit']==0: r[m]=run(n,m,[str(exe)],dict(os.environ,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='halt_on_error=1'))
 r['js_compile']=run(n,'js-compile',['/workspace/scratch/optional-next-adamic','js',str(f)])
 if r['js_compile']['exit']==0:
  js=s/(n+'.mjs'); js.write_text(r['js_compile']['stdout']); r['js_compile']['stdout']='generated JavaScript: '+n+'.js-compile.out.log'; r['js']=run(n,'js',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(js)])
 cs=[r[m+'_compile'] for m in ['native','sanitized','js']]
 if all(x['exit']==1 and ("stage 0 can't lower" in x['stderr'] or 'Adamic 0.1 refuses' in x['stderr']) for x in cs): r['classification']='REFUSED'
 elif all(m in r and obs(r[m])==obs(r['node']) for m in ['native','sanitized','js']): r['classification']='FIXED'
 else: r['classification']='STILL WRONG'
 rows.append(r); (p/'results.json').write_text(json.dumps(rows,indent=2)+'\n'); print(n,r['classification'],flush=True)

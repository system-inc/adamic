from pathlib import Path
import subprocess,json,time
root=Path.cwd();ev=root/'review/compiler/loom-implied-target';source=ev/'objects-sequential.a';cli='/tmp/implied-adamic';rows=[]
def run(name,args,limit=90):
 start=time.monotonic();p=subprocess.run(list(map(str,args)),cwd=root,capture_output=True,timeout=limit)
 r=dict(name=name,command=list(map(str,args)),exit=p.returncode,stdout=p.stdout.decode(),stderr=p.stderr.decode(),seconds=round(time.monotonic()-start,3));rows.append(r);return r
try:
 node=run('node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',source],40)
 js=run('javascript-compile',[cli,'js',source]);(ev/'control.mjs').write_text(js['stdout'])
 javascript=run('javascript',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',ev/'control.mjs'],40)
 for name,flags in [('native',[]),('sanitized',['--sanitize'])]:
  binary='/tmp/implied-control-'+name
  built=run(name+'-compile',[cli,'build',source,'-o',binary,*flags])
  if built['exit']!=0:raise RuntimeError(built)
  got=run(name,[binary],40)
  assert all(got[k]==node[k] for k in ['stdout','stderr','exit']),got
 assert all(javascript[k]==node[k] for k in ['stdout','stderr','exit']),javascript
 assert node['exit']==0
 print('all backends agree:',node['stdout'].strip())
finally:
 for r in rows:
  if r['name']=='javascript-compile':r.pop('stdout',None)
 (ev/'backend-results.json').write_text(json.dumps(rows,indent=2)+'\n')

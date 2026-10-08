from pathlib import Path
import subprocess,json,os,time
base=Path('/tmp/class-refusals'); cases=json.loads((base/'supplement-cases.json').read_text()); results=[]
def run(args):
 p=subprocess.run(args,capture_output=True,text=True,timeout=90)
 return dict(exit=p.returncode,stdout=p.stdout,stderr=p.stderr)
for case in cases:
 name=case['name']; path=case['path']; r=dict(case)
 r['node']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',path])
 r['compile']=run([str(base/'adamic'),'js',path])
 if r['compile']['exit']==0:
  script=base/(name+'.mjs'); script.write_text(r['compile']['stdout']); r['backend']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(script)])
  for mode,flags in [('release',[]),('sanitized',['--sanitize'])]:
   binary=base/(name+'-'+mode)
   # Use the CLI source entry point explicitly for each successful probe.
   r[mode+'_build']=run(['go','run','./cmd/adamic','build',path,'-o',str(binary)]+flags)
   if r[mode+'_build']['exit']==0: r[mode]=run([str(binary)])
 results.append(r); (base/'supplement-results.json').write_text(json.dumps(results,indent=2))
 print(name,case['kind'],'ACCEPT' if r['compile']['exit']==0 else r['compile']['stderr'].strip(),flush=True)
 if 'release' in r: print('  Node',repr(r['node']['stdout']),'release',repr(r['release']['stdout']),'sanitized',repr(r.get('sanitized')),'backend',repr(r['backend']['stdout']),flush=True)

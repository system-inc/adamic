import json, os, subprocess, sys
from pathlib import Path
root=Path(__file__).resolve().parent
scratch=Path('/tmp/coverage-oct8-typedarrays');scratch.mkdir(exist_ok=True)
results={}
compiler=os.environ.get('ADAMIC_COVERAGE_COMPILER','/tmp/coverage-adamic')
def run(cmd, tag, env=None):
 p=subprocess.run(cmd,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=60,env=env)
 (scratch/(tag+'.stdout')).write_bytes(p.stdout);(scratch/(tag+'.stderr')).write_bytes(p.stderr)
 return {'command':cmd,'exit':p.returncode,'stdout':p.stdout.decode(errors='replace'),'stderr':p.stderr.decode(errors='replace')}
for source in sorted(root.glob('*.a')):
 name=source.stem
 if len(sys.argv)>1 and name not in sys.argv[1:]:continue
 r={};r['source']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(source)],name+'-source')
 for mode in ['sanitized','release','javascript']:
  target=scratch/(name+('.mjs' if mode=='javascript' else '-'+mode))
  cmd=[compiler,'js',str(source)] if mode=='javascript' else [compiler,'build',str(source),'-o',str(target)]+(['--sanitize'] if mode=='sanitized' else [])
  build=run(cmd,name+'-'+mode+'-build');r[mode+'-build']=build
  if build['exit']==0:
   if mode=='javascript':target.write_text(build['stdout'])
   env=dict(os.environ,ASAN_OPTIONS='detect_leaks=0',UBSAN_OPTIONS='halt_on_error=1')
   r[mode]=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(target)] if mode=='javascript' else [str(target)],name+'-'+mode,env)
   if mode=='sanitized' and r[mode]['exit']==0:r['leaks']=run([str(target)],name+'-leaks',dict(env,ASAN_OPTIONS='detect_leaks=1'))
  else:r[mode]={'exit':build['exit'],'stdout':build['stdout'],'stderr':build['stderr'],'phase':'compiler'}
 comparable=lambda v:(v['exit'],v['stdout'],v['stderr'])
 r['agree']=all(comparable(r[m])==comparable(r['source']) for m in ['sanitized','release','javascript'])
 
 for mode in ['sanitized','release','javascript']:
  if r[mode+'-build']['exit']==0:r[mode+'-build']['stdout']=''
 results[name]=r;print(name,r['agree'],[(m,r[m]['exit']) for m in ['source','sanitized','release','javascript']],flush=True)
Path(os.environ.get('ADAMIC_COVERAGE_RESULTS',str(root/'results.json'))).write_text(json.dumps(results,indent=2)+'\n')
failed=False
for name,r in results.items():
 if name in ['buffer-gap','readiness-callback']:
  expected='ArrayBuffer' if name=='buffer-gap' else 'refuses the non-null assertion'
  passed=r['source']['exit']==0 and all(r[mode].get('phase')=='compiler' and r[mode]['exit']==1 and expected in r[mode]['stderr'] for mode in ['sanitized','release','javascript'])
 else:
  passed=r['agree'] and r['source']['exit']==0 and r['leaks']['exit']==0 and r['leaks']['stderr']==''
 failed=failed or not passed
sys.exit(1 if failed else 0)

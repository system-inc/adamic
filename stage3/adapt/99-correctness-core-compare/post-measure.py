import json,os,signal,subprocess,time,shutil,hashlib
from pathlib import Path
s=Path('/workspace/scratch/core-compare');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-compare';phases={};env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source')
assert json.loads((s/'CONTROL.json').read_text())['complete'];assert json.loads((s/'RECONCILED.json').read_text())['complete']
def run(name,args,cwd=repo):
 print('start '+name,flush=True);begin=time.monotonic()
 with (s/(name+'.log')).open('w') as log:
  p=subprocess.Popen(args,cwd=cwd,env=env,stdout=log,stderr=log,start_new_session=True)
  try:code=p.wait(timeout=1700)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-begin};(s/'POST-MEASURE.json').write_text(json.dumps(phases,indent=2)+'\n');assert code==0,(name,code);print('pass '+name,flush=True)
run('construct',['node',str(s/'construct.cjs')])
shutil.copytree(s/'before',s/'after',symlinks=True)
run('adapt-after',['node',str(out/'adapt.cjs'),str(s/'after')])
run('contract',['node',str(s/'contract.cjs')]);run('witness',['node',str(s/'witness.cjs')])
observed=set(json.loads((s/'COUNT.json').read_text())['counts']);witnesses=json.loads((s/'witness.log').read_text());covered={r['typeofA']+'/'+r['typeofB'] for r in witnesses['rows'] if r['same']};assert observed<=covered,(observed-covered)
for side in ['before','after']:
 (s/side/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
 run('build-'+side,['npm','run','build:compiler'],cwd=s/side)
(s/'comparison').mkdir();run('comparison',['node',str(s/'compare.cjs'),str(s/'before'),str(s/'after'),str(s/'comparison')])
run('types-after',['python3',str(s/'check-types.py'),'after'])
run('census-after',['python3',str(s/'census.py'),'after'])
(s/'mutant').mkdir();shutil.copytree(s/'after/src',s/'mutant/src',symlinks=True);(s/'mutant/node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
pin=json.loads((out/'owners.json').read_text())[0];file=s/'mutant/src/compiler/core.ts';data=file.read_bytes();nl=b'\r\n' if b'\r\n' in data else b'\n';a=pin['after'].encode().replace(b'\n',nl);b=pin['before'].encode().replace(b'\n',nl);assert data.count(a)==1;file.write_bytes(data.replace(a,b));assert file.read_bytes()==(s/'before/src/compiler/core.ts').read_bytes()
run('census-mutant',['python3',str(s/'census.py'),'mutant']);run('census-comparison',['python3',str(s/'compare-census.py')])
run('paired-replay',['python3',str(s/'replay.py')])
def manifest(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(tree.rglob('*')) if p.is_file()}
a,b=manifest(s/'replay-before/src'),manifest(s/'replay-after/src');assert a==b;(s/'replay/source-equality.json').write_text(json.dumps({'files':len(a),'byteIdentical':True},indent=2)+'\n')
print('POST MEASUREMENT PROOFS COMPLETE',flush=True)

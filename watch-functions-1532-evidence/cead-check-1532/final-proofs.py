from pathlib import Path
import os,json,subprocess,time,hashlib,difflib
s=Path('/tmp/watch-file-check-cead');repo=s/'work';after=s/'after-tree';before=Path('/tmp/watch-functions-1532/main-tree')
while not (s/'COUNT_OK').exists():
 if (s/'count-exit.txt').exists():
  r=json.loads((s/'count-oracle/report.json').read_text())
  if r.get('counts')!={'passing':106366,'failing':1,'pending':0}:raise RuntimeError('quiet counting oracle did not meet the exact expected result')
 time.sleep(1)
assert (s/'APPLY_OK').exists()
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/watch-functions-1532/TypeScript',STAGE3_CACHE=str(s/'cache'),ADAMIC_BUILD_CACHE_DIR=str(s/'cache'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE='/tmp/watch-file-check-1532/go-cache',NODE_PATH=str(repo/'stage3/api/node_modules'),npm_config_cache='/tmp/watch-functions-1532/npm-cache',PYTHONDONTWRITEBYTECODE='1',NODE_OPTIONS='--max-old-space-size=2048')
commands=[]
def run(name,cmd,cwd=repo,variables=env,expected=0):
 start=time.monotonic()
 with (s/(name+'.log')).open('w') as log:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=variables,stdout=log,stderr=subprocess.STDOUT)
 receipt={'name':name,'command':list(map(str,cmd)),'cwd':str(cwd),'exit':r.returncode,'seconds':time.monotonic()-start,'environment':{k:variables[k] for k in ['ADAMIC_TYPESCRIPT_SOURCE','STAGE3_CACHE','TMPDIR','GOTMPDIR','GOCACHE','NODE_PATH','npm_config_cache','NODE_OPTIONS'] if k in variables}}
 commands.append(receipt);(s/'proof-commands.json').write_text(json.dumps(commands,indent=2)+'\n');assert r.returncode==expected,receipt
run('after-install',['npm','ci','--no-audit','--no-fund'],after)
run('after-build',['npm','run','build:compiler'],after)
run('after-dts',['node_modules/.bin/hereby','dts'],after)
run('scope-stock',['node',s/'scope-stock.cjs'])
run('contracts',['node',s/'contracts.cjs'])
for label,tree in [('main',before),('after',after)]:
 roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79;run('types-'+label,[s/'adamic','types',*roots])
run('witness',['node',s/'witness.cjs'])
declarations=[];diff=''
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(s/'baseline-declarations'/name).read_bytes();b=(after/'built/local'/name).read_bytes();diff+=''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='main/'+name,tofile='after/'+name));declarations.append({'file':name,'mainSha256':hashlib.sha256(a).hexdigest(),'afterSha256':hashlib.sha256(b).hexdigest(),'bytes':len(a),'equal':a==b})
(s/'declarations.diff').write_text(diff);(s/'declarations.json').write_text(json.dumps(declarations,indent=2)+'\n');assert all(d['equal'] for d in declarations)
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip()=='ceadba1d9baf55dc0917eaf27dec69ebada795db'
run('new-branch',['git','switch','-c','typescript/watch-file-check-1532-1'])
run('stage-implementation',['git','add','stage3/adapt/93-builtin-imports','stage3/adapt/99-correctness-core-multimap','stage3/adapt/99-correctness-single-watcher-forwarder','stage3/adapt/99-correctness-timer-closures','stage3/adapt/99-wasm-crypto-import','stage3/upstream/LEDGER.md'])
run('commit-implementation',['git','commit','-m','stage3: check dynamic watch intervals at their true boundary (#cyjqp5f)'])
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip();assert subprocess.check_output(['git','rev-parse','HEAD^'],cwd=repo,text=True).strip()=='ceadba1d9baf55dc0917eaf27dec69ebada795db';(s/'implementation-sha.txt').write_text(sha+'\n')
(s/'PRE_LANE_OK').write_text('all pre-lane proofs passed\n')
# Run the whole lane with its normal Node defaults, no test filter and no competing proof work.
lane_env=dict(env);lane_env.pop('NODE_OPTIONS',None)
(s/'memory-events-before-lane.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
run('whole-lane',[repo/'stage3/lane/run.sh',s/'lane'],variables=lane_env)
(s/'memory-events-after-lane.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
run('push-implementation',['git','push','origin','HEAD:refs/heads/typescript/watch-file-check-1532-1'])
(s/'ALL_PROOFS_OK').write_text(sha+'\n')

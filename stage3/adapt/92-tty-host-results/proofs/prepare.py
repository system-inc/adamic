from pathlib import Path
import subprocess as s,os,json,time,shutil,difflib
R=Path('/tmp/tty-adaptation/head');D=Path('/tmp/tty-adaptation');P=R/'stage3/adapt/92-tty-host-results/proofs';records=[]
env=dict(os.environ,PATH='/workspace/adamic-tools/bin:/opt/codex/runtimes/codex-primary-runtime/dependencies/node/bin:/usr/bin:/bin',NODE_PATH=str(R/'stage3/api/node_modules'),NODE_OPTIONS='--max-old-space-size=3072',npm_config_cache=str(D/'npm-cache'),HOME=str(D/'home'),GOMAXPROCS='2',PYTHONDONTWRITEBYTECODE='1');(D/'home').mkdir(exist_ok=True)
def run(name,command,cwd=R,expected=0,seconds=600):
 start=time.monotonic()
 with (P/'logs'/(name+'.log')).open('w') as f:r=s.run([str(x) for x in command],cwd=cwd,env=env,stdout=f,stderr=s.STDOUT,timeout=seconds)
 row={'name':name,'command':list(map(str,command)),'cwd':str(cwd),'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'limit':seconds};records.append(row);(P/'commands.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps(row),flush=True)
 assert r.returncode==expected,(name,r.returncode)
for label in ['control','after']:run(label+'-install',['npm','ci','--no-audit','--no-fund'],D/label,seconds=600)
run('checks',['node',P/'check.cjs',D/'control',D/'after',P/'checks.json'],seconds=300)
run('undone-checker',['node',P/'check.cjs','--require-repaired',D/'control'],expected=1,seconds=180)
run('restored-checker',['node',P/'check.cjs','--require-repaired',D/'after'],seconds=180)
run('idempotence',['node',P.parent/'adapt.cjs',D/'after'],seconds=180)
assert json.loads((P/'logs/idempotence.log').read_text())['files']==0
# The movement guard validates both methods before the sole write.
f=D/'after/src/compiler/sys.ts';saved=f.read_bytes()
try:
 f.write_bytes(saved+b'\n');run('moved-target-mutant',['node',P.parent/'adapt.cjs',D/'after'],expected=1,seconds=180)
 assert f.read_bytes()==saved+b'\n'
finally:f.write_bytes(saved)
# A runtime-only mutant must fail the actual reader witness, not a source hash.
try:
 assert saved.count(b'columns ?? 0')==1;f.write_bytes(saved.replace(b'columns ?? 0',b'columns ?? 1'))
 run('wrong-fallback-mutant',['node','-e',f"require({str(P/'check.cjs')!r}).TestTTYConsumerEquivalence({str(D/'control')!r},{str(D/'after')!r})"],expected=1,seconds=60)
finally:f.write_bytes(saved)
a=(D/'control/src/compiler/sys.ts').read_text();b=f.read_text();(P/'source.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/src/compiler/sys.ts',tofile='b/src/compiler/sys.ts')))
for label in ['control','after']:run(label+'-build',['npm','run','build'],D/label,seconds=1200)
apiA=(D/'control/built/local/typescript.d.ts').read_bytes();apiB=(D/'after/built/local/typescript.d.ts').read_bytes();assert apiA==apiB,'public API changed'
shutil.copyfile(D/'logs/control-apply.log',P/'logs/control-apply.log');shutil.copyfile(D/'logs/after-apply.log',P/'logs/after-apply.log');shutil.copyfile(D/'after/patch-set.md',P/'patch-set.md')
(P/'preparation.json').write_text(json.dumps({'status':'pass','apiBytes':len(apiA),'mutants':['undone with A types','moved source','wrong numeric fallback']},indent=2)+'\n')
print('PASS bounded preparation',flush=True)

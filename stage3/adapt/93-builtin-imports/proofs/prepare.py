from pathlib import Path
import os,subprocess as s,time,json,hashlib,shutil,signal,difflib
R=Path(__file__).resolve().parents[4];D=Path(os.environ.get('BUILTIN_PROOF_SCRATCH','/workspace/scratch/adapt-93-proof'));P=Path(__file__).resolve().parent
env=dict(os.environ,NODE_PATH=str(R/'stage3/api/node_modules'),NODE_OPTIONS='--max-old-space-size=3072',PYTHONDONTWRITEBYTECODE='1',npm_config_cache=str(D/'npm-cache'))
records=[]
def run(name,cmd,cwd=R,expected=0,limit=600):
 start=time.monotonic()
 with (P/'logs'/f'{name}.log').open('w') as f:
  proc=s.Popen(list(map(str,cmd)),cwd=cwd,env=env,stdout=f,stderr=s.STDOUT,start_new_session=True)
  minimum=shutil.disk_usage(D).free
  while proc.poll() is None:
   minimum=min(minimum,shutil.disk_usage(D).free)
   if time.monotonic()-start>limit or minimum<3*1024**3:
    os.killpg(proc.pid,signal.SIGTERM);proc.wait(timeout=10);raise RuntimeError((name,'time/disk bound'))
   time.sleep(1)
 row=dict(name=name,command=list(map(str,cmd)),cwd=str(cwd),exit=proc.returncode,seconds=round(time.monotonic()-start,3),limit=limit,minFreeBytes=minimum)
 records.append(row);(P/'commands.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps(row),flush=True)
 assert proc.returncode==expected,(name,proc.returncode)
 return row
# Pin the fully adapted main input, retaining its exact newline bytes.
f=D/'control/src/compiler/sys.ts';before=f.read_bytes();after=before
for module,name in [('fs','adamicNodeFs93'),('path','adamicNodePath93'),('os','adamicNodeOs93')]:
 needle=f'require("{module}")'.encode();assert after.count(needle)==1;after=after.replace(needle,name.encode())
# Main's final import is the hostErrors helper. Use its exact complete line.
needle=b'import { errorCode } from "./hostErrors.js";';assert after.count(needle)==1
nl=b'\r\n' if b'\r\n' in before else b'\n'
imports=b''.join(nl+f'import * as {name} from "node:{module}";'.encode() for module,name in [('fs','adamicNodeFs93'),('path','adamicNodePath93'),('os','adamicNodeOs93')])
after=after.replace(needle,needle+imports)
h=lambda b:hashlib.sha256(b).hexdigest()
(P/'targets.json').write_text(json.dumps([dict(file='src/compiler/sys.ts',beforeSha256=h(before),afterSha256=h(after))],indent=2)+'\n')
run('after-clone',['git','clone','--shared',D/'upstream',D/'after'])
run('after-apply',['python3',R/'stage3/apply.py','--adapt-checkout',D/'after'],limit=1800)
assert (D/'after/src/compiler/sys.ts').read_bytes()==after
shutil.copyfile(D/'after/patch-set.md',R/'stage3/patch-set.md')
shutil.copyfile(D/'after/patch-set.md',P/'patch-set.md')
shutil.copyfile(D/'control/patch-set.md',P/'patch-set-control.md')
shutil.copyfile(D/'logs/control-apply.log',P/'logs/control-apply.log')
run('checks',['node',P/'check.cjs',D/'control',D/'after',P/'checks.json'],limit=300)
run('idempotence',['node',P.parent/'adapt.cjs',D/'after'],limit=180)
assert json.loads((P/'logs/idempotence.log').read_text())['files']==0
file=D/'after/src/compiler/sys.ts';saved=file.read_bytes();stamp=file.stat().st_mtime_ns
# Prove rerun performs no write, rather than merely reporting zero.
assert saved==after
run('idempotence-no-write',['node',P.parent/'adapt.cjs',D/'after'],limit=180)
assert file.stat().st_mtime_ns==stamp
try:
 file.write_bytes(saved+b'\n');stamp=file.stat().st_mtime_ns
 run('moved-target-mutant',['node',P.parent/'adapt.cjs',D/'after'],expected=1,limit=180)
 assert file.read_bytes()==saved+b'\n' and file.stat().st_mtime_ns==stamp
finally:file.write_bytes(saved)
# A real extra runtime edit: outside the sanctioned imports and initializers.
try:
 needle=b'return process.stdout.columns ?? 0;';assert saved.count(needle)==1
 file.write_bytes(saved.replace(needle,b'return process.stdout.columns ?? 1;'))
 run('scope-mutant',['node',P/'check.cjs',D/'control',D/'after'],expected=1,limit=300)
 assert 'TestBuiltinSourceScope: edit outside' in (P/'logs/scope-mutant.log').read_text()
finally:file.write_bytes(saved)
run('restored-checks',['node',P/'check.cjs',D/'control',D/'after',P/'checks.json'],limit=300)
(P/'source.diff').write_text(''.join(difflib.unified_diff(before.decode().replace('\r\n','\n').splitlines(True),after.decode().replace('\r\n','\n').splitlines(True),fromfile='a/src/compiler/sys.ts',tofile='b/src/compiler/sys.ts',n=0)))
for label in ['control','after']:
 run(label+'-install',['npm','ci','--no-audit','--no-fund'],D/label,limit=600)
 run(label+'-build',['npm','run','build'],D/label,limit=1200)
a=(D/'control/built/local/typescript.d.ts').read_bytes();b=(D/'after/built/local/typescript.d.ts').read_bytes();assert a==b
(P/'preparation.json').write_text(json.dumps(dict(status='pass',apiBytes=len(a),apiSha256=h(a),beforeSysSha256=h(before),afterSysSha256=h(after),rerunDoesNotWrite=True,movedTargetRejectedBeforeWrite=True,scopeMutantCaught=True),indent=2)+'\n')
print('PASS bounded preparation',flush=True)

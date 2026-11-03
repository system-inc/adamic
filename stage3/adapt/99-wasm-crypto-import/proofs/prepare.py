#!/usr/bin/env python3
"""Main's official apply first, then a fresh last-adapter apply and bounded proofs."""
from pathlib import Path
import os,subprocess as s,time,json,hashlib,shutil,signal,difflib
R=Path(__file__).resolve().parents[4];P=Path(__file__).resolve().parent
D=Path(os.environ.get('CRYPTO_PROOF_SCRATCH','/workspace/scratch/wasm-crypto-proof'))
env=dict(os.environ,NODE_PATH=str(R/'stage3/api/node_modules'),NODE_OPTIONS='--max-old-space-size=3072',PYTHONDONTWRITEBYTECODE='1',npm_config_cache=str(D/'npm-cache'))
records=json.loads((D/'control-commands.json').read_text())
assert records[-1]['name']=='control-apply' and records[-1]['exit']==0
(P/'logs').mkdir(exist_ok=True)
for row in records:shutil.copyfile(D/(row['name']+'.log'),P/'logs'/(row['name']+'.log'))
def run(name,cmd,cwd=R,expected=0,limit=600):
 start=time.monotonic();minimum=shutil.disk_usage(D).free
 with (P/'logs'/f'{name}.log').open('w') as f:
  proc=s.Popen(list(map(str,cmd)),cwd=cwd,env=env,stdout=f,stderr=s.STDOUT,start_new_session=True)
  while proc.poll() is None:
   minimum=min(minimum,shutil.disk_usage(D).free)
   if time.monotonic()-start>limit or minimum<3*1024**3:
    os.killpg(proc.pid,signal.SIGTERM);proc.wait(timeout=10);raise RuntimeError((name,'time/disk bound'))
   time.sleep(1)
 row=dict(name=name,command=list(map(str,cmd)),cwd=str(cwd),exit=proc.returncode,seconds=round(time.monotonic()-start,3),limit=limit,minFreeBytes=minimum)
 records.append(row);(P/'commands.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps(row),flush=True)
 assert proc.returncode==expected,(name,proc.returncode)
 return row
# These bytes are the FULL official main tree, including 98-e and 99-i, not 93's snapshot.
f=D/'control/src/compiler/sys.ts';before=f.read_bytes();after=before
nl=b'\r\n' if b'\r\n' in before else b'\n'
assert b'createSHA256Hash?: { createSHA256Hash(data: string): string; }["createSHA256Hash"] | undefined;' in before
block=nl.join(x.encode() for x in [
 '        // crypto can be absent on reduced node installations',
 '        let _crypto: typeof import("crypto") | undefined;',
 '        try {','            _crypto = require("crypto");','        }','        catch {','            _crypto = undefined;','        }',
])+nl
assert after.count(block)==1;after=after.replace(block,b'')
for old,new in [
 (b'import * as _os from "node:os";',b'import * as _os from "node:os";'+nl+b'import * as _crypto from "node:crypto";'),
 (b'createHash: _crypto ? createSHA256Hash : generateDjb2Hash,',b'createHash: createSHA256Hash,'),
 (b'createSHA256Hash: _crypto ? createSHA256Hash : undefined,',b'createSHA256Hash,'),
 (b'_crypto!.createHash("sha256")',b'_crypto.createHash("sha256")'),
]:
 assert after.count(old)==1,old;after=after.replace(old,new)
h=lambda b:hashlib.sha256(b).hexdigest()
(P/'targets.json').write_text(json.dumps([dict(file='src/compiler/sys.ts',beforeSha256=h(before),afterSha256=h(after))],indent=2)+'\n')
run('after-apply',['bash','stage3/apply.sh',D/'after'],limit=1800)
assert (D/'after/src/compiler/sys.ts').read_bytes()==after
# Generated reports are retained here, never copied over main's committed table or plans.
for label in ['control','after']:shutil.copyfile(D/label/'patch-set.md',P/('patch-set.md' if label=='after' else 'patch-set-control.md'))
run('checks',['node',P/'check.cjs',D/'control',D/'after',P/'checks.json'],limit=300)
run('bindings',['node',P/'bindings.cjs',D/'control',D/'after',P/'bindings.json'],limit=300)
file=D/'after/src/compiler/sys.ts';saved=file.read_bytes();stamp=file.stat().st_mtime_ns
run('idempotence-no-write',['node',P.parent/'adapt.cjs',D/'after'],limit=180)
assert json.loads((P/'logs/idempotence-no-write.log').read_text())['files']==0
assert file.read_bytes()==saved and file.stat().st_mtime_ns==stamp
try:
 file.write_bytes(saved+b'\n');stamp=file.stat().st_mtime_ns
 run('moved-target-mutant',['node',P.parent/'adapt.cjs',D/'after'],expected=1,limit=180)
 assert 'target source moved: sys.ts' in (P/'logs/moved-target-mutant.log').read_text()
 assert file.read_bytes()==saved+b'\n' and file.stat().st_mtime_ns==stamp
finally:file.write_bytes(saved)
run('scope-only-mutant',['node',P/'scope-mutant.cjs',D/'control',D/'after'],expected=1,limit=300)
assert 'PASS TestCryptoTranspileScope:' in (P/'logs/scope-only-mutant.log').read_text()
assert 'TestCryptoSourceScope: edit outside' in (P/'logs/scope-only-mutant.log').read_text()
assert file.read_bytes()==saved
try:
 needle=b'return process.stdout.columns ?? 0;';assert saved.count(needle)==1
 file.write_bytes(saved.replace(needle,b'return process.stdout.columns ?? 1;'))
 run('runtime-scope-mutant',['node',P/'check.cjs',D/'control',D/'after'],expected=1,limit=300)
 assert 'TestCryptoSourceScope: edit outside' in (P/'logs/runtime-scope-mutant.log').read_text()
finally:file.write_bytes(saved)
run('ternary-mutant',['node',P/'ternary-mutant.cjs',D/'after'],expected=1,limit=300)
assert 'collision or bare namespace use: _crypto' in (P/'logs/ternary-mutant.log').read_text();assert file.read_bytes()==saved
# A documented earlier adapter's input is deliberately changed only AFTER its one official pass.
run('earlier-reader',['node',R/'stage3/adapt/98-axolotl-e/adapt.cjs',D/'after'],expected=1,limit=300)
assert 'the nodeSystem store: want one, got 0' in (P/'logs/earlier-reader.log').read_text()
assert file.read_bytes()==saved
run('restored-checks',['node',P/'check.cjs',D/'control',D/'after',P/'checks.json'],limit=300)
(P/'source.diff').write_text(''.join(difflib.unified_diff(before.decode().replace('\r\n','\n').splitlines(True),after.decode().replace('\r\n','\n').splitlines(True),fromfile='a/src/compiler/sys.ts',tofile='b/src/compiler/sys.ts',n=0)))
for label in ['control','after']:
 run(label+'-install',['npm','ci','--no-audit','--no-fund'],D/label,limit=600)
 run(label+'-build',['npm','run','build'],D/label,limit=1200)
a=(D/'control/built/local/typescript.d.ts').read_bytes();b=(D/'after/built/local/typescript.d.ts').read_bytes();assert a==b
(P/'preparation.json').write_text(json.dumps(dict(status='pass',apiBytes=len(a),apiSha256=h(a),beforeSysSha256=h(before),afterSysSha256=h(after),rerunDoesNotWrite=True,movedTargetRejectedBeforeWrite=True,scopeMutantsCaught=True,ternaryMutantCaught=True,earlierReaderFailsOnlyOnNewShape=True,shorthandTypechecks=True),indent=2)+'\n')
print('PASS bounded preparation',flush=True)

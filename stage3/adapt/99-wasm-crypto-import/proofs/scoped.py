#!/usr/bin/env python3
"""Fresh official apply, owner-scoped negatives, a foreign edit and one actual lane unit."""
from pathlib import Path
import os,subprocess as s,time,json,hashlib,shutil,signal,gzip
R=Path(__file__).resolve().parents[4];P=Path(__file__).resolve().parent
D=Path(os.environ.get('CRYPTO_SCOPED_SCRATCH','/workspace/scratch/crypto-scoped-proof'));base='4b8634a4a498a39d254ee8c325eb2711f84359d9';records=[]
h=lambda b:hashlib.sha256(b).hexdigest()
def run(name,cmd,cwd=R,expected=0,limit=600,env=None):
 start=time.monotonic();minimum=shutil.disk_usage(D).free;log=P/'logs'/('scoped-'+name+'.log')
 with log.open('w') as f:
  child=s.Popen(list(map(str,cmd)),cwd=cwd,env=env,stdout=f,stderr=s.STDOUT,start_new_session=True)
  while child.poll() is None:
   minimum=min(minimum,shutil.disk_usage(D).free)
   if time.monotonic()-start>limit or minimum<3*1024**3:
    os.killpg(child.pid,signal.SIGTERM);child.wait(timeout=10);raise RuntimeError(name+' time/disk bound')
   time.sleep(1)
 row=dict(name=name,command=list(map(str,cmd)),cwd=str(cwd),exit=child.returncode,seconds=round(time.monotonic()-start,3),limit=limit,minFreeBytes=minimum)
 records.append(row);(P/'scoped-commands.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps(row),flush=True)
 assert child.returncode==expected,(name,child.returncode,log)
 return log
# The original main adapter supplies both real snapshots; no hand-authored owner bytes.
main=D/'main';afterBytes=(main/'src/compiler/sys.ts').read_bytes();beforeBytes=s.check_output(['git','-C',str(main),'show','refs/stage3/99-wasm-crypto-import-before:src/compiler/sys.ts'])
(D/'main-before.ts').write_bytes(beforeBytes)
run('pins',['node',P/'scoped-hashes.cjs',D/'main-before.ts',main/'src/compiler/sys.ts',P/'targets.json'])
pins=json.loads((P/'targets.json').read_text())[0]
run('apply',['bash','stage3/apply.sh',D/'scoped'],limit=1800)
scoped=D/'scoped';assert (scoped/'src/compiler/sys.ts').read_bytes()==afterBytes
for f in (main/'src').rglob('*'):
 if f.is_file():assert f.read_bytes()==(scoped/f.relative_to(main)).read_bytes(),f
assert sum(f.is_file() for f in (scoped/'src').rglob('*'))==746
file=scoped/'src/compiler/sys.ts';saved=file.read_bytes();stamp=file.stat().st_mtime_ns
log=run('rerun',['node',P.parent/'adapt.cjs',scoped]);assert json.loads(log.read_text())['files']==0;assert file.read_bytes()==saved and file.stat().st_mtime_ns==stamp
# Input owner and import pins each have their own direct negative.
needle=b'const isMacOs = process.platform === "darwin";';assert saved.count(needle)==1
mutant=saved.replace(needle,b'const isMacOs = process.platform === "darw1n";')
try:
 file.write_bytes(mutant);stamp=file.stat().st_mtime_ns
 log=run('moved-owner',['node',P.parent/'adapt.cjs',scoped],expected=1)
 assert 'uninventoried owner: getNodeSystem' in log.read_text();assert file.read_bytes()==mutant and file.stat().st_mtime_ns==stamp
finally:file.write_bytes(saved)
try:
 mutant=saved+b'\r\nimport * as cryptoScopedProbe from "node:util";\r\n';file.write_bytes(mutant);stamp=file.stat().st_mtime_ns
 log=run('moved-imports',['node',P.parent/'adapt.cjs',scoped],expected=1)
 assert 'uninventoried node: import block' in log.read_text();assert file.read_bytes()==mutant and file.stat().st_mtime_ns==stamp
finally:file.write_bytes(saved)
# A private copy uses hard links for immutable input, but its sys.ts is detached before any write.
probe=D/'probe';shutil.copytree(main,probe,symlinks=True,copy_function=os.link,ignore=shutil.ignore_patterns('.git'))
f=probe/'src/compiler/sys.ts';f.unlink();f.write_bytes(beforeBytes)
run('bindings',['node',P/'bindings.cjs',probe,scoped,P/'scoped-bindings.json'])
run('scope',['node',P/'check.cjs',probe,scoped,P/'scoped-checks.json'])
site=json.loads((P/'foreign-site.json').read_text());old=site['before'].encode();new=site['after'].encode();assert beforeBytes.count(old)==1
try:
 f.write_bytes(beforeBytes.replace(old,new));log=run('foreign-edit',['node',P.parent/'adapt.cjs',probe]);assert json.loads(log.read_text())['files']==1
 assert f.read_bytes()==afterBytes.replace(old,new)
 actual=json.loads(s.check_output(['node',str(P/'scoped-hashes.cjs'),str(f)]));assert actual['ownerSha256']==pins['ownerAfterSha256'] and actual['importsSha256']==pins['importsAfterSha256']
 foreign=dict(site=site,ownerAfterSha256=actual['ownerSha256'],importsAfterSha256=actual['importsSha256'],foreignEditPreserved=True,wholeFileSha256=h(f.read_bytes()))
finally:f.write_bytes(beforeBytes)
# The deliberately broken copy adds an otherwise invisible insertion in invokeCallbacks.
source=(P.parent/'adapt.cjs').read_text();marker='    let importInsertions = 0;'
extra="    const outside = file.text.indexOf('(existing as string[]).push(...fileNames);');\n    assert(outside >= 0);\n    edits.push({start: outside, end: outside, text: ' '});\n"
assert source.count(marker)==1;broken=source.replace(marker,extra+marker)
spanline="        assert(inOwner || importInsertion, 'edit outside getNodeSystem and node: import insertion');";assert broken.count(spanline)==1
copy=P.parent/'.span-escape-adapter.cjs';unguarded=P.parent/'.span-escape-no-guard.cjs'
try:
 copy.write_text(broken);stamp=f.stat().st_mtime_ns
 log=run('span-escape',['node',copy,probe],expected=1);assert 'edit outside getNodeSystem and node: import insertion' in log.read_text();assert f.read_bytes()==beforeBytes and f.stat().st_mtime_ns==stamp
 unguarded.write_text(broken.replace(spanline,"        // Deliberate control: disable only the new edit-span assertion."))
 log=run('span-disabled-control',['node',unguarded,probe]);assert json.loads(log.read_text())['files']==1
 assert f.read_bytes()!=afterBytes and f.read_bytes().replace(b' (existing as string[]).push(...fileNames);',old)==afterBytes
finally:
 f.write_bytes(beforeBytes);copy.unlink(missing_ok=True);unguarded.unlink(missing_ok=True)
# Produce a real adapted-tree view and run one of the lane's ordinary units, not a full oracle.
run('install',['npm','ci','--no-audit','--no-fund'],scoped)
run('build',['npm','run','build'],scoped,limit=1200)
run('harness',['node',scoped/'node_modules/hereby/bin/hereby.js','tests'],scoped)
assert (scoped/'package-lock.json').read_bytes()==(R/'stage3/typescript/package-lock.json').read_bytes()
shutil.copytree(scoped/'node_modules',R/'stage3/typescript/node_modules',symlinks=True,copy_function=os.link)
(scoped/'node_modules').rename(D/'scoped-modules')
for p in (scoped/'built').rglob('*.tsbuildinfo'):p.unlink()
log=run('lane-unit',['python3','stage3/lane/shard.py','stage3-lane-file-000',D/'lane-results'],env=dict(os.environ,STAGE3_ADAPTED_TREE=str(scoped)),limit=600)
assert (scoped/'src/compiler/sys.ts').read_bytes()==afterBytes
result=dict(base=base,pins=pins,wholeFileSha256=h(afterBytes),byteIdentical=True,all746SourceFilesIdentical=True,foreignEdit=foreign,movedOwnerRejectedBeforeWrite=True,movedImportsRejectedBeforeWrite=True,spanEscapeCaughtBeforeWrite=True,spanDisabledControlWritesExtraEdit=True,rerunFiles=0,laneLine=log.read_text().strip(),fullOraclesRerun=False,reason='compiled source bytes are identical to the landed adapter')
(P/'scoped-results.json').write_text(json.dumps(result,indent=2)+'\n')
for log in (P/'logs').glob('scoped-*.log'):
 with (Path(str(log)+'.gz')).open('wb') as raw:
  with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=0) as out:out.write(log.read_bytes())
 log.unlink()
print('PASS scoped pins, unchanged source bytes, independent span negative and actual lane unit',flush=True)

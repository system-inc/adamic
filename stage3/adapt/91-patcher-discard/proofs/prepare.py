#!/usr/bin/env python3
"""Run bounded proofs serially before either upstream suite."""
import difflib
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

repo=Path(__file__).resolve().parents[4]
proof=Path(__file__).resolve().parent
scratch=Path('/workspace/scratch')
before=scratch/'adapt91-control'; after=scratch/'adapt91-after'; mutant=scratch/'adapt91-mutant'
logs=proof/'logs'; logs.mkdir(exist_ok=True)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),PYTHONDONTWRITEBYTECODE='1')
commands=[]
def run(name,cmd,limit=300,cwd=repo,expected=0,extra=None):
 start=time.monotonic()
 with (logs/(name+'.log')).open('w') as log:
  result=subprocess.run(list(map(str,cmd)),cwd=cwd,env=dict(env,**(extra or {})),stdout=log,stderr=subprocess.STDOUT,timeout=limit)
 commands.append(dict(name=name,command=list(map(str,cmd)),cwd=str(cwd),limit=limit,exit=result.returncode,seconds=round(time.monotonic()-start,3)))
 (proof/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 assert result.returncode==expected,(name,result.returncode,(logs/(name+'.log')).read_text()[-4000:])
run('idempotence',['node',proof.parent/'adapt.cjs',after],120)
assert json.loads((logs/'idempotence.log').read_text())['files']==0
run('references',['node',proof/'references.cjs',after,proof/'references.json'])
run('scope',['node',proof/'scope.cjs',before,after,scratch/'adapt91-scope.json'])
scope=json.loads((scratch/'adapt91-scope.json').read_text());assert len(scope['sites'])==2
(proof/'scope.json').write_text(json.dumps(scope,indent=2)+'\n')
patch=''
for site in scope['sites']:
 file=site['file'];a=(before/file).read_text();b=(after/file).read_text()
 patch+=''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/'+file,tofile='b/'+file,n=0))
(proof/'source.diff').write_text(patch)
# The sole semantic mutant restores nodeFactory's concise arrow, keeping utilities adapted.
shutil.copytree(after/'src',mutant/'src')
file=mutant/'src/compiler/factory/nodeFactory.ts'
a=file.read_bytes();old=b'fn => { fn(factory); }';assert a.count(old)==1
file.write_bytes(a.replace(old,b'fn => fn(factory)'))
run('mutant-scope',['node',proof/'scope.cjs',before,mutant,scratch/'adapt91-mutant-scope.json'])
m=json.loads((scratch/'adapt91-mutant-scope.json').read_text());assert m['changed']==['src/compiler/utilities.ts']
(proof/'mutant-scope.json').write_text(json.dumps(m,indent=2)+'\n')
emitdiff=''
for row in scope['sites']:
 name=Path(row['file']).name
 a=Path(str(scratch/'adapt91-scope.json')+'.'+name+'.before.js.txt').read_text()
 b=Path(str(scratch/'adapt91-scope.json')+'.'+name+'.after.js.txt').read_text()
 emitdiff+=''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before/'+name+'.js',tofile='after/'+name+'.js'))
(proof/'transpile.diff').write_text(emitdiff)
# Prove the exact source scope and movement guards fail for unrelated changes.
file=after/'src/compiler/core.ts';saved=file.read_bytes()
try:
 file.write_bytes(saved+b'\n// unrelated scope mutant\n')
 run('scope-mutant',['node',proof/'scope.cjs',before,after,scratch/'adapt91-invalid-scope.json'],expected=1)
finally: file.write_bytes(saved)
file=after/'src/compiler/utilities.ts';saved=file.read_bytes()
try:
 file.write_bytes(saved+b'\n')
 run('moved-target-mutant',['node',proof.parent/'adapt.cjs',after],120,expected=1)
 assert 'target source moved' in (logs/'moved-target-mutant.log').read_text()
finally: file.write_bytes(saved)
file=after/'src/testRunner/adapt91Registrant.ts'
try:
 file.write_text('import { addNodeFactoryPatcher } from "../compiler/factory/nodeFactory.js";\naddNodeFactoryPatcher(() => {});\n')
 run('registrant-mutant',['node',proof/'references.cjs',after,scratch/'adapt91-invalid-references.json'],expected=1)
 assert 'registrant or value alias exists' in (logs/'registrant-mutant.log').read_text()
finally: file.unlink()
run('overlay',['python3',repo/'stage3/census/latent/make_overlay.py',repo,scratch/'adapt91-latent-overlay'])
run('scan-overlay',['python3',proof/'make-scan.py',repo,scratch/'adapt91-latent-overlay/overlay.json',scratch/'adapt91-scan-overlay'])
run('scan-build',['go','build','-buildvcs=false','-overlay='+str(scratch/'adapt91-scan-overlay/overlay.json'),'-o',scratch/'adapt91-void-scan','./stage3/census/latent/tool'],300)
for label,tree in [('before',before),('after',after),('mutant',mutant)]:
 run('scan-'+label,[scratch/'adapt91-void-scan',tree/'src/compiler',scratch/('adapt91-scan-'+label+'.jsonl')],300,extra={'LATENT_ASSERT_NO_OUTPUT':'1'})
run('scan-compare',['python3',proof/'compare-scans.py',*[scratch/('adapt91-scan-'+label+'.jsonl') for label in ['before','after','mutant']],proof/'scan-delta.json'])
for label,tree in [('control',before),('after',after)]:
 run(label+'-install',['npm','ci','--no-audit','--no-fund'],300,cwd=tree)
 run(label+'-build',['npm','run','build'],600,cwd=tree)
run('original-hooks',['node',proof/'hooks.cjs',before/'built/local/typescript.js',proof/'hooks-before.json'],60)
run('hooks-before',['node',proof/'hooks-check.cjs',before/'built/local/typescript.js','[1,0,1,0]',proof/'hooks-control.json'],60)
run('hooks-after',['node',proof/'hooks-check.cjs',after/'built/local/typescript.js','[1,1,1,1]',proof/'hooks-after.json'],60)
run('original-hooks-after',['node',proof/'hooks.cjs',after/'built/local/typescript.js',scratch/'adapt91-invalid-hooks.json'],60,expected=1)
a=(before/'built/local/typescript.d.ts').read_bytes();b=(after/'built/local/typescript.d.ts').read_bytes()
assert a==b
run('api-compare',['python3',proof/'compare-api.py',before/'built/local/typescript.d.ts',after/'built/local/typescript.d.ts',proof/'api.json'],60)
api_mutant=scratch/'adapt91-api-mutant.d.ts.txt'
api_mutant.write_bytes(b+b'\ndeclare const unintendedApi: number;\n')
run('api-mutant',['python3',proof/'compare-api.py',before/'built/local/typescript.d.ts',api_mutant,scratch/'adapt91-invalid-api.json'],60,expected=1)
assert 'public typescript.d.ts bytes differ' in (logs/'api-mutant.log').read_text()
(proof/'api.json').write_text(json.dumps(dict(byteIdentical=True,bytes=len(a),sha256=hashlib.sha256(a).hexdigest(),extraDeclarationMutantCaught=True),indent=2)+'\n')
(proof/'preparation.json').write_text(json.dumps(dict(status='pass',mutants=['one-site reversion','unrelated source','moved target','registrant','original early-stop witness on adapted output','extra API declaration']),indent=2)+'\n')
print('PASS serial preparation')

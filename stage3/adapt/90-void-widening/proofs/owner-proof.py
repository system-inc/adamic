#!/usr/bin/env python3
"""Replay bounded 90 proofs from official apply snapshots. No full latent claim."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tarfile
import time
sys.dont_write_bytecode = True
repo = Path(__file__).resolve().parents[4]
proof = Path(__file__).resolve().parent
control, scratch, scanner = map(lambda p: Path(p).resolve(), sys.argv[1:])
scratch.mkdir(parents=True, exist_ok=False)
tree = scratch / 'tree'
tree.mkdir()
env = dict(os.environ, NODE_PATH=str(repo/'stage3/api/node_modules'), LATENT_ASSERT_NO_OUTPUT='1')
commands = []
def run(name, command, cwd=repo, limit=600):
    started = time.monotonic()
    with (scratch/(name+'.log')).open('w') as log:
        result = subprocess.run(list(map(str, command)), cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT, timeout=limit)
    commands.append(dict(name=name, command=list(map(str, command)), cwd=str(cwd), exit=result.returncode, limit=limit, seconds=round(time.monotonic()-started, 3)))
    (proof/'owner-proof-commands.json').write_text(json.dumps(commands, indent=2)+'\n')
    assert result.returncode == 0, name
    return (scratch/(name+'.log')).read_text()
with (scratch/'before.tar').open('wb') as archive:
    subprocess.run(['git','-C',str(control),'archive','refs/stage3/90-void-widening-before'], stdout=archive, check=True)
with tarfile.open(scratch/'before.tar') as archive:
    archive.extractall(tree, filter='data')
run('install', ['npm','ci','--no-audit','--no-fund'], tree)
sys.path.insert(0, str(repo/'stage3/adapt/91-patcher-discard/proofs'))
from extract import extract
def scan(label):
    output = scratch/(label+'.jsonl')
    run(label, [scanner, tree/'src/compiler', output])
    records = [json.loads(line) for line in output.read_text().splitlines()]
    assert len(records[1:]) == 79
    assert all(f['kind'] != 'panic' for record in records for f in record.get('findings') or [])
    return len(extract(output))
assert scan('scan-before') == 116
core = tree/'src/compiler/core.ts'
saved = core.read_bytes()
try:
    core.write_bytes(saved+b'\nexport function ownerScopedVoidMutant(): void | number { return ownerScopedVoidProducer(); }\nfunction ownerScopedVoidProducer(): void {}\n')
    assert scan('scan-planted-mutant') == 117
finally:
    core.write_bytes(saved)
checker = """const ts=require('typescript'),path=require('path');const {compiler}=require(process.argv[1]);const tree=process.argv[2];console.log(JSON.stringify(ts.getPreEmitDiagnostics(compiler(tree,ts)).map(d=>({file:d.file?path.relative(tree,d.file.fileName):'',code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\\n'),position:d.file&&d.start!==undefined?d.file.getLineAndCharacterOfPosition(d.start):null}))))"""
def diagnostics(label):
    return json.loads(run(label, ['node','-e',checker,repo/'stage3/adapt/47-host-errors/adapt.cjs',tree]))
before_diagnostics = diagnostics('checker-before')
run('build-before', ['npm','run','build'], tree, 900)
names = [row['file'] for row in json.loads((proof/'source-api-proof.json').read_text())['artifacts']]
artifacts = {name:(tree/'built/local'/name).read_bytes() for name in names}
manifest = json.loads((proof.parent/'targets.json').read_text())
originals = {}
for entry in manifest['files']:
    destination = scratch/'original'/entry['file']
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_bytes((tree/entry['file']).read_bytes())
    originals[entry['file']] = destination
run('adapt', ['node',proof.parent/'adapt.cjs',tree])
assert scan('scan-after') == 5
binder = tree/'src/compiler/binder.ts'
saved = binder.read_bytes()
try:
    needle = b'function bindFunctionOrConstructorType(node: SignatureDeclaration | JSDocSignature): undefined'
    assert saved.count(needle) == 1
    binder.write_bytes(saved.replace(needle,needle.replace(b': undefined',b': void')))
    assert scan('scan-binder-reversion') == 6
finally:
    binder.write_bytes(saved)
after_diagnostics = diagnostics('checker-after')
assert before_diagnostics == after_diagnostics
saved = core.read_bytes()
try:
    core.write_bytes(saved+b'\nexport const ownerScopedNumberMutant: number = "not a number";\n')
    mutant_diagnostics = diagnostics('checker-mutant')
    assert len(mutant_diagnostics) == len(after_diagnostics)+1
    assert sum(d['code']==2322 for d in mutant_diagnostics) == sum(d['code']==2322 for d in after_diagnostics)+1
finally:
    core.write_bytes(saved)
erasure = """const ts=require('typescript'),fs=require('fs'),assert=require('assert/strict');const a=fs.readFileSync(process.argv[1],'utf8'),b=fs.readFileSync(process.argv[2],'utf8');for(const removeComments of [false,true]){const emit=text=>ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;assert.equal(emit(a),emit(b));assert.notEqual(emit(a),emit(b+'\\nvoid 0;'));}console.log('PASS identity and runtime mutant in both comment modes');"""
for index, (file, original) in enumerate(originals.items()):
    run('erasure-'+str(index), ['node','-e',erasure,original,tree/file])
shutil.rmtree(tree/'built')
run('build-after', ['npm','run','build'], tree, 900)
for name, data in artifacts.items():
    assert (tree/'built/local'/name).read_bytes() == data, name
assert json.loads(run('rerun', ['node',proof.parent/'adapt.cjs',tree]))['files'] == 0
result = dict(status='pass', diagnostics=dict(before=len(before_diagnostics), after=len(after_diagnostics), mutant=len(mutant_diagnostics)), sourceScan=dict(before=116, after=5, plantedMutant=117, binderReversion=6, fullLatentClaim=False), erasureFiles=len(originals), commentModes=[False, True], runtimeMutantsCaught=len(originals), artifacts=[dict(file=name, bytes=len(data), sha256=hashlib.sha256(data).hexdigest(), byteIdentical=True) for name,data in artifacts.items()])
(proof/'owner-proof-results.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))

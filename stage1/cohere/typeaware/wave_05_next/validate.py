#!/usr/bin/env python3
"""Independent wave-owned driver. Every subprocess writes directly to a log file."""
import gzip
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import statistics
import subprocess
import sys
import time

REPOSITORY = Path(__file__).resolve().parents[4]
HERE = Path(__file__).resolve().parent
ARTIFACTS = Path(os.environ.get('WAVE05_NEXT_ARTIFACTS', '/workspace/wave-05-next-validation'))
ARTIFACTS.mkdir(parents=True, exist_ok=True)
SEQUENCE = 0

def run(name, command, cwd=REPOSITORY, environment=None, expected=0):
    global SEQUENCE
    SEQUENCE += 1
    prefix = ARTIFACTS / f'{SEQUENCE:03d}-{name}'
    started = time.monotonic()
    with open(str(prefix)+'.stdout', 'wb') as output, open(str(prefix)+'.stderr', 'wb') as error:
        result = subprocess.run([str(arg) for arg in command], cwd=cwd, env=environment, stdout=output, stderr=error)
    elapsed = time.monotonic()-started
    stdout = Path(str(prefix)+'.stdout').read_bytes()
    stderr = Path(str(prefix)+'.stderr').read_bytes()
    assert result.returncode == expected, (name, result.returncode, stderr.decode(errors='replace'))
    return stdout, stderr, elapsed

def compare(name, config, manifest, executable):
    truth, _, _ = run(name+'-go', [oracle, config, manifest])
    actual, stderr, _ = run(name+'-native', [executable, config, manifest])
    assert not stderr, (name, stderr)
    assert actual == truth, (name, next((i for i, pair in enumerate(zip(actual, truth)) if pair[0] != pair[1]), min(len(actual), len(truth))))
    print(name, 'identical bytes', len(truth), 'sha256', hashlib.sha256(truth).hexdigest(), truth.splitlines()[-1].decode(), flush=True)
    (ARTIFACTS/(name+'.oracle.gz')).write_bytes(gzip.compress(truth, mtime=0))
    return truth

stage0 = ARTIFACTS/'adamic'
run('stage0', ['go', 'build', '-o', stage0, './cmd/adamic'])
archive = ARTIFACTS/'checker.a'
run('archive', ['go', 'build', '-buildmode=c-archive', '-o', archive, './bridge/tsgo/archive'])
native = ARTIFACTS/'native'
run('native-build', [stage0, 'build', HERE/'main.a', '-o', native, '--tsgo', archive])
virtual = REPOSITORY/'cohere/adamic_wave05_next_oracle.go'
overlay = ARTIFACTS/'oracle-overlay.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(HERE/'testdata/oracle.go')}}))
oracle = ARTIFACTS/'oracle'
run('oracle-build', ['go', 'build', '-overlay', overlay, '-o', oracle, virtual], REPOSITORY/'cohere')

prelude = 'declare function work():Promise<string>; declare function use(x:unknown):void; declare const ms:number;\n'
controls = [
    'Promise.race([work(),new Promise((r,j)=>{setTimeout(j,ms);})]);',
    'Promise.race([work(),new Promise((r,j)=>setTimeout(j,ms))]);',
    'const timeout=new Promise((r,j)=>{setTimeout(j,ms);});Promise.race([work(),timeout]);',
    'Promise.race([work(),new Promise((r,j)=>{const timer=setTimeout(j,ms);})]);',
    'let timer;Promise.race([work(),new Promise((r,j)=>{timer=setTimeout(j,ms);})]);',
    'let timer;Promise.race([work(),new Promise((r,j)=>{timer=setTimeout(j,ms);})]);clearTimeout(timer);',
    'Promise.race([work(),new Promise((r,j)=>{globalThis.setTimeout(j,ms);})]);',
    'Promise.race([work(),new Promise((r,j)=>{void setTimeout(j,ms);})]);',
    'Promise.race([new Promise((r)=>{setTimeout(r,ms);}),new Promise((r,j)=>{setTimeout(j,2*ms);})]);',
    'Promise.race([new Promise((r,j)=>{use(setTimeout(j,ms));})]);',
    'Promise.race([new Promise((r,j)=>{const timer=setTimeout(j,ms);use({timer});})]);',
    'Promise.race([new Promise((r,j)=>{work().then(()=>{setTimeout(j,ms);});})]);',
    'Promise.all([new Promise((r,j)=>{setTimeout(j,ms);})]);',
    'function setTimeout(f:unknown,ms:number){return 0;}Promise.race([new Promise((r,j)=>{setTimeout(j,ms);})]);',
    'const pool={race(a:unknown){return a;}};pool.race([new Promise((r,j)=>{setTimeout(j,ms);})]);',
    'let timeout=new Promise((r,j)=>{setTimeout(j,ms);});Promise.race([work(),timeout]);',
    'Promise.race([new Promise((r,j)=>{const timer=setTimeout(j,ms);const f=(timer:number)=>timer;})]);',
    '/* 世界 🌍 */\r\nPromise.race([new Promise((r,j)=>{ (setTimeout(j,ms)); })]);\r\n',
    'Promise.race([new Promise((r,j)=>{return setTimeout(j,ms);})]);',
    'let a,b;Promise.race([new Promise((r,j)=>{a=b=setTimeout(j,ms);})]);',
    'Promise.race([new Promise((r,j)=>{class C {m(){setTimeout(j,ms);}}})]);',
    'Promise.race([new Promise((r,j)=>{const t=setTimeout(j,ms);t=0;})]);',
]
paths=[]
for index, text in enumerate(controls):
    path=ARTIFACTS/f'control-{index:03d}.a';path.write_text(prelude+text+'\nexport {};\n');paths.append(str(path))
manifest=ARTIFACTS/'controls.manifest';manifest.write_text('\n'.join(paths)+'\n')
dom=ARTIFACTS/'dom.json';dom.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022','DOM']},'files':[paths[0]]}))
node=ARTIFACTS/'node.json'
platform=ARTIFACTS/'timers.d.ts'
platform.write_text('export {};\ndeclare global { namespace NodeJS { interface Timeout { unref():this; } } function setTimeout(f:(...a:any[])=>void, ms?:number):NodeJS.Timeout; namespace setTimeout { const extra:unknown; } function clearTimeout(t:unknown):void; }\n')
node.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022']},'files':[str(platform)]}))
truth=compare('dom-controls', dom, manifest, native)
assert b'unclearedRaceTimeout' in truth, 'no positive control'
compare('node-controls', node, manifest, native)

sanitized_archive=ARTIFACTS/'sanitized.a';environment=dict(os.environ, CC='clang', CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
run('asan-archive', ['go','build','-buildmode=c-archive','-o',sanitized_archive,'./bridge/tsgo/archive'], environment=environment)
sanitized=ARTIFACTS/'sanitized'
run('asan-build',[stage0,'build',HERE/'main.a','-o',sanitized,'--tsgo',sanitized_archive,'--sanitize'])
compare('dom-asan',dom,manifest,sanitized)
compare('node-asan',node,manifest,sanitized)

mutant_source=ARTIFACTS/'mutant-source';mutant_source.mkdir(exist_ok=True)
for file in HERE.glob('*.a'):
    text=file.read_text()
    if file.name=='no_uncleared_race_timeout.a':
        pattern=r'this\.lost\(executor,\s*index\)'
        assert len(re.findall(pattern,text))==1
        text=re.sub(pattern,'!this.lost(executor, index)',text)
    text=text.replace("'../", "'"+str(HERE.parent)+'/').replace("'../../../typescript/", "'"+str(REPOSITORY/'stage1/typescript')+'/')
    (mutant_source/file.name).write_text(text)
mutant=ARTIFACTS/'mutant'
run('mutant-build',[stage0,'build',mutant_source/'main.a','-o',mutant,'--tsgo',archive])
wrong, stderr, _=run('mutant-run',[mutant,dom,manifest]);assert not stderr and wrong!=truth
print('timer mutant exits 0, empty stderr; independent bytes catch it',flush=True)

released=ARTIFACTS/'released'
run('released-build',[stage0,'build',HERE/'testdata/released.a','-o',released,'--tsgo',archive])
probe=ARTIFACTS/'released-probe.a';probe.write_text('clock;\n')
_, stderr, _=run('released-run',[released,dom,probe],expected=70)
assert stderr==b'adamic: panic: invalid or released checker handle\n'
print('released symbol-ancestry handle: exit 70, exact refusal',flush=True)

_, refusal, _=run('cfg-refusal',[stage0,'build',HERE/'gaps/cfg_probe.a','-o',ARTIFACTS/'cfg-probe','--tsgo',archive],expected=1)
assert b'bindings.ts:49:33' in refusal and b'escaping a constructor before every field is set' in refusal
print('CFG reuse blocked: stage 0 constructor refusal at shared bindings.ts:49:33',flush=True)

corpora=[('repository',REPOSITORY/'tsconfig.json',os.environ.get('ADAMIC_WAVE05_REPOSITORY_MANIFEST')),('compiler',Path(os.environ.get('ADAMIC_TYPESCRIPT_SOURCE','/workspace/wave-05-typescript'))/'src/compiler/tsconfig.json',os.environ.get('ADAMIC_WAVE05_COMPILER_MANIFEST'))]
for name,config,paths in corpora:
    if paths:
        compare(name,config,paths,native);compare(name+'-asan',config,paths,sanitized)
for name,config,paths in corpora:
    if not paths or os.environ.get("WAVE05_NEXT_SKIP_BENCH"):continue
    medians={}
    for round_ in range(3):
        for implementation in ([oracle,native] if round_%2==0 else [native,oracle]):
            label='go' if implementation==oracle else 'native'
            stdout, stderr, elapsed=run(f'{name}-{label}-round-{round_}',[implementation,config,paths,'--count'],environment=dict(os.environ,ADAMIC_TSGO_TIMING='1'))
            medians.setdefault(label,[]).append(elapsed)
            print(name,label,'round',round_,f'{elapsed:.6f}',stdout.strip().decode(),stderr.strip().decode(),flush=True)
    print(name,'medians',json.dumps({key:statistics.median(value) for key,value in medians.items()}),flush=True)
print('PASS timer comparison, mutant and sanitizers',flush=True)

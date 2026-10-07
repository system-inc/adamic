#!/usr/bin/env python3
"""Rule-local comparison, never edits the shared lint harness."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

repo = Path(__file__).resolve().parents[4]
source = Path(__file__).resolve().parent
artifacts = Path(sys.argv[1]).resolve()
artifacts.mkdir(parents=True, exist_ok=True)
compiler = artifacts / 'adamic'
commands = []

def run(name, argv, cwd=repo, expected=0):
    begin = time.monotonic_ns()
    with (artifacts / (name + '.stdout')).open('wb') as out, (artifacts / (name + '.stderr')).open('wb') as err:
        result = subprocess.run([str(x) for x in argv], cwd=cwd, stdout=out, stderr=err)
    commands.append({'name': name, 'argv': [str(x) for x in argv], 'exit': result.returncode, 'ns': time.monotonic_ns()-begin})
    if result.returncode != expected:
        raise RuntimeError(f'{name} exit {result.returncode}, expected {expected}; see {artifacts / (name + ".stderr")}')
    return (artifacts / (name + '.stdout')).read_bytes()

run('stage0', ['go', 'build', '-o', compiler, './cmd/adamic'])
archive = artifacts / 'checker.a'
run('archive', ['go', 'build', '-buildmode=c-archive', '-o', archive, './bridge/tsgo/archive'])
fixtures = json.loads(run('extract', ['go', 'run', source / 'testdata/extract.go', repo]))
fixtures += [
 {'rule': 'id-denylist', 'source': '/* 世界 🌍 */\r\nconst é=1;let obj={};obj.é+=é;\r\n', 'names': ['é'], 'pattern': '', 'flags': {}},
 {'rule': 'id-match', 'source': '/* 世界 🌍 */\r\nconst bad_é=1; class C {#bad_é=1;method(){return this.#bad_é;}}\r\n', 'names': [], 'pattern': '^[^_]+$', 'flags': {'classFields': True}},
]
(artifacts / 'fixtures.json').write_text(json.dumps(fixtures, ensure_ascii=False, indent=2)+'\n')
paths = []
profiles = []
for i, fixture in enumerate(fixtures):
    path = artifacts / f'input-{i:03d}.a'
    # Match the production fixture sourceType:module; imports and globals do not
    # accidentally merge between unrelated cases in this one checker program.
    path.write_text(fixture['source'] + '\nexport {};\n')
    paths.append(str(path))
    flags = ''.join(letter for key, letter in [('properties','p'),('classFields','c'),('onlyDeclarations','o'),('ignoreDestructuring','i')] if fixture['flags'].get(key, False))
    profiles.append(fixture['rule']+'\t'+(','.join(fixture['names'] or []) if fixture['rule']=='id-denylist' else fixture['pattern'])+'\t'+flags)
(artifacts / 'manifest').write_text('\n'.join(paths)+'\n')
(artifacts / 'profiles').write_text('\n'.join(profiles)+'\n')
(artifacts / 'tsconfig.json').write_text(json.dumps({'compilerOptions': {'strict': True, 'target': 'ESNext', 'module': 'NodeNext', 'moduleResolution': 'NodeNext'}, 'files': paths}))
virtual = repo / 'cohere/adamic_wave29_configured_oracle.go'
(artifacts / 'overlay.json').write_text(json.dumps({'Replace': {str(virtual): str(source / 'testdata/oracle.go')}}))
oracle = artifacts / 'oracle'
run('oracle-build', ['go','build','-overlay',artifacts / 'overlay.json','-o',oracle,virtual], cwd=repo / 'cohere')
run('native-build',[compiler,'build', source / 'runner.a','-o',artifacts / 'native','--tsgo',archive])
args = [artifacts / 'tsconfig.json', artifacts / 'manifest']
truth = run('go',[oracle,*args,artifacts / 'fixtures.json'])
native = run('native',[artifacts / 'native',*args,artifacts / 'profiles'])
if truth != native:
    a=truth.splitlines(); b=native.splitlines()
    for i,(x,y) in enumerate(zip(a,b)):
        if x!=y:
            print('first differing line',i, 'Go',x, 'native',y)
            print('Go nearby',a[max(0,i-2):i+3]); print('native nearby',b[max(0,i-2):i+3]);break
    raise RuntimeError('configured byte mismatch')
print('configured inputs',len(fixtures),'bytes',len(truth),'sha256',hashlib.sha256(truth).hexdigest(),truth.splitlines()[-1].decode())
# A comparison mutant must compile and exit normally with no stderr.
for name, filename, before, after in [
    ('denylist','id_denylist.a',"return !this.global(index);",'return false;'),
    ('match','id_match.a',"if(!this.check(index, flags) || matches(name)) { continue; }",'if(matches(name)) { continue; }'),
]:
    directory=artifacts / (name+'-mutant'); directory.mkdir(exist_ok=True)
    for module in ['id_denylist.a','id_match.a']:
        text=(source.parent / module).read_text()
        if module==filename:
            assert text.count(before)==1
            text=text.replace(before,after)
        for dependency in ['rules.ts','diagnostic.ts','symbol_provenance.a','regexp_program.a']:
            text=text.replace("'./"+dependency+"'", "'"+str(source.parent / dependency)+"'")
        (directory / module).write_text(text)
    text=(source / 'runner.a').read_text()
    for dependency in ['unary_minus.ts','rules.ts']:
        text=text.replace("'../"+dependency+"'", "'"+str(source.parent / dependency)+"'")
    for dependency in ['id_denylist.a','id_match.a']:
        text=text.replace("'../"+dependency+"'", "'./"+dependency+"'")
    text=text.replace("'../../../typescript/", "'"+str(repo / 'stage1/typescript')+'/')
    (directory / 'runner.a').write_text(text)
    binary=artifacts / name
    run(name+'-build',[compiler,'build',directory/'runner.a','-o',binary,'--tsgo',archive])
    got=run(name,[binary,*args,artifacts/'profiles'])
    assert got!=truth and not (artifacts / (name+'.stderr')).read_bytes(), name+' mutant survived'
    print(name,'mutant compiled, exit 0, empty stderr; bytes differ')
# Instrument both the native compiler and the checker library.
oldcc=os.environ.get('CC'); oldflags=os.environ.get('CGO_CFLAGS')
os.environ['CC']='clang'; os.environ['CGO_CFLAGS']='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'
run('asan-archive',['go','build','-buildmode=c-archive','-o',artifacts/'asan-checker.a','./bridge/tsgo/archive'])
if oldcc is None: os.environ.pop('CC',None)
else: os.environ['CC']=oldcc
if oldflags is None: os.environ.pop('CGO_CFLAGS',None)
else: os.environ['CGO_CFLAGS']=oldflags
run('asan-build',[compiler,'build',source/'runner.a','-o',artifacts/'asan-native','--tsgo',artifacts/'asan-checker.a','--sanitize'])
os.environ['ASAN_OPTIONS']='detect_leaks=1:halt_on_error=1'
os.environ['UBSAN_OPTIONS']='halt_on_error=1'
asan=run('asan',[artifacts/'asan-native',*args,artifacts/'profiles'])
assert asan==truth and not (artifacts/'asan.stderr').read_bytes(), 'sanitizer mismatch'
print('ASan/UBSan/LeakSanitizer bytes agree; empty stderr')
run('regexp-gap',[compiler,'build',source/'gaps/regexp.a','-o',artifacts/'regexp-gap'],expected=1)
assert "stage 0 can't lower new an Identifier yet" in (artifacts/'regexp-gap.stderr').read_text()
print('regexp gap reproduced: compile exit 1')
(artifacts/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')

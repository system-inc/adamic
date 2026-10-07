"""Native listener agreement using explicit Go-captured numeric syntax inputs."""
from pathlib import Path
import hashlib
import json
import os
import re
import shutil
import subprocess
import time

repo = Path('/workspace/adamic'); source = repo/'stage1/cohere/typeaware/wave_01_fourth'; out = Path('/workspace/wave-01-fourth-validation');out.mkdir(exist_ok=True)
records = []
def run(name, command, cwd=repo, env=None, expected=0):
    started=time.perf_counter_ns()
    with (out/(name+'.stdout')).open('wb') as stdout, (out/(name+'.stderr')).open('wb') as stderr:
        p=subprocess.run(list(map(str,command)),cwd=cwd,env={**os.environ,**(env or {})},stdout=stdout,stderr=stderr)
    data=(out/(name+'.stdout')).read_bytes();error=(out/(name+'.stderr')).read_bytes()
    records.append({'name':name,'command':list(map(str,command)),'exit':p.returncode,'wall_ns':time.perf_counter_ns()-started,'sha256':hashlib.sha256(data).hexdigest(),'stderr':error.decode(errors='replace')});(out/'records.json').write_text(json.dumps(records,indent=2)+'\n')
    if p.returncode!=expected:raise RuntimeError((name,p.returncode,error.decode()))
    return data,error
virtual=repo/'cohere/adamic_wave01fourth.go';overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/oracle.go')}}));oracle=out/'oracle'
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],cwd=repo/'cohere')
cases=['Symbol();','let x = Symbol();','Symbol("Foo");','new Symbol();','globalThis.Symbol();','Symbol(undefined);','Symbol(...[]);','(Symbol)();','((Symbol))();','var Symbol = function() {}; Symbol();','function f(Symbol:any){Symbol();}','function f(){let Symbol=()=>0;Symbol();}','const x={Symbol(){}};x.Symbol();','Symbol?.();','const s="😀"; Symbol();','Symbol.for("x");','Symbol["for"]("x");']
paths=[]
for i,text in enumerate(cases):
    p=out/('control-%02d.a'%i);p.write_text(text+'\nexport {};\n');paths.append(p)
manifest=out/'controls.manifest';manifest.write_text('\n'.join(map(str,paths))+'\n');config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'NodeNext','noEmit':True},'files':list(map(str,paths)),'sourceExtensions':['.a']}))
binary=Path('/workspace/wave-01-fourth-symbol')
truths={}
def compare(name,binary,config,manifest):
    packet=out/(name+'.nodes');data,_=run(name+'-nodes',[oracle,config,manifest,'--nodes']);packet.write_bytes(data)
    want,_=run(name+'-go',[oracle,config,manifest]);got,error=run(name+'-native',[binary,config,manifest,packet])
    assert not error and got==want,(name,got[:100],want[:100],error)
    truths[name]=(config,manifest,packet,want)
    print(name,got.splitlines()[-1].decode(),len(got),'identical bytes',flush=True)
compare('controls',binary,config,manifest)
for name,base,cfg in [('compiler',Path('/workspace/wave-01-typescript'),Path('/workspace/wave-01-typescript/src/compiler/tsconfig.json')),('repository',repo,repo/'tsconfig.json')]:
    manifest=out/(name+'.manifest');manifest.write_text('\n'.join(str(base/line) for line in (repo/('stage1/cohere/typeaware/validation-coverage/'+name+'.manifest')).read_text().splitlines() if line)+'\n');compare(name,binary,cfg,manifest)
archive=out/'checker-asan.a';run('asan-archive',['go','build','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'],env={'CC':'clang','CGO_CFLAGS':'-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'})
asan=out/'symbol-asan';run('asan-build',['/workspace/wave-01-fourth-adamic','build',source/'testdata/symbol_suite.a','-o',asan,'--tsgo',archive,'--sanitize'])
for name,(cfg,manifest,packet,want) in list(truths.items()):
    got,error=run(name+'-asan',[asan,cfg,manifest,packet]);assert got==want and not error;print(name,'sanitizers PASS',flush=True)
mutantdir=out/'mutant';mutantdir.mkdir(exist_ok=True)
for original in (source/'symbol_description').glob('*.a'):
    text=original.read_text()
    if original.name=='rule.a':
        assert text.count('shape.arguments !== 0')==1;text=text.replace('shape.arguments !== 0','shape.arguments === 0')
    text=re.sub(r"from '([^']+)'",lambda m:m.group(0) if m.group(1)=='adamic' or m.group(1).startswith('./') else "from '%s'"%(original.parent/m.group(1)).resolve(),text);(mutantdir/original.name).write_text(text)
suite=(source/'testdata/symbol_suite.a').read_text();suite=re.sub(r"from '([^']+)'",lambda m:m.group(0) if m.group(1)=='adamic' else "from '%s'"%((mutantdir/'rule.a') if m.group(1)=='../symbol_description/rule.a' else (source/'testdata'/m.group(1)).resolve()),suite);entry=out/'mutant-suite.a';entry.write_text(suite)
mutant=out/'symbol-mutant';run('mutant-build',['/workspace/wave-01-fourth-adamic','build',entry,'-o',mutant,'--tsgo','/workspace/wave-01-fourth-checker.a'])
cfg,manifest,packet,want=truths['controls'];got,error=run('mutant-run',[mutant,cfg,manifest,packet]);assert not error and got!=want;print('symbol mutant exits 0; empty stderr; byte comparison catches',flush=True)
probe=out/'released.a';probe.write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,8,'CallExpression','call-symbol-shape'));\n")
released=out/'released';run('released-build',['/workspace/wave-01-fourth-adamic','build',probe,'-o',released,'--tsgo','/workspace/wave-01-fourth-checker.a']);data,error=run('released-run',[released,config,paths[0]],expected=70);assert error==b'adamic: panic: invalid or released checker handle\n';print('released question panic 70 PASS',flush=True)
print('PASS symbol listener only; native parser/driver integration remains blocked',flush=True)

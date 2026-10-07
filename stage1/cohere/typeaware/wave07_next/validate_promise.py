#!/usr/bin/env python3
"""Exercise a private registration overlay; never edit the shared dispatcher."""
import argparse, json, os, pathlib, subprocess, time, re

parser = argparse.ArgumentParser()
parser.add_argument('artifacts', type=pathlib.Path)
parser.add_argument('--compiler', type=pathlib.Path, required=True)
args = parser.parse_args()
root = pathlib.Path(__file__).resolve().parents[4]
unit = pathlib.Path(__file__).resolve().parent
out = args.artifacts.resolve()
out.mkdir(parents=True, exist_ok=True)
records = []

def run(name, command, cwd=root, environment=None, expected=0):
    started = time.perf_counter_ns()
    with (out/(name+'.stdout')).open('wb') as stdout, (out/(name+'.stderr')).open('wb') as stderr:
        result = subprocess.run([str(x) for x in command], cwd=cwd, env=environment, stdout=stdout, stderr=stderr)
    elapsed = time.perf_counter_ns()-started
    if result.returncode != expected:
        raise RuntimeError(f'{name}: exit {result.returncode}, expected {expected}; see saved stderr')
    record = dict(name=name, command=[str(x) for x in command], process_ns=elapsed, exit=result.returncode)
    records.append(record)
    (out/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
    return (out/(name+'.stdout')).read_bytes(), (out/(name+'.stderr')).read_bytes()

def overlay(name, replacements):
    path=out/(name+'.json')
    path.write_text(json.dumps({'Replace':{str(k):str(v) for k,v in replacements.items()}}))
    return path

run('stage0',['go','build','-o',out/'adamic','./cmd/adamic'])
run('archive',['go','build','-buildmode=c-archive','-o',out/'checker.a','./bridge/tsgo/archive'])
entry=unit/'promise_suite.a'
run('native-build',[out/'adamic','build',entry,'-o',out/'timer','--tsgo',out/'checker.a'])
virtual=root/'cohere/adamic_wave07_timer_oracle.go'
oracle_overlay=overlay('oracle',{virtual:unit/'testdata/oracle_promise.go'})
run('oracle-build',['go','build','-overlay',oracle_overlay,'-o',out/'oracle',virtual],root/'cohere')
prelude=out/'ambient.d.ts'
prelude.write_text('export {};')
config=out/'tsconfig.json'
config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['es2022']},'files':['ambient.d.ts']}))
common='declare let error:Error;declare const flag:boolean;declare function use(x:unknown):void;\n'
controls=['Promise.reject(5);', 'Promise.reject();', 'Promise.reject(undefined);', 'Promise.reject(new Error());', 'Promise.reject(error);', 'Promise.reject({code:1});', 'Promise.reject(error&&5);', 'Promise.reject(5&&error);', 'Promise.reject(flag?5:error);', 'Promise.reject(flag?5:6);', 'Promise.reject((5,error));', 'Promise.reject((error,5));', 'Promise.reject((error as Error));', 'Promise["reject"](5);', '(Promise?.reject)(5);', 'Promise[`reject`](5);', 'function f(Promise:{reject(x:unknown):void}){Promise.reject(5);}', 'function f(undefined:unknown){Promise.reject(undefined);}', 'new Promise((resolve,reject)=>reject(5));', 'new Promise((resolve,reject)=>{function nested(){reject(5);}});', 'new Promise((resolve,reject)=>{function nested(reject:(v:unknown)=>void){reject(5);}});', 'new Promise((resolve,reject)=>{const reject=(x:unknown)=>{};reject(5);});', 'new Promise(function(reject,reject){reject(5);});', 'new Promise((resolve,{apply})=>apply(5));', 'new Promise(({foo},reject)=>reject(5));', 'new Promise((resolve,reject,x=reject(5))=>{});', 'new Promise((resolve,reject)=>resolve(5,reject));', 'Promise.reject([1]);', 'Promise.reject(()=>5);', 'Promise.reject(1+2);', 'Promise.reject(error||5);', 'Promise.reject(error??5);', 'Promise.reject(error=5);', 'Promise.reject(error&&=5);', 'Promise.reject(error||=5);', '/*世界 🌍*/\r\nPromise.reject("bad");', 'Promise.reject((undefined));', 'new Promise();', 'new Promise(()=>{});', 'new (Promise)((r,j)=>j(5));']
paths=[]
for i,source in enumerate(controls):
    path=out/f'control-{i:03d}.a';path.write_text(common+source+'\nexport {};\n');paths.append(path)
manifest=out/'controls.manifest';manifest.write_text(''.join(str(p)+'\n' for p in paths))

def compare(name, native, config, manifest):
    truth,go_error=run(name+'-go',[out/'oracle',config,manifest])
    actual,native_error=run(name+'-native',[native,config,manifest])
    if truth!=actual or native_error: raise RuntimeError(f'{name}: diagnostic disagreement or native stderr')
    print(f'{name}: {len(actual)} identical bytes, {actual.splitlines()[-1].decode()}',flush=True)
    return truth

truth=compare('controls',out/'timer',config,manifest)
if b'\tprefer-promise-reject-errors\t' not in truth: raise RuntimeError('no positive control')
# DOM-only declarations exercise window and library member merging independently.
(out/'dom-prelude.d.ts').write_text('export {};')
dom_config=out/'dom.json';dom_config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['es2022','dom']},'files':['dom-prelude.d.ts']}))
compare('controls-dom',out/'timer',dom_config,manifest)
environment=dict(os.environ,CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
run('archive-asan',['go','build','-buildmode=c-archive','-o',out/'checker-asan.a','./bridge/tsgo/archive'],environment=environment)
run('native-asan-build',[out/'adamic','build',entry,'-o',out/'timer-asan','--tsgo',out/'checker-asan.a','--sanitize'])
compare('controls-asan',out/'timer-asan',config,manifest)
for name,corpus_root,corpus_config in [('repository',root,root/'tsconfig.json'),('compiler',args.compiler,args.compiler/'src/compiler/tsconfig.json')]:
    lines=(root/f'stage1/cohere/typeaware/validation-volume/{name}.manifest').read_text().splitlines()
    corpus_manifest=out/(name+'.manifest');corpus_manifest.write_text(''.join(str(corpus_root/line)+'\n' for line in lines))
    compare(name,out/'timer',corpus_config,corpus_manifest)
    compare(name+'-asan',out/'timer-asan',corpus_config,corpus_manifest)
mutant=out/'mutant.a';rule=(unit/'prefer_promise_reject_errors.a').read_text();token="'rejectAnError'"
if rule.count(token)!=1: raise RuntimeError('nonunique mutant')
rule=rule.replace(token,"'rejectAnErrorMutant'").replace("'../","'"+str(unit.parent)+"/").replace("'../../../typescript/","'"+str(root/'stage1/typescript')+"/").replace("'./helpers.a'","'"+str(unit/'helpers.a')+"'")
mutant.write_text(rule)
mutant_entry=out/'mutant-suite.a';source=entry.read_text().replace("'../","'"+str(unit.parent)+"/").replace("'../../../typescript/","'"+str(root/'stage1/typescript')+"/").replace("'./prefer_promise_reject_errors.a'","'"+str(mutant)+"'")
mutant_entry.write_text(source)
run('mutant-build',[out/'adamic','build',mutant_entry,'-o',out/'timer-mutant','--tsgo',out/'checker.a'])
changed,errors=run('mutant',[out/'timer-mutant',config,manifest])
if errors or truth==changed: raise RuntimeError('rule mutant survived')
print('rule mutant: exit 0, empty stderr; caught only by independent Go bytes',flush=True)
print('PASS normal-archive gate',flush=True)

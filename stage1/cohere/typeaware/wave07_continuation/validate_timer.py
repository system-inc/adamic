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

facts=root/'bridge/tsgo/checker/facts.go'
private=out/'facts.go'
source=facts.read_text()
if source.count('switch mode {') != 1: raise RuntimeError('nonunique dispatcher')
private.write_text(source.replace('switch mode {','switch mode {\n case "wave07-symbol-context": return p.wave07SymbolContext(c,node,question)'))
registration=overlay('registration',{facts:private})
run('question-test',['go','test','./bridge/tsgo/checker','-run','^TestWave07SymbolContext$','-count=1','-v'])
run('stage0',['go','build','-o',out/'adamic','./cmd/adamic'])
run('archive',['go','build','-overlay',registration,'-buildmode=c-archive','-o',out/'checker.a','./bridge/tsgo/archive'])
entry=unit/'timer_suite.a'
run('native-build',[out/'adamic','build',entry,'-o',out/'timer','--tsgo',out/'checker.a'])
virtual=root/'cohere/adamic_wave07_timer_oracle.go'
oracle_overlay=overlay('oracle',{virtual:unit/'testdata/oracle_timer.go'})
run('oracle-build',['go','build','-overlay',oracle_overlay,'-o',out/'oracle',virtual],root/'cohere')
prelude=out/'ambient.d.ts'
prelude.write_text('export {}; declare global { namespace NodeJS { interface Timeout { unref():this; } } function setTimeout(cb:(...a:any[])=>void,ms?:number,...a:any[]):NodeJS.Timeout; namespace setTimeout {const __promisify__: unknown;} function clearTimeout(t:NodeJS.Timeout|undefined):void;}')
config=out/'tsconfig.json'
config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['es2022']},'files':['ambient.d.ts']}))
common='declare function work():Promise<string>;declare function use(x:unknown):void;declare const ms:number;\n'
controls=[
 'Promise.race([work(),new Promise((r,j)=>setTimeout(j,ms))]);',
 'Promise.race([work(),new Promise((r,j)=>{setTimeout(j,ms);})]);',
 'Promise.race([work(),new Promise((r,j)=>{const timer=setTimeout(j,ms);})]);',
 'const timeout=new Promise((r,j)=>{setTimeout(j,ms);});Promise.race([work(),timeout]);',
 'let timer;Promise.race([work(),new Promise((r,j)=>{timer=setTimeout(j,ms);})]);',
 'Promise.race([work(),new Promise((r,j)=>{globalThis.setTimeout(j,ms);})]);',
 'Promise.race([work(),new Promise((r,j)=>{void setTimeout(j,ms);})]);',
 'Promise.race([new Promise(r=>{setTimeout(r,ms);}),new Promise((r,j)=>{setTimeout(j,ms);})]);',
 'Promise.race([work(),new Promise((r,j)=>{setTimeout(j,ms).unref();})]);',
 'let timer:NodeJS.Timeout|undefined;Promise.race([work(),new Promise((r,j)=>{timer=setTimeout(j,ms);})]);clearTimeout(timer);',
 'Promise.race([work(),new Promise((r,j)=>{const timer=setTimeout(j,ms);use(timer);})]);',
 'Promise.race([work(),new Promise((r,j)=>{const timer=setTimeout(j,ms);use({timer});})]);',
 'Promise.race([work(),new Promise((r,j)=>{function nested(){setTimeout(j,ms);}})]);',
 'Promise.race([work(),new Promise((r,j)=>{return setTimeout(j,ms);})]);',
 'let timeout=new Promise((r,j)=>{setTimeout(j,ms);});Promise.race([work(),timeout]);',
 'Promise.race([work(),new Promise((r,j)=>{const setTimeout=(f:unknown,n:unknown)=>1;setTimeout(j,ms);})]);',
 'const Promise={race:(v:unknown)=>v};Promise.race([work()]);',
 '/*世界 🌍*/\r\nPromise.race([work(),new Promise((r,j)=>{const é=setTimeout(j,ms);})]);',
 'let a:NodeJS.Timeout,b:NodeJS.Timeout;Promise.race([work(),new Promise((r,j)=>{a=b=setTimeout(j,ms);})]);',
 'let timer:NodeJS.Timeout;Promise.race([work(),new Promise((r,j)=>{(timer)=setTimeout(j,ms);})]);',
]
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
if b'\tnexus/correctness-no-uncleared-race-timeout\t' not in truth: raise RuntimeError('no positive control')
# DOM-only declarations exercise window and library member merging independently.
(out/'dom-prelude.d.ts').write_text('export {};')
dom_config=out/'dom.json';dom_config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['es2022','dom']},'files':['dom-prelude.d.ts']}))
compare('controls-dom',out/'timer',dom_config,manifest)
environment=dict(os.environ,CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
run('archive-asan',['go','build','-overlay',registration,'-buildmode=c-archive','-o',out/'checker-asan.a','./bridge/tsgo/archive'],environment=environment)
run('native-asan-build',[out/'adamic','build',entry,'-o',out/'timer-asan','--tsgo',out/'checker-asan.a','--sanitize'])
compare('controls-asan',out/'timer-asan',config,manifest)
for name,corpus_root,corpus_config in [('repository',root,root/'tsconfig.json'),('compiler',args.compiler,args.compiler/'src/compiler/tsconfig.json')]:
    lines=(root/f'stage1/cohere/typeaware/validation-volume/{name}.manifest').read_text().splitlines()
    corpus_manifest=out/(name+'.manifest');corpus_manifest.write_text(''.join(str(corpus_root/line)+'\n' for line in lines))
    compare(name,out/'timer',corpus_config,corpus_manifest)
    compare(name+'-asan',out/'timer-asan',corpus_config,corpus_manifest)
mutant=out/'mutant.a';rule=(unit/'correctness_no_uncleared_race_timeout.a').read_text();token="'unclearedRaceTimeout'"
if rule.count(token)!=1: raise RuntimeError('nonunique mutant')
rule=rule.replace(token,"'unclearedRaceTimeoutMutant'").replace("'../","'"+str(unit.parent)+"/").replace("'../../../typescript/","'"+str(root/'stage1/typescript')+"/").replace("'./wave07_symbol_context.a'","'"+str(unit/'wave07_symbol_context.a')+"'")
mutant.write_text(rule)
mutant_entry=out/'mutant-suite.a';source=entry.read_text().replace("'../","'"+str(unit.parent)+"/").replace("'../../../typescript/","'"+str(root/'stage1/typescript')+"/").replace("'./correctness_no_uncleared_race_timeout.a'","'"+str(mutant)+"'")
mutant_entry.write_text(source)
run('mutant-build',[out/'adamic','build',mutant_entry,'-o',out/'timer-mutant','--tsgo',out/'checker.a'])
changed,errors=run('mutant',[out/'timer-mutant',config,manifest])
if errors or truth==changed: raise RuntimeError('rule mutant survived')
print('rule mutant: exit 0, empty stderr; caught only by independent Go bytes',flush=True)
print('PASS private-overlay timer gate; normal dispatcher remains unregistered',flush=True)

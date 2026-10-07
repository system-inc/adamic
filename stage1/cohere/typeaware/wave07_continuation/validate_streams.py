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
private.write_text(source.replace('switch mode {','switch mode {\n case "wave07-symbol-context": return p.wave07SymbolContext(c,node,question)\n case "wave07-control-flow": return p.wave07ControlFlow(node,question)\n case "wave07-program-modules": return p.wave07ProgramModules(node,question)'))
registration=overlay('registration',{facts:private})
run('question-test',['go','test','./bridge/tsgo/checker','-run','^TestWave07SymbolContext$','-count=1','-v'])
run('stage0',['go','build','-o',out/'adamic','./cmd/adamic'])
run('archive',['go','build','-overlay',registration,'-buildmode=c-archive','-o',out/'checker.a','./bridge/tsgo/archive'])
entry=unit/'streams_suite.a'
run('native-build',[out/'adamic','build',entry,'-o',out/'timer','--tsgo',out/'checker.a'])
virtual=root/'cohere/adamic_wave07_timer_oracle.go'
oracle_overlay=overlay('oracle',{virtual:unit/'testdata/oracle_streams.go'})
run('oracle-build',['go','build','-overlay',oracle_overlay,'-o',out/'oracle',virtual],root/'cohere')
prelude=out/'ambient.d.ts'
prelude.write_text('declare namespace NodeJS {interface Process {stdout:{write(text:string,callback?:()=>void):boolean};stderr:{write(text:string):boolean};exit(code?:number):never;exitCode:number;}} declare var process:NodeJS.Process; declare var console:{log(...args:unknown[]):void;error(...args:unknown[]):void;info(...args:unknown[]):void;};')
config=out/'tsconfig.json'
config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['es2022']},'files':['ambient.d.ts']}))
common='declare const flag:boolean;declare const n:number;declare function run():void;\n'
controls=["console.log('hi');process.exit(0);", "if(flag){console.log('hi');process.exit(0);}", "if(flag)process.exit(1);console.log('hi');", "console.log('hi');process.exit(0);process.exit(1);", "while(flag){process.exit(0);console.log('hi');}", "for(let i=0;i<3;i++){if(flag)process.exit(0);console.log('hi');}", "console.log('hi');try{run();}catch{process.exit(1);}", "try{run();console.log('hi');}catch{process.exit(1);}", "try{run();}catch{console.error('err');process.exit(1);}", "try{console.log('hi');}finally{process.exit(1);}", "console.log('hi');try{run();}catch{process.exit(1);}finally{process.exit(2);}", "function help(){console.log('hi');}help();process.exit(0);", "const help=()=>console.log('hi');help();process.exit(0);", "function help(){process.exit(1);console.log('hi');}help();process.exit(0);", "function* help(){console.log('hi');}help();process.exit(0);", "async function help(){console.log('hi');}help();process.exit(0);", "async function help(){console.log('hi');}async function main(){await help();process.exit(0);}", 'console.log(process.exit(0));process.exit(1);', "process.stdout.write('hi',()=>{process.exit(0);});", "process.stdout.write('hi');process.exit(0);", "process.stderr.write('hi');process.exit(0);", "function local(console:{log(x:string):void}){console.log('hi');process.exit(0);}", "function main(){return;console.log('hi');process.exit(0);}", "switch(n){case 1:console.log('hi');break;default:break;}process.exit(0);", "console.log('hi');try{throw new Error();}catch({code=process.exit(1)}){process.exit(0);}", "/*世界 🌍*/\r\nconsole.log('hi');process.exit(0);"]

# These are TypeScript oracle inputs, not Adamic implementation files. The
# production rule deliberately recognizes the legacy StandardStreams.ts path.
streams=out/'source/system/StandardStreams.ts'
streams.parent.mkdir(parents=True,exist_ok=True)
streams.write_text('export function blockStandardStreams():void {}')
load_block=out/'BlockOnLoad.ts';load_block.write_text("import {blockStandardStreams} from './source/system/StandardStreams.js';blockStandardStreams();")
controls.extend([
 "import {blockStandardStreams} from './source/system/StandardStreams.js';console.log('hi');process.exit(0);blockStandardStreams();",
 "import {blockStandardStreams} from './source/system/StandardStreams.js';blockStandardStreams();console.log('hi');process.exit(0);",
 "import {blockStandardStreams} from './source/system/StandardStreams.js';function main(){console.log('hi');process.exit(0);blockStandardStreams();}main();",
 "import {blockStandardStreams} from './source/system/StandardStreams.js';function main(){blockStandardStreams();console.log('hi');process.exit(0);}main();",
 "import './BlockOnLoad.js';console.log('hi');process.exit(0);",
 "import {blockStandardStreams} from './source/system/StandardStreams.js';async function main(){await run();console.log('hi');process.exit(0);blockStandardStreams();}main();",
 "import {blockStandardStreams} from './source/system/StandardStreams.js';process.stdout.write('hi',()=>{console.log('hi');process.exit(0);});blockStandardStreams();",
 "#!/usr/bin/env node\nconsole.log('hi');process.exit(0);",
 "export function library(){console.log('hi');process.exit(0);}",
 "import {blockStandardStreams} from './source/system/StandardStreams.js';function recurse(){if(flag){console.log('hi');process.exit(0);}else{recurse();blockStandardStreams();}}recurse();",
])

paths=[]
for i,source in enumerate(controls):
    path=out/f'control-{i:03d}.a';path.write_text((source.split('\n',1)[0]+'\n'+common+source.split('\n',1)[1] if source.startswith('#!') else common+source)+'\nexport {};\n');paths.append(path)
importer=out/'Importer.ts';importer.write_text("import {library} from './control-034.a';void library;");paths.append(importer)
manifest=out/'controls.manifest';manifest.write_text(''.join(str(p)+'\n' for p in paths))

def compare(name, native, config, manifest):
    truth,go_error=run(name+'-go',[out/'oracle',config,manifest])
    actual,native_error=run(name+'-native',[native,config,manifest])
    if truth!=actual or native_error: raise RuntimeError(f'{name}: diagnostic disagreement or native stderr')
    print(f'{name}: {len(actual)} identical bytes, {actual.splitlines()[-1].decode()}',flush=True)
    return truth

truth=compare('controls',out/'timer',config,manifest)
if b'\tnexus/correctness-require-blocking-standard-streams\t' not in truth: raise RuntimeError('no positive control')
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
mutant=out/'mutant.a';rule=(unit/'correctness_require_blocking_standard_streams.a').read_text();token="'exitBeforeBlockingStandardStreams'"
if rule.count(token)!=1: raise RuntimeError('nonunique mutant')
rule=rule.replace(token,"'exitBeforeBlockingStandardStreamsMutant'").replace("'../","'"+str(unit.parent)+"/").replace("'../../../typescript/","'"+str(root/'stage1/typescript')+"/").replace("'./","'"+str(unit)+"/")
mutant.write_text(rule)
mutant_entry=out/'mutant-suite.a';source=entry.read_text().replace("'../","'"+str(unit.parent)+"/").replace("'../../../typescript/","'"+str(root/'stage1/typescript')+"/").replace("'./correctness_require_blocking_standard_streams.a'","'"+str(mutant)+"'").replace("'./","'"+str(unit)+"/")
mutant_entry.write_text(source)
run('mutant-build',[out/'adamic','build',mutant_entry,'-o',out/'timer-mutant','--tsgo',out/'checker.a'])
changed,errors=run('mutant',[out/'timer-mutant',config,manifest])
if errors or truth==changed: raise RuntimeError('rule mutant survived')
print('rule mutant: exit 0, empty stderr; caught only by independent Go bytes',flush=True)
print('PASS private-overlay timer gate; normal dispatcher remains unregistered',flush=True)

#!/usr/bin/env python3
"""Wave-owned output-rule parity runner; subprocess output always goes to files."""
import json, os, subprocess, time, gzip, shutil, statistics
from pathlib import Path
ROOT=Path(__file__).resolve().parents[4]
HERE=Path(__file__).resolve().parent
OUT=Path(os.environ.get('WAVE05_OUTPUT_ARTIFACTS','/workspace/wave-05-output-validation'))
OUT.mkdir(exist_ok=True)
def run(name,args,cwd=ROOT,env=None,expected=0):
    begin=time.monotonic()
    with (OUT/(name+'.stdout')).open('wb') as stdout,(OUT/(name+'.stderr')).open('wb') as stderr:
        done=subprocess.run(list(map(str,args)),cwd=cwd,stdout=stdout,stderr=stderr,env=env)
    assert done.returncode==expected,(name,done.returncode,(OUT/(name+'.stderr')).read_text())
    return (OUT/(name+'.stdout')).read_bytes(),(OUT/(name+'.stderr')).read_bytes(),time.monotonic()-begin
stage0=OUT/'adamic';archive=OUT/'checker.a';native=OUT/'native';oracle=OUT/'oracle'
run('stage0',['go','build','-o',stage0,'./cmd/adamic'])
run('archive',['go','build','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'])
run('native-build',[stage0,'build',HERE/'main.a','-o',native,'--tsgo',archive])
virtual=ROOT/'cohere/adamic_wave05_next_oracle.go';overlay=OUT/'oracle-overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'testdata/oracle.go')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],ROOT/'cohere')
platform=OUT/'node.d.ts'
platform.write_text('declare var process:NodeJS.Process; declare namespace NodeJS {interface WriteStream {write(s:unknown,callback?:()=>void):boolean;} interface Process {stdout:WriteStream;stderr:WriteStream;exit(n?:number):never;exitCode:number;on(e:string,f:()=>void):this;}} declare module "node:process" {export=process;}\n')
controls=[
"console.log('x');process.exit(0);",
"if(flag){console.log('x');process.exit(1);}",
"if(!flag)process.exit(1);console.log('x');",
"console.log('x');process.exitCode=0;",
"process.stdout.write('x');process.exit(0);",
"import {stderr,exit} from 'node:process';stderr.write('x');exit(0);",
"import * as p from 'node:process';p.stderr.write('x');p.exit(0);",
"const console={log(s:string){}};console.log('x');process.exit(0);",
"const p={stdout:{write(s:string){}},exit(n:number){}};p.stdout.write('x');p.exit(0);",
"const s={write(s:string){}};s.write('x');process.exit(0);",
"process.stdout.write('x',()=>process.exit(0));",
"function help(){console.log('x');}help();process.exit(0);",
"async function help(){console.log('x');}help();process.exit(0);",
"async function main(){async function help(){console.log('x');}await help();process.exit(0);}main();",
"function * help(){console.log('x');}help();process.exit(0);",
"function help():never{console.log('x');throw Error();}help();process.exit(0);",
"function help(){console.log('x');process.exit(1);}help();process.exit(0);",
"try{console.log('x');}catch{process.exit(1);}",
"console.log('x');try{work();}catch{process.exit(1);}",
"try{work();console.log('x');}catch{process.exit(1);}",
"try{work();}catch{console.error('x');process.exit(1);}",
"while(flag){process.exit(1);console.log('x');}",
"while(flag){if(flag)process.exit(1);console.log('x');}",
"for(const x of rows){console.log(x);}process.exit(0);",
"console.log(process.exit(1));",
"console['log']('x');process.exit(0);",
"console.log('x');process['exit'](0);",
"function help(){()=>console.log('x');}help();process.exit(0);",
"/* 世界 🌍 */\r\nconsole.log('x');process.exit(0);",
"console.log('x');try{return;}finally{process.exit(0);}",
]
paths=[]
for at,text in enumerate(controls):
    p=OUT/f'output-{at:03d}.a'
    # return is wrapped, all other cases retain their own scope and imports.
    if 'try{return;}' in text:text='function main(){'+text+'}main();'
    p.write_text('declare const flag:boolean;declare const rows:string[];declare function work():void;\n'+text+'\nexport {};\n');paths.append(p)
manifest=OUT/'controls.manifest';manifest.write_text(''.join(str(p)+'\n' for p in paths))
config=OUT/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022','DOM']},'files':[str(platform)]}))
def compare(name,conf,files,exe):
    truth,_,_=run(name+'-go',[oracle,conf,files])
    actual,stderr,_=run(name+'-native',[exe,conf,files]);assert not stderr,(name,stderr)
    (OUT/(name+'.oracle.gz')).write_bytes(gzip.compress(truth,mtime=0))
    if actual!=truth:
        (OUT/(name+'.actual')).write_bytes(actual);(OUT/(name+'.expected')).write_bytes(truth)
    assert actual==truth,(name,'byte mismatch')
    print(name,len(truth),truth.splitlines()[-1].decode(),flush=True)
    return truth
truth=compare('controls',config,manifest,native)
assert b'exitAfterOutput' in truth
# Module fixtures are TypeScript oracle inputs, not Adamic implementation files.
module_controls=[
("none", "console.log('x');process.exit(0);", {}),
("first", "import {blockStandardStreams as block} from './source/system/StandardStreams';block();console.log('x');process.exit(0);", {}),
("late", "import {blockStandardStreams as block} from './source/system/StandardStreams';console.log('x');process.exit(0);block();", {}),
("local-main", "import {blockStandardStreams as block} from './source/system/StandardStreams';function main(){block();console.log('x');process.exit(0);}main();", {}),
("guard-main", "import {blockStandardStreams as block} from './source/system/StandardStreams';function main(){if(flag){console.log('x');process.exit(0);}block();}main();", {}),
("unstarted", "export function main(){console.log('x');process.exit(0);}", {}),
("callback", "process.on('exit',()=>{console.log('x');process.exit(0);});", {}),
("ordered-callback", "import {blockStandardStreams as block} from './source/system/StandardStreams';process.on('exit',()=>{console.log('x');process.exit(0);});block();", {}),
("await", "import {blockStandardStreams as block} from './source/system/StandardStreams';async function main(){await Promise.resolve();console.log('x');process.exit(0);}main();", {}),
("unawaited", "async function main(){await Promise.resolve();console.log('x');process.exit(0);}main();", {}),
("unknown", "import {blockStandardStreams as block} from './source/system/StandardStreams';unknownCall();console.log('x');process.exit(0);", {}),
("computed-import", "declare const moduleName:string;import(moduleName);console.log('x');process.exit(0);", {}),
("two-exits", "if(flag){console.log('x');process.exit(0);}else{console.error('x');process.exit(1);}", {}),
("const-arrow", "const main=()=>{console.log('x');process.exit(0);};main();", {}),
("let-arrow", "let main=()=>{console.log('x');process.exit(0);};main();", {}),
("generator", "function * main(){console.log('x');process.exit(0);}main();", {}),
("ordered-generator", "import {blockStandardStreams as block} from './source/system/StandardStreams';function * main(){console.log('x');process.exit(0);}main();block();", {}),
("imported", "console.log('x');process.exit(0);", {'importer.ts':"import './entry';"}),
("shebang-imported", "#!/usr/bin/env tsx\nconsole.log('x');process.exit(0);", {'importer.ts':"import './entry';"}),
("type-only", "console.log('x');process.exit(0);", {'importer.ts':"import type {} from './entry';"}),
("load-block", "import './loader';console.log('x');process.exit(0);", {'loader.ts':"import {blockStandardStreams} from './source/system/StandardStreams';blockStandardStreams();"}),
("export-helper", "import {help} from './help';help();process.exit(0);", {'help.ts':"export function help(){console.log('x');}"}),
("foreign-global", "help();process.exit(0);", {'global.ts':"function help(){console.log('x');}"}),
("foreign-overload", "import {help} from './help';help();process.exit(0);", {'help.ts':"export function help():void;export function help(){console.log('x');}"}),
("ordered-argument", "import {blockStandardStreams as block} from './source/system/StandardStreams';function main(x:unknown){console.log('x');process.exit(0);}main(block());", {}),
("for-await", "import {blockStandardStreams as block} from './source/system/StandardStreams';async function main(){for await(const x of rows){}console.log('x');process.exit(0);}main();", {}),
("recursive", "import {blockStandardStreams as block} from './source/system/StandardStreams';function recur(){if(flag)recur();block();}recur();console.log('x');process.exit(0);", {}),
("ambient", "import {blockStandardStreams as block} from './source/system/StandardStreams';declare function f():void;f();console.log('x');process.exit(0);", {}),
("catch-binding", "console.log('x');try{work();}catch({x=process.exit(1)}){process.exit(2);}", {}),
("class-initializer", "import {blockStandardStreams as block} from './source/system/StandardStreams';class C{x=block();}console.log('x');process.exit(0);", {}),
]
for at,(name,text,others) in enumerate(module_controls):
    folder=OUT/f'module-{at:03d}-{name}';folder.mkdir(exist_ok=True)
    streams=folder/'source/system/StandardStreams.ts';streams.parent.mkdir(parents=True,exist_ok=True)
    streams.write_text("export function blockStandardStreams():void{Reflect.get(process.stdout,'_handle');}\n")
    if text.startswith('#!'):
        line,text=text.split('\n',1);text=line+'\n'+"declare const flag:boolean;declare const rows:string[];declare function work():void;\n"+text
    else:text="declare const flag:boolean;declare const rows:string[];declare function work():void;\n"+text
    entry=folder/'entry.ts';entry.write_text(text+'\nexport {};\n')
    roots=[entry]
    for filename,body in others.items():
        file=folder/filename;file.parent.mkdir(parents=True,exist_ok=True);file.write_text(body+'\n');roots.append(file)
    module_manifest=folder/'manifest';module_manifest.write_text(''.join(str(file)+'\n' for file in roots))
    module_config=folder/'tsconfig.json';module_config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022','DOM']},'files':[str(platform)]}))
    compare('module-'+name,module_config,module_manifest,native)
print('PASS output controls and module cases',flush=True)

# Existing timer fixtures must retain their original findings with the new rules loaded.
base=Path('/workspace/wave-05-next-validation')
for mode in ('dom','node'):
    compare('timer-'+mode,base/(mode+'.json'),base/'controls.manifest',native)
asan_archive=OUT/'sanitized.a'
environment=dict(os.environ,CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
run('asan-archive',['go','build','-buildmode=c-archive','-o',asan_archive,'./bridge/tsgo/archive'],env=environment)
san=OUT/'sanitized'
run('asan-build',[stage0,'build',HERE/'main.a','-o',san,'--tsgo',asan_archive,'--sanitize'])
compare('controls-asan',config,manifest,san)
for at,(name,_,_) in enumerate(module_controls):
    folder=OUT/f'module-{at:03d}-{name}'
    compare('module-'+name+'-asan',folder/'tsconfig.json',folder/'manifest',san)
for mode in ('dom','node'):
    compare('timer-'+mode+'-asan',base/(mode+'.json'),base/'controls.manifest',san)
for name,file,pattern,replacement in [
    ('process','no_process_exit_after_output.a','state.length > 0 && !reported.includes(block.index)','state.length === 0 && !reported.includes(block.index)'),
    ('blocking','require_blocking_standard_streams.a','state.length > 0 && !result.exits.includes(block.index)','state.length === 0 && !result.exits.includes(block.index)'),
    ('cfg','output_cfg.a','this.link(this.current, node.kind === \'DoStatement\' ? bodyEntry : head);','this.link(this.current, after);'),
]:
    import re
    folder=OUT/('mutant-'+name);folder.mkdir(exist_ok=True)
    for source in HERE.glob('*.a'):
        text=source.read_text()
        if source.name==file:
            assert text.count(pattern)==1,(name,'mutation drift');text=text.replace(pattern,replacement)
        text=re.sub(r"from '([^']+)'",lambda m: "from '"+str((HERE/m[1]).resolve())+"'" if m[1].startswith('../') else m[0],text)
        (folder/source.name).write_text(text)
    exe=OUT/('mutant-'+name+'-bin')
    run(name+'-mutant-build',[stage0,'build',folder/'main.a','-o',exe,'--tsgo',archive])
    wrong,stderr,_=run(name+'-mutant-run',[exe,config,manifest])
    assert not stderr and wrong!=truth,(name,'comparison did not catch mutant')
    print(name,'mutant exits 0, empty stderr, caught only by Go byte comparison',flush=True)
release=OUT/'released'
run('released-build',[stage0,'build',HERE/'testdata/released_output.a','-o',release,'--tsgo',archive])
probe=OUT/'released.a';probe.write_text('clock;\n')
for question in ['output-symbol','output-callee','runtime-modules']:
    _,stderr,_=run('released-'+question,[release,config,probe,question],expected=70)
    assert stderr==b'adamic: panic: invalid or released checker handle\n'
    print(question,'released handle exact refusal 70',flush=True)
corpora=[('repository',ROOT/'tsconfig.json',Path('/workspace/wave-05-repository.manifest')),('compiler',Path('/workspace/wave-05-typescript/src/compiler/tsconfig.json'),Path('/workspace/wave-05-compiler.manifest'))]
for name,conf,files in corpora:
    compare(name,conf,files,native);compare(name+'-asan',conf,files,san)
if not os.environ.get('WAVE05_OUTPUT_SKIP_BENCH'):
    timings={}
    for name,conf,files in corpora:
        rounds={'go':[],'native':[]}
        for at in range(3):
            for label,exe in ([('go',oracle),('native',native)] if at%2==0 else [('native',native),('go',oracle)]):
                output,stderr,elapsed=run(f'{name}-{label}-round-{at}',[exe,conf,files,'--count'],env=dict(os.environ,ADAMIC_TSGO_TIMING='1'))
                rounds[label].append(elapsed);print(name,label,at,elapsed,output.strip().decode(),stderr.strip().decode(),flush=True)
        timings[name]=rounds
        print(name,'median',json.dumps({label:statistics.median(values) for label,values in rounds.items()}),flush=True)
    (OUT/'timings.json').write_text(json.dumps(timings,indent=2)+'\n')
print('PASS both output rules, timer regression, corpora, comparison mutants, release and sanitizers',flush=True)

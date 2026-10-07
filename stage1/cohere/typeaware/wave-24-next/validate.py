#!/usr/bin/env python3
"""Wave-24-owned integration runner. Never edits the shared lint harness."""
import argparse
import gzip
import json
import os
from pathlib import Path
import subprocess
import time

ROOT = Path(__file__).resolve().parents[4]
OWN = Path(__file__).resolve().parent
CONTROLS = [
    "while(true){try{break;}finally{console.log('x');}}process.exit(1);",
    "try{while(true){break;}process.exit(1);}finally{console.log('x');}",
    "console.log('x');try{try{throw 1;}finally{process.exit(1);}}catch{}",
    "console.log('x');process.\\u0065xit(1);",

    "for(/*;*/let x=0;/*;*/x<2;/*;*/x++){console.log('x');}process.exit(0);",
    "declare const target:string;function loader(){import(target);}declare function block():void;function main(){console.log('x');process.exit(1);block();}main();",
    "declare const target:string;function loader(){import(target);}declare function block():void;function main(){block();console.log('x');process.exit(1);}main();",
    "import {blockStandardStreams} from './source/system/StandardStreams.js';console.log('x');process.exit(0);",

    "console.log('start');try{declareRun();process.exit(1);}catch{process.exit(2);}declare function declareRun():void;",
    "function f(){console.log('x');process.exit(1);return;process.exit(2);}f();",
    "declare const flag:boolean;if(!(flag && console.log('x')))process.exit(1);",

    "console.log('ready');process.exit(0);",
    "process.exit(0);console.log('late');process.exit(1);",
    "declare const flag:boolean;if(flag){console.warn('x');process.exit(1);}",
    "declare const flag:boolean;if(flag)console.log('x');else process.exit(1);process.exit(2);",
    "declare const flag:boolean;while(flag){process.exit(1);console.log('again');}",
    "declare const flag:boolean;while(flag){console.log('x');break;}process.exit(1);",
    "declare const flag:boolean;do{console.log('x');}while(flag);process.exit(0);",
    "declare function run():void;console.log('start');try{run();}catch{process.exit(1);}",
    "declare function run():void;try{run();console.log('end');}catch{process.exit(1);}",
    "try{console.log('x');}finally{process.exit(1);}",
    "function f(){try{console.log('x');return;}finally{process.exit(1);}}f();",
    "function help(){console.log('help');}help();process.exit(0);",
    "function deep(){console.log('help');}function help(){deep();}help();process.exit(0);",
    "async function help(){console.log('help');}async function f(){await help();process.exit(0);}f();",
    "async function help(){console.log('help');}help();process.exit(0);",
    "function* help(){console.log('help');}help();process.exit(0);",
    "function help():never{console.log('help');throw 1;}help();process.exit(0);",
    "process.stdout.write('x');process.exit(0);",
    "process.stderr.write('x',()=>process.exit(0));",
    "const console={log(_x:string){}};console.log('x');process.exit(0);",
    "const fake={exit(_code:number){}};console.log('x');fake.exit(1);",
    "console.log(process.exit(1));process.exit(2);",
    "declare const flag:boolean;switch(flag){case true:console.log('x');break;default:break;}process.exit(0);",
    "declare const flag:boolean;outer:while(flag){console.log('x');break outer;}process.exit(0);",
    "for(let x=0;x<2;x++){console.log('x');}process.exit(0);",
    "for(const x of [1,2]){console.log(x);}process.exit(0);",
    "const f=()=>{console.log('x');process.exit(1);};f();",
    "declare function register(callback:()=>void):void;register(()=>{console.log('x');process.exit(1);});",
    "console.log('x');if(Math.random()){process.exit(1);}process.exit(2);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{setTimeout(reject,100);})]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>setTimeout(reject,100))]);",
    "declare const work:Promise<number>;const timeout=new Promise((_r,reject)=>{const timer=setTimeout(reject,100);});Promise.race([work,timeout]);",
    "declare const work:Promise<number>;let timer:number;Promise.race([work,new Promise((_r,reject)=>{timer=setTimeout(reject,100);})]);clearTimeout(timer!);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{const timer=setTimeout(reject,100);console.log(timer);})]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{globalThis.setTimeout(reject,100);})]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{window.setTimeout(reject,100);})]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{void setTimeout(reject,100);})]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{(()=>setTimeout(reject,100))();})]);",
    "declare const work:Promise<number>;let timeout=new Promise((_r,reject)=>{setTimeout(reject,100);});Promise.race([work,timeout]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{const setTimeout=(_f:unknown,_ms:number)=>0;setTimeout(reject,100);})]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{const timer=setTimeout(reject,100);const out={timer};})]);",
    "declare const work:Promise<number>;Promise.race([work,new Promise((_r,reject)=>{let timer;timer=setTimeout(reject,100);})]);",
    "/* 世界 🌍 */\r\nconsole.log('漢');process.exit(1);\r\n",
    "#!/usr/bin/env tsx\nconsole.log('x');process.exit(0);",
    "declare const moduleName:string;console.log('x');process.exit(0);import(moduleName);",
    "declare const moduleName:string;import(moduleName);console.log('x');process.exit(0);",
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--compiler', type=Path)
    parser.add_argument('--stage0', type=Path, required=True)
    parser.add_argument('--controls-only', action='store_true')
    args = parser.parse_args()
    dest = args.directory.resolve(); dest.mkdir(parents=True, exist_ok=True)
    serial = 0
    def run(label, command, cwd=ROOT, environment=None, expected=0):
        nonlocal serial
        serial += 1; stem = dest / f'{serial:03}-{label}'
        started = time.perf_counter_ns()
        with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
            result = subprocess.run(list(map(str, command)), cwd=cwd, env=environment, stdout=stdout, stderr=stderr)
        elapsed = time.perf_counter_ns() - started
        assert result.returncode == expected, (label, result.returncode, stem.with_suffix('.stderr').read_text())
        return stem.with_suffix('.stdout').read_bytes(), stem.with_suffix('.stderr').read_bytes(), elapsed
    def archive(label, sanitizer=False, overlay=None):
        output = dest / (label + '.a'); command = ['go','build','-buildmode=c-archive','-o',output]
        if overlay: command += ['-overlay', overlay]
        command += ['./bridge/tsgo/archive']; environment = dict(os.environ)
        if sanitizer: environment.update(CC='clang', CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
        run(label,command,environment=environment); return output
    def build(label, entry, linked, sanitizer=False):
        output = dest/label; command = [args.stage0,'build',entry,'-o',output,'--tsgo',linked]
        if sanitizer: command += ['--sanitize']
        run(label,command); return output
    linked=archive('checker'); native=build('native',OWN/'suite.a',linked)
    virtual=ROOT/'cohere/adamic_wave24_next_oracle.go'; overlay=dest/'oracle-overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/oracle.go')}}))
    oracle=dest/'oracle';run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],cwd=ROOT/'cohere')
    streams = dest/'source/system/StandardStreams.d.ts';streams.parent.mkdir(parents=True,exist_ok=True);streams.write_text('export declare function blockStandardStreams():void;\n')
    declarations = dest/'globals.d.ts'
    declarations.write_text("declare namespace NodeJS { interface Process { exit(code?:number):never; stdout:{write(text:string,callback?:()=>void):void}; stderr:{write(text:string,callback?:()=>void):void}; } }\ndeclare var process:NodeJS.Process;\n")
    config=dest/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'NodeNext','lib':['ES2022','DOM']},'files':[str(declarations)]}))
    paths=[]
    for index,source in enumerate(CONTROLS):
        path=dest/f'control-{index:03}.a';path.write_text(source+'\nexport {};\n');paths.append(path)
    manifest=dest/'controls.manifest';manifest.write_text(''.join(str(path)+'\n' for path in paths))
    def compare(label, config, manifest, binary):
        truth,_,_=run(label+'-go',[oracle,config,manifest]); got,stderr,_=run(label+'-native',[binary,config,manifest])
        assert not stderr,(label,stderr)
        if got != truth:
            offset=next((i for i,(a,b) in enumerate(zip(got,truth)) if a!=b),min(len(got),len(truth)))
            print(label,'MISMATCH',offset,repr(truth[max(0,offset-80):offset+500]),repr(got[max(0,offset-80):offset+500]),flush=True)
            raise AssertionError((label,'diagnostic byte mismatch'))
        for mode,data in [('go',truth),('native',got)]:
            with gzip.GzipFile(filename=str(dest/f'{label}-{mode}.findings.gz'),mode='wb',mtime=0) as output: output.write(data)
        print(label,len(got),'identical bytes',got.splitlines()[-1].decode(),flush=True);return truth
    truth=compare('controls',config,manifest,native)
    for name in ['no-process-exit-after-output','no-uncleared-race-timeout','require-blocking-standard-streams']:
        assert ('\tnexus/correctness-'+name+'\t').encode() in truth,('missing positive control',name)
    if args.controls_only: return
    sanitized=archive('checker-asan',True);asan=build('native-asan',OWN/'suite.a',sanitized,True)
    compare('controls-asan',config,manifest,asan)
    changes=[('exit','no_process_exit_after_output.a','this.walk(root))','this.walk(root, true))'),
             ('timer','no_uncleared_race_timeout.a','&& this.lost(executor, index)','&& !this.lost(executor, index)'),
             ('blocking','require_blocking_standard_streams.a','if(!shebang && graph.imported.has(this.rules.path))','if(!shebang && !graph.imported.has(this.rules.path))')]
    # Blocking's native entry judgment is separately mutated; the underlying
    # per-exit rule stays unchanged in this mutant.
    for label,file,original,replacement in changes:
        directory=dest/(label+'-source');directory.mkdir(exist_ok=True)
        for path in OWN.glob('*.a'):
            source=path.read_text()
            if path.name==file:
                assert source.count(original)==1,(label,'nonunique mutation');source=source.replace(original,replacement)
            source=source.replace("'../", "'"+str(OWN.parent)+'/')
            (directory/path.name).write_text(source)
        mutant=build(label+'-mutant',directory/'suite.a',linked)
        got,stderr,_=run(label+'-run',[mutant,config,manifest]);assert not stderr and got!=truth,(label,'mutant survived')
        offset=next((i for i,(a,b) in enumerate(zip(got,truth)) if a!=b),min(len(got),len(truth)))
        print(label,'mutant exits 0, empty stderr; byte oracle catches',offset,flush=True)
    for label,file,original,replacement in [
        ('provenance-question','symbol_provenance.go','out.yes(source.IsDeclarationFile)','out.yes(source.IsDeclarationFile && false)'),
        ('callee-question','resolved_call_origin.go','out.number(flags)','out.number(flags | uint64(checker.TypeFlagsNever))'),
        ('modules-question','program_module_edges.go','out.yes(computed)','out.yes(computed && false)'),
    ]:
        originalPath=ROOT/'bridge/tsgo/checker'/file
        source=originalPath.read_text();assert source.count(original)==1,(label,'nonunique mutation')
        side=dest/(label+'.go');side.write_text(source.replace(original,replacement))
        overlay=dest/(label+'.json');overlay.write_text(json.dumps({'Replace':{str(originalPath):str(side)}}))
        mutantArchive=archive(label,overlay=overlay);mutant=build(label+'-native',OWN/'suite.a',mutantArchive)
        got,stderr,_=run(label+'-run',[mutant,config,manifest]);assert not stderr and got!=truth,(label,'mutant survived')
        offset=next((i for i,(a,b) in enumerate(zip(got,truth)) if a!=b),min(len(got),len(truth)))
        print(label,'mutant exits 0, empty stderr; byte oracle catches',offset,flush=True)
    for corpus,sourceRoot,sourceConfig in [('repository',ROOT,ROOT/'tsconfig.json'),('compiler',args.compiler,args.compiler/'src/compiler/tsconfig.json' if args.compiler else None)]:
        if sourceRoot is None: continue
        portable=ROOT/'stage1/cohere/typeaware/validation-coverage'/f'{corpus}.manifest'
        resolved=dest/f'{corpus}.manifest';resolved.write_text(''.join(str(sourceRoot/path)+'\n' for path in portable.read_text().splitlines()))
        compare(corpus,sourceConfig,resolved,native);compare(corpus+'-asan',sourceConfig,resolved,asan)
    released=dest/'released.a';released.write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,4,'ArrowFunction','symbol-provenance'));\n")
    probe=dest/'probe.a';probe.write_text('x=>x;\n');stale=build('released',released,linked)
    _,stderr,_=run('released-run',[stale,config,probe],expected=70);assert stderr==b'adamic: panic: invalid or released checker handle\n'
    print('released provenance query: exact panic 70',flush=True)
    print('PASS',flush=True)


if __name__ == '__main__': main()

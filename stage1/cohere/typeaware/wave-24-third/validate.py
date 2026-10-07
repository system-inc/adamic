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
CONTROLS = ['Promise.reject();',
 'Promise.reject(undefined);',
 "Promise.reject('x');",
 'Promise.reject(1);',
 'Promise.reject(null);',
 'Promise.reject(false);',
 'Promise.reject({code:1});',
 'Promise.reject([]);',
 "Promise.reject(new Error('x'));",
 'declare const error:Error;Promise.reject(error);',
 'declare const error:Error;Promise.reject(error as Error);',
 'declare const error:Error;Promise.reject((error));',
 'Promise.reject(void 1);',
 'declare let error:Error;Promise.reject(error && 5);',
 'declare let error:Error;Promise.reject(5 && error);',
 'declare let error:Error;Promise.reject(error || 5);',
 'declare let error:Error;Promise.reject(5 ?? error);',
 'declare let error:Error;Promise.reject((error,5));',
 'declare let error:Error;Promise.reject((5,error));',
 'declare let error:Error;Promise.reject(true?error:5);',
 'declare let error:any;Promise.reject(error+=5);',
 "Promise['reject'](5);",
 '(Promise?.reject)(5);',
 'Promise?.reject?.(5);',
 'function f(Promise:any){Promise.reject(5);}',
 'function f(undefined:Error){Promise.reject(undefined);}',
 'const Promise={reject(_value:unknown){}};Promise.reject(5);',
 'new Promise((resolve,reject)=>reject(5));',
 'new Promise((resolve,reject)=>reject(new Error()));',
 'new Promise((resolve,reject)=>{function inner(){reject(5);}inner();});',
 'new Promise((resolve,reject)=>{function inner(reject:(x:number)=>void){reject(5);}});',
 'new Promise((resolve,reject)=>{const reject2=reject;reject2(5);});',
 'new Promise((resolve,reject)=>{resolve(5,reject);});',
 'new Promise((resolve,{apply})=>apply(5));',
 'new Promise(({foo},reject)=>reject(5));',
 'new Promise(function(reject,reject){reject(5);});',
 'new Promise((resolve,reject,third=reject(5))=>{});',
 "/* 世界 🌍 */\r\nPromise.reject('漢');\r\n",
 'arguments;',
 'const f=()=>arguments;',
 'function f(){arguments;}',
 'function f(){arguments[0];}',
 'function f(){arguments.length;arguments.callee;}',
 'function f(arguments:unknown){arguments;}',
 'function f(){var arguments;arguments;}',
 'function f(){function g(){arguments;}}',
 'function f(arguments:unknown){function g(){arguments;}}',
 'function f(){const g=()=>arguments;}',
 'function f(){try{}catch(arguments){arguments;}}',
 'function f(){const obj={arguments:1};obj.arguments;}',
 "function f(){arguments['length'];}",
 'function f(){let arguments:unknown;arguments;}',
 'function f(){if(true){let arguments;arguments;}arguments;}',
 'function f(){const obj={arguments};}']
CONTROLS += [
    "new RegExp('abc');", "RegExp('abc','g');", "new RegExp('');", "new RegExp('a/b');",
    "RegExp('a\\n');", "RegExp('漢');", "new RegExp(/* note */'abc');",
    "new RegExp('+');", "new RegExp('[');", "new RegExp('a','gg');", "new RegExp('a','uv');",
    "new RegExp('a'+'b');", "new RegExp();", "new RegExp(/a/);", "function f(RegExp:any){new RegExp('a');}",
    "RegExp(String.raw`\\d`,'g');", "RegExp(String['raw']`\\w{1, 2`,'u');", "new RegExp('a*?');",
    "new RegExp('[a-z/]');", "const R=RegExp;new R('a');", "let R=RegExp;R('a');R=1 as any;R('b');",
    "RegExp=1 as any;new RegExp('a');", "globalThis.RegExp('a');", "const {RegExp:R}=globalThis;new R('a');",
    "const R=(Math.random()?RegExp:RegExp);R('a');", "const R=(RegExp as typeof RegExp);R('a');",
    "RegExp(String.raw`[/\\d]`);", "a/RegExp('abc');declare const a:number;", "RegExp('a')in {};",
    "/* 世界 🌍 */\r\nRegExp('a','g');\r\n", "new RegExp(String.raw`\\p{Script=Greek}`,'u');",
    "RegExp(String.raw`[\\B]`,'u');", "RegExp(String.raw`[a-\\d]`);", "RegExp('x\\t');",
    "RegExp('a','v');", "RegExp(String.raw`[[a]--[/]]`,'v');"
]

CONTROLS += [
    "declare const error:Error;Promise?.reject?.(error);", "Promise.reject<Error>(5 as any);", "new Promise<void>((resolve,reject)=>reject(5));",
    "RegExp<never>('a');", "new RegExp<never>('a');", "globalThis[`Reg${'Exp'}`]('a');",
    "RegExp++;RegExp('a');", "[...RegExp]=[] as any;RegExp('a');", "let R;({RegExp:R}=globalThis);R('a');",
    "const key='Exp';globalThis[`Reg${key}`]('a');", "RegExp(String.raw`[a-/]`);", "RegExp(String.raw`[a-\\n]`);"
]
CONTROLS += json.loads((OWN/'testdata/imported.json').read_text())


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
    virtual=ROOT/'cohere/adamic_wave24_third_oracle.go'; overlay=dest/'oracle-overlay.json'
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
    for name in ['prefer-promise-reject-errors','prefer-rest-params','prefer-regex-literals']:
        assert ('\t'+name+'\t').encode() in truth,('missing positive control',name)
    if args.controls_only: return
    sanitized=archive('checker-asan',True);asan=build('native-asan',OWN/'suite.a',sanitized,True)
    compare('controls-asan',config,manifest,asan)
    changes=[('promise','prefer_promise_reject_errors.a','if(reason >= 0 && this.error(reason) && !undefinedGlobal)', 'if(reason >= 0 && !this.error(reason) && !undefinedGlobal)'),
             ('rest','prefer_rest_params.a','symbol.count !== 0', 'symbol.count === 0'),
             ('regex','prefer_regex_literals.a', 'escaped === \'\' ? \'(?:)\' : escaped', 'escaped === \'\' ? \'(?:)\' : escaped + \'x\'')]
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
        ('direct-symbol-question','raw_symbol_declaration_count.go','out.number(uint64(count))','out.number(uint64(count + 1))'),
        ('provenance-question','symbol_provenance.go','out.yes(source.IsDeclarationFile)','out.yes(source.IsDeclarationFile && false)'),
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
    released=dest/'released.a';released.write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,9,'Identifier','raw-symbol-declaration-count'));\n")
    probe=dest/'probe.a';probe.write_text('arguments;\n');stale=build('released',released,linked)
    _,stderr,_=run('released-run',[stale,config,probe],expected=70);assert stderr==b'adamic: panic: invalid or released checker handle\n'
    print('released direct symbol query: exact panic 70',flush=True)
    print('PASS',flush=True)


if __name__ == '__main__': main()

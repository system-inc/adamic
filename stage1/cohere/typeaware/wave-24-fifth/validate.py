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
CONTROLS = ['Symbol();', 'Symbol(undefined);', "Symbol('x');", 'new Symbol();', 'globalThis.Symbol();', '(Symbol)();', '((Symbol))();', 'Symbol?.();', 'function f(Symbol:()=>void){Symbol();}', 'const Symbol=()=>{};Symbol();', 'declare const Symbol:()=>symbol;Symbol();', 'Symbol(...[]);', 'Symbol<never>();', 'typeof value === "array";', 'typeof value === "string";', 'typeof value === undefined;', 'typeof value !== null;', 'typeof value === 4;', 'typeof value === true;', 'typeof value === 2n;', 'typeof value === /a/;', 'typeof value === `array`;', 'typeof value === `string`;', 'typeof value === typeof other;', 'typeof value === name;', '"array" === (typeof value);', '(undefined) === typeof value;', 'function f(undefined:unknown){typeof value===undefined;}', 'const undefined=4;typeof value===undefined;', '/* 世界 🌍 */\r\nSymbol();typeof value===undefined;', 'typeof value === -1;', 'typeof value === `a${value}`;', 'typeof value === void 0;']
CONTROLS += ['async function f(){return 1;}', 'async function f(){}', 'async function f(){;}', 'async function f(){await x;}', 'async function f(){async function inner(){await x;}return 1;}', 'const f=async()=>1;', 'const f=async()=>await x;', 'const f=async(x:number)=>x;', 'const f=async function named(){return 1;};', 'const o={run:async function(){return 1;}};', 'const o={run:async()=>1};', 'const o={async run(){return 1;}};', 'class C{static async run(){return 1;}}', 'class C{field=async()=>1;}', 'class C{field=async function(){return 1;};}', 'async function* f(){yield 1;}', 'async function f(){for await(const x of [] as number[]){}}', 'async function f(){const x=1;}', 'export async function f(){return 1;}', 'export default async function named(){return 1;}', 'const f:()=>Promise<number>=async()=>1;', 'const f:()=>number|Promise<number>=async()=>1;', '[1,2].map(async x=>x+1);', 'function id<T>(x:T):T{return x;}id(async()=>1);', 'function invoke<T>(f:()=>Promise<T>):void{}invoke(async()=>1);', 'function invoke<T>(...f:(()=>T)[]):void{}invoke(async()=>1);', 'interface I{run():Promise<number>;}class C implements I{async run(){return 1;}}', 'interface I{run():number|Promise<number>;}class C implements I{async run(){return 1;}}', 'class C{a=0\nasync ["x"](){return 1;}}', 'class C{a=0\nasync in(){return 1;}}', 'class C{a\nasync ["x"](){return 1;}}', 'x\nasync()=>1;', '/* 世界 🌍 */\r\nasync function 漢(){return 1;}', 'async /* keep */ function f(){return 1;}']

CONTROLS += json.loads((OWN/'testdata/imported.json').read_text())

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--compiler', type=Path)
    parser.add_argument('--stage0', type=Path, required=True)
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
    listener=build('listener-probe',OWN/'listener_probe.a',linked)
    declared,_,_=run('listener-native',[listener]);expected,_,_=run('listener-go',[oracle,'--kinds']);assert declared==expected
    metadata=b''
    for name in ['symbol-description','valid-typeof','require-await']:
        record=json.loads((OWN/name/'rule.json').read_text());assert record['name']==name and all(type(kind) is str for kind in record['kinds'])
        metadata+=(name+'\t'+','.join(str(kind) for kind in sorted(record['kinds']))+'\n').encode()
    assert metadata==expected, 'rule.json listener mismatch'
    changed=json.loads((OWN/'symbol-description/rule.json').read_text());changed['kinds']=['BinaryExpression']
    (dest/'mutant-rule.json').write_text(json.dumps(changed))
    changed=json.loads((dest/'mutant-rule.json').read_text())
    mutated=(changed['name']+'\t'+','.join(str(kind) for kind in sorted(changed['kinds']))+'\n').encode()+metadata.split(b'\n',1)[1]
    assert mutated!=expected, 'JSON listener mutant survived'
    print('native/JSON production listener kinds match; JSON relevance mutant caught',flush=True)
    streams = dest/'source/system/StandardStreams.d.ts';streams.parent.mkdir(parents=True,exist_ok=True);streams.write_text('export declare function blockStandardStreams():void;\n')
    declarations = dest/'globals.d.ts'
    declarations.write_text("declare namespace NodeJS { interface Process { exit(code?:number):never; stdout:{write(text:string,callback?:()=>void):void}; stderr:{write(text:string,callback?:()=>void):void}; } }\ndeclare var process:NodeJS.Process;\n")
    config=dest/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'NodeNext','lib':['ES2022','DOM']},'files':[str(declarations)]}))
    paths=[]
    for index,source in enumerate(CONTROLS):
        path=dest/f'control-{index:03}.a';path.write_text(source+'\nexport {};\n');paths.append(path)
    manifest=dest/'controls.manifest';manifest.write_text(''.join(str(path)+'\n' for path in paths))
    valid,_,_=run('controls-valid-go',[oracle,config,manifest,'--valid-sources'])
    rejected=set(map(str,paths))-set(valid.decode().splitlines())
    (dest/'rejected-go-sources.json').write_text(json.dumps(sorted(rejected),indent=2)+'\n')
    manifest.write_bytes(valid)
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
    for name in ['symbol-description','valid-typeof','require-await']:
        assert name.encode() in truth, (name,'missing positive control')
    sanitized=archive('checker-asan',True);asan=build('native-asan',OWN/'suite.a',sanitized,True)
    compare('controls-asan',config,manifest,asan)
    changes=[('await','require_await.a',"Remove `async`. This changes","Remove `async`. MUTANT This changes"),('symbol','symbol_description.a','declarations[0]!==true','declarations[0]===true'),
             ('typeof','valid_typeof.a',"this.nonString('undefined')", "this.nonString('undefined')+' MUTANT'")]
    for label,file,original,replacement in changes:
        directory=dest/(label+'-source');directory.mkdir(exist_ok=True)
        for path in OWN.glob('*.a'):
            source=path.read_text()
            if path.name==file:
                assert source.count(original)==1
                source=source.replace(original,replacement)
            source=source.replace("'../", "'"+str(OWN.parent)+'/')
            (directory/path.name).write_text(source)
        mutant=build(label+'-mutant',directory/'suite.a',linked)
        got,stderr,_=run(label+'-run',[mutant,config,manifest]);assert not stderr and got!=truth,(label,'mutant survived')
        print(label,'mutant exits 0, byte oracle catches',flush=True)
    for label,file,original,replacement in [
        ('generic-question','generic_call_signature.go','declared = resolved.Target()', 'declared = nil'),
        ('index-question','type_index.go','index := checker.Checker_numberType(c)', 'index := checker.Checker_stringType(c)'),
        ('heritage-question','heritage_types.go','roots = append(roots, g.add(value))', 'g.add(value)'),
    ]:
        originalPath=ROOT/'bridge/tsgo/checker'/file
        source=originalPath.read_text();assert source.count(original)==1,(label,'nonunique mutation')
        side=dest/(label+'.go');side.write_text(source.replace(original,replacement))
        overlay=dest/(label+'.json');overlay.write_text(json.dumps({'Replace':{str(originalPath):str(side)}}))
        mutantArchive=archive(label,overlay=overlay);mutant=build(label+'-native',OWN/'suite.a',mutantArchive)
        got,stderr,_=run(label+'-run',[mutant,config,manifest]);assert not stderr and got!=truth,(label,'mutant survived')
        print(label,'mutant exits 0, byte oracle catches',flush=True)
    for corpus,sourceRoot,sourceConfig in [('repository',ROOT,ROOT/'tsconfig.json'),('compiler',args.compiler,args.compiler/'src/compiler/tsconfig.json' if args.compiler else None)]:
        if sourceRoot is None: continue
        portable=ROOT/'stage1/cohere/typeaware/validation-coverage'/f'{corpus}.manifest'
        resolved=dest/f'{corpus}.manifest';resolved.write_text(''.join(str(sourceRoot/path)+'\n' for path in portable.read_text().splitlines()))
        compare(corpus,sourceConfig,resolved,native);compare(corpus+'-asan',sourceConfig,resolved,asan)
    released=dest/'released.a'
    released.write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,6,'CallExpression',args[2]??''));\n")
    probe=dest/'probe.a';probe.write_text('id(1);\n');stale=build('released',released,linked)
    for question in ['generic-call-signature','type-index\n1\nnumber','heritage-types']:
        _,stderr,_=run('released-run',[stale,config,probe,question],expected=70)
        assert stderr==b'adamic: panic: invalid or released checker handle\n'
    print('released new questions: exact panic 70',flush=True)
    print('PASS',flush=True)


if __name__ == '__main__': main()

#!/usr/bin/env python3
"""Compare owned numeric JSX handlers with unmodified production Go rules."""
import argparse
import hashlib
import json
import os
import pathlib
import re
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('artifacts', type=pathlib.Path)
parser.add_argument('--compiler', type=pathlib.Path, required=True)
args = parser.parse_args()
ROOT = pathlib.Path(__file__).resolve().parents[4]
UNIT = pathlib.Path(__file__).resolve().parent
OUT = args.artifacts.resolve()
OUT.mkdir(exist_ok=True)
PRIVATE = OUT/'native'
PRIVATE.mkdir(exist_ok=True)
DEPENDENCY = pathlib.Path('/workspace/wave-07-jsx-dependency')
PARSER_REVISION = 'a8a62d62ca49db7415e14c3887dd305022b17309'
STAGE0 = pathlib.Path('/workspace/wave-07-next-rest/adamic')
ARCHIVE = pathlib.Path('/workspace/wave-07-next-rest/checker.a')
SAN_ARCHIVE = pathlib.Path('/workspace/wave-07-next-rest/checker-asan.a')
records = []

def run(name, command, cwd=ROOT, expected=0):
    started = time.monotonic_ns()
    with (OUT/(name+'.stdout')).open('wb') as stdout, (OUT/(name+'.stderr')).open('wb') as stderr:
        result = subprocess.run([str(value) for value in command], cwd=cwd, stdout=stdout, stderr=stderr)
    records.append(dict(name=name,command=[str(value) for value in command],exit=result.returncode,elapsed_ns=time.monotonic_ns()-started))
    (OUT/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
    assert result.returncode == expected, (name,result.returncode,expected)
    return (OUT/(name+'.stdout')).read_bytes(), (OUT/(name+'.stderr')).read_bytes()

# Reproduce the exact published dependency, never modifying shared branch files.
files=subprocess.check_output(['git','ls-tree','-r','--name-only',PARSER_REVISION,'stage1/typescript'],cwd=ROOT,text=True).splitlines()
pins={}
for name in files:
    if not name.endswith(('.ts','.a')): continue
    content=subprocess.check_output(['git','show',PARSER_REVISION+':'+name],cwd=ROOT)
    target=DEPENDENCY/name;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(content)
    pins[name]=hashlib.sha256(content).hexdigest()
(OUT/'parser-pins.json').write_text(json.dumps(dict(revision=PARSER_REVISION,sha256=pins),indent=2)+'\n')

def private_text(text):
    for relative in ['../../../typescript/parser/parser.ts','../../../typescript/scanner/scanner.ts']:
        text=text.replace(relative,str(DEPENDENCY/'stage1/typescript'/('/'.join(relative.split('/')[4:]))))
    for helper in ['unary_minus','diagnostic','frames']:
        text=text.replace("'../"+helper+".ts'","'"+str(UNIT.parent/(helper+'.ts'))+"'")
    return text.replace("'../../../typescript/parser/nodes.ts'","'"+str(ROOT/'stage1/typescript/parser/nodes.ts')+"'")

for path in UNIT.rglob('*.a'):
    target=PRIVATE/path.relative_to(UNIT);target.parent.mkdir(parents=True,exist_ok=True)
    target.write_text(private_text(path.read_text()))

run('native-build',[STAGE0,'build',PRIVATE/'suite.a','-o',OUT/'native-suite','--tsgo',ARCHIVE])
run('native-asan-build',[STAGE0,'build',PRIVATE/'suite.a','-o',OUT/'native-asan','--tsgo',SAN_ARCHIVE,'--sanitize'])
virtual=ROOT/'cohere/adamic_wave07_jsx.go'
(OUT/'oracle.json').write_text(json.dumps({'Replace':{str(virtual):str(UNIT/'testdata/oracle.go')}}))
run('oracle-build',['go','build','-overlay',OUT/'oracle.json','-o',OUT/'oracle',virtual],ROOT/'cohere')
(OUT/'ambient.d.ts').write_text('export {};\n')
config=OUT/'tsconfig.json'
config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','jsx':'preserve','lib':['es2022','dom'],'allowJs':True},'files':['ambient.d.ts']}))
common='declare const React:any;declare const flag:boolean;declare const app:any;\n'
controls=[
    'const x=<React.Fragment/>;',
    'const x=<React.Fragment><div/></React.Fragment>;',
    'const x=<React.Fragment key="k"/>;',
    'const x=<React.Fragment {...{}}/>;',
    'const x=<React.Fragment<number>/>;',
    'const x=<><div/></>;',
    'const x=<></>;',
    'const x=<A.React.Fragment/>;',
    'const x=<React.Other/>;',
    'import {Fragment} from "react";const x=<Fragment/>;',
    'import {Fragment as F} from "react";const x=<F/>;',
    'import type {Fragment as F} from "react";const x=<F/>;',
    'import {type Fragment as F} from "react";const x=<F/>;',
    'import {Other as F} from "react";const x=<F/>;',
    'import {Fragment as F} from "preact";const x=<F/>;',
    'const F=React.Fragment;const x=<F/>;',
    'const F=React;const x=<F/>;',
    'const F=React.Other;const x=<F/>;',
    'const F=require("react");const x=<F/>;',
    'const F=require(`react`);const x=<F/>;',
    'const F=require("react").Fragment;const x=<F/>;',
    'const {Fragment}=React;const x=<Fragment/>;',
    'const {Fragment:F}=React;const x=<F/>;',
    'const {Anything}=React;const x=<Anything/>;',
    'const {Fragment}=require("react");const x=<Fragment/>;',
    'const {Fragment}=require("preact");const x=<Fragment/>;',
    'const F:unknown = /* comment */ React.Fragment;const x=<F/>;',
    'let F;const x=<F/>;',
    'const F=React.Fragment;function nested(){const F=null;return <F/>;}',
    'declare function Fragment():void;import {Fragment} from "react";const x=<Fragment/>;',
    'import {Fragment} from "react";declare function Fragment():void;const x=<Fragment/>;',
    'const x=<Missing/>;',
    'const x=<div/>;',
    'const x=<x-gif/>;',
    'const x=<Foo-bar/>;',
    'const x=<_foo/>;',
    'const x=<$foo/>;',
    'const x=<테스트/>;',
    'const x=<Map/>;',
    'const x=<appp.Foo/>;',
    'const x=<appp.foo.Bar/>;',
    'const x=<app.Foo/>;',
    'const x=<Apppp:Foo/>;',
    'const x=<this.props.tag/>;',
    'const x=<this/>;',
    'enum A{App}const x=<App/>;',
    'enum A{App}var App;const x=<App/>;',
    'import App=require("./app");const x=<App/>;',
    'declare namespace Foo {class App{}}import App=Foo.App;const x=<App/>;',
    '{const App=null;const x=<App/>;}const y=<App/>;',
    'declare global {var Global:any;}const x=<Global/>;',
    'function f<T>(){return <T/>;}',
    'const x=<div><a/><a/></div>;',
    'const x=<div><a/> <a/></div>;',
    'const x=<div><a/>x<a/></div>;',
    'const x=<div><a/>&nbsp;<a/></div>;',
    'const x=<div><a/>{flag}<a/></div>;',
    'const x=<div><a/>{/*comment*/}<a/></div>;',
    'const x=<div><a/><span/><a/></div>;',
    'const x=<div><a/><span/>x<a/><span/></div>;',
    'const x=<div><div/><a/></div>;',
    'const x=<div><App/><App/></div>;',
    'const x=<div><a.b/><a.b/></div>;',
    'const x=<><a/><span/></>;',
    'const x=<React.Fragment><a/><span/></React.Fragment>;',
    'const x=<div><a/><><span/></></div>;',
    'React.createElement("div",undefined,[React.createElement("a"),React.createElement("span")]);',
    'React.createElement("div",undefined,[NotReact.createElement("a"),NotReact.createElement("span")]);',
    'NotReact.createElement("div",undefined,[React.createElement("a"),React.createElement("span")]);',
    'React.createElement("div",undefined,[foo(),React.createElement("a")]);',
    'React.createElement("div",undefined,React.createElement("a"));',
    'React.createElement("div");',
    '(React.createElement)("div",undefined,[foo("a"),foo("span")]);',
    'import {createElement} from "react";createElement("div",undefined,[foo("a"),foo("span")]);',
    'import {createElement} from "preact";createElement("div",undefined,[foo("a"),foo("span")]);',
    'import {other as createElement} from "react";createElement("div",undefined,[foo("a"),foo("span")]);',
    'const {createElement}=React;createElement("div",undefined,[foo("a"),foo("span")]);',
    'const {createElement}=require("react");createElement("div",undefined,[foo("a"),foo("span")]);',
    'const createElement=React.whatever;createElement("div",undefined,[foo("a"),foo("span")]);',
    'const createElement=require("react").whatever;createElement("div",undefined,[foo("a"),foo("span")]);',
    'declare function createElement(a:string):void;import {createElement} from "react";createElement("div",undefined,[foo("a"),foo("span")]);',
    'import {createElement} from "react";declare function createElement(a:string):void;createElement("div",undefined,[foo("a"),foo("span")]);',
    'React.createElement("div",undefined,[,foo("a"),foo("span")]);',
    '/*世界 🌍*/\r\nconst é=<React.Fragment><a/><span/></React.Fragment>;',
]
for literal in ['""','"x"','" "','"x "','" x"','1','0x10','123n','true','false','null','/x/','`x`','-1']:
    controls.append('React.createElement("div",undefined,['+literal+',foo("a")]);')
for code in [9,10,13,32,133,160,5760,8192,8202,8232,8233,8239,8287,12288,65279]:
    for edge in ['leading','trailing']:
        value=chr(code)+'x' if edge=='leading' else 'x'+chr(code)
        controls.append('React.createElement("div",undefined,['+json.dumps(value)+',foo("a")]);')
paths=[]
for index,source in enumerate(controls):
    path=OUT/f'control-{index:03}.tsx';path.write_text(common+source+'\nexport {};\n');paths.append(path)
# A script declaration in another root tests declaration origins and private source loading.
shared=OUT/'shared.a';shared.write_text('declare const React:any;const GlobalFragment=React.Fragment;\n')
paths.append(shared)
path=OUT/'external.tsx';path.write_text('const x=<GlobalFragment/>;\nexport {};\n');paths.append(path)
path=OUT/'commonjs.cjs';path.write_text('const x=<Map/>;\n');paths.append(path)
manifest=OUT/'controls.manifest';manifest.write_text(''.join(str(path)+'\n' for path in paths))

def compare(label, native, config, manifest, selected='', options=()):
    truth,_=run(label+'-go',[OUT/'oracle',config,manifest,selected,*options])
    actual,error=run(label+'-native',[native,config,manifest,selected,*options])
    if truth!=actual or error:
        mismatch=next((index for index,(a,b) in enumerate(zip(truth,actual)) if a!=b),min(len(truth),len(actual)))
        raise RuntimeError(f'{label}: mismatch at {mismatch}, stderr {len(error)} bytes')
    print(label+': '+str(len(actual))+' identical bytes; '+actual.splitlines()[-1].decode(),flush=True)
    return truth

truth=compare('controls',OUT/'native-suite',config,manifest)
for name in ['jsx-fragments','jsx-no-undef','no-adjacent-inline-elements']:
    assert ('\treact/'+name+'\t').encode() in truth, name+' has no positive control'
compare('controls-asan',OUT/'native-asan',config,manifest)
for label,options in [('element',['--element']),('globals',['--allow-globals']),('both',['--element','--allow-globals'])]:
    compare('controls-'+label,OUT/'native-suite',config,manifest,options=options)
    compare('controls-'+label+'-asan',OUT/'native-asan',config,manifest,options=options)
for name,base,config_path in [('repository',ROOT,ROOT/'tsconfig.json'),('compiler',args.compiler,args.compiler/'src/compiler/tsconfig.json')]:
    lines=(ROOT/f'stage1/cohere/typeaware/validation-volume/{name}.manifest').read_text().splitlines()
    corpus=OUT/(name+'.manifest');corpus.write_text(''.join(str(base/line)+'\n' for line in lines))
    compare(name,OUT/'native-suite',config_path,corpus)
    compare(name+'-asan',OUT/'native-asan',config_path,corpus)
for name,token in [('jsx-fragments',"'preferFragment'"),('jsx-no-undef',"'jsxIdentifierNotDefined'"),('no-adjacent-inline-elements',"'inlineElement'")]:
    path=PRIVATE/'rules'/name/'rule.a';original=path.read_text()
    assert original.count(token)==1
    path.write_text(original.replace(token,token[:-1]+"Mutant'"))
    run(name+'-mutant-build',[STAGE0,'build',PRIVATE/'suite.a','-o',OUT/'mutant','--tsgo',ARCHIVE])
    changed,error=run(name+'-mutant',[OUT/'mutant',config,manifest])
    assert not error and changed!=truth, name+' mutant survived'
    path.write_text(original)
    print(name+': message mutant exits 0; caught only by production Go bytes',flush=True)
# Hold declarations to actual live production Run listener keys, and to their manifests.
expected,_=run('listeners-go',[OUT/'oracle',config,manifest,'--listeners'])
run('listeners-build',[STAGE0,'build',PRIVATE/'listeners_probe.a','-o',OUT/'listeners','--tsgo',ARCHIVE])
actual,error=run('listeners-native',[OUT/'listeners'])
assert expected==actual and not error
for line in expected.decode().splitlines():
    name,*keys=line.split('\t')
    descriptor=json.loads((UNIT/'rules'/name.split('/')[1]/'rule.json').read_text())
    assert descriptor['kinds']==list(map(int,keys))
print('listener manifests and compiled declarations: production Go keys match',flush=True)
probe_source=OUT/'released.a';probe_source.write_text('declare let value:number;value;\nexport {};\n')
run('released-build',[STAGE0,'build',PRIVATE/'released_probe.a','-o',OUT/'released','--tsgo',ARCHIVE])
output,error=run('released',[OUT/'released',config,probe_source],expected=70)
assert output==b'live symbol succeeds\n' and b'invalid or released checker handle' in error
print('symbol metadata: live success, released handle panic 70',flush=True)
# A listener mutation must alter real dispatch, not merely metadata output.
for name in ['jsx-fragments','jsx-no-undef','no-adjacent-inline-elements']:
    path=PRIVATE/'rules'/name/'rule.a';original=path.read_text()
    keys=re.search(r'listenerKinds: readonly number\[\] = \[([^]]+)\]',original)
    assert keys
    mutated=original[:keys.start(1)]+original[keys.start(1):keys.end(1)].replace(keys.group(1).split(',')[0],'0',1)+original[keys.end(1):]
    path.write_text(mutated)
    run(name+'-listener-mutant-build',[STAGE0,'build',PRIVATE/'suite.a','-o',OUT/'mutant','--tsgo',ARCHIVE])
    changed,error=run(name+'-listener-mutant',[OUT/'mutant',config,manifest])
    assert not error and changed!=truth, name+' listener mutant survived'
    path.write_text(original)
    print(name+': numeric listener mutant exits 0; Go bytes catch missed dispatch',flush=True)
print('PASS isolated JSX rule gate; published parser integration remains pending.',flush=True)

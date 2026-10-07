"""Compare actual handed JSX nodes and raw checker facts with independent Go."""
import argparse,json,pathlib,subprocess,time
OWN=pathlib.Path(__file__).resolve().parent;ROOT=OWN.parents[3]
p=argparse.ArgumentParser();p.add_argument('--prepared',required=True);p.add_argument('--parser-root',required=True);args=p.parse_args();S=pathlib.Path(args.prepared);records=[]
def run(name,cmd,expect=0):
 t=time.monotonic()
 with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:r=subprocess.run([str(c) for c in cmd],cwd=ROOT,stdout=out,stderr=err)
 records.append(dict(name=name,command=[str(c) for c in cmd],exit=r.returncode,seconds=time.monotonic()-t));(S/'all-source-runs.json').write_text(json.dumps(records,indent=2)+'\n');assert r.returncode==expect,(name,r.returncode,(S/(name+'.stderr')).read_text());return (S/(name+'.stdout')).read_bytes()
q=json.dumps
imports=f'import {{Parser}} from {q(str(pathlib.Path(args.parser_root)/"stage1/typescript/parser/parser.ts"))};import {{SourceFragments}} from {q(str(OWN/"jsx_fragments/source.a"))};import {{SourceNoUndef}} from {q(str(OWN/"jsx_no_undef/source.a"))};import {{CheckedSourceAdjacent}} from {q(str(OWN/"no_adjacent_inline_elements/checked_source.a"))};import {{panic,tsgoProgram,tsgoRelease,programArguments,readTextFile}} from "adamic";\n'
dispatch='''const parser=new Parser(text,file);parser.file();const lines:string[]=[];for(const node of parser.nodes){
if(node.kind==="JsxElement"){const opening=parser.nodes[node.children[0]??-1]??panic("missing opening");for(const finding of fragments.named(node,opening,parser.nodes,text,file,program)){lines.push(finding.written());count++;}for(const finding of adjacent.jsx(node,parser.nodes,text)){lines.push(finding.written());count++;}}
else if(node.kind==="JsxSelfClosingElement"){for(const finding of fragments.named(node,node,parser.nodes,text,file,program)){lines.push(finding.written());count++;}for(const finding of undef.visit(node,parser.nodes,text,file,program)){lines.push(finding.written());count++;}}
else if(node.kind==="JsxOpeningElement"){for(const finding of undef.visit(node,parser.nodes,text,file,program)){lines.push(finding.written());count++;}}
else if(node.kind==="JsxFragment"){for(const finding of fragments.fragment(node,text)){lines.push(finding.written());count++;}}
else if(node.kind==="CallExpression"){for(const finding of adjacent.call(node,parser.nodes,text,file,program)){lines.push(finding.written());count++;}}
}lines.sort((left,right)=>left<right?-1:left>right?1:0);for(const line of lines){console.log(line);}
'''
extra=[
 'import {"Fragment" as F} from "react";const value=<F />;',
 'import {"createElement" as createElement} from "react";const value=createElement("div",null,[React.createElement("a"),React.createElement("span")]);',
 "import {createElement} from 'react';const value=createElement('div',null,[createElement('a'),createElement('span')]);",
 "import {createElement} from 'other';const value=createElement('div',null,[createElement('a'),createElement('span')]);",
 "const {createElement}=React;const value=createElement('div',null,[React.createElement('a'),React.createElement('span')]);",
 "const createElement=require('react').createElement;const value=createElement('div',null,[React.createElement('a'),React.createElement('span')]);",
 "const createElement=require('other').createElement;const value=createElement('div',null,[React.createElement('a'),React.createElement('span')]);",
 "const {Fragment:F}=require('react');const value=<F />;",
 "const {Fragment:F}=require('other');const value=<F />;",
 "const F=require('react');const value=<F />;",
 "const F=React;const value=<F />;",
 "const F=React.NotFragment;const value=<F />;",
 "const F=require(`react`);const value=<F />;",
 "const F=React.Fragment;const value=<F key='k' />;",
 "/* 🦊한 */const value=<Missing />;",
 "/* 🦊한 */const value=<div><a/><span/></div>;",
 "/* 🦊한 */const value=<React.Fragment />;",
]
for literal in ['0','1n','true','false','null','/ x /','`x`','(1)']:
 extra.append('const value=React.createElement("div",null,['+literal+',React.createElement("span")]);')
extra.extend([
 'const value=(React.createElement)("div",null,[React.createElement("a"),React.createElement("span")]);',
 'declare function createElement(a:string):void;import {createElement} from "react";const value=createElement("div",null,[React.createElement("a"),React.createElement("span")]);',
 'import {createElement} from "react";declare function createElement(a:string):void;const value=createElement("div",null,[React.createElement("a"),React.createElement("span")]);',
 'const {NotFragment:F}=React;const value=<F />;',
 'const [F]=React;const value=<F />;',
])
files=[S/f'control-{i}.tsx' for i in range(61)]
for at,body in enumerate(extra):
 f=S/f'source-extra-{at}.tsx';f.write_text('declare const React:any;declare const require:any;'+body+'\nexport {};\n');files.append(f)
manifest=S/'all-source.manifest';manifest.write_text(''.join(str(f)+'\n' for f in files));truth=run('all-source-go',[S/'oracle',S/'tsconfig.json',manifest]);archive=S/'checker.a';run('all-source-archive',['go','build','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'])
setup='let count=0;const fragments=new SourceFragments();const undef=new SourceNoUndef();const adjacent=new CheckedSourceAdjacent();\n'
source=imports+'const program=tsgoProgram('+q(str(S/'tsconfig.json'))+','+q([str(f) for f in files])+');'+setup
for f in files:source+='{const file='+q(str(f))+';const text='+q(f.read_text())+';console.log("file\\t"+file);'+dispatch+'}\n'
source+='tsgoRelease(program);console.log(`findings ${count}`);';entry=S/'all_source_controls.a';entry.write_text(source)
for name,sanitize in [('all-source-native',False),('all-source-asan',True)]:
 exe=S/name;run(name+'-build',[S/'adamic','build',entry,'-o',exe,'--tsgo',archive]+(['--sanitize'] if sanitize else []));actual=run(name,[exe]);assert actual==truth,(name,actual.decode(),truth.decode());assert (S/(name+'.stderr')).read_bytes()==b''
print('PASS all actual source/checker:',len(files),'controls;',truth.decode().splitlines()[-1],';',len(truth),'identical Go/native/sanitized bytes',flush=True)
# The new question must be capable of producing an observable wrong answer.
go=ROOT/'bridge/tsgo/checker/wave06_jsx_bindings.go';text=go.read_text();assert 'out.text(imported)' in text;mutant=S/'jsx-bindings-mutant.go';mutant.write_text(text.replace('out.text(imported)','out.text(imported + "wrong")'));overlay=S/'jsx-bindings-mutant.json';overlay.write_text(json.dumps({'Replace':{str(go):str(mutant)}}));mutant_archive=S/'jsx-bindings-mutant.a';run('all-source-question-mutant-archive',['go','build','-overlay='+str(overlay),'-buildmode=c-archive','-o',mutant_archive,'./bridge/tsgo/archive']);exe=S/'all-source-question-mutant';run('all-source-question-mutant-build',[S/'adamic','build',entry,'-o',exe,'--tsgo',mutant_archive]);assert run('all-source-question-mutant',[exe])!=truth;assert (S/'all-source-question-mutant.stderr').read_bytes()==b''
print('Raw imported-name question mutant exits zero and is caught only by Go bytes',flush=True)
# A quoted import must retain StringLiteral rather than masquerade as Identifier.
text=go.read_text();assert 'out.text(importedKind)' in text;mutant=S/'jsx-binding-kind-mutant.go';mutant.write_text(text.replace('out.text(importedKind)','if importedKind == "StringLiteral" { importedKind = "Identifier" }; out.text(importedKind)'));overlay=S/'jsx-binding-kind-mutant.json';overlay.write_text(json.dumps({'Replace':{str(go):str(mutant)}}));mutant_archive=S/'jsx-binding-kind-mutant.a';run('all-source-kind-mutant-archive',['go','build','-overlay='+str(overlay),'-buildmode=c-archive','-o',mutant_archive,'./bridge/tsgo/archive']);exe=S/'all-source-kind-mutant';run('all-source-kind-mutant-build',[S/'adamic','build',entry,'-o',exe,'--tsgo',mutant_archive]);assert run('all-source-kind-mutant',[exe])!=truth;assert (S/'all-source-kind-mutant.stderr').read_bytes()==b''
print('Raw imported-kind mutant exits zero; quoted import control catches it only through Go bytes',flush=True)

# Released handles must be rejected by the newly added question too.
f=files[4];stale=S/'bindings_released.a';stale.write_text(imports+'const file='+q(str(f))+';const text='+q(f.read_text())+';const program=tsgoProgram('+q(str(S/'tsconfig.json'))+',[file]);const parser=new Parser(text,file);parser.file();tsgoRelease(program);for(const node of parser.nodes){if(node.kind==="JsxSelfClosingElement"){new SourceFragments().named(node,node,parser.nodes,text,file,program);}}')
exe=S/'bindings-released';run('bindings-released-build',[S/'adamic','build',stale,'-o',exe,'--tsgo',archive]);assert run('bindings-released',[exe],70)==b'';assert b'invalid or released checker handle' in (S/'bindings-released.stderr').read_bytes()
retained=S/'bindings-retained.a';run('bindings-retained-archive',['go','build','-overlay=/workspace/wave-06-landing3-original/released-registry.json','-buildmode=c-archive','-o',retained,'./bridge/tsgo/archive']);exe=S/'bindings-retained';run('bindings-retained-build',[S/'adamic','build',stale,'-o',exe,'--tsgo',retained]);assert run('bindings-retained',[exe])==b'';assert (S/'bindings-retained.stderr').read_bytes()==b''
print('New question rejects released handles with exit 70; retained registry mutant exits zero',flush=True)
# Full source pipeline on the frozen compiler and repository manifests.
source=imports+'const args=programArguments();const config=args[0]??panic("missing config");const manifest=readTextFile(args[1]??panic("missing manifest"));if(manifest.kind!=="Ok"){panic("missing manifest");}const files:string[]=[];for(const file of manifest.text.split("\\n")){if(file!==""){files.push(file);}}const program=tsgoProgram(config,files);'+setup+'for(const file of files){const input=readTextFile(file);if(input.kind!=="Ok"){panic("missing source");}const text=input.text;console.log("file\\t"+file);'+dispatch+'}tsgoRelease(program);console.log(`findings ${count}`);'
entry=S/'all_source_corpora.a';entry.write_text(source)
for name,sanitize in [('all-corpus-native',False),('all-corpus-asan',True)]:run(name+'-build',[S/'adamic','build',entry,'-o',S/name,'--tsgo',archive]+(['--sanitize'] if sanitize else []))
for corpus,base,config in [('compiler',pathlib.Path('/workspace/wave-06-typescript'),pathlib.Path('/workspace/wave-06-typescript/src/compiler/tsconfig.json')),('repository',ROOT,ROOT/'tsconfig.json')]:
 paths=(ROOT/f'stage1/cohere/typeaware/validation-coverage/{corpus}.manifest').read_text().splitlines();manifest=S/(corpus+'-all-source.manifest');manifest.write_text(''.join(str(base/path)+'\n' for path in paths if path));truth=run(corpus+'-all-source-go',[S/'oracle',config,manifest]);assert truth.endswith(b'findings 0\n')
 for name in ['all-corpus-native','all-corpus-asan']:
  assert run(corpus+'-'+name,[S/name,config,manifest])==truth,corpus+' '+name;assert (S/(corpus+'-'+name+'.stderr')).read_bytes()==b''
 print(corpus,len(paths),'source roots;',len(truth),'identical Go/native/sanitized bytes',flush=True)

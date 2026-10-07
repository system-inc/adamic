"""Independent Go comparison for private prepared-node kernels, not source integration."""
import argparse,json,pathlib,re,subprocess,time,shutil,gzip,hashlib
OWN=pathlib.Path(__file__).resolve().parent
ROOT=OWN.parents[3]
p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);args=p.parse_args();S=pathlib.Path(args.scratch).resolve();S.mkdir(parents=True,exist_ok=True);runs=[]
def run(name,cmd,expect=0):
 t=time.monotonic()
 with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:r=subprocess.run([str(x) for x in cmd],cwd=ROOT,stdout=out,stderr=err)
 runs.append(dict(name=name,command=[str(x) for x in cmd],exit=r.returncode,seconds=time.monotonic()-t));(S/'runs.json').write_text(json.dumps(runs,indent=2)+'\n')
 assert r.returncode==expect,(name,r.returncode,(S/(name+'.stderr')).read_text());return (S/(name+'.stdout')).read_bytes()
K={name:int(value) for name,value in re.findall(r'export const (\w+) = (\d+);',(OWN/'kinds.a').read_text())}
q=lambda x:json.dumps(x,ensure_ascii=True)
# Raw fixture facts are constructed independently from source; Go findings are never inputs.
cases=[]
def case(source,code):cases.append((source,code))
def span(source,text):
 start=source.index(text);return len(source[:start].encode()),len(source[:start+len(text)].encode())
def element(source,text,tag,attrs=0,decl=''):
 start,end=span(source,text)
 return f'const n=new ElementInput();n.start={start};n.end={end};n.attributes={attrs};'+tag+decl
pref='declare const React:any;'
for text,tag,attrs in [('<React.Fragment />',"n.tag.kind=K.PropertyAccessExpression;n.tag.receiverKind=K.Identifier;n.tag.receiverText='React';n.tag.text='Fragment';n.tag.depth=1;",0),('<React.Fragment key="k" />',"n.tag.kind=K.PropertyAccessExpression;n.tag.receiverKind=K.Identifier;n.tag.receiverText='React';n.tag.text='Fragment';n.tag.depth=1;",1),('<React.Fragment></React.Fragment>',"n.tag.kind=K.PropertyAccessExpression;n.tag.receiverKind=K.Identifier;n.tag.receiverText='React';n.tag.text='Fragment';n.tag.depth=1;",0),('<React.Other />',"n.tag.kind=K.PropertyAccessExpression;n.tag.receiverKind=K.Identifier;n.tag.receiverText='React';n.tag.text='Other';n.tag.depth=1;",0)]:
 source=pref+'const value='+text+';';case(source,element(source,text,tag,attrs)+"emit(new JsxFragments().named(n));")
for binding,decl in [("import {Fragment as F} from 'react';","d.kind=K.ImportSpecifier;d.imported='Fragment';d.module='react';"),("import {Fragment as F} from 'preact';","d.kind=K.ImportSpecifier;d.imported='Fragment';d.module='preact';"),("const F=React.Fragment;","d.kind=K.VariableDeclaration;d.initializerKind=K.PropertyAccessExpression;d.receiverKind=K.Identifier;d.receiverName='React';d.initializerName='Fragment';"),("const {Fragment:F}=React;","d.kind=K.BindingElement;d.objectBinding=true;d.variableParent=true;d.initializerKind=K.Identifier;d.initializerName='React';")]:
 source=pref+binding+'const value=<F />;';case(source,element(source,'<F />',"n.tag.kind=K.Identifier;n.tag.text='F';",0,'const d=new DeclarationInput();'+decl+'n.declarations.push(d);')+'emit(new JsxFragments().named(n));')
# Symbol facts include same-file declarations, external declarations, absent symbols and member roots.
for name,declared,external in [('Missing',False,False),('Known',True,False),('div',False,False),('Foo-bar',False,False),('_foo',False,False),('$foo',False,False),('테스트',False,False),('Map',True,True),('app.Foo',False,False),('app.Foo.Bar',False,False),('this.foo',False,False)]:
 source=('declare const Known:any;' if name=='Known' else '')+'const value=<'+name+' />;'
 root=name.split('.')[0];start,end=span(source,root)
 tag='const t=new TagInput();'+f't.kind=K.{"PropertyAccessExpression" if "." in name else "Identifier"};t.text={q(name)};t.receiverKind={0 if root=="this" else K["Identifier"]};t.receiverText={q(root)};t.start={start};t.end={end};'
 case(source,tag+f'emit(new JsxNoUndef().visit(t,true,{str(declared).lower()},DECLFILES,FILE));'.replace('DECLFILES', '[FILE]' if declared and not external else "['external.d.ts']" if external else '[]'))
inline=['a','b','big','i','small','tt','abbr','acronym','cite','code','dfn','em','kbd','strong','samp','time','var','bdo','br','img','map','object','q','script','span','sub','sup','button','input','label','select','textarea']
for name in inline:
 text=f'<div><{name} /><span /></div>';source='const value='+text+';';start,end=span(source,text)
 code=f'const n=new ContainerInput();n.start={start};n.end={end};'
 for tag in [name,'span']:code+=f'{{const c=new ChildInput();c.kind=K.JsxSelfClosingElement;c.tagKind=K.Identifier;c.tagText={q(tag)};n.children.push(c);}}'
 case(source,code+'emit(new NoAdjacentInlineElements().jsx(n));')
for mid in ['x',' ','{0}']:
 text='<div><a />'+mid+'<span /></div>';source='const value='+text+';';start,end=span(source,text)
 code=f'const n=new ContainerInput();n.start={start};n.end={end};'
 for kind,tag in [('JsxSelfClosingElement','a'),('Identifier',''),('JsxSelfClosingElement','span')]:code+=f'{{const c=new ChildInput();c.kind=K.{kind};c.tagKind=K.Identifier;c.tagText={q(tag)};n.children.push(c);}}'
 case(source,code+'emit(new NoAdjacentInlineElements().jsx(n));')
for literal in ['', 'x', ' ', 'x ', ' x', '\u0085', '\ufeff']:
 text='React.createElement("div",null,[React.createElement("a"),'+q(literal)+',React.createElement("span")])';source=pref+'const value='+text+';';start,end=span(source,text)
 code=f'const n=new ContainerInput();n.start={start};n.end={end};n.argumentsCount=3;n.thirdKind=K.ArrayLiteralExpression;n.callee.kind=K.PropertyAccessExpression;n.callee.receiverKind=K.Identifier;n.callee.receiverText="React";n.callee.text="createElement";'
 for tag in ['a',None,'span']:
  code+=' {const c=new ChildInput();'+(f'c.kind=K.StringLiteral;c.text={q(literal)};' if tag is None else f'c.kind=K.CallExpression;c.firstKind=K.StringLiteral;c.firstText={q(tag)};')+'n.children.push(c);}'
 case(source,code+'emit(new NoAdjacentInlineElements().call(n));')
anchor=S/'anchor.d.ts';anchor.write_text('export {};');config=S/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'ESNext','jsx':'preserve'},'files':[str(anchor)]}))
files=[]
for i,(source,code) in enumerate(cases):f=S/f'control-{i}.tsx';f.write_text(source+'\nexport {};\n');files.append(f)
manifest=S/'controls.manifest';manifest.write_text(''.join(str(f)+'\n' for f in files))
virtual=ROOT/'cohere/adamic_wave06jsxoracle.go';overlay=S/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/oracle.go')}}))
oracle=S/'oracle';stage0=S/'adamic';run('stage0-build',['go','build','-o',stage0,'./cmd/adamic']);run('oracle-build',['go','-C',ROOT/'cohere','build','-overlay',overlay,'-o',oracle,virtual]);truth=run('go-controls',[oracle,config,manifest])
assert b'findings 0\n' not in truth
source=''
for cls,folder in [('JsxFragments','jsx_fragments'),('JsxNoUndef','jsx_no_undef'),('NoAdjacentInlineElements','no_adjacent_inline_elements')]:source+=f'import {{{cls}}} from {q(str(OWN/folder/"index.a"))};\n'
source+=f'import {{ElementInput,TagInput,ContainerInput,ChildInput,DeclarationInput}} from {q(str(OWN/"inputs.a"))};import * as K from {q(str(OWN/"kinds.a"))};import {{Diagnostic}} from {q(str(ROOT/"stage1/cohere/typeaware/diagnostic.ts"))};\n'
source=re.sub(r'import \* as K from ([^;]+);',lambda m:'import {'+', '.join(sorted(K))+'} from '+m.group(1)+';',source)
source+='let total=0;function emit(findings:Diagnostic[]):void{for(const finding of findings){console.log(finding.written());total++;}}\n'
for f,(_,code) in zip(files,cases):source+='{const FILE='+q(str(f))+';console.log("file\\t"+FILE);'+code+'}\n'
source+='console.log(`findings ${total}`);';source=re.sub(r'K\.(\w+)',r'\1',source);probe=S/'prepared_controls.a';probe.write_text(source)
for sanitize in [False,True]:
 name='native-asan' if sanitize else 'native';exe=S/name;run(name+'-build',[stage0,'build',probe,'-o',exe]+(['--sanitize'] if sanitize else []));output=run(name,[exe]);assert output==truth,(name,output.decode(),truth.decode());assert (S/(name+'.stderr')).read_bytes()==b''
print('prepared controls:',len(files),'sources;',truth.decode().splitlines()[-1],';',len(truth),'identical Go/native/sanitizer bytes',flush=True)
assert run('source-node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',probe])==truth
emitted=S/'prepared_controls.mjs';emitted.write_bytes(run('emit-js',[stage0,'js',probe]));assert run('emitted-node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',emitted])==truth
for folder,old,new in [('jsx_fragments','if(!match)','if(match)'),('jsx_no_undef',"first>122","first>=97"),('no_adjacent_inline_elements','if(previous&&current)','if(previous&&!current)')]:
 original=OWN/folder/'index.a';text=original.read_text();assert old in text
 text=re.sub(r"(['\"])([^'\"]+\.(?:a|ts))\1",lambda m:q(str((original.parent/m.group(2)).resolve())),text)
 mutant=S/(folder+'-mutant.a');mutant.write_text(text.replace(old,new));entry=S/(folder+'-probe.a');entry.write_text(source.replace(str(original),str(mutant)));exe=S/(folder+'-mutant');run(folder+'-mutant-build',[stage0,'build',entry,'-o',exe]);assert run(folder+'-mutant',[exe])!=truth;assert (S/(folder+'-mutant.stderr')).read_bytes()==b'';print(folder,'semantic mutant exits zero and differs only on independent Go bytes',flush=True)
 cls={'jsx_fragments':'JsxFragments','jsx_no_undef':'JsxNoUndef','no_adjacent_inline_elements':'NoAdjacentInlineElements'}[folder];entry=S/(folder+'-refusal.a');entry.write_text(f'import {{{cls}}} from {q(str(original))};new {cls}().analyze();');exe=S/(folder+'-refusal');run(folder+'-refusal-build',[stage0,'build',entry,'-o',exe]);run(folder+'-refusal',[exe],70);assert b'NotYet:' in (S/(folder+'-refusal.stderr')).read_bytes()
# The U+0085 fixture distinguishes Go's space class from JavaScript's built-in \s.
original=OWN/'no_adjacent_inline_elements/index.a';text=original.read_text();assert r'\u0085' in text
text=re.sub(r"(['\"])([^'\"]+\.(?:a|ts))\1",lambda m:q(str((original.parent/m.group(2)).resolve())),text)
mutant=S/'edge-whitespace-mutant.a';mutant.write_text(text.replace(r'\u0085',''));entry=S/'edge-whitespace-probe.a';entry.write_text(source.replace(str(original),str(mutant)));exe=S/'edge-whitespace-mutant';run('edge-whitespace-mutant-build',[stage0,'build',entry,'-o',exe]);assert run('edge-whitespace-mutant',[exe])!=truth;assert (S/'edge-whitespace-mutant.stderr').read_bytes()==b''
print('edge whitespace regex mutant exits zero; Go bytes catch missing U+0085',flush=True)
print('PARTIAL PASS: prepared nodes only; source JSX/numeric/checker adapter missing',flush=True)

"""Source adapter validation using the isolated named shared JSX parser."""
import argparse,pathlib,json,subprocess,time
OWN=pathlib.Path(__file__).resolve().parent;ROOT=OWN.parents[3]
p=argparse.ArgumentParser();p.add_argument('--prepared',required=True);p.add_argument('--parser-root',required=True);args=p.parse_args();S=pathlib.Path(args.prepared);runs=[]
def run(name,cmd):
 t=time.monotonic()
 with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:r=subprocess.run([str(c) for c in cmd],cwd=ROOT,stdout=out,stderr=err)
 runs.append(dict(name=name,command=[str(c) for c in cmd],exit=r.returncode,seconds=time.monotonic()-t));(S/'source-runs.json').write_text(json.dumps(runs,indent=2)+'\n');assert r.returncode==0,(name,(S/(name+'.stderr')).read_text());return (S/(name+'.stdout')).read_bytes()
files=[S/f'control-{i}.tsx' for i in range(19,61)]
manifest=S/'source-controls.manifest';manifest.write_text(''.join(str(f)+'\n' for f in files));truth=run('source-go',[S/'oracle',S/'tsconfig.json',manifest])
q=json.dumps
source=f'import {{Parser}} from {q(str(pathlib.Path(args.parser_root)/"stage1/typescript/parser/parser.ts"))};import {{SourceAdjacent}} from {q(str(OWN/"no_adjacent_inline_elements/source.a"))};\n'
source+=f'import {{Diagnostic}} from {q(str(ROOT/"stage1/cohere/typeaware/diagnostic.ts"))};\n'
source+='let count=0;const rule=new SourceAdjacent();\n'
for f in files:
 source+=' {console.log('+q('file\t'+str(f))+');const text='+q(f.read_text())+';const parser=new Parser(text,"fixture.tsx");parser.file();const lines:string[]=[];for(const node of parser.nodes){'
 # Kind relevance dispatch is once in the driver, outside the rule.
 source+='let findings:Diagnostic[]=[];if(node.kind==="JsxElement"){findings=rule.jsx(node,parser.nodes,text);}else if(node.kind==="CallExpression"){findings=rule.call(node,parser.nodes,text);}for(const finding of findings){lines.push(finding.written());count++;}}lines.sort((left,right)=>left<right?-1:left>right?1:0);for(const line of lines){console.log(line);}}\n'
source+='console.log(`findings ${count}`);';entry=S/'source_controls.a';entry.write_text(source)
for name,sanitized in [('source-native',False),('source-asan',True)]:
 exe=S/name;run(name+'-build',[S/'adamic','build',entry,'-o',exe]+(['--sanitize'] if sanitized else []));assert run(name,[exe])==truth,name;assert (S/(name+'.stderr')).read_bytes()==b''
assert run('source-node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',entry])==truth
emitted=S/'source_controls.mjs';emitted.write_bytes(run('source-js',[S/'adamic','js',entry]));assert run('source-emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',emitted])==truth
# Mutate the actual adapter's byte span, while leaving native rule decisions intact.
original=OWN/'no_adjacent_inline_elements/source.a';text=original.read_text();import re
text=re.sub(r"(['\"])([^'\"]+\.(?:a|ts))\1",lambda m:q(str((original.parent/m.group(2)).resolve())),text)
assert 'result.start=utf8Length' in text;text=text.replace('result.start=utf8Length(source.slice(0,scanner.start));','result.start=utf8Length(source.slice(0,scanner.start))+1;')
mutant=S/'source-mutant.a';mutant.write_text(text);probe=S/'source_mutant_controls.a';probe.write_text(source.replace(str(original),str(mutant)));exe=S/'source-mutant';run('source-mutant-build',[S/'adamic','build',probe,'-o',exe]);assert run('source-mutant',[exe])!=truth;assert (S/'source-mutant.stderr').read_bytes()==b''
print('PASS source adapter:',len(files),'controls;',truth.decode().splitlines()[-1],';',len(truth),'identical Go/native/ASan/Node/JS bytes',flush=True)
print('Adapter span mutant exits zero and is caught only by independent Go bytes',flush=True)
# Compare actual source parsing over the original frozen corpora as well.
corpus_source=f'import {{Parser}} from {q(str(pathlib.Path(args.parser_root)/"stage1/typescript/parser/parser.ts"))};import {{SourceAdjacent}} from {q(str(OWN/"no_adjacent_inline_elements/source.a"))};import {{Diagnostic}} from {q(str(ROOT/"stage1/cohere/typeaware/diagnostic.ts"))};import {{programArguments,readTextFile,panic}} from "adamic";\n'
corpus_source+='const args=programArguments();const manifest=readTextFile(args[1]??" ");if(manifest.kind!=="Ok"){panic("missing manifest");}let count=0;const rule=new SourceAdjacent();for(const file of manifest.text.split("\\n")){if(file===""){continue;}const input=readTextFile(file);if(input.kind!=="Ok"){panic("missing source");}console.log("file\\t"+file);const parser=new Parser(input.text,file);parser.file();const lines:string[]=[];for(const node of parser.nodes){let findings:Diagnostic[]=[];if(node.kind==="JsxElement"){findings=rule.jsx(node,parser.nodes,input.text);}else if(node.kind==="CallExpression"){findings=rule.call(node,parser.nodes,input.text);}for(const finding of findings){lines.push(finding.written());count++;}}lines.sort((left,right)=>left<right?-1:left>right?1:0);for(const line of lines){console.log(line);}}console.log(`findings ${count}`);'
entry=S/'source_corpora.a';entry.write_text(corpus_source)
for name,sanitized in [('corpus-native',False),('corpus-asan',True)]:run(name+'-build',[S/'adamic','build',entry,'-o',S/name]+(['--sanitize'] if sanitized else []))
emitted=S/'source_corpora.mjs';emitted.write_bytes(run('corpus-js',[S/'adamic','js',entry]))
for corpus,base,config in [('compiler',pathlib.Path('/workspace/wave-06-typescript'),pathlib.Path('/workspace/wave-06-typescript/src/compiler/tsconfig.json')),('repository',ROOT,ROOT/'tsconfig.json')]:
 paths=(ROOT/f'stage1/cohere/typeaware/validation-coverage/{corpus}.manifest').read_text().splitlines();manifest=S/(corpus+'-source.manifest');manifest.write_text(''.join(str(base/path)+'\n' for path in paths if path))
 truth=run(corpus+'-source-go',[S/'oracle',config,manifest]);assert truth.endswith(b'findings 0\n')
 for name in ['corpus-native','corpus-asan']:
  assert run(corpus+'-'+name,[S/name,config,manifest])==truth,corpus+' '+name
  assert (S/(corpus+'-'+name+'.stderr')).read_bytes()==b''
 assert run(corpus+'-source-node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',entry,config,manifest])==truth
 assert run(corpus+'-source-emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',emitted,config,manifest])==truth
 print(corpus,len(paths),'source roots;',len(truth),'identical Go/native/ASan/Node/JS bytes',flush=True)

"""Actual source/checker comparison for jsx-no-undef using the existing raw bridge."""
import pathlib,argparse,json,subprocess,time
OWN=pathlib.Path(__file__).resolve().parent;ROOT=OWN.parents[3]
p=argparse.ArgumentParser();p.add_argument('--prepared',required=True);p.add_argument('--parser-root',required=True);args=p.parse_args();S=pathlib.Path(args.prepared);records=[]
def run(name,cmd,expect=0):
 t=time.monotonic()
 with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:r=subprocess.run([str(c) for c in cmd],cwd=ROOT,stdout=out,stderr=err)
 records.append(dict(name=name,command=[str(c) for c in cmd],exit=r.returncode,seconds=time.monotonic()-t));(S/'undef-source-runs.json').write_text(json.dumps(records,indent=2)+'\n');assert r.returncode==expect,(name,r.returncode,(S/(name+'.stderr')).read_text());return (S/(name+'.stdout')).read_bytes()
q=json.dumps;files=[S/f'control-{i}.tsx' for i in range(8,19)];manifest=S/'undef-source.manifest';manifest.write_text(''.join(str(f)+'\n' for f in files));truth=run('undef-source-go',[S/'oracle',S/'tsconfig.json',manifest]);archive=S/'checker.a';run('undef-archive',['go','build','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'])
source=f'import {{Parser}} from {q(str(pathlib.Path(args.parser_root)/"stage1/typescript/parser/parser.ts"))};import {{SourceNoUndef}} from {q(str(OWN/"jsx_no_undef/source.a"))};import {{Diagnostic}} from {q(str(ROOT/"stage1/cohere/typeaware/diagnostic.ts"))};import {{tsgoProgram,tsgoRelease}} from "adamic";\n'
source+='const program=tsgoProgram('+q(str(S/'tsconfig.json'))+','+q([str(f) for f in files])+');let count=0;const rule=new SourceNoUndef();\n'
for f in files:
 source+='{const file='+q(str(f))+';const text='+q(f.read_text())+';console.log("file\\t"+file);const parser=new Parser(text,file);parser.file();const lines:string[]=[];for(const node of parser.nodes){if(node.kind==="JsxOpeningElement"||node.kind==="JsxSelfClosingElement"){for(const finding of rule.visit(node,parser.nodes,text,file,program)){lines.push(finding.written());count++;}}}lines.sort((left,right)=>left<right?-1:left>right?1:0);for(const line of lines){console.log(line);}}\n'
source+='tsgoRelease(program);console.log(`findings ${count}`);';entry=S/'undef_source_controls.a';entry.write_text(source)
for name,sanitize in [('undef-source-native',False),('undef-source-asan',True)]:
 exe=S/name;run(name+'-build',[S/'adamic','build',entry,'-o',exe,'--tsgo',archive]+(['--sanitize'] if sanitize else []));assert run(name,[exe])==truth,name;assert (S/(name+'.stderr')).read_bytes()==b''
# Released checker queries must fail before producing a finding.
f=files[0];stale=S/'undef_source_released.a';stale.write_text(f'import {{Parser}} from {q(str(pathlib.Path(args.parser_root)/"stage1/typescript/parser/parser.ts"))};import {{SourceNoUndef}} from {q(str(OWN/"jsx_no_undef/source.a"))};import {{tsgoProgram,tsgoRelease}} from "adamic";const file={q(str(f))};const text={q(f.read_text())};const program=tsgoProgram({q(str(S/"tsconfig.json"))},[file]);const parser=new Parser(text,file);parser.file();tsgoRelease(program);for(const node of parser.nodes){{if(node.kind==="JsxSelfClosingElement"){{new SourceNoUndef().visit(node,parser.nodes,text,file,program);}}}}')
exe=S/'undef-source-released';run('undef-released-build',[S/'adamic','build',stale,'-o',exe,'--tsgo',archive]);assert run('undef-released',[exe],70)==b'';assert b'invalid or released checker handle' in (S/'undef-released.stderr').read_bytes()
mutant_archive=pathlib.Path('/workspace/wave-06-landing3-original/released-registry.a');exe=S/'undef-source-release-mutant';run('undef-release-mutant-build',[S/'adamic','build',stale,'-o',exe,'--tsgo',mutant_archive]);assert run('undef-release-mutant',[exe])==b'';assert (S/'undef-release-mutant.stderr').read_bytes()==b''
print('PASS jsx-no-undef actual source/checker:',len(files),'controls;',truth.decode().splitlines()[-1],';',len(truth),'identical Go/native/sanitized bytes',flush=True)
print('Released checker panic 70; retained-live registry mutant exits zero and loses required panic',flush=True)

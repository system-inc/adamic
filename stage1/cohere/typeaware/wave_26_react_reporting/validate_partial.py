# Reporting validator adapted from wave 06. This never proves source analysis.
"""Validate reporting and explicit refusal only; this is not a completed rule port."""
import pathlib,subprocess,json,os,shutil,re,time,argparse
ROOT=pathlib.Path(__file__).resolve().parents[4];OWN=pathlib.Path(__file__).resolve().parent
p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);args=p.parse_args();scratch=pathlib.Path(args.scratch).resolve();scratch.mkdir(parents=True,exist_ok=True);records=[]
def run(name,command,expect=0):
 start=time.monotonic()
 with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:r=subprocess.run([str(c) for c in command],cwd=ROOT,stdout=out,stderr=err)
 records.append(dict(name=name,command=[str(c) for c in command],exit=r.returncode,seconds=time.monotonic()-start));(scratch/'runs.json').write_text(json.dumps(records,indent=2)+'\n')
 assert r.returncode==expect,(name,r.returncode,(scratch/(name+'.stderr')).read_text())
 return (scratch/(name+'.stdout')).read_bytes()
stage0=scratch/'adamic';run('stage0-build',['go','build','-o',stage0,'./cmd/adamic'])
virtual=ROOT/'cohere/adamic_wave26reactreportoracle.go';overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/oracle.go'),str(ROOT/'cohere/internal/lint/rules/react/adamic_wave26_reporting.go'):str(OWN/'testdata/static_reporting.go')}}));oracle=scratch/'oracle';run('oracle-build',['go','-C',ROOT/'cohere','build','-overlay',overlay,'-o',oracle,virtual])
anchor=scratch/'anchor.d.ts';anchor.write_text('export {};');config=scratch/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'ESNext','jsx':'preserve'},'files':[str(anchor)]}))
prelude='type Dispatch<S>=(value:S)=>void;declare function useState(v:number):[number,Dispatch<number>];declare function useEffect(fn:()=>void):void;declare function useMemo(fn:()=>number):number;'
examples=[('effect',prelude+'function Component(){const[x,setX]=useState(0);useEffect(()=>{setX(1)});return x;}'),('render',prelude+'function Component(){const[x,setX]=useState(0);setX(1);return x;}'),('static','function Component(){const Inner = () => null;return <Inner/>;}'),('memo',prelude+'function Component(){const[x,setX]=useState(0);useMemo(()=>{setX(1);return 1;});return x;}')]
examples += [('static-unicode', "function Component(){const Inner = () => '世界🌍';return <Inner/>;}"), ('static-fallback', 'export {};')]
files=[]
for name,source in examples:
 f=scratch/(name+'.tsx');f.write_text(source+'\nexport {};\n');files.append(f)
manifest=scratch/'controls.manifest';manifest.write_text(''.join(str(f)+'\n' for f in files));truth=run('go-controls',[oracle,config,manifest]);rows=[];path=''
for line in truth.decode().splitlines():
 fields=line.split('\t')
 if fields[0]=='file':path=fields[1]
 elif len(fields)>=7:rows.append((path,int(fields[0]),int(fields[1]),fields[3]))
assert len(rows)==6,truth.decode()
imports="import {SetStateInEffect} from 'EFFECT';import {SetStateInRender} from 'RENDER';import {StaticComponents} from 'STATIC';"
for key,folder in [('EFFECT','set_state_in_effect'),('RENDER','set_state_in_render'),('STATIC','static_components')]:imports=imports.replace(key,str(OWN/folder/'index.a'))
source=imports+"import {written} from '"+str(ROOT/'stage1/typescript/parser/nodes.ts')+"';\n"
for f in files:
 source+='console.log('+json.dumps('file\t'+str(f))+');\n'
 created = '' if f.stem=='static-fallback' else "() => '世界🌍'" if f.stem=='static-unicode' else '() => null'
 for path,start,end,id in rows:
  if path!=str(f):continue
  call=(f'new SetStateInEffect().render({start},{end})' if id=='setStateInEffect' else f'new SetStateInRender().render({start},{end})' if id=='setStateInRender' else f'new SetStateInRender().memo({start},{end})' if id=='setStateInUseMemo' else f'new StaticComponents().render({json.dumps(created)},{start},{end})')
  source+='console.log('+call+'.written());\n'
source+='console.log("findings 6");\n';probe=scratch/'reporting_probe.a';probe.write_text(source);binary=scratch/'reporting';run('reporting-build',[stage0,'build',probe,'-o',binary]);output=run('reporting-run',[binary]);assert (scratch/'reporting-run.stderr').read_bytes()==b'';assert output==truth,(output.decode(),truth.decode());print('reporting only: five production Go findings plus one production-helper record match '+str(len(truth))+' bytes; ranges supplied by Go, no native source analysis',flush=True)
assert run('source-node-run',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',probe])==truth;assert (scratch/'source-node-run.stderr').read_bytes()==b''
javascript=run('javascript-emit',[stage0,'js',probe]);js=scratch/'reporting.js';js.write_bytes(javascript);assert run('javascript-run',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',js])==truth;assert (scratch/'javascript-run.stderr').read_bytes()==b''
asan=scratch/'reporting-asan';run('asan-build',[stage0,'build',probe,'-o',asan,'--sanitize']);assert run('asan-run',[asan])==truth;assert (scratch/'asan-run.stderr').read_bytes()==b''
for folder,cls,id in [('set_state_in_effect','SetStateInEffect','setStateInEffect'),('set_state_in_render','SetStateInRender','setStateInRender'),('static_components','StaticComponents','staticComponents')]:
 original=OWN/folder/'index.a';text=original.read_text();text=re.sub(r"(['\"])([^'\"]+\.ts)\1",lambda m:repr(str((original.parent/m.group(2)).resolve())),text)
 mutant=scratch/(folder+'-mutant.a');mutant.write_text(text.replace("'"+id+"'","'"+id+"Mutant'"));mutantProbe=scratch/(folder+'-report.a');mutantProbe.write_text(source.replace(str(original),str(mutant)));exe=scratch/(folder+'-mutant');run(folder+'-mutant-build',[stage0,'build',mutantProbe,'-o',exe]);assert run(folder+'-mutant-run',[exe])!=truth;assert (scratch/(folder+'-mutant-run.stderr')).read_bytes()==b''
 refusal=scratch/(folder+'-refusal.a');refusal.write_text(f'import {{{cls}}} from {str(original)!r};new {cls}().analyze();');exe=scratch/(folder+'-refusal');run(folder+'-refusal-build',[stage0,'build',refusal,'-o',exe]);run(folder+'-refusal-run',[exe],70);assert b'NotYet:' in (scratch/(folder+'-refusal-run.stderr')).read_bytes();assert run(folder+'-refusal-node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',refusal],70)==b'';assert (scratch/(folder+'-refusal-node.stderr')).read_bytes()==(scratch/(folder+'-refusal-run.stderr')).read_bytes()
 noRefusal=scratch/(folder+'-no-refusal.a');noRefusal.write_text(re.sub(r"panic\('NotYet:[^']+'\);",'return;',text));noRefusalProbe=scratch/(folder+'-no-refusal-probe.a');noRefusalProbe.write_text(f'import {{{cls}}} from {str(noRefusal)!r};new {cls}().analyze();');exe=scratch/(folder+'-no-refusal');run(folder+'-no-refusal-build',[stage0,'build',noRefusalProbe,'-o',exe]);assert run(folder+'-no-refusal-run',[exe])==b'';assert (scratch/(folder+'-no-refusal-run.stderr')).read_bytes()==b''
 print(folder+': rendering mutant caught by Go bytes; removed refusal caught by required exit 70',flush=True)
original=OWN/'static_components/index.a';text=original.read_text();text=text.replace('../../diagnostic.ts',str(ROOT/'stage1/cohere/typeaware/diagnostic.ts'))
assert text.count("created!==''")==1
mutant=scratch/'static-fallback-mutant.a';mutant.write_text(text.replace("created!==''",'true'))
mutantProbe=scratch/'static-fallback-mutant-probe.a';mutantProbe.write_text(source.replace(str(original),str(mutant)));exe=scratch/'static-fallback-mutant'
run('static-fallback-mutant-build',[stage0,'build',mutantProbe,'-o',exe]);assert run('static-fallback-mutant-run',[exe])!=truth;assert (scratch/'static-fallback-mutant-run.stderr').read_bytes()==b''
print('static-components empty-creator branch: diagnostic-only mutant caught by production message bytes',flush=True)
print('PARTIAL PASS: reporting, JavaScript, sanitizer and refusal only. All three source-analysis ports remain blocked.',flush=True)

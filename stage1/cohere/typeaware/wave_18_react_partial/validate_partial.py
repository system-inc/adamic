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
virtual=ROOT/'cohere/adamic_wave18reactoracle.go';overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/oracle.go')}}));oracle=scratch/'oracle';run('oracle-build',['go','-C',ROOT/'cohere','build','-overlay',overlay,'-o',oracle,virtual])
anchor=scratch/'anchor.d.ts';anchor.write_text('export {};');config=scratch/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'ESNext','jsx':'preserve'},'files':[str(anchor)]}))
prelude='type Dispatch<S>=(value:S)=>void;declare function useState(v:number):[number,Dispatch<number>];declare function useEffect(fn:()=>void):void;declare function useMemo(fn:()=>number):number;'
examples=[('effect',prelude+'function Component(){const[x,setX]=useState(0);useEffect(()=>{setX(1)});return x;}'),('render',prelude+'function Component(){const[x,setX]=useState(0);setX(1);return x;}'),('static','function Component(){const Inner = () => null;return <Inner/>;}'),('memo',prelude+'function Component(){const[x,setX]=useState(0);useMemo(()=>{setX(1);return 1;});return x;}')]
files=[]
for name,source in examples:
 f=scratch/(name+'.tsx');f.write_text(source+'\nexport {};\n');files.append(f)
manifest=scratch/'controls.manifest';manifest.write_text(''.join(str(f)+'\n' for f in files));truth=run('go-controls',[oracle,config,manifest]);rows=[];path=''
for line in truth.decode().splitlines():
 fields=line.split('\t')
 if fields[0]=='file':path=fields[1]
 elif len(fields)>=7:rows.append((path,int(fields[0]),int(fields[1]),fields[3]))
assert len(rows)==4,truth.decode()
imports="import {SetStateInEffect} from 'EFFECT';import {SetStateInRender} from 'RENDER';import {StaticComponents} from 'STATIC';"
for key,folder in [('EFFECT','set_state_in_effect'),('RENDER','set_state_in_render'),('STATIC','static_components')]:imports=imports.replace(key,str(OWN/folder/'index.a'))
source=imports+"import {written} from '"+str(ROOT/'stage1/typescript/parser/nodes.ts')+"';\n"
for f in files:
 source+='console.log('+json.dumps('file\t'+str(f))+');\n'
 for path,start,end,id in rows:
  if path!=str(f):continue
  call=(f'new SetStateInEffect().render({start},{end})' if id=='setStateInEffect' else f'new SetStateInRender().render({start},{end})' if id=='setStateInRender' else f'new SetStateInRender().memo({start},{end})' if id=='setStateInUseMemo' else f'new StaticComponents().render("() => null",{start},{end})')
  source+='console.log('+call+'.written());\n'
source+='console.log("findings 4");\n';probe=scratch/'reporting_probe.a';probe.write_text(source);binary=scratch/'reporting';run('reporting-build',[stage0,'build',probe,'-o',binary]);output=run('reporting-run',[binary]);assert output==truth,(output.decode(),truth.decode());print('reporting only: four production Go findings match '+str(len(truth))+' bytes; ranges supplied by Go, no native source analysis',flush=True)
asan=scratch/'reporting-asan';run('asan-build',[stage0,'build',probe,'-o',asan,'--sanitize']);assert run('asan-run',[asan])==truth;assert (scratch/'asan-run.stderr').read_bytes()==b''
for folder,cls,id in [('set_state_in_effect','SetStateInEffect','setStateInEffect'),('set_state_in_render','SetStateInRender','setStateInRender'),('static_components','StaticComponents','staticComponents')]:
 original=OWN/folder/'index.a';text=original.read_text();text=re.sub(r"(['\"])([^'\"]+\.ts)\1",lambda m:repr(str((original.parent/m.group(2)).resolve())),text)
 mutant=scratch/(folder+'-mutant.a');mutant.write_text(text.replace("'"+id+"'","'"+id+"Mutant'"));mutantProbe=scratch/(folder+'-report.a');mutantProbe.write_text(source.replace(str(original),str(mutant)));exe=scratch/(folder+'-mutant');run(folder+'-mutant-build',[stage0,'build',mutantProbe,'-o',exe]);assert run(folder+'-mutant-run',[exe])!=truth;assert (scratch/(folder+'-mutant-run.stderr')).read_bytes()==b''
 refusal=scratch/(folder+'-refusal.a');refusal.write_text(f'import {{{cls}}} from {str(original)!r};new {cls}().analyze();');exe=scratch/(folder+'-refusal');run(folder+'-refusal-build',[stage0,'build',refusal,'-o',exe]);run(folder+'-refusal-run',[exe],70);assert b'NotYet:' in (scratch/(folder+'-refusal-run.stderr')).read_bytes()
 noRefusal=scratch/(folder+'-no-refusal.a');removed,count=re.subn(r"panic\(\s*'NotYet:[^']+'\s*,?\s*\);",'return;',text);assert count==1,(folder,count);noRefusal.write_text(removed);noRefusalProbe=scratch/(folder+'-no-refusal-probe.a');noRefusalProbe.write_text(f'import {{{cls}}} from {str(noRefusal)!r};new {cls}().analyze();');exe=scratch/(folder+'-no-refusal');run(folder+'-no-refusal-build',[stage0,'build',noRefusalProbe,'-o',exe]);assert run(folder+'-no-refusal-run',[exe])==b'';assert (scratch/(folder+'-no-refusal-run.stderr')).read_bytes()==b''
 print(folder+': rendering mutant caught by Go bytes; removed refusal caught by required exit 70',flush=True)
parser=scratch/'parser-probe';run('parser-build',[stage0,'build',OWN/'parser_probe.a','-o',parser]);r=run('parser-run',[parser]);assert r==b'2\n';assert not (scratch/'parser-run.stderr').read_bytes()
parserMutant=scratch/'parser-count-mutant.a';text=(OWN/'parser_probe.a').read_text().replace('../../../typescript/parser/parser.ts',str(ROOT/'stage1/typescript/parser/parser.ts')).replace('.length}', '.length + 1}');assert text!=(OWN/'parser_probe.a').read_text();parserMutant.write_text(text);mutant=scratch/'parser-count-mutant';run('parser-count-mutant-build',[stage0,'build',parserMutant,'-o',mutant]);assert run('parser-count-mutant-run',[mutant])==b'3\n';assert not (scratch/'parser-count-mutant-run.stderr').read_bytes();print('base JSX parser: two nodes; compiling count mutant caught by expected output',flush=True)
print('PARTIAL PASS: reporting, sanitizer and refusal only. All three source-analysis ports remain blocked.',flush=True)

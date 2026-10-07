#!/usr/bin/env python3
"""Measure the native parser blocker against positive production Go controls."""
import pathlib,subprocess,json
root=pathlib.Path(__file__).resolve().parents[4];unit=pathlib.Path(__file__).resolve().parent
out=pathlib.Path('/workspace/wave-07-react-probe');out.mkdir(exist_ok=True);records=[]
def run(name,command,cwd=root):
 with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
  result=subprocess.run([str(value) for value in command],cwd=cwd,stdout=stdout,stderr=stderr)
 records.append(dict(name=name,command=[str(value) for value in command],exit=result.returncode))
 (out/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
 return result.returncode
stage0=pathlib.Path('/workspace/wave-07-next-rest/adamic')
if run('build-parser',[stage0,'build',unit/'parser_probe.a','-o',out/'parser']):raise RuntimeError('parser probe did not compile')
virtual=root/'cohere/adamic_wave07_react_probe.go';overlay=out/'oracle.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(unit/'testdata/oracle_react.go')}}))
if run('build-oracle',['go','build','-overlay',overlay,'-o',out/'oracle',virtual],root/'cohere'):raise RuntimeError('oracle did not compile')
prelude=out/'ambient.d.ts';prelude.write_text('declare namespace JSX {interface IntrinsicElements {div:unknown;}} type Dispatch<T>=(value:T)=>void;declare function useState<T>(x:T):[T,Dispatch<T>];declare function useEffect(callback:()=>void):void;declare function make():()=>unknown;')
config=out/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'target':'ES2022','strict':True,'jsx':'preserve','lib':['es2022']},'files':['ambient.d.ts']}))
controls=['function Widget(){const Inner=make();return <Inner/>;}','function Widget(){const [state,setState]=useState(0);setState(1);return <div/>;}','function Widget(){const [state,setState]=useState(0);useEffect(()=>{setState(1);});return <div/>;}']
paths=[]
for index,source in enumerate(controls):
 path=out/f'control-{index}.tsx';path.write_text(source+'\nexport {};\n');paths.append(path)
manifest=out/'controls.manifest';manifest.write_text(''.join(str(path)+'\n' for path in paths))
if run('oracle',[out/'oracle',config,manifest]):raise RuntimeError('production oracle failed')
for index,path in enumerate(paths):
 code=run('parser-'+str(index),[out/'parser',path])
 if code!=70 or b'expected GreaterThanToken, got SlashToken' not in (out/('parser-'+str(index)+'.stderr')).read_bytes():raise RuntimeError('recorded JSX blocker changed')
 print(f'parser control {index}: exit {code}',flush=True)
truth=(out/'oracle.stdout').read_text()
for name in ['react-hooks/set-state-in-effect','react-hooks/set-state-in-render','react-hooks/static-components']:
 if '\t'+name+'\t' not in truth:raise RuntimeError('missing production positive control '+name)
plain=out/'plain.tsx';plain.write_text('function Widget(){return 1;}\nexport {};\n')
if run('parser-plain',[out/'parser',plain]):raise RuntimeError('native parser also failed ordinary TypeScript')
print('ordinary TypeScript: native parser exits 0',flush=True)
print(truth,flush=True)

"""Go-prepared React HIR core validation, not native source-analysis parity."""
import argparse,pathlib,subprocess,json,os,time,re,shutil
ROOT=pathlib.Path(__file__).resolve().parents[4];OWN=pathlib.Path(__file__).resolve().parent
p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);p.add_argument('--compiler',required=True);args=p.parse_args();scratch=pathlib.Path(args.scratch).resolve();scratch.mkdir(parents=True,exist_ok=True);records=[]
def run(name,cmd,expect=0):
 start=time.monotonic()
 with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:r=subprocess.run([str(x) for x in cmd],cwd=ROOT,stdout=out,stderr=err)
 records.append(dict(name=name,command=[str(x) for x in cmd],exit=r.returncode,seconds=time.monotonic()-start));(scratch/'runs.json').write_text(json.dumps(records,indent=2)+'\n')
 assert r.returncode==expect,(name,r.returncode,(scratch/(name+'.stderr')).read_text())
 return (scratch/(name+'.stdout')).read_bytes()
stage0=scratch/'adamic';run('stage0-build',['go','build','-o',stage0,'./cmd/adamic']);virtual=ROOT/'cohere/adamic_wave06cores.go';overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/cores_oracle.go')}}));oracle=scratch/'oracle';run('oracle-build',['go','-C',ROOT/'cohere','build','-overlay',overlay,'-o',oracle,virtual])
declarations='''export type SetStateAction<S> = S | ((prev:S)=>S);
export type Dispatch<A> = (value:A)=>void;
export interface RefObject<T> {current:T}
export declare function useState<S>(initial?:S):[S,Dispatch<SetStateAction<S>>];
export declare function useEffect(callback:()=>void,deps?:unknown[]):void;
export declare function useLayoutEffect(callback:()=>void,deps?:unknown[]):void;
export declare function useInsertionEffect(callback:()=>void,deps?:unknown[]):void;
export declare function useEffectEvent<T extends Function>(callback:T):T;
export declare function useRef<T>(initial:T):RefObject<T>;
export declare function useCallback<T>(callback:T,deps:unknown[]):T;
export declare function useMemo<T>(callback:()=>T,deps:unknown[]):T;
export type ActionDispatch<A extends unknown[]>=(...args:A)=>void;
export declare function useReducer<S,A>(r:(s:S,a:A)=>S,i:S):[S,ActionDispatch<[A]>];
'''
(scratch/'react.d.ts').write_text(declarations);(scratch/'ambient.d.ts').write_text("declare module 'react' {\n"+declarations.replace('export declare ','export ')+'}\n');config=scratch/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'NodeNext','moduleDetection':'force','lib':['ES2022'],'jsx':'preserve','noEmit':True},'include':['*.d.ts']}))
paths=[]
for i,row in enumerate(json.loads((OWN/'testdata/core_controls.json').read_text())):
 file=scratch/('fixture-'+str(i)+'.tsx');file.write_text(row['source']);paths.append(file)
manifest=scratch/'all.manifest';manifest.write_text(''.join(str(x)+'\n' for x in paths));valid=run('source-filter',[oracle,config,manifest,'--valid-sources']).decode().splitlines();assert len(valid)==91,len(valid);print('91/100 valid upstream and independent source controls, with no admitted input omitted',flush=True)
batches=[];finding_count=0;total_bytes=0
for start in range(0,len(valid),8):
 name='batch-'+str(start//8);manifest=scratch/(name+'.manifest');manifest.write_text('\n'.join(valid[start:start+8])+'\n');generated=run(name+'-prepare',[oracle,config,manifest,'--native-graphs',OWN]);entry=scratch/(name+'.a');entry.write_bytes(generated);truth=run(name+'-go',[oracle,config,manifest]);binary=scratch/name
 run(name+'-build',[stage0,'build',entry,'-o',binary]);assert run(name+'-native',[binary])==truth;assert (scratch/(name+'-native.stderr')).read_bytes()==b''
 asan=scratch/(name+'-asan');run(name+'-asan-build',[stage0,'build',entry,'-o',asan,'--sanitize']);assert run(name+'-asan-native',[asan])==truth;assert (scratch/(name+'-asan-native.stderr')).read_bytes()==b''
 count=int(truth.decode().splitlines()[-1].split()[1]);finding_count+=count;total_bytes+=len(truth);batches.append(dict(name=name,entry=str(entry),truth=str(scratch/(name+'-go.stdout'))));print(name,len(truth),'identical bytes',count,'findings, normal and ASan/UBSan',flush=True)
assert finding_count==48,finding_count
(scratch/'batches.json').write_text(json.dumps(batches,indent=2)+'\n')
for name,file,old,new,marker in [('effect-setter','set_state_in_effect.a','if(!local.has(i.p0.id)&&!function_.setter(i.p0.id))','if(!local.has(i.p0.id)&&function_.setter(i.p0.id))','setStateInEffect'),('render-unconditional','set_state_in_render.a','if(!unconditional.has(block.id))','if(unconditional.has(block.id))','setStateInRender'),('static-creator','static_components.a','dynamic.set(target,i.p0.id);','dynamic.delete(target);','staticComponents'),('adjacency-alias','postdominator.a','const fresh:number[]=[];','const fresh:number[]=this.empty;','')]:
 selected=next(b for b in batches if ((marker.encode() in pathlib.Path(b['truth']).read_bytes()) if marker else 'for(const x of props)' in '\n'.join(pathlib.Path(x).read_text() for x in (scratch/(b['name']+'.manifest')).read_text().splitlines())))
 directory=scratch/(name+'-source');shutil.copytree(OWN,directory,ignore=shutil.ignore_patterns('validation','landing_evidence','core_evidence'),dirs_exist_ok=True)
 for copied in directory.rglob('*.a'):
  original=OWN/copied.relative_to(directory);text=copied.read_text();text=re.sub(r"(['\"])([^'\"]+\.(?:ts|a))\1",lambda m:repr(m.group(2) if m.group(2).endswith('.a') and (original.parent/m.group(2)).resolve().is_relative_to(OWN) else str((original.parent/m.group(2)).resolve())),text)
  if copied==directory/'cores'/file:assert text.count(old)==1;text=text.replace(old,new)
  copied.write_text(text)
 entry=scratch/(name+'-main.a');entry.write_text(pathlib.Path(selected['entry']).read_text().replace(str(OWN),str(directory)));binary=scratch/name;run(name+'-mutant-build',[stage0,'build',entry,'-o',binary]);assert run(name+'-mutant-run',[binary])!=pathlib.Path(selected['truth']).read_bytes();assert (scratch/(name+'-mutant-run.stderr')).read_bytes()==b'';print(name,'semantic mutant exits 0, independent Go bytes catch it',flush=True)
compiler=pathlib.Path(args.compiler).resolve();assert subprocess.check_output(['git','-C',compiler,'rev-parse','HEAD'],text=True).strip()=='050880ce59e30b356b686bd3144efe24f875ebc8'
for name,base,configuration in [('compiler',compiler,compiler/'src/compiler/tsconfig.json'),('repository',ROOT,ROOT/'tsconfig.json')]:
 manifest=scratch/(name+'.manifest');manifest.write_text(''.join(str(base/x)+'\n' for x in (OWN.parent/'validation-coverage'/(name+'.manifest')).read_text().splitlines() if x));entry=scratch/(name+'.a');entry.write_bytes(run(name+'-prepare',[oracle,configuration,manifest,'--native-graphs',OWN]));truth=run(name+'-go',[oracle,configuration,manifest]);binary=scratch/name;run(name+'-build',[stage0,'build',entry,'-o',binary]);assert run(name+'-native',[binary])==truth
 asan=scratch/(name+'-asan');run(name+'-asan-build',[stage0,'build',entry,'-o',asan,'--sanitize']);assert run(name+'-asan-native',[asan])==truth;assert (scratch/(name+'-asan-native.stderr')).read_bytes()==b'';print(name,len(truth),'identical prepared-input bytes; native source lowering remains untested',flush=True)
print('PARTIAL CORE PASS:',finding_count,'findings',total_bytes,'bytes; prepared HIR only, source adapter absent',flush=True)

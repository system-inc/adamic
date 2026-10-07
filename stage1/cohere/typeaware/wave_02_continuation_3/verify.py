#!/usr/bin/env python3
"""Standalone byte oracle; edits no shared harness, registrations, or compiler sources."""
import argparse,gzip,hashlib,json,os,statistics,subprocess,time
from pathlib import Path
p=argparse.ArgumentParser()
p.add_argument('--scratch',required=True);p.add_argument('--stage0',required=True);p.add_argument('--archive',required=True);p.add_argument('--sanitized-archive',required=True);p.add_argument('--typescript-source',required=True)
a=p.parse_args();own=Path(__file__).resolve().parent;root=own.parents[3];scratch=Path(a.scratch).resolve();scratch.mkdir(parents=True,exist_ok=True)
commands=[]
def run(name,args,cwd=root,expected=0):
 start=time.perf_counter_ns()
 with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:
  result=subprocess.run([str(x) for x in args],cwd=cwd,stdout=out,stderr=err)
 elapsed=time.perf_counter_ns()-start;commands.append({'name':name,'args':[str(x) for x in args],'cwd':str(cwd),'exit':result.returncode,'process_ns':elapsed})
 (scratch/'commands.json').write_text(json.dumps(commands,indent=2))
 if result.returncode!=expected:raise RuntimeError(f'{name}: exit {result.returncode}; see saved streams')
 return (scratch/(name+'.stdout')).read_bytes(),elapsed

def compare(name,config,manifest):
 truth,_=run(name+'-go',[scratch/'oracle',config,manifest]);actual,_=run(name+'-native',[scratch/'native',config,manifest]);sanitized,_=run(name+'-asan',[scratch/'native-asan',config,manifest])
 assert truth==actual==sanitized,name+' diagnostic bytes differ'
 assert not (scratch/(name+'-native.stderr')).read_bytes() and not (scratch/(name+'-asan.stderr')).read_bytes(),name+' native stderr'
 print(name,len(truth),'identical bytes',truth.splitlines()[-1].decode(),flush=True)
 return truth

virtual=root/'cohere/adamic_wave02_continuation3_oracle.go';overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'testdata/oracle.go')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',scratch/'oracle',virtual],root/'cohere')
for name,archive,flags in [('native',a.archive,[]),('native-asan',a.sanitized_archive,['--sanitize'])]:
 run(name+'-build',[a.stage0,'build',own/'suite.a','-o',scratch/name,'--tsgo',archive,*flags])
fixtures=json.loads((own/'testdata/fixtures.json').read_text());paths=[]
for index,source in enumerate(fixtures):
 file=scratch/f'control-{index:03}.tsx';file.write_text(source+'\nexport {};\n');paths.append(str(file))
config=scratch/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','module':'NodeNext','jsx':'preserve','lib':['ESNext'],'noEmit':True},'include':['*.ts','*.tsx']}))
(scratch/'globals.d.ts').write_text('declare function useEffect(...args:any[]):any;declare function useLayoutEffect(...args:any[]):any;declare function useMemo(...args:any[]):any;declare function useCallback(...args:any[]):any;declare function useImperativeHandle(...args:any[]):any;declare function useState(value:any):[any,(value:any)=>void];declare function useRef(...args:any[]):any;declare function useEffectEvent(...args:any[]):any;declare const React:any;declare function log(...args:any[]):void;declare const ref:any,deps:any,flag:any,other:any,window:any,a:any;declare function wrap(...args:any[]):any;')
manifest=scratch/'controls.manifest';manifest.write_text('\n'.join(paths)+'\n')
valid,_=run('valid-sources',[scratch/'oracle',config,manifest,'--valid-sources']);valid_paths=valid.decode().splitlines();supported=[];blocked=[]
for index,file in enumerate(valid_paths):
 single=scratch/'single.manifest';single.write_text(file+'\n')
 start=time.perf_counter_ns()
 with (scratch/f'parse-{index:03}.stdout').open('wb') as out,(scratch/f'parse-{index:03}.stderr').open('wb') as err:
  result=subprocess.run([str(scratch/'native'),str(config),str(single)],cwd=root,stdout=out,stderr=err)
 if result.returncode:
  message=(scratch/f'parse-{index:03}.stderr').read_text()
  assert result.returncode==70 and ('parser slice expected' in message or 'shared parser cannot represent JSX' in message),message
  blocked.append({'file':file,'exit':result.returncode,'stderr':message})
 else:supported.append(file)
(scratch/'blocked.json').write_text(json.dumps(blocked,indent=2));manifest=scratch/'supported.manifest';manifest.write_text('\n'.join(supported)+'\n')
print('controls',len(fixtures),'Go parse-valid',len(valid_paths),'native supported',len(supported),'blocked',len(blocked),flush=True)
truth=compare('controls',config,manifest)
for name in ['prefer-rest-params','prefer-regex-literals','react-hooks/exhaustive-deps']:assert ('\t'+name+'\t').encode() in truth,name+' has no positive control'
compiler=Path(a.typescript_source).resolve();compiler_paths=sorted((compiler/'src/compiler').rglob('*.ts'));compiler_manifest=scratch/'compiler.manifest';compiler_manifest.write_text('\n'.join(str(file) for file in compiler_paths)+'\n');assert len(compiler_paths)==77,len(compiler_paths)
repository_manifest=root/'stage1/cohere/typeaware/validation-volume/repository.manifest';assert len(repository_manifest.read_text().splitlines())==287
compare('compiler',compiler/'src/compiler/tsconfig.json',compiler_manifest)
compare('repository',root/'tsconfig.json',repository_manifest)
mutants=[('rest','prefer_rest_params.a','symbol.declarations.length !== 0','symbol.declarations.length === 0'),('regex','prefer_regex_literals.a','syntax.valid()','!syntax.valid()'),('react','exhaustive_deps.a','!effect && recommendation.missing.length > 0','effect && recommendation.missing.length > 0')]
for name,filename,before,after in mutants:
 directory=scratch/(name+'-source');directory.mkdir(exist_ok=True)
 for file in own.glob('*.a'):
  source=file.read_text()
  if file.name==filename:
   assert source.count(before)==1,(filename,before);source=source.replace(before,after,1)
  source=source.replace("from '../", "from '"+str(own.parent)+'/').replace("from '../../../typescript/", "from '"+str(root/'stage1/typescript')+'/')
  (directory/file.name).write_text(source)
 run(name+'-mutant-build',[a.stage0,'build',directory/'suite.a','-o',scratch/(name+'-mutant'),'--tsgo',a.archive])
 output,_=run(name+'-mutant-run',[scratch/(name+'-mutant'),config,manifest]);assert output!=truth and not (scratch/(name+'-mutant-run.stderr')).read_bytes(),name+' mutant survived'
 first=next((i for i,(x,y) in enumerate(zip(output,truth)) if x!=y),min(len(output),len(truth)))
 print(name,'mutant exits 0, byte comparison catches byte',first,flush=True)
# Prove the owned fail-closed JSX guard distinguishes a silent wrong AST.
directory=scratch/'jsx-source';directory.mkdir(exist_ok=True)
for file in own.glob('*.a'):
 source=file.read_text()
 if file.name=='suite.a':
  assert source.count("path.endsWith('.tsx')")==1
  source=source.replace("path.endsWith('.tsx')", "false && path.endsWith('.tsx')",1)
 source=source.replace("from '../", "from '"+str(own.parent)+'/').replace("from '../../../typescript/", "from '"+str(root/'stage1/typescript')+'/')
 (directory/file.name).write_text(source)
run('jsx-mutant-build',[a.stage0,'build',directory/'suite.a','-o',scratch/'jsx-mutant','--tsgo',a.archive])
jsx_file=scratch/'control-467.tsx';jsx_manifest=scratch/'jsx.manifest';jsx_manifest.write_text(str(jsx_file)+'\n')
run('jsx-refusal',[scratch/'native',config,jsx_manifest],expected=70)
jsx_truth,_=run('jsx-go',[scratch/'oracle',config,jsx_manifest]);jsx_wrong,_=run('jsx-mutant-run',[scratch/'jsx-mutant',config,jsx_manifest])
assert jsx_truth!=jsx_wrong and not (scratch/'jsx-mutant-run.stderr').read_bytes()
print('JSX refusal mutant exits 0; independent Go construction finding catches the silent AST error',flush=True)
probe=scratch/'probe.ts';probe.write_text('x;\n')
run('released-build',[a.stage0,'build',own/'released.a','-o',scratch/'released','--tsgo',a.archive])
run('released-run',[scratch/'released',config,probe],expected=70)
assert (scratch/'released-run.stderr').read_text()=='adamic: panic: invalid or released checker handle\n'
print('released handle: panic 70',flush=True)
original=root/'bridge/tsgo/archive/main.go';source=original.read_text();before='delete(programs.live, uint64(handle))';assert source.count(before)==1
side=scratch/'released-main.go';side.write_text(source.replace(before,'// Mutant retains the released program.',1));overlay=scratch/'released-overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(side)}}))
run('released-registry-build',['go','build','-buildmode=c-archive','-overlay',overlay,'-o',scratch/'released-registry.a','./bridge/tsgo/archive'])
run('released-mutant-build',[a.stage0,'build',own/'released.a','-o',scratch/'released-mutant','--tsgo',scratch/'released-registry.a'])
run('released-mutant-run',[scratch/'released-mutant',config,probe])
assert not (scratch/'released-mutant-run.stderr').read_bytes()
print('released registry mutant exits 0; expected panic 70 catches it',flush=True)

timings=[]
for corpus,cfg,mf in [('compiler',compiler/'src/compiler/tsconfig.json',compiler_manifest),('repository',root/'tsconfig.json',repository_manifest)]:
 for index in range(3):
  for implementation in (['native','go'] if index%2 else ['go','native']):
   binary=scratch/('oracle' if implementation=='go' else 'native')
   output,elapsed=run(f'{corpus}-{index+1}-{implementation}-timing',[binary,cfg,mf,'--count']);assert output.splitlines()[-1]==b'findings 0'
   timings.append({'corpus':corpus,'round':index+1,'implementation':implementation,'process_ns':elapsed})
 for implementation in ['native','go']:
  print(corpus,implementation,'median_seconds',statistics.median(x['process_ns'] for x in timings if x['corpus']==corpus and x['implementation']==implementation)/1e9,flush=True)
(scratch/'timings.json').write_text(json.dumps(timings,indent=2));print('PASS: complete bytes, three compiling mutants, sanitizers, released handle; JSX exclusions remain',flush=True)

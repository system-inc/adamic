#!/usr/bin/env python3
"""Isolated wave 27 third-batch validation. Does not edit the shared harness."""
from pathlib import Path
import os,sys,subprocess,json,time,shutil,hashlib,statistics
repo=Path(__file__).resolve().parents[4]
source=Path(__file__).resolve().parent
art=Path(os.environ['ADAMIC_WAVE27_THIRD_ARTIFACTS']);art.mkdir(parents=True,exist_ok=True)
stage0=art/'adamic';archive=art/'checker.a';native=art/'native';oracle=art/'oracle'
serial=0

def run(label,args,cwd=repo,env=None,expected=0):
 global serial
 serial+=1;stem=art/f'{serial:03}-{label}'
 with stem.with_suffix('.stdout').open('wb') as stdout,stem.with_suffix('.stderr').open('wb') as stderr:
  start=time.perf_counter_ns();result=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=stdout,stderr=stderr);elapsed=time.perf_counter_ns()-start
 out=stem.with_suffix('.stdout').read_bytes();err=stem.with_suffix('.stderr').read_bytes()
 if result.returncode!=expected:raise AssertionError((label,result.returncode,expected,err[-3000:]))
 return out,err,elapsed

def build(label,entry,output,lib=archive,sanitize=False):
 args=[stage0,'build',entry,'-o',output,'--tsgo',lib]
 if sanitize:args+=['--sanitize']
 run(label,args)

def lib(label,path,sanitize=False,overlay=None):
 args=['go','build','-buildmode=c-archive','-o',path]
 if sanitize:args+=['-asan']
 if overlay:args+=['-overlay',overlay]
 args+=['./bridge/tsgo/archive'];run(label,args)

def compare(label,binary,config,manifest):
 truth,_,_=run(label+'-go',[oracle,config,manifest]);got,err,_=run(label+'-native',[binary,config,manifest]);assert not err,(label,err)
 assert truth==got,(label,next((i for i,(x,y) in enumerate(zip(truth,got)) if x!=y),min(len(truth),len(got))))
 print(label,len(got),'identical bytes',got.rsplit(b'\n',2)[-2].decode(),flush=True);return truth

run('stage0',['go','build','-o',stage0,'./cmd/adamic'])
lib('checker',archive)
build('native',source/'suite.a',native)
overlay=art/'oracle-overlay.json';virtual=repo/'cohere/adamic_wave27_third_oracle.go';overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'oracle.go.txt')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],cwd=repo/'cohere')
run('make-controls',[sys.executable,source/'make_controls.py',art])
config=art/'tsconfig.json';manifest=art/'controls.manifest'
candidates=art/'control-candidates.manifest';shutil.copyfile(manifest,candidates)
valid,err,_=run('control-source-parsing',[oracle,config,candidates,'--valid-sources']);assert not err
manifest.write_bytes(valid)
excluded=sorted(set(candidates.read_text().splitlines())-set(valid.decode().splitlines()))
(art/'excluded-controls.json').write_text(json.dumps(excluded,indent=2)+'\n')
print(len(candidates.read_text().splitlines()),'control candidates;',len(valid.decode().splitlines()),'parseable;',len(excluded),'excluded by independent Go parser',flush=True)
truth=compare('controls',native,config,manifest)
for name in ['prefer-rest-params','prefer-regex-literals','react-hooks/exhaustive-deps']: assert ('\t'+name+'\t').encode() in truth
asanarchive=art/'checker-asan.a';asan=art/'native-asan';lib('checker-asan',asanarchive,True);build('native-asan',source/'suite.a',asan,asanarchive,True);compare('controls-asan',asan,config,manifest)
populations=[]
for name,setting,corpus_config in [('repository','ADAMIC_WAVE27_REPOSITORY_MANIFEST',repo/'tsconfig.json'),('compiler','ADAMIC_WAVE27_COMPILER_MANIFEST',Path(os.environ.get('ADAMIC_TYPESCRIPT_SOURCE','/missing'))/'src/compiler/tsconfig.json')]:
 if setting in os.environ:
  corpus_manifest=Path(os.environ[setting]);compare(name,native,corpus_config,corpus_manifest);compare(name+'-asan',asan,corpus_config,corpus_manifest);populations.append((name,corpus_config,corpus_manifest))
 else:print(name,'corpus not supplied',flush=True)

mutants=[('rest','prefer_rest_params.a','symbol.declarations.length !== 0','symbol.declarations.length === 0'),('hooks','exhaustive_deps.a','result.missing.length > 0 ?', 'result.missing.length === 0 ?'),('regex','prefer_regex_literals.a',"if(args.length !== 1 && args.length !== 2)","if(args.length === 1 || args.length === 2)")]
for name,file,before,after in mutants:
 directory=art/(name+'-source');directory.mkdir(exist_ok=True)
 for original in source.glob('*.a'):
  text=original.read_text()
  if original.name==file:
   import re
   pattern=r'\s*'.join(re.escape(char) for char in before)
   text,count=re.subn(pattern,after,text);assert count==1,(name,count)
  import re
  def dependency(match):return "'"+str((source/match.group(1)).resolve())+"'"
  text=re.sub(r"'((?:\.\./)+[^']+)'",dependency,text)
  (directory/original.name).write_text(text)
 binary=art/(name+'-mutant');build(name+'-mutant',directory/'suite.a',binary)
 got,err,_=run(name+'-run',[binary,config,manifest]);assert not err and got!=truth
 offset=next((i for i,(x,y) in enumerate(zip(truth,got)) if x!=y),min(len(truth),len(got)));print(name,'mutant: exit 0, empty stderr; Go byte oracle catches byte',offset,flush=True)

question_mutants=[('wave-27-declaration-ancestry','wave_27_declaration_ancestry.go','',''),('binding-declarations','','','')]
probe=art/'probe.a';probe.write_text('work();\n');release=art/'released.a';release.write_text("""import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const mode=args[2]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
const root=mode==='wave-27-syntax-flow'||mode==='wave-27-module-sources';const call=mode==='wave-27-call-declaration';console.log(tsgoInspect(program,file,0,root?8:call?6:4,root?'SourceFile':call?'CallExpression':'Identifier',mode));
""")
stale=art/'released';build('released-build',release,stale)
for mode,_,_,_ in question_mutants:
 _,err,_=run('released-'+mode,[stale,config,probe,mode],expected=70);assert err==b'adamic: panic: invalid or released checker handle\n';print(mode,'released handle: required panic 70',flush=True)
original=repo/'bridge/tsgo/archive/main.go';text=original.read_text();before='delete(programs.live, uint64(handle))';assert text.count(before)==1;mutant=art/'released-registry.go';mutant.write_text(text.replace(before,'// Mutant retains the released handle.'));overlay=art/'released-registry.json';overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}));mutantarchive=art/'released-registry.a';lib('released-registry',mutantarchive,overlay=overlay);binary=art/'released-registry';build('released-registry-build',release,binary,mutantarchive)
for mode,_,_,_ in question_mutants:
 run('released-mutant-'+mode,[binary,config,probe,mode]);print(mode,'retained-registry mutant: exit 0; required-panic check catches it',flush=True)

measurements={}
for name,corpus_config,corpus_manifest in populations:
 samples={'native':[],'go':[]};streams=[]
 for round in range(3):
  for engine in (['native','go'] if round%2==0 else ['go','native']):
   env=dict(os.environ,ADAMIC_TSGO_TIMING='1');out,err,elapsed=run(f'{name}-timed-{round}-{engine}',[native if engine=='native' else oracle,corpus_config,corpus_manifest],env=env);samples[engine].append(elapsed);streams.append(out)
 assert all(stream==streams[0] for stream in streams)
 medians={engine:statistics.median(values) for engine,values in samples.items()};measurements[name]={'process_ns':samples,'median_ns':medians,'native_over_go':medians['native']/medians['go'],'bytes':len(streams[0]),'sha256':hashlib.sha256(streams[0]).hexdigest()}
(art/'measurements.json').write_text(json.dumps(measurements,indent=2)+'\n');print(json.dumps(measurements,indent=2),flush=True);print('PASS',flush=True)

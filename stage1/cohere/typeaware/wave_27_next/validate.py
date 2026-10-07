#!/usr/bin/env python3
"""Isolated wave 27 second-batch validation. Does not edit the shared harness."""
from pathlib import Path
import os,sys,subprocess,json,time,shutil,hashlib,statistics
repo=Path(__file__).resolve().parents[4]
source=Path(__file__).resolve().parent
art=Path(os.environ['ADAMIC_WAVE27_NEXT_ARTIFACTS']);art.mkdir(parents=True,exist_ok=True)
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
overlay=art/'oracle-overlay.json';virtual=repo/'cohere/adamic_wave27_next_oracle.go';overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'oracle.go.txt')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],cwd=repo/'cohere')
run('make-controls',[sys.executable,source/'make_controls.py',art])
config=art/'tsconfig.json';manifest=art/'controls.manifest';truth=compare('controls',native,config,manifest)
for name in ['correctness-no-process-exit-after-output','correctness-no-uncleared-race-timeout','correctness-require-blocking-standard-streams']:assert ('\tnexus/'+name+'\t').encode() in truth
asanarchive=art/'checker-asan.a';asan=art/'native-asan';lib('checker-asan',asanarchive,True);build('native-asan',source/'suite.a',asan,asanarchive,True);compare('controls-asan',asan,config,manifest)
populations=[]
for name,setting,corpus_config in [('repository','ADAMIC_WAVE27_REPOSITORY_MANIFEST',repo/'tsconfig.json'),('compiler','ADAMIC_WAVE27_COMPILER_MANIFEST',Path(os.environ.get('ADAMIC_TYPESCRIPT_SOURCE','/missing'))/'src/compiler/tsconfig.json')]:
 if setting in os.environ:
  corpus_manifest=Path(os.environ[setting]);compare(name,native,corpus_config,corpus_manifest);compare(name+'-asan',asan,corpus_config,corpus_manifest);populations.append((name,corpus_config,corpus_manifest))
 else:print(name,'corpus not supplied',flush=True)

mutants=[('process','correctness_no_process_exit_after_output.a','if(state.length>0){reported.add(index);}','if(state.length===0){reported.add(index);}'),('race','correctness_no_uncleared_race_timeout.a','this.timer(index)&&this.lost(executor,index)','this.timer(index)&&!this.lost(executor,index)'),('blocking','correctness_require_blocking_standard_streams.a','if(found.length===0){return;}','if(found.length>=0){return;}')]
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

question_mutants=[('wave-27-declaration-ancestry','wave_27_declaration_ancestry.go','out.yes(source.IsDeclarationFile)','out.yes(false)'),('wave-27-syntax-flow','wave_27_syntax_flow.go','out.yes(block.Reachable)','out.yes(false)'),('wave-27-call-declaration','wave_27_call_declaration.go','flags = uint64(t.Flags())','flags = 131072'),('wave-27-module-sources','wave_27_module_sources.go','out.yes(edge.typeOnly)','out.yes(true)')]
question_mutants.append(('ancestry-name-kind','wave_27_declaration_ancestry.go','out.text(nameKind)','out.text("Identifier" + nameKind[:0])'))
for name,file,before,after in question_mutants:
 original=repo/'bridge/tsgo/checker'/file;text=original.read_text();assert text.count(before)==1,(name,before);mutant=art/(name+'-mutant.go');mutant.write_text(text.replace(before,after));overlay=art/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}));mutantarchive=art/(name+'-mutant.a');lib(name+'-archive',mutantarchive,overlay=overlay);binary=art/(name+'-mutant');build(name+'-build',source/'suite.a',binary,mutantarchive);got,err,_=run(name+'-run',[binary,config,manifest]);assert not err and got!=truth,(name,err);offset=next((i for i,(x,y) in enumerate(zip(truth,got)) if x!=y),min(len(truth),len(got)));print(name,'fact mutant: exit 0, empty stderr; Go bytes catch byte',offset,flush=True)
 binary.unlink();mutantarchive.unlink()

probe=art/'probe.a';probe.write_text('work();\n');release=art/'released.a';release.write_text("""import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const mode=args[2]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
const root=mode==='wave-27-syntax-flow'||mode==='wave-27-module-sources';const call=mode==='wave-27-call-declaration';console.log(tsgoInspect(program,file,0,root?8:call?6:4,root?'SourceFile':call?'CallExpression':'Identifier',mode));
""")
stale=art/'released';build('released-build',release,stale)
for mode,_,_,_ in question_mutants[:4]:
 _,err,_=run('released-'+mode,[stale,config,probe,mode],expected=70);assert err==b'adamic: panic: invalid or released checker handle\n';print(mode,'released handle: required panic 70',flush=True)
original=repo/'bridge/tsgo/archive/main.go';text=original.read_text();before='delete(programs.live, uint64(handle))';assert text.count(before)==1;mutant=art/'released-registry.go';mutant.write_text(text.replace(before,'// Mutant retains the released handle.'));overlay=art/'released-registry.json';overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}));mutantarchive=art/'released-registry.a';lib('released-registry',mutantarchive,overlay=overlay);binary=art/'released-registry';build('released-registry-build',release,binary,mutantarchive)
for mode,_,_,_ in question_mutants[:4]:
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

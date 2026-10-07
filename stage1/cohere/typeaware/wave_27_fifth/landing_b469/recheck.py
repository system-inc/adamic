from pathlib import Path
import subprocess,os,re,time,json,statistics
root=Path(__file__).resolve().parents[5];out=Path(__file__).resolve().parent
scratch=Path('/workspace/wave-27-scratch');env=dict(os.environ);env.pop('ADAMIC_TSGO_TIMING',None)
# These compiler/parser/bridge and owned rule inputs must remain unchanged.
changes=subprocess.check_output(['git','diff','--name-only','b84a9d931','origin/area/stage1-lint','--','cmd','internal','bridge','cohere','go.mod','go.work','stage1/typescript','stage1/cohere/typeaware'],cwd=root)
assert not changes,changes
print('Unchanged compiler, bridge, parser and owned rules verified',flush=True)
def run(label,args,code=0,go=False):
 with (out/(label+'.stdout')).open('wb') as stdout,(out/(label+'.stderr')).open('wb') as stderr:
  start=time.perf_counter_ns();r=subprocess.run([str(a) for a in args],stdout=stdout,stderr=stderr,env=env);elapsed=time.perf_counter_ns()-start
 data=(out/(label+'.stdout')).read_bytes();error=(out/(label+'.stderr')).read_bytes();assert r.returncode==code,(label,r.returncode,error)
 if go:assert re.fullmatch(rb'cohere: load_ns=\d+ rule_ns=\d+ run_ns=\d+\n',error),error
 elif code==70:assert error==b'adamic: panic: invalid or released checker handle\n',error
 else:assert not error,(label,error)
 return data,elapsed
batches=[('first','f801-first','wave27','wave27-asan','wave27-oracle','controls.manifest',['orm','serializable','array']),('next','f801-next','native','native-asan','oracle','controls.manifest',['process','race','blocking']),('third','f801-third','native','native-asan','oracle','controls.manifest',['effect-regex','rest','hooks','regex']),('fifth','fifth-c019','native','native-asan','oracle','valid.manifest',['symbol','typeof','await'])]
for batch,d,n,a,o,manifest,mutants in batches:
 art=scratch/d;config=art/'tsconfig.json'
 for corpus,setting,m in [('controls',config,art/manifest),('repository',root/'tsconfig.json',scratch/'repository.manifest'),('compiler',scratch/'typescript/src/compiler/tsconfig.json',scratch/'compiler.manifest')]:
  truth,_=run(batch+'-'+corpus+'-go',[art/o,setting,m],go=True)
  for engine in [n,a]:
   actual,_=run(batch+'-'+corpus+'-'+engine,[art/engine,setting,m]);assert truth==actual,(batch,corpus,engine)
  print(batch,corpus,len(truth),'identical bytes normal and sanitized',flush=True)
  if corpus=='controls':control=truth
 for mutant in mutants:
  actual,_=run(batch+'-'+mutant+'-mutant',[art/(mutant+'-mutant'),config,art/manifest]);assert actual!=control
  offset=next((i for i,(x,y) in enumerate(zip(control,actual)) if x!=y),min(len(control),len(actual)))
  print(batch,mutant,'mutant exit 0, empty stderr; Go catches byte',offset,flush=True)
 if batch=='first':
  run(batch+'-released',[art/'released',config,art/'probe.a'],70);run(batch+'-retained',[art/'released-mutant',config,art/'probe.a'])
 elif batch in ['next','third']:
  modes=['wave-27-declaration-ancestry','binding-declarations'] if batch=='third' else ['wave-27-declaration-ancestry','wave-27-syntax-flow','wave-27-call-declaration','wave-27-module-sources']
  for mode in modes:
   run(batch+'-released-'+mode,[art/'released',config,art/'probe.a',mode],70);run(batch+'-retained-'+mode,[art/'released-registry',config,art/'probe.a',mode])
 else:
  for mode in ['symbol','call','type','heritage']:
   run(batch+'-released-'+mode,[art/'released',config,(art/manifest).read_text().splitlines()[0],mode],70)
  run(batch+'-retained',[art/'retained',config,art/'released-symbol.a'])
 print(batch,'released panic 70; retained-registry mutants exit 0 and fail required refusal',flush=True)
measurements={};art=scratch/'fifth-c019'
for corpus,setting,m in [('repository',root/'tsconfig.json',scratch/'repository.manifest'),('compiler',scratch/'typescript/src/compiler/tsconfig.json',scratch/'compiler.manifest')]:
 samples={'native':[],'oracle':[]};streams=[]
 for round in range(3):
  for engine in ['native','oracle'] if round%2==0 else ['oracle','native']:
   data,elapsed=run(corpus+'-timed-'+str(round)+'-'+engine,[art/engine,setting,m],go=engine=='oracle');samples[engine].append(elapsed);streams.append(data)
 assert all(s==streams[0] for s in streams)
 medians={k:statistics.median(v) for k,v in samples.items()};measurements[corpus]={'samples_ns':samples,'median_ns':medians,'native_over_go':medians['native']/medians['oracle']}
(out/'measurements.json').write_text(json.dumps(measurements,indent=2)+'\n');print(json.dumps(measurements),flush=True);print('PASS',flush=True)

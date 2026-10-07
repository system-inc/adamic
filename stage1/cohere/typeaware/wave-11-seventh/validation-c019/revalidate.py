import concurrent.futures,json,os,pathlib,subprocess,time
root='/workspace/adamic';logs=pathlib.Path('/workspace/wave-11-logs');env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/workspace/wave-11-typescript'
suites=[('first','TestWave11AgreementAndMutants','ADAMIC_WAVE_11_ARTIFACTS'),('next','TestWave11NextAgreementAndMutants','ADAMIC_WAVE_11_NEXT_ARTIFACTS'),('third','TestWave11ThirdAgreementAndMutants','ADAMIC_WAVE_11_THIRD_ARTIFACTS'),('fourth','TestWave11FourthAgreementAndMutants','ADAMIC_WAVE_11_FOURTH_ARTIFACTS'),('fifth','TestWave11FifthAgreementAndMutants','ADAMIC_WAVE_11_FIFTH_ARTIFACTS'),('listeners','TestWave11ListenerDeclarationsAndMutants','ADAMIC_WAVE_11_LISTENER_ARTIFACTS')]
jobs=[]
for name,test,key in suites:
 e=env.copy();e[key]='/tmp/wave-11-c019-'+name;jobs.append((name,['go','test','./stage1/cohere/typeaware','-run','^'+test+'$','-count=1','-timeout','15m','-v'],e))
jobs.append(('bridge',['go','test','./bridge/tsgo/...','-count=1','-timeout','10m','-v'],env))
filt=r'^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(sorting|string_index|functions|closures|devirtualize|call_targets_(closure|element|region|reuse|sort)|library_map_set_iterator_.*|047cb0d_n_.*)\.a$'
jobs.append(('compiler',['go','test','./internal/oracle','-run',filt,'-count=1','-timeout','10m','-v'],env))
jobs.append(('vet',['go','vet','./stage1/cohere/typeaware','./bridge/tsgo/...'],env))
def run(job):
 name,args,e=job;start=time.monotonic();log=logs/('c019-'+name+'.log')
 with log.open('w') as out: result=subprocess.run(args,cwd=root,env=e,stdout=out,stderr=subprocess.STDOUT)
 r={'name':name,'argv':args,'exit':result.returncode,'seconds':round(time.monotonic()-start,3),'log':str(log)}
 print(json.dumps(r),flush=True);return r
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool: results=list(pool.map(run,jobs))
(logs/'c019-results.json').write_text(json.dumps(results,indent=2)+'\n')
raise SystemExit(any(r['exit']!=0 for r in results))

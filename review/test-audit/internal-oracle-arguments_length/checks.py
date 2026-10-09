import pathlib,json,subprocess,os,time,difflib
p=pathlib.Path('review/test-audit/internal-oracle-arguments_length');(p/'checks').mkdir(exist_ok=True);plans=[]
f='internal/oracle/cache_test.go';source=pathlib.Path(f).read_text()
def add(id,old,new,kind):
 assert source.count(old)==1,(id,old);plans.append(dict(id=id,file=f,line=source[:source.index(old)].count('\n')+1,old=old,new=new,kind=kind))
add('S01','return cacheKey("adamic-native-result-v1", code, runtimeKey, nodeVersion, context)','return cacheKey("adamic-native-result-v1", "", runtimeKey, nodeVersion, context)','change constant')
add('S02','return cacheKey("adamic-native-result-v1", code, runtimeKey, nodeVersion, context)','return cacheKey("adamic-native-result-v1", code, "", nodeVersion, context)','change constant')
add('S03','return cacheKey("adamic-node-result-v1", source, javascript, nodeVersion, context)','return cacheKey("adamic-node-result-v1", "", javascript, nodeVersion, context)','change constant')
add('S04','\tif err := os.Rename(temporary.Name(), filepath.Join(cache.directory, key+".json")); err != nil {\n\t\tt.Fatal(err)\n\t}\n','', 'drop statement')
add('S05','if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {\n\t\tcache.misses[kind].Add(1)','if os.Getenv("ADAMIC_GATE_UNCACHED") == "2" {\n\t\tcache.misses[kind].Add(1)','change constant')
wf='internal/oracle/oracle_test.go';ws=pathlib.Path(wf).read_text();h='func disagreement(oracle run, native run) string {';plans.append(dict(id='W01',file=wf,line=ws[:ws.index(h)].count('\n')+1,old=h,new=h+'\n\tif true {return ""}',kind='weaken comparison at entry'))
(p/'check-plan.json').write_text(json.dumps(plans,indent=2));results=[];cachetests='TestGateCacheGeneratedC TestGateCacheToolchains TestGateCacheNodeInputs TestGateCacheAtomicEvidence TestGateCacheUncached'.split()
for q in plans:
 file=pathlib.Path(q['file']);before=file.read_text();after=before.replace(q['old'],q['new']);(p/'checks'/(q['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+q['file'],tofile='b/'+q['file'])))
 try:
  file.write_text(after)
  with (p/'logs'/('vet-'+q['id']+'.log')).open('w') as out:subprocess.run(['go','vet','./internal/oracle/'],stdout=out,stderr=subprocess.STDOUT,check=True)
  tests=cachetests if q['id'].startswith('S') else ['TestArgumentsLengthWrongSlotMutant','TestCheckedCastMutants','TestGateCacheAtomicEvidence'];cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(tests)+')$'];start=time.monotonic()
  with (p/'logs'/(q['id']+'.log')).open('w') as out:r=subprocess.run(cmd,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u055/cache/'+q['id']),stdout=out,stderr=subprocess.STDOUT)
  es=[]
  for s in (p/'logs'/(q['id']+'.log')).read_text().splitlines():
   if s.startswith('{'):
    try:es.append(json.loads(s))
    except:pass
  results.append(dict(id=q['id'],exit=r.returncode,command=' '.join(cmd),wall_seconds=time.monotonic()-start,events=es,kills=sorted(set(e['Test'].split('/')[0] for e in es if e['Action']=='fail' and e.get('Test')))));(p/'check-results.json').write_text(json.dumps(results,indent=2));print(q['id'],results[-1]['kills'],flush=True)
 finally:file.write_text(before)

import pathlib,json,subprocess,os,time,difflib
p=pathlib.Path('review/test-audit/internal-oracle-arguments_length');(p/'probes').mkdir(exist_ok=True);rows=json.loads((p/'rows.json').read_text());names=[r['test'] for r in rows];normal=[names[i] for i in [0,7,8,9,10,11,12]];plans=[]
def add(id,f,h,ans,tests):
 s=pathlib.Path(f).read_text();assert s.count(h)==1,(id,h);plans.append(dict(id=id,file=f,entry=h,line=s[:s.index(h)].count('\n')+1,empty=ans,tests=tests))
add('P01','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return &ir.Program{}, nil',normal)
add('P02','internal/native/emit.go','func C(program *ir.Program) string {','return ""',[n for n in normal if n not in [names[10],names[11]]])
add('P03','internal/javascript/javascript.go','func JavaScript(program *ir.Program) string {','return ""',[n for n in normal if n!=names[10]])
add('P04','internal/oracle/cache_test.go','func nativeResultKey(code, runtimeKey, nodeVersion, context string) string {','return ""',names[2:4])
add('P05','internal/oracle/cache_test.go','func reusableResult[T any](t *testing.T, cache *resultCache, kind int, key string, execute func() T) T {','var value T; return value',[names[2],names[3],names[5]])
add('P06','internal/oracle/cache_test.go','func nodeResultKey(source, javascript, nodeVersion, context string) string {','return ""',[names[3],names[4]])
add('P07','internal/oracle/cache_test.go','func sourceIdentity(t *testing.T, path string) string {','return ""',[names[4]])
add('P08','internal/oracle/cache_test.go','func cacheKey(parts ...string) string {','return ""',[names[5],names[6]])
add('P09','internal/oracle/cache_test.go','func cachedResult[T any](t *testing.T, cache *resultCache, kind int, key string, execute func() T) T {','var value T; return value',[names[6]])
(p/'probe-plan.json').write_text(json.dumps(plans,indent=2));original={q['file']:pathlib.Path(q['file']).read_text() for q in plans};switch=dict(original);results=[]
try:
 for q in plans:
  before=original[q['file']];after=before.replace(q['entry'],q['entry']+'\n\tif true { '+q['empty']+' }');(p/'probes'/(q['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+q['file'],tofile='b/'+q['file'])))
  pathlib.Path(q['file']).write_text(after)
  with (p/'logs'/('vet-'+q['id']+'.log')).open('w') as out:subprocess.run(['go','vet','./'+str(pathlib.Path(q['file']).parent)+'/'],stdout=out,stderr=subprocess.STDOUT,check=True)
  pathlib.Path(q['file']).write_text(before)
  # os is already imported in oracle's cache file; production packages get a private helper.
  cond='os.Getenv("ADAMIC_MUTANT")=="'+q['id']+'"' if q['file'].startswith('internal/oracle/') else 'auditProbe("'+q['id']+'")'
  switch[q['file']]=switch[q['file']].replace(q['entry'],q['entry']+'\n\tif '+cond+' { '+q['empty']+' }')
 for f,s in switch.items():pathlib.Path(f).write_text(s)
 for package in ['lower','native','javascript']:pathlib.Path('internal/'+package+'/audit_probe.go').write_text('package '+package+'\nimport "os"\nfunc auditProbe(id string)bool{return os.Getenv("ADAMIC_MUTANT")==id}\n')
 with (p/'logs'/'compile-probes.log').open('w') as out:subprocess.run(['go','test','-c','-o','/tmp/u055-probes.test','./internal/oracle/'],stdout=out,stderr=subprocess.STDOUT,check=True)
 for q in plans:
  for n in q['tests']:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+n+'$'];log=p/'logs'/(q['id']+'-'+n+'.log');start=time.monotonic()
   with log.open('w') as out:r=subprocess.run(cmd,env=dict(os.environ,ADAMIC_MUTANT=q['id'],ADAMIC_BUILD_CACHE_DIR='/tmp/u055/cache/'+q['id']),stdout=out,stderr=subprocess.STDOUT)
   es=[]
   for s in log.read_text().splitlines():
    if s.startswith('{'):
     try:es.append(json.loads(s))
     except:pass
   results.append(dict(id=q['id'],test=n,exit=r.returncode,command=' '.join(cmd),wall_seconds=time.monotonic()-start,events=es));(p/'probe-results.json').write_text(json.dumps(results,indent=2));print(q['id'],n,r.returncode,flush=True)
finally:
 for f,s in original.items():pathlib.Path(f).write_text(s)
 for package in ['lower','native','javascript']:pathlib.Path('internal/'+package+'/audit_probe.go').unlink(missing_ok=True)

import os,json,pathlib,subprocess,time,signal,re
R=pathlib.Path('review/test-defend/deletion-set/internal-fuzz');items=json.load(open(R/'mutant-list.json'));C=json.load(open(R/'skip.json'));results=[]
baseline_events=[json.loads(s) for s in (R/'logs/baseline.log').read_text().splitlines() if s.startswith('{')];last=baseline_events[-1];assert last['Action']=='pass'
baseline=dict(command="ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/deletion-fuzz/cache/baseline go test -json -count=1 -timeout 30m ./internal/fuzz/ -skip '^("+'|'.join(C)+")($|_|/)'",binary_seconds=last['Elapsed'],failed=[],log=str(R/'logs/baseline.log'))
WITNESS=re.compile(r'Mutants?(?:Unit\d+)?$|MutantKilled|PlantedFailure');PINS=set()
def run(name,extra):
 cache='/tmp/deletion-fuzz/cache/'+name;os.makedirs(cache,exist_ok=True);env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=cache)
 skip='^('+'|'.join(C+extra)+')($|_|/)';cmd=['go','test','-json','-count=1','-timeout','30m','./internal/fuzz/','-skip',skip];log=R/'logs'/(name+'.log');start=time.monotonic();failed=[];witness=[];pins=[];panic=None;active=None;stopped=False
 with log.open('w') as out:
  p=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  with log.open() as inp:
   while True:
    s=inp.readline()
    if not s:
     if p.poll() is not None:break
     time.sleep(.1);continue
    try:e=json.loads(s)
    except ValueError:continue
    t=e.get('Test','').split('/')[0]
    if e.get('Action')=='run' and t:active=t
    if e.get('Output','').startswith('panic:'):panic=t or active
    if e.get('Action')=='fail' and t and t not in failed:
     failed.append(t)
     if WITNESS.search(t):witness.append(t)
     elif t in PINS:pins.append(t)
     elif not panic and not stopped:
      os.killpg(p.pid,signal.SIGTERM);stopped=True
   p.wait()
 a=dict(command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+cache+' '+ ' '.join(cmd),wall_seconds=round(time.monotonic()-start,3),exit=p.returncode,failed=failed,witness_failures=witness,pin_failures=pins,panic=panic,stopped_after_catch=stopped,log=str(log))
 (R/'logs'/(name+'.json')).write_text(json.dumps(a,indent=2)+'\n');print(name,json.dumps(a),flush=True);return a
for e in items:
 result=dict(e,still_caught_by=[],runs=[],witness_failures=[],pin_failures=[],panicking_tests=[])
 if not e['stale']:
  subprocess.run(['git','apply',e['diff']],check=True)
  try:
   extra=[]
   for retry in range(100):
    a=run(e['replay']+'-'+str(retry),extra);result['runs'].append(a)
    result['witness_failures']=sorted(set(result['witness_failures']+a['witness_failures']));result['pin_failures']=sorted(set(result['pin_failures']+a['pin_failures']))
    if a['panic']:
     if a['panic'] in extra:result['broken']='unresolved panic';break
     extra.append(a['panic']);result['panicking_tests'].append(a['panic']);continue
    result['still_caught_by']=[t for t in a['failed'] if t not in a['witness_failures']+a['pin_failures']]
    result['completed_without_valid_catch']=not result['still_caught_by'] and not a['stopped_after_catch'] and (a['exit']==0 or bool(a['failed']))
    if not a['failed'] and a['exit']!=0:result['broken']='build or execution failure without a top-level test result'
    break
  finally:subprocess.run(['git','apply','-R',e['diff']],check=True)
 results.append(result);(R/'matrix.json').write_text(json.dumps(dict(baseline=baseline,mutants=results),indent=2)+'\n')
print('COMPLETE',flush=True)

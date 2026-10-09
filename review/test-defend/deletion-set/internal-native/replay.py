import os,json,pathlib,subprocess,time,signal,re
ROOT=pathlib.Path('review/test-defend/deletion-set/internal-native');ROOT.joinpath('logs').mkdir(exist_ok=True)
items=json.loads((ROOT/'mutant-list.json').read_text()); candidates=json.loads((ROOT/'skip.json').read_text())
WITNESS=re.compile(r'Mutants?(?:Unit\d+)?$|MutantKilled|PlantedFailure|Catches|TestOptionalMethodThunksMatchNode$')
def run(name,extra=[],early=False):
 cache='/tmp/deletion-native/cache/'+name;os.makedirs(cache,exist_ok=True)
 skip='^('+'|'.join(candidates+extra)+')($|_|/)'
 cmd=['go','test','-json','-count=1','-timeout','30m','./internal/native/','-skip',skip]
 env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=cache)
 log=ROOT/'logs'/(name+'.log');started=time.monotonic();failed=[];witness=[];panic=None;active=None;stopped=False
 with log.open('w') as out:
  p=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  with log.open() as inp:
   while True:
    line=inp.readline()
    if not line:
     if p.poll() is not None:break
     time.sleep(.2);continue
    try:e=json.loads(line)
    except ValueError:continue
    t=e.get('Test','').split('/')[0]
    if e.get('Action')=='run' and t:active=t
    output=e.get('Output','')
    if output.startswith('panic:'):
     panic=t or active
    if e.get('Action')=='fail' and t and t not in failed:
     failed.append(t)
     if WITNESS.search(t):witness.append(t)
     elif early and not panic:
      os.killpg(p.pid,signal.SIGTERM);stopped=True
   p.wait()
 result=dict(name=name,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+cache+' '+' '.join(cmd),seconds=round(time.monotonic()-started,3),exit=p.returncode,failed=failed,witness_failures=witness,panic=panic,stopped_after_catch=stopped,log=str(log))
 (ROOT/'logs'/(name+'.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True);return result
# Check every diff before baseline, keeping stale evidence untouched.
for e in items:
 p=subprocess.run(['git','apply','--check',e['diff']],capture_output=True,text=True)
 (ROOT/'logs'/(e['replay']+'-apply.log')).write_text(p.stdout+p.stderr)
 e['stale']=p.returncode!=0
(ROOT/'mutant-list.json').write_text(json.dumps(items,indent=2)+'\n')
baseline=run('baseline')
if baseline['exit']!=0:
 (ROOT/'matrix.json').write_text(json.dumps(dict(baseline=baseline,mutants=[]),indent=2)+'\n');raise SystemExit('RED BASELINE, STOPPED')
results=[]
for e in items:
 result=dict(e,still_caught_by=[],runs=[],panicking_tests=[],witness_failures=[])
 if not e['stale']:
  subprocess.run(['git','apply',e['diff']],check=True)
  try:
   extra=[]
   for attempt in range(100):
    r=run(e['replay']+'-'+str(attempt),extra,True);result['runs'].append(r)
    result['witness_failures']=sorted(set(result['witness_failures']+r['witness_failures']))
    if r['panic']:
     if r['panic'] in extra:result['broken']='unresolved panic';break
     extra.append(r['panic']);result['panicking_tests'].append(r['panic']);continue
    result['still_caught_by']=[t for t in r['failed'] if t not in r['witness_failures']]
    result['completed_pass']=r['exit']==0
    if 'error: unused function' in pathlib.Path(r['log']).read_text():
     result['build_failure_rows']=result['still_caught_by'];result['still_caught_by']=[];result['broken']='C compilation fails under -Werror'
     break
    if not result['still_caught_by'] and r['exit']!=0:result['broken']='nonpassing run without a clean non-witness catch'
    break
  finally:subprocess.run(['git','apply','-R',e['diff']],check=True)
 results.append(result)
 (ROOT/'matrix.json').write_text(json.dumps(dict(baseline=baseline,mutants=results),indent=2)+'\n')
print('REPLAY COMPLETE',flush=True)

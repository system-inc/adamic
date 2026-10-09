import os,sys,json,subprocess,time,signal,re
from pathlib import Path
P=Path('review/test-audit/bridge-tsgo-registered_units'); S=Path('/tmp/u004'); ROOT=Path.cwd(); PACKAGE='github.com/system-inc/adamic/bridge/tsgo'
rows=re.findall(r'func (Test\w+)\(t \*testing.T\)',Path('bridge/tsgo/registered_units_test.go').read_text())
(P/'scope.json').write_text(json.dumps({'rows':rows,'starting_commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'missing':[x for x in rows if x not in (S/'list.log').read_text()],'corpus_commit':subprocess.check_output(['git','-C',str(S/'typescript'),'rev-parse','HEAD'],text=True).strip()},indent=2))
def run(label,pattern,corpus=False,selector='',overlay=None,cache=None):
 env=dict(os.environ,ADAMIC_MUTANT=selector);env['ADAMIC_BUILD_LOG']=str(S/(label+'-builds.log'));env.pop('ADAMIC_TEST_SHARD',None);env.pop('ADAMIC_UNIT_BUDGET',None)
 if corpus:env['ADAMIC_TSGO_CORPUS']=str(S/'typescript')
 else:env.pop('ADAMIC_TSGO_CORPUS',None)
 if cache:env['ADAMIC_BUILD_CACHE_DIR']=cache
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s']
 if overlay:cmd+=['-overlay',overlay]
 cmd+=['./bridge/tsgo/','-run',pattern];start=time.monotonic(); log=S/(label+'.log')
 with open(log,'w') as f:
  proc=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  try:code=proc.wait(timeout=120)
  except subprocess.TimeoutExpired:os.killpg(proc.pid,signal.SIGKILL);code=124
  # End orphaned subprocesses after the test process exited.
  try:os.killpg(proc.pid,signal.SIGTERM)
  except ProcessLookupError:pass
 events=[]
 for line in log.read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 status={e['Test']:e['Action'] for e in events if e.get('Test') and '/' not in e['Test'] and e['Action'] in ['pass','fail','skip']};elapsed=next((e.get('Elapsed') for e in reversed(events) if not e.get('Test') and e['Action'] in ['pass','fail']),None)
 d={'label':label,'command':' '.join(cmd),'corpus':corpus,'selector':selector,'cache':cache,'exit':code,'wall_seconds':time.monotonic()-start,'binary_seconds':elapsed,'status':status,'events':events,'cooked':any('test timed out' in e.get('Output','') for e in events) or code==124}
 (P/(label+'.json')).write_text(json.dumps(d,indent=2)); print(label,code,round(d['wall_seconds'],3),status,flush=True);return d
if __name__=='__main__':
 if sys.argv[1]=='baseline':
  for row in rows:
   corpus=any(row.endswith(x) for x in ['Checker','Parser','Types','Utilities'])
   for trial in range(1,4):
    d=run(row+'-timing-'+str(trial),'^'+row+'$',corpus)
    if d['exit'] and not d['cooked']:sys.exit('RED BASELINE '+row)
    if d['cooked']:break

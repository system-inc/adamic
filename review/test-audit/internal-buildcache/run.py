import os,subprocess,time,json,signal,re,sys
from pathlib import Path
P=Path('review/test-audit/internal-buildcache');S=Path('/tmp/u015');PACKAGE='./internal/buildcache/'
def run(label,selector='',pattern='.',extra=[]):
 env=dict(os.environ,ADAMIC_MUTANT=selector);cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s']+extra+[PACKAGE,'-run',pattern];s=time.monotonic()
 with open(S/(label+'.log'),'w') as f:
  proc=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  try:code=proc.wait(timeout=120)
  except subprocess.TimeoutExpired:os.killpg(proc.pid,signal.SIGKILL);code=124
  try:os.killpg(proc.pid,signal.SIGTERM)
  except ProcessLookupError:pass
 events=[]
 for l in (S/(label+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except ValueError:pass
 status={e['Test']:e['Action'] for e in events if e.get('Test') and '/' not in e['Test'] and e['Action'] in ['pass','fail','skip']};subcases={e['Test']:e['Action'] for e in events if '/' in e.get('Test','') and e['Action'] in ['pass','fail','skip']}
 d={'label':label,'selector':selector,'command':'ADAMIC_MUTANT='+repr(selector)+' '+' '.join(cmd),'exit':code,'wall_seconds':time.monotonic()-s,'binary_seconds':next((e.get('Elapsed') for e in reversed(events) if not e.get('Test') and e['Action'] in ['pass','fail']),None),'status':status,'subcases':subcases,'events':events,'cooked':code==124 or any('test timed out' in e.get('Output','') for e in events)};(P/(label+'.json')).write_text(json.dumps(d,indent=2));print(label,code,round(d['wall_seconds'],3),status,flush=True);return d
if __name__=='__main__':
 rows=[x for x in (S/'list.log').read_text().splitlines() if x.startswith('Test')]
 if sys.argv[-1]=='timing':
  for row in rows:
   for n in [1,2,3]:
    d=run(row+'-timing-'+str(n),pattern='^'+row+'$')
    if d['exit']:raise SystemExit('timing failed')
 else:
  for m in json.load(open(P/'plan.json')):
   d=run(m['id'],m['id'])
   if len(d['status'])!=len(rows) or any('panic:' in e.get('Output','') for e in d['events']):
    for row in rows:run(m['id']+'-'+row,m['id'],'^'+row+'$')

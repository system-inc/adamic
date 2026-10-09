from pathlib import Path
import os,subprocess,time,json,re,signal
root=Path(__file__).resolve().parents[3]
out=Path(__file__).resolve().parent
names=[x.strip() for x in Path('/tmp/u007-test-list.log').read_text().splitlines() if x.startswith('Test')]
mutants=json.loads((out/'mutants.json').read_text())
records=[]
def invoke(mid,pattern,scope):
 env=dict(os.environ,ADAMIC_MUTANT=mid,ADAMIC_NATIVE_SPLIT='u007-'+mid)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic/','-run',pattern]
 log=out/(mid+'-'+scope+'.log'); start=time.monotonic(); cooked=False
 with log.open('w') as stream:
  process=subprocess.Popen(cmd,cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
  try: code=process.wait(timeout=90)
  except subprocess.TimeoutExpired:
   cooked=True; os.killpg(process.pid,signal.SIGTERM); code=process.wait(timeout=5)
 events=[]
 for line in log.read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 status={}; output={}
 for e in events:
  name=e.get('Test','')
  if name:
   if e['Action'] in ('pass','fail','skip'): status[name]=e['Action']
   if e['Action']=='output': output.setdefault(name,[]).append(e.get('Output',''))
 failed=sorted({n.split('/')[0] for n,s in status.items() if s=='fail'})
 evidence={}
 for n,texts in output.items():
  if n.split('/')[0] in failed:
   lines=''.join(texts).splitlines()
   relevant=[x.strip() for x in lines if re.search(r'_test.go:\d+:|^panic:',x)]
   if relevant: evidence[n]=relevant
 top={n:status.get(n,'unknown') for n in names if scope=='package' or pattern=='^'+n+'$'}
 pkg=next((e for e in reversed(events) if 'Test' not in e and e['Action'] in ('pass','fail')),{})
 result=dict(mutant=mid,scope=scope,command='ADAMIC_MUTANT='+mid+' ADAMIC_NATIVE_SPLIT=u007-'+mid+' '+' '.join(cmd),wall_seconds=time.monotonic()-start,binary_seconds=pkg.get('Elapsed'),exit=code,cooked=cooked,status=top,failed_rows=failed,failing_lines=evidence,log=str(log.relative_to(root)))
 records.append(result); (out/'matrix-runs.json').write_text(json.dumps(records,indent=2)+'\n')
 print(mid,scope,round(result['wall_seconds'],3),failed,top,flush=True)
 return result
for m in mutants:
 r=invoke(m['id'],'.','package')
 if any(v=='unknown' for v in r['status'].values()):
  for n in names:invoke(m['id'],'^'+n+'$',n)

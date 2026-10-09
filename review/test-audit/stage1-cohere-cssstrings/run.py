import subprocess,time,json,os,signal,sys
from pathlib import Path
E=Path('/workspace/adamic/review/test-audit/stage1-cohere-cssstrings')
def run(label,pattern='.',environment=None):
 command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/cssstrings/','-run',pattern]
 start=time.monotonic();env=os.environ.copy();env.update(environment or {})
 with (E/(label+'.log')).open('w') as f:
  process=subprocess.Popen(command,stdout=f,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  try:code=process.wait(timeout=90);cooked=False
  except subprocess.TimeoutExpired:
   children={process.pid};parents={}
   for path in Path('/proc').iterdir():
    if path.name.isdigit():
     try:parents[int(path.name)]=int((path/'stat').read_text().split(') ')[1].split()[1])
     except (OSError,ValueError,IndexError):pass
   while True:
    more={pid for pid,parent in parents.items() if parent in children}
    if more.issubset(children):break
    children.update(more)
   for pid in reversed(sorted(children)):
    try:os.kill(pid,signal.SIGKILL)
    except ProcessLookupError:pass
   process.wait();code=124;cooked=True
 elapsed=time.monotonic()-start
 events=[]
 for line in (E/(label+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except json.JSONDecodeError:pass
 record=dict(label=label,command=command,environment=environment or {},wall_seconds=round(elapsed,3),exit=code,cooked=cooked,binary_seconds=next((x['Elapsed'] for x in reversed(events) if x.get('Action') in ['pass','fail'] and not x.get('Test')),None),failed_tests=[x['Test'] for x in events if x.get('Action')=='fail' and x.get('Test') and '/' not in x['Test']])
 with (E/'runs.jsonl').open('a') as f:f.write(json.dumps(record)+'\n')
 print(json.dumps({k:v for k,v in record.items() if k != "failed_tests"} | {"failed_test_count":len(record["failed_tests"])}),flush=True)
 return record
if __name__=='__main__':
 if sys.argv[1]=='timings':
  for name,pattern in [('family','^TestCSSStrings(_[0-9]{3}|_Setup|Union)$'),('gap','^TestMultiPushGap$'),('witness','^TestCSSStringsPlantedDisagreement$')]:
   for i in range(3):run('timing-'+name+'-'+str(i+1),pattern)

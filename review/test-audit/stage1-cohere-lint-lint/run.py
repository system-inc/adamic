import subprocess,time,json,os,signal
from pathlib import Path
E=Path(__file__).resolve().parent
def run(label,pattern,env=None,extra=None):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',pattern]+(extra or [])
 start=time.monotonic();environment=os.environ.copy();environment.update(env or {})
 with (E/(label+'.log')).open('w') as f:
  p=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,env=environment,start_new_session=True)
  try:rc=p.wait(timeout=90);cooked=False
  except subprocess.TimeoutExpired:
   children={p.pid};parents={}
   for q in Path('/proc').iterdir():
    if q.name.isdigit():
     try:parents[int(q.name)]=int((q/'stat').read_text().split(') ')[1].split()[1])
     except:pass
   while True:
    new={pid for pid,parent in parents.items() if parent in children};old=len(children);children|=new
    if len(children)==old:break
   for pid in sorted(children,reverse=True):
    try:os.kill(pid,signal.SIGKILL)
    except ProcessLookupError:pass
   p.wait();rc=124;cooked=True
 events=[]
 for x in (E/(label+'.log')).read_text().splitlines():
  try:events.append(json.loads(x))
  except:pass
 r=dict(label=label,command=cmd,environment=env or {},exit=rc,cooked=cooked,wall_seconds=round(time.monotonic()-start,3),binary_seconds=next((e['Elapsed'] for e in reversed(events) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None),failed_tests=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')],passed_tests=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test')],skipped_tests=[e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')])
 with (E/'runs.jsonl').open('a') as f:f.write(json.dumps(r)+'\n')
 print(json.dumps({k:v for k,v in r.items() if not k.endswith('_tests')}),flush=True)
 return r

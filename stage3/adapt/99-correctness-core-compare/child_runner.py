import json,os,signal,subprocess,time,threading
from pathlib import Path
ROOT=Path('/workspace/scratch/core-compare-amend');EVENTS=json.loads((ROOT/'CHILD-EXITS.json').read_text()) if (ROOT/'CHILD-EXITS.json').exists() else [];LOCK=threading.Lock()
class ChildExit(Exception):
 def __init__(self,code,label):self.code=code;self.label=label;super().__init__(f'{label}: child exit {code}')
def run(label,args,cwd='/workspace/adamic',env=None,accept=(0,),timeout=1700):
 log=ROOT/(label+'.log');log.parent.mkdir(parents=True,exist_ok=True);begin=time.monotonic();timed=False
 with log.open('w') as stream:
  p=subprocess.Popen([str(x) for x in args],cwd=cwd,env=env,stdout=stream,stderr=stream,start_new_session=True)
  try:p.wait(timeout=timeout)
  except subprocess.TimeoutExpired:
   timed=True
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 code=p.returncode
 event={'label':label,'command':[str(x) for x in args],'cwd':str(cwd),'childExitCode':code,'seconds':time.monotonic()-begin,'timedOut':timed,'log':str(log),'acceptedCodes':list(accept) if accept is not None else 'recorded validation subject'}
 with LOCK:
  EVENTS.append(event);(ROOT/'CHILD-EXITS.json').write_text(json.dumps(EVENTS,indent=2)+'\n');print(f'{label}: childExitCode={code}',flush=True)
 if timed or accept is not None and code not in accept:
  if code==0:raise RuntimeError(f'{label}: validation failed despite childExitCode=0; expected {accept}; timedOut={timed}')
  raise ChildExit(code,label)
 return event

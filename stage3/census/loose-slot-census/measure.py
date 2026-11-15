"""One child per measurement. Linux wait4 supplies peak RSS in KiB."""
import json,os,subprocess,time
from pathlib import Path
def run(args,dest,cwd=None,env=None):
 dest=Path(dest);dest.parent.mkdir(parents=True,exist_ok=True)
 start=time.monotonic()
 with dest.with_suffix('.log').open('wb') as log:
  p=subprocess.Popen(list(map(str,args)),stdout=log,stderr=log,cwd=cwd,env=env)
  _,status,usage=os.wait4(p.pid,0);p.returncode=os.waitstatus_to_exitcode(status)
 result=dict(command=list(map(str,args)),seconds=time.monotonic()-start,peak_rss_KiB=usage.ru_maxrss,exit_code=p.returncode)
 dest.with_suffix('.time.json').write_text(json.dumps(result,indent=2)+'\n')
 if p.returncode: raise RuntimeError(result)
 return result
if __name__=='__main__':
 import sys
 run(sys.argv[2:],sys.argv[1])

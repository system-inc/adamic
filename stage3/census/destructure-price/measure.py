import argparse,subprocess,os,time,json,resource,signal
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--out',required=True,type=Path);p.add_argument('--cwd',type=Path);p.add_argument('--seconds',type=int);p.add_argument('command',nargs=argparse.REMAINDER);a=p.parse_args();cmd=a.command[1:] if a.command[:1]==['--'] else a.command
a.out.parent.mkdir(parents=True,exist_ok=True);start=time.monotonic();peak=0;samples=0;timed_out=False
with a.out.with_suffix('.log').open('wb') as log:
 child=subprocess.Popen(cmd,cwd=a.cwd,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 while True:
  total=0; processes={}
  for proc in Path('/proc').iterdir():
   if not proc.name.isdigit():continue
   try:
    stat=(proc/'stat').read_text();rest=stat[stat.rfind(')')+2:].split()
    rss=0
    for line in (proc/'status').read_text().splitlines():
     if line.startswith('VmRSS:'):rss=int(line.split()[1]);break
    processes[int(proc.name)]=(int(rest[1]),int(rest[2]),rss)
   except (FileNotFoundError,ProcessLookupError,PermissionError):pass
  descendants={child.pid}
  while True:
   expanded=descendants|{pid for pid,(parent,group,rss) in processes.items() if parent in descendants}
   if expanded==descendants:break
   descendants=expanded
  total=sum(processes[pid][2] for pid in descendants if pid in processes)
  groups={processes[pid][1] for pid in descendants if pid in processes}
  peak=max(peak,total);samples+=1
  code=child.poll()
  if code is not None:break
  if a.seconds and time.monotonic()-start>=a.seconds:
   timed_out=True
   for group in groups:
    try:os.killpg(group,signal.SIGTERM)
    except ProcessLookupError:pass
   try:child.wait(timeout=10)
   except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait()
   code=child.returncode;break
  time.sleep(.2)
 result={'command':cmd,'cwd':str(a.cwd) if a.cwd else None,'exit':code,'timeout':timed_out,'limit_seconds':a.seconds,'wall_seconds':round(time.monotonic()-start,3),'peak_process_tree_rss_kib':peak,'sampling_interval_seconds':.2,'samples':samples,'max_descendant_rss_kib':resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss}
 a.out.with_suffix('.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
raise SystemExit(code if code>=0 else 128-code)

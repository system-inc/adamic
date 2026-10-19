import json,os,signal,subprocess,sys,time
from pathlib import Path
root=Path(sys.argv[1]).resolve();label=sys.argv[2];e=Path(sys.argv[3]).resolve();e.mkdir(parents=True,exist_ok=True)
cpu=len(os.sched_getaffinity(0)); env={**os.environ,'GOMAXPROCS':str(cpu)}
for key in ['ADAMIC_TEST_SHARD','ADAMIC_STAGE3_FIXTURE','ADAMIC_JSON_NODE_ONLY','ADAMIC_JSON_BENCH','ADAMIC_JSON_ARTIFACTS']:env.pop(key,None)
command=['go','test','./stage1/cohere/json','-count=1','-parallel='+str(cpu),'-timeout=30m']
cg=Path('/sys/fs/cgroup')
def read(name):
 try:return (cg/name).read_text().strip()
 except OSError:return None
def counters(name):
 text=read(name)
 return {k:int(v) for k,v in (l.split() for l in text.splitlines())} if text else {}
def descendants_rss(pid):
 records={}
 for p in Path('/proc').iterdir():
  if not p.name.isdecimal():continue
  try:
   fields=(p/'stat').read_text().rsplit(')',1)[1].split();records[int(p.name)]=(int(fields[1]),int(fields[21])*os.sysconf('SC_PAGE_SIZE'))
  except (OSError,ValueError,IndexError):continue
 own={pid};changed=True
 while changed:
  before=len(own);own.update(i for i,(parent,_) in records.items() if parent in own);changed=len(own)!=before
 return sum(records.get(i,(0,0))[1] for i in own)
start=time.monotonic();before=counters('memory.events');peak_current=0;peak_rss=0;peak_anon=0;peak_file=0;samples=0;timed_out=False;baseline=read('memory.current');baseline_stat=counters('memory.stat')
with (e/(label+'.log')).open('w') as log:
 proc=subprocess.Popen(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 while True:
  current=int(read('memory.current') or 0);stat=counters('memory.stat');peak_current=max(peak_current,current);peak_anon=max(peak_anon,stat.get('anon',0));peak_file=max(peak_file,stat.get('file',0));peak_rss=max(peak_rss,descendants_rss(proc.pid));samples+=1
  if proc.poll() is not None:break
  if time.monotonic()-start>1860:
   timed_out=True;os.killpg(proc.pid,signal.SIGKILL);proc.wait();break
  time.sleep(.05)
result=dict(label=label,commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),command=command,GOMAXPROCS=cpu,affinity_cpus=sorted(os.sched_getaffinity(0)),cpu_max=read('cpu.max'),memory_limit_bytes=read('memory.max'),memory_baseline_bytes=int(baseline),memory_baseline_anon_bytes=baseline_stat.get('anon'),memory_baseline_file_bytes=baseline_stat.get('file'),peak_cgroup_current_bytes=peak_current,peak_cgroup_anon_bytes=peak_anon,peak_cgroup_file_bytes=peak_file,peak_descendant_rss_bytes=peak_rss,samples=samples,sample_interval_seconds=.05,seconds=round(time.monotonic()-start,3),exit=proc.returncode,timed_out=timed_out,memory_events_before=before,memory_events_after=counters('memory.events'))
result['memory_events_delta']={k:v-before.get(k,0) for k,v in result['memory_events_after'].items()}
(e/(label+'.json.txt')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

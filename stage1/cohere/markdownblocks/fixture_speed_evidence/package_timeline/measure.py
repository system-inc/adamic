import subprocess,time,json,pathlib,datetime,os
out=pathlib.Path('/tmp/markdown-timeline')
start=time.time(); mono=time.monotonic()
with (out/'package.jsonl').open('w') as log,(out/'stderr.log').open('w') as err,(out/'samples.jsonl').open('w') as samples:
 p=subprocess.Popen(['go','test','-count=1','-json','-v','-timeout=30m','./stage1/cohere/markdownblocks'],stdout=log,stderr=err)
 n=0
 while True:
  stat=dict(line.split() for line in pathlib.Path('/sys/fs/cgroup/cpu.stat').read_text().splitlines())
  row={'epoch':time.time(),'command_seconds':time.monotonic()-mono,'cpu_stat':{k:int(v) for k,v in stat.items()},'memory_current':int(pathlib.Path('/sys/fs/cgroup/memory.current').read_text())}
  if n%5==0:
   procs={}
   for d in pathlib.Path('/proc').iterdir():
    if not d.name.isdigit():continue
    try:
     fields=(d/'stat').read_text().rsplit(')',1)[1].split();cmd=(d/'cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace').strip()
     procs[int(d.name)]={'pid':int(d.name),'ppid':int(fields[1]),'ticks':int(fields[11])+int(fields[12]),'command':cmd}
    except (FileNotFoundError,ProcessLookupError,PermissionError):pass
   descendants={p.pid}
   while True:
    more={pid for pid,r in procs.items() if r['ppid'] in descendants}
    if more<=descendants:break
    descendants|=more
   row['processes']=[r for pid,r in procs.items() if pid in descendants]
  samples.write(json.dumps(row)+'\n');samples.flush();n+=1
  if p.poll() is not None:break
  time.sleep(1)
 result={'command_start_epoch':start,'command_wall_seconds':time.monotonic()-mono,'exit':p.returncode,'memory_peak':int(pathlib.Path('/sys/fs/cgroup/memory.peak').read_text())}
 (out/'run.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

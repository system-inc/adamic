from pathlib import Path
import json,os,subprocess,time,sys
label,root=sys.argv[1:3]
out=Path('/tmp/v3-json-measure')/label; out.mkdir(exist_ok=True)
logs=out/'children';logs.mkdir(exist_ok=True)
env=dict(os.environ,PATH='/tmp/v3-json-measure/bin:'+os.environ['PATH'],ADAMIC_RSS_DIR=str(logs),ADAMIC_GATE_UNCACHED='1',GOMAXPROCS='4')
(out/'memory-events-before.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
with (out/'test.log').open('w') as log:
 p=subprocess.Popen(['/tmp/v3-json-measure/rss',str(out/'test.rss'),'/tmp/v3-json-measure/'+label+'.test','-test.run=^TestPortMatchesGoCohere$','-test.count=1','-test.v','-test.timeout=30m'],cwd=Path(root)/'stage1/cohere/json',env=env,stdout=log,stderr=subprocess.STDOUT)
 seen={}; maximum=0;started=time.monotonic();known={p.pid}
 while p.poll() is None:
  snapshot={}
  for entry in Path('/proc').iterdir():
   if not entry.name.isdecimal():continue
   try:
    status=entry.joinpath('status').read_text(); fields=dict(line.split(':',1) for line in status.splitlines() if ':' in line)
    pid=int(entry.name); parent=int(fields['PPid']);snapshot[pid]=(parent,fields,entry)
   except (OSError,ValueError,KeyError):continue
  descendants={p.pid}
  while True:
   new={pid for pid,(parent,_,_) in snapshot.items() if parent in descendants}
   if new<=descendants:break
   descendants|=new
  known|=descendants
  for pid in known & snapshot.keys():
   _,fields,entry=snapshot[pid]
   try:cmd=entry.joinpath('cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace')
   except OSError:cmd=''
   row=seen.setdefault(pid,{'pid':pid,'command':cmd,'peak_hwm_kib':0,'peak_rss_kib':0})
   if cmd: row['command']=cmd
   row['peak_hwm_kib']=max(row['peak_hwm_kib'],int(fields.get('VmHWM','0 kB').split()[0]))
   row['peak_rss_kib']=max(row['peak_rss_kib'],int(fields.get('VmRSS','0 kB').split()[0]))
  maximum=max(maximum,int(Path('/sys/fs/cgroup/memory.current').read_text()))
  time.sleep(.025)
 result=p.wait()
(out/'processes.json').write_text(json.dumps(list(seen.values()),indent=2))
(out/'summary.json').write_text(json.dumps({'exit':result,'wall_seconds':time.monotonic()-started,'sampled_cgroup_peak_bytes':maximum,'memory_limit_bytes':int(Path('/sys/fs/cgroup/memory.max').read_text())},indent=2))
(out/'memory-events-after.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
print(label,(out/'summary.json').read_text(),flush=True)

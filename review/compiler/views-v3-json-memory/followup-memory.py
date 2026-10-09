import time,json
from pathlib import Path
out=Path('/workspace/v3-json-artifacts/followup/harness-memory.jsonl')
with out.open('w') as f:
 while True:
  processes=[]
  for p in Path('/proc').iterdir():
   if not p.name.isdecimal():continue
   try:
    cmd=p.joinpath('cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace')
    if 'followup-v3.test' not in cmd and 'TestPortMatchesGoCohere' not in cmd:continue
    if cmd.startswith(('/bin/bash','bwrap','codex','/tmp/v3-json-measure/rss')):continue
    stat=dict(x.split(':',1) for x in p.joinpath('status').read_text().splitlines() if ':' in x)
    processes.append({'pid':int(p.name),'rss_kib':int(stat.get('VmRSS','0 kB').split()[0]),'hwm_kib':int(stat.get('VmHWM','0 kB').split()[0]),'cmd':cmd})
   except (OSError,ValueError):pass
  f.write(json.dumps({'monotonic':time.monotonic(),'cgroup_bytes':int(Path('/sys/fs/cgroup/memory.current').read_text()),'processes':processes})+'\n');f.flush()
  if not any('followup-v3.test' in p['cmd'] for p in processes):break
  time.sleep(1)

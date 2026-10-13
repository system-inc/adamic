"""Record bounded-run progress, live RSS and cgroup pressure every 30 seconds.
Usage: watch_stream.py RUN CHECKER_DIRECTORY OUTPUT_JSONL DEADLINE_EPOCH
"""
import json
from pathlib import Path
import sys
import time
from datetime import datetime,timezone
run,checker,output=(Path(x) for x in sys.argv[1:4]);deadline=float(sys.argv[4])
with output.open('a') as log:
 while time.time()<deadline:
  workers=[]
  for entry in Path('/proc').glob('[0-9]*'):
   try:
    command=(entry/'cmdline').read_bytes().split(b'\0')
    name=Path(command[0].decode()).name
    if name not in ('stream-census','progress-v2-census'):continue
    stat=(entry/'stat').read_text().rsplit(')',1)[1].split()
    if stat[0]=='Z':continue
    workers.append(dict(pid=int(entry.name),binary=name,rss_kib=int(stat[21])*4,
                        output=command[2].decode()))
   except (FileNotFoundError,ProcessLookupError,PermissionError,ValueError,IndexError):continue
  records=list((run/'records').glob('*.jsonl'))
  row=dict(utc=datetime.now(timezone.utc).isoformat(),completed_files=len(records),workers=workers)
  progress=checker/'progress.json'
  if progress.exists():
   try:row['checker_progress']=json.loads(progress.read_text())
   except json.JSONDecodeError:pass
  for name in ('memory.current','memory.events'):
   row[name]=(Path('/sys/fs/cgroup')/name).read_text().strip()
  log.write(json.dumps(row)+'\n');log.flush()
  time.sleep(min(30,max(0,deadline-time.time())))

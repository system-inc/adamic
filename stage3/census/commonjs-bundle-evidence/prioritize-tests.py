"""Temporarily free CPU for the bounded whole-package retry; always resume roots."""
import datetime,json,os,signal,time
from pathlib import Path
out=Path('/workspace/scratch/995spev');started=time.monotonic();pids=[]
for entry in Path('/proc').iterdir():
 if not entry.name.isdigit():continue
 try: command=(entry/'cmdline').read_bytes().split(b'\0')[0].decode()
 except (OSError,UnicodeError):continue
 if command in [str(out/'main-census'),str(out/'head-census')]:
  try:os.kill(int(entry.name),signal.SIGSTOP);pids.append(int(entry.name))
  except ProcessLookupError:pass
record=dict(started_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),pids=pids,reason='Give the pending 24-minute whole lower-package test the four-core CPU quota; first-stop measurements have no timing claim.',max_seconds=300)
(out/'measurement-pause.json').write_text(json.dumps(record,indent=2)+'\n')
def interrupted(signum,frame):raise KeyboardInterrupt
signal.signal(signal.SIGTERM,interrupted);signal.signal(signal.SIGINT,interrupted)
try:
 while time.monotonic()-started<300 and not (out/'lower-tests-warm-status.json').exists():time.sleep(1)
finally:
 for pid in pids:
  try:os.kill(pid,signal.SIGCONT)
  except ProcessLookupError:pass
 record.update(seconds=time.monotonic()-started,ended_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),resumed=True)
 (out/'measurement-pause.json').write_text(json.dumps(record,indent=2)+'\n')
 print(json.dumps(record),flush=True)

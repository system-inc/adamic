#!/usr/bin/env python3
"""Finish the independent paired recovery only after both measurements complete."""
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time

run = Path(__file__).resolve().parents[1]
repository = Path('/workspace/adamic')
scratch = Path('/tmp/adamic-gate/stage3-meter.RXmuND')
pin = '295df3bc6aa3d47933ccf675da1c33e745856313'
child, controller = 11038, 10615
while True:
    fields = Path(f'/proc/{child}/stat').read_text().split()
    if fields[2] == 'Z':
        break
    time.sleep(5)
assert len(fields) >= 52 and fields[1] == '(latent-census)'
raw_exit = int(fields[51])
assert os.waitstatus_to_exitcode(raw_exit) == 0, fields
sys.path.insert(0, str(repository/'stage3/meter'))
from report import latent_summary
main = latent_summary(scratch/'main-adapted', run/'main')
assert main['source_files'] == 79 and main['latent_mode'] == 'first-error'
completion = dict(exit=0, proc_wait_status=raw_exit, pid=child, totals=main['totals'], coverage='79 sources; validated by latent_summary', completed_utc=datetime.now(timezone.utc).isoformat())
(run/'validation/main-latent-completion.json').write_text(json.dumps(completion,indent=2)+'\n')
print('Main latent completed and coverage validated:',completion,flush=True)
# The paused controller would otherwise launch area a second time. Its child has
# finished; deliver TERM before continuing it. No compiler process is interrupted.
state = Path(f'/proc/{controller}/stat').read_text().split()[2]
assert state in ('T','t'), state
os.kill(controller,signal.SIGTERM)
os.kill(controller,signal.SIGCONT)
print('Retired paused controller after main latent exited 0; area continues independently.',flush=True)
while not (run/'validation/area-recovery.exit').exists():
    time.sleep(5)
assert (run/'validation/area-recovery.exit').read_text().strip() == '0'
with (run/'validation/report-generation.log').open('w') as log:
    subprocess.run(['python3',str(repository/'stage3/meter/report.py'),str(scratch/'main-adapted'),str(scratch/'area-adapted'),str(run),'20261019T233755Z',pin,pin],cwd=repository,stdout=log,stderr=log,check=True,timeout=300)
paths = [run/label/name for label in ['main','area'] for name in ['tsc/census.jsonl','tsc/latent.jsonl','census.jsonl','latent.jsonl']]
subprocess.run(['gzip', *map(str,paths)],check=True,timeout=300)
started = int((run/'validation/initial-overlay-failure/meter-start.epoch').read_text())
finished = int(time.time())
result = dict(exit=0, measured_commit=pin, started_epoch=started, finished_epoch=finished, elapsed_seconds=finished-started,
              completed_utc=datetime.now(timezone.utc).isoformat(), controller='independent paired recovery; original controller retired only after main child exit 0',
              ordinary_main_exit=int((run/'validation/main-ordinary.exit').read_text()), area_exit=0)
(run/'validation/measurement-completion.json').write_text(json.dumps(result,indent=2)+'\n')
print('Paired measurement complete:',json.dumps(result),flush=True)

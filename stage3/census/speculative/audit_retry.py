"""Prove four-process concurrency, reuse and the global launch deadline.
Usage: audit_retry.py OUTPUT
The synthetic child only witnesses process scheduling; compiler parity is audited
separately on the existing dependency project.
"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time
output=Path(sys.argv[1]).resolve();output.mkdir(exist_ok=True,parents=True)
source=output/'source';source.mkdir(exist_ok=True)
for i in range(4):(source/(str(i)+'.a')).write_text('// empty\n')
binary=output/'child.py'
binary.write_text('''#!/usr/bin/env python3
import fcntl,json,os,sys,time
from pathlib import Path
state=Path(os.environ['RETRY_WITNESS'])
def change(delta):
 with Path(str(state)+'.lock').open('w') as lock:
  fcntl.flock(lock,fcntl.LOCK_EX)
  data=json.loads(state.read_text()) if state.exists() else dict(active=0,peak=0,total=0)
  data['active']+=delta
  if delta>0:data['total']+=1
  data['peak']=max(data['peak'],data['active'])
  state.write_text(json.dumps(data))
change(1)
time.sleep(1)
source=Path(os.environ['LATENT_ONLY_FILE'])
Path(sys.argv[2]).write_text(json.dumps(dict(latent_mode='speculative'))+'\\n'+json.dumps(dict(file=str(source),findings=[],speculative_coverage=dict(source_bytes=source.stat().st_size,unvisited_nodes=0)))+'\\n')
change(-1)
''')
binary.chmod(0o755)
baseline=output/'baseline';(baseline/'records').mkdir(exist_ok=True,parents=True)
identity=dict(binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),root=str(source),
 files=[dict(file=p.name,bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()) for p in sorted(source.glob('*.a'))],
 seconds=10.0,rss_mib=256,project='whole compiler; only the census walk is filtered')
(baseline/'INPUT.json').write_text(json.dumps(identity))
def run(label,workers,deadline):
 witness=output/(label+'-state.json')
 target=output/label
 with (output/(label+'.log')).open('w') as log:
  subprocess.run([sys.executable,str(Path(__file__).with_name('retry_stream.py')),str(binary),str(source),str(baseline),str(target),'20','256',str(workers),str(deadline)],
                 env=dict(os.environ,RETRY_WITNESS=str(witness)),stdout=log,stderr=subprocess.STDOUT,timeout=30,check=True)
 return json.loads(witness.read_text()) if witness.exists() else dict(active=0,peak=0,total=0)
green=run('four',4,time.time()+25)
assert green['peak']==4 and green['total']==4 and green['active']==0,green
resume=run('four',4,time.time()+25)
assert resume==green,'resume launched completed files'
record=output/'four/records/0.a.jsonl'
checksum=record.with_suffix('.sha256');saved=checksum.read_bytes()
try:
    checksum.write_text('mutant\n')
    try:run('four',4,time.time()+25)
    except subprocess.CalledProcessError:print('corrupted completed checksum mutant caught before resume')
    else:raise AssertionError('checksum mutant survived')
finally:checksum.write_bytes(saved)
backup=record.with_suffix('.backup')
try:
    record.rename(backup)
    try:run('four',4,time.time()+25)
    except subprocess.CalledProcessError:print('dropped completed record mutant caught before resume')
    else:raise AssertionError('dropped record mutant survived')
finally:backup.rename(record)
paused=output/'paused';paused.mkdir(exist_ok=True);(paused/'PAUSE_QUEUE').touch()
assert run('paused',4,time.time()+25)['total']==0
mutant=run('serial-mutant',1,time.time()+25)
try:assert mutant['peak']==4
except AssertionError:print('one-worker mutant caught by independent live-process overlap witness')
else:raise AssertionError('serial mutant survived')
expired=run('expired',4,time.time()-1)
assert expired['total']==0 and not list((output/'expired/records').glob('*.jsonl'))
deadline_mutant=run('deadline-mutant',4,time.time()+25)
try:assert deadline_mutant['total']==0
except AssertionError:print('future-deadline mutant caught by independent child-launch witness')
else:raise AssertionError('deadline mutant survived')
print('PASS: four concurrent children; completed records reused with zero relaunches; expired deadline starts no children')

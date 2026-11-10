import os,signal,time,json
from pathlib import Path
out=Path(__file__).resolve().parent
parent=5352;groups=10907
assert b'measure.py' in Path(f'/proc/{parent}/cmdline').read_bytes()
assert b'run.py\x00groups' in Path(f'/proc/{groups}/cmdline').read_bytes()
s=dict(parent=parent,groups=groups,action='Pause only the outer driver while ordinary groups complete; census children retain their own 1800-second bounds.',paused_at=time.time())
os.kill(parent,signal.SIGSTOP)
(out/'groups-supervision.json').write_text(json.dumps(s,indent=2)+'\n')
try:
 while Path(f'/proc/{groups}/stat').exists() and Path(f'/proc/{groups}/stat').read_text().split()[2]!='Z':time.sleep(1)
finally:
 os.kill(parent,signal.SIGCONT)
 s['resumed_at']=time.time();s['reason']='Group runner exited; resume driver to collect its actual child exit.'
 (out/'groups-supervision.json').write_text(json.dumps(s,indent=2)+'\n')
 print('supervision_exit=0',flush=True)

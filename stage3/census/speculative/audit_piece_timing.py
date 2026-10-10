"""Prove the sampled prepass/body timestamp can detect a shifted timing tag.
Usage: audit_piece_timing.py OUTPUT
"""
from copy import deepcopy
import json
import os
from pathlib import Path
import subprocess
import sys
output=Path(sys.argv[1]).resolve()
output.mkdir(parents=True,exist_ok=True)
progress=output/'progress.json'
progress.unlink(missing_ok=True)
child="""import json,os,time
from pathlib import Path
p=Path(os.environ['LATENT_PROGRESS_FILE'])
p.write_text(json.dumps({'phase':'registration'}))
time.sleep(0.35)
p.write_text(json.dumps({'phase':'walk'}))
time.sleep(0.35)
"""
with (output/'supervisor.log').open('w') as log:
    subprocess.run([sys.executable,str(Path(__file__).with_name('timed_run.py')),'5','64',str(output/'metrics.json'),str(output/'child.log'),sys.executable,'-c',child],
                   env=dict(os.environ,LATENT_PROGRESS_FILE=str(progress)),stdout=log,stderr=subprocess.STDOUT,timeout=10,check=True)
metrics=json.loads((output/'metrics.json').read_text())
def check(m):
    phases=m['phase_first_wall_seconds']
    assert 0 <= phases['registration'] < phases['walk'] < m['wall_seconds']
    assert phases['walk'] >= .35
    assert m['user_cpu_seconds'] + m['system_cpu_seconds'] > 0
check(metrics)
for name,field in [('shifted-prepass-time','phase_first_wall_seconds'),('erased-CPU-time','cpu')]:
    mutant=deepcopy(metrics)
    if field=='cpu':mutant['user_cpu_seconds']=mutant['system_cpu_seconds']=0
    else:mutant[field]['walk']=0
    try:check(mutant)
    except AssertionError:print(name+' mutant caught by independent timed workload')
    else:raise AssertionError(name+' survived')
print('PASS: prepass/body timestamp and CPU usage match an independently timed workload')

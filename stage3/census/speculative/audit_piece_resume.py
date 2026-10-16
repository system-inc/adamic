"""Exercise piece resume and rejection guards with an independent launch witness.
Usage: audit_piece_resume.py ROOT PLAN COMPLETED_RUN OUTPUT
Only verified completed records are copied; the stand-in binary records any launch.
"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
root,plan,source,output=(Path(p).resolve() for p in sys.argv[1:5])
output.mkdir(parents=True,exist_ok=True)
launches=output/'launches.txt'
launches.unlink(missing_ok=True)
standin=output/'launch_witness.py'
standin.write_text('#!/usr/bin/env python3\nfrom pathlib import Path\np=Path('+repr(str(launches))+')\np.write_text("launched\\n")\nraise SystemExit(77)\n')
standin.chmod(0o755)
records=sorted((source/'records').glob('*.jsonl'))[:2]
assert len(records)==2
ids=[json.loads(p.read_text().splitlines()[1])['piece']['id'] for p in records]
identity=json.loads((source/'INPUT.json').read_text())
identity['binary_sha256']=hashlib.sha256(standin.read_bytes()).hexdigest()

def stage(name):
    target=output/name
    (target/'records').mkdir(parents=True,exist_ok=True)
    (target/'INPUT.json').write_text(json.dumps(identity,indent=2)+'\n')
    for path in records:
        shutil.copy2(path,target/'records'/path.name)
        shutil.copy2(path.with_suffix('.sha256'),target/'records'/path.with_suffix('.sha256').name)
        shutil.copy2(source/(path.stem+'.metrics.json'),target/(path.stem+'.metrics.json'))
    return target

def execute(target):
    with (target/'control.log').open('w') as log:
        code=subprocess.run([sys.executable,str(Path(__file__).with_name('run_pieces.py')),str(standin),str(root),str(plan),str(target),str(identity['seconds']),str(identity['rss_mib']),'32'],
                            env=dict(os.environ,LATENT_PIECES=','.join(map(str,ids))),stdout=log,stderr=subprocess.STDOUT,timeout=10).returncode
    return code,(target/'control.log').read_text()

positive=stage('resume')
before={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (positive/'records').iterdir()}
code,log=execute(positive)
assert code==0 and log.count('RESUME ')==2 and 'START ' not in log and not launches.exists()
after={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (positive/'records').iterdir()}
assert before==after
print('PASS: resume with 32 workers launches zero binaries and preserves completed records')
corrupt=stage('checksum-mutant')
(corrupt/'records'/records[0].with_suffix('.sha256').name).write_text('0'*64+'\n')
code,log=execute(corrupt)
assert code!=0 and 'piece checksum changed' in log and not launches.exists()
print('changed-checksum mutant caught before any compiler launch')
dropped=stage('dropped-mutant')
(dropped/'records'/records[0].name).unlink()
code,log=execute(dropped)
assert code!=0 and 'completed piece dropped' in log and not launches.exists()
print('dropped-completed-piece mutant caught before any compiler launch')

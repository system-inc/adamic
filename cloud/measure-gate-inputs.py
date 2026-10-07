#!/usr/bin/env python3
"""Interleave three fresh input installations with three validated warm checks."""
import importlib.util
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import time

source=Path(__file__).resolve().parent
repository,output=sys.argv[1:];output=Path(output);output.mkdir(parents=True,exist_ok=True)
scratch=Path(tempfile.mkdtemp(prefix='gate-input-timings-',dir='/workspace'))
node=str(Path(os.environ.get('ADAMIC_TOOLS','/opt/adamic-tools'))/'bin/node')

def command(args):return subprocess.check_output(args,cwd=repository,timeout=60).decode().strip()
def flags():return dict(commit=command(['git','rev-parse','HEAD']),nproc=command(['nproc']),cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),go=command(['go','version']),clang=command(['clang','--version']).splitlines()[0],node=command([node,'--version']),load=Path('/proc/loadavg').read_text().strip())
records=[]
for loop in range(1,4):
    cold=scratch/str(loop);warm=cold if loop!=2 else scratch/'1'
    for mode,root in ([('cold',cold),('warm',warm)] if loop!=2 else [('warm',warm),('cold',cold)]):
        for phase in ['npm','corpora','archive']:
            args=['python3',str(source/'setup-gate-inputs.py'),phase,repository,str(root),node]
            before=flags();started=time.monotonic();log=output/f'{loop}-{mode}-{phase}.log'
            with log.open('wb') as stream:
                result=subprocess.run(args,env=dict(os.environ,ADAMIC_GATE_UNCACHED='0'),stdout=stream,stderr=subprocess.STDOUT,timeout=1800)
            elapsed=time.monotonic()-started;after=flags();text=log.read_text();assert result.returncode==0,text
            steps=[dict(name=name,answer=answer,seconds=float(seconds)) for name,answer,seconds in re.findall(r'setup: gate input (.+?) (installed|generated|built|skipped).*?step-duration=([0-9.]+)s',text)]
            assert all((step['answer']=='skipped')==(mode=='warm') for step in steps),steps
            record=dict(loop=loop,mode=mode,phase=phase,seconds=elapsed,steps=steps,command=args,before=before,after=after,cached=mode=='warm')
            records.append(record);(output/'timings.json').write_text(json.dumps(records,indent=2)+'\n')
            with log.open('a') as stream:stream.write('build-flags '+json.dumps(record)+'\n')
            print(loop,mode,phase,f'{elapsed:.3f}s',flush=True)
print('retained fresh directories:',scratch,flush=True)

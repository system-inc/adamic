#!/usr/bin/env python3
"""Reproduce the fresh main control before prepare.py adds the last adapter."""
from pathlib import Path
import os,subprocess as s,time,json,shutil,signal
R=Path(__file__).resolve().parents[4]
D=Path(os.environ.get('CRYPTO_PROOF_SCRATCH','/workspace/scratch/wasm-crypto-proof'))
base='a0ba9dd52b23326880b16a7898d446597cfb5a88'
D.mkdir(parents=True,exist_ok=True);(D/'logs').mkdir(exist_ok=True)
(D/'base.json').write_text(json.dumps(dict(base=base),indent=2)+'\n')
assert not (D/'main').exists() and not (D/'control').exists(),'use a fresh scratch directory'
rows=[]
def run(name,cmd,cwd,limit):
 start=time.monotonic();minimum=shutil.disk_usage(D).free
 with (D/(name+'.log')).open('w') as f:
  child=s.Popen(cmd,cwd=cwd,stdout=f,stderr=s.STDOUT,start_new_session=True)
  while child.poll() is None:
   minimum=min(minimum,shutil.disk_usage(D).free)
   if time.monotonic()-start>limit or minimum<3*1024**3:
    os.killpg(child.pid,signal.SIGTERM);child.wait(timeout=10);raise RuntimeError('control time/disk bound')
   time.sleep(1)
 row=dict(name=name,command=cmd,cwd=str(cwd),exit=child.returncode,seconds=round(time.monotonic()-start,3),limit=limit,minFreeBytes=minimum)
 rows.append(row);(D/'control-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print(row,flush=True)
 assert child.returncode==0,row
run('control-source',['git','worktree','add','--detach',str(D/'main'),base],R,300)
run('api-install',['npm','ci','--prefix','stage3/api','--no-audit','--no-fund'],D/'main',600)
run('control-apply',['bash','stage3/apply.sh',str(D/'control')],D/'main',1800)

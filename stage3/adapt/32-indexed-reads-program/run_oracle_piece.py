#!/usr/bin/env python3
"""Run one bounded chunk of the unchanged lane manifest using its normal product path."""
from concurrent.futures import ThreadPoolExecutor, wait, FIRST_COMPLETED
import json
import os
import shutil
from pathlib import Path
import subprocess
import sys
import time
repo,artifact,results=map(lambda s:Path(s).resolve(),sys.argv[1:4]);chunk=int(sys.argv[4]);size=40
plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=plan['units'][chunk*size:(chunk+1)*size];assert units
product=results/'product'
if not product.exists():
    product.mkdir(parents=True)
    for p in (artifact/'tree').iterdir():
        if p.name!='node_modules':(product/p.name).symlink_to(p,target_is_directory=p.is_dir())
modules=repo/'stage3/typescript/node_modules';modules.parent.mkdir(exist_ok=True)
if not modules.exists():modules.symlink_to(artifact/'tree/node_modules',target_is_directory=True)
assert (repo/'stage3/typescript/package-lock.json').read_bytes()==(artifact/'tree/package-lock.json').read_bytes()
env=dict(os.environ,STAGE3_ADAPTED_TREE=str(product));env.pop('STAGE3_ARTIFACT',None)
chunkdir=results/('piece-'+str(chunk));chunkdir.mkdir(exist_ok=False)
started=time.monotonic();rows=[]
def run(unit):
    command=['python3',str(repo/'stage3/lane/shard.py'),unit['name'],str(results/'units')]
    with (chunkdir/(unit['name']+'.log')).open('w') as log:r=subprocess.run(command,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
    report=results/'units'/unit['name']/'report.json'
    result=dict(name=unit['name'],exit=r.returncode,report=json.loads(report.read_text()) if report.exists() else None)
    if r.returncode==0 and unit['name']!='stage3-lane-byte-comparisons':
        shutil.rmtree(results/'units'/unit['name']/'adapted-tree')
    return result
pending=iter(units);failed=False
with ThreadPoolExecutor(max_workers=2) as pool:
    active={pool.submit(run,next(pending)) for _ in range(min(2,len(units)))}
    while active:
        completed,active=wait(active,return_when=FIRST_COMPLETED)
        for f in completed:
            row=f.result();rows.append(row);failed|=row['exit']!=0
            print(json.dumps(dict(name=row['name'],exit=row['exit'],seconds=row['report'].get('seconds') if row['report'] else None)),flush=True)
        if not failed:
            for _ in completed:
                unit=next(pending,None)
                if unit:active.add(pool.submit(run,unit))
        (chunkdir/'results.json').write_text(json.dumps(dict(chunk=chunk,seconds=time.monotonic()-started,units=rows,failed=failed),indent=2)+'\n')
        if time.monotonic()-started>1400:raise RuntimeError('piece approaching 25-minute bound')
raise SystemExit(1 if failed else 0)

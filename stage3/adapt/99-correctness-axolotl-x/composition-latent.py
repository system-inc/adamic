#!/usr/bin/env python3
"""Measure exactly the six CW051 units in main and main plus 32 plus 99x."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time
p=Path(__file__).resolve().parent;repo=p.parents[2];scratch=Path('/workspace/scratch/bind2');out=p/'composition/latent';out.mkdir(parents=True,exist_ok=True)
allrows=[]
for label in ['main','32x']:
 receipt=p/'composition/matrix'/(label+'.json')
 while not receipt.exists():time.sleep(5)
 assert json.loads(receipt.read_text())['status']=='pass'
 tree=Path('/workspace/bind2-'+label+'-tree');plans=scratch/('latent-plans-'+label)
 env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
 subprocess.run(['node',str(p/'composition-latent-plan.cjs'),str(tree),str(plans)],cwd=repo,env=env,check=True)
 for index in range(6):
  plan=json.loads((plans/(str(index)+'.json')).read_text());raw=scratch/('latent-raw-'+label+'-'+str(index));raw.mkdir(exist_ok=True);records=raw/'records.jsonl'
  env=dict(os.environ,LATENT_FULL='0',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(raw/'times.jsonl'),ANY_MEMBERS_CENSUS_ROOT=str(tree),ANY_MEMBERS_CENSUS_PIECE=str(plans/(str(index)+'.json')))
  measured=scratch/('latent-'+label+'-'+str(index));subprocess.run(['python3',str(p/'measure_piece.py'),str(measured),'--',str(scratch/'latent-build/latent'),str(tree/'src/compiler'),str(records)],cwd=repo,env=env,check=True)
  rows=[json.loads(s) for s in records.read_text().splitlines()];assert rows[0]['checker_rejected'] is False
  files=[r for r in rows if 'units' in r];assert len(files)==1 and len(files[0]['units'])==1
  unit=files[0]['units'][0];norm=lambda s:s.removeprefix(str(tree)+'/');assert norm(unit['where'])==plan['expected_units'][0]
  findings=[f for f in files[0]['findings'] if f.get('phase')=='lowering'];assert len(findings)<=1 and all(norm(f['unit'])==plan['expected_units'][0] for f in findings)
  stop={k:findings[0].get(k) for k in ['kind','where','reason']} if findings else dict(kind=unit['status'],where=unit['where'],reason='no lowering finding');stop['where']=norm(stop['where'])
  metrics=json.loads((measured/'metrics.json').read_text());assert metrics['exit']==0
  summary=dict(label=label,index=index,reference=plan['reference'],name=plan['name'],actual=norm(unit['where']),checker_rejected=False,stop=stop,metrics=metrics,binary_sha256=hashlib.sha256((scratch/'latent-build/latent').read_bytes()).hexdigest());allrows.append(summary)
  for f in measured.iterdir():shutil.copyfile(f,raw/f.name)
  shutil.copyfile(plans/(str(index)+'.json'),raw/'selection.json');(raw/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
  archive=out/(label+'-'+str(index)+'.tar.gz');subprocess.run(['python3',str(p/'composition-pack.py'),str(raw),str(archive)],check=True)
  summaryfile=out/(label+'-'+str(index)+'.json');summaryfile.write_text(json.dumps(summary,indent=2)+'\n')
  subprocess.run(['python3',str(p/'publish_paths.py'),'99x CW051 '+label+' unit '+str(index),*[str(x.relative_to(repo)) for x in [archive,archive.with_suffix('').with_suffix('.sha256'),Path(str(archive)+'.sha256'),summaryfile]]],cwd=repo,check=True)
comparison=[]
for i in range(6):
 before=next(r for r in allrows if r['label']=='main' and r['index']==i);after=next(r for r in allrows if r['label']=='32x' and r['index']==i)
 comparison.append(dict(name=before['name'],reference=before['reference'],before=before['stop'],after=after['stop']))
summary=out/'summary.json';summary.write_text(json.dumps(dict(status='pass',pieces=12,comparison=comparison),indent=2)+'\n')
subprocess.run(['python3',str(p/'publish_paths.py'),'99x CW051 table for main plus 32 plus 99x',str(summary.relative_to(repo))],cwd=repo,check=True)

#!/usr/bin/env python3
"""Run the frozen whole lane and ten bounded pieces, publishing each receipt plainly."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import tarfile

repo,frozen,artifact,results=map(lambda s:Path(s).resolve(),sys.argv[1:])
unit=repo/'stage3/adapt/32-indexed-reads-program'
out=unit/'proof/maybebind/continuation'
out.mkdir(parents=True,exist_ok=True)

def run(command,cwd=repo):
    subprocess.run([str(s) for s in command],cwd=cwd,check=True)

def pack(source,name):
    files=sorted(f for f in source.rglob('*') if f.is_file() and 'adapted-tree' not in f.relative_to(source).parts and 'product' not in f.relative_to(source).parts)
    archive=out/(name+'.tar.gz')
    with tarfile.open(archive,'w:gz') as tar:
        for f in files:tar.add(f,arcname=f.relative_to(source),recursive=False)
    (out/(name+'-files.sha256')).write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+str(f.relative_to(source))+'\n' for f in files))
    (out/(name+'.sha256')).write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')

def publish(message):
    run(['git','add','--',str(unit.relative_to(repo))])
    names=subprocess.check_output(['git','diff','--cached','--name-only'],cwd=repo,text=True).splitlines()
    assert names and all(n.startswith('stage3/adapt/32-indexed-reads-program/') for n in names)
    run(['git','diff','--cached','--check'])
    run(['git','commit','-m',message])
    run(['git','push','origin','HEAD:refs/heads/typescript/32-maybebind-fold'])
    print(json.dumps(dict(published=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),message=message)),flush=True)

whole=results/'whole'
run(['python3',unit/'measure_piece.py',results/'whole-piece','--',frozen/'stage3/lane/run.sh',whole,'--artifact',artifact],cwd=frozen)
report=json.loads((whole/'report.json').read_text())
assert report['status']=='pass'
pack(whole,'whole')
(out/'whole-summary.json').write_text(json.dumps(dict(report=report,metrics=json.loads((results/'whole-piece/metrics.json').read_text())),indent=2)+'\n')
publish('Record the whole lane PASS for the twelve-site fold on main 12fec995')
shards=results/'shards'
for index in range(10):
    piece=results/('piece-'+str(index))
    run(['python3',unit/'measure_piece.py',piece,'--','python3',unit/'run_oracle_piece.py',frozen,artifact,shards,index])
    record=json.loads((shards/('piece-'+str(index))/'results.json').read_text())
    metrics=json.loads((piece/'metrics.json').read_text())
    assert not record['failed'] and metrics['exit']==0 and metrics['seconds']<1500
    summary=dict(piece=index,metrics=metrics,seconds=record['seconds'],units=[dict(name=r['name'],exit=r['exit'],seconds=r['report']['seconds'],status=r['report']['status'],counts=r['report']['counts']) for r in record['units']])
    assert all(r['exit']==0 and r['status']=='pass' for r in summary['units'])
    (out/('piece-'+str(index)+'.json')).write_text(json.dumps(summary,indent=2)+'\n')
    pack(shards,'shards')
    publish('Record passing oracle shard piece '+str(index)+' for the twelve-site fold')
run(['python3',unit/'check_oracle_union.py',frozen,shards,whole,out/'oracle-union.json'])
publish('Prove all 398 fold shards equal the whole lane by identity and API bytes')

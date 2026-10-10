"""Retry missing records with four independent single-CPU whole-project walks.
Usage: retry_stream.py BINARY ROOT BASELINE OUTPUT SECONDS RSS_MIB WORKERS DEADLINE_EPOCH
Completed baseline records are checksum verified before reuse. Input sources and
binary must be identical; only the retry time budget changes. All attempts are
journaled and constrained by the global deadline, including queued jobs.
"""
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import time

binary, root, baseline, output = (Path(x).resolve() for x in sys.argv[1:5])
seconds, rss, workers, deadline = float(sys.argv[5]),int(sys.argv[6]),int(sys.argv[7]),float(sys.argv[8])
assert workers > 0 and seconds > 0
old=json.loads((baseline/'INPUT.json').read_text())
assert old['binary_sha256']==hashlib.sha256(binary.read_bytes()).hexdigest()
assert old['root']==str(root)
for item in old['files']:
    source=root/item['file']
    assert source.stat().st_size==item['bytes'] and hashlib.sha256(source.read_bytes()).hexdigest()==item['sha256']
output.mkdir(parents=True,exist_ok=True)
(output/'records').mkdir(exist_ok=True)
identity=dict(old,seconds=seconds,rss_mib=rss)
manifest=output/'INPUT.json'
if manifest.exists():assert json.loads(manifest.read_text())==identity
else:manifest.write_text(json.dumps(identity,indent=2)+'\n')
(output/'INPUT.baseline.json').write_text(json.dumps(old,indent=2)+'\n')
for record in (baseline/'records').glob('*.jsonl'):
    assert record.with_suffix('.sha256').read_text().strip()==hashlib.sha256(record.read_bytes()).hexdigest()
    destination=output/'records'/record.name
    if not destination.exists():
        shutil.copyfile(record,destination)
        shutil.copyfile(record.with_suffix('.sha256'),destination.with_suffix('.sha256'))
        metrics=baseline/(record.stem+'.metrics.json')
        shutil.copyfile(metrics,output/metrics.name)
for metrics in baseline.glob('*.metrics.json'):
    value=json.loads(metrics.read_text())
    if value['exit']==0:
        assert (baseline/'records'/(metrics.name.removesuffix('.metrics.json')+'.jsonl')).exists(), 'completed baseline record dropped'
    destination=output/metrics.name
    if not destination.exists():shutil.copyfile(metrics,destination)
for record in (output/'records').glob('*.jsonl'):
    assert record.with_suffix('.sha256').read_text().strip()==hashlib.sha256(record.read_bytes()).hexdigest(), 'completed record checksum differs'
    rows=[json.loads(line) for line in record.read_text().splitlines()]
    assert len(rows)==2 and rows[1]['file'].startswith(str(root)+'/')
for metrics in output.glob('*.metrics.json'):
    if json.loads(metrics.read_text())['exit']==0:
        assert (output/'records'/(metrics.name.removesuffix('.metrics.json')+'.jsonl')).exists(), 'completed output record dropped'
remaining=[x for x in old['files'] if not (output/'records'/(x['file'].replace('/','__')+'.jsonl')).exists()]
selection=os.environ.get('LATENT_RETRY_FILES')
if selection:
    selected=set(selection.split(','))
    assert selected <= {x['file'] for x in old['files']}
    remaining=[x for x in remaining if x['file'] in selected]
attempts=[]
def measure(item):
    name=item['file'].replace('/','__')
    if time.time() >= deadline-3:return dict(file=item['file'],status='not started: unit deadline')
    started=time.time()
    print('start retry '+item['file'],flush=True)
    env=dict(os.environ,LATENT_RUN_ONLY=item['file'],LATENT_FILE_CPUS='1',LATENT_FILE_DEADLINE=str(deadline))
    with (output/(name+'.supervisor.log')).open('w') as log:
        result=subprocess.run([sys.executable,str(Path(__file__).with_name('stream.py')),str(binary),str(root),str(output),str(seconds),str(rss)],
                              env=env,stdout=log,stderr=subprocess.STDOUT,timeout=min(seconds+15,max(1,deadline-started+10)))
    metrics=output/(name+'.metrics.json')
    status=dict(file=item['file'],started_epoch=started,finished_epoch=time.time(),runner_exit=result.returncode,
                metrics=json.loads(metrics.read_text()) if metrics.exists() else None,
                completed=(output/'records'/(name+'.jsonl')).exists())
    print('finish retry '+item['file']+' completed='+str(status['completed']),flush=True)
    return status
# Largest first makes the long tail visible early.
with ThreadPoolExecutor(max_workers=workers) as pool:
    futures=[pool.submit(measure,x) for x in sorted(remaining,key=lambda x:-x['bytes'])]
    while futures:
        ready=[f for f in futures if f.done()]
        for f in ready:
            attempts.append(f.result());futures.remove(f)
            temporary=output/'ATTEMPTS.pending'
            temporary.write_text(json.dumps(dict(workers=workers,deadline_epoch=deadline,attempts=attempts),indent=2)+'\n')
            os.replace(temporary,output/'ATTEMPTS.json')
        time.sleep(.2)
print('retry finished '+str(sum(x.get('completed',False) for x in attempts))+'/'+str(len(remaining)),flush=True)

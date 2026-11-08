from pathlib import Path
import os,sys,json,shutil,tarfile,hashlib
sys.dont_write_bytecode=True
from run import run,s
repo=Path('/workspace/adamic');counts=s/'after-unit-counts';env=dict(os.environ,STAGE3_ADAPTED_TREE=str(s/'measure-after'),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(s/'pair-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',GFN_PAIR_DIR=str(counts));run('moved-real-six-unit',['python3',repo/'stage3/lane/shard.py','stage3-lane-file-293',s/'after-unit-results'],env=env)
rows=[json.loads(p.read_text()) for p in sorted(counts.glob('*.json'))];calls=[r for process in rows for r in process['records']];assert len(calls)==6 and all(r['equal'] for r in calls);assert all(r['castFailureCalls']==0 for r in rows);report=json.loads((s/'after-unit-results/stage3-lane-file-293/report.json').read_text());assert report['status']=='pass' and not report['errors'];(s/'raw/moved-real-six-unit.json').write_text(json.dumps({'passed':True,'unit':'stage3-lane-file-293','ordinaryGuardReport':report,'actualCalls':6,'counterexamples':0,'records':calls},indent=2)+'\n')
with tarfile.open(s/'raw/moved-real-six-unit.tar.gz','w:gz') as tf:
 for p in sorted(counts.glob('*')):tf.add(p,arcname='counts/'+p.name)
 for p in sorted((s/'after-unit-results/stage3-lane-file-293').rglob('*')):
  if p.is_file() and 'adapted-tree' not in p.parts:tf.add(p,arcname='ordinary/'+str(p.relative_to(s/'after-unit-results/stage3-lane-file-293')))
print('all six real moved deprecate calls return the original body names; ordinary unit guards pass')

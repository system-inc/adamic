import os,subprocess,sys,time,json,hashlib,signal,gzip
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];adapted=Path('/workspace/scratch/bar-adapted')
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_ASSERT_NO_OUTPUT='1',LATENT_RELATION_PER_UNIT='1',GOMEMLIMIT='5GiB',GOGC='70')
active={}
try:
 for side in ['base','head']:
  name=side+'-entry-run';runenv=dict(env,LATENT_TIMES=str(out/(name+'-units.jsonl')),LATENT_RELATION_TIMES=str(out/(name+'-relations.jsonl')))
  active[name]=subprocess.Popen([sys.executable,str(out/'run.py'),name,str(repo),'/workspace/scratch/bar-'+side+'-bin',str(adapted/'src/tsc/tsc.ts'),str(out/(name+'.jsonl'))],env=runenv,start_new_session=True)
 while active:
  for name,p in list(active.items()):
   if p.poll() is not None:
    rc=p.wait();print(name,'wrapper_exit='+str(rc),flush=True);del active[name]
    assert rc==0,(name,rc)
    for f in out.glob(name+'*'):
     if f.is_file() and f.stat().st_size>1000000:
      with f.open('rb') as src,gzip.open(str(f)+'.gz','wb') as dst:
       import shutil;shutil.copyfileobj(src,dst)
      f.unlink()
    rc=subprocess.run([sys.executable,str(out/'checkpoint.py'),name],cwd=repo).returncode
    assert rc==0,('checkpoint',rc)
  if active:time.sleep(1)
finally:
 for name,p in active.items():
  try:os.killpg(p.pid,signal.SIGTERM)
  except ProcessLookupError:pass
  try:p.wait(timeout=15)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
assert {str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((adapted/'src').rglob('*')) if p.is_file()}==json.loads((out/'source-hashes.json').read_text())
def stream(side):
 p=out/(side+'-entry-run.jsonl');return p if p.exists() else Path(str(p)+'.gz')
# The pinned pair tool accepts source locations, while the relation trailer has
# where=program. Preserve original streams and pair explicit source-only copies.
for side in ['base','head']:
 source=stream(side)
 with (gzip.open(source,'rt') if source.suffix=='.gz' else source.open()) as f:
  records=[json.loads(line) for line in f]
 assert records[-1]['status']=='relation_report'
 with gzip.open(out/(side+'-pair-input.jsonl.gz'),'wt') as target:
  for record in records[:-1]:
   for unit in record.get('units',[]):unit.pop('relation_overflow',None)
   target.write(json.dumps(record,separators=(',',':'))+'\n')
cmd=['python3',str(repo/'stage3/census/pair/pair.py'),str(out/'base-pair-input.jsonl.gz'),str(out/'head-pair-input.jsonl.gz'),str(adapted),'--typescript',str(repo/'stage3/api/node_modules/typescript'),'--json',str(out/'entry-paired.json'),'--report',str(out/'entry-paired.md')]
rc=subprocess.run([sys.executable,str(out/'run.py'),'entry-pair',str(repo),*cmd],env=env).returncode
print('pair wrapper_exit='+str(rc),flush=True);sys.exit(rc)

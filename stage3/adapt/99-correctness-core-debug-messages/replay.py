"""Replay every adapter, retaining exits, exact output, writes, and write contents."""
from pathlib import Path
import sys,os,json,hashlib,shutil,subprocess,signal,time,tarfile
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/debug-messages');u=repo/'stage3/adapt/99-correctness-core-debug-messages';side=sys.argv[1];tree=s/('replay-'+side);source=s/side;out=s/('replay-raw-'+side);out.mkdir();
shutil.copytree(source,tree,ignore=lambda p,n:[x for x in ['.git','built','node_modules','tests'] if Path(p)==source and x in n]);(tree/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
def manifest():return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
initial=manifest();(out/'initial-manifest.json').write_text(json.dumps(initial,sort_keys=True)+'\n');records=[]
for adapter in sorted((repo/'stage3/adapt').glob('*/adapt.cjs')):
 name=adapter.parent.name;log=out/(name+'.log');journal=out/(name+'.writes.jsonl');journal.write_text('');env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--require='+str(u/'replay-io.cjs'),BIND_REPLAY_ROOT=str(tree),BIND_REPLAY_JOURNAL=str(journal))
 begin=time.monotonic()
 with log.open('w') as f:
  p=subprocess.Popen(['node',str(adapter),str(tree)],cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
  try:
   try:code=p.wait(timeout=240)
   except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait();code=p.returncode
  finally:
   if p.poll() is None:os.killpg(p.pid,signal.SIGKILL);p.wait()
 text=log.read_text().replace(str(tree),'<tree>');events=[json.loads(l) for l in journal.read_text().splitlines()]
 records.append({'adapter':name,'exit':code,'stdout_stderr':text,'writes':events});(out/'records.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps({'side':side,'adapter':name,'child_exit':code,'seconds':time.monotonic()-begin,'write_calls':len(events)}),flush=True)
final=manifest();observed={e['file'] for r in records for e in r['writes']};changed={k for k in initial.keys()|final.keys() if initial.get(k)!=final.get(k)};assert not changed-observed
(out/'coverage.json').write_text(json.dumps({'changed_files':sorted(changed),'journaled_files':sorted(observed),'untracked_changes':[],'final_manifest':final},indent=2)+'\n')
evidence=s/'evidence';evidence.mkdir(exist_ok=True);archive=evidence/(side+'-replay.tar.gz')
with tarfile.open(archive,'w:gz') as tf:
 for p in sorted(out.iterdir()):tf.add(p,arcname=p.name)
assert archive.stat().st_size<20_000_000
print('COMPLETE',side,len(records),flush=True)

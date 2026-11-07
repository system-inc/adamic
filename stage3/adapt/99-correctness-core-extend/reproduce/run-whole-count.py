import os,time,subprocess,json,signal,sys
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic')
while True:
 control=s/'whole-control-result.json';rows=s/'paired-replay.json';ready=control.exists() and rows.exists() and len(json.loads(rows.read_text()))==sum(p.is_dir() for p in (repo/'stage3/adapt').iterdir())
 active=any(line.strip().split(None,1)[1]=='python3 /workspace/scratch/core-extend/replay.py' for line in subprocess.check_output(['ps','-eo','pid,args'],text=True).splitlines()[1:])
 if ready and not active:break
 print('waiting for control and replay before bounded whole-oracle count',flush=True);time.sleep(20)
assert json.loads(control.read_text())['exit']==0
with (s/'reconcile-replay.log').open('w') as log:subprocess.run(['python3',str(s/'reconcile-replay.py')],stdout=log,stderr=subprocess.STDOUT,check=True)
out=s/'full-count';out.mkdir();(out/'adapted-tree').symlink_to(s/'counted',target_is_directory=True);(out/'calls').mkdir()
env=dict(os.environ,COPY_CALL_DIR=str(out/'calls'),COPY_PIECE='whole-oracle',STAGE3_TASK_PROFILE=str(out/'tasks.jsonl'),NODE_OPTIONS='--require='+str(s/'preload.cjs')+' --require='+str(repo/'stage3/oracle/profile-tasks.cjs'),NODE_DISABLE_COMPILE_CACHE='1',PYTHONDONTWRITEBYTECODE='1')
start=time.monotonic()
with (out/'oracle-driver.log').open('w') as log:
 child=subprocess.Popen(['bash',str(repo/'stage3/oracle/run.sh'),str(s/'counted'),str(out/'oracle'),'--limit-seconds','1400'],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=child.wait(timeout=1500)
 except subprocess.TimeoutExpired:
  os.killpg(child.pid,signal.SIGKILL);child.wait();code=124
platform=json.loads(subprocess.check_output(['node','-p','JSON.stringify({os:process.platform,arch:process.arch,node:process.version})'],text=True));(out/'execution.json').write_text(json.dumps({'commit':json.loads((s/'recut-base.json').read_text())['base'],'platform':platform,'apply_exit':0,'oracle_exit':code,'wall_seconds':time.monotonic()-start,'measurement':'Full current-main apply copied exactly; only complete extend/copyProperties bodies wrapped for observation'},indent=2)+'\n')
(out/'apply.log').write_text((s/'apply-before.log').read_text());(out/'patch-set.md').write_bytes((s/'before/patch-set.md').read_bytes())
sys.path.insert(0,str(repo/'stage3/lane'));from check import check_results,write_verdict
report=check_results(out);result=write_verdict(out,report);(s/'full-count-result.json').write_text(json.dumps({'exit':result,'oracleExit':code,'seconds':time.monotonic()-start,'status':report['status']},indent=2)+'\n');print('bounded count',report['status'],flush=True);raise SystemExit(result)

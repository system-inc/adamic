import hashlib,json,os,signal,subprocess,time,sys
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/takewhile');raw=root/'raw';control=root/'whole-control';main=root/'main-repo';profile=raw/'whole-tasks.jsonl';begin=time.monotonic()
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(profile),NODE_OPTIONS='--require='+str(main/'stage3/oracle/profile-tasks.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
for key in ['TW_CALL_DIR','STAGE3_ADAPTED_TREE','STAGE3_ARTIFACT']: env.pop(key,None)
with (raw/'whole-control.log').open('w') as log:
    child=subprocess.Popen(['bash',str(main/'stage3/lane/run.sh'),str(control)],cwd=main,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    try: code=child.wait(timeout=1400)
    finally:
        rows=[row.split(None,2) for row in subprocess.check_output(['ps','-eo','pid,ppid,comm'],text=True).splitlines()[1:]];ids={child.pid}
        while True:
            expanded=ids|{int(pid) for pid,parent,_ in rows if int(parent) in ids}
            if expanded==ids: break
            ids=expanded
        for pid in sorted(ids,reverse=True):
            try: os.kill(pid,signal.SIGKILL)
            except ProcessLookupError: pass
        child.wait()
result={'exit':code,'seconds':time.monotonic()-begin,'counted':False,'command':'unchanged main stage3/lane/run.sh','processesCleaned':True}
(raw/'whole-control-result.json').write_text(json.dumps(result,indent=2)+'\n');assert code==0,result
sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import names_digest
observed=[dict(record,task=row['task'],status=status) for row in map(json.loads,profile.read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
counted=json.loads((raw/'count-observations.json').read_text());assert names_digest(observed)==names_digest(counted)
assert (raw/'count-baseline.diff').read_bytes()==(control/'oracle/baseline.diff').read_bytes()
# Bind the count to the fully applied main compiler used by this whole control.
manifest=lambda tree:{str(p.relative_to(tree/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
assert manifest(root/'before')==manifest(control/'adapted-tree')
result.update(identityMultisetIdentical=True,tests=len(observed),namesDigest=names_digest(observed),baselineDiffByteIdentical=True,allAppliedSourcesIdentical=True)
(raw/'whole-control-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

import hashlib,json,os,signal,subprocess,time,sys
from pathlib import Path
root=Path('/workspace/scratch/takewhile');raw=root/'raw'
for side in (sys.argv[1:] or ['before','after','mutant']):
    tree=root/side;file=tree/'src/compiler/core.ts'
    before=hashlib.sha256(file.read_bytes()).hexdigest();begin=time.monotonic()
    env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
    env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(file),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(raw/(side+'-unit-times.jsonl')),GOMAXPROCS='2',GOMEMLIMIT='5GiB')
    with (raw/(side+'-census.log')).open('w') as log:
        child=subprocess.Popen([str(root/'bin/latent'),str(tree/'src/compiler'),str(raw/(side+'-census.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try: code=child.wait(timeout=1400)
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    result={'exit':code,'seconds':time.monotonic()-begin,'sourceSha256':before,'sourceUnchanged':before==hashlib.sha256(file.read_bytes()).hexdigest(),'processGroupCleaned':True}
    (raw/(side+'-census-result.json')).write_text(json.dumps(result,indent=2)+'\n')
    print(side,json.dumps(result),flush=True)
    assert code==0 and result['sourceUnchanged']

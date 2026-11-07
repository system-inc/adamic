import json,os,signal,subprocess,time,sys
from pathlib import Path
root=Path('/workspace/scratch/takewhile');raw=root/'raw'
for side in (sys.argv[1:] or ['before','after']):
    tree=root/side;roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
    (raw/(side+'-roots.json')).write_text(json.dumps(roots,indent=2)+'\n');begin=time.monotonic()
    with (raw/('types-'+side+'.log')).open('w') as log:
        child=subprocess.Popen([str(root/'bin/adamic'),'types',*roots],env=dict(os.environ,GOMAXPROCS='2',GOMEMLIMIT='5GiB'),stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try: code=child.wait(timeout=900)
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    result={'roots':len(roots),'exit':code,'seconds':time.monotonic()-begin,'diagnosticsLogBytes':(raw/('types-'+side+'.log')).stat().st_size,'processGroupCleaned':True}
    (raw/('types-'+side+'.json')).write_text(json.dumps(result,indent=2)+'\n');assert code==0,result;print(side,json.dumps(result),flush=True)

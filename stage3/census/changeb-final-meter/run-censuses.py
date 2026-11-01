#!/usr/bin/env python3
"""Concurrent first-error runs; wait4 captures independent child resource usage."""
import datetime,json,os,pathlib,subprocess,time
out=pathlib.Path('/workspace/scratch/changeb-final-meter'); out.mkdir(exist_ok=True)
(out/'start-utc.txt').write_text(datetime.datetime.now(datetime.timezone.utc).isoformat()+'\n')
pending={}
for side in ('A','B'):
    env=dict(os.environ,LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/(side+'-times.jsonl')))
    command=['timeout','--kill-after=5s','5400','/workspace/scratch/census-'+side,'/workspace/scratch/tree/src/compiler',str(out/(side+'.jsonl'))]
    log=(out/(side+'.log')).open('w')
    start=time.monotonic()
    process=subprocess.Popen(command,env=env,stdout=log,stderr=subprocess.STDOUT)
    pending[process.pid]=(side,process,start,log,command)
(out/'pids.txt').write_text(json.dumps({side:pid for pid,(side,*_) in pending.items()})+'\n')
results={}
while pending:
    pid,status,usage=os.wait4(-1,0)
    side,process,start,log,command=pending.pop(pid)
    code=os.waitstatus_to_exitcode(status); process.returncode=code; log.close()
    result={'command':command,'exit_code':code,'wall_seconds':time.monotonic()-start,'peak_rss_kib':usage.ru_maxrss,'user_seconds':usage.ru_utime,'system_seconds':usage.ru_stime}
    (out/(side+'-resource.json')).write_text(json.dumps(result,indent=2)+'\n')
    (out/(side+'-exit.txt')).write_text(str(code)+'\n')
    results[side]=result
(out/'end-utc.txt').write_text(datetime.datetime.now(datetime.timezone.utc).isoformat()+'\n')
raise SystemExit(0 if all(r['exit_code']==0 for r in results.values()) else 1)

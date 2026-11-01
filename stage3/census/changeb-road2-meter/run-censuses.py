#!/usr/bin/env python3
"""Concurrent bounded census, with wait4 peak RSS and monotonic wall time."""
import datetime,json,os,pathlib,signal,subprocess,time
out=pathlib.Path('/workspace/scratch/paired')
out.mkdir(exist_ok=True)
def utc(): return datetime.datetime.now(datetime.timezone.utc).isoformat()
box={'hostname':os.uname().nodename,'nproc':subprocess.check_output(['nproc'],text=True).strip(),'memory':pathlib.Path('/proc/meminfo').read_text(),'cpu_max':pathlib.Path('/sys/fs/cgroup/cpu.max').read_text(),'loadavg_start':pathlib.Path('/proc/loadavg').read_text()}
(out/'box.json').write_text(json.dumps(box,indent=2)+'\n')
active={}
for side in ('A','B'):
    env=dict(os.environ,LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/f'{side}-times.jsonl'))
    cmd=[f'/workspace/scratch/census-{side}','/workspace/scratch/tree/src/compiler',str(out/f'{side}.jsonl')]
    log=(out/f'{side}.log').open('w')
    start=time.monotonic()
    p=subprocess.Popen(cmd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    active[p.pid]=dict(side=side,p=p,log=log,start=start,started_utc=utc(),command=cmd,timed_out=False,term=None)
while active:
    for pid,r in list(active.items()):
        done,status,usage=os.wait4(pid,os.WNOHANG)
        now=time.monotonic()
        if done:
            code=os.waitstatus_to_exitcode(status)
            result={k:r[k] for k in ('side','started_utc','command','timed_out')}
            result.update(exit_code=code,wall_seconds=now-r['start'],peak_rss_kib=usage.ru_maxrss,user_seconds=usage.ru_utime,system_seconds=usage.ru_stime,ended_utc=utc())
            (out/f"{r['side']}-resources.json").write_text(json.dumps(result,indent=2)+'\n')
            (out/f"{r['side']}-exit.txt").write_text(str(code)+'\n')
            r['p'].returncode=code
            r['log'].close()
            print(json.dumps(result),flush=True)
            del active[pid]
        elif not r['timed_out'] and now-r['start']>=5400:
            os.killpg(pid,signal.SIGTERM);r['timed_out']=True;r['term']=now
        elif r['timed_out'] and now-r['term']>=60:
            os.killpg(pid,signal.SIGKILL)
    if active: time.sleep(0.25)
box['loadavg_end']=pathlib.Path('/proc/loadavg').read_text()
(out/'box.json').write_text(json.dumps(box,indent=2)+'\n')

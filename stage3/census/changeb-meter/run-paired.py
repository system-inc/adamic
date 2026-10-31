#!/usr/bin/env python3
import datetime, hashlib, json, os, subprocess, time
from pathlib import Path
out=Path(__file__).resolve().parent
subprocess.run(['python3','/workspace/scratch/change-a/stage3/meter/box.py',str(out/'box.json'),'start'],check=True)
(out/'start.utc').write_text(datetime.datetime.now(datetime.timezone.utc).isoformat()+'\n')
processes={}
for side in ('A','B'):
 env=dict(os.environ,LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/f'{side}.times.jsonl'))
 log=open(out/f'{side}.log','w')
 started=time.monotonic()
 p=subprocess.Popen(['timeout','--kill-after=60s','5400',f'/workspace/scratch/census-{side.lower()}','/workspace/scratch/tree/src/compiler',str(out/f'{side}.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT)
 processes[p.pid]=(side,p,started,log)
exits={}
while processes:
 pid,status,usage=os.wait4(-1,0)
 side,p,started,log=processes.pop(pid)
 p.returncode=os.waitstatus_to_exitcode(status); log.close()
 exits[side]=p.returncode
 (out/f'{side}.resources.json').write_text(json.dumps(dict(exit_code=p.returncode,wall_seconds=time.monotonic()-started,peak_rss_kib=usage.ru_maxrss,user_seconds=usage.ru_utime,system_seconds=usage.ru_stime),indent=2)+'\n')
(out/'exit.json').write_text(json.dumps(exits,indent=2)+'\n')
(out/'end.utc').write_text(datetime.datetime.now(datetime.timezone.utc).isoformat()+'\n')
subprocess.run(['python3','/workspace/scratch/change-a/stage3/meter/box.py',str(out/'box.json'),'end'],check=True)

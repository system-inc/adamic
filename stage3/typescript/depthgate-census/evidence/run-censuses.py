#!/usr/bin/env python3
"""Run each canonical census separately; preserve raw stderr, enforce 30 minutes, wait4 RSS."""
import os,sys,time,json,subprocess,signal,threading,hashlib
from pathlib import Path
BASE='9d64615714e5beb154effd374bef22bcce5cfd8a'
HEAD='54210cd4c9b850b7c2df13b30f4406c78451ea9b'
ROOT=Path('/workspace/depthgate-runs')
TREE=Path('/workspace/depthgate-adapted')
RUNS=[('base','utilitiesPublic'),('head','utilitiesPublic'),('base','transformers-utilities'),('head','transformers-utilities'),('base','utilities'),('head','utilities')]
FILES={'utilitiesPublic':'src/compiler/utilitiesPublic.ts','utilities':'src/compiler/utilities.ts','transformers-utilities':'src/compiler/transformers/utilities.ts'}
ROOT.mkdir(exist_ok=True)
for label,name in RUNS:
 out=ROOT/(label+'-'+name)
 out.mkdir(exist_ok=False)
 binary=Path('/workspace/depthgate-tools-'+label+'/entry-census')
 selected=TREE/FILES[name]
 env={k:v for k,v in os.environ.items() if not k.startswith(('LATENT_','ADAMIC_PROBE_'))}
 env.update(CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TAIL='1',LATENT_ONLY_FILE=str(selected),LATENT_TIMES='/dev/stderr')
 if label=='head':env['ADAMIC_PROBE_DEPTHGATE']='1'
 cmd=[str(binary),str(TREE/'src/tsc/tsc.ts'),str(out/'latent.jsonl')]
 meta={'loadavg_start':list(os.getloadavg()),'cpu_quota':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'memory_limit_bytes':Path('/sys/fs/cgroup/memory.max').read_text().strip(),'label':label,'name':name,'sha':BASE if label=='base' else HEAD,'command':cmd,'cwd':'/workspace/depthgate-'+label,'environment':{k:env[k] for k in env if k.startswith(('LATENT_','CENSUS_MODULE_FORMAT','ADAMIC_PROBE_'))},'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'selected_source_sha256':hashlib.sha256(selected.read_bytes()).hexdigest(),'timeout_seconds':1800,'started_utc':time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime())}
 (out/'started.json').write_text(json.dumps(meta,indent=2)+'\n')
 print('START',label,name,meta['started_utc'],flush=True)
 start=time.monotonic()
 with (out/'stdout.log').open('wb') as stdout,(out/'stderr.log').open('wb',buffering=0) as stderr:
  p=subprocess.Popen(cmd,cwd=meta['cwd'],env=env,stdout=stdout,stderr=subprocess.PIPE,start_new_session=True)
  def pump():
   while True:
    data=p.stderr.read1(65536)
    if not data:break
    stderr.write(data)
  thread=threading.Thread(target=pump);thread.start()
  timeout=False;term=None;killed=False
  while True:
   pid,status,usage=os.wait4(p.pid,os.WNOHANG)
   if pid:break
   elapsed=time.monotonic()-start
   if elapsed>=1800 and term is None:
    timeout=True;term=time.monotonic();os.killpg(p.pid,signal.SIGTERM)
   elif term is not None and time.monotonic()-term>=5 and not killed:
    killed=True;os.killpg(p.pid,signal.SIGKILL)
   time.sleep(.2)
  p.returncode=os.waitstatus_to_exitcode(status);thread.join();p.stderr.close()
 meta['loadavg_end']=list(os.getloadavg())
 meta.update(wall_seconds=time.monotonic()-start,exit_code=p.returncode,timed_out=timeout,sigkill_required=killed,peak_rss_kib=usage.ru_maxrss,user_cpu_seconds=usage.ru_utime,system_cpu_seconds=usage.ru_stime,finished_utc=time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()))
 (out/'status.json').write_text(json.dumps(meta,indent=2)+'\n')
 print('DONE',label,name,'exit',p.returncode,'timeout',timeout,'wall',round(meta['wall_seconds'],3),'peak_RSS_KiB',usage.ru_maxrss,flush=True)

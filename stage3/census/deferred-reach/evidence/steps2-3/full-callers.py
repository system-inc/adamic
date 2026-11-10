import json,os,subprocess,time,signal
from pathlib import Path
w=Path('/workspace/scratch/abrq61p');rs=json.loads((w/'classification-draft.json').read_text());files=sorted({'src/compiler/core.ts'}|{c['unit'].rsplit(':',2)[0] for r in rs if r['where'].startswith('src/compiler/core.ts:') and r['classification'] and r['classification'].startswith('unreached hidden') for c in r['census_callers']});selected='\n'.join('/workspace/scratch/abrq61p-adapted/'+f for f in files);(w/'full-core-callers-files.json').write_text(json.dumps(files,indent=2)+'\n')
env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=selected,LATENT_TIMES=str(w/'full-core-callers-times.jsonl'))
start=time.monotonic();last=start;size=None;peak=0;reason=None
with (w/'full-core-callers.log').open('w') as log:
 p=subprocess.Popen([str(w/'head-census'),'/workspace/scratch/abrq61p-adapted/src/compiler',str(w/'full-core-callers.jsonl')],cwd='/workspace/adamic',env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:
  while p.poll() is None:
   current=int(Path('/sys/fs/cgroup/memory.current').read_text());peak=max(peak,current)
   now=tuple((w/n).stat().st_size if (w/n).exists() else 0 for n in ['full-core-callers.log','full-core-callers-times.jsonl'])
   if now!=size:size=now;last=time.monotonic()
   if current>14*1024**3:reason='cgroup memory guard at 14 GiB'
   elif time.monotonic()-last>1800:reason='30 minutes without log or per-unit timing progress'
   elif time.monotonic()-start>1200:reason='20 minute bounded full-caller experiment budget'
   if reason:
    os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
    break
   time.sleep(1)
 finally:
  if p.poll() is None:os.killpg(p.pid,signal.SIGKILL);p.wait()
(w/'full-core-callers.exit').write_text(str(p.returncode)+'\n');result=dict(child_exit=p.returncode,stop_reason=reason,peak_cgroup_bytes=peak,seconds=time.monotonic()-start,selected_files=files,complete=p.returncode==0);(w/'full-core-callers-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

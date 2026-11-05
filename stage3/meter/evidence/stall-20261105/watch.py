import os, subprocess, time, json, signal
from pathlib import Path
root=Path('/tmp/meter-stall-20261105')
while not (root/'entry-census').exists() or Path('/proc/833').exists(): time.sleep(5)
out=root/'latent.jsonl'; log=root/'latent.log'
env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(root/'unit-times.jsonl'),LATENT_PROGRESS_FILE=str(root/'progress.json'))
with log.open('wb') as f:
 p=subprocess.Popen([str(root/'entry-census'),str(root/'adapted/src/tsc/tsc.ts'),str(out)],stdout=f,stderr=f,env=env,start_new_session=True)
 (root/'census.pid').write_text(str(p.pid))
 start=last=time.monotonic(); sizes=None
 while p.poll() is None:
  current=tuple(x.stat().st_size if x.exists() else 0 for x in [out,log])
  if current!=sizes: last=time.monotonic(); sizes=current
  if time.monotonic()-last>=900:
   with (root/'proc-stat.txt').open('w') as stats:
    for _ in range(4):
     stats.write(str(time.time())+' '+Path(f'/proc/{p.pid}/stat').read_text()+'\n'); stats.flush(); time.sleep(2)
   p.send_signal(signal.SIGQUIT)
   try:p.wait(timeout=30)
   except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
   break
  time.sleep(1)
 (root/'full-timing.json').write_text(json.dumps(dict(pid=p.pid,seconds=time.monotonic()-start,exit=p.returncode,no_output_seconds=time.monotonic()-last))+'\n')

from pathlib import Path
import os,json,subprocess,time,signal,datetime
r=Path('/tmp/meter-stall-20261105');d=r/'slowdown'
for label,sha in [('good','f87d7884807161162c98a2d0ef3b5339e16b315f'),('bad','1f7df6a2e8145391dea19da63a172d5385638e39')]:
 out=d/label;out.mkdir(exist_ok=True)
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/'times.jsonl'),METER_CPU_PROFILE=str(out/'cpu.pprof'))
 started=time.monotonic();utc=datetime.datetime.now(datetime.timezone.utc).isoformat()
 with (out/'census.log').open('wb') as log:
  p=subprocess.Popen([str(d/(label+'-census')),str(r/'adapted/src/tsc/tsc.ts'),str(out/'latent.jsonl')],env=env,stdout=log,stderr=log,start_new_session=True)
  (out/'pid').write_text(str(p.pid))
  (out/'start.json').write_text(json.dumps(dict(commit=sha,pid=p.pid,started_utc=utc,budget_seconds=1800))+'\n')
  try:p.wait(timeout=1800);deadline=False
  except subprocess.TimeoutExpired:
   deadline=True;p.send_signal(signal.SIGQUIT)
   try:p.wait(timeout=20)
   except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
 (out/'result.json').write_text(json.dumps(dict(commit=sha,started_utc=utc,finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),wall_seconds=time.monotonic()-started,deadline=deadline,exit=p.returncode))+'\n')

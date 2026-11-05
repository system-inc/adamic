from pathlib import Path
import json,os,subprocess,time,sys
r=Path('/tmp/meter-stall-20261105');s=r/'slowdown';settings=json.loads((s/'predicate.json').read_text())
for sha in ['b2b32fb8914d28e5ae911ec1dd8a475933f2b10c','4a8171994b6a63a42962ddaa462aa28d8e9caa3d']:
 base=s/'unit-tests'/sha/'confirm-1';d=base.parent/'confirm-2';d.mkdir(exist_ok=True)
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',METER_UNIT_CPU_PROFILE=str(d/'cpu.pprof'))
 started=time.monotonic()
 with (d/'record.jsonl').open('wb') as out,(d/'run.log').open('wb') as log:
  result=subprocess.run(['timeout','--signal=QUIT','--kill-after=20s',str(settings['hard_timeout_seconds'])+'s',str(base/'unit'),str(r/'adapted/src/tsc/tsc.ts'),str(r/'adapted'/settings['unit']),str(d/'times.jsonl')],env=env,stdout=out,stderr=log)
 row=dict(commit=sha,label='confirm-2',reused_binary=str(base/'unit'),run_seconds=time.monotonic()-started,exit=result.returncode,threshold_seconds=settings['threshold_seconds'],hard_timeout_seconds=settings['hard_timeout_seconds'])
 if result.returncode!=0:raise SystemExit('confirmation did not complete: '+str(row))
 a=[json.loads(x) for x in (d/'times.jsonl').read_text().splitlines()];assert len(a)==1
 record=json.loads((d/'record.jsonl').read_text());assert len(record['units'])==1
 row['unit_time']=a[0];row['bisect_exit']=int(a[0]['wall_seconds']>settings['threshold_seconds']);(d/'result.json').write_text(json.dumps(row,indent=2)+'\n')
 with (s/'test-timings.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(json.dumps(row),flush=True)

from pathlib import Path
import json,subprocess,os,time,math,sys
r=Path('/tmp/meter-stall-20261105');s=r/'slowdown';source=r/'source';settings=json.loads((s/'predicate.json').read_text())
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=source,text=True).strip();label=sys.argv[1] if len(sys.argv)>1 else 'bisect'
d=s/'unit-tests'/sha/label;d.mkdir(parents=True,exist_ok=True)
env=dict(os.environ,PATH='/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:'+os.environ['PATH'],TMPDIR='/tmp/adamic-gate',GOTOOLCHAIN='auto')
started=time.monotonic();row={'commit':sha,'label':label,'threshold_seconds':settings['threshold_seconds'],'hard_timeout_seconds':3*settings['bad_seconds']};code=125
try:
 with (d/'build.log').open('wb') as log:
  for cmd in [
   ['python3','stage3/census/latent/make_overlay.py',str(source),str(d/'latent')],
   ['python3','stage3/meter/entry_overlay.py',str(d/'latent'),str(d/'entry')],
   ['python3',str(s/'prepare-unit.py'),str(d/'observed'),str(d/'entry')]]:
   subprocess.run(cmd,cwd=source,env=env,stdout=log,stderr=log,check=True,timeout=300)
  m=json.loads((d/'observed/overlay.json').read_text())
  k=next(k for k in m['Replace'] if k.endswith('/stage3/census/latent/tool/main.go'));m['Replace'][k]=str(s/'unit-main.go.txt');(d/'unit-overlay.json').write_text(json.dumps(m))
  subprocess.run(['go','build','-buildvcs=false','-overlay='+str(d/'unit-overlay.json'),'-o',str(d/'unit'),'./stage3/census/latent/tool'],cwd=source,env=env,stdout=log,stderr=log,check=True,timeout=600)
 row['build_seconds']=time.monotonic()-started;t=time.monotonic()
 runenv=dict(env,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1')
 if label.startswith('confirm'):runenv['METER_UNIT_CPU_PROFILE']=str(d/'cpu.pprof')
 with (d/'record.jsonl').open('wb') as out,(d/'run.log').open('wb') as log:
  result=subprocess.run(['timeout','--signal=QUIT','--kill-after=20s',str(row['hard_timeout_seconds'])+'s',str(d/'unit'),str(r/'adapted/src/tsc/tsc.ts'),str(r/'adapted'/settings['unit']),str(d/'times.jsonl')],env=runenv,stdout=out,stderr=log)
 row['run_seconds']=time.monotonic()-t;row['exit']=result.returncode
 if result.returncode==0:
  a=[json.loads(x) for x in (d/'times.jsonl').read_text().splitlines()];assert len(a)==1,a
  record=json.loads((d/'record.jsonl').read_text());assert len(record['units'])==1
  row['unit_time']=a[0];code=int(a[0]['wall_seconds']>settings['threshold_seconds'])
 elif result.returncode in [124,137]:code=1
except (subprocess.CalledProcessError,subprocess.TimeoutExpired,AssertionError,ValueError) as e:row['error']=str(e)
row['bisect_exit']=code;(d/'result.json').write_text(json.dumps(row,indent=2)+'\n')
with (s/'test-timings.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
print(json.dumps(row),flush=True);sys.exit(code)

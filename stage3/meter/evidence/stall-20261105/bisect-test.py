import subprocess, time, os, json, pathlib, sys
r=pathlib.Path('/tmp/meter-stall-20261105'); source=r/'source'
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=source,text=True).strip()
d=r/'bisect'/sha;d.mkdir(parents=True,exist_ok=True)
start=time.monotonic(); env=dict(os.environ,PATH='/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:'+os.environ['PATH'],GOTOOLCHAIN='auto',TMPDIR='/tmp/adamic-gate')
steps=[['python3','stage3/census/latent/make_overlay.py',str(source),str(d/'latent')],['python3','stage3/meter/entry_overlay.py',str(d/'latent'),str(d/'entry')]]
code=125
with (d/'build.log').open('wb') as log:
 try:
  for cmd in steps:subprocess.run(cmd,cwd=source,env=env,stdout=log,stderr=log,check=True,timeout=300)
  m=json.loads((d/'entry/overlay.json').read_text())
  for k in m['Replace']:
   if k.endswith('/stage3/census/latent/tool/main.go'):m['Replace'][k]=str(r/'unit-main.go.txt')
  (d/'unit-overlay.json').write_text(json.dumps(m))
  subprocess.run(['go','build','-buildvcs=false','-overlay='+str(d/'unit-overlay.json'),'-o',str(d/'unit'),'./stage3/census/latent/tool'],cwd=source,env=env,stdout=log,stderr=log,check=True,timeout=600)
  build=time.monotonic()-start; teststart=time.monotonic()
  with (d/'stdout.jsonl').open('wb') as out,(d/'stderr.log').open('wb') as err:
   result=subprocess.run(['timeout','--signal=QUIT','--kill-after=15s','180s',str(d/'unit'),str(r/'adapted/src/tsc/tsc.ts'),str(r/'adapted/src/compiler/tracing.ts:218:5')],env=dict(env,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1'),stdout=out,stderr=err)
  test=time.monotonic()-teststart
  code=0 if result.returncode==0 else 1 if result.returncode in [124,137] else 125
  row=dict(commit=sha,build_seconds=build,test_seconds=test,exit=result.returncode,bisect_exit=code)
 except (subprocess.CalledProcessError,subprocess.TimeoutExpired) as e:row=dict(commit=sha,build_seconds=time.monotonic()-start,error=str(e),bisect_exit=125)
with (r/'per-commit-timings.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
print(json.dumps(row),flush=True)
sys.exit(code)

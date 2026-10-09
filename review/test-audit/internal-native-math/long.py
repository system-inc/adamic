import os,subprocess,json,pathlib,time,signal
out=pathlib.Path('review/test-audit/internal-native-math');results=[]
for mid in ['P09','M12']:
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_NORMALIZE_BENCH_LOG']=str(out.resolve()/('long-'+mid+'.jsonl'));env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u049/cache/long-'+mid
 cmd=['go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestNormalizeLongMeasurements$'];t=time.monotonic()
 with open(out/('long-'+mid+'.log'),'w') as log:
  p=subprocess.Popen(cmd,stdout=log,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  try:code=p.wait(timeout=95)
  except subprocess.TimeoutExpired:code=124
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 results.append(dict(id=mid,command='ADAMIC_MUTANT='+mid+' ADAMIC_NORMALIZE_BENCH_LOG='+env['ADAMIC_NORMALIZE_BENCH_LOG']+' '+' '.join(cmd),exit=code,wall=time.monotonic()-t));(out/'long.json').write_text(json.dumps(results,indent=2));print(mid,code,flush=True)

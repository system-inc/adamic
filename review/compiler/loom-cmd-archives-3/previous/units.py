import json, os, pathlib, resource, subprocess, sys, tempfile, time
out=pathlib.Path(__file__).resolve().parent
binary,mode=sys.argv[1:]
rows=[]
for test in ['TestExplainChecksOutput','TestNonNullExplainChecks','TestRequestNativeStaysCommand','TestWASIRequest','TestWASIRequestThrows']:
 env=dict(os.environ,ADAMIC_BUILD_STORE='off',ADAMIC_ORACLE_WASI='1')
 if mode=='before':
  env.pop('ADAMIC_BUILD_CACHE_DIR',None); env.pop('ADAMIC_BUILD_CACHE',None)
  env['XDG_CACHE_HOME']=tempfile.mkdtemp(prefix='adamic-unit-audit-')
 else: env['ADAMIC_BUILD_CACHE']='read'
 usage=resource.getrusage(resource.RUSAGE_CHILDREN); start=time.monotonic()
 with (out/f'units-{mode}-{test}.log').open('w') as log:
  result=subprocess.run(['timeout','60',binary,'-test.timeout=55s','-test.v','-test.run=^'+test+'$'],cwd='cmd/adamic',env=env,stdout=log,stderr=subprocess.STDOUT)
 after=resource.getrusage(resource.RUSAGE_CHILDREN)
 row=dict(test=test,seconds=time.monotonic()-start,core_seconds=after.ru_utime+after.ru_stime-usage.ru_utime-usage.ru_stime,exit=result.returncode)
 rows.append(row);print(json.dumps(row),flush=True)
 (out/f'units-{mode}.json').write_text(json.dumps(rows,indent=2)+'\n')
 if result.returncode:sys.exit(result.returncode)

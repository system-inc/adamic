import json, os, pathlib, resource, subprocess, sys, tempfile, time
root = pathlib.Path(__file__).resolve().parent
checks = ['overload_some_empty','overload_some_false_read','overload_some_true_only','overload_erased','overload_assertion']
nonnull = ['narrowed','present','undefined','null','initializer','let_initializer','refuse_undefined','refuse_null']
tests = ['TestExplainChecksOutput/'+x+'/build' for x in checks] + ['TestNonNullExplainChecks/'+x+'/build' for x in nonnull] + ['TestRequestNativeStaysCommand','TestWASIRequest/request.a','TestWASIRequest/request_alias.a','TestWASIRequestThrows']
binary, mode = sys.argv[1:]
if os.environ.get('AUDIT_ONLY'):
    tests=[test for test in tests if test==os.environ['AUDIT_ONLY']]
rows=[]
for index, test in enumerate(tests):
    cache=tempfile.mkdtemp(prefix='adamic-audit-') if mode.startswith('before') else os.environ['ADAMIC_BUILD_CACHE_DIR']
    for temperature in ['cold','warm'] if mode.startswith('before') else ['ahead']:
        env=dict(os.environ, ADAMIC_BUILD_CACHE_DIR=cache, ADAMIC_BUILD_STORE='off', ADAMIC_ORACLE_WASI='1')
        if not mode.startswith('before'): env['ADAMIC_BUILD_CACHE']='read'
        pattern='/'.join('^'+part.replace('.',r'\.')+'$' for part in test.split('/'))
        usage=resource.getrusage(resource.RUSAGE_CHILDREN); start=time.monotonic()
        with (root/f'{mode}-{index}-{temperature}.log').open('w') as log:
            result=subprocess.run(['timeout','60',binary,'-test.timeout=55s','-test.v','-test.run='+pattern],cwd='cmd/adamic',env=env,stdout=log,stderr=subprocess.STDOUT)
        after=resource.getrusage(resource.RUSAGE_CHILDREN)
        row=dict(test=test,temperature=temperature,seconds=time.monotonic()-start,core_seconds=after.ru_utime+after.ru_stime-usage.ru_utime-usage.ru_stime,exit=result.returncode)
        rows.append(row); print(json.dumps(row),flush=True)
        (root/f'{mode}.json').write_text(json.dumps(rows,indent=2)+'\n')
        if result.returncode: sys.exit(result.returncode)

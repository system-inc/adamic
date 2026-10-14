import json, os, pathlib, subprocess, time
out=pathlib.Path(__file__).resolve().parent
binary='/tmp/adamic-after.test'
listed=subprocess.run([binary,'-test.list=^Test'],capture_output=True,text=True,check=True).stdout.splitlines()
skip={'TestExplainChecksOutput','TestNonNullExplainChecks','TestRequestNativeStaysCommand','TestWASIRequest','TestWASIRequestThrows'}
rows=[]
for test in listed:
 if not test.startswith('Test') or test in skip or test.startswith('TestProduct_'):continue
 start=time.monotonic()
 with (out/('verify-'+test+'.log')).open('w') as log:
  result=subprocess.run(['timeout','60',binary,'-test.timeout=55s','-test.v','-test.run=^'+test+'$'],cwd='cmd/adamic',stdout=log,stderr=subprocess.STDOUT)
 rows.append(dict(test=test,seconds=time.monotonic()-start,exit=result.returncode))
 if result.returncode:raise RuntimeError(rows[-1])
(out/'remaining.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))

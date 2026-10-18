import concurrent.futures,json,os,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-v4-2';work=Path('/workspace/scratch/loom-v4-2-validation')
cases=[('stage1/cohere/markdowninline','TestMarkdownInline_000'+str(i)) for i in range(2)]+[('stage1/typescript/scanner','TestScannerAgreesWithTypescriptGo_00'+str(i)) for i in range(3)]
def run(case):
 package,test=case;start=time.monotonic();log=e/(package.replace('/','-')+'-'+test+'-clean.log')
 with log.open('w') as out:
  try:r=subprocess.run([str(e/(package.replace('/','-')+'-consumer.test')),'-test.run','^'+test+'$','-test.timeout=90s','-test.parallel=1','-test.v'],cwd=work/package,stdout=out,stderr=subprocess.STDOUT,timeout=95,env={**os.environ,'ADAMIC_TYPESCRIPT_SOURCE':'/workspace/scratch/loom-v4-2-typescript'});code=r.returncode
  except subprocess.TimeoutExpired:code=124
 seconds=round(time.monotonic()-start,3);row=dict(package=package,test=test,seconds=seconds,exit=code,grain=seconds<60);print(row,flush=True);return row
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:rows=list(pool.map(run,cases))
(e/'clean-consumer-results.json').write_text(json.dumps(rows,indent=2)+'\n')

import hashlib,json,os,subprocess,time
from pathlib import Path
e=Path(__file__).resolve().parent;art=e/'artifacts';art.mkdir(exist_ok=True)
env={**os.environ,'GOMAXPROCS':'4'}
rows=[]
def command(args,cwd,log,limit,extra=None):
 start=time.monotonic()
 with (e/log).open('w') as out:r=subprocess.run(args,cwd=cwd,env={**env,**(extra or {})},stdout=out,stderr=subprocess.STDOUT,timeout=limit)
 if r.returncode:raise RuntimeError(str(args)+' failed: '+str(r.returncode))
 return time.monotonic()-start
# Export corpus after full-package runs, restoring the exact checkout immediately.
root=Path('/tmp/adamic-json-82');probe=root/'stage1/cohere/json/diagnostic_v4_export_test.go';assert not probe.exists()
try:
 probe.write_bytes((e/'export-corpus.go.txt').read_bytes())
 command(['go','test','./stage1/cohere/json','-run','^TestDiagnosticExportJSONCorpus$','-count=1','-timeout','90s','-json'],root,'export.jsonl',120,{'V4_JSON_MEASURE':str(art)})
finally:probe.unlink()
# Build the same compiler-driver source separately against each exact revision.
for side,tree in [('base','/tmp/adamic-v4-2-base'),('v4','/tmp/adamic-json-82')]:
 root=Path(tree);temp=root/'.v4-json-measure';temp.mkdir(exist_ok=False)
 try:
  (temp/'main.go').write_bytes((e/'measure-port.go.txt').read_bytes())
  driver=art/(side+'-measure')
  command(['go','build','-o',str(driver),'./.v4-json-measure'],root,side+'-measure-build.log',180)
 finally:(temp/'main.go').unlink();temp.rmdir()
 command([str(driver),str(art/side)],root,side+'-compile.log',240)
# One source-Node observation, then three interleaved rounds of release and sanitizer binaries.
inputs=art/'cases.txt';measurements=[]
def run(side,mode,round_,args,cwd,sanitize=False):
 output=art/(side+'-'+mode+'.stdout');error=art/(side+'-'+mode+'.stderr');start=time.monotonic()
 with output.open('wb') as out,error.open('wb') as err:
  r=subprocess.run(args,cwd=cwd,env={**env,**({'ASAN_OPTIONS':'detect_leaks=1'} if sanitize else {})},stdout=out,stderr=err,timeout=180)
 seconds=time.monotonic()-start
 row=dict(side=side,mode=mode,round=round_,seconds=seconds,exit=r.returncode,stdout_bytes=output.stat().st_size,stdout_sha256=hashlib.sha256(output.read_bytes()).hexdigest(),stderr_bytes=error.stat().st_size)
 measurements.append(row);(e/'run-times.json').write_text(json.dumps(measurements,indent=2)+'\n');print(row,flush=True)
 if r.returncode or error.stat().st_size:raise RuntimeError('port failed '+str(row))
run('source','node',0,['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','stage1/cohere/json/main.ts','--cases',str(inputs)],'/tmp/adamic-v4-2-base')
for round_ in range(3):
 for side in (['base','v4'] if round_%2==0 else ['v4','base']):
  for mode in ['release','sanitized']:
   run(side,mode,round_,[str(art/side/mode),'--cases',str(inputs)],'/tmp/adamic-v4-2-base' if side=='base' else '/tmp/adamic-json-82',mode=='sanitized')
assert len({r['stdout_sha256'] for r in measurements})==1,'source Node/native output mismatch'
(e/'measurement-input.json').write_text(json.dumps(dict(cases=3919,bytes=inputs.stat().st_size,sha256=hashlib.sha256(inputs.read_bytes()).hexdigest()),indent=2)+'\n')
print('All bulk-corpus stdout hashes match source Node',flush=True)

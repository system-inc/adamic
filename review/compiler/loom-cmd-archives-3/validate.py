import json, os, pathlib, subprocess, time
root=pathlib.Path('/tmp/adamic-recut3-evidence'); root.mkdir(exist_ok=True)
cache='/tmp/adamic-recut3-cache-final'
env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR=cache,ADAMIC_BUILD_STORE='off',ADAMIC_ORACLE_WASI='1')
binary='/tmp/adamic-recut3.test'
rows=[]
def run(name,cwd,binary,trace=False,mode='build'):
 e=env.copy();e['ADAMIC_BUILD_CACHE']='read' if mode=='read' else ''
 command=['timeout','60',binary,'-test.timeout=55s','-test.v','-test.run=^'+name+'$']
 label=mode+'-'+name
 if trace:
  label='trace-'+name;e['ADAMIC_TRACE_REPORT']='/tmp/'+label+'.jsonl';command=['timeout','60','/tmp/traced3','--']+command[2:]
 start=time.monotonic()
 with (root/(label+'.log')).open('w') as log:r=subprocess.run(command,cwd=cwd,env=e,stdout=log,stderr=subprocess.STDOUT)
 row=dict(test=name,mode='trace' if trace else mode,seconds=time.monotonic()-start,exit=r.returncode);rows.append(row);print(json.dumps(row),flush=True);(root/'validation.json').write_text(json.dumps(rows,indent=2)+'\n')
 if trace:
  p=pathlib.Path(e['ADAMIC_TRACE_REPORT'])
  if p.exists():(root/(label+'.jsonl')).write_bytes(p.read_bytes())
 if r.returncode:raise SystemExit(r.returncode)
for n in ['TestProduct_NativeRuntimesSanitized0','TestProduct_NativeRuntimesWASI0']:run(n,'internal/oracle','/tmp/oracle3.test',True)
products=['TestProduct_CommandRuntimeSanitized','TestProduct_CommandRuntimeWASI','TestProduct_CommandRuntimeWASICounted']
for n in products:run(n,'cmd/adamic',binary,True)
listing=subprocess.check_output([binary,'-test.list=^Test'],cwd='cmd/adamic',env=env,text=True)
for mode in ['build','read']:
 for name in listing.splitlines():
  if name.startswith('Test'):run(name,'cmd/adamic',binary,mode=mode)

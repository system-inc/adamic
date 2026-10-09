import pathlib,json,subprocess,time,os
p=pathlib.Path('/tmp/u150/evidence');plan=json.loads((p/'plan.json').read_text());results=json.loads((p/'standalone-validation.json').read_text())
entries={'M1':'cst_main.ts','M2':'compose_main.ts','M3':'main.ts','M4':'main.ts','P1':'compose_main.ts','P2':'cst_main.ts','P3':'main.ts'}
for item in plan:
 id=item['id'];
 if any(x['id']==id and x['exit']==0 for x in results):continue
 diff=p/(id+'.diff');r=subprocess.run(['git','apply','--check',str(diff)],capture_output=True,text=True);assert r.returncode==0,(id,r.stderr);subprocess.run(['git','apply',str(diff)],check=True)
 try:
  if item['file'].endswith('.go'):cmd=['go','vet','./stage1/cohere/yaml/']
  else:cmd=['timeout','90','/tmp/u150/adamic','build','stage1/cohere/yaml/'+entries[id],'-o','/tmp/u150/validate-'+id,'--sanitize']
  start=time.monotonic()
  with (p/(id+'-standalone-build.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u150/cache/validate-'+id))
  result=dict(id=id,exit=r.returncode,seconds=time.monotonic()-start,command=cmd);results.append(result);(p/'standalone-validation.json').write_text(json.dumps(results,indent=2));print(result,flush=True)
  assert r.returncode==0,id
  if id=='P4':
   with (p/'P4-alone.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^TestFileDriver_Setup$'],stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_YAML_LIBRARY='/tmp/u150/library',ADAMIC_BUILD_CACHE_DIR='/tmp/u150/cache/P4-alone'))
   assert r.returncode==1
 finally:subprocess.run(['git','apply','-R',str(diff)],check=True)
assert subprocess.run(['git','diff','--exit-code']).returncode==0
# Final restoration checks cover both native compilation and the external library gap.
start=time.monotonic()
with (p/'restored-baseline.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^(TestCSTMatchesGo|TestComposeMatchGo|TestBundledParserDifference)$'],stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_YAML_LIBRARY='/tmp/u150/library'))
(p/'restored-timing.json').write_text(json.dumps({'exit':r.returncode,'seconds':time.monotonic()-start}));print('restored',r.returncode,flush=True)

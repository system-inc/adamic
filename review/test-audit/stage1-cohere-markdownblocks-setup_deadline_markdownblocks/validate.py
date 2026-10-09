import pathlib,subprocess,json,time,os
p=pathlib.Path('/tmp/u129/evidence');results=[]
for diff in sorted(p.glob('[SP][0-9].diff')):
 s=time.monotonic();r=subprocess.run(['git','apply','--check',str(diff)],capture_output=True,text=True);assert r.returncode==0,(diff,r.stderr)
 subprocess.run(['git','apply',str(diff)],check=True)
 try:
  with (p/(diff.stem+'-standalone-vet.log')).open('w') as log:r=subprocess.run(['go','vet','./stage1/cohere/markdownblocks/'],stdout=log,stderr=subprocess.STDOUT)
  results.append({'id':diff.stem,'vet_exit':r.returncode,'seconds':time.monotonic()-s});assert r.returncode==0,diff
  if diff.stem=='P1':
   with (p/'P1-alone.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^TestMarkdownLayoutSetupHasNoDeadline$'],stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u129/cache/P1-alone'))
   assert r.returncode==1
 finally:subprocess.run(['git','apply','-R',str(diff)],check=True)
 (p/'standalone-validation.json').write_text(json.dumps(results,indent=2));print(results[-1],flush=True)
assert subprocess.run(['git','diff','--exit-code']).returncode==0
s=time.monotonic();pattern=json.loads((p/'slice-baseline-timing.json').read_text())['command_pattern']
with (p/'restored-slice-baseline.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',pattern],stdout=log,stderr=subprocess.STDOUT)
(p/'restored-slice-timing.json').write_text(json.dumps({'exit':r.returncode,'seconds':time.monotonic()-s}));print('restored',r.returncode,flush=True)

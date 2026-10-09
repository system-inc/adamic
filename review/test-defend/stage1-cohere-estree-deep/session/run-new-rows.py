import pathlib,subprocess,json,os,time,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-estree-deep/session'
while 'matrix completed' not in (p/'matrix-driver.log').read_text():
 if 'Baseline batch failed' in (p/'matrix-driver.log').read_text():raise SystemExit('Main baseline blocked')
 time.sleep(2)
subprocess.run(['git','diff','--exit-code','--','stage1/cohere/estree/convert.ts'],cwd=root,check=True)
names=json.loads((p/'scope-difference.json').read_text())['added'];sel='^('+'|'.join(names)+')$';env=os.environ.copy();env.update(ADAMIC_ESTREE_LIBRARY='/tmp/estree-deep-defense/library',ADAMIC_NATIVE_SPLIT='1',ADAMIC_ESTREE_BENCHMARK='1')
old=(root/'stage1/cohere/estree/convert.ts').read_bytes();out=[]
for mid in ['clean','D1','D2','D3']:
 try:
  if mid!='clean':
   env['ADAMIC_BUILD_CACHE_DIR']='/tmp/estree-deep-defense/cache/'+mid;subprocess.run(['git','apply',str(p/(mid+'.diff'))],cwd=root,check=True)
  else:env.pop('ADAMIC_BUILD_CACHE_DIR',None)
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',sel];start=time.monotonic();log=mid+'-new-rows.log'
  with (p/log).open('w') as f:r=subprocess.run(command,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in (p/log).read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  result=dict(id=mid,command=command,cache=env.get('ADAMIC_BUILD_CACHE_DIR','default'),wall=time.monotonic()-start,exit=r.returncode,cooked=r.returncode==124 or any('test timed out' in e.get('Output','') for e in events),statuses={e['Test']:e['Action'] for e in events if e.get('Test') and '/' not in e['Test'] and e['Action'] in ['pass','fail','skip']},errors=[dict(test=e.get('Test'),line=e['Output'].strip()) for e in events if e.get('OutputType')=='error'],log=log)
  out.append(result);(p/'new-rows-matrix.json').write_text(json.dumps(out,indent=2))
  if mid=='clean' and r.returncode:raise SystemExit('New-row baseline blocked')
 finally:(root/'stage1/cohere/estree/convert.ts').write_bytes(old)
print('new rows completed')

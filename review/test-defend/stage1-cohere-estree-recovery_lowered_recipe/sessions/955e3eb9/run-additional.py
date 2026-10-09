import pathlib,json,subprocess,os,time
r=pathlib.Path('/tmp/defend-estree-scalars');src=pathlib.Path('stage1/cohere/estree/values.ts');original=src.read_text();assert 'base === 16 && digits.length > 16' in original
rows=json.loads((r/'additional-matrix-rows.json').read_text());cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^('+'|'.join(rows)+')$'];results=[]
try:
 for mode in ['baseline','D1']:
  env=os.environ.copy()
  if mode=='D1':
   src.write_text(original.replace('base === 16 && digits.length > 16','base === 16 && digits.length > 17'));env['ADAMIC_BUILD_CACHE_DIR']=str(r/'cache/D1')
  start=time.monotonic()
  with (r/(mode+'-additional.log')).open('w') as log:result=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  results.append({'mode':mode,'command':cmd,'cache':env.get('ADAMIC_BUILD_CACHE_DIR'),'exit':result.returncode,'wall_seconds':time.monotonic()-start});(r/'additional-timings.json').write_text(json.dumps(results,indent=2));print(mode,result.returncode,flush=True)
  if mode=='baseline' and result.returncode:break
finally:src.write_text(original)

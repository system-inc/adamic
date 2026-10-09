import pathlib,json,subprocess,os,time
p=pathlib.Path('review/test-audit/internal-oracle-private_generic_mutant');names=['TestRuntimeLastIndexOfMatchesNode','TestScannerNestedReferences'];pattern='^('+'|'.join(names)+')$'
for id in ['CONTROL','M01','M02','M03','M04','P02','P03','P01']:
 patterns=['^'+n+'$' for n in names] if id=='P01' else [pattern]
 for i,pat in enumerate(patterns):
  logid=id+('-'+str(i+1) if id=='P01' else '')
  env=os.environ.copy();env.update(ADAMIC_MUTANT=id,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/u068/cache/'+id)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pat];start=time.monotonic()
  with (p/(logid+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  (p/(logid+'-run.json')).write_text(json.dumps({'command':'ADAMIC_MUTANT='+id+' ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start},indent=2))

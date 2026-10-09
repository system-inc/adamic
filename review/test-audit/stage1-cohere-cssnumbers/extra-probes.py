import pathlib,json,subprocess,time,os,difflib
root=pathlib.Path('/tmp/u082');out=pathlib.Path('review/test-audit/stage1-cohere-cssnumbers')
while not (root/'matrix-done').exists():time.sleep(1)
base=(out/'base.txt').read_text().strip();paths=['stage1/cohere/cssnumbers/port_test.go','stage1/cohere/cssnumbers/shards_test.go','stage1/cohere/cssnumbers/setup_products_test.go'];original={p:subprocess.check_output(['git','show',base+':'+p],text=True) for p in paths};scratch=original.copy()
probes=[('P2',paths[0],'func prepareCSSNumbersSetup(t *testing.T) {','return','^TestCSSNumbers_Setup$'),('P3',paths[1],'func numbersUnits(corpus numbersCorpus) []numbersUnit {','return nil','^TestCSSNumbersUnion$'),('P4',paths[0],'func numbersPreparedOracle(t *testing.T) string {','return ""','^TestProduct_CSSNumbersGoOracle$'),('P5',paths[1],'func numbersInput(texts []string) []byte {','return nil','^TestProduct_CSSNumbersGoAnswers$'),('P6',paths[2],'func numbersPreparedProgram(t *testing.T, mutation int) numbersProgram {','return numbersProgram{}','^TestProduct_CSSNumbers(Port|Mutant[012])Lowered$'),('P7',paths[2],'func numbersPreparedNative(t *testing.T, mutation int, sanitize bool) string {','return ""','^TestProduct_CSSNumbers((Port|Mutant[012])Native|FastNative)$')]
try:
 scratch[paths[2]]=scratch[paths[2]].replace('import (','import (\n\t"os"',1)
 for mid,path,entry,ret,pattern in probes:scratch[path]=scratch[path].replace(entry,entry+'\n\tif os.Getenv("ADAMIC_AUDIT_MODE") == "'+mid+'" { '+ret+' }',1)
 for p,s in scratch.items():pathlib.Path(p).write_text(s)
 subprocess.run(['gofmt','-w',*paths],check=True)
 with (out/'construction-empty-switch.diff').open('w') as log:subprocess.run(['git','diff','--',*paths],stdout=log,check=True)
 for mid,path,entry,ret,pattern in probes:
  env=os.environ.copy();env['ADAMIC_CSSNUMBERS_LIBRARY']='/tmp/u082/library';env['ADAMIC_AUDIT_MODE']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u082/cache/'+mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/cssnumbers/','-run',pattern];start=time.monotonic()
  with (root/(mid+'.log')).open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  (root/(mid+'.meta')).write_text(json.dumps(dict(exit=p.returncode,wall=time.monotonic()-start,command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],mode=mid)))
 (out/'construction-probes.json').write_text(json.dumps([dict(id=mid,file=path,line=original[path][:original[path].index(entry)].count('\n')+1,return_value=ret,pattern=pattern) for mid,path,entry,ret,pattern in probes],indent=2)+'\n')
finally:
 for p,s in original.items():pathlib.Path(p).write_text(s)
(root/'extra-done').write_text('done')

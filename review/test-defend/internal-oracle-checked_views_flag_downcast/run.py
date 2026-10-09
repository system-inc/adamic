import subprocess,pathlib,os,json,time
root=pathlib.Path('/workspace/adamic');out=pathlib.Path('/tmp/def-view')
mutants=[('D1','internal/javascript/view_unions_read.go','!contract.FixedTuple','len(contract.Tuple) == 0','change dispatch option from tuple-identity flag to retained tuple metadata'),('D2','internal/javascript/view_unions_untagged.go','\tif !certified {\n\t\treturn recorded\n\t}\n','','drop entire uncertified-producer early return'),('D3','internal/javascript/javascript.go','if statement.CheckAfter {\n\t\te.line("if (!%s) break;"','if statement.CheckAfter {\n\t\te.line("if (%s) break;"','flip post-test loop exit condition')]
results=[]
for mid,file,old,new,desc in mutants:
 p=root/file;original=p.read_text();assert original.count(old)==1
 line=original[:original.index(old)].count('\n')+1
 try:
  p.write_text(original.replace(old,new))
  (out/(mid+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file],cwd=root))
  t=time.monotonic()
  with open(out/(mid+'.vet.log'),'w') as log: vet=subprocess.run(['go','vet','./internal/javascript/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0
  env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/def-view/cache/'+mid)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^TestCheckedView']
  with open(out/(mid+'.log'),'w') as log:run=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[json.loads(x) for x in (out/(mid+'.log')).read_text().splitlines() if x.startswith('{')]
  fails=sorted(set(x['Test'].split('/')[0] for x in events if x.get('Action')=='fail' and x.get('Test')))
  passes=sorted(set(x['Test'] for x in events if x.get('Action')=='pass' and x.get('Test') and '/' not in x['Test']))
  errors=[x['Output'].strip() for x in events if '.go:' in x.get('Output','') and x.get('Test','').split('/')[0] in fails]
  results.append(dict(mutant=mid,file_line=file+':'+str(line),change=desc,rows_failed=fails,rows_passed=passes,errors=errors,seconds=time.monotonic()-t,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)))
  (out/'matrix.json').write_text(json.dumps(results,indent=2))
 finally:p.write_text(original)

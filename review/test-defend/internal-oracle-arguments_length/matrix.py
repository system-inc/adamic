import subprocess,pathlib,os,json,time
root=pathlib.Path('/workspace/adamic');out=pathlib.Path('/tmp/def-args')
mutants=[('D1','internal/lower/cast_proof.go','l.constant(message)}},','l.constant(message[:len(message)-1])}},','off by one: omit final byte of class-cast panic text'),('D2','internal/javascript/javascript.go','\t\te.statements(statement.Catch)\n','','drop catch-body emission statement'),('D3','internal/javascript/javascript.go','\t\t\te.nested(statement.Finally)\n','','drop finally-body emission statement'),('D4','internal/javascript/javascript.go','\t\te.statements(statement.Catch)\n','\t\te.statements(statement.Finally)\n','swap catch and finally body arguments')]
results=[]
for mid,file,old,new,desc in mutants:
 p=root/file;original=p.read_text();assert original.count(old)==1,(mid,original.count(old));line=original[:original.index(old)].count('\n')+1
 changed=original.replace(old,new)
 if mid=='D4':changed=changed.replace('\t\t\te.nested(statement.Finally)\n','\t\t\te.nested(statement.Catch)\n')
 try:
  p.write_text(changed);(out/(mid+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file],cwd=root))
  t=time.monotonic()
  with open(out/(mid+'.vet.log'),'w') as f:vet=subprocess.run(['go','vet','./'+str(pathlib.Path(file).parent)+'/'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
  assert vet.returncode==0,mid
  env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/def-args/cache/'+mid)
  runs=[];events=[]
  for suffix,pattern in [('casts','^(TestCheckedCast|TestUncheckableCast|TestCallTarget|TestDirectClosure|TestInterfaceCast)'),('neighbors','^TestNativeAgreesWithNode$/internal/oracle/testdata/(cast|exceptions|try|finally)')]:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern]
   with open(out/(mid+'-'+suffix+'.log'),'w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   events += [json.loads(x) for x in (out/(mid+'-'+suffix+'.log')).read_text().splitlines() if x.startswith('{')]
   runs.append(dict(command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),returncode=r.returncode,log=mid+'-'+suffix+'.log'))
  fails=sorted(set(x['Test'].split('/')[0] for x in events if x.get('Action')=='fail' and x.get('Test')))
  passes=sorted(set(x['Test'] for x in events if x.get('Action')=='pass' and x.get('Test') and '/' not in x['Test']))
  errors=[x['Output'].strip() for x in events if '.go:' in x.get('Output','') and x.get('Test','').split('/')[0] in fails]
  results.append(dict(mutant=mid,file_line=file+':'+str(line),change=desc,rows_failed=fails,rows_passed=passes,errors=errors,seconds=time.monotonic()-t,runs=runs))
  (out/'matrix.json').write_text(json.dumps(results,indent=2))
 finally:p.write_text(original)

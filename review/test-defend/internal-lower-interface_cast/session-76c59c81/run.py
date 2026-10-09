import pathlib,subprocess,time,json,os
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-defend/internal-lower-interface_cast'
plans=[('D01','internal/lower/iteration_origin.go','if isCallee(where) { // Calls carry','if isCallee(where) && false { // Calls carry','Disable callee exemption while keeping destructured-field refusal'),('D02','internal/lower/class_inheritance.go','if mismatch == nil {\n\t\treturn nil\n\t}\n\ttarget = mismatch','if mismatch == nil || node.Kind == ast.KindNewExpression {\n\t\treturn nil\n\t}\n\ttarget = mismatch','Skip nominal rejection for constructor-expression contextual assignments')]
(out/'plans.json').write_text(json.dumps(plans,indent=2))
for ident,file,old,new,why in plans:
 p=root/file;orig=p.read_text();assert orig.count(old)==1
 line=orig[:orig.index(old)].count('\n')+1
 p.write_text(orig.replace(old,new,1))
 try:
  diff=subprocess.check_output(['git','diff','--',file],cwd=root);(out/(ident+'.diff')).write_bytes(diff)
  start=time.monotonic()
  with (out/(ident+'.vet.log')).open('w') as f: vet=subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-interface-cast/cache/'+ident)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.']
  with (out/(ident+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for s in (out/(ident+'.log')).read_text().splitlines():
   try:events.append(json.loads(s))
   except:pass
  failed=sorted({x['Test'].split('/')[0] for x in events if x.get('Test') and x['Action']=='fail'})
  passed=sorted({x['Test'] for x in events if x.get('Test') and '/' not in x['Test'] and x['Action']=='pass'})
  data=dict(id=ident,file_line=file+':'+str(line),change=why,rows_failed=failed,rows_passed=passed,wall_seconds=time.monotonic()-start,vet_exit=vet.returncode,exit=r.returncode)
  (out/(ident+'.results.json')).write_text(json.dumps(data,indent=2));print(ident,failed,flush=True)
 finally:p.write_text(orig)

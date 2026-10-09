import pathlib,subprocess,os,time,json,re
root=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/fresh-defense/evidence')
base={f:subprocess.check_output(['git','show','HEAD:'+f],cwd=root).decode() for f in ['internal/fresh/fresh.go','internal/fresh/node_fs_file.go']}
mutants=[('D1','internal/fresh/fresh.go','\t\t\tif !a.proof.direct[target] {\n\t\t\t\t// Its parameters are outside to it, so it may keep them anywhere.\n\t\t\t\ta.call(operands, 0)\n\t\t\t}','','drop the conservative escape statement for non-direct method calls'),('D2','internal/fresh/node_fs_file.go','func (a *analysis) nodeFSFile(expression ir.NodeFSFile) value {','func (a *analysis) nodeFSFile(expression ir.NodeFSFile) value {\n\tif !mutable(expression.Type()) {\n\t\treturn a.call(a.operands(expression), expression.Type())\n\t}','return early through a retaining, clobbering call for primitive host operations'),('D3','internal/fresh/node_fs_file.go','\tfor _, argument := range expression.Arguments {\n\t\ta.value(argument)\n\t}','','drop the complete filesystem argument-evaluation loop'),('D4','internal/fresh/fresh.go','\tcase ir.NodeFSFile:\n\t\treturn a.nodeFSFile(expression)','\tcase ir.NodeFSFile:','drop the recognized host-return statement, reaching the existing unknown fallback')]
(p/'planned-mutants.json').write_text(json.dumps([{'mutant':i,'file_line':f+':'+str(base[f][:base[f].index(old)].count('\n')+1+(1 if i == 'D4' else 0)),'change':desc} for i,f,old,new,desc in mutants],indent=2))
results=[]
try:
 for i,f,old,new,desc in mutants:
  for file,s in base.items():root.joinpath(file).write_text(s)
  assert base[f].count(old)==1
  root.joinpath(f).write_text(base[f].replace(old,new))
  subprocess.run(['gofmt','-w',f],cwd=root,check=True)
  (p/(i+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',f],cwd=root))
  env=dict(os.environ,TMPDIR='/tmp/fresh-defense/tmp',ADAMIC_BUILD_CACHE_DIR='/tmp/fresh-defense/cache/'+i)
  start=time.monotonic()
  with (p/(i+'-vet.log')).open('w') as out:vet=subprocess.run(['timeout','90','go','vet','./internal/fresh/'],cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  r={'mutant':i,'change':desc,'vet_status':vet.returncode,'vet_wall':round(time.monotonic()-start,3)}
  if vet.returncode:results.append(r);break
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/fresh/','-run','.'];start=time.monotonic()
  with (p/(i+'.log')).open('w') as out:run=subprocess.run(command,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  r.update(command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command),status=run.returncode,wall=round(time.monotonic()-start,3),rows_failed=[],rows_passed=[],first_assertions={},panic=False,timeout=False)
  for line in (p/(i+'.log')).read_text().splitlines():
   try:x=json.loads(line)
   except:continue
   t=x.get('Test','');out=x.get('Output','')
   if t and '/' not in t and x.get('Action') in ['fail','pass']:r['rows_'+('failed' if x['Action']=='fail' else 'passed')].append(t)
   if t and re.search(r'\w+_test.go:\d+:',out) and t not in r['first_assertions']:r['first_assertions'][t]=out.strip()
   if 'panic:' in out:r['panic']=True
   if 'test timed out' in out:r['timeout']=True
  results.append(r);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(i,r['wall'],r['rows_failed'],'passes',len(r['rows_passed']),flush=True)
finally:
 for f,s in base.items():root.joinpath(f).write_text(s)
 (p/'matrix.json').write_text(json.dumps(results,indent=2))

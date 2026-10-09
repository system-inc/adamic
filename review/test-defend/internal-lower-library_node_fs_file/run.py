import pathlib,subprocess,json,os,time,difflib
r=pathlib.Path('/workspace/adamic');o=r/'review/test-defend/internal-lower-library_node_fs_file'
plans=[('D1','internal/lower/library_node.go','parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindInterfaceDeclaration','parent.Kind == ast.KindEnumDeclaration || parent.Kind == ast.KindInterfaceDeclaration','TestNodeLibraryDistinguishesReceiverOwners'),('Q1','internal/lower/proven_relations.go','exact := origin != nil && origin.Kind == ast.KindObjectLiteralExpression','exact := origin != nil && origin.Kind == ast.KindNewExpression','TestNodeLibraryQualifiedTypeAssertion'),('Q2','internal/lower/proven_relations.go','''\tif field := l.optionalRelationFailure(source, target, expression, map[[2]*checker.Type]bool{}); field != "" {
\t\treturn refuse("optional field "+field+" has no proven compatible presence/type", "keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)")
\t}
''','','TestNodeLibraryQualifiedTypeAssertion'),('Q3','internal/lower/proven_relations.go','exact := origin != nil && origin.Kind == ast.KindObjectLiteralExpression','exact := origin != nil && origin.Kind != ast.KindObjectLiteralExpression','TestNodeLibraryQualifiedTypeAssertion')]
original={f:(r/f).read_text() for _,f,_,_,_ in plans};plan=[]
for mid,f,a,b,target in plans:
 assert original[f].count(a)==1;plan.append({'id':mid,'file':f,'line':original[f][:original[f].index(a)].count('\n')+1,'old':a,'new':b,'target':target})
(o/'plan.json').write_text(json.dumps(plan,indent=2));results=[];defended=set()
try:
 for mid,f,a,b,target in plans:
  if target in defended:continue
  s=original[f];edited=s.replace(a,b);(r/f).write_text(edited)
  (o/(mid+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),edited.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
  with (o/(mid+'-vet.log')).open('w') as out:rc=subprocess.run(['go','vet','./internal/lower/'],cwd=r,stdout=out,stderr=subprocess.STDOUT).returncode
  if rc:raise RuntimeError('vet failed '+mid)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-nodefs/cache/'+mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];start=time.monotonic()
  with (o/(mid+'.log')).open('w') as out:rc=subprocess.run(cmd,cwd=r,env=env,stdout=out,stderr=subprocess.STDOUT).returncode
  events=[]
  for line in (o/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  failed=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']];passed=[e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']]
  result={'id':mid,'wall_seconds':time.monotonic()-start,'exit':rc,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'rows_failed':failed,'rows_passed':passed,'failures':[{'test':e.get('Test'),'output':e['Output'].strip()} for e in events if e.get('OutputType')=='error'],'package_seconds':[e.get('Elapsed') for e in events if e.get('Action')=='fail' and 'Test' not in e]};results.append(result);(o/'results.json').write_text(json.dumps(results,indent=2));print(mid,result['wall_seconds'],failed,flush=True)
  (r/f).write_text(s)
  if failed==[target]:defended.add(target)
finally:
 for f,s in original.items():(r/f).write_text(s)

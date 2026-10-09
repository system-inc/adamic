import pathlib,subprocess,json,os,time,difflib
r=pathlib.Path('/workspace/adamic');o=r/'review/test-defend/internal-lower-prototype'
plans=[('D1','internal/lower/object.go','''\tif l.inheritedLibraryMember(node) && !l.regexRuntimeProperty(access.Expression, name) && name != "length" && name != "size" && !(l.isLibraryType(l.checker.GetTypeAtLocation(access.Expression), "Error") && (name == "name" || name == "message")) {
\t\treturn nil, l.prototypeRead(node, name)
\t}
''',''),('D2','internal/lower/proven_relations.go','''\t\t\t\t\tif field := l.optionalRelationFailure(toArguments[index], fromArguments[index], nil, seen); field != "" {
\t\t\t\t\t\treturn "type argument." + field
\t\t\t\t\t}
''',''),('D3','internal/lower/expression.go','\t\treturn l.expression(satisfies.Expression)\n','')]
original={f:(r/f).read_text() for _,f,_,_ in plans};plan=[]
for mid,f,a,b in plans:
 assert original[f].count(a)==1,(mid,original[f].count(a));plan.append({'id':mid,'file':f,'line':original[f][:original[f].index(a)].count('\n')+1,'old':a,'new':b})
(o/'plan.json').write_text(json.dumps(plan,indent=2));results=[]
try:
 for mid,f,a,b in plans:
  s=original[f];edited=s.replace(a,b);(r/f).write_text(edited)
  (o/(mid+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),edited.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
  with (o/(mid+'-vet.log')).open('w') as out:rc=subprocess.run(['go','vet','./internal/lower/'],cwd=r,stdout=out,stderr=subprocess.STDOUT).returncode
  if rc:raise RuntimeError('vet failed '+mid)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-prototype/cache/'+mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];start=time.monotonic()
  with (o/(mid+'.log')).open('w') as out:rc=subprocess.run(cmd,cwd=r,env=env,stdout=out,stderr=subprocess.STDOUT).returncode
  events=[]
  for line in (o/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  failed=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']];passed=[e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']]
  result={'id':mid,'wall_seconds':time.monotonic()-start,'exit':rc,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'rows_failed':failed,'rows_passed':passed,'failures':[{'test':e.get('Test'),'output':e['Output'].strip()} for e in events if e.get('OutputType')=='error'],'package_seconds':[e.get('Elapsed') for e in events if e.get('Action')=='fail' and 'Test' not in e]};results.append(result);(o/'results.json').write_text(json.dumps(results,indent=2));print(mid,result['wall_seconds'],failed,flush=True)
  (r/f).write_text(s)
finally:
 for f,s in original.items():(r/f).write_text(s)

import pathlib,subprocess,json,time,os,difflib
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/internal-lower-module_namespace'
C='TestNamespaceClassEarlyConstructionStaysLoud';A='TestNamespaceAmbientHostInitialization';S='TestTscNamespaceDeclarationShapes'
plans=[
 dict(id='C1',test=C,file='internal/lower/namespaces_call_graph.go',old='if node == nil || ast.IsTypeNode(node) || node != function && ast.IsFunctionLike(node) {',new='if node == nil || ast.IsTypeNode(node) || node.Kind == ast.KindNewExpression || node != function && ast.IsFunctionLike(node) {',kind='change condition: return early for constructor expressions',lead='constructor expression inside reachable factory, unlike callable namespace property read'),
 dict(id='A1',test=A,file='internal/lower/namespaces.go',old='return l.notYet(node, "a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration")',new='return l.notYet(node, "a namespace read before initialization, directly or through a reachable call; move that read or call after the namespace declaration")',kind='change diagnostic constant',lead='direct read preflight checks runtime initialization wording; callable subsumer uses reachable-call branch'),
 dict(id='S1',test=S,file='internal/lower/namespaces.go',old='declaration == node || declaration.Kind == ast.KindInterfaceDeclaration || declaration.Kind == ast.KindTypeAliasDeclaration',new='declaration == node || declaration.Kind == ast.KindTypeAliasDeclaration',kind='change condition: stop admitting interface/namespace merges',lead='BuilderState interface shares namespace symbol'),
 dict(id='C2',test=C,file='internal/lower/namespaces.go',old='case ast.KindClassDeclaration:\n\t\t\treturn declaration',new='case ast.KindClassDeclaration:\n\t\t\treturn nil',kind='return early empty callable class target',lead='class target discovery exclusive line 363'),
 dict(id='S2',test=S,file='internal/lower/namespaces.go',old='declaration == node || declaration.Kind == ast.KindInterfaceDeclaration || declaration.Kind == ast.KindTypeAliasDeclaration',new='declaration == node || declaration.Kind == ast.KindInterfaceDeclaration',kind='change condition: stop admitting type-alias/namespace merges',lead='BinaryExpressionState alias shares namespace symbol'),
 dict(id='C3',test=C,file='internal/lower/namespaces_call_graph.go',old='if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression {',new='if node.Kind == ast.KindCallExpression {',kind='change condition: omit constructor call graph edges',lead='construction edges in factory versus ordinary callable property reads'),
 dict(id='S3',test=S,file='internal/lower/namespace_callable.go',old='"toString":',new='"log":',kind='change intrinsic name constant',lead='Debug.log nested callable namespace exports log')]
orig={x['file']:(R/x['file']).read_text() for x in plans}
for x in plans:
 s=orig[x['file']];assert s.count(x['old'])==1,(x['id'],s.count(x['old']));x['line']=s[:s.index(x['old'])].count('\n')+1
 (P/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(x['old'],x['new'],1).splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
(P/'plan.json').write_text(json.dumps(plans,indent=2));runs=[];defended=set();matrix={}
try:
 for x in plans:
  if x['test'] in defended:continue
  f=R/x['file'];f.write_text(orig[x['file']].replace(x['old'],x['new'],1));id=x['id']
  with (P/(id+'.vet.log')).open('w') as log:r=subprocess.run(['timeout','90','go','vet','./internal/lower/'],cwd=R,stdout=log,stderr=subprocess.STDOUT)
  assert r.returncode==0,id
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/workspace/defend-module-cache/'+id
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];t=time.monotonic()
  with (P/(id+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=R,env=env,stdout=log,stderr=subprocess.STDOUT)
  es=[]
  for l in (P/(id+'.log')).read_text().splitlines():
   try:es.append(json.loads(l))
   except:pass
  states={e['Test']:e['Action'] for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']};fails=[n for n,a in states.items() if a=='fail'];run=dict(id=id,exit=r.returncode,seconds=time.monotonic()-t,command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],vet_exit=0,binary_seconds=next((e.get('Elapsed') for e in reversed(es) if not e.get('Test') and e.get('Action') in ['pass','fail']),None));runs.append(run);matrix[id]=dict(states=states,failed=fails,passed=[n for n,a in states.items() if a=='pass'],skipped=[n for n,a in states.items() if a=='skip']);(P/'runs.json').write_text(json.dumps(runs,indent=2));(P/'matrix.json').write_text(json.dumps(matrix,indent=2));print(id,run['seconds'],len(states),fails,flush=True)
  f.write_text(orig[x['file']]);assert len(states)==275,(id,'incomplete matrix')
  if fails==[x['test']]:defended.add(x['test'])
finally:
 for f,s in orig.items():(R/f).write_text(s)

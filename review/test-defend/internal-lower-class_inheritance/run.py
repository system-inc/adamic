import pathlib,subprocess,json,difflib,time,shutil
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-defend/internal-lower-class_inheritance'; tmp=pathlib.Path('/tmp/defend-inherit');out.mkdir(parents=True,exist_ok=True)
pairs=[('TestInheritanceRefusesGrowingGenericClasses','TestInheritanceHasClassIdentity'),('TestInheritanceConditionalThisRules','TestInheritanceRefusesThisBeforeSuperReturns'),('TestInheritanceGenericViewsKeepNominalArguments','TestInheritanceGenericNominalConstraints')]
def cov(name):
 d={}
 for line in (tmp/(name+'.cover')).read_text().splitlines()[1:]:
  loc,n,count=line.split();d[loc]=int(count)
 return d
coverage={a:[k for k,v in cov(a).items() if v and not cov(b).get(k,0)] for a,b in pairs}
(out/'exclusive-coverage.json').write_text(json.dumps(coverage,indent=2))
for p in tmp.glob('*.cover'):shutil.copy(p,out/p.name)
for n in ['clean.log','list.log','npm.log']:shutil.copy(tmp/n,out/n)
recipes=[('D1','internal/lower/class.go','keep recursive type arguments unchanged, or write a class per type','write a class per type'),('D2','internal/lower/class_super.go','\tfor _, parameter := range constructor.Parameters() {\n\t\tif err := flow.reads(parameter, superUninitialized); err != nil {\n\t\t\treturn 0, err\n\t\t}\n\t}\n',''),('D3','internal/lower/class_inheritance.go','\t\tfor index := range from {\n\t\t\tif !l.enumAssignable(from[index], to[index]) || !l.enumAssignable(to[index], from[index]) || l.nominalMismatch(from[index], to[index], seen) != nil || l.nominalMismatch(to[index], from[index], seen) != nil {\n\t\t\t\treturn false\n\t\t\t}\n\t\t}\n','')]
(out/'plan.json').write_text(json.dumps(recipes,indent=2))
results=[]
for mid,file,old,new in recipes:
 p=root/file;original=p.read_text();assert original.count(old)==1
 changed=original.replace(old,new);line=original[:original.index(old)].count('\n')+1
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 try:
  p.write_text(changed)
  with (out/(mid+'-vet.log')).open('w') as log: vet=subprocess.run(['bash','-c','source /workspace/adamic-tools/env.sh; timeout 90 go vet ./internal/lower/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0
  cmd=f'source /workspace/adamic-tools/env.sh; ADAMIC_BUILD_CACHE_DIR=/tmp/defend-inherit/cache/{mid} timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .'
  start=time.monotonic()
  with (out/(mid+'.log')).open('w') as log:r=subprocess.run(['bash','-c',cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for s in (out/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(s))
   except:pass
  failed=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test'in e and '/'not in e['Test']]
  passed=[e['Test'] for e in events if e.get('Action')=='pass' and 'Test'in e and '/'not in e['Test']]
  result=dict(mutant=mid,file_line=f'{file}:{line}',command=cmd,wall_seconds=time.monotonic()-start,returncode=r.returncode,rows_failed=failed,rows_passed=passed,failing_output=[e['Output'].strip() for e in events if e.get('Test') in failed and (' .go:'in e.get('Output','') or '.go:'in e.get('Output',''))])
  results.append(result);(out/'results.json').write_text(json.dumps(results,indent=2));print(mid,result['wall_seconds'],failed,flush=True)
 finally:p.write_text(original)

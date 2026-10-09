import subprocess,pathlib,json,difflib,time,os
root=pathlib.Path('/workspace/adamic'); p=root/'review/test-audit/stage1-cohere-lint-regex'; pkg='./stage1/cohere/lint/regex/'
names=['TestFixedPatterns','TestDynamicPatternGap','TestOptionDialectGap','TestInventoryMatchesPinnedSource','TestShapeFixtures']
changes=[('M2','stage1/cohere/lint/regex/patterns.a','/^no default(?![\\s\\S])/gui','/^no default(?![\\s\\S])/gu'),('M3','stage1/cohere/lint/regex/patterns.a','[a-z]{2,3}','[a-z]{2,2}'),('M4','stage1/cohere/lint/regex/patterns.a','[AEIOUY]','[aeiouy]'),('M1','internal/lower/regexp.go','"RegExp with a nonconstant pattern"','"RegExp pattern unavailable"'),('S1','stage1/cohere/lint/regex/testdata/inventory.go','sel.Sel.Name != "Compile"','sel.Sel.Name != "CompileBROKEN"'),('S2','stage1/cohere/lint/regex/testdata/inventory.go','depth > 20','depth > 0'),('S3','stage1/cohere/lint/regex/testdata/inventory.go','Line: fs.Position(c.Pos()).Line,','Line: fs.Position(c.Pos()).Line + 1,'),('W1','stage1/cohere/lint/regex/regex_test.go','bytes.Equal(expected, actual)','true'),('W2','stage1/cohere/lint/regex/testdata/shapes/gate.go','equal(f, "source Node", want, actual)','equal(f, "source Node", actual, actual)'),('P1','stage1/cohere/lint/regex/patterns.a','export function fixedPatterns(): RegExp[] {','export function fixedPatterns(): RegExp[] { return [];'),('P2','internal/lower/lower.go','files := program.Files()','if program != nil { return nil, nil }\n\tfiles := program.Files()')]
results=json.loads((p/"runs.json").read_text()) if os.environ.get("U118_RESUME")=="1" else []
for ident,file,old,new in changes:
 if ident in {r["id"] for r in results}: continue
 target=root/file; base=subprocess.check_output(['git','show','HEAD:'+file],cwd=root).decode(); assert old in base
 changed=base.replace(old,new,2) if ident=='W1' else base.replace(old,new,1)
 target.write_text(changed)
 if file.endswith('.go'): subprocess.run(['gofmt','-w',str(target)],check=True)
 changed=target.read_text(); (p/(ident+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u118/cache/'+ident
 start=time.monotonic()
 try:
  with (p/(ident+'-compile.log')).open('w') as f:
   vet=['go','vet',pkg] if not file.startswith('internal/lower') else ['go','vet','./internal/lower/']
   v=subprocess.run(['timeout','90']+vet,stdout=f,stderr=subprocess.STDOUT,env=env)
   if '/testdata/' in file and file.endswith('.go'): v2=subprocess.run(['go','vet',file],stdout=f,stderr=subprocess.STDOUT,env=env);assert v2.returncode==0
  assert v.returncode==0,(ident,v.returncode)
  compile_seconds=time.monotonic()-start;start=time.monotonic()
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.']
  with (p/(ident+'.log')).open('w') as f: r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
  duration=time.monotonic()-start
  events=[]
  for l in (p/(ident+'.log')).read_text().splitlines():
   try: events.append(json.loads(l))
   except: pass
  panic=any('panic:' in x.get('Output','') for x in events)
  if panic or r.returncode==124:
   for n in names:
    with (p/(ident+'-'+n+'.log')).open('w') as f: subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','^'+n+'$'],stdout=f,stderr=subprocess.STDOUT,env=env)
  results.append({'id':ident,'file':file,'line':base[:base.index(old)].count('\n')+1,'change':old+' -> '+new,'compile_seconds':compile_seconds,'run_wall_seconds':duration,'exit':r.returncode,'panic':panic,'command':' '.join(cmd),'fail_rows':[x['Test'] for x in events if x.get('Action')=='fail' and x.get('Test') in names]})
  (p/'runs.json').write_text(json.dumps(results,indent=2)+'\n');print(ident,results[-1],flush=True)
 finally: target.write_text(base)

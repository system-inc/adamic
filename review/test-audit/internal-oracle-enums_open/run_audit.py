import pathlib,subprocess,time,json,difflib,os,re
root=pathlib.Path.cwd(); out=root/'review/test-audit/internal-oracle-enums_open'; scope=json.loads((out/'scope.json').read_text()); rows=scope['tests']; regex='^('+ '|'.join(rows)+')$'; meta=[]
files=['internal/native/emit.go','internal/lower/expression.go','internal/lower/diagnostics.go','internal/lower/lower.go','internal/javascript/javascript.go','internal/oracle/oracle_test.go']; originals={f:(root/f).read_text() for f in files}
variants=[('M1','internal/native/emit.go','\temitter.block(program.Main, nil)\n',''),('M2','internal/lower/expression.go','if to == ir.Union && value != nil && value.Type() != ir.Union {','if to == ir.Union && value != nil && value.Type() == ir.Union {'),('M3','internal/lower/diagnostics.go','"%s: Adamic 0.1 refuses %s; %s"','"%s: Adamic 0.1 refuses %s; %.0s"'),('M4','internal/lower/diagnostics.go','"%s: stage 0 can\'t lower %s yet"','"%s: stage 0 can\'t lower %s now"'),('P1','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\treturn nil, nil'),('P2','internal/native/emit.go','func C(program *ir.Program) string {','func C(program *ir.Program) string {\n\treturn ""'),('P3','internal/javascript/javascript.go','func JavaScript(program *ir.Program) string {','func JavaScript(program *ir.Program) string {\n\treturn ""'),('W1','internal/oracle/oracle_test.go','func disagreement(oracle run, native run) string {','func disagreement(oracle run, native run) string {\n\treturn ""'),('W2','internal/oracle/oracle_test.go','command.Env = append(os.Environ(), environment...)','command.Env = append(append(os.Environ(), environment...), "ASAN_OPTIONS=detect_leaks=0")')]
def run(label,command,env=None):
 start=time.monotonic()
 with (out/(label+'.log')).open('w') as log: result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,env=env)
 item={'label':label,'command':command,'seconds':round(time.monotonic()-start,3),'exit':result.returncode}
 if env:item['selector']=env.get('ADAMIC_MUTANT'); item['cache']=env.get('ADAMIC_BUILD_CACHE_DIR')
 meta.append(item); (out/'commands.json').write_text(json.dumps(meta,indent=2)+'\n'); print(json.dumps(item),flush=True); return result.returncode
try:
 for row in rows:
  for n in [1,2,3]:
   assert run('timing-'+row+'-'+str(n),['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row+'$'])==0,'clean target red: '+row
 for name,file,old,new in variants:
  text=originals[file]; assert old in text; changed=text.replace(old,new,1); (root/file).write_text(changed)
  (out/(name+'.diff')).write_text(''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
  if name.startswith('M'):
   assert run(name+'-vet',['timeout','90','go','vet','./'+str(pathlib.Path(file).parent)+'/'])==0
  (root/file).write_text(text)
 switched=dict(originals)
 def replace(file,old,new):
  assert old in switched[file]; switched[file]=switched[file].replace(old,new,1)
 replace('internal/native/emit.go','\temitter.block(program.Main, nil)\n','\tif os.Getenv("ADAMIC_MUTANT") != "M1" { emitter.block(program.Main, nil) }\n')
 replace('internal/lower/expression.go','if to == ir.Union && value != nil && value.Type() != ir.Union {','if to == ir.Union && value != nil && (value.Type() != ir.Union) != (os.Getenv("ADAMIC_MUTANT") == "M2") {')
 replace('internal/lower/diagnostics.go','return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %s", r.Where, r.What, r.Fix)','if os.Getenv("ADAMIC_MUTANT") == "M3" { return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %.0s", r.Where, r.What, r.Fix) }; return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %s", r.Where, r.What, r.Fix)')
 replace('internal/lower/diagnostics.go','return fmt.Sprintf("%s: stage 0 can\'t lower %s yet", n.Where, n.What)','if os.Getenv("ADAMIC_MUTANT") == "M4" { return fmt.Sprintf("%s: stage 0 can\'t lower %s now", n.Where, n.What) }; return fmt.Sprintf("%s: stage 0 can\'t lower %s yet", n.Where, n.What)')
 for name,file,old,new in variants:
  if name.startswith('P') or name=='W1': replace(file,old,old+'\n\tif os.Getenv("ADAMIC_MUTANT") == "'+name+'" { '+ ('return nil, nil' if name=='P1' else 'return ""')+' }')
 replace('internal/oracle/oracle_test.go','command.Env = append(os.Environ(), environment...)','command.Env = append(os.Environ(), environment...)\n\t\tif os.Getenv("ADAMIC_MUTANT") == "W2" { command.Env = append(command.Env, "ASAN_OPTIONS=detect_leaks=0") }')
 for file,text in switched.items():
  if 'os.Getenv' in text and '"os"' not in text:text=text.replace('import (','import (\n\t"os"',1)
  (root/file).write_text(text)
 assert run('switch-gofmt',['gofmt','-w',*files])==0
 assert run('switch-compile',['timeout','90','go','test','-c','-o','/tmp/u060-oracle.test','./internal/oracle/'])==0
 (out/'switch.diff').write_bytes(subprocess.check_output(['git','diff','--',*files]))
 for name,*_ in variants:
  env=os.environ.copy(); env['ADAMIC_MUTANT']=name; env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u060/cache/'+name; env['ADAMIC_GATE_UNCACHED']='1'
  target=regex
  if name=='W1':target='^(TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestFallthroughMutants|TestFieldReadinessRepresentation)$'
  if name=='W2':target='^TestEnumCleanupMutant$'
  run(name,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',target],env)
  content=(out/(name+'.log')).read_text()
  if 'panic:' in content and ('goroutine' in content or 'test timed out' in content):
   for row in rows:run(name+'-isolated-'+row,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row+'$'],env)
finally:
 for f,text in originals.items():(root/f).write_text(text)

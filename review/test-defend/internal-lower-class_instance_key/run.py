import json,os,pathlib,subprocess,time,sys
p=pathlib.Path("review/test-defend/internal-lower-class_instance_key")
menu=[
("D1","internal/lower/empty_literal.go","\tproven = l.concrete(proven)","\tif l.typeMapper != nil { return nil }\n\tproven = l.concrete(proven)","Return early for generic contextual empty-array representation"),
("D2","internal/lower/class_inheritance.go","\t\t\tvalue, err = l.expression(member.AsPropertyDeclaration().Initializer)","","Drop field initializer evaluation, retaining zero value"),
("D3","internal/lower/assignments.go","return append(statements, ir.Assign{Local: local, Value: fit(value, l.result.Locals[local].Type), Checked: l.checked(local) && !l.result.Locals[local].NamespaceState}), nil","return statements, nil","Drop final local-assignment emission"),
("D4","internal/lower/class.go","field.Declarations[0].Kind == ast.KindPropertyDeclaration || parameterProperty(field.Declarations[0])","field.Declarations[0].Kind == ast.KindParameter || parameterProperty(field.Declarations[0])","Change allowed constructor property-declaration kind")]
records=json.loads((p/'matrix.json').read_text()) if (p/'matrix.json').exists() else []
for ident,file,before,after,change in menu:
 if len(sys.argv)>1 and ident not in sys.argv[1:]: continue
 f=pathlib.Path(file); original=f.read_text(); assert original.count(before)==1,(ident,original.count(before)); line=original[:original.index(before)].count('\n')+1
 try:
  f.write_text(original.replace(before,after)); subprocess.run(['gofmt','-w',file],check=True)
  (p/(ident+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file]))
  with (p/(ident+'-vet.log')).open('w') as log: vet=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0,ident+' vet'
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-lower-class-key/cache/'+ident)
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.']; start=time.monotonic()
  with (p/(ident+'.log')).open('w') as log: run=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT)
  results={}; evidence={}
  for raw in (p/(ident+'.log')).read_text().splitlines():
   try: e=json.loads(raw)
   except: continue
   test=e.get('Test','');action=e.get('Action')
   if test and '/' not in test and action in ['pass','fail','skip']:results[test]=action
   if test and e.get('Output') and any(s in e['Output'] for s in ['empty_literal_test.go:','definite_assignment_test.go:']): evidence.setdefault(test,e['Output'].strip())
  expected={l.strip() for l in (p/'list.log').read_text().splitlines() if l.startswith('Test')}
  record=dict(mutant=ident,file_line=file+':'+str(line),change=change,command=' '.join(command),cache=env['ADAMIC_BUILD_CACHE_DIR'],wall_seconds=time.monotonic()-start,exit=run.returncode,rows_failed=sorted(k for k,v in results.items() if v=='fail'),rows_passed=sorted(k for k,v in results.items() if v=='pass'),rows_skipped=sorted(k for k,v in results.items() if v=='skip'),rows_unknown=sorted(expected-set(results)),evidence=evidence)
  records.append(record); (p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n'); print(ident,record['rows_failed'],len(record['rows_unknown']),flush=True)
 finally: f.write_text(original)

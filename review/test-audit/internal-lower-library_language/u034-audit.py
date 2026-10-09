import pathlib,json,subprocess,difflib,time,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-lower-library_language'
rows=['TestLibraryLanguageBoundaries','TestLibraryMapSetGapsStayRefused','TestLibraryMapSetIteratorCopyTypesRefused','TestLibraryMethodValues','TestLibraryMethodValueBoundaries','TestLibraryMethodValueSafety','TestNodeBufferRefusals']
listed=(out/'u034-list.log').read_text().splitlines();assert all(r in listed for r in rows)
functions=[l for l in (out/'u034-functions.log').read_text().splitlines() if l.startswith('github.') and float(l.split()[-1].rstrip('%'))>0]
(out/'reached-functions.txt').write_text('\n'.join(functions)+'\n')
menu=[]
def add(id,file,old,new,kind='expr',op='flip condition'):
 f='internal/lower/'+file;s=(root/f).read_text();assert s.count(old)==1,(id,s.count(old));menu.append(dict(id=id,file=f,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind,operator=op))
add('M01','library_for_in.go','l.checker.IsArrayType(proven) || checker.IsTupleType(proven)','!l.checker.IsArrayType(proven) && !checker.IsTupleType(proven)')
add('M02','library_for_in.go','!l.plainEnumerableObject(statement.Expression, map[*ast.Symbol]bool{})','l.plainEnumerableObject(statement.Expression, map[*ast.Symbol]bool{})')
add('M03','library_for_in.go','safe = false','safe = true',kind='stmt',op='change constant')
add('M04','library_globals.go','parent.Kind == ast.KindTypeOfExpression','parent.Kind != ast.KindTypeOfExpression')
add('M05','library_function_expressions.go','parameter.Name().Text() == "this"','parameter.Name().Text() == "that"',op='change constant')
add('M06','library_map_set.go','element != otherElement','element == otherElement')
add('M07','library_map_set.go','ast.SkipParentheses(arguments[0]).Kind != ast.KindArrowFunction','ast.SkipParentheses(arguments[0]).Kind == ast.KindArrowFunction')
add('M08','library_map_set.go','!known || held != element','!known || held == element')
add('M09','library_map_set.go','func (l *lowering) libraryIteratorUnsupportedUse(node *ast.Node) error {','return nil',kind='entry',op='return early')
add('M10','library_method_values.go','declaration.Parent.Flags&ast.NodeFlagsConst == 0','declaration.Parent.Flags&ast.NodeFlagsConst != 0')
add('M11','library_method_values.go','!constMethodInitializer(node)','constMethodInitializer(node)')
# Use complete conditional as selector to avoid evaluating spread expressions differently.
add('M12','library_method_values.go','if hasSpread(node) {','if !hasSpread(node) {',kind='condition')
add('M13','library_method_values.go','len(written) != 1 || ast.SkipParentheses(written[0]).Kind != ast.KindArrayLiteralExpression','len(written) != 2 || ast.SkipParentheses(written[0]).Kind != ast.KindArrayLiteralExpression',op='change bound')
add('M14','library_method_values.go','receiver.Type() != ir.String || l.mayBeUndefined(written[0])','receiver.Type() != ir.String || !l.mayBeUndefined(written[0])')
add('M15','library_method_values.go','if l.mayBeUndefined(receiver) {','if !l.mayBeUndefined(receiver) {',kind='condition')
add('M16','library_method_values.go','if value.Type() != ir.Number {','if value.Type() == ir.Number {',kind='condition')
add('M17','library_node_buffer.go','args[0].Text() != "sha256"','args[0].Text() != "sha1"',op='change constant')
add('M18','library_node_buffer.go','args[0].Text() != "hex"','args[0].Text() != "base64"',op='change constant')
add('M19','library_node_buffer.go','\t\t\treturn false\n\t\t}\n\t}\n\treturn true\n}\n\nfunc (l *lowering) nodeBufferEncoding','\t\t\treturn true\n\t\t}\n\t}\n\treturn true\n}\n\nfunc (l *lowering) nodeBufferEncoding',kind='block',op='change constant')
add('M20','library_method_values.go','arguments = append(arguments, reads[1])','arguments = append(arguments, reads[0])',kind='stmt',op='change constant index')
add('P01','lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return nil, nil',kind='entry',op='empty-answer probe')
(out/'menu.json').write_text(json.dumps(menu,indent=2))
(out/'code-and-oracle.md').write_text('CODE UNDER TEST: Adamic internal/lower entry and reached lowering functions listed in reached-functions.txt, generated from scoped coverage before mutants. Load is preparation.\nORACLE: self-written acceptance/rejection, NotYet identity, and reason substrings. No scoped row runs an external comparator. Safety accepts any nonnil error and can be masked by unrelated refusal.\nFixed 20-mutant menu in menu.json recorded before any mutant runtime results. Only production code is changed. One distinct entry: Lower, P01 returns nil,nil.\n')
base={m['file']:(root/m['file']).read_text() for m in menu}
validation=[]
for m in menu:
 s=base[m['file']]
 if m['kind']=='entry':
  start=s.index(m['old']);end=s.index('\n}',start)+2;new=s[:start]+m['old']+'\n\t'+m['new']+'\n}'+s[end:]
  if m['id']=='P01':
   for imp in ['fmt','path/filepath']:new=new.replace('\t"'+imp+'"\n','')
 else:new=s.replace(m['old'],m['new'])
 (root/m['file']).write_text(new)
 (out/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 start=time.monotonic()
 with (out/(m['id']+'-vet.log')).open('w') as log:rc=subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90).returncode
 (root/m['file']).write_text(s);validation.append(dict(id=m['id'],exit=rc,seconds=time.monotonic()-start))
 assert rc==0,m['id']
(out/'validation.json').write_text(json.dumps(validation,indent=2))
for file,s in base.items():
 for m in [m for m in menu if m['file']==file]:
  id=m['id'];old=m['old'];new=m['new'];kind=m['kind']
  if kind=='entry':rep=old+'\n\tif auditMutant("'+id+'") { '+new+' }'
  elif kind=='stmt':rep='if auditMutant("'+id+'") { '+new+' } else { '+old+' }'
  elif kind=='condition':rep='if auditChoose("'+id+'", '+old[3:-2].strip()+', '+new[3:-2].strip()+') {'
  elif kind=='block':rep=old.replace('return false','return auditMutant("'+id+'")')
  else:rep='auditChoose("'+id+'", '+old+', '+new+')'
  s=s.replace(old,rep)
 (root/file).write_text(s)
(root/'internal/lower/u034_switch.go').write_text('package lower\nimport "os"\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT")==id }\nfunc auditChoose[T any](id string, normal, mutant T) T { if auditMutant(id) { return mutant }; return normal }\n')
subprocess.run(['gofmt','-w',*[str(root/f) for f in base],str(root/'internal/lower/u034_switch.go')],check=True)
start=time.monotonic()
with (out/'switch-build.log').open('w') as log:subprocess.run(['go','test','-c','-o','/tmp/u034-lower.test','./internal/lower/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=90)
(out/'switch-build-seconds.json').write_text(json.dumps(time.monotonic()-start))
with (out/'inactive-switch-baseline.log').open('w') as log:
 env=os.environ.copy();env['ADAMIC_MUTANT']='';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u034/cache/inactive'
 subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
runs=[]
for m in menu:
 id=m['id'];env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u034/cache/'+id
 # The probe is allowed to panic consumers; run only the seven requested rows individually.
 pattern='.' if id!='P01' else '^('+'|'.join(rows)+')$'
 start=time.monotonic()
 with (out/(id+'.log')).open('w') as log:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',pattern],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 data=(out/(id+'.log')).read_text();panic='panic:' in data;timeout='test timed out' in data or rc==124
 runs.append(dict(id=id,exit=rc,wall_seconds=time.monotonic()-start,pattern=pattern,panic=panic,timeout=timeout))
 print(id,rc,round(runs[-1]['wall_seconds'],3),'panic',panic,'timeout',timeout,flush=True)
 if panic or timeout:
  for row in rows:
   with (out/(id+'-'+row+'.log')).open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 (out/'runs.json').write_text(json.dumps(runs,indent=2))
for file,s in base.items():(root/file).write_text(s)
(root/'internal/lower/u034_switch.go').unlink()

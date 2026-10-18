import ast,difflib,json,pathlib,subprocess,time
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-r-views-fx-d';rows=[]
mutations=[]
for script in ['run-mutants.py','run-extra-mutants.py','run-static-mutant.py']:
 tree=ast.parse((root/'review/compiler/loom-views-fx-d'/script).read_text())
 for expression in tree.body:
  if isinstance(expression,ast.Expr) and isinstance(expression.value,ast.Call) and isinstance(expression.value.func,ast.Name) and expression.value.func.id=='mutant':
   args=[ast.literal_eval(value) for value in expression.value.args]
   if args[0] not in ['cast-metadata','clear-callable-contract']:mutations.append(args)
assert len(mutations)==16
for entry in mutations:
 if entry[0]=='javascript-depth-128': entry[3]='^TestFX7Depth65JavaScript$'
 if entry[0]=='tuple-unmarked': entry[3]='^TestLoomViewsDTupleUnionJavaScript$'
 if entry[0]=='tuple-rejected-object': entry[3]='^TestLoomViewsDTupleObjectJavaScript$'
 if entry[0] in ['native-union-kind','native-static-kind']:
  filename,old,new=entry[1][0]
  if entry[0]=='native-union-kind': old,new='if (object->tuple) { value.kind = adamic_view_union_array; }','if (false) { value.kind = adamic_view_union_array; }'
  else: old,new='else if (object->class != NULL && object->class->is_static)','else if (false && object->class != NULL && object->class->is_static)'
  entry[1][0]=('internal/native/runtime/view_unions_untagged.c',old,new)
for name,changes,pkg,test in mutations:
 originals={}
 try:
  patches=[]
  for filename,old,new in changes:
   path=root/filename;originals.setdefault(path,path.read_text());text=path.read_text();assert old in text,(name,filename,old)
   changed=text.replace(old,new,1);path.write_text(changed)
   patches.extend(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/'+filename,tofile='b/'+filename,n=0))
  (ev/(name+'.patch')).write_text(''.join(patches))
  command=['timeout','120','go','test',pkg,'-run',test,'-count=1','-v','-timeout','55s']
  start=time.monotonic()
  with (ev/(name+'.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
  output=(ev/(name+'.log')).read_text();caught=r.returncode==1 and '--- FAIL:' in output and '[build failed]' not in output and 'error:' not in output and 'panic: test timed out' not in output
  row=dict(name=name,test=test,package=pkg,exit=r.returncode,caught=caught,seconds=time.monotonic()-start);rows.append(row);(ev/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
  assert caught,(name,output[-4000:])
 finally:
  for path,text in originals.items():path.write_text(text)

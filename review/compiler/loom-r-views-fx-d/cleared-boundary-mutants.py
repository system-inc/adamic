import pathlib,subprocess,difflib,json,time
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-r-views-fx-d';rows=[]
p=root/'internal/native/runtime/view_unions_untagged.c'
original=p.read_text()
start=original.index('    // Only this own-data adapter')
end=original.index('    return adamic_view_mixed_union_select(&value',start)
block=original[start:end]
mutated=original[:start]+original[end:]
anchor='    if (id != 0 && id <= count) {'
mutated=mutated.replace(anchor,block+anchor,1)
cases=[('kind-before-matching',p,mutated,'^TestFX7TupleSelectionBeforeDiagnosticKindRelease$')]
p2=root/'internal/native/runtime/view_unions_mixed.c';text=p2.read_text()
unsafe='    if (value->kind == adamic_view_union_object && value->payload.reference != NULL) {\n        const adamic_object *object = value->payload.reference;\n        if (object->tuple) { found = "array"; }\n        else if (object->class != NULL && object->class->is_static) { found = "function"; }\n    }\n'
anchor2='    const char *found = adamic_view_union_kind_name(value->kind);\n'
cases.append(('generic-payload-read',p2,text.replace(anchor2,anchor2+unsafe,1),'^TestViewMixedUnionUnknownAndUnavailable$/true'))
for name,path,changed,test in cases:
 before=path.read_text()
 try:
  assert changed!=before
  (ev/(name+'.patch')).write_text(''.join(difflib.unified_diff(before.splitlines(True),changed.splitlines(True),fromfile='a/'+str(path.relative_to(root)),tofile='b/'+str(path.relative_to(root)))))
  path.write_text(changed);starttime=time.monotonic()
  with (ev/(name+'.log')).open('w') as log:
   result=subprocess.run(['timeout','150','go','test','./internal/native','-run',test,'-count=1','-v','-timeout','55s'],stdout=log,stderr=subprocess.STDOUT)
  output=(ev/(name+'.log')).read_text()
  caught=result.returncode==1 and '--- FAIL:' in output and '[build failed]' not in output and 'panic: test timed out' not in output
  assert caught,output[-3000:]
  if name=='kind-before-matching': assert 'found array' in output
  else: assert 'AddressSanitizer: stack-buffer-overflow' in output
  rows.append(dict(name=name,caught=caught,seconds=time.monotonic()-starttime,test=test))
  (ev/'cleared-boundary-mutants.json').write_text(json.dumps(rows,indent=2)+'\n');print(rows[-1],flush=True)
 finally:path.write_text(before)

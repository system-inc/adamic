import difflib, os, pathlib, subprocess
root=pathlib.Path(__file__).resolve().parent
path=pathlib.Path('cmd/adamic/runtimes_test.go')
original=path.read_text()
mutants=[('undeclared-sanitized','\t{"sanitized", native.Options{Sanitize: true}, nil},\n','', '^TestExplainChecksOutput$/^overload_some_empty$/^build$', 'no runtime product declared'),('missing-key-part','slices.Clone(request.Portable),','slices.Clone(request.Portable[1:]),','^TestCommandRuntimeProductKeyHoldsEveryPart$', 'native key part')]
try:
 for name,old,new,test,message in mutants:
  assert original.count(old)==1, (name,original.count(old))
  changed=original.replace(old,new)
  (root/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/cmd/adamic/runtimes_test.go',tofile='b/cmd/adamic/runtimes_test.go')))
  path.write_text(changed)
  with (root/(name+'.log')).open('w') as log:
   result=subprocess.run(['timeout','90','go','test','./cmd/adamic','-run',test,'-count=1','-timeout','60s'],stdout=log,stderr=subprocess.STDOUT)
  text=(root/(name+'.log')).read_text()
  assert result.returncode==1 and message in text, (name,result.returncode,text)
  print(name+': caught by '+test,flush=True)
  path.write_text(original)
finally:
 path.write_text(original)

import difflib, pathlib, subprocess
root=pathlib.Path(__file__).resolve().parent
runtime=pathlib.Path('cmd/adamic/runtimes_test.go')
order=pathlib.Path('cloud/fast-gate/product-order.json')
originals={runtime:runtime.read_text(),order:order.read_text()}
mutants=[
 ('second-builder',runtime,'buildcache.BuiltElsewhere(t, "the "+product.name+" command runtime archive", commandRuntimeInputs(request), product.owner, build)','buildcache.Product(t, commandRuntimeInputs(request), build)','TestSharedCommandRuntimeCannotBuildItsOwnersArchive','a shared runtime product must fail for its missing owner'),
 ('second-owner',runtime,'{"sanitized", native.Options{Sanitize: true}, nil, "internal/oracle TestProduct_NativeRuntimesSanitized0"}','{"sanitized", native.Options{Sanitize: true}, nil, ""}','TestCommandRuntimeArchivesHaveOneBuilder','has 2 builder packages [internal/oracle cmd/adamic]'),
 ('missing-wait',order,'    "cmd/adamic TestProduct_CommandRuntimeSanitized": [\n      "internal/oracle TestProduct_NativeRuntimesSanitized0"\n    ],\n','','TestCommandRuntimeProductsWaitForOwners','must wait for internal/oracle TestProduct_NativeRuntimesSanitized0'),
 ('missing-key-part',runtime,'return runtimeproducts.Inputs(request)','request.Portable = request.Portable[1:]\n return runtimeproducts.Inputs(request)','TestCommandRuntimeProductKeyHoldsEveryPart','native key part "adamic-runtime-v6"'),
 ('undeclared-sanitized',runtime,'\t{"sanitized", native.Options{Sanitize: true}, nil, "internal/oracle TestProduct_NativeRuntimesSanitized0"},\n','','TestExplainChecksOutput$/^overload_some_empty$/^build','no runtime product declared'),
]
try:
 for name,path,old,new,test,message in mutants:
  original=originals[path]
  assert original.count(old)==1,(name,original.count(old))
  changed=original.replace(old,new)
  (root/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+str(path),tofile='b/'+str(path),n=0)))
  path.write_text(changed)
  with (root/(name+'.log')).open('w') as log:
   result=subprocess.run(['timeout','90','go','test','./cmd/adamic','-run','^'+test+'$','-count=1','-timeout','60s'],stdout=log,stderr=subprocess.STDOUT)
  output=(root/(name+'.log')).read_text()
  assert result.returncode==1 and message in output,(name,result.returncode,output)
  print(name+': caught by '+test,flush=True)
  path.write_text(original)
finally:
 for path,original in originals.items():path.write_text(original)

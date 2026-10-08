import json,pathlib,subprocess
root=pathlib.Path('/workspace/adamic')
results=[]
for name,file,needle,replacement,package,test in [
 ('required-undefined-rejected','internal/lower/phantom_array_brands.go','phantomField(l.checker.GetTypeOfSymbol(field), true)','phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0)','./internal/oracle','TestPhantomSortedArrayProbe'),
 ('required-cast-not-erased','internal/lower/cast.go','return value, err\n\t}\n\tif l.phantomCast', 'return ir.Defined{Value: value, Message: "mutant cast"}, err\n\t}\n\tif l.phantomCast','./internal/lower','TestPhantomArrayRequiredCastsAreErased'),
]:
 p=root/file; original=p.read_text(); assert needle in original
 try:
  p.write_text(original.replace(needle,replacement))
  with open('/tmp/phantom-presence-mutant-'+name+'.log','w') as log:
   result=subprocess.run(['go','test',package,'-run','^'+test+'$','-count=1'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  results.append({'mutant':name,'test':test,'exit':result.returncode})
 finally: p.write_text(original)
pathlib.Path('/tmp/phantom-presence-extra-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(x['exit']==1 for x in results),results

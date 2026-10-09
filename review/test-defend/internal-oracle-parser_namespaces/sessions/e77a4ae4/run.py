from pathlib import Path
import subprocess,time,json,os
root=Path('/workspace/adamic');p=Path('/tmp/defend-parser-namespaces')
mutants=[('D1','internal/lower/namespace_callable.go','if parent.Kind == ast.KindTypeOfExpression {\n\t\treturn true','if parent.Kind == ast.KindTypeOfExpression {\n\t\treturn false',33,'Change the typeof-use acceptance constant true to false'),('D2','internal/lower/namespace_receiver_scope.go','node.Kind != ast.KindArrowFunction','node.Kind != ast.KindMethodDeclaration',13,'Change the nested-function scope exemption from ArrowFunction to MethodDeclaration'),('D3','internal/lower/namespaces.go','Condition: ir.Unary{Operator: ir.Not, Operand: b.read(b.parameters[0])}','Condition: b.read(b.parameters[0])',418,'Flip the Boolean namespace-readiness condition by removing negation')]
r1='^(TestParserNamespaceReceiver|TestParserNamespaceClass|TestParserCallableNamespace)$'
r2='^TestNativeAgreesWithNode$/(stage3|internal)/(namespace-parser-stops|oracle)/(native-(namespace|callable)|testdata)/(namespace_(method_receiver|class_registration|callable_properties))'
(p/'plan.json').write_text(json.dumps({'mutants':mutants,'code_under_test':'Adamic namespace lowering in internal/lower','oracle':'Node executing the unmodified original sources','matrix_regexes':[r1,r2]},indent=2)+'\n')
results=[]
for mid,file,before,after,line,desc in mutants:
 f=root/file; original=f.read_text(); assert original.count(before)==1,(mid,original.count(before))
 try:
  f.write_text(original.replace(before,after))
  (p/(mid+'.diff')).write_text(subprocess.check_output(['git','diff','--',file],cwd=root,text=True))
  env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/mid)
  cmds=[('vet',['go','vet','./internal/lower/']),('target',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',r1]),('subsumer',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',r2])]
  for label,cmd in cmds:
   started=time.monotonic()
   with open(p/(mid+'.'+label+'.log'),'w') as out:ret=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT).returncode
   result={'mutant':mid,'label':label,'command':cmd,'exit':ret,'wall':time.monotonic()-started,'cache':env['ADAMIC_BUILD_CACHE_DIR']};results.append(result);(p/'commands.json').write_text(json.dumps(results,indent=2)+'\n');print(result,flush=True)
   if label=='vet' and ret:raise RuntimeError('vet failed')
 finally:f.write_text(original)

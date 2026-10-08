#!/usr/bin/env python3
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[4]
logs=Path('/tmp/views-callables-factory-generic-mutants');logs.mkdir(exist_ok=True)
cases=[
('generic-method-identity','internal/native/view_callables_methods.go','e.line("%s = NULL;", recorded)','e.line("(void)%s;", recorded)','./internal/oracle','^TestCheckedViewCallableFactoryGenericMethodDeclaration$','generic method negative exit=0 stdout="7\\n" stderr=""'),
('generic-instantiation-key','internal/lower/generic.go','key += "," + l.genericTypeKey(concrete)','key += "," + typeName(held)','./internal/oracle','^TestCheckedViewCallableFactoryGenerics$/^cloneNode$/^conversion$','generic instantiations collapsed: got 1 bodies, want 4'),
('generic-result-identity','internal/lower/view_callables_generic.go','return identicalTypes(l.checker, instantiateType(l.checker, l.checker.GetReturnTypeOfSignature(producer), mapper), l.checker.GetReturnTypeOfSignature(wanted))','return true','./internal/lower','^TestGenericCallableDeclarationIdentity$','generic declaration badResult compatibility true, want false'),
('generic-result-instantiation','internal/lower/view_callables_generic.go','if value == nil || !l.checker.IsTypeAssignableTo(l.concrete(l.checker.GetTypeAtLocation(value)), promised) {','if false {','./internal/lower','^TestGenericCallableResultInstantiation$','constraint substituted for result instantiation: <nil>'),
('generic-call-effects','internal/ir/call_targets.go','if len(call.GenericInstances) != 0 {','if false {','./internal/ir','^TestGenericClosureTargetsUseInstantiatedBodies$','generic call followed template rather than instantiated effects'),
]
for name,relative,before,after,package,test,failure in cases:
 path=root/relative;original=path.read_bytes()
 try:
  source=original.decode();assert source.count(before)==1
  path.write_text(source.replace(before,after,1))
  log=logs/(name+'.log')
  with log.open('w') as output: result=subprocess.run(['go','test',package,'-run',test,'-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  assert result.returncode!=0 and failure in log.read_text(),str(log)
  if name=='generic-method-identity':assert 'sanitized generic method exit=0 stdout="7\\n" stderr=""' in log.read_text(),str(log)
  print(name+': caught by '+failure)
 finally:path.write_bytes(original)

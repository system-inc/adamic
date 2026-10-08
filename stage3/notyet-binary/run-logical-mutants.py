"""Kill independent semantic and ownership mutants, restoring the source after each."""
import os
import pathlib
import subprocess

root=pathlib.Path.cwd()
scratch=pathlib.Path('/tmp/notyet-binary-logical-mutants')
scratch.mkdir(exist_ok=True)
mutants=[
 ('and-branch','internal/native/emit_branches.go','if coalesce.Logical == ir.And {','if coalesce.Logical == ir.And && false {','binary_logical_number_number'),
 ('or-branch','internal/native/emit_branches.go','if coalesce.Logical == ir.And {','if coalesce.Logical == ir.And || coalesce.Logical == ir.Or {','binary_logical_number_number'),
 ('nan-truthiness','internal/native/census_small.go','fmt.Sprintf("(%s != 0.0 && !isnan(%s))", value, value)','fmt.Sprintf("(%s != 0.0)", value)','binary_logical_number_number'),
 ('empty-string-truthiness','internal/native/census_small.go','fmt.Sprintf("(%s != NULL && %s->length != 0)", value, value)','fmt.Sprintf("(%s != NULL)", value)','binary_logical_union'),
 ('unselected-left-transfer','internal/native/emit_branches.go','coalesce.Logical == 0 && coalesce.Of.IsReference() && !fresh && e.taken(unwrapped)','coalesce.Of.IsReference() && !fresh && e.taken(unwrapped)','binary_logical_union'),
 ('js-operand-selection','internal/javascript/javascript.go','operator = "||"','operator = "&&"','binary_logical_number_boolean'),
]
for name,file,before,after,fixture in mutants:
 path=root/file
 original=path.read_text()
 assert original.count(before)==1, (name,original.count(before))
 try:
  path.write_text(original.replace(before,after))
  log=scratch/(name+'.log')
  with log.open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/'+fixture,'-count=1','-timeout','10m'],stdout=output,stderr=output,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'))
  text=log.read_text()
  assert result.returncode!=0, ('mutant survived',name)
  assert 'stdout differs' in text or 'heap-use-after-free' in text or 'LeakSanitizer' in text or 'native stdout' in text or 'JavaScript stdout' in text, ('not a semantic/sanitizer kill',name,text)
  assert 'error:' not in text and '[build failed]' not in text, ('build kill',name)
  print(name,'caught, exit',result.returncode,'log',log,flush=True)
 finally:
  path.write_text(original)

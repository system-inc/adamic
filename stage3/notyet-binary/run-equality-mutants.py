"""Prove strict equality normalization and value comparison against source Node."""
import os
import pathlib
import subprocess

scratch=pathlib.Path('/tmp/notyet-binary-equality-mutants');scratch.mkdir(exist_ok=True)
mutants=[
 ('wrong-right-operand','internal/lower/expression.go','left, right = fit(left, ir.Union), fit(right, ir.Union)','left, right = fit(left, ir.Union), fit(left, ir.Union)'),
 ('boxed-pointer-equality','internal/native/emit_expressions.go','return fmt.Sprintf("adamic_union_equal(%s, %s)", left, right)','return fmt.Sprintf("(%s == %s)", left, right)'),
]
for name,file,before,after in mutants:
 path=pathlib.Path(file);original=path.read_text();assert original.count(before)==1
 try:
  path.write_text(original.replace(before,after))
  log=scratch/(name+'.log')
  with log.open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/binary_equality','-count=1','-timeout','10m'],stdout=output,stderr=output,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'))
  text=log.read_text()
  assert result.returncode!=0 and 'stdout differs' in text, (name,text)
  assert '[build failed]' not in text and 'error:' not in text, (name,text)
  print(name,'caught, exit',result.returncode,'log',log,flush=True)
 finally:path.write_text(original)

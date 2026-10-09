"""Break Error subclass rules serially; source Node and ASan must reject each."""
import os
from pathlib import Path
import subprocess
ROOT=Path.cwd()
LOGS=ROOT/'review/compiler/lowering-chain/source-member-17/evidence'
mutants=[
 ('error-ancestry-missing','internal/native/class_inheritance.go','base = "&adamic_" + map[string]string{"Error": "error", "RangeError": "range_error", "TypeError": "type_error"}[class.BuiltinError] + "_class"','base = "NULL"','stdout differs'),
 ('error-default-name-wrong','internal/lower/error_subclasses.go','l.constant(name)','l.constant("Subclass")','stdout differs'),
 ('error-prefix-released-twice','internal/lower/error_subclasses.go','OwnStart: 2','OwnStart: 0','AddressSanitizer: heap-use-after-free'),
 ('error-fields-out-of-order','internal/lower/class_inheritance.go','append(l.result.Functions[initializer].Body, initialized...)','append(initialized, l.result.Functions[initializer].Body...)','stdout differs'),
]
for name,relative,old,new,expected in mutants:
 if os.environ.get("MUTANT") and name != os.environ["MUTANT"]: continue
 path=ROOT/relative
 original=path.read_bytes()
 source=original.decode()
 assert old in source,name
 try:
  path.write_text(source.replace(old,new))
  log=LOGS/(name+'.log.txt')
  with log.open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/step21_error_subclasses[.]a$','-count=1','-v','-timeout=5m'],cwd=ROOT,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=output,stderr=subprocess.STDOUT,timeout=360)
  observed=log.read_text()
  assert result.returncode!=0 and expected in observed,(name,result.returncode,observed)
  assert 'clang failed' not in observed and '[build failed]' not in observed,(name,observed)
  print(f'{name}: exit {result.returncode}, caught by {expected}',flush=True)
 finally:
  path.write_bytes(original)

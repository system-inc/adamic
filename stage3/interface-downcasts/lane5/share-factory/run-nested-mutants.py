#!/usr/bin/env python3
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[4]
logs=Path('/tmp/views-callables-factory-nested-mutants');logs.mkdir(exist_ok=True)
cases=[
('native-nested-comparison','internal/native/runtime/view_callables_contract.h','return given != NULL && wanted != NULL && adamic_view_callable_signatures_match(given, wanted);','(void)given; (void)wanted; return true;'),
('javascript-nested-comparison','internal/javascript/view_callables_contract.go','return given !== undefined && wanted !== undefined && adamicViewCallableSignaturesMatch(given, wanted);','return true;'),
]
for name,relative,before,after in cases:
 path=root/relative;original=path.read_bytes()
 try:
  source=original.decode();assert source.count(before)==1
  path.write_text(source.replace(before,after,1))
  log=logs/(name+'.log')
  with log.open('w') as output:result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewCallableFactoryHigherOrder$/^wrong-nested$','-count=1','-v','-timeout','5m'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text();failure='negative exit=0 stdout="done\\n" stderr=""'
  assert result.returncode!=0 and failure in observed,str(log)
  if name.startswith('native'):assert 'sanitized nested exit=0 stdout="done\\n" stderr=""' in observed,str(log)
  print(name+': caught by forbidden clean output done instead of exit 70')
 finally:path.write_bytes(original)

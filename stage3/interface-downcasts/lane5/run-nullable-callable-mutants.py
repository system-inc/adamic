#!/usr/bin/env python3
"""Nullable dispatch omissions must lose a named refusal in valid executable code."""
from pathlib import Path
import subprocess
root = Path(__file__).resolve().parents[3]
logs = Path(__file__).resolve().parent / 'logs/nullable-callable-mutants'
logs.mkdir(parents=True, exist_ok=True)
for name,filename,before,after in [
 ('native','internal/native/view_nullish.go','e.viewCallableNullishCertificate(property, value)','// omit callable certificate'),
 ('javascript','internal/javascript/view_nullish.go','return e.viewCallableNullishCertificate(property, e.nullishMemberSelection(property, value))','return e.nullishMemberSelection(property, value)'),
]:
 path=root/filename
 original=path.read_text()
 assert original.count(before)==1
 try:
  path.write_text(original.replace(before,after))
  with (logs/(name+'.log')).open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewCallableNullableArity$','-count=1','-timeout','5m'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  evidence=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'error:' not in evidence,evidence
  assert 'exitCode:0' in evidence,evidence
  print(name+': nullable callable omission ran successfully and lost the required refusal',flush=True)
 finally:
  path.write_text(original)

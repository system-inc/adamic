#!/usr/bin/env python3
"""Run executable adapter and variance erasures, restoring each before the next."""
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
logs=Path(__file__).resolve().parent/'logs/boxing'
mutants=[
 ('result-value','internal/native/view_callables_boxing.go','adapted, _ := converted(function.Returns, returns, value)','adapted, _ := converted(function.Returns, returns, value)\n if function.Returns == ir.Number && returns == ir.Union { adapted = "adamic_box_number(0.0)" }','^TestCheckedViewCallableBoxing$/^number-result$','disagreement'),
 ('parameter-value','internal/native/view_callables_boxing.go','adapted, fresh := converted(from[i], to, value)','adapted, fresh := converted(from[i], to, value)\n if from[i] == ir.Number && to == ir.Union { adapted = "adamic_box_number(0.0)" }','^TestCheckedViewCallableBoxing$/^(parameter|array-predicate-parameter)$','disagreement'),
 ('parameter-release','internal/native/view_callables_boxing.go','release = append(release, fmt.Sprintf','_ = fmt.Sprint(0)\n release = append([]string{}, fmt.Sprintf','^TestCheckedViewCallableBoxing$/^parameter-throw$','leak'),
 ('native-members','internal/native/runtime/view_callables_contract.h','(given & wanted) == given','true','^TestCheckedViewCallableBoxingRefusals$/^view-.*wrong-members$','exitCode:0'),
 ('javascript-members','internal/javascript/view_callables_contract.go','(given & wanted) === given','true','^TestCheckedViewCallableBoxingRefusals$/^view-.*wrong-members$','exitCode:0'),
 ('native-contravariance','internal/native/runtime/view_callables_contract.h','(given & wanted) == given','(given & wanted) == wanted','^TestCheckedViewCallableBoxingRefusals$/^view-parameter-wrong$','exitCode:0'),
 ('javascript-contravariance','internal/javascript/view_callables_contract.go','(given & wanted) === given','(given & wanted) === wanted','^TestCheckedViewCallableBoxingRefusals$/^view-parameter-wrong$','exitCode:0'),
]
for name,filename,before,after,selection,kind in mutants:
 path=root/filename; original=path.read_text(); assert original.count(before)==1,(name,original.count(before))
 # A release omission must remove the actual cleanup, rather than replace it by another release.
 if name=='parameter-release':
  before='for _, line := range release {\n\t\t\tbody.WriteString(line)\n\t\t}'
  after='for _, line := range release { _ = line }'
  assert original.count(before)==1
 try:
  path.write_text(original.replace(before,after))
  with (logs/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run',selection,'-count=1','-timeout','5m'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  evidence=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'clang failed:' not in evidence,evidence
  if kind=='exitCode:0': assert kind in evidence,evidence
  elif kind=='leak': assert 'LeakSanitizer' in evidence or 'leak' in evidence.lower(),evidence
  else: assert 'stdout' in evidence or 'stdout' in evidence.lower(),evidence
  print(name+': executable mutant killed',flush=True)
 finally:path.write_text(original)

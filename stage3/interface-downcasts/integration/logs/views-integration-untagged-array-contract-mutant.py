from pathlib import Path
import subprocess
p=Path('internal/lower/view_contracts.go')
original=p.read_bytes()
try:
 s=original.decode();before=' && !l.untaggedPlainArrayUnion(target)';assert s.count(before)==1
 p.write_text(s.replace(before,''))
 with open('/tmp/views-integration-untagged-cd32-array-contract-mutant.log','wb') as out:
  r=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewUntaggedArrayUnion$/mixed$','-count=1','-v'],stdout=out,stderr=subprocess.STDOUT)
 log=Path('/tmp/views-integration-untagged-cd32-array-contract-mutant.log').read_text()
 assert r.returncode!=0 and '--- FAIL: TestCheckedViewUntaggedArrayUnion' in log and 'exit=0 stdout="true' in log and '[build failed]' not in log and 'clang failed' not in log,log
 print('plain array union descriptor erasure caught by both compiled backend refusal pins')
finally:
 p.write_bytes(original)

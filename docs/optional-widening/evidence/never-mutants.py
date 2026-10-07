from pathlib import Path
import subprocess
p=Path('internal/lower/never.go')
original=p.read_text()
mutants=[
 ('no-op', 'body = append(body, ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}})', 'body = append(body, ir.Return{Value: ir.NumberConstant{Value: float64(len(message))*0}})', '^TestNever'),
 ('skip-call', 'if effect != nil {', 'if effect != nil && false {', 'TestNativeAgreesWithNode/internal/oracle/testdata/never_call'),
]
for name,before,after,test in mutants:
 assert original.count(before)==1
 try:
  p.write_text(original.replace(before,after))
  with open('/tmp/never-mutant-'+name+'.log','w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run',test,'-count=1','-timeout','30m'],stdout=log,stderr=subprocess.STDOUT)
  print(name,'exit',result.returncode,flush=True)
  if result.returncode==0: raise SystemExit('mutant escaped')
  output=Path('/tmp/never-mutant-'+name+'.log').read_text()
  marker='exit 0, stdout \"continued' if name=='no-op' else 'stdout differs'
  if marker not in output or 'build failed' in output or 'clang:' in output: raise SystemExit('mutant did not fail its semantic assertion')
 finally:
  p.write_text(original)

"""Prove saved Error/rethrow identity and quiet uncaught backend behavior."""
import os
from pathlib import Path
import subprocess
ROOT=Path(__file__).resolve().parents[2]
LOGS=ROOT/'docs/step-21-exceptions/evidence'
mutants=[
 ('rethrow-wrapped','internal/native/exceptions.go','e.line("adamic_thrown = adamic_retain(%s);", value)','if read, ok := statement.Value.(ir.Read); ok && e.program.Locals[read.Local].Name == "error" { e.line("adamic_thrown = (adamic_heap *)adamic_error_new(((adamic_object *)%s)->slots[1].reference);", value) } else { e.line("adamic_thrown = adamic_retain(%s);", value) }','TestNativeAgreesWithNode/internal/oracle/testdata/step21_saved_error[.]a$','stdout differs'),
 ('builtin-narrow-forgets-subtype','internal/lower/locals.go','matches = ir.InstanceOf{Value: held, Class: identity}','_ = identity; matches = ir.InstanceOf{Value: held, Class: -1}','TestNativeAgreesWithNode/internal/oracle/testdata/step21_builtin_narrow_terminal[.]a$','want the inserted check to fire'),
 ('saved-error-wrapped','internal/native/exceptions.go','e.line("adamic_thrown = adamic_retain(%s);", value)','e.line("adamic_thrown = (adamic_heap *)adamic_error_new(((adamic_object *)%s)->slots[1].reference);", value)','TestNativeAgreesWithNode/internal/oracle/testdata/step21_saved_error[.]a$','stdout differs'),
 ('backend-renders-stack','internal/javascript/javascript.go','builder.WriteString("process.on(\'uncaughtException\', () => process.exit(1));\\n")','// mutant: renderer handles uncaught payload','^TestStep21Uncaught$','uncaught renderer wrote stderr'),
]
for name,relative,old,new,test,expected in mutants:
 if os.environ.get("MUTANT") and name != os.environ["MUTANT"]: continue
 path=ROOT/relative
 original=path.read_bytes()
 source=original.decode()
 assert source.count(old)==1,(name,source.count(old))
 try:
  path.write_text(source.replace(old,new,1))
  log=LOGS/(name+'.log.txt')
  with log.open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run',test,'-count=1','-v','-timeout=5m'],cwd=ROOT,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=output,stderr=subprocess.STDOUT,timeout=360)
  observed=log.read_text()
  assert result.returncode!=0 and expected in observed,(name,result.returncode,observed)
  assert 'clang failed' not in observed and '[build failed]' not in observed,(name,observed)
  if name in ('saved-error-wrapped', 'rethrow-wrapped'): assert 'false false' in observed,observed
  print(f'{name}: exit {result.returncode}, caught by {expected}',flush=True)
 finally:
  path.write_bytes(original)

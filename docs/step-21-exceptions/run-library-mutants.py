"""Prove catchable library failures and terminal soundness checks; restore all edits."""
import os
from pathlib import Path
import subprocess
ROOT = Path(__file__).resolve().parents[2]
LOGS = ROOT/'docs/step-21-exceptions/evidence'
mutants = [
 ('range-error-is-error','internal/lower/library_exceptions.go','Constructor: "RangeError"','Constructor: "Error"','library_failures','stdout differs'),
 ('wrong-range-message','internal/lower/library_exceptions.go','Invalid count value: ','Wrong count: ','library_failures','stdout differs'),
 ('type-error-is-error','internal/lower/library_exceptions.go','Constructor: "TypeError"','Constructor: "Error"','library_types','stdout differs'),
 ('hash-failure-is-panic','internal/native/runtime/node_crypto.c','adamic_thrown = &error->heap;','adamic_panic("Digest already called", 21); adamic_thrown = &error->heap;','library_host','exit codes differ'),
 ('host-type-error-is-error','internal/native/runtime/node_fs_file.c','strcmp(name, "TypeError") == 0 ? &adamic_host_type_error_class','strcmp(name, "TypeError") == 0 ? &adamic_host_error_class','library_host','stdout differs'),
 ('soundness-check-is-catchable','internal/lower/expression.go','ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}','ir.Throw{Value: ir.Box{Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant(message)}}}}','soundness_terminal','want the inserted check to fire'),
]
for name, relative, old, new, fixture, expected in mutants:
 if os.environ.get("MUTANT") and name != os.environ["MUTANT"]: continue
 path=ROOT/relative
 original=path.read_bytes()
 source=original.decode()
 assert old in source,name
 try:
  path.write_text(source.replace(old,new))
  log=LOGS/(name+'.log.txt')
  with log.open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run',f'TestNativeAgreesWithNode/internal/oracle/testdata/step21_{fixture}[.]a$','-count=1','-v','-timeout=5m'],cwd=ROOT,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=output,stderr=subprocess.STDOUT,timeout=360)
  observed=log.read_text()
  assert result.returncode != 0 and expected in observed,(name,result.returncode,observed)
  assert 'clang failed' not in observed and '[build failed]' not in observed,(name,observed)
  print(f'{name}: exit {result.returncode}, caught by {expected}',flush=True)
 finally:
  path.write_bytes(original)

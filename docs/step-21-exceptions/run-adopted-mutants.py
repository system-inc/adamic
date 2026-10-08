"""Run adopted exception rule mutants serially and restore every source file."""
import os
from pathlib import Path
import subprocess
ROOT = Path(__file__).resolve().parents[2]
LOGS = ROOT / 'docs/step-21-exceptions/evidence'
mutants = [
 ('implicit-string-conversion','internal/native/runtime/exceptions.c',
  'adamic_output_flush();', 'adamic_release(adamic_union_to_string(adamic_thrown)); adamic_output_flush();',
  '^TestStep21Uncaught/step21_object_uncaught$', 'uncaught lifetime check: exit 70'),
 ('pending-undefined-lost','internal/native/exceptions.go',
  'e.line("adamic_exception_pending = true;")\n\te.end()',
  'e.line("adamic_exception_pending = (adamic_thrown != NULL);")\n\te.end()',
  'TestNativeAgreesWithNode/internal/oracle/testdata/step21_dynamic[.]a$', 'stdout differs'),
 ('catch-assumes-error','internal/lower/exceptions.go',
  'return ir.InstanceOf{Value: value, Class: -1}, true',
  '_ = value; return ir.BooleanConstant{Value: true}, true',
  'TestNativeAgreesWithNode/internal/oracle/testdata/step21_dynamic[.]a$', 'exit codes differ'),
 ('typeof-boolean-is-number','internal/native/runtime/union.c',
  'case adamic_kind_boolean:\n\t\treturn &adamic_typeof_boolean;',
  'case adamic_kind_boolean:\n\t\treturn &adamic_typeof_number;',
  'TestNativeAgreesWithNode/internal/oracle/testdata/step21_dynamic[.]a$', 'stdout differs'),
 ('null-is-undefined','internal/native/union.go',
  'return "&adamic_null"', 'return "NULL"',
  'TestNativeAgreesWithNode/internal/oracle/testdata/step21_dynamic[.]a$', 'stdout differs'),
 ('finally-releases-twice','internal/native/exceptions.go',
  'e.hold(pending)', 'e.hold(pending)\n\t\te.hold(pending)',
  'TestNativeAgreesWithNode/internal/oracle/testdata/step21_dynamic[.]a$', 'AddressSanitizer: heap-use-after-free'),
 ('region-payload-freed','internal/native/region.go',
  'case ir.Box:\n\t\t\t// Boxing preserves reachability. A thrown object must escape the statement arena.\n\t\t\treturn derived(expression.Value)',
  'case ir.Box:\n\t\t\treturn false',
  'TestNativeAgreesWithNode/internal/oracle/testdata/step21_region_payload[.]a$', 'AddressSanitizer: heap-use-after-free'),
 ('uncaught-is-panic','internal/native/runtime/exceptions.c',
  'exit(1);', 'exit(70);', '^TestStep21Uncaught$', 'uncaught lifetime check: exit 70'),
]
# Unknown assertions have two independent barriers: the up-front proof and lowering.
# Mutate both to demonstrate that the admission check rejects an unchecked cast.
assertion_files = [ROOT/'internal/lower/cast_proof.go', ROOT/'internal/lower/cast.go']
originals = [path.read_bytes() for path in assertion_files]
try:
 for path, original in zip(assertion_files, originals):
  source = original.decode()
  anchor = 'as := node.AsAsExpression()'
  assert source.count(anchor) == 1
  if path.name == 'cast_proof.go':
   replacement = anchor + '; if l.checker.GetTypeAtLocation(as.Expression).Flags() & checker.TypeFlagsUnknown != 0 { return castProof{}, nil }'
  else:
   replacement = anchor + '; if l.checker.GetTypeAtLocation(as.Expression).Flags() & checker.TypeFlagsUnknown != 0 { value, err := l.expression(as.Expression); if err != nil { return nil, err }; return ir.Narrow{Value: value, To: ir.Object}, nil }'
  path.write_text(source.replace(anchor,replacement,1))
 log = LOGS/'unknown-authorizes-assertion.log.txt'
 with log.open('w') as output:
  result = subprocess.run(['go','test','./internal/oracle','-run','^TestStep21UnknownAssertions$','-count=1','-v'],cwd=ROOT,stdout=output,stderr=subprocess.STDOUT,timeout=360)
 observed=log.read_text()
 assert result.returncode != 0 and 'want Refused, got <nil>' in observed,observed
 print('unknown-authorizes-assertion: exit 1, caught by admission check',flush=True)
finally:
 for path,original in zip(assertion_files,originals): path.write_bytes(original)
for name, relative, old, new, test, expected in mutants:
 target = ROOT / relative
 original = target.read_bytes()
 source = original.decode()
 assert source.count(old) == 1, (name, source.count(old))
 try:
  target.write_text(source.replace(old,new,1))
  log = LOGS / (name+'.log.txt')
  with log.open('w') as output:
   result = subprocess.run(['go','test','./internal/oracle','-run',test,'-count=1','-v','-timeout','5m'],cwd=ROOT,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=output,stderr=subprocess.STDOUT,timeout=360)
  observed=log.read_text()
  assert result.returncode != 0 and expected in observed, (name,result.returncode,observed)
  assert 'clang failed' not in observed and '[build failed]' not in observed,(name,observed)
  print(f'{name}: exit {result.returncode}, caught by {expected}',flush=True)
 finally:
  target.write_bytes(original)

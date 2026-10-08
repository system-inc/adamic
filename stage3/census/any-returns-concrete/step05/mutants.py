"""Mutate one loader condition at a time and restore the source."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[4]
logs = Path('/tmp/timeout-mutants')
logs.mkdir(exist_ok=True)
guard = 'if symbol == nil || symbol.CheckFlags&ast.CheckFlagsUnresolved != 0 {'
cases = [
 ('ignore-namespace', 'internal/load/node_library.go',
  'if needed {', 'if needed && false {', 'TestNodeNamespaceAnnotationUsesPinnedTimeout',
  'qualified Timeout annotation must load pinned Node declarations'),
 ('ignore-unresolved-alias', 'internal/load/node_library.go',
  guard, 'if symbol == nil {', 'TestNodeNamespaceAnnotationUsesPinnedTimeout',
  'qualified Timeout annotation must load pinned Node declarations'),
 ('replace-local-namespace', 'internal/load/node_library.go',
  guard, 'if symbol != nil || true {', 'TestNodeNamespaceLocalDeclarationIsNotReplaced',
  'a resolved local namespace must not request ambient Node declarations'),
 ('accept-unresolved-type', 'internal/load/load.go',
  'if diagnostics := loaded.diagnostics(context.Background()); len(diagnostics) > 0 {',
  'if diagnostics := loaded.diagnostics(context.Background()); false {',
  'TestNodeNamespaceUnresolvedTypesStayRejected',
  'unresolved type must retain its named checker refusal'),
]
for name, relative, before, after, test, witness in cases:
 path = root / relative
 original = path.read_text()
 assert original.count(before) == 1, (name, 'mutation anchor')
 try:
  path.write_text(original.replace(before, after))
  with (logs / (name + '.log')).open('w') as log:
   result = subprocess.run(['go','test','./internal/load','-run','^'+test+'$','-count=1','-v'], cwd=root, stdout=log, stderr=log)
  output = (logs / (name + '.log')).read_text()
  assert result.returncode != 0 and witness in output and '[build failed]' not in output, (name, output)
  print(name + ': caught by ' + test)
 finally:
  path.write_text(original)

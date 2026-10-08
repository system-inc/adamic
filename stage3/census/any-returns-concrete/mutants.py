"""Run each generic-return mutant independently and restore every source."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/any-returns-mutants')
logs.mkdir(exist_ok=True)
cases = [
 ('unsubstituted-return', 'internal/lower/generic.go',
  'returns := l.concrete(l.checker.GetReturnTypeOfSignature(signature))',
  'returns := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(signature.Declaration()))',
  'TestGenericResolvedReturnContracts', 'memoize resolved return: got () => T'),
 ('trust-any', 'internal/lower/generic.go',
  'if returns.Flags()&checker.TypeFlagsAny != 0 {', 'if false {',
  'TestGenericResolvedReturnContracts', 'any return must stay refused'),
 ('opaque-mapper', 'internal/lower/instantiate.go',
  'type typeMapper = checker.TypeMapper', 'type typeMapper struct{}',
  'TestGenericMapperKeepsCheckerIdentity', 'generic mapper must retain the checker type identity'),
]
for name, relative, before, after, test, witness in cases:
 path = root / relative
 original = path.read_text()
 assert original.count(before) == 1, (name, 'mutation anchor')
 try:
  path.write_text(original.replace(before, after))
  with (logs / (name + '.log')).open('w') as log:
   result = subprocess.run(['go','test','./internal/lower','-run','^'+test+'$','-count=1','-v'], cwd=root, stdout=log, stderr=log)
  output = (logs / (name + '.log')).read_text()
  assert result.returncode != 0 and witness in output and '[build failed]' not in output, (name, output)
  print(name + ': caught by ' + test)
 finally:
  path.write_text(original)

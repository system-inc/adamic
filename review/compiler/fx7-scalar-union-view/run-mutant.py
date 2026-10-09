from pathlib import Path
import difflib
import subprocess

path = Path('internal/native/view_unions_read.go')
original = path.read_text()
before = '\t\te.checkMaybeUnionView(property, value, snapshot)\n'
assert original.count(before) == 1
mutated = original.replace(before, '', 1)
evidence = Path('review/compiler/fx7-scalar-union-view')
(evidence / 'skip-conversion-check.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile='a/'+str(path), tofile='b/'+str(path))))
try:
 path.write_text(mutated)
 with (evidence / 'mutant.log').open('w') as log:
  result = subprocess.run(['go','test','./internal/oracle','-run','^TestScalarUnionViewMisfit','-count=1','-v','-timeout','90s'], stdout=log, stderr=subprocess.STDOUT, timeout=180)
 output = (evidence / 'mutant.log').read_text()
 assert result.returncode == 1 and 'native exit=0' in output and 'sanitized exit=0' in output and 'exit codes differ' in output and '[build failed]' not in output and 'Sanitizer' not in output, output
 print('skip-conversion-check: caught by native and sanitized misfit fixture runtime comparisons; exit', result.returncode)
finally:
 path.write_text(original)

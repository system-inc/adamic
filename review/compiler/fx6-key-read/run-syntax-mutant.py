from pathlib import Path
import difflib
import subprocess
import sys
name, filename, before, after, test = sys.argv[1:]
path = Path(filename)
evidence = Path('review/compiler/fx6-key-read')
original = path.read_text()
assert original.count(before) == 1
mutated = original.replace(before, after, 1)
(evidence / (name + '-bypass.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile='a/'+filename, tofile='b/'+filename)))
try:
 path.write_text(mutated)
 with (evidence / (name + '-mutant.log')).open('w') as log:
  result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^'+test+'$', '-count=1', '-v', '-timeout', '90s'], stdout=log, stderr=subprocess.STDOUT, timeout=180)
 output = (evidence / (name + '-mutant.log')).read_text()
 print(name, 'mutant test exit', result.returncode)
 assert result.returncode != 0 and 'exit codes differ' in output and 'got exit 0' in output, output
 assert 'Sanitizer' not in output, output
finally:
 path.write_text(original)

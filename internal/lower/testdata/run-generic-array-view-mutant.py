from pathlib import Path
import subprocess
root = Path(__file__).resolve().parents[3]
source = root / 'internal/lower/invariance.go'
original = source.read_text()
needle = '\tfrom, to = l.concrete(from), l.concrete(to)'
mutant = '''
    rawFrom, rawTo := l.checker.GetTypeArguments(from), l.checker.GetTypeArguments(to)
    if len(rawFrom) == 1 && len(rawTo) == 1 && l.deferredNullableArrayView(from, to, rawFrom[0], rawTo[0]) { return nil }
'''
assert original.count(needle) == 1
try:
 source.write_text(original.replace(needle,mutant+needle))
 log = Path('/tmp/generic-array-bypass-mutant.log')
 with log.open('w') as output:
  result = subprocess.run(['go','test','./internal/lower','-run','^TestGenericNullableArrayViewRefused$','-count=1'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
 text=log.read_text()
 assert result.returncode != 0 and 'want pinned call-site nullable type argument refusal, got <nil>' in text, text
 print('accept-without-instantiation-comparison: caught; nullable fixture incorrectly lowered, pinned refusal test failed')
finally:
 source.write_text(original)

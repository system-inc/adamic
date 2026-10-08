#!/usr/bin/env python3
"""Restore the legacy preflight independently for each deferred parser family."""
import pathlib, subprocess
root = pathlib.Path(__file__).resolve().parents[3]
path = root / 'internal/lower/view_cast_preflight.go'
logs = root / 'stage3/interface-downcasts/lane2/parser-logs'
logs.mkdir(exist_ok=True)
original = path.read_bytes()
cases = [
 ('views-preflight', 'return l.viewCastCandidate(source, target)', 'TestCheckedViewGenericArrayCast'),
 ('phantom-preflight', 'return l.phantomArrayBase(source) != nil || l.phantomArrayBase(target) != nil || l.phantomCast(source, target)', 'TestPhantomSortedArrayProbe'),
]
for name, before, test in cases:
 assert original.decode().count(before) == 1
 try:
  path.write_text(original.decode().replace(before, 'return false'))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^'+test+'$', '-count=1'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in observed and '[build failed]' not in observed, observed
  assert "a cast the runtime can't check" in observed and 'adamic/no-unchecked-cast' in observed, observed
  print(name + ': caught by ' + test, flush=True)
 finally:
  path.write_bytes(original)

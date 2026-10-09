#!/usr/bin/env python3
"""Prove the finite signature, null distinction and dispatch controls can fail."""
from pathlib import Path
import json, os, subprocess, tempfile
root = Path(__file__).resolve().parents[2]
logs = Path('/tmp/area-stack-clock-generic-guard-mutants')
logs.mkdir(exist_ok=True)
proof = root / 'internal/lower/clock_generic_returns_t_01.go'
source = proof.read_text()
functions = root / 'internal/lower/functions.go'
function_source = functions.read_text()
hook = '\t\t\tif !isKnown {\n\t\t\t\tvalueType, isKnown = l.clockGenericReturnsT01(returns)\n\t\t\t}\n'
assert function_source.count(hook) == 1
mutants = [
 ('skip-finite-shape', proof, source.replace('if !l.clockGenericReturnsT01Shape(object, map[*checker.Type]bool{}) {', 'if false {', 1), '^TestClockGenericReturnsT01', 'indexed object admitted by the finite-shape proof'),
 ('erase-null-distinction', proof, source.replace('len(result.Types()) != 2', 'len(result.Types()) < 2', 1).replace('} else {\n\t\t\tobject = member', '} else if member.Flags()&checker.TypeFlagsNull == 0 {\n\t\t\tobject = member', 1), '^TestClockGenericReturnsT01RejectsNullBeforeBody$', 'null and undefined admitted with one representation'),
 ('remove-signature-dispatch', functions, function_source.replace(hook, '', 1), '^TestClockGenericReturnsT01Shapes/nested_readonly_fields$', 'a function returning'),
]
with tempfile.TemporaryDirectory(prefix='clock-generic-mutants-', dir=os.environ.get('TMPDIR')) as scratch:
 for name, path, changed, pattern, catch in mutants:
  assert changed != path.read_text(), name
  replacement = Path(scratch) / (name + '.go')
  replacement.write_text(changed)
  overlay = Path(scratch) / (name + '.json')
  overlay.write_text(json.dumps({'Replace': {str(path): str(replacement)}}))
  log = logs / (name + '.log')
  with log.open('w') as output:
   result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/lower', '-run', pattern, '-count=1', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  output = log.read_text()
  assert result.returncode != 0 and catch in output and 'build failed' not in output, name + ' not caught: ' + output
  print(name + ': caught by ' + catch, flush=True)

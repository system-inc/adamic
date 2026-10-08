#!/usr/bin/env python3
"""Metadata checks only. These do not claim the three runtime selection mutants."""
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
source = root / 'internal/lower/view_unions_mixed.go'
original = source.read_text()
logs = pathlib.Path(__file__).resolve().parent / 'logs'
mutants = [
 ('omit-member', 'contract.Members = append(contract.Members, child)', '_ = child', 'TestMixedUnionContractGraph', 'union lost a member contract'),
 ('accept-unknown', ' || l.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown', '', 'TestMixedUnionContractUnknownMemberFails', 'unavailable member certified'),
 ('keep-failed-graph', 'if err == nil {', 'if true {', 'TestMixedUnionContractFailureDoesNotCertifyRetry', 'failed graph remains memoized as certified'),
]
try:
 for name, before, after, test, catcher in mutants:
  assert original.count(before) == 1, name
  source.write_text(original.replace(before, after))
  with (logs / ('contract-mutant-' + name + '.log')).open('w') as log:
   result = subprocess.run(['go', 'test', './internal/lower', '-run', '^' + test + '$', '-count=1', '-timeout', '10m'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
  output = (logs / ('contract-mutant-' + name + '.log')).read_text()
  assert result.returncode == 1 and catcher in output, (name, result.returncode, output)
  print(name + ': caught by ' + catcher)
  source.write_text(original)
finally:
 source.write_text(original)

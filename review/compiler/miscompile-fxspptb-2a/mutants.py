#!/usr/bin/env python3
"""Revert each narrow stop separately; its pinned test must fail."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / 'internal/lower/known_lowering_gaps.go'
original = source.read_text()
probes = [
 ('error-spread', 'case ast.KindSpreadAssignment:', 'case ast.KindUnknown:', '(ErrorSpread|ErrorSubclassSpread)'),
 ('optional-error', 'case ast.KindNewExpression:', 'case ast.KindUnknown:', 'OptionalError'),
 ('maybe-setter', 'case ast.KindSetAccessor:', 'case ast.KindUnknown:', 'MaybeSetter'),
 ('long-name', 'len(plainPropertyName(node)) > 4095', 'len(plainPropertyName(node)) > 1000000', 'LongName'),
 ('computed-name', 'case ast.KindClassDeclaration:', 'case ast.KindUnknown:', '(DerivedSymbol|OverrideSource)'),
 ('constructor-capture', 'case ast.KindThisKeyword:', 'case ast.KindUnknown:', '(ConstructorArrow|ConstructorNumberArrow)'),
]
try:
 for label, before, after, selector in probes:
  assert original.count(before) == 1, label
  source.write_text(original.replace(before, after))
  log = Path(__file__).resolve().parent / ('mutant-' + label + '.log')
  with log.open('w') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestMiscompile2A' + selector + '$', '-count=1', '-timeout', '90s', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  text = log.read_text()
  assert result.returncode != 0, label + ' survived'
  assert 'want path-bearing stop' in text or 'panic: Unhandled case in Node.Text' in text, label + ' failed outside its pin'
  print(label + ': caught by pinned stop; ' + str(log), flush=True)
finally:
 source.write_text(original)

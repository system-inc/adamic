#!/usr/bin/env python3
"""Revert each containment guard or Error fix separately; its pinned test must fail."""
from pathlib import Path
import subprocess
import sys
import os

root = Path.cwd()
source = root / 'internal/lower/known_lowering_gaps.go'
original = source.read_text()
probes = [
 ('error-spread', 'case ast.KindSpreadAssignment:', 'case ast.KindUnknown:', '(ErrorSpread|ErrorSubclassSpread)'),
 ('maybe-setter', 'case ast.KindSetAccessor:', 'case ast.KindUnknown:', 'MaybeSetter'),
 ('long-name', 'len(plainPropertyName(node)) > 4095', 'len(plainPropertyName(node)) > 1000000', 'LongName'),
 ('computed-name', 'case ast.KindClassDeclaration:', 'case ast.KindUnknown:', '(DerivedSymbol|OverrideSource)'),
 ('constructor-capture', 'case ast.KindThisKeyword:', 'case ast.KindUnknown:', '(ConstructorArrow|ConstructorNumberArrow)'),
]
try:
 for label, before, after, selector in ([] if '--optional-only' in sys.argv else probes):
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

# Restore the old null native storage. Both the supplied witness and the census
# program's omitted-argument driver must fail under sanitizers, not at clang.
source = root / 'internal/native/runtime/exceptions.c'
original = source.read_text()
before = 'adamic_retain(message == NULL ? &empty_message : message)'
after = 'adamic_retain(message)'
assert original.count(before) == 1
try:
 source.write_text(original.replace(before, after).replace('static adamic_string empty_message = ADAMIC_STRING("");\n', ''))
 log = Path(__file__).resolve().parent / 'mutant-optional-error.log'
 with log.open('w') as output:
  for test in ['TestMiscompile2AOptionalError', 'TestMiscompile2AMarkerUndefined']:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^' + test + '$', '-count=1', '-timeout', '90s', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT, env={**os.environ, 'UBSAN_OPTIONS': 'symbolize=0', 'ADAMIC_GATE_UNCACHED': '1'})
   assert result.returncode != 0, test + ' survived'
 text = log.read_text()
 assert 'runtime error:' in text and 'member access within null pointer' in text, 'optional-error failed outside the native read'
 for test in ['TestMiscompile2AOptionalError', 'TestMiscompile2AMarkerUndefined']:
  assert '--- FAIL: ' + test in text, test + ' did not catch the old read'
 print('optional-error: old native null read caught by supplied witness and census omitted-argument driver; ' + str(log), flush=True)
finally:
 source.write_text(original)

# Giving an undefined message an own public field must fail Node parity even
# when the fallback string itself is correct and all sanitizer checks pass.
source = root / 'internal/native/runtime/exceptions.c'
original = source.read_text()
before = 'adamic_object_new(message == NULL ? &absent_message_shape : &error_shape)'
after = 'adamic_object_new(&error_shape)'
assert original.count(before) == 1
try:
 source.write_text(original.replace(before, after))
 log = Path(__file__).resolve().parent / 'mutant-error-own-message.log'
 with log.open('w') as output:
  result = subprocess.run(['go', 'test', './internal/native', '-run', '^TestErrorUndefinedMessageHasNoOwnProperty$', '-count=1', '-timeout', '90s', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
 text = log.read_text()
 assert result.returncode != 0, 'error-own-message survived'
 assert 'native ' in text and 'Node ' in text and '[] true true' in text, 'error-own-message failed outside Node parity'
 print('error-own-message: incorrect own presence caught by runtime Node parity; ' + str(log), flush=True)
finally:
 source.write_text(original)

#!/usr/bin/env python3
"""Semantic mutants for nullable returns and readonly conditional consumers."""
import pathlib, subprocess, sys
root = pathlib.Path(__file__).resolve().parents[3]
logs = root / 'stage3/interface-downcasts/lane2/parser-return-logs'
logs.mkdir(exist_ok=True)
cases = [
 ('nullable-dispatch', 'internal/lower/view_cast_preflight.go', ' || l.nullableReadonlyArrayCast(node, source, target)', '', './internal/oracle', 'TestCheckedViewParserReturns/native-same-map-return-cast', "a cast the runtime can't check"),
 ('readonly-consumer', 'internal/lower/invariance.go', 'if target := l.readonlyArrayConsumer(parent); target != nil {\n\t\t\t\treturn target\n\t\t\t}', '', './internal/oracle', 'TestCheckedViewParserReturns/native-sorted-empty-conditional', 'adamic/invariant-mutable'),
 ('mutable-admission', 'internal/lower/invariance.go', 'mutable := !l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(to) && to.TargetTupleType().IsReadonly())', 'mutable := false', './internal/lower', 'TestViewArrayConsumerRefusals', 'want mutable target refusal, got <nil>'),
 ('string-element-check', 'internal/native/runtime/view_arrays.c', 'adamic_value *slot = checked ? adamic_view_array_at', 'adamic_value *slot = false && checked ? adamic_view_array_at', './internal/oracle', 'TestCheckedViewArrays/nullable-array-string', '--- FAIL:'),
 ('string-literal-check', 'internal/native/runtime/view_arrays.c', 'if (slot != NULL && checked && allowed_count != 0)', 'if (false && slot != NULL && checked && allowed_count != 0)', './internal/oracle', 'TestCheckedViewArrays/array-string-literal', '--- FAIL:'),
 ('undefined-reference-sort', 'internal/lower/view_array_consumers.go', 'if element == ir.String && l.includesUndefined(declared)', 'if false && element == ir.String && l.includesUndefined(declared)', './internal/lower', 'TestViewArrayOptionalComparatorRefusal', 'want undefined-reference comparator stop, got <nil>'),
 ('numeric-default-sort', 'internal/native/runtime/view_arrays.c', 'int result = adamic_string_compare(a, b);', 'int result = element == 1 ? (left.number < right.number ? -1 : left.number > right.number ? 1 : 0) : adamic_string_compare(a, b);', './internal/oracle', 'TestCheckedViewParserReturns/optional-sort', '--- FAIL:'),
]
for name, relative, before, after, package, test, marker in cases:
 if len(sys.argv) > 1 and name not in sys.argv[1:]: continue
 path = root / relative
 original = path.read_bytes()
 assert original.decode().count(before) == 1
 try:
  path.write_text(original.decode().replace(before, after))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   result = subprocess.run(['go', 'test', package, '-run', '^'+test+'$', '-count=1', '-timeout', '10m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in observed and '[build failed]' not in observed and 'clang failed' not in observed and marker in observed, observed
  print(name + ': caught by ' + test, flush=True)
 finally:
  path.write_bytes(original)

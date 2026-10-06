#!/usr/bin/env python3
"""Run one compiler mutant at a time, restore sources, and reject build-only failures."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
report = Path(__file__).resolve().parent
expression = 'internal/lower/expression.go'
prelude = 'internal/lower/prelude.go'
number = 'internal/lower/library_math_number.go'
json = 'internal/native/runtime/json_stringify.c'
nullable = 'internal/lower/nullable.go'
generic = 'internal/lower/generic.go'
refusals = 'internal/lower/refusals.go'
files = {name: (root / name).read_text() for name in [expression, prelude, number, json, nullable, generic, refusals]}

oracle = 'TestNativeAgreesWithNode/internal/oracle/testdata/nullable_references'
mutants = [
 ('console_null_as_undefined', {prelude: [('value = l.spelled(arguments[0], value)', 'if value.Type() == ir.String && l.includesNull(l.checker.GetTypeAtLocation(arguments[0])) { value = ir.Coalesce{Value: value, Fallback: ir.StringConstant{Index: l.constant("undefined")}, Of: ir.String} } else { value = l.spelled(arguments[0], value) }')]}, oracle, 'stdout differs'),
 ('typeof_null_as_undefined', {expression: [('l.nullableObservation("typeof", operand, ir.StringConstant{Index: l.constant("object")}', 'l.nullableObservation("typeof", operand, ir.StringConstant{Index: l.constant("undefined")}')]}, oracle, 'stdout differs'),
 ('number_null_as_nan', {number: [('l.nullableObservation("number", value, ir.NumberConstant{}', 'l.nullableObservation("number", value, ir.NumberConstant{Value: math.NaN()}')]}, oracle, 'stdout differs'),
 ('json_omits_null', {json: [('return (json_scalar){adamic_json_null, value};', 'return (json_scalar){adamic_json_undefined, value};')]}, oracle, 'stdout differs'),
 ('generic_undefined_true_for_null', {expression: [('test = ir.IsNull{Value: value, AlwaysFalse: true}', 'test = ir.IsUndefined{Value: value}')]}, oracle, 'stdout differs'),
 ('generic_null_true_for_undefined', {expression: [('AlwaysFalse: !l.includesNull(l.checker.GetTypeAtLocation(operand))', 'AlwaysFalse: !l.includesNull(l.checker.GetTypeAtLocation(operand)) && false')]}, oracle, 'stdout differs'),
 ('generic_keys_merge_nullable_strings', {generic: [('return l.checker.TypeToString(proven)', 'if l.checker.GetNonNullableType(proven).Flags()&checker.TypeFlagsStringLike != 0 { return "nullable_string" }; return l.checker.TypeToString(proven)')]}, oracle, 'stdout differs'),
 ('ignore_nullable_views', {nullable: [('if !l.nullableViewsMatch(own, contextual, map[[2]*checker.Type]bool{}) {', 'if false && !l.nullableViewsMatch(own, contextual, map[[2]*checker.Type]bool{}) {')]}, 'TestNullableReferenceViewsCannotChangeTheEmptyCase', 'got <nil>'),
 ('empty_string_is_truthy', {expression: [('Left: ir.StringLength{Value: read}, Right: ir.NumberConstant{}', 'Left: ir.StringLength{Value: read}, Right: ir.NumberConstant{Value: -1}')]}, oracle, 'stdout differs'),
 ('admit_two_empty_cases', {
  expression: [('l.includesNull(proven) && l.includesUndefined(proven)', 'false && l.includesNull(proven) && l.includesUndefined(proven)'), ('if l.includesNull(proven) {', 'if false && l.includesNull(proven) {')],
  nullable: [('if l.includesNull(own) && l.includesUndefined(own) {', 'if false && l.includesNull(own) && l.includesUndefined(own) {')],
  generic: [('if l.includesNull(concrete) && l.includesUndefined(concrete) {', 'if false && l.includesNull(concrete) && l.includesUndefined(concrete) {')],
  refusals: [('if l.includesNull(proven) && l.includesUndefined(proven) {', 'if false && l.includesNull(proven) && l.includesUndefined(proven) {')],
 }, 'TestNullableReferencesNeedAnEmptyCaseTag', 'got <nil>'),
]

rows = []
try:
 for name, changes, test, evidence in mutants:
  for path, replacements in changes.items():
   source = files[path]
   for before, after in replacements:
    if before not in source: raise RuntimeError(f'{name}: replacement missing in {path}: {before}')
    source = source.replace(before, after)
   (root / path).write_text(source)
  package = './internal/lower' if test.startswith('TestNullable') else './internal/oracle'
  command = ['go', 'test', package, '-run', test, '-count=1', '-timeout', '10m']
  with (report / (name + '.log')).open('w') as log:
   result = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT)
  output = (report / (name + '.log')).read_text()
  for path in changes: (root / path).write_text(files[path])
  if result.returncode == 0 or evidence not in output or '[build failed]' in output or 'clang failed' in output:
   raise RuntimeError(f'{name}: not killed by the intended check, see {name}.log')
  row = f'{name}: caught by {test}: {evidence}'
  rows.append(row)
  print(row, flush=True)
finally:
 for path, source in files.items(): (root / path).write_text(source)
(report / 'mutants.txt').write_text('\n'.join(rows) + '\n')

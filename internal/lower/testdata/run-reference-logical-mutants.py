from pathlib import Path
import os, subprocess, sys
root = Path(__file__).resolve().parents[3]
mutants = [
 ('trace-refused-fixtures', 'internal/flow/flow_test.go', 'refusedOracleFixture(filepath.Base(path)) || ', '', './internal/flow', '^TestEveryFunctionIsInSingleAssignment$', 'Lower:'),
 ('treat-structural-primitives-as-references', 'internal/lower/reference_logical.go', 'if of == ir.Object &&', 'if false && of == ir.Object &&', './internal/lower', '^TestReferenceLogicalStructuralPrimitivesStayOnExistingPaths$', 'got <nil>'),
 ('borrow-and-as-coalesce', 'internal/native/element_borrow.go', 'coalesce.ReferenceAnd || coalesce.Panic', 'coalesce.Panic', './internal/oracle', '^TestNativeAgreesWithNode/internal/oracle/testdata/reference_logical.a$', 'stdout'),
 ('lose-narrowing-check', 'internal/lower/narrowed.go', '&& (l.includesUndefined(whole) || l.includesNull(whole))', '&& (whole != nil)', './internal/oracle', '^TestNativeAgreesWithNode/internal/oracle/testdata/reference_logical_narrowed.a$', 'want the inserted check to fire'),
 ('collapse-lookup-absences', 'internal/lower/reference_logical.go', 'if !l.includesNull(proven) ||', 'if true || !l.includesNull(proven) ||', './internal/lower', '^TestNullableReferenceLookupNeedsTag$', 'got <nil>'),
 ('use-string-presence', 'internal/lower/reference_logical.go', 'of == ir.Array || of == ir.Object', 'of == ir.String || of == ir.Array || of == ir.Object', './internal/lower', '^TestReferenceLogicalPrimitivesStayOnExistingPaths$', 'got <nil>'),
 ('double-left', 'internal/native/emit_branches.go', 'if coalesce.ReferenceAnd {\n\t\treturn e.referenceAnd(coalesce, value)', 'if coalesce.ReferenceAnd {\n\t\te.value(coalesce.Value)\n\t\treturn e.referenceAnd(coalesce, value)', './internal/oracle', '^TestNativeAgreesWithNode/internal/oracle/testdata/reference_logical.a$', 'stdout'),
 ('empty-array-falsy', 'internal/native/emit_branches.go', 'e.line("if (%s != NULL) {", value)', 'if expression.Value.Type() == ir.Array { e.line("if (%s != NULL && %s->length > 0) {", value, value) } else { e.line("if (%s != NULL) {", value) }', './internal/oracle', '^TestNativeAgreesWithNode/internal/oracle/testdata/reference_logical.a$', 'stdout'),
 ('lose-and-view', 'internal/lower/reference_logical.go', 'return binary.Left == node &&', 'return false && binary.Left == node &&', './internal/oracle', '^TestNativeAgreesWithNode/internal/oracle/testdata/reference_logical.a$', 'refuses'),
 ('accept-writable-result', 'internal/lower/invariance.go', 'if l.referenceAndTest(node) {', 'if node.Kind == ast.KindIdentifier || node.Kind == ast.KindBinaryExpression || l.referenceAndTest(node) {', './internal/lower', '^TestReferenceLogicalKeepsMutableViews$', 'want mutable view refusal, got <nil>'),
]
for name, file, old, new, package, test, catcher in mutants:
 if len(sys.argv) > 1 and name not in sys.argv[1:]: continue
 source = root / file
 original = source.read_text()
 assert original.count(old) == 1, (name, original.count(old))
 log = Path('/tmp/reference-logical-mutant-' + name + '.log')
 try:
  source.write_text(original.replace(old, new))
  with log.open('w') as output:
   result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-timeout', '30m'], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
  text = log.read_text()
  assert result.returncode != 0 and catcher in text, (name, text)
  assert 'clang:' not in text and 'runtime error:' not in text, (name, text)
  print(name + ': caught by ' + catcher, flush=True)
 finally:
  source.write_text(original)

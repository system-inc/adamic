#!/usr/bin/env python3
"""Run each computed-name mutant through an isolated source overlay."""
from pathlib import Path
import os
import json
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/destructuring-computed-mutants')
logs.mkdir(exist_ok=True)
helper = 'internal/lower/computed_field_name.go'
mutants = [
 ('string-key', helper, 'return left + right, true', 'return left + "wrong" + right, true', 'oracle', 'stdout differs'),
 ('enum-key', helper, "strconv.FormatFloat(value.Value, 'f', -1, 64)", "strconv.FormatFloat(value.Value + 1, 'f', -1, 64)", 'oracle', 'stdout differs'),
 ('binding-key', 'internal/lower/collections.go', 'field, known = l.constantFieldName(declared.PropertyName.AsComputedPropertyName().Expression)', 'field, known = l.constantFieldName(declared.PropertyName.AsComputedPropertyName().Expression)\n if field == "-2" { field = "3" }', 'oracle', 'stdout differs'),
 ('runtime-key-evaluation', 'internal/lower/object.go', 'value = ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: key, Right: constant}, WhenTrue: value, WhenNot: value, Of: value.Type()}', 'value = ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: constant, Right: constant}, WhenTrue: value, WhenNot: value, Of: value.Type()}', 'oracle', 'stdout differs'),
 ('own-proto', 'internal/javascript/javascript.go', 'if field.Name == "__proto__" {', 'if field.Name == "__proto__" && false {', 'oracle', 'stdout differs'),
 ('singleton-call', helper, 'node = ast.SkipParentheses(node)', 'node = ast.SkipParentheses(node)\n if node.Kind == ast.KindCallExpression { return "label", true }', 'lower', 'want the computed-field gap'),
 ('numeric-concatenation', helper, 'leftKnown && rightKnown && l.checker.GetTypeAtLocation(binary.Left).Flags()&checker.TypeFlagsStringLike != 0 && l.checker.GetTypeAtLocation(binary.Right).Flags()&checker.TypeFlagsStringLike != 0', 'leftKnown && rightKnown && (true || l.checker.GetTypeAtLocation(binary.Left).Flags()&checker.TypeFlagsStringLike != 0 || l.checker.GetTypeAtLocation(binary.Right).Flags()&checker.TypeFlagsStringLike != 0)', 'lower', 'want the computed-field gap'),
]
for name, file, before, after, package, catcher in mutants:
 path = root / file
 original = path.read_text()
 assert original.count(before) == 1, (name, original.count(before))
 source = logs / (name + '.go')
 source.write_text(original.replace(before, after, 1))
 overlay = logs / (name + '.json')
 overlay.write_text(json.dumps({'Replace': {str(path): str(source)}}))
 pattern = '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_computed_' if package == 'oracle' else '^TestComputedFieldNameGapsStayExplicit$'
 if name == 'enum-key': pattern = '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_computed_numeric_names.a$'
 with (logs / (name + '.log')).open('w') as log:
  result = subprocess.run(['go','test','-overlay='+str(overlay),'./internal/'+package,'-run',pattern,'-count=1','-timeout','10m'],cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=log,stderr=subprocess.STDOUT)
 output = (logs / (name + '.log')).read_text()
 assert result.returncode != 0 and catcher in output and '[build failed]' not in output, (name,result.returncode,output)
 print(name + ': caught by ' + catcher,flush=True)

#!/usr/bin/env python3
"""Run focused overload mutants through scratch Go overlays, never editing production."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[2]
source = root / 'internal/lower/overload_results.go'
original = source.read_text()
evidence = Path(__file__).parent / 'evidence'
key_start = original.index('\tkey := fmt.Sprintf("overload-result:')
key_end = original.index('\tfor _, argument := range invoked.Arguments', key_start)
mutations = [
 ('trust-optional-property', original.replace('sourceProperty == nil || sourceProperty.Flags&ast.SymbolFlagsOptional != 0 || accessorSymbol(sourceProperty)', 'sourceProperty == nil || accessorSymbol(sourceProperty)'), './internal/lower', '^TestOverloadResultsRefuses$/optional_result_property$'),
 ('trust-result', original.replace('if !l.censusProveOverloadResult(implementation, overload) && !l.censusRelated(produced, promised) {', 'if false && !l.censusProveOverloadResult(implementation, overload) && !l.censusRelated(produced, promised) {'), './internal/lower', '^TestOverloadResultsLiarStops$'),
 ('drop-specialization', original[:key_start] + '\tkey := fmt.Sprintf("overload-result:%d:%d:erased", function, ordinal)\n' + original[key_end:], './internal/oracle', '^TestNativeAgreesWithNode$/internal/oracle/testdata/overload_results_binding.a$'),
 ('drop-parameter-check', original.replace('checked.Arguments[position] = ir.Coalesce{Value: argument, Of: held, Panic: ir.StringConstant{Index: l.constant(message)}}', '_ = message\nchecked.Arguments[position] = fit(argument, held)'), './internal/lower', '^TestOverloadResultsLiarStops$/parameter-liar$'),
 ('trust-shared-result', original.replace('if !l.overloadFreshResult(implementation, map[*ast.Node]bool{}) {', 'if false && !l.overloadFreshResult(implementation, map[*ast.Node]bool{}) {'), './internal/lower', '^TestOverloadResultsRefuses$'),
]
results = []
for name, changed, package, selector in mutations:
 assert changed != original
 with tempfile.TemporaryDirectory(prefix='overload-mutant-') as scratch:
  scratch = Path(scratch)
  replacement = scratch / 'overload_results.go'
  replacement.write_text(changed)
  overlay = scratch / 'overlay.json'
  overlay.write_text(json.dumps({'Replace': {str(source): str(replacement)}}))
  command = ['go', 'test', '-overlay', str(overlay), package, '-run', selector, '-count=1', '-v', '-timeout', '10m']
  with (evidence / (name + '.log')).open('w') as log:
   result = subprocess.run(command, cwd=root, env={**os.environ, 'ADAMIC_GATE_UNCACHED':'1'}, stdout=log, stderr=subprocess.STDOUT)
  output = (evidence / (name + '.log')).read_text()
  caught = result.returncode != 0 and '--- FAIL:' in output and '[build failed]' not in output
  results.append({'mutant':name, 'selector':selector, 'exit':result.returncode, 'caught':caught})
  print(name, 'caught' if caught else 'NOT CAUGHT', flush=True)
(evidence / 'mutants.json').write_text(json.dumps(results, indent=2)+'\n')
if not all(row['caught'] for row in results):
 raise SystemExit(1)

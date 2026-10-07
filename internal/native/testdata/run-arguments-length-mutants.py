#!/usr/bin/env python3
"""Run actual-count faults against the external Node oracle, restoring every source."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
os.chdir(root)
mutants = [
 ('native-declared-count', 'internal/native/emit_functions.go', 'count := fmt.Sprint(len(call.Arguments))', 'count := fmt.Sprint(len(function.Parameters))', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length.a'),
 ('native-function-value-count', 'internal/native/emit_functions.go', 'count := fmt.Sprint(len(expression.Arguments))', 'count := "0"', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length_value_count.a'),
 ('javascript-declared-count', 'internal/javascript/javascript.go', 'count = values.length - (fn.adamicReceiver ? 1 : 0)', 'count = fn.length - (fn.adamicCount ? 1 : 0) - (fn.adamicReceiver ? 1 : 0)', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length.a'),
 ('native-ignore-count-reader-type', 'internal/native/emit_functions.go', 'if !e.program.ClosureReadsArgumentsCount(expression) && !e.program.ClosureNeedsArgumentSlots(expression) {', 'if !e.program.ClosureNeedsArgumentSlots(expression) {', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length_value_count.a'),
 ('javascript-function-value-count', 'internal/javascript/javascript.go', 'closure.code(closure, values, values.length)', 'closure.code(closure, values, 0)', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length_value_count.a'),
]
for name, file, before, after, test in mutants:
 path = root / file
 original = path.read_text()
 if original.count(before) != 1:
  raise RuntimeError(f'{name}: mutation anchor is not unique')
 log = Path('/tmp/arguments-length-' + name + '.log')
 try:
  path.write_text(original.replace(before, after))
  with log.open('w') as output:
   run = subprocess.run(['go','test','./internal/oracle','-run',test,'-count=1','-timeout','30m'], stdout=output, stderr=subprocess.STDOUT, env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'})
  observation = log.read_text()
  if run.returncode == 0 or 'stdout differs' not in observation:
   raise RuntimeError(f'{name}: not killed by observable output; see {log}')
  print(f'{name}: killed by Node stdout disagreement ({log})', flush=True)
 finally:
  path.write_text(original)

# Refusal messages and conservative unknown targets are also required evidence.
checks = [
 ('arrow-refusal', 'internal/lower/arguments_length.go', 'if owner.Kind == ast.KindArrowFunction {', 'if false {', './internal/lower', 'TestArgumentsLengthRefusals/arrow', 'want pinned refusal'),
 ('write-refusal', 'internal/lower/arguments_length.go', 'if !ast.IsAssignmentTarget(length) {', 'if true {', './internal/lower', 'TestArgumentsLengthRefusals/writing_length', 'want pinned refusal'),
 ('unknown-count-target', 'internal/ir/call_targets.go', 'for _, function := range p.Functions {\n\t\tif function.ArgumentsCount != 0 {', 'for _, function := range p.Functions {\n\t\tif false && function.ArgumentsCount != 0 {', './internal/ir', 'TestClosureArgumentsCountTargets', 'type 12: false, want true'),
]
for name, file, before, after, package, test, evidence in checks:
 path = root / file
 original = path.read_text()
 if original.count(before) != 1:
  raise RuntimeError(f'{name}: mutation anchor is not unique')
 log = Path('/tmp/arguments-length-' + name + '.log')
 try:
  path.write_text(original.replace(before, after))
  with log.open('w') as output:
   run = subprocess.run(['go','test',package,'-run',test,'-count=1','-timeout','30m'], stdout=output, stderr=subprocess.STDOUT)
  if run.returncode == 0 or evidence not in log.read_text():
   raise RuntimeError(f'{name}: not killed by required check; see {log}')
  print(f'{name}: killed by {test} ({log})', flush=True)
 finally:
  path.write_text(original)

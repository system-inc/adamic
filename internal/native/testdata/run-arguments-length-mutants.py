#!/usr/bin/env python3
"""Run actual-count faults against the external Node oracle, restoring every source."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
os.chdir(root)
mutants = [
 ('native-declared-count', 'internal/native/emit_functions.go', 'count := fmt.Sprint(len(call.Arguments))', 'count := fmt.Sprint(len(function.Parameters))', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length.a'),
 ('native-function-value-count', 'internal/native/arguments_length.go', 'return e.closureSlots(call, slots, "", fmt.Sprint(len(call.Arguments)))', 'return e.closureSlots(call, slots, "", "0")', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length_value_count.a'),
 ('javascript-declared-count', 'internal/javascript/javascript.go', 'count = values.length - (fn.adamicReceiver ? 1 : 0)', 'count = fn.length - (fn.adamicCount ? 1 : 0) - (fn.adamicReceiver ? 1 : 0)', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length.a'),
 ('javascript-function-value-count', 'internal/javascript/javascript.go', 'count := "values.length"', 'count := "0"', 'TestNativeAgreesWithNode/internal/oracle/testdata/arguments_length_value_count.a'),
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
 ('unknown-count-target', 'internal/ir/argument_slots.go', 'if bounded.Unknown {', 'if false {', './internal/ir', 'TestArgumentLayouts', "unknown type must include"),
 ('ignore-count-reader-type', 'internal/ir/argument_slots.go', 'layout.Count = layout.Count || function.ReadsArguments', 'layout.Count = false', './internal/ir', 'TestArgumentLayouts', 'reader layout'),
 ('nonreader-count', 'internal/native/arguments_length.go', 'func (e *emitter) closureSlots(call ir.CallClosure, slots []string, source, count string) string {\n\tlayout := e.program.ClosureArgumentLayout(call)', 'func (e *emitter) closureSlots(call ir.CallClosure, slots []string, source, count string) string {\n\tlayout := e.program.ClosureArgumentLayout(call)\n\tlayout.Count = true', './internal/native', 'TestNoReaderCallingConvention', 'hidden count slots: got'),
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

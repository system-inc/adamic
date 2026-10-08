#!/usr/bin/env python3
"""Compile isolated source mutants with Go overlays, preserving repository files."""
import argparse
import difflib
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('names', nargs='*')
args = parser.parse_args()
root = Path(__file__).resolve().parents[3]
unit = Path(__file__).resolve().parent
scratch = Path('/tmp/coverage-closure-mutants')
scratch.mkdir(exist_ok=True)
mutants = [
 ('count-zero', 'internal/native/arguments_length.go', [
  ('return e.closureSlots(call, slots, "", fmt.Sprint(len(call.Arguments))), fmt.Sprint(len(call.Arguments))',
   'return e.closureSlots(call, slots, "", fmt.Sprint(len(call.Arguments))), "0"')],
  './internal/native', 'TestNoReaderCallingConvention', ['count_default_evaluation']),
 ('packed-leak', 'internal/native/arguments_length.go', [
  ('packed := e.own(ir.Array, "adamic_array_new(0, false)")',
   'packed := e.temporary()\n\te.line("adamic_array *%s = adamic_array_new(0, false);", packed)')],
  './internal/native', 'TestNoReaderCallingConvention', ['nested_argument_counts']),
 ('mixed-tuple', 'internal/lower/arguments_length.go', [
  ('for position, provenElement := range elements {', 'for _, provenElement := range elements {'),
  ('!known || slotless(of) || position > 0 && of != element', '!known || slotless(of)')],
  './internal/lower', '', ['mixed_tuple_spread_guard']),
 ('delayed-fill', 'internal/lower/library_method_values.go', [
  ('\t\t\tif ast.SkipParentheses(receiver).Kind == ast.KindNewExpression && method.name == "fill" {\n\t\t\t\treturn nil, true, l.notYet(node, "delayed fill of an array with holes; write the direct filled construction")\n\t\t\t}\n', '')],
  './internal/lower', '', ['delayed_fill_guard']),
 ('unnamed-nested', 'internal/lower/nested_functions.go', [
  ('\t\tif node.Name() == nil {\n\t\t\treturn nil, l.notYet(node, "a generic or unnamed nested function declaration")\n\t\t}\n', '')],
  './internal/lower', '', []),
 ('bind-presence', 'internal/lower/library_method_values.go', [
  ('receiver.Type() != ir.String || l.mayBeUndefined(written[0])', 'receiver.Type() != ir.String')],
  './internal/lower', '', ['bound_receiver_escape']),
 ('spread-snapshot', 'internal/native/arguments_length.go', [
  ('snapshot := e.own(ir.Array, fmt.Sprintf("adamic_array_slice(%s, 0, 0, false)", source))', 'snapshot := source')],
  './internal/native', 'TestNoReaderCallingConvention', ['nested_argument_counts']),
 ('rest-retain', 'internal/native/arguments_length.go', [
  ('{.reference = adamic_retain(%s[%s].reference)}', '{.reference = %s[%s].reference}')],
  './internal/native', 'TestNoReaderCallingConvention', ['rest_capture_lifetime', 'nested_argument_counts']),
]
records = {}
for name, path, replacements, package, pattern, probes in mutants:
 if args.names and name not in args.names: continue
 folder = scratch / name
 folder.mkdir(exist_ok=True)
 original = (root / path).read_text()
 mutated = original
 for before, after in replacements:
  assert mutated.count(before) == 1, (name, before, mutated.count(before))
  mutated = mutated.replace(before, after, 1)
 replacement = folder / Path(path).name
 replacement.write_text(mutated)
 overlay = folder / 'overlay.json'
 overlay.write_text(json.dumps({'Replace': {str(root/path): str(replacement)}}))
 (folder/'mutation.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile=path,tofile=path)))
 command = ['go', 'test', '-overlay', str(overlay), '-count=1', '-timeout', '30m', package]
 if pattern: command += ['-run', pattern]
 with (folder/'package.log').open('wb') as log:
  test = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=1900)
 compiler = folder/'adamic'
 with (folder/'build.log').open('wb') as log:
  build = subprocess.run(['go','build','-overlay',str(overlay),'-o',str(compiler),'./cmd/adamic'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=180)
 record = {'file':path, 'package_command':command, 'package_exit':test.returncode, 'build_exit':build.returncode, 'probes':probes}
 if build.returncode == 0 and probes:
  with (folder/'probes.log').open('wb') as log:
   probe = subprocess.run(['python3',str(unit/'run.py'),'--compiler',str(compiler),'--results',str(folder/'results')]+probes,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=300)
  record['probe_runner_exit'] = probe.returncode
 (folder/'record.json').write_text(json.dumps(record,indent=2)+'\n')
 records[name]=record
 print(f"{name}: package_exit={test.returncode} build_exit={build.returncode}",flush=True)
(scratch/('summary-'+('-'.join(args.names) or 'all')+'.json')).write_text(json.dumps(records,indent=2)+'\n')

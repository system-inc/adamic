#!/usr/bin/env python3
"""Check focused survivors against the full native package and the new witness."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
unit = Path(__file__).resolve().parent
evidence = unit / 'evidence'
results = []
for mutant in json.loads((evidence / 'mutants.json').read_text()):
 if mutant['exit'] != 0:
  continue
 path = root / mutant['file']
 original = path.read_text()
 try:
  assert original.count(mutant['before']) == 1
  path.write_text(original.replace(mutant['before'], mutant['after'], 1))
  argv = ['go', 'test', './internal/native', '-count=1', '-timeout', '30m']
  with (evidence / ('full-' + mutant['name'] + '.log')).open('wb') as log:
   result = subprocess.run(argv, cwd=root, stdout=log, stderr=log, timeout=1800)
  results.append({'name': mutant['name'], 'command': argv, 'exit': result.returncode})
  print(mutant['name'], 'full-package-exit=' + str(result.returncode), flush=True)
  if mutant['name'] == 'record-two-index-sort':
   with tempfile.TemporaryDirectory(dir='/tmp/adamic-gate') as scratch:
    compiler = str(Path(scratch) / 'adamic')
    with (evidence / 'sort-witness-build.log').open('wb') as log:
     subprocess.run(['go', 'build', '-o', compiler, './cmd/adamic'], cwd=root, stdout=log, stderr=log, check=True)
    source = str(unit / 'entries-two-indices.a')
    witness = {}
    for mode in ['sanitize', 'O2', 'javascript']:
     output = str(Path(scratch) / ('program-' + mode + ('.mjs' if mode == 'javascript' else '')))
     build = [compiler, 'js', source] if mode == 'javascript' else [compiler, 'build', source, '-o', output] + (['--sanitize'] if mode == 'sanitize' else [])
     with (evidence / ('sort-witness-' + mode + '-compile.stdout')).open('wb') as out, (evidence / ('sort-witness-' + mode + '-compile.stderr')).open('wb') as err:
      subprocess.run(build, cwd=root, stdout=out, stderr=err, check=True)
     if mode == 'javascript':
      Path(output).write_bytes((evidence / ('sort-witness-' + mode + '-compile.stdout')).read_bytes())
      argv = ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', output]
     else:
      argv = [output]
     with (evidence / ('sort-witness-' + mode + '.stdout')).open('wb') as out, (evidence / ('sort-witness-' + mode + '.stderr')).open('wb') as err:
      ran = subprocess.run(argv, cwd=root, stdout=out, stderr=err, env=dict(os.environ, ASAN_OPTIONS='detect_leaks=1', UBSAN_OPTIONS='halt_on_error=1:print_stacktrace=1'), timeout=120)
     witness[mode] = {'exit': ran.returncode, 'stdout': (evidence / ('sort-witness-' + mode + '.stdout')).read_text(), 'stderr': (evidence / ('sort-witness-' + mode + '.stderr')).read_text()}
    (evidence / 'sort-witness.json').write_text(json.dumps(witness, indent=2) + '\n')
 finally:
  path.write_text(original)
(evidence / 'survivors.json').write_text(json.dumps(results, indent=2) + '\n')

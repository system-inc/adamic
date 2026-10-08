#!/usr/bin/env python3
"""Temporarily remove production guards, run checks, and always restore files."""
import json
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
evidence = Path(__file__).resolve().parent / 'evidence'
mutants = [
 ('numeric-index-guard', 'internal/lower/optional_indexing.go', 'if index.Type() != ir.Number {', 'if false {', ['./internal/lower'], 'TestOptionalIndexing'),
 ('structural-map-guard', 'internal/lower/optional_indexing_map.go', 'if l.optionalMapStructuralReceiver(access.Expression) {', 'if false {', ['./internal/lower'], 'TestOptionalIndexing'),
 ('substr-missing-argument-guard', 'internal/lower/string_substr.go', 'if index >= len(reads) {', 'if false {', ['./internal/oracle'], 'TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr'),
 ('record-two-index-sort', 'internal/native/runtime/record.c', 'if (count > 1) {', 'if (count > 2) {', ['./internal/native'], 'TestRecordsAgainstNode'),
 ('record-allocation-size-guard', 'internal/native/runtime/record.c', 'if (count > SIZE_MAX / sizeof *indices) {\n\t\t\tadamic_panic("out of memory", sizeof "out of memory" - 1);\n\t\t}', '', ['./internal/native'], 'TestRecordsAgainstNode'),
 ('record-deleted-key-filter', 'internal/native/runtime/record.c', 'if (!entry->deleted && !array_index(entry->key.reference, &number)) {', 'if (!array_index(entry->key.reference, &number)) {', ['./internal/native'], 'TestRecordsAgainstNode'),
]
results = []
for name, file, before, after, packages, pattern in mutants:
 path = root / file
 original = path.read_text()
 assert original.count(before) == 1, name
 try:
  path.write_text(original.replace(before, after, 1))
  command = ['go', 'test', *packages, '-run', pattern, '-count=1', '-timeout', '30m', '-v']
  with (evidence / ('mutant-' + name + '.log')).open('wb') as log:
   result = subprocess.run(command, cwd=root, stdout=log, stderr=log, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), timeout=1800)
  results.append({'name': name, 'file': file, 'before': before, 'after': after, 'command': command, 'exit': result.returncode})
  print(name, 'exit=' + str(result.returncode), flush=True)
 finally:
  path.write_text(original)
(evidence / 'mutants.json').write_text(json.dumps(results, indent=2) + '\n')

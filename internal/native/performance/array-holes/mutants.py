#!/usr/bin/env python3
"""Run the five requested semantic mutants and restore each source in finally."""
import argparse
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--logs', required=True)
parser.add_argument('--only', choices=['foreach', 'map', 'join', 'bound', 'count'])
args = parser.parse_args()
root = Path(__file__).resolve().parents[4]
logs = Path(args.logs).resolve()
logs.mkdir(parents=True, exist_ok=True)
mutants = [
 ('foreach', 'internal/native/emit_arrays.go',
  'e.line("if (%s == NULL) continue;", slot)',
  '''if visit.Method == "forEach" {
   missing := "{0}"
   if visit.Element == ir.MaybeNumber { missing = "{.number = adamic_maybe_number_pack((adamic_maybe_number){0})}" }
   e.line("adamic_value %s_missing = %s;", slot, missing)
   e.line("if (%s == NULL) %s = &%s_missing;", slot, slot, slot)
  } else { e.line("if (%s == NULL) continue;", slot) }''',
  'library_array_holes_callbacks'),
 ('map', 'internal/native/library_array_holes.go',
  'e.line("if (%s == NULL) continue;", slot)',
  'e.line("if (%s == NULL) { adamic_array_holes_set(%s, (double)%s, (adamic_value){0}); continue; }", slot, result, index)',
  'library_array_holes_callbacks'),
 ('join', 'internal/native/runtime/array_holes.c',
  'hole_text = ADAMIC_STRING("")', 'hole_text = ADAMIC_STRING("undefined")',
  'library_array_holes_callbacks'),
 ('bound', 'internal/native/runtime/array_holes.c',
  'length > 4294967295.0', 'length >= 4294967295.0',
  'array-holes-boundaries/library_array_holes_range'),
 ('count', 'internal/native/runtime/array_holes.c',
  'array->capacity--;', '/* mutant: do not decrement the hole count */',
  'library_array_holes_callbacks'),
]
for name, filename, old, new, fixture in mutants:
 if args.only and name != args.only:
  continue
 path = root / filename
 original = path.read_bytes()
 try:
  source = original.decode()
  if old not in source:
   raise RuntimeError(f'{name}: mutation anchor missing')
  path.write_text(source.replace(old, new, 1))
  log = logs / f'mutant-{name}.log'
  with log.open('wb') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run',
    'TestArrayHolesMilestone/' + fixture, '-count=1', '-v', '-timeout', '10m'],
    cwd=root, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  if result.returncode == 0 or 'stdout differs' not in observed:
   raise RuntimeError(f'{name}: no Node stdout disagreement; see {log}')
  if any(reason in observed for reason in ['error: ', 'AddressSanitizer:', 'runtime error:', 'signal: segmentation']):
   raise RuntimeError(f'{name}: compiler or sanitizer failure is not a semantic kill; see {log}')
  print(f'{name}: killed by Node stdout comparison', flush=True)
 finally:
  path.write_bytes(original)

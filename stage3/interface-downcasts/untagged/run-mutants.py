#!/usr/bin/env python3
"""Run each semantic defect separately; always restore the lane-owned files."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = root / 'stage3/interface-downcasts/untagged/logs'
logs.mkdir(exist_ok=True)
c = root / 'internal/native/runtime/view_unions_untagged.c'
js = root / 'internal/javascript/view_unions_untagged.go'
oracle = root / 'internal/oracle/checked_views_untagged_unions_test.go'
changes = [
 ('skip-check-native', c,
  'if (match(context, &descriptor, value)) { return member->contract; }',
  'if (true) { return member->contract; }'),
 ('skip-check-javascript', js,
  "if (match({kind:'object', contract:member.contract}, snapshot)) return member.contract;",
  'if (true) return member.contract;'),
 ('accept-wrong-shape-native', c,
  'if (!candidate) { continue; }', '/* mutant: ignore member tags */'),
 ('accept-wrong-shape-javascript', js,
  'if (!candidate) continue;', '/* mutant: ignore member tags */'),
 ('drop-transitive-native', oracle,
  'return member->contract!=1 && !object->missing && object->initialized && object->label.kind==adamic_view_union_string;',
  'return member->contract!=1 && !object->missing && object->initialized;'),
 ('drop-transitive-javascript', oracle,
  "return member.contract!==1 && label!==undefined && label.initialized && label.snapshot.kind==='string';",
  'return member.contract!==1 && label!==undefined && label.initialized;'),
]
original = {path: path.read_bytes() for _, path, _, _ in changes}
try:
 for name, path, before, after in changes:
  source = original[path].decode()
  assert source.count(before) == 1, (name, 'mutation anchor changed')
  path.write_text(source.replace(before, after))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestCheckedViewUntaggedSelection$', '-count=1', '-v', '-timeout', '10m'], cwd=root, env={**os.environ, 'ADAMIC_GATE_UNCACHED':'1'}, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL: TestCheckedViewUntaggedSelection' in observed, (name, observed)
  assert 'native: clang failed' not in observed and '[build failed]' not in observed, (name, observed)
  assert 'exitCode:0' in observed or 'stdout differs' in observed, (name, observed)
  print(name + ': caught by semantic output/exit pin', flush=True)
  path.write_bytes(original[path])
finally:
 for path, content in original.items():
  path.write_bytes(content)

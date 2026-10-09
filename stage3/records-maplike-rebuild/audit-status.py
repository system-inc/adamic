#!/usr/bin/env python3
"""Assert that status changes touch only stage0, and list each changed result."""
from pathlib import Path
import json
import re
import subprocess

root = Path(__file__).resolve().parents[2]
decoder = json.JSONDecoder()


def outside_stage0(text):
    result = []
    cursor = 0
    for match in re.finditer(r'"stage0"\s*:\s*', text):
        _, end = decoder.raw_decode(text[match.end():])
        result.append(text[cursor:match.end()])
        result.append('STAGE0')
        cursor = match.end() + end
    result.append(text[cursor:])
    return ''.join(result).encode('utf-8')


changes = []
for path in sorted((root / 'stage3/fixtures').glob('*/status.json')):
    relative = str(path.relative_to(root))
    original = subprocess.check_output(['git', 'show', 'origin/main:' + relative], cwd=root).decode('utf-8')
    current = path.read_bytes().decode('utf-8')
    assert outside_stage0(original) == outside_stage0(current), 'Node or other evidence changed: ' + relative
    before, after = json.loads(original), json.loads(current)
    assert len(before) == len(after)
    for old, new in zip(before, after):
        assert old['file'] == new['file']
        if old['stage0'] != new['stage0']:
            changes.append({'file': str(path.parent.relative_to(root) / old['file']), 'before': old['stage0'], 'after': new['stage0']})
regressions = [change for change in changes if change['before']['outcome'] == 'Compiles' and change['after']['outcome'] in ('Refused', 'NotYet')]
assert not regressions, 'Previously compiling fixtures became gaps: ' + repr(regressions)
print(json.dumps({'node_and_all_fields_outside_stage0_byte_identical': True, 'deliberate_regressions': regressions, 'changes': changes}, indent=2))

#!/usr/bin/env python3
"""Keep ordinary evidence byte-identical; observe named host profiles explicitly."""
from pathlib import Path
import json
import re
import subprocess
import sys

root = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(root / 'internal/hostconfiguration'))
sys.dont_write_bytecode = True
from configuration import node_for

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


def raw_rows(text):
    result = {}
    cursor = text.index('[') + 1
    while True:
        while text[cursor].isspace() or text[cursor] == ',':
            cursor += 1
        if text[cursor] == ']':
            return result
        row, end = decoder.raw_decode(text, cursor)
        assert row['file'] not in result, 'duplicate fixture ' + row['file']
        result[row['file']] = text[cursor:end]
        cursor = end


changes = []
skipped = []
checked = []
all_identical = True
for path in sorted((root / 'stage3/fixtures').glob('*/status.json')):
    relative = str(path.relative_to(root))
    original = subprocess.check_output(['git', 'show', 'origin/main:' + relative], cwd=root).decode('utf-8')
    current = path.read_bytes().decode('utf-8')
    before, after = json.loads(original), json.loads(current)
    identical = outside_stage0(original) == outside_stage0(current)
    all_identical = all_identical and identical
    if not any(row.get('configuration') is not None for row in before + after):
        # The original audit, unchanged, for ordinary records.
        assert identical, 'Node or other evidence changed: ' + relative
        assert len(before) == len(after)
        pairs = list(zip(before, after))
        assert all(old['file'] == new['file'] for old, new in pairs)
    else:
        old_rows, new_rows = raw_rows(original), raw_rows(current)
        old_by_file = {row['file']: row for row in before}
        assert set(old_rows) <= set(new_rows), 'fixture deleted: ' + relative
        ordinary_before = [row['file'] for row in before if not row.get('configuration')]
        ordinary_after = [row['file'] for row in after if not row.get('configuration') and row['file'] in old_by_file and not old_by_file[row['file']].get('configuration')]
        # A declared profile transition may remove an ordinary row from this
        # list; it may not reorder the ordinary records that remain.
        assert [name for name in ordinary_before if name in ordinary_after] == ordinary_after
        pairs = []
        for index, row in enumerate(after):
            name = row['file']
            old = old_by_file.get(name)
            configuration = row.get('configuration')
            identity = str(path.parent.relative_to(root) / name)
            if configuration is None:
                assert old is not None and not old.get('configuration'), 'configuration removed: ' + identity
                assert outside_stage0(old_rows[name]) == outside_stage0(new_rows[name]), 'Node or other evidence changed: ' + identity
            else:
                binary = node_for(row, index + 1)
                if binary is None:
                    skipped.append(identity)
                else:
                    observation = subprocess.run([binary, '--disable-warning=ExperimentalWarning', str(root/'oracle/node.mjs'), str(path.parent/name)], cwd=root, capture_output=True)
                    actual = dict(stdout=observation.stdout.decode(), stderr=observation.stderr.decode(), exit=observation.returncode)
                    assert actual == row['node'], 'configured Node differs: ' + identity
                    checked.append(identity)
                if old is not None:
                    if old.get('configuration') == configuration:
                        assert outside_stage0(old_rows[name]) == outside_stage0(new_rows[name]), 'same-profile evidence changed: ' + identity
                    else:
                        # Only a named profile and its freshly observed golden
                        # may differ; provenance stays the same.
                        keep = lambda value: {key: item for key, item in value.items() if key not in ('configuration', 'node', 'stage0')}
                        assert keep(old) == keep(row), 'fixture provenance changed: ' + identity
            if old is not None:
                pairs.append((old, row))
    for old, new in pairs:
        if old['stage0'] != new['stage0']:
            changes.append({'file': str(path.parent.relative_to(root) / old['file']), 'before': old['stage0'], 'after': new['stage0']})
regressions = [change for change in changes if change['before']['outcome'] == 'Compiles' and change['after']['outcome'] in ('Refused', 'NotYet')]
assert not regressions, 'Previously compiling fixtures became gaps: ' + repr(regressions)
report = {'node_and_all_fields_outside_stage0_byte_identical': all_identical, 'deliberate_regressions': regressions, 'changes': changes}
if skipped or checked:
    report.update(skipped_configurations=skipped, checked_configurations=checked)
print(json.dumps(report, indent=2))

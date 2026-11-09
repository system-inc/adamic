"""Validate paired selection observations; never turn a blocked group into a zero."""
import hashlib
import json
from pathlib import Path
root = Path(__file__).resolve().parent
data = json.loads((root / 'dump.json').read_text())
assert data['header']['entry'].endswith('/src/tsc/tsc.ts')
assert data['header']['config'].endswith('/src/compiler/tsconfig.json')
pairs = {}
for group in data['groups']:
    pair = pairs.setdefault(group['entry'], {})
    mode = group['control_through_record_false']
    assert mode not in pair
    pair[mode] = group

def canonical(components):
    # Checker IDs are process-local allocation metadata. Compare all source facts.
    components = json.loads(json.dumps(components))
    for component in components:
        for member in component['members']:
            member.pop('type_id', None)
        component['members'].sort(key=lambda m: json.dumps(m, sort_keys=True))
    return sorted(components, key=lambda c: json.dumps(c, sort_keys=True))

assert len(pairs) == data['header']['compiler_files']

for name, pair in pairs.items():
    assert set(pair) == {False, True}, name
    a, b = pair[False], pair[True]
    for flag in ['cycles_reached', 'selection_reached', 'selection_completed']:
        assert a[flag] == b[flag], (name, flag)
    assert canonical(a['selected_components']) == canonical(b['selected_components']), name
    if not a['selection_reached']:
        assert a.get('failure') == b.get('failure'), name
        assert a.get('panic') == b.get('panic'), name
        assert a['last_phase'] == b['last_phase'], name
        assert not a['selected_components'], name
provenance=json.loads((root/'provenance.json').read_text())
for key,file in [('test_sha256','record_select_census_test.go.txt'),('scratch_patch_sha256','scratch.patch'),('source_manifest_sha256','source-sha256.json')]:
    assert hashlib.sha256((root/file).read_bytes()).hexdigest()==provenance[key], file
status=json.loads((root/'run-exit.json').read_text())
assert status['child_exit_code']==0 and not status['silent_30_minutes']
assert provenance['measured_run']==status
assert all(code==0 for code in provenance['successful_child_exit_codes'].values())
fixture = json.loads((root / 'fixture.json').read_text())
assert len(fixture) == 2
assert all(g['selection_reached'] and g['selection_completed'] for g in fixture)
assert canonical(fixture[0]['selected_components']) == canonical(fixture[1]['selected_components'])
assert any(c['has_record'] for c in fixture[0]['selected_components'])
assert 'a record can\'t join a graph region yet' in fixture[0]['failure']
assert 'a record can\'t join a graph region yet' not in fixture[1]['failure']
print(json.dumps({'groups': len(pairs), 'control_components_identical': True,
                  'selection_reached_groups': sum(p[False]['selection_reached'] for p in pairs.values()),
                  'selection_completed_groups': sum(p[False]['selection_completed'] for p in pairs.values()),
                  'fixture_selected_record_and_control_verified': True}, indent=2))

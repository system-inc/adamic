"""Check exact event identity against the independent ledger, then report reviewed extras."""
import collections
import json
import sys
from pathlib import Path

independent = json.loads(Path(sys.argv[1]).read_text())
ours = json.loads(Path(sys.argv[2]).read_text())


def key(event):
    return event['file'], event['start'], event['end'], event['property'], event['form']


expected = {key(event): event for event in independent['writes']}
actual = {key(event): event for event in ours['writes']}
assert len(expected) == len(independent['writes']), 'duplicate independent event'
assert len(actual) == len(ours['writes']), 'duplicate adjudicated event'
assert expected.keys() == actual.keys(), 'supplied event sets differ'
for event_key, original in expected.items():
    current = actual[event_key]
    assert all(current[field] == value for field, value in original.items()), event_key
    if current.get('actual_value_type'):
        assert current['actual_value_type'] == original['value_type'], event_key

extras = ours['additional']
assert len({(w['file'], w['start'], w['end'], w['property']) for w in extras}) == len(extras)
assert all(w['construction'] in ('inside', 'outside local construction') for w in extras)
assert all(w['explanation'] and w['source_sha256'] and w['scope_difference'] for w in extras)
assert all(not match['receiver_assignable_to_member'] for w in extras for match in w['receiver_to_member'])
assert ours['diagnostics'] == 0 and ours['sourceFiles'] == 77
summary = {
    'supplied_events': len(expected),
    'missing': 0,
    'duplicated': 0,
    'original_fields_changed': 0,
    'rechecked_assignment_value_type_differences': 0,
    'rechecked_assignment_update_events': sum(bool(w.get('actual_value_type')) for w in actual.values()),
    'supplied_construction': dict(collections.Counter(w['construction'] for w in actual.values())),
    'supplied_kind': dict(collections.Counter(w['construction'] for w in actual.values() if w['property'] == 'kind')),
    'supplied_rule': ours['count'],
    'extra_events': len(extras),
    'extra_construction': dict(collections.Counter(w['construction'] for w in extras)),
    'extra_rule': ours['extra_count'],
    'root_declaration_witness_missing': sum(not w['declaration_witnesses'] for w in extras),
    'receiver_to_member_filter_differs': sum(bool(w['declaration_witnesses']) for w in extras),
    'combined': ours['combined'],
}
print(json.dumps(summary, indent=2))

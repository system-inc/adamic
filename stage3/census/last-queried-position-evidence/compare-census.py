from collections import Counter
import json
from pathlib import Path
import re
import sys

proof = Path(sys.argv[1])
mode = sys.argv[2]
logs = proof / 'evidence'
relative = 'src/compiler/parser.ts'
original = b'let lastQueriedPosition = InvalidPosition.Value;'
edited = b'let lastQueriedPosition: number = InvalidPosition.Value;'
before = (proof / 'before' / relative).read_bytes()
after = (proof / 'after' / relative).read_bytes()
assert before.count(original) == 1 and after == before.replace(original, edited)

lines = before.decode().splitlines()
declaration, = (number for number, text in enumerate(lines, 1) if original.decode() in text)
write, = (number for number, text in enumerate(lines, 1) if 'lastQueriedPosition = position;' in text)
column = lines[declaration - 1].index('lastQueriedPosition') + len('lastQueriedPosition') + 1
delta = len(edited) - len(original)

def census(label):
    root = str(proof / label) + '/'
    stream = [json.loads(line) for line in (logs / f'{label}-census.jsonl').read_text().splitlines()]
    header, record, trailer = stream
    assert header['status'] == 'measurement' and header['latent_mode'] == 'full'
    assert header['module_format'] == 'CommonJSBundle'
    assert record['file'] == root + relative and record['units']
    assert trailer['status'] == 'relation_report'
    coordinate = re.compile(re.escape(relative) + rf':{declaration}:(\d+)(?!\d)')

    def normalized(text):
        text = text.replace(root, '')
        if label == 'after':
            def mapped(match):
                queried = int(match[1])
                if queried >= column:
                    queried = max(column, queried - delta)
                return f'{relative}:{declaration}:{queried}'
            text = coordinate.sub(mapped, text)
        return text

    def key(finding):
        return tuple(normalized(finding.get(field, '')) for field in ('unit', 'phase', 'kind', 'where', 'reason', 'text'))

    findings = Counter(map(key, (record['findings'] or []) + (trailer.get('findings') or [])))
    units = {normalized(unit['where']): unit['status'] for unit in record['units']}
    return findings, units

def target(event):
    unit, phase, kind, where, reason, text = event
    return kind in {'Refused', 'NotYet'} and where.startswith(f'{relative}:{write}:')

before_events, before_units = census('before')
after_events, after_units = census('after')
if mode == 'mutant':
    label = sys.argv[3]
    assert (proof / label / relative).read_bytes() == before
    mutant_events, mutant_units = census(label)
    assert any(map(target, mutant_events)), 'mutant did not restore the write stop'
    assert mutant_events == before_events and mutant_units == before_units, 'mutant census drift'
    print('Restored write stop', f'{relative}:{write}')
    print('All baseline events and unit statuses match the mutant.')
    raise SystemExit(0)
assert mode == 'before-after'
stops = {'Refused', 'NotYet'}
gone, added = before_events - after_events, after_events - before_events
prior_blockers = {event[0] for event in before_events if event[2] in stops}
target_units = {event[0] for event in gone if target(event)}
reached, novel = [], []
for event in added.elements():
    unit, phase, kind, where, reason, text = event
    if unit in prior_blockers and unit in target_units and kind in stops | {'Boundary'} and where:
        reached.append(event)
    else:
        novel.append(event)

result = {
    'target': f'{relative}:{write}',
    'counts': {'before': before_events.total(), 'after': after_events.total()},
    'removed': list(gone.elements()),
    'reached_in_previously_blocked_units': reached,
    'new': novel,
}
(logs / 'census-delta.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result, indent=2))
assert any(map(target, before_events)), 'baseline write stop missing'
assert not any(map(target, after_events)), 'candidate still stops at the write'
assert not novel, 'unexplained new findings'
assert before_units.keys() == after_units.keys(), 'attempted unit set changed'

import json
import subprocess
from pathlib import Path

root = Path(__file__).resolve().parent

def table(text):
    rows = {}
    for line in text.splitlines():
        cells = [cell.strip() for cell in line.split('|')[1:-1]]
        if len(cells) == 7 and all(cell.isdigit() for cell in cells[1:]):
            assert cells[0] not in rows
            rows[cells[0]] = list(map(int, cells[1:]))
    return rows

def revision(rev):
    return table(subprocess.check_output(['git', 'show', rev + ':internal/oracle/counts.md'], text=True))

main = revision('0bf6186d')
previous = revision('1fc069f0')
base = revision('5e33a17b')
current = table(Path('internal/oracle/counts.md').read_text())
own = {row['fixture']: row for row in json.loads(Path('review/compiler/chain-slice-4/counts-attribution.json').read_text())['rows'] if row['changed']}
getters = {name for name in main if name not in base}
assert getters == {'internal/oracle/testdata/getters_census_' + name + '.a' for name in ['class_map', 'throwing_object', 'binary_inline']}
assert set(current) == set(main) | set(previous)
result = {'main': '0bf6186d', 'previous_slice': '1fc069f0', 'rows_total': len(current), 'changes_against_main': [], 'changes_against_previous_slice': []}
for name in sorted(current):
    value = current[name]
    if main.get(name) != value:
        assert name in own, ('unattributed row', name)
        assert own[name]['after'] == value, ('unexpected slice count', name, value)
        result['changes_against_main'].append({'fixture': name, 'before': main.get(name), 'after': value, 'owner': own[name]['owner'], 'reason': own[name]['reason']})
    if previous.get(name) != value:
        assert name in getters, ('unexpected change since slice', name, previous.get(name), value)
        assert value == main[name]
        result['changes_against_previous_slice'].append({'fixture': name, 'before': previous.get(name), 'after': value, 'owner': 'main getters census'})
assert len(result['changes_against_main']) == 59
assert len(result['changes_against_previous_slice']) == 3
(root / 'counts-attribution.json').write_text(json.dumps(result, indent=2) + '\n')
print('PASS: 1069 rows; 59 slice-owned deltas against main (55 additions, 4 existing changes); only 3 main getters additions against previous slice')
for row in result['changes_against_previous_slice']:
    print(row['fixture'], '/'.join(map(str, row['after'])))
for row in result['changes_against_main']:
    if row['before'] is not None:
        print(row['fixture'], '/'.join(map(str, row['before'])), '->', '/'.join(map(str, row['after'])), row['owner'])

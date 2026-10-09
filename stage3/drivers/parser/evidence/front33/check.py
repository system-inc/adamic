"""Verify recorded scope and comparisons; reject counting discovery artifacts."""
from pathlib import Path
import copy
import hashlib
import json

root = Path(__file__).resolve().parent

def check(rows):
    assert 1 <= len(rows) <= 15
    for index, row in enumerate(rows, 1):
        assert row['order'] == index
        assert row['found_behind_stubs'] == (index > 1)
        assert row['file'].startswith('src/compiler/')
        assert row['line'] > 0 and row['column'] > 0
        assert (root.parent.parent / row['minimal_program']).is_file()
        assert not any(text in row['message'] for text in
                       ['TS2393', 'TS2454', "Property 'length' does not exist on type 'void'"])

rows = json.loads((root / 'parser-front33-ordered-stops.json').read_text())['rows']
check(rows)
mutant = copy.deepcopy(rows)
mutant[-1]['message'] = 'error TS2454: Variable enumMemberCache is used before being assigned.'
try:
    check(mutant)
except AssertionError:
    pass
else:
    raise AssertionError('discovery artifact entered the source blocker list')
probes = json.loads((root / 'parser-front33-minimals/report.json').read_text())
assert probes['total'] == 20 and probes['passing'] == 6
for probe in probes['probes']:
    assert probe['node']['exit'] == 0
    if probe['pass']:
        assert probe['build']['exit'] == probe['native']['exit'] == 0
        assert probe['native_byte_mutant_cmp'] == 1
        assert all(value == 0 for value in probe['comparison'].values())
        assert probe['node']['stdout'] == probe['native']['stdout']
        assert probe['node']['stderr'] == probe['native']['stderr'] == ''
    else:
        assert probe['build']['exit'] != 0
for name in ['parser-front33-full-node', 'parser-front33-slice-node']:
    report = json.loads((root / name / 'report.json').read_text())
    assert report['node']['bytes'] == 36429231
    assert report['node']['sha256'] == '686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615'
    assert report['node']['stderr_bytes'] == 0
    assert report['mutation']['changed_lines'] == 1
    jsdoc = json.loads((root / name / 'jsdoc-mutants/report.json').read_text())
    assert all(row['comparison_exit'] == 1 for row in jsdoc.values())
metrics = json.loads((root / 'parser-front33-build-modes/report.json').read_text())
assert metrics['generated_c'] is None and metrics['emit_exit'] != 0
assert all(row['exit'] != 0 and row['clang_wall_seconds'] is None for row in metrics['attempts'])
print(json.dumps({'source_stops': len(rows), 'artifact_count_mutant_caught': True,
                  'native_probes_matching_node': 6, 'native_byte_mutants_caught': 6,
                  'extended_node_identity': True, 'parser_native_acceptance': 'unreached'}))

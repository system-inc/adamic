from pathlib import Path
import copy
import gzip
import hashlib
import json

root = Path(__file__).resolve().parent

def load(name):
    return json.loads((root / name).read_text())

def check_stops(report):
    rows = report['rows']
    assert len(rows) == 16
    assert [r['order'] for r in rows] == list(range(1, 17))
    assert 'no parser C' in report['native_acceptance']
    assert rows[0]['file'] == 'src/compiler/debug.ts'
    assert (rows[0]['line'], rows[0]['column']) == (113, 19)
    assert not rows[0]['pending']
    for r in rows:
        assert r['owner'] and r['awaits'] and r['message']
        assert r['status_against_blockers'] in ['still there', 'new', 'gone']
        assert r['pending'] == (r['order'] > 1)
        target = root / r['minimal'] if r['minimal'].startswith('minimals/') else root.parents[1] / r['minimal']
        assert target.is_file(), target
    artifacts = report['excluded_artifacts']
    assert len(artifacts) == 3
    assert all(r['discovery_artifact'] for r in artifacts)
    assert all(r['message'].startswith(('error TS2322:', 'error TS2345:')) for r in artifacts)

def check_modes(report):
    assert [r['split'] for r in report] == [0, 1]
    for r in report:
        assert r['exit'] != 0
        assert 'debug.ts:113:19:' in r['stderr']
        assert 'debug.ts:114:19:' in r['stderr']
        assert 'captureStackTrace' in r['stderr']

def check_probes(report, expected_pass, total):
    assert report['total'] == total
    assert report['passing'] == expected_pass
    assert sum(r['pass'] for r in report['probes']) == expected_pass
    for r in report['probes']:
        assert r['node']['exit'] == 0 and r['node']['stderr'] == ''
        if r['build']['exit'] == 0:
            assert r['pass'], 'silent miscompile: ' + r['file']
            assert r['native']['exit'] == 0
            assert r['node']['stdout'] == r['native']['stdout']
            assert r['node']['stderr'] == r['native']['stderr']
            assert all(c == 0 for c in r['comparison'].values())
            assert r['native_byte_mutant_cmp'] == 1
        else:
            assert not r['pass']

stops = load('stops.json')
check_stops(stops)
modes = load('build-modes/report.json')
check_modes(modes)
check_probes(load('minimals/report.json'), 12, 21)
check_probes(load('extra-minimals/report.json'), 2, 6)
check_probes(load('container-final-minimal/report.json'), 0, 1)
merges = load('merges/report.json')
assert merges['base'] == '54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8'
assert len(merges['rows']) == 8
assert [r['order'] for r in merges['rows'] if r['status'] == 'merged'] == [1, 4]
for r in merges['rows']:
    if r['status'] == 'skipped conflict':
        assert r['merge_exit'] != 0 and r['abort_exit'] == 0 and r['conflicts']
        assert r['before'] == r['after']
assert (root / 'parser-rehearsal-compiler-sha').read_text().strip() == merges['rows'][-1]['after']
assert hashlib.sha256((root.parents[1] / 'BLOCKERS.md').read_bytes()).hexdigest() == (root / 'baseline-blockers.sha256').read_text().strip()
summary = load('slice-summary.json')
assert (summary['declaration_files'], summary['code_declarations'], summary['declaration_lines']) == (27, 1994, 41861)
assert summary['copied_spans'] == 2083 and summary['preserves_evaluation']
for name in ['full-node', 'slice-node']:
    report = load(name + '/report.json')
    assert report['files'] == 81
    assert report['node']['bytes'] == 36429231
    assert report['node']['sha256'] == '686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615'
    assert report['node']['stderr_bytes'] == 0
    assert report['mutation']['changed_lines'] == 1
    mutants = load(name + '/jsdoc-mutants/report.json')
    assert all(row['comparison_exit'] == 1 for row in mutants.values())
with gzip.open(root / 'parser-rehearsal-node-cmp.log.gz', 'rb') as f:
    assert f.read() == b''
with gzip.open(root / 'parser-rehearsal-lane-tests.log.gz', 'rt') as f:
    log = f.read()
    assert 'Ran 32 tests' in log and '\nOK\n' in log

# Mutate each acceptance boundary separately; none may become a green claim.
mutants = []
def rejected(name, fn, value):
    try:
        fn(value)
    except AssertionError:
        mutants.append(name)
    else:
        raise AssertionError('uncaught evidence mutant: ' + name)
changed = copy.deepcopy(stops)
changed['rows'].pop()
rejected('drop one ordered source stop', check_stops, changed)
changed = copy.deepcopy(stops)
changed['rows'][1]['awaits'] = ''
rejected('erase placeholder pending dependency', check_stops, changed)
changed = copy.deepcopy(stops)
changed['rows'][1]['pending'] = False
rejected('claim a stubbed row is unconditional', check_stops, changed)
changed = copy.deepcopy(modes)
changed[0]['exit'] = 0
rejected('claim native parser green', check_modes, changed)
changed = copy.deepcopy(load('minimals/report.json'))
row = next(r for r in changed['probes'] if r['pass'])
row['native']['stdout'] = 'x' + row['native']['stdout']
rejected('one-byte matching-probe metadata lie', lambda r: check_probes(r, 12, 21), changed)
print(json.dumps({'real_stops': 16, 'artifacts_excluded': 3, 'primary_matching': 12, 'primary_total': 21, 'supplemental_matching': 2, 'supplemental_total': 7, 'native_byte_mutants_caught': 14, 'node_tree_identity': True, 'parser_native': 'checker red; pending rehearsal', 'evidence_mutants_caught': mutants}, indent=2))

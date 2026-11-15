"""Verify the archived phase-aware single-file census comparison; no compiler build."""
import collections
import gzip
import json
import re
from pathlib import Path

root = Path(__file__).resolve().parent
artifacts = root / 'evidence'


def load(name):
    return json.loads((artifacts / name).read_text())


def read(name):
    rows = [json.loads(line) for line in gzip.decompress(
        (artifacts / (name + '.jsonl.gz')).read_bytes()).decode().splitlines()]
    assert len(rows) == 3
    assert rows[0]['status'] == 'measurement'
    assert rows[0]['latent_mode'] == 'full'
    assert rows[0]['module_format'] == 'CommonJSBundle'
    assert rows[0]['checker_rejected'] is False
    assert rows[0]['diagnostics'] == []
    assert rows[-1]['status'] == 'relation_report'
    assert rows[1]['file'].endswith('/adapted/src/compiler/sys.ts')
    assert len(rows[1]['units']) == 136
    return rows


def key(finding):
    message = re.sub(r'^.*?:\d+:\d+: ', '', finding['text'], count=1)
    return (finding['phase'], finding['kind'], finding['where'],
            finding['reason'], message)


def findings(rows):
    return {key(finding) for row in rows for finding in row.get('findings', [])}


base, head = read('base'), read('head')
assert base[0] == head[0]
assert {k: v for k, v in base[1].items() if k != 'findings'} == {
    k: v for k, v in head[1].items() if k != 'findings'}
before, after = findings(base), findings(head)
assert (len(before), len(after)) == (109, 107)
assert (len(base[1]['findings']), len(head[1]['findings'])) == (180, 178)
gone, new = before - after, after - before
assert len(gone) == 2 and not new
reason = ('node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, '
          'not (timeout: unknown) => void with no slot')
expected_sites = {'1501:57', '1614:57'}
assert {f[2].split('sys.ts:')[1] for f in gone} == expected_sites
assert all(f[0:2] == ('refusal_scan', 'NotYet') and f[3] == reason
           and f[4] == "stage 0 can't lower " + reason + ' yet' for f in gone)
assert {(f['phase'], f['kind'], f['where'], f['reason'], f['message'])
        for f in load('gone.json')} == gone
assert load('new.json') == []
assert collections.Counter(k[:2] for k in before) == {
    ('lowering', 'Boundary'): 55, ('lowering', 'NotYet'): 38,
    ('lowering', 'Refused'): 10, ('refusal_scan', 'NotYet'): 3,
    ('refusal_scan', 'Refused'): 3}
assert collections.Counter(k[:2] for k in after) == {
    ('lowering', 'Boundary'): 55, ('lowering', 'NotYet'): 38,
    ('lowering', 'Refused'): 10, ('refusal_scan', 'NotYet'): 1,
    ('refusal_scan', 'Refused'): 3}
for rows, expected in [(base, '1501:57'), (head, '1548:13')]:
    scan = sorted((key(f) for f in rows[1]['findings']
                   if f['phase'] == 'refusal_scan'),
                  key=lambda k: tuple(map(int, k[2].rsplit(':', 2)[1:])))
    assert scan[0][2].endswith('sys.ts:' + expected)
    assert not any(k[2].endswith('sys.ts:1387:5') for k in scan)
first_head = min((k for k in after if k[0] == 'refusal_scan'),
                 key=lambda k: tuple(map(int, k[2].rsplit(':', 2)[1:])))
assert first_head in before and first_head[1] == 'Refused'
boundary = {k for k in before if k[0:2] == ('lowering', 'Boundary')
            and k[2].endswith('sys.ts:1387:5')}
assert len(boundary) == 1 and boundary <= after
ast = load('gone-ast.json')
assert len(ast) == 2
assert {a['finding']['where'] for a in ast} == {k[2] for k in gone}
assert all(a['inside_type_query'] and a['node_kind'] == 'Identifier'
           and a['node_text'] == 'clearTimeout'
           and a['type_query'] == 'typeof clearTimeout'
           and a['ancestors'][:4] == [
               'Identifier', 'TypeQuery', 'TypeReference', 'IndexedAccessType']
           for a in ast)
pins = load('provenance.json')
assert pins['base'] == pins['head_parent'] == '89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825'
assert pins['head'] == 'fc677461c7d593cb4eb9bfca86102f001a91d6e1'
assert pins['base_stage3_tree'] == pins['head_stage3_tree'] == 'c4981d46fb8998476032918d2e0003964bc60e3d'
assert pins['typescript'] == '050880ce59e30b356b686bd3144efe24f875ebc8'
for name in ['base', 'head']:
    run = load('run-' + name + '.json')
    assert run['exit'] == 0 and 0 < run['wall_seconds'] < run['bound_seconds'] == 3600
    assert run['env'] == {
        'LATENT_FULL': '1', 'LATENT_ASSERT_NO_OUTPUT': '1',
        'LATENT_ONLY_FILE': base[1]['file'],
        'CENSUS_MODULE_FORMAT': 'CommonJSBundle'}
summary = load('summary.json')
assert summary['new'] == summary['changed'] == []
assert summary['checker_headers_identical'] and summary['nonfinding_file_fields_identical']
manifest = (artifacts / 'adapted-src.sha256').read_text().splitlines()
assert len(manifest) == 747
assert summary['sys_sha256'] + '  compiler/sys.ts' in manifest
print('PASS: 2 type-query findings gone; 0 new; 0 changed; exact pins, completed runs, '
      'identical metadata, and measured first-scan/coordinate qualifications verified.')

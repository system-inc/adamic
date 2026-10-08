#!/usr/bin/env python3
"""Compare full, guarded censuses of exactly the same adapted compiler."""
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
PROJECT = Path('/tmp/optional-indexing-adapted/src/compiler')
BEFORE = Path('/tmp/optional-indexing-before.jsonl')
AFTER = Path('/tmp/optional-indexing-delivery.jsonl')
OUT = ROOT / 'review/optional-indexing'
REASONS = [
    '?.[] on a value',
    'a call through ?. (an optional call)',
    'an optional chain longer than one step',
    'optional chaining to .size on a value',
    '?. to a number, which would be number | undefined',
]


def normalize(value):
    return value.replace(str(PROJECT) + '/', '')


def read(path):
    records = [json.loads(line) for line in path.read_text().splitlines()]
    header, files = records[0], records[1:]
    assert header['latent_mode'] == 'full'
    assert header['measurement'] == 'measured on a checker-rejected program'
    assert header['checker_rejected'] is True
    assert len(files) == 79, (str(path), len(files))
    roots = {}
    for record in files:
        for finding in record['findings']:
            if finding['kind'] not in ('NotYet', 'Refused'):
                continue
            key = tuple(normalize(finding[field]) for field in ('kind', 'where', 'reason', 'text'))
            roots.setdefault(key, []).append({field: normalize(finding[field]) for field in finding})
    return header, files, roots


before_header, before_files, before = read(BEFORE)
after_header, after_files, after = read(AFTER)
assert before_header == after_header, 'checker metadata changed'
assert [r['file'] for r in before_files] == [r['file'] for r in after_files], 'file coverage changed'
manifest = {str(path.relative_to(PROJECT)): hashlib.sha256(path.read_bytes()).hexdigest()
            for path in sorted(PROJECT.rglob('*.ts'))}
assert len(manifest) == 79, len(manifest)
rows = []
for reason in REASONS:
    old = {key for key in before if key[2] == reason}
    new = {key for key in after if key[2] == reason}
    disappeared = []
    for key in sorted(old - new):
        filename, line, _column = key[1].rsplit(':', 2)
        replacements = sorted(k for k in after if k[1].rsplit(':', 2)[:2] == [filename, line])
        disappeared.append({'root': dict(zip(('kind', 'where', 'reason', 'text'), key)),
                            'before_attempts': before[key],
                            'remaining_same_line_roots': [dict(zip(('kind', 'where', 'reason', 'text'), r)) for r in replacements]})
    rows.append({'reason': reason, 'before': len(old), 'after': len(new),
                 'disappeared_exact_roots': len(old - new),
                 'roots_without_a_same_line_replacement': sum(not r['remaining_same_line_roots'] for r in disappeared),
                 'disappeared': disappeared,
                 'introduced': [dict(zip(('kind', 'where', 'reason', 'text'), key)) for key in sorted(new - old)],
                 'remaining': [dict(zip(('kind', 'where', 'reason', 'text'), key)) for key in sorted(new)]})
compiler_paths = ['internal/lower/object.go', 'internal/lower/optional_intrinsic.go',
                  'internal/lower/typed_arrays.go', 'internal/lower/optional_indexing.go',
                  'internal/lower/optional_indexing_chain.go', 'internal/lower/optional_indexing_map.go']
result = {'measurement': before_header['measurement'], 'mode': 'full',
          'base': '4885cec50290686df487b62aac47c85d871ed40c',
          'upstream_typescript': '050880ce59e30b356b686bd3144efe24f875ebc8',
          'files': 79, 'checker_metadata_identical': True,
          'source_hashes': manifest,
          'compiler_hashes': {p: hashlib.sha256((ROOT / p).read_bytes()).hexdigest() for p in compiler_paths},
          'guard_environment': {'LATENT_FULL': '1', 'LATENT_ASSERT_NO_OUTPUT': '1'},
          'rows': rows}
# Check every preserved compiler byte against the measurement binary's build manifest.
assert result['compiler_hashes'] == json.loads((OUT / 'compiler-hashes.json').read_text())
if '--audit' in sys.argv:
    assert result == json.loads((OUT / 'census.json').read_text()), 'recorded census changed'
    for name, path in [('before', BEFORE), ('after', AFTER)]:
        assert gzip.decompress((OUT / (name + '.jsonl.gz')).read_bytes()) == path.read_bytes()
else:
    (OUT / 'census.json').write_text(json.dumps(result, indent=2) + '\n')
    for name, path in [('before', BEFORE), ('after', AFTER)]:
        (OUT / (name + '.jsonl.gz')).write_bytes(gzip.compress(path.read_bytes(), mtime=0))
for row in rows:
    print(f"{row['reason']}: {row['before']} -> {row['after']}; {row['disappeared_exact_roots']} disappeared; {row['roots_without_a_same_line_replacement']} without same-line replacement")
print('Same checker metadata, same 79 measured files, 79 source hashes; audit passed.')

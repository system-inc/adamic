#!/usr/bin/env python3
"""Reuse the existing input selection and materialization, without compiler feedback."""
import argparse
import hashlib
import json
from pathlib import Path
import sys

SCOUT = Path(__file__).resolve().parent
STAGE = SCOUT.parents[1]
sys.path.insert(0, str(STAGE / 'drivers/tsc'))
from corpus import materialize, PIN

parser = argparse.ArgumentParser()
parser.add_argument('output', type=Path)
parser.add_argument('--upstream', type=Path)
args = parser.parse_args()
if args.output.exists():
    parser.error('output must be new')
selection = json.loads((STAGE / 'drivers/tsc/selection.json').read_text())
projects = []
for row in selection['cases']:
    source = STAGE / 'drivers/tsc' / row['path']
    raw = source.read_bytes()
    if hashlib.sha256(raw).hexdigest() != row['source_sha256']:
        raise RuntimeError('acceptance source hash changed: ' + str(source))
    text, options = materialize(raw)
    if options != row['options']:
        raise RuntimeError('acceptance options changed')
    projects.append({'id': row['id'], 'options': {'types': [], 'skipDefaultLibCheck': True,
        'noErrorTruncation': True, 'ignoreDeprecations': '6.0', **options},
        'files': [{'path': source.name, 'text': text}]})
files = []
for source in sorted((STAGE / 'drivers/tsc/tiny').glob('*.a')):
    raw = source.read_bytes()
    if raw.startswith(b'// a-check:'):
        raw = raw.split(b'\n', 1)[1]
    text, _ = materialize(raw)
    files.append({'path': source.with_suffix('.ts').name, 'text': text})
projects.append({'id': 'tiny', 'options': {'strict': True, 'target': 'es2020'}, 'files': files})
if args.upstream:
    import subprocess
    if subprocess.check_output(['git', '-C', str(args.upstream), 'rev-parse', 'HEAD'], text=True).strip() != PIN:
        raise RuntimeError('upstream pin changed')
    selection = json.loads((STAGE / 'verdict/selection.json').read_text())
    for row in selection['cases']:
        raw = (args.upstream / row['source']).read_bytes()
        if hashlib.sha256(raw).hexdigest() != row['source_sha256']:
            raise RuntimeError('upstream source hash changed: ' + row['source'])
        text, options = materialize(raw)
        if options != row['options']:
            raise RuntimeError('upstream options changed')
        projects.append({'id': row['source'], 'options': {'types': [], 'skipDefaultLibCheck': True,
            'noErrorTruncation': True, 'ignoreDeprecations': '6.0', **options},
            'files': [{'path': row['name'], 'text': text}]})
args.output.write_text(json.dumps({'projects': projects}, ensure_ascii=True) + '\n')
print(f'{len(projects)} projects written to {args.output}')

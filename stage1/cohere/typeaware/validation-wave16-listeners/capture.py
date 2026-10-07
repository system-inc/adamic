#!/usr/bin/env python3
"""Retain fresh oracle streams, without binary/archive artifacts."""
from pathlib import Path
import gzip
import hashlib
import json
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[4]
SCRATCH = Path(sys.argv[1]).resolve()
OUT = Path(__file__).resolve().parent
metadata = {
    'source_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
    'main': subprocess.check_output(['git', 'rev-parse', 'origin/main'], cwd=ROOT, text=True).strip(),
    'cohere': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT / 'cohere', text=True).strip(),
    'streams': {}, 'sources': {},
}
for group in ['first', 'followup', 'third', 'fourth', 'fifth']:
    target = OUT / 'streams' / group
    target.mkdir(parents=True, exist_ok=True)
    for path in sorted((SCRATCH / group).glob('*')):
        if path.suffix not in {'.stdout', '.stderr'}:
            continue
        data = path.read_bytes()
        metadata['streams'][group + '/' + path.name] = {
            'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest(),
            'original_path': str(path),
        }
        if path.suffix == '.stdout':
            (target / (path.name + '.gz')).write_bytes(gzip.compress(data, mtime=0))
        else:
            (target / path.name).write_bytes(data)
for path in sorted((ROOT / 'stage1/cohere/typeaware').glob('*.a')):
    metadata['sources'][str(path.relative_to(ROOT))] = hashlib.sha256(path.read_bytes()).hexdigest()
(OUT / 'provenance.json').write_text(json.dumps(metadata, indent=2, sort_keys=True) + '\n')
print(f"Captured {len(metadata['streams'])} oracle streams; source {metadata['source_commit']}")

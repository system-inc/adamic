#!/usr/bin/env python3
"""Verify inventory UTF-16 spans against an independent pinned tsc checkout."""
import json
import subprocess
import sys
from pathlib import Path
root = Path(__file__).resolve().parent
source = Path(sys.argv[1])
pin = '050880ce59e30b356b686bd3144efe24f875ebc8'
assert subprocess.check_output(['git', '-C', str(source), 'rev-parse', 'HEAD'], text=True).strip() == pin
pairs = json.loads((root / 'callable-pairs-ranked.json').read_text())
verified = []
count = 0
for pair in pairs:
    for site in pair['sites']:
        text = (source / site['file']).read_bytes().decode('utf-8').encode('utf-16-le')
        actual = text[site['start'] * 2:site['end'] * 2].decode('utf-16-le')
        assert actual == site['text'], (site, actual)
        count += 1
    verified.append({key: pair[key] for key in ['rank', 'type_id', 'type', 'field', 'read_count']} | {'witness': pair['sites'][0]})
assert count == 1503 and len(verified) == 308
(root / 'original-witnesses.json').write_text(json.dumps({'upstream': pin, 'verified_reads': count, 'pairs': verified}, indent=2) + '\n')
print(f'{len(verified)} pairs, {count} original read spans verified')

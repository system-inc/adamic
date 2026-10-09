"""Measure every authored .a in place and update only its expected-error header."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--update', action='store_true')
args = parser.parse_args()
root = Path(__file__).resolve().parent
records = []
for file in sorted(root.rglob('*.a')):
    source = file.read_text()
    body = re.sub(r'^// a-check: [^\n]*\n', '', source)
    result = subprocess.run([str(args.compiler.resolve()), 'c', str(file)], capture_output=True, text=True, timeout=30)
    reason = re.search(r'Adamic 0\.1 refuses (.*?); ', result.stderr)
    code = re.search(r'\bTS(\d+)\b', result.stderr)
    header = '// a-check: refused ' + reason[1] if reason else '// a-check: type error TS' + code[1] if code else ''
    if result.returncode and not reason and not code:
        assert any(word in result.stderr.lower() for word in ['not yet', 'notyet', "can't lower"]), result.stderr
    old = source.splitlines()[0] if source.startswith('// a-check: ') else ''
    if args.update and header != old:
        file.write_text((header + '\n' if header else '') + body)
    records.append(dict(file=str(file.relative_to(root)), header=header, previous_header=old,
                        body_sha256=hashlib.sha256(body.encode()).hexdigest(), exit=result.returncode, stderr=result.stderr,
                        matches=header == old or args.update))
    print(file.relative_to(root), 'header changed' if header != old else 'header matches')
if args.update:
    manifest = json.loads((root / 'manifest.json').read_text())
    by_name = {row['file']: row['header'] for row in records}
    for family in manifest['families']:
        for side, name in family['files'].items():
            family['a_check'][side] = by_name[name]
    (root / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    (root / 'headers.json').write_text(json.dumps(records, indent=2) + '\n')
assert all(row['matches'] for row in records)
print(f'PASS {len(records)} in-place headers')

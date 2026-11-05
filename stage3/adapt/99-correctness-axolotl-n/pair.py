"""Project unchanged H10 finding coordinates onto main, then use the real pair tool.

The only edit expands one signature line. No finding or unit coordinate may
land at or after that edit on that line. Other lines retain exact columns.
Raw streams are preserved separately; this is an explicit coordinate projection,
not a claim that the two measured source files have identical bytes.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess

p = argparse.ArgumentParser()
p.add_argument('before', type=Path)
p.add_argument('after', type=Path)
p.add_argument('base_tree', type=Path)
p.add_argument('head_tree', type=Path)
p.add_argument('output', type=Path)
a = p.parse_args()
base, head = a.base_tree.resolve(), a.head_tree.resolve()
before = (base / 'src/compiler/debug.ts').read_bytes()
after = (head / 'src/compiler/debug.ts').read_bytes()
bl, al = before.decode().splitlines(), after.decode().splitlines()
assert len(bl) == len(al)
changes = [(i+1, x, y) for i, (x, y) in enumerate(zip(bl, al)) if x != y]
assert len(changes) == 1
line, x, y = changes[0]
start = next(i for i, (c, d) in enumerate(zip(x, y)) if c != d)
assert x.startswith('        function hasNode(')
coordinate = re.compile(re.escape(str(head)) + r'/src/compiler/debug\.ts:(\d+):(\d+)')
raw = a.after.read_text()
rows = [json.loads(row) for row in raw.splitlines()]
coordinates = [match for row in rows[1:] for item in (row.get('findings') or []) + (row.get('units') or []) for key in ('where', 'unit') for match in coordinate.findall(item.get(key, ''))]
assert all(int(l) != line or int(c)-1 < start for l, c in coordinates)
a.output.mkdir(parents=True, exist_ok=True)
projected = a.output / 'head-on-main.jsonl'
# Units retain parser offsets as well as line/column coordinates. The changed
# signature precedes its opening brace; every unit body offset lies outside it.
head_text = after.decode()
line_start = sum(len(part) for part in head_text.splitlines(keepends=True)[:line-1])
brace = line_start + y.rindex('{')
shift = len(y)-len(x)
for row in rows[1:]:
    for unit in row.get('units') or []:
        for field in ('body_start', 'body_end'):
            if field in unit:
                position = unit[field]
                assert position < line_start + start or position >= brace
                if position >= brace:
                    unit[field] = position-shift
projected.write_text(''.join(json.dumps(row).replace(str(head), str(base))+'\n' for row in rows))
provenance = dict(projection='tree prefix and unit body offsets; every finding/unit column unchanged',
    edited_line=line, first_changed_column=start+1, unit_body_offset_shift=len(y)-len(x),
    before_sha256=hashlib.sha256(before).hexdigest(),
    after_sha256=hashlib.sha256(after).hexdigest(),
    base_stream_sha256=hashlib.sha256(a.before.read_bytes()).hexdigest(),
    head_stream_sha256=hashlib.sha256(a.after.read_bytes()).hexdigest())
(a.output / 'projection.json').write_text(json.dumps(provenance, indent=2)+'\n')
repository = Path(__file__).resolve().parents[2]
subprocess.run(['python3', str(repository / 'census/pair/pair.py'), str(a.before),
    str(projected), str(base), '--typescript', str(repository / 'api/node_modules/typescript'),
    '--json', str(a.output / 'delta.json'), '--report', str(a.output / 'REPORT.md')], check=True)

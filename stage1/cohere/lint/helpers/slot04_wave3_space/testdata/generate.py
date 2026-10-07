"""Regenerate the pinned consumer subset and deterministic string controls."""
import json
from pathlib import Path
import random
HERE = Path(__file__).resolve().parent
BASE = HERE.parents[1]
ledger = json.loads((BASE / 'readiness.json').read_text())
suffixes = ('.isJavaScriptSpace', '.isBlank', '.isValueSeparator')
consumers = {r['rule'] for r in ledger['remaining'] if any(s.endswith(suffixes) for s in r['remaining_helpers'])}
rows = json.loads((BASE / 'testdata/slot04/tailwind-consumers.json').read_text())
rows = [r for r in rows if r['name'].split(':')[0] in consumers]
assert {r['name'].split(':')[0] for r in rows} == consumers
(HERE / 'consumers.json').write_text(json.dumps(rows, ensure_ascii=True, indent=2) + '\n')
white = [9, 10, 11, 12, 13, 32, 160, 5760, *range(8192, 8203), 8232, 8233, 8239, 8287, 12288, 65279]
other = [0, 8, 14, 133, 8203, 6158, 8288, 65, 47, 58, 44, 61, 62, 60, 128512, 1114111]
values = [''] + [chr(c) for c in white + other]
values += [''.join(chr(c) for c in white), '\r\n', ' / ', ' \ufeff\u2000 ', 'x\ufeff', '\ufeffx', '\ufeffx\ufeff']
values += [chr(a) + chr(b) for a in white + other for b in white + other]
rng = random.Random(4)
values.extend(''.join(chr(rng.choice(white + other)) for _ in range(rng.randrange(0, 30))) for _ in range(200))
(HERE / 'witnesses.json').write_text(json.dumps([{'Name': 'control:' + str(i), 'Source': s} for i, s in enumerate(values)], ensure_ascii=True, indent=2) + '\n')
print('consumer inputs', len(rows), 'rules', len(consumers), 'string controls', len(values))

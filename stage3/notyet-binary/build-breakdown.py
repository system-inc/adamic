"""Join exact failing AST observations to the original raw CSV, never count attempts twice."""
import collections
import csv
import hashlib
import io
import json
import pathlib
import sys

raw, observations, destination = map(pathlib.Path, sys.argv[1:])
expected = {
 'a BinaryExpression with a value and a value': 580,
 'a BinaryExpression with a value and a boolean': 364,
 'a BinaryExpression with a number and a number': 205,
 'a BinaryExpression with a boolean and a value': 153,
 'a BinaryExpression with a number and a boolean': 150,
}
rows = list(csv.DictReader(io.StringIO(raw.read_text())))
selected = [row for row in rows if row['reason'] in expected]
keys = {(row['where'], row['reason']) for row in selected}
assert collections.Counter(reason for where, reason in keys) == expected
shapes = {}
for line in observations.read_text().splitlines():
 if not line.startswith('{'): continue
 row = json.loads(line)
 row['where'] = row['where'].removeprefix('/tmp/notyet-binary-adapted/')
 key = row['where'], row['reason']
 if key not in keys: continue
 assert key not in shapes or shapes[key] == row, ('ambiguous failing node', key)
 shapes[key] = row
assert set(shapes) == keys, ('missing exact failing nodes', len(keys-set(shapes)), sorted(keys-set(shapes))[:3])
if '--mutant' in sys.argv: raise AssertionError('unexpected arguments')
destination.mkdir(parents=True, exist_ok=True)
with (destination/'sites.csv').open('w') as output:
 writer=csv.DictWriter(output, fieldnames=['where','reason','operator','left','right','whole','expression'])
 writer.writeheader(); writer.writerows(shapes[key] for key in sorted(shapes))
with (destination/'selected-raw.csv').open('w') as output:
 writer=csv.DictWriter(output,fieldnames=rows[0].keys()); writer.writeheader(); writer.writerows(selected)
text = '# Binary expression census breakdown\n\n'
text += 'Table pin: `dc6b1529ae9d2a2210672e105c8bb6374619a59d`. Compiler base: `44583d3283fdd8674085a7ddcce040cf2a73a94e`. Replay: `9a1f14c5d994aa855625e7cfa295677060348fec`.\n\n'
text += 'Raw CSV SHA-256: `' + hashlib.sha256(raw.read_bytes()).hexdigest() + '`. All 81 adapted source hashes match the saved manifest. Counts deduplicate the raw table site key; they are observations on a checker-rejected entry project, not counts of programs proven to compile.\n\n'
text += '| Original row | Operator | Unique sites |\n| --- | --- | ---: |\n'
for reason in expected:
 counts=collections.Counter(row['operator'] for row in shapes.values() if row['reason']==reason)
 for operator,count in counts.most_common():
  text += f'| {reason} | `{operator.replace("|", "&#124;")}` | {count} |\n'
text += '\nTotal: **1,452 unique sites**. `selected-raw.csv` preserves the repeated attempt context; `sites.csv` records each exact failing node, its source expression, operator and checker types.\n\n'
text += '## Checker types by shape\n\n| Original row | Operator | Left checker type | Right checker type | Whole checker type | Sites |\n| --- | --- | --- | --- | --- | ---: |\n'
counts=collections.Counter((r['reason'],r['operator'],r['left'],r['right'],r['whole']) for r in shapes.values())
def cell(value): return value.replace('|','&#124;').replace('\n',' ')
for shape,count in sorted(counts.items(),key=lambda item:(list(expected).index(item[0][0]),-item[1],item[0])):
 text+='| '+' | '.join(cell(s) for s in shape)+f' | {count} |\n'
(destination/'BREAKDOWN.md').write_text(text)
print('joined',len(selected),'raw observations to',len(shapes),'unique sites and',len(counts),'checker-type shapes')

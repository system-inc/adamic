#!/usr/bin/env python3
# Own-file diagnostics for named files, attributed the way stage3/meter/report.py does:
# every diagnostic in every census record, by its primary file:line:col, deduplicated.
import json, re, sys
from pathlib import Path

tree, census, output = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
names = sys.argv[4:]
wanted = {str((tree / 'src/compiler' / name).resolve()): 'src/compiler/' + name for name in names}
located = {key: set() for key in wanted}
for line in census.read_text().splitlines():
    for diagnostic in json.loads(line).get('diagnostics', []):
        match = re.match(r'^(.+?):[0-9]+:[0-9]+: error TS[0-9]+:', diagnostic)
        if match and str(Path(match[1]).resolve()) in located:
            located[str(Path(match[1]).resolve())].add(diagnostic)
lines = []
for key, relative in wanted.items():
    print(f'{relative} {len(located[key])}')
    for diagnostic in sorted(located[key], key=lambda d: [int(p) if p.isdigit() else p for p in re.split(r'(\d+)', d)]):
        lines.append(diagnostic.replace(str(tree.resolve()) + '/', '', 1))
print('total', sum(len(v) for v in located.values()))
output.write_text('\n'.join(lines) + '\n')

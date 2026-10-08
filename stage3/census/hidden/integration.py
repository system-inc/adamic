"""Measure real nested recovery and compare hidden bytes with a byte-set oracle."""
import json
import os
from pathlib import Path
import subprocess
import sys

import hidden

binary, output = Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve()
output.mkdir(parents=True, exist_ok=True)
source = output / 'source'
source.mkdir(exist_ok=True)
(source / 'probe.a').write_bytes(Path(__file__).with_name('testdata').joinpath('probe.a.txt').read_bytes())
with (output / 'census.log').open('w') as log:
    subprocess.run([str(binary), str(source), str(output / 'full.jsonl')], check=True,
        env=dict(os.environ, LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1'), stdout=log, stderr=log)
with (output / 'stock.log').open('w') as log:
    subprocess.run(['node', str(Path(__file__).with_name('units.cjs')), str(source), str(output / 'stock.json')],
        check=True, stdout=log, stderr=log)
rows = hidden.read_rows(output / 'full.jsonl')
stock = json.loads((output / 'stock.json').read_text())
result = hidden.calculate(rows, stock, source)
units = rows[1]['units']
parent = next(unit for unit in units if unit.get('name') == 'diagnosed')
clean = next(unit for unit in units if unit.get('name') == 'clean')
failed = next(unit for unit in units if unit.get('name') == 'failed')
signature = next(unit for unit in units if unit.get('name') == 'signature')
assert parent['status'] == 'split_checker_body'
assert clean['status'] == failed['status'] == signature['status'] == 'attempted'
boundaries = [finding for finding in rows[1]['findings'] if finding['kind'] == 'Boundary']
assert any(finding['unit'] == failed['where'] for finding in boundaries), 'failed child must have boundary'
assert any(finding['unit'] == signature['where'] and finding['start'] == signature['body_start']
    and finding['end'] == signature['body_end'] for finding in boundaries), 'whole-body signature boundary'
expected = set(range(parent['body_start'], parent['body_end']))
for name in ('clean', 'failed'):
    unit = next(unit for unit in stock['probe.a']['units'] if unit['where'] == hidden.local_where(
        next(unit for unit in units if unit.get('name') == name)['where'], source))
    expected -= set(range(unit['start'], unit['end']))
for finding in boundaries:
    expected |= set(range(finding['start'], finding['end']))
assert result['hidden_bytes'] == len(expected), (result['hidden_bytes'], len(expected))
assert result['independently_examined_bytes'] > 0, 'nested successful coverage must be subtracted'
print(f'PASS: real nested recovery, failed child retained, whole-body signature boundary; {len(expected)} hidden bytes against independent byte-set oracle')

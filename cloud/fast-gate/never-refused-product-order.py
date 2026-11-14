#!/usr/bin/env python3
"""Record the four independent Never-refusal mutant producers."""
import json
import sys
from pathlib import Path
path = Path(__file__).resolve().parent / 'product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
for name in ['Any', 'Eval', 'Function', 'Array']:
    order['waits']['internal/lower TestProduct_NeverRefusal' + name + 'Mutant'] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('run cloud/fast-gate/never-refused-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

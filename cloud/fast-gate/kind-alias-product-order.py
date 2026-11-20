#!/usr/bin/env python3
"""Regenerate U7's independent kind-alias mutant producer rows."""
import json
import sys
from pathlib import Path
path = Path(__file__).resolve().parent / 'product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
for name in ['Initializer', 'RunsFirst', 'Rewritten', 'Let']:
    order['waits']['internal/lower TestProduct_KindAlias' + name + 'Mutant'] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('run cloud/fast-gate/kind-alias-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

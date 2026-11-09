#!/usr/bin/env python3
"""Record the Stats guarded-path mutant producer, which builds private Go binaries."""
import json
import sys
from pathlib import Path
root = Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
for name in ('StatsOptionalReceiverMarkMutant', 'DirentOptionalReceiverMarkMutant'):
    order['waits']['internal/lower TestProduct_' + name] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('run cloud/fast-gate/stats-optional-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

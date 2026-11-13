#!/usr/bin/env python3
"""Record the independent Never-refusal mutant producer."""
import json
import sys
from pathlib import Path
path = Path(__file__).resolve().parent / 'product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
order['waits']['internal/lower TestProduct_NeverRefusalMutants'] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('run cloud/fast-gate/never-refused-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

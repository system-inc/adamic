#!/usr/bin/env python3
"""Regenerate family 17's independent enumeration accessor reach mutant producer rows."""
import json
import sys
from pathlib import Path
path = Path(__file__).resolve().parent / 'product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
for name in ['WholeGraph', 'UnionMembers', 'DeepReach', 'DiscriminatedUnion', 'GenericLiteral', 'LiteralTypeParameter', 'OneWay']:
    order['waits']['internal/lower TestProduct_EnumerationReach' + name + 'Mutant'] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('run cloud/fast-gate/enumeration-reach-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

#!/usr/bin/env python3
"""Regenerate U1's independent kind-set mutant producer rows."""
import json
import sys
from pathlib import Path
path = Path(__file__).resolve().parent / 'product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
for name in ['Disjunction', 'Switch', 'SwitchTail', 'Alias', 'Guard', 'DoubleGuard', 'IfGuard', 'FalseSide', 'FirstKindFit', 'NotEqual', 'EnumIdentity', 'PrimitiveGuard', 'TrueDefault', 'DuplicateCase', 'UnionClaimFit']:
    order['waits']['internal/lower TestProduct_KindSet' + name + 'Mutant'] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('run cloud/fast-gate/kind-set-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

#!/usr/bin/env python3
"""Regenerate the inspector doomed-use mutant product rows."""
import json
import sys
from pathlib import Path
root = Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
for mode in ('AnyRequire', 'Let', 'Receiver', 'ReadDrop', 'Fallthrough'):
    order['waits']['internal/lower TestProduct_InspectorDoomed' + mode + 'Mutant'] = []
for mode in ('Axiom', 'ReadDropOracle'):
    order['waits']['internal/oracle TestProduct_InspectorDoomed' + mode + 'Mutant'] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('inspector doomed product order is stale; regenerate with inspector-doomed-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

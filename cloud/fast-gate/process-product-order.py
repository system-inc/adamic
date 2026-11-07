#!/usr/bin/env python3
"""Regenerate the optional process-path mutant's product order row."""
import json
import sys
from pathlib import Path
root = Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
order['waits']['internal/lower TestProduct_ProcessOptionalEverywhereMutant'] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('process product order is stale; run cloud/fast-gate/process-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

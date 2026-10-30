#!/usr/bin/env python3
"""Regenerate the stdout-columns probe product's runtime-owner dependencies."""
import json
from pathlib import Path
root = Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
order['waits']['internal/oracle TestProduct_StdoutColumnsProbes'] = [
    'internal/oracle TestProduct_NativeRuntimesSanitized' + str(shard)
    for shard in range(3)
] + [
    'internal/oracle TestProduct_NativeRuntimesRelease0',
    'internal/oracle TestProduct_NativeRuntimesCounted0',
]
path.write_text(json.dumps(order, indent=2) + chr(10))

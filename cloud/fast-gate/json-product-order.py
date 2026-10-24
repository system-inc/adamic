#!/usr/bin/env python3
"""Regenerate JSON oracle runtime-owner waits from json_products_test.go."""
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parents[2]
source = (root / 'internal/oracle/json_products_test.go').read_text()
variants = re.findall(r'\{"(sanitized|release|counted)", "(\w+)"\}', source)
if len(variants) != 3 or {name for name, _ in variants} != {'sanitized', 'release', 'counted'}:
    raise SystemExit('unsupported JSON runtime variant table; no order was written')
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
if order.get('version') != 1:
    raise SystemExit('unsupported product order version')
order['waits']['internal/oracle TestProduct_JSONOracleRuntimes'] = [
    'internal/oracle TestProduct_NativeRuntimes' + owner + '0' for _, owner in variants
]
path.write_text(json.dumps(order, indent=2) + '\n')

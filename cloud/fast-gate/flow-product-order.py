#!/usr/bin/env python3
"""Regenerate flow corpus product declarations from the family table."""
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parents[2]
source = (root / 'internal/flow/families_test.go').read_text()
products = re.findall(r'\{"(TestProduct_FlowCorpusRemainder)", "\w+", (\d+), true\}', source)
if len(products) != 1:
    raise SystemExit('unsupported flow product family; no order was written')
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
for key in list(order['waits']):
    if key.startswith('internal/flow TestProduct_FlowCorpusRemainder_'):
        del order['waits'][key]
for name, shards in products:
    for shard in range(int(shards)):
        order['waits'][f'internal/flow {name}_{shard:03}'] = []
path.write_text(json.dumps(order, indent=2) + '\n')

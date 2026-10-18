#!/usr/bin/env python3
"""Regenerate command runtime owner waits from runtimes_test.go's declarations."""
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parents[2]
source = (root / 'cmd/adamic/runtimes_test.go').read_text()
products = dict(re.findall(
    r'func (TestProduct_\w+)\(t \*testing.T\)\s*\{\s*t.Parallel\(\);?\s*'
    r'commandRuntimeProduct\(t, "([^"]+)"\)', source))
tests = {name: test for test, name in products.items()}
table = source.split('var commandRuntimes = []commandRuntimeSpec{', 1)[1].split('\n}', 1)[0]
owners = re.findall(r'\{"([^"]+)", native.Options\{[^}]*\}, nil, "([^"]*)"\}', table)
if not owners or len(owners) != len(table.strip().splitlines()):
    raise SystemExit('unsupported command runtime declaration; no order was written')
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
for name, owner in owners:
    test = 'cmd/adamic ' + tests[name]
    if owner:
        waits = order['waits'].setdefault(test, [])
        if owner not in waits:
            waits.append(owner)
path.write_text(json.dumps(order, indent=2) + '\n')

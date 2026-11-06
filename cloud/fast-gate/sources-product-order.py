#!/usr/bin/env python3
"""Regenerate the readiness-source harness products' rows in the product order."""
import json
import re
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
prefixes = ('internal/native TestProduct_Sources', 'internal/native TestProduct_SourceMutant')
for key in list(order['waits']):
    if key.startswith(prefixes):
        del order['waits'][key]
source = (root / 'internal/native/sources_test.go').read_text()
names = re.findall(r'^func (TestProduct_Source\w+)\(', source, re.M)
if not names:
    raise SystemExit('no readiness-source products in sources_test.go')
for name in names:
    if name not in ('TestProduct_SourcesSanitized', 'TestProduct_SourcesRelease',
                    'TestProduct_SourcesCounted') and not name.startswith('TestProduct_SourceMutant'):
        raise SystemExit('unclassified readiness-source product: ' + name)
    # The harness builds its runtime archive inside its own product, as the
    # timer and watchFile harnesses do, and a mutant builds its private copy
    # there too, so none reads another product's archive or waits on one.
    order['waits']['internal/native ' + name] = []
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('readiness-source product order is stale; run cloud/fast-gate/sources-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

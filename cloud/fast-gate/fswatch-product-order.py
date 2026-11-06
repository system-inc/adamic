#!/usr/bin/env python3
"""Regenerate fs.watch products' rows in the product order, and their waits on the shared runtime archive owners."""
import json
import re
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
for key in list(order['waits']):
    if key.startswith(('internal/native TestProduct_FSWatch', 'internal/oracle TestProduct_FSWatch')):
        del order['waits'][key]
sets_source = (root / 'internal/native/runtimeproducts/runtimeproducts.go').read_text()
groups = int(re.search(r'const Groups = (\d+)', sets_source)[1])
for package, filename in [('internal/native', 'fswatch_test.go'),
                          ('internal/oracle', 'library_node_watch_test.go')]:
    source = (root / package / filename).read_text()
    names = re.findall(r'^func (TestProduct_FSWatch\w+)\(', source, re.M)
    if not names:
        raise SystemExit('no fs.watch products in ' + filename)
    for name in names:
        waits = []
        if package == 'internal/native':
            # The harness builds its runtime archive inside its own product, as the timer, watchFile
            # and readiness-source harnesses do, and its mutant a private copy there too.
            if name not in ('TestProduct_FSWatchHarnessRelease', 'TestProduct_FSWatchHarnessSanitized', 'TestProduct_FSWatchSharedHarness') and 'Mutant' not in name:
                raise SystemExit('unclassified native fs.watch product: ' + name)
        elif 'Mutant' not in name:
            # A program's control links the shared Node-host runtime archive, as watchFile's do.
            waits = ['internal/oracle TestProduct_NativeRuntimesSanitized' + str(i) for i in range(groups)]
        # Mutants build their own private archive inside their product.
        order['waits'][package + ' ' + name] = waits
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('fs.watch product order is stale; run cloud/fast-gate/fswatch-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

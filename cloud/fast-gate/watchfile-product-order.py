#!/usr/bin/env python3
"""Regenerate watchFile products' waits on the shared runtime archive owners."""
import json
import re
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
assert order['version'] == 1
for key in list(order['waits']):
    if key.startswith(('internal/native TestProduct_WatchFile',
                       'internal/oracle TestProduct_WatchFile')):
        del order['waits'][key]
sets_source = (root / 'internal/native/runtimeproducts/runtimeproducts.go').read_text()
sets = sets_source.split('var Sets = []Set{', 1)[1].split('\n}', 1)[0]
entries = re.findall(r'^\s*\{(nil|\[\]string\{[^}]*\}), "[^"]+"\}', sets, re.M)
if entries.count('[]string{"ADAMIC_NODE_HOST"}') != 1:
    raise SystemExit('cannot identify the Node-host runtime owner')
groups = int(re.search(r'const Groups = (\d+)', sets_source)[1])
host_group = str(entries.index('[]string{"ADAMIC_NODE_HOST"}') % groups)
for package, filename in [('internal/native', 'watchfile_test.go'),
                          ('internal/oracle', 'library_node_watch_file_test.go')]:
    source = (root / package / filename).read_text()
    names = re.findall(r'^func (TestProduct_WatchFile\w+)\(', source, re.M)
    if not names:
        raise SystemExit('no watchFile products in ' + filename)
    for name in names:
        waits = []
        if package == 'internal/native':
            # Native's private harness builds its archive in this product, like
            # the timer harness. It does not make native's test binary a lean
            # runtimeproducts builder (the runtime census checks that boundary).
            if name not in ('TestProduct_WatchFileCustomHooks',
                            'TestProduct_WatchFileProviderSanitized') and not name.endswith('Mutant'):
                raise SystemExit('unclassified native watchFile product: ' + name)
        elif 'Mutant' not in name:
            waits = ['internal/oracle TestProduct_NativeRuntimesSanitized' + str(i)
                     for i in range(groups)]
        # Mutants build their own private archive inside their product.
        order['waits'][package + ' ' + name] = waits
if '--check' in sys.argv:
    if json.loads(path.read_text()) != order:
        raise SystemExit('watchFile product order is stale; run cloud/fast-gate/watchfile-product-order.py')
else:
    path.write_text(json.dumps(order, indent=2) + '\n')

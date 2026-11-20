#!/usr/bin/env python3
"""Regenerate U2's independent kind-helper mutant producer rows."""
import json
import sys
from pathlib import Path
path=Path(__file__).resolve().parent/'product-order.json'
order=json.loads(path.read_text())
assert order['version']==1
for name in ['Depth', 'Cycle', 'NonHelper', 'WrittenParameter', 'WrongArgument', 'Call', 'Scalar', 'Default', 'Tail', 'ReturnBody', 'Arrow', 'DefaultMask', 'DefaultParameter', 'Enum', 'Overload', 'Let', 'AliasArgument', 'MixedDisjunction', 'DuplicateCase', 'EmptyDefaultTail']:
    order['waits']['internal/lower TestProduct_KindHelper'+name+'Mutant']=[]
if '--check' in sys.argv:
    if json.loads(path.read_text())!=order:
        raise SystemExit('run cloud/fast-gate/kind-helper-product-order.py')
else:
    path.write_text(json.dumps(order,indent=2)+'\n')

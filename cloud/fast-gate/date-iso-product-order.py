#!/usr/bin/env python3
"""Record the independent GCC audit producer for Date formatting."""
import json,sys
from pathlib import Path
path=Path(__file__).resolve().parent/'product-order.json'
order=json.loads(path.read_text())
order['waits']['internal/native TestProduct_GCCRuntimeDate']=[]
order['waits']['internal/lower TestProduct_DateISOStringMethodMutant']=[]
if '--check' in sys.argv:
    if json.loads(path.read_text())!=order: raise SystemExit('run cloud/fast-gate/date-iso-product-order.py')
else: path.write_text(json.dumps(order,indent=2)+'\n')

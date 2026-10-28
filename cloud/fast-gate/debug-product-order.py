#!/usr/bin/env python3
"""Declare the debug corpus products from their generated family wrappers."""
import json
import pathlib
import re
root = pathlib.Path(__file__).resolve().parents[2]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
for name in re.findall(r'func (TestProduct_Debug\w+)\(', (root/'internal/native/source_lines_families_generated_test.go').read_text()):
    order['waits']['internal/native '+name] = []
order['waits']['internal/native TestProduct_DebugLineTable'] = ['internal/native TestProduct_WASICompiler']
order['waits'].pop('internal/native TestProduct_DebugTwoDirectories', None)
for name in ['Release', 'Sanitized', 'Debug', 'DebugSanitized']:
    order['waits']['internal/native TestProduct_DebugRuntime'+name] = []
order['waits']['internal/native TestProduct_DebugCommandDriver'] = ['cmd/adamic TestProduct_CommandRuntimeDebug']
order['waits']['internal/native TestProduct_DebugRuntimeSharing'] = ['internal/native TestProduct_WASICompiler']
order['waits']['cmd/adamic TestProduct_CommandRuntimeDebug'] = []
path.write_text(json.dumps(order, indent=2)+'\n')

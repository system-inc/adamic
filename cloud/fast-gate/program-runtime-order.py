#!/usr/bin/env python3
"""Regenerate Program product waits from the current product-test inventory."""
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parents[2]
products = set()
for package in ('internal/lower', 'internal/oracle'):
    for source in (root / package).glob('*_test.go'):
        products.update(package + ' ' + name for name in re.findall(
            r'^func (TestProduct_\w+)\(t \*testing.T\)', source.read_text(), re.M))

path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())

def named(package, prefix):
    return sorted(p for p in products if p.startswith(package + ' ' + prefix))

def wait(product, owners):
    if product not in products or any(owner not in products for owner in owners):
        raise SystemExit('unknown Program product dependency: ' + product)
    waits = order['waits'].setdefault(product, [])
    for owner in owners:
        if owner not in waits:
            waits.append(owner)

program_counted = named('internal/oracle', 'TestProduct_ProgramRuntimesCounted')
parse = ['internal/oracle TestProduct_ProgramParseOn', 'internal/oracle TestProduct_ProgramParseOff']
for product in named('internal/oracle', 'TestProduct_CountMeasurements'):
    wait(product, parse + program_counted)
for product in named('internal/lower', 'TestProduct_ProgramRegionCensus'):
    wait(product, ['internal/oracle TestProduct_ProgramTscSource'])
for enabled in ('On', 'Off'):
    prefix = 'TestProduct_ProgramRuntimes' if enabled == 'On' else 'TestProduct_NativeRuntimes'
    for sanitized in (False, True):
        variant = 'Sanitized' if sanitized else 'Counted'
        product = 'internal/oracle TestProduct_ProgramParse' + ('Sanitized' if sanitized else '') + enabled
        wait(product, ['internal/oracle TestProduct_ProgramParseLower' + enabled] + named('internal/oracle', prefix + variant))
    wait('internal/oracle TestProduct_ProgramHost' + enabled, named('internal/oracle', prefix + 'Counted'))
    for variant in ('Counted', 'Sanitized'):
        product = 'internal/oracle TestProduct_ProgramMillionMeasured' + ('Sanitized' if variant == 'Sanitized' else '') + enabled
        wait(product, named('internal/oracle', prefix + variant))
path.write_text(json.dumps(order, indent=2) + '\n')

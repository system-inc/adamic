#!/usr/bin/env python3
"""Regenerate Worker product waits from the Worker family declarations."""
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parents[2]
source = (root / 'internal/worker/families_test.go').read_text()
products = dict((name, int(count)) for name, count in re.findall(
    r'\{"(TestProduct_\w+)", "\w+", (\d+), \w+, true\}', source))
expected = {'TestProduct_WorkerWasm', 'TestProduct_WorkerMutants', 'TestProduct_WorkerWasmMutants', 'TestProduct_WorkerJSON'}
if set(products) != expected:
    raise SystemExit('unsupported Worker product families; no order was written')
prefix = 'internal/worker '
compiler = prefix + 'TestProduct_WorkerCompiler'
checks = prefix + 'TestProduct_WorkerChecks'
seed = prefix + 'TestProduct_WorkerWasmTransform'
waits = {compiler: ['internal/native TestProduct_WASICompiler'], checks: [compiler], seed: [compiler]}
waits[prefix + "TestProduct_WorkerJSONTypes"] = []
waits[prefix + "TestProduct_WorkerDecodeJSONCorpus"] = []
waits[prefix + "TestProduct_WorkerEncodeJSONCorpus"] = []
waits[prefix + "TestProduct_WorkerComputeCorpus"] = [compiler]
waits[prefix + "TestProduct_WorkerComputeSieveMutant"] = [compiler]
waits[prefix + "TestProduct_WorkerdJavaScript"] = [compiler]
waits[prefix + "TestProduct_WorkerdWebAssembly"] = [compiler]
wasm = []
for index in range(products['TestProduct_WorkerWasm']):
    name = prefix + f'TestProduct_WorkerWasm_{index:03}'
    waits[name] = [seed, compiler] + (wasm[-1:] if index else [])
    wasm.append(name)
for index in range(products['TestProduct_WorkerMutants']):
    waits[prefix + f'TestProduct_WorkerMutants_{index:03}'] = [compiler, checks]
for index in range(products['TestProduct_WorkerJSON']):
    waits[prefix + f'TestProduct_WorkerJSON_{index:03}'] = [compiler]
waits['internal/worker/wasm TestProduct_WorkersWASMInputs'] = [compiler]
waits['internal/worker/wasm TestProduct_GeneratedCrossingInputs'] = [compiler, 'internal/worker/wasm TestProduct_WorkersWASMInputs']
waits['internal/split TestProduct_SplitOracleFixtures'] = ['internal/oracle TestProduct_OracleCompilingFixtures']
previous = []
for index in range(products['TestProduct_WorkerWasmMutants']):
    name = prefix + f'TestProduct_WorkerWasmMutants_{index:03}'
    waits[name] = [compiler, checks, seed] + wasm + previous
    previous = [name]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
order['waits'] = {key: value for key, value in order['waits'].items() if not key.startswith(prefix)}
order['waits'].update(waits)
path.write_text(json.dumps(order, indent=2) + '\n')

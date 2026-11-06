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
# The Wasm shards build different variants with the one unmutated compiler and the one wasm32-wasi runtime, so each
# waits on what it reads and nothing else: the compiler, and the transform seed, which builds that runtime first and
# which the counted variant copies (products_test.go's workerWasmProduct). The shards ran one after another, each on
# the last, to keep two processes from compiling the identical runtime at once; cachedRuntime's lock per archive key
# (internal/native/library.go, #0p6x7t7) does that now, and every product builds in a directory of its own, so they
# run beside each other (#a87r8sj: the chain set a compiler tree's products span, 953 s, its last 150 s one at a time).
for index in range(products['TestProduct_WorkerWasm']):
    waits[prefix + f'TestProduct_WorkerWasm_{index:03}'] = [seed, compiler]
for index in range(products['TestProduct_WorkerMutants']):
    waits[prefix + f'TestProduct_WorkerMutants_{index:03}'] = [compiler, checks]
for index in range(products['TestProduct_WorkerJSON']):
    waits[prefix + f'TestProduct_WorkerJSON_{index:03}'] = [compiler]
waits['internal/worker/wasm TestProduct_WorkersWASMInputs'] = [compiler]
waits['internal/worker/wasm TestProduct_GeneratedCrossingInputs'] = [compiler, 'internal/worker/wasm TestProduct_WorkersWASMInputs']
waits['internal/split TestProduct_SplitOracleFixtures'] = ['internal/oracle TestProduct_OracleCompilingFixtures']
# A Wasm mutant shard builds its own mutated compiler, worker.test and, for two checks, its own mutated transform
# and counted Workers, each keyed by its mutation and laid in an overlay and a build directory of its own (goplan's
# Run, buildcache's place). It reads the unmutated compiler, the checks' worker.test and the runtime the seed builds,
# and no other shard's product, so it waits on those three and runs beside the other shards (#a87r8sj).
for index in range(products['TestProduct_WorkerWasmMutants']):
    waits[prefix + f'TestProduct_WorkerWasmMutants_{index:03}'] = [compiler, checks, seed]
path = root / 'cloud/fast-gate/product-order.json'
order = json.loads(path.read_text())
order['waits'] = {key: value for key, value in order['waits'].items() if not key.startswith(prefix)}
order['waits'].update(waits)
path.write_text(json.dumps(order, indent=2) + '\n')

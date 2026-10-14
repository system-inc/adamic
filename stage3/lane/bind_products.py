#!/usr/bin/env python3
"""Rebind existing shard inventories to the lane runtime, without repartitioning."""
import json
from pathlib import Path
import sys
sys.dont_write_bytecode = True
from shard_plan import LANE, digest, runtime_hash, gate_manifest

plan = json.loads((LANE / 'shard-plan.json').read_text())
if len(plan['units']) != 151:
    raise ValueError('expected the existing 151-unit inventory ')
plan.pop('artifact', None)
plan.pop('api_digest', None)
plan.update(runtime=runtime_hash())
for unit in plan['units']:
    definition = {key: value for key, value in unit.items() if key != 'inputs_hash'}
    unit['inputs_hash'] = digest(dict(runtime=plan['runtime'], unit=definition))
(LANE / 'shard-plan.json').write_text(json.dumps(plan, indent=2) + '\n')
(LANE / 'shards.json').write_text(json.dumps(gate_manifest(plan), indent=2) + '\n')
print(json.dumps(dict(runtime=plan['runtime'], units=len(plan['units']))))

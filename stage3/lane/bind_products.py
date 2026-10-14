#!/usr/bin/env python3
"""Rebind existing shard inventories to a measured product artifact, without repartitioning."""
import argparse
import json
from pathlib import Path
import sys
sys.dont_write_bytecode = True
from shard_plan import LANE, digest, runtime_hash, gate_manifest

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('artifact', type=Path)
args = parser.parse_args()
ready = json.loads((args.artifact / 'ready.json').read_text())
api = json.loads((args.artifact / 'shards-ready.json').read_text())
plan = json.loads((LANE / 'shard-plan.json').read_text())
if len(plan['units']) != 151 or api['artifact'] != ready['key']:
    raise ValueError('expected the existing 151-unit inventory and matching artifact markers')
plan.update(artifact=ready['key'], api_digest=api['api_digest'], runtime=runtime_hash())
for unit in plan['units']:
    definition = {key: value for key, value in unit.items() if key != 'inputs_hash'}
    unit['inputs_hash'] = digest(dict(artifact=plan['artifact'], api=plan['api_digest'], runtime=plan['runtime'], unit=definition))
(LANE / 'shard-plan.json').write_text(json.dumps(plan, indent=2) + '\n')
(LANE / 'shards.json').write_text(json.dumps(gate_manifest(plan), indent=2) + '\n')
print(json.dumps(dict(artifact=plan['artifact'], api=plan['api_digest'], runtime=plan['runtime'], units=len(plan['units']))))

import json
from pathlib import Path
import subprocess
import sys

repo = Path(sys.argv[1])
sys.path.insert(0, str(repo / 'stage3/lane'))
import shard_plan

plan = shard_plan.load_plan()
assert shard_plan.runtime_hash() == plan['runtime'], 'runtime input hash drift'
rebuilt, _ = shard_plan.build_plan(plan['api_digest'])
assert rebuilt == plan, 'plan inputs have moved'
manifest = json.loads((shard_plan.LANE / 'shards.json').read_text())
assert shard_plan.gate_manifest(plan) == manifest, 'manifest inputs have moved'
subprocess.run([sys.executable, str(shard_plan.LANE / 'generate_shards.py'), '--check'], check=True)
subprocess.run(['git', '-C', str(repo), 'diff', '--exit-code', 'HEAD^', 'HEAD', '--',
                'stage3/lane/shard-plan.json', 'stage3/lane/shards.json',
                'stage3/lane/shards_generated_test.go'], check=True)
print('Runtime, calibration, partition and generated declaration inputs match the checked-in files.')
print('API input digest', plan['api_digest'])

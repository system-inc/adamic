"""Reproduce standalone accessor builds and record each command and observation."""
from concurrent.futures import ThreadPoolExecutor
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]
SCRATCH = Path('/tmp/adamic-accessors')
SCRATCH.mkdir(parents=True, exist_ok=True)


def run(command):
    result = subprocess.run(command, cwd=ROOT, capture_output=True, text=True, timeout=120)
    return {'command': command, 'exit': result.returncode,
            'stdout': result.stdout, 'stderr': result.stderr}


def observe(path):
    relative = str(path.relative_to(ROOT))
    binary = SCRATCH / path.stem
    result = {'program': relative}
    result['node'] = run(['node', '--disable-warning=ExperimentalWarning',
                          'oracle/node.mjs', relative])
    result['build'] = run(['go', 'run', './cmd/adamic', 'build', relative,
                           '-o', str(binary)])
    if result['build']['exit'] == 0:
        result['native'] = run([str(binary)])
        result['javascript_build'] = run(['go', 'run', './cmd/adamic', 'js', relative])
        if result['javascript_build']['exit'] == 0:
            javascript = binary.with_suffix('.mjs')
            javascript.write_text(result['javascript_build']['stdout'])
            result['javascript'] = run(['node', '--disable-warning=ExperimentalWarning',
                                        'oracle/node.mjs', str(javascript)])
    print(relative, 'build exit', result['build']['exit'], flush=True)
    return result


paths = sorted((ROOT / 'internal/oracle/testdata').glob('accessors_coverage_*.a'))
paths += sorted((ROOT / 'notes/accessors').glob('*/*.a'))
with ThreadPoolExecutor(max_workers=3) as pool:
    observations = list(pool.map(observe, paths))

for observation in observations:
    name = observation['program']
    if '/unsupported/' in name:
        assert observation['build']['exit'] != 0, name
        continue
    assert observation['build']['exit'] == 0, name
    assert observation['javascript_build']['exit'] == 0, name
    behavior = lambda run: (run['exit'], run['stdout'], run['stderr'])
    if '/differences/' in name:
        assert behavior(observation['node']) != behavior(observation['native']), name
    else:
        assert behavior(observation['node']) == behavior(observation['native']), name
        assert behavior(observation['node']) == behavior(observation['javascript']), name

# The generated JavaScript is reproducible; keep only its exit and command here.
for observation in observations:
    if 'javascript_build' in observation:
        del observation['javascript_build']['stdout']
(ROOT / 'notes/accessors/observations.json').write_text(json.dumps(observations, indent=2) + '\n')

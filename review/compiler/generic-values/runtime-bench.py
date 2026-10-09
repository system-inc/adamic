import json, pathlib, subprocess, time
root = pathlib.Path('/workspace/generic-values-proof-tmp')
results = {'load_before': pathlib.Path('/proc/loadavg').read_text().strip(), 'runs': {'base': [], 'head': []}}
for round in range(7):
 for name in (['base', 'head'] if round % 2 == 0 else ['head', 'base']):
  start = time.perf_counter()
  run = subprocess.run([str(root / ('runtime-' + name + '-set'))], capture_output=True, timeout=20, check=True)
  elapsed = time.perf_counter() - start
  assert run.stdout == b'30720000\n' and run.stderr == b'', run
  results['runs'][name].append(elapsed)
results['load_after'] = pathlib.Path('/proc/loadavg').read_text().strip()
results['best'] = {name: min(values) for name, values in results['runs'].items()}
results['percent'] = 100 * (results['best']['head'] / results['best']['base'] - 1)
print(json.dumps(results, indent=2))

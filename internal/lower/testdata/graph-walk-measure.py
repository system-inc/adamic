"""Run from the repository root with cloud/setup.sh's environment sourced.

The baseline is a Go source overlay, never a production compiler option. It
skips IR copying/site IDs and the corresponding ID selection, leaving the rest
of the flow analysis intact. These binaries measure lowering only, not execution.
"""
import hashlib
import json
import os
from pathlib import Path
import re
import statistics
import subprocess
import sys

root = Path.cwd()
out = Path(sys.argv[1] if len(sys.argv) > 1 else '/tmp/graph-walk-bench')
out.mkdir(parents=True, exist_ok=True)
source = root / 'internal/lower/graph_flow.go'
original = source.read_text()
start = original.index('func graphAllocationSites(')
end = original.index('// graphFlows works backwards', start)
baseline = original[:start] + '''func graphAllocationSites(value reflect.Value, next *int) reflect.Value {
    return value
}

''' + original[end:]
start = baseline.index('\t\tif field := value.FieldByName("GraphTypes"); field.IsValid() {')
end = baseline.index('\n\t\tswitch value := expression.(type)', start)
baseline = baseline[:start] + '''\t\tif field := value.FieldByName("GraphTypes"); field.IsValid() {
            return
        }
''' + baseline[end:]
(out / 'without-walk.go').write_text(baseline)
overlay = out / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(source): str(out / 'without-walk.go')}}))

# Only implementation imports, excluding libraries and adamic's prelude.
def manifest(entry):
    seen = {}
    def visit(path):
        path = path.resolve()
        if path in seen:
            return
        text = path.read_text()
        seen[path] = {'path': str(path.relative_to(root)), 'lines': len(text.splitlines()),
                      'bytes': len(text.encode()), 'sha256': hashlib.sha256(text.encode()).hexdigest()}
        for name in re.findall(r"(?:from\s+|import\s*)['\"]([^'\"]+)['\"]", text):
            if name.startswith('.'):
                visit(path.parent / name)
    visit(root / entry)
    return sorted(seen.values(), key=lambda row: row['path'])

results = {'head': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
           'graph_flow_sha256': hashlib.sha256(original.encode()).hexdigest(),
           'benchmark_sha256': hashlib.sha256((root / 'internal/lower/graph_flow_walk_test.go').read_bytes()).hexdigest(),
           'cpu_max': Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
           'nproc': subprocess.check_output(['nproc'], text=True).strip(),
           'baseline': 'skip allocation-site copy/IDs and ID selection only, using Go overlay',
           'samples': [], 'inputs': {},
           'rounds': int(os.environ.get('ADAMIC_GRAPH_BENCH_ROUNDS', '5')),
           'iterations': int(os.environ.get('ADAMIC_GRAPH_BENCH_ITERATIONS', '5'))}
entries = os.environ.get('ADAMIC_GRAPH_BENCH_ENTRIES', 'stage1/typescript/parser/main.ts,stage1/cohere/lint/main.ts,stage1/cohere/typeaware/volume_suite.ts').split(',')
for entry in entries:
    results['inputs'][entry] = manifest(entry)
    for cycle in [False, True]:
        env = dict(os.environ, ADAMIC_GRAPH_BENCH_SOURCE=str(root / entry),
                   ADAMIC_GRAPH_BENCH_CYCLE='1' if cycle else '0',
                   ADAMIC_GRAPH_BENCH_TSGO='1' if '/typeaware/' in entry else '0')
        # Separate precompiled binaries avoid any builds during timed rounds.
        binaries = {}
        for mode in ['with', 'without']:
            binary = out / f'{Path(entry).parent.name}-{mode}.test'
            cmd = ['go', 'test', '-c', '-o', str(binary), './internal/lower']
            if mode == 'without':
                cmd.insert(2, '-overlay=' + str(overlay))
            with (out / f'build-{Path(entry).parent.name}-{mode}.log').open('w') as log:
                subprocess.run(cmd, stdout=log, stderr=subprocess.STDOUT, check=True)
            binaries[mode] = binary
        for round_index in range(int(os.environ.get('ADAMIC_GRAPH_BENCH_ROUNDS', '5'))):
            order = ['with', 'without'] if round_index % 2 == 0 else ['without', 'with']
            for mode in order:
                name = f'{Path(entry).parent.name}-cycle{int(cycle)}-{round_index}-{mode}'
                cmd = [str(binaries[mode]), '-test.run=^$', '-test.bench=^BenchmarkGraphLowering$',
                       '-test.benchtime=' + os.environ.get('ADAMIC_GRAPH_BENCH_ITERATIONS', '5') + 'x', '-test.count=1', '-test.timeout=5m']
                log_path = out / (name + '.log')
                with log_path.open('w') as log:
                    subprocess.run(cmd, cwd=root / 'internal/lower', env=env,
                                   stdout=log, stderr=subprocess.STDOUT, check=True)
                text = log_path.read_text()
                match = re.search(r'BenchmarkGraphLowering-\d+\s+\d+\s+(\d+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op', text)
                if not match:
                    raise RuntimeError(f'missing measurement: {log_path}')
                row = {'entry': entry, 'cycle_overlay': cycle, 'round': round_index,
                       'mode': mode, 'ns': int(match[1]), 'bytes': int(match[2]),
                       'allocations': int(match[3]), 'log': str(log_path),
                       'graph_ids': int(re.search(r'graph IDs=(\d+)', text)[1])}
                results['samples'].append(row)
                print(name, row['ns'], row['bytes'], row['allocations'], flush=True)
results['isolated'] = []
for entry in results['inputs']:
    env = dict(os.environ, ADAMIC_GRAPH_BENCH_SOURCE=str(root / entry),
               ADAMIC_GRAPH_BENCH_TSGO='1' if '/typeaware/' in entry else '0')
    binary = out / f'{Path(entry).parent.name}-with.test'
    log_path = out / f'{Path(entry).parent.name}-isolated.log'
    with log_path.open('w') as log:
        subprocess.run([str(binary), '-test.run=^$', '-test.bench=^BenchmarkGraphAllocationSiteWalk$',
                        '-test.benchtime=1s', '-test.count=5', '-test.timeout=5m'],
                       cwd=root / 'internal/lower', env=env, stdout=log,
                       stderr=subprocess.STDOUT, check=True)
    samples = [{'ns': int(ns), 'bytes': int(size), 'allocations': int(allocations)}
               for ns, size, allocations in re.findall(r'BenchmarkGraphAllocationSiteWalk-\d+\s+\d+\s+(\d+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op', log_path.read_text())]
    if len(samples) != 5:
        raise RuntimeError(f'missing isolated measurements: {log_path}')
    results['isolated'].append({'entry': entry, 'samples': samples, 'log': str(log_path),
                                'median': {key: statistics.median(row[key] for row in samples)
                                           for key in ['ns', 'bytes', 'allocations']}})
results['summary'] = []
for entry in results['inputs']:
    for cycle in [False, True]:
        row = {'entry': entry, 'cycle_overlay': cycle}
        for mode in ['with', 'without']:
            rows = [sample for sample in results['samples'] if sample['entry'] == entry and sample['cycle_overlay'] == cycle and sample['mode'] == mode]
            row[mode] = {key: statistics.median(sample[key] for sample in rows)
                         for key in ['ns', 'bytes', 'allocations']}
        row['delta_ns'] = row['with']['ns'] - row['without']['ns']
        row['delta_percent'] = 100 * row['delta_ns'] / row['without']['ns']
        results['summary'].append(row)
(out / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
print(json.dumps(results['summary'], indent=2), flush=True)

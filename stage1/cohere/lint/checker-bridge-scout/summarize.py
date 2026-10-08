"""Render measured costs; do not forecast batching speedups."""
import argparse
import json
from pathlib import Path
from statistics import median

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('directory', type=Path)
args = parser.parse_args()

def records(name):
    value = json.loads((args.directory / name / 'results.json').read_text())
    return value['Records'] if isinstance(value, dict) else value

cold = records('perfile-1')
repeated = records('perfile-100')
profiled = records('profile')
index = {(r['file'], r['round']): r for r in cold}
files = list(dict.fromkeys(r['file'] for r in cold))
print(f'The controlled project loads {cold[0]["program_files"]} source files: 23 roots plus bundled libraries. '
      'Cold and repeated per-file runs each have 69 records (23 files × 3 rounds). '
      'Profiling has 23 separate records. Source hashes and all counters are retained in evidence/.\n')
print('| Whole program, median of 3 release runs | Native pilot | Independent Go rule |')
print('| --- | ---: | ---: |')
whole = records('whole-1')
native = [r for r in whole if r['implementation'] == 'native']
go = [r for r in whole if r['implementation'] == 'go']
print(f'| One pass, process | {median(r["process_ns"] for r in native)/1e6:.3f} ms | {median(r["process_ns"] for r in go)/1e6:.3f} ms |')
for label, key in [('Create/pool initialization', 'load_ns'), ('38 question adapters, including lazy Go work', 'query_ns'), ('After-create run', 'run_ns')]:
    print(f'| {label} | {median(r[key] for r in native)/1e6:.3f} ms | not separately instrumented |')
print('\nThe 100-pass whole-program native run makes 3,800 asks; its Go comparison executes one pass '
      'to check the first-pass output. Their repeated process times are not comparable throughput figures.\n')
print('| Per-file experiment, all 23 files × 3 rounds | Raw native ABI | Production Checker + rule |')
print('| --- | ---: | ---: |')
for label, rows, statistic in [
        ('Median complete-program create', cold, lambda r, g: r[g]['load_ns']/1e6),
        ('Median cold question time per file (µs)', cold, lambda r, g: r[g]['query_ns']/1e3),
        ('Median 100-pass question time / file / pass (µs)', repeated, lambda r, g: r[g]['query_ns']/100/1e3)]:
    values = [median(statistic(r, g) for r in rows) for g in ['native', 'pilot']]
    suffix = ' ms' if 'create' in label else ''
    print(f'| {label} | {values[0]:.3f}{suffix} | {values[1]:.3f}{suffix} |')
for group, label in [('native', 'Raw ABI'), ('pilot', 'Production pilot')]:
    delta = sum(r[group]['query_ns']-index[(r['file'], r['round'])][group]['query_ns'] for r in repeated)
    questions = sum(r['questions']*99 for r in repeated)
    print(f'\n{label}: complete first-pass subtraction yields **{delta/questions/1e3:.3f} µs per remaining question** '
          '(aggregate measured difference, separate cold/repeated runs). This includes Go work and scheduling. '
          'It is not a pure crossing latency or a proposed batch speedup.')
print('\n| Separate profiled run, weighted ns/question | Raw native ABI | Production pilot |')
print('| --- | ---: | ---: |')
for label, keys in [('Input conversion', ['input_ns']), ('Output decode/free', ['output_ns']), ('C-call minus timed Go inspection', ['call_ns', 'inspect_ns']), ('Timed Go inspection', ['inspect_ns'])]:
    values = []
    for group in ['native', 'pilot']:
        count = sum(r[group]['queries'] for r in profiled)
        total = sum(r[group][keys[0]]-(r[group][keys[1]] if len(keys)>1 else 0) for r in profiled)
        values.append(total/count)
    print(f'| {label} | {values[0]:.1f} | {values[1]:.1f} |')
print('\nThe larger production residual can include registry-lock/scheduling/Go-runtime effects while '
      'the surrounding Adamic rule allocates and executes. It must not be attributed entirely to cgo. '
      'Program creation and surrounding rule work are visible separately.\n')
print('| Public sample (full pins/paths in testdata/public-fixtures.json) | Asks | Findings | Median cold adapter time/file (µs) | Remaining-question difference (µs/q) |')
print('| --- | ---: | ---: | ---: | ---: |')
for file in files:
    c = [r for r in cold if r['file'] == file]
    warm = [r for r in repeated if r['file'] == file]
    name = file.split('/samples/', 1)[1]
    delta = median((r['pilot']['query_ns']-index[(file, r['round'])]['pilot']['query_ns'])/(99*r['questions'])/1e3 for r in warm)
    print(f'| `{name}` | {c[0]["questions"]} | {c[0]["findings"]} | {median(r["pilot"]["query_ns"] for r in c)/1e3:.3f} | {delta:.3f} |')

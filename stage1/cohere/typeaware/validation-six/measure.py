"""Recompute medians from the retained, unmodified test log."""
from pathlib import Path
import re
from statistics import median

log = Path(__file__).with_name('suite.log').read_text()
for name in ('suite', 'oracle'):
    rows = re.findall(r'all-77 round \d ' + name + r': process_s=([\d.]+) findings (\d+); (.+)', log)
    assert len(rows) == 3, (name, len(rows))
    values = [dict(re.findall(r'(\w+)=(\d+)', row[2])) for row in rows]
    process = median(float(row[0]) for row in rows)
    load = median(int(row['load_ns']) for row in values) / 1e9
    run = median(int(row['run_ns']) for row in values) / 1e9
    print(f'{name}: load_s={load:.9f} run_s={run:.9f} process_s={process:.6f} run_findings_s={1763/run:.3f} process_findings_s={1763/process:.3f}')
    if name == 'suite':
        queries = {int(row['queries']) for row in values}
        assert len(queries) == 1
        query = median(int(row['query_ns']) for row in values) / 1e9
        print(f'  queries={queries.pop()} query_s={query:.9f} average_query_us={query*1e6/131755:.3f}')
print('Warm identical-facts probes: median ns/query excluding first, 9999 remaining calls')
for question in ('assignable', 'declarations', 'signature', 'raw-type', 'nullable', 'union', 'options'):
    rows = re.findall(r'fact cost ' + question + r' round \d: native (.+); Go (.+); identical units=(\d+)', log)
    assert len(rows) == 3
    means = []
    for side in (0, 1):
        values = [dict(re.findall(r'(\w+)=(\d+)', row[side])) for row in rows]
        means.append(median((int(v['query_ns'])-int(v['first_query_ns']))/9999 for v in values))
    print(f'{question}: native={means[0]:.3f} direct_Go={means[1]:.3f} delta={means[0]-means[1]:.3f}')

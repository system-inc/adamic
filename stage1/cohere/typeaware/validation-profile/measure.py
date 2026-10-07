"""Recompute benchmark medians and profile attribution from retained logs."""
from pathlib import Path
import re
from statistics import median

root = Path(__file__).parent
log = root.joinpath('paired-bench.log').read_text()
for name in ('before', 'after', 'oracle'):
    rows = re.findall(r'round=\d name=' + name + r' process_s=([\d.]+) (.+)', log)
    assert len(rows) == 3
    values = [dict(re.findall(r'(\w+)=(\d+)', row[1])) for row in rows]
    process = median(float(row[0]) for row in rows)
    load = median(int(row['load_ns']) for row in values) / 1e9
    run = median(int(row['run_ns']) for row in values) / 1e9
    print(f'{name}: load_s={load:.9f} run_s={run:.9f} process_s={process:.9f} run_findings_s={1763/run:.3f} process_findings_s={1763/process:.3f}')
    if name != 'oracle':
        queries = {int(row['queries']) for row in values}
        assert len(queries) == 1
        count = queries.pop()
        query = median(int(row['query_ns']) for row in values) / 1e9
        print(f'  queries={count} query_s={query:.9f} average_query_us={query*1e6/count:.3f}')
for side in ('before', 'after'):
    log = root.joinpath(f'phases-{side}.stderr.log').read_text()
    phases = {k: int(v)/1e9 for k, v in re.findall(r'native_phase: adamic_function_\d+_(\w+)=(\d+)', log)}
    go = dict(re.findall(r'(\w+)=(\d+)', re.search(r'tsgo_go_profile: (.+)', log)[1]))
    c = dict(re.findall(r'(\w+)=(\d+)', re.search(r'tsgo_c_profile: (.+)', log)[1]))
    timing = dict(re.findall(r'(\w+)=(\d+)', re.search(r'tsgo: (.+)', log)[1]))
    gap = (int(c['call_ns'])-int(go['inspect_ns'])-int(go['parts_ns']))/1e9
    query = int(timing['query_ns'])/1e9
    run = int(timing['run_ns'])/1e9
    other = run-query-phases['types']-phases['Parser_file']-phases['Rules_links']-phases['byteOffsets']
    print(f'{side} profile: run_s={run:.6f} query_s={query:.6f} decode_s={phases["types"]:.6f} parse_s={phases["Parser_file"]:.6f} parents_s={phases["Rules_links"]:.6f} offsets_s={phases["byteOffsets"]:.6f} other_s={other:.6f}')
    print(f'  C-public-call minus Go-body: {gap:.6f}s, {gap*1e6/int(timing["queries"]):.3f}us/query; input_s={int(c["input_ns"])/1e9:.6f} output_s={int(c["output_ns"])/1e9:.6f}')
print('Warm identical-facts probes, median us/query excluding first, 9999 remaining calls')
log = root.joinpath('suite.log').read_text()
for question in ('assignable', 'declarations', 'signature', 'raw-type', 'nullable', 'union', 'signature-shape', 'raw-shape', 'type-shape', 'options'):
    rows = re.findall(r'fact cost ' + question + r' round \d: native (.+); Go (.+); identical units=(\d+)', log)
    assert len(rows) == 3
    means = []
    for side in (0, 1):
        values = [dict(re.findall(r'(\w+)=(\d+)', row[side])) for row in rows]
        means.append(median((int(v['query_ns'])-int(v['first_query_ns']))/9999 for v in values)/1000)
    print(f'{question}: native={means[0]:.3f} direct_Go={means[1]:.3f} delta={means[0]-means[1]:.3f}')

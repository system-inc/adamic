"""Recount observed diagnostics; unknown lowering counts remain null."""
import collections
import gzip
import hashlib
import json
import pathlib
import re
import shutil
import sys

scratch = pathlib.Path(sys.argv[1]).resolve()
adapted = pathlib.Path(sys.argv[2]).resolve()
out = pathlib.Path(__file__).resolve().parent
runs = json.loads((scratch / 'runs.json').read_text())
source = adapted / 'src/compiler'
files = sorted(p for p in source.rglob('*.ts'))
manifest = [{'file': str(p.relative_to(adapted)), 'sha256': hashlib.sha256(p.read_bytes()).hexdigest(), 'generated': '.generated.' in p.name} for p in files]
data = out / 'data'
data.mkdir(exist_ok=True)
result = {'complete': False, 'source': {'typescript': '6.0.3', 'commit': '050880ce59e30b356b686bd3144efe24f875ebc8', 'files': manifest}, 'runs': runs}

def normalize(value):
    if isinstance(value, str): return value.replace(str(adapted) + '/', '')
    if isinstance(value, list): return [normalize(v) for v in value]
    if isinstance(value, dict): return {k: normalize(v) for k, v in value.items()}
    return value

for row in runs:
    name = row['name']
    row['lowering_counts'] = None
    row['feature_lowering_delta'] = None
    row['per_file'] = {f['file']: {'NotYet': None, 'Refused': None, 'status': row['status']} for f in manifest}
    log = scratch / (name + '-' + ('run' if row['status'] == 'complete' else row['status']) + '.log')
    if log.exists(): shutil.copyfile(log, data / (name + '-' + log.name.split('-', 1)[1]))
    if row['status'] != 'complete': continue
    raw = [normalize(json.loads(line)) for line in (scratch / (name + '.jsonl')).read_text().splitlines()]
    with gzip.open(data / (name + '.jsonl.gz'), 'wt') as output:
        for record in raw: output.write(json.dumps(record) + '\n')
    header = raw[0]
    row['gate'] = header['status']
    diagnostics = header.get('diagnostics', [])
    row['checker_total'] = len(diagnostics)
    counts = collections.Counter()
    for diagnostic in diagnostics:
        code = re.search(r'error (TS\d+):', diagnostic)
        assert code, diagnostic
        counts[code[1]] += 1
    row['per_reason'] = dict(sorted(counts.items()))
    for record in raw[1:]:
        name = record['file']
        assert name in row['per_file'], name
        entry = row['per_file'][name]
        entry['status'] = record['status']
        entry['checker'] = len(record.get('checker_diagnostics', []))
        if record['status'] == 'measured':
            entry['NotYet'] = sum(f['kind'] == 'NotYet' for f in record['findings'])
            entry['Refused'] = sum(f['kind'] == 'Refused' for f in record['findings'])
    assert len(raw[1:]) == len(manifest)
    if header['status'] == 'measurement':
        lower_counts = collections.Counter()
        lower_reasons = collections.Counter()
        for record in raw[1:]:
            for finding in record['findings']:
                lower_counts[finding['kind']] += 1
                lower_reasons[finding['kind'] + ': ' + finding['reason']] += 1
        row['lowering_counts'] = dict(lower_counts)
        row['per_reason'] = dict(sorted(lower_reasons.items()))

baseline = runs[0]
for row in runs[1:]:
    if 'checker_total' in row and 'checker_total' in baseline:
        row['checker_delta'] = row['checker_total'] - baseline['checker_total']
        reasons = baseline['per_reason'].keys() | row['per_reason'].keys()
        row['checker_delta_per_reason'] = {reason: row['per_reason'].get(reason, 0) - baseline['per_reason'].get(reason, 0) for reason in sorted(reasons)}
(out / 'REPORT.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps([{k: v for k, v in row.items() if k not in ['per_file']} for row in runs], indent=2))

"""Render the pinned acceptance targets separately from compiler observations."""
import collections
import csv
import hashlib
import json
import pathlib
import subprocess

PIN = '3027b839'
RESOURCE = 'stage3/scouts/step24/factories/data/factories.json'
data = subprocess.check_output(['git', 'show', f'{PIN}:{RESOURCE}'])
rows = [r for r in json.loads(data) if r['factoryName'] and r['nodeResult']
        and not r['parser'] and r['location']['file'].startswith('src/compiler/factory/')]
assert len(rows) == 492
names = {f['name'] for r in rows for f in r['requiredFields']
         if f['status'] == 'written-undefined-before-escape'}
assert len(names) == 15
assert sum(f['status'] == 'written-undefined-before-escape'
           for r in rows for f in r['requiredFields']) == 428
out = pathlib.Path(__file__).parent
counts = collections.Counter()
with (out / 'factory-pairs.tsv').open('w') as stream:
    writer = csv.writer(stream, delimiter='\t', lineterminator='\n')
    writer.writerow(['factory', 'field', 'acceptance_target', 'source_evidence',
                     'observed_status', 'observed_boundary'])
    for row in rows:
        for field in row['requiredFields']:
            if field['phantom']:
                counts['excluded_type_brands'] += 1
                continue
            target = 'checked' if field['name'] in names else 'proven'
            counts['target_' + target] += 1
            counts['observed_refused'] += 1
            writer.writerow([row['id'], field['name'], target, field['status'],
                             'refused', 'builder.ts:1246:69 TS2345 before factory lowering'])
summary = {'evidence_pin': PIN, 'evidence_sha256': hashlib.sha256(data).hexdigest(),
           'factories': len(rows), 'explicit_undefined_pairs': 428,
           'deferred_names': sorted(names), 'counts': dict(counts),
           'observation': 'Full parser driver refused during Load; no source-only status is a native proof.'}
(out / 'factory-summary.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary['counts'], sort_keys=True))

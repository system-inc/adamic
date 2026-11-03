import collections, gzip, importlib.util, json, pathlib, sys
repo = pathlib.Path('/workspace/adamic')
run = repo/'stage3/meter/runs/20261103T200643Z.GK5jxe'
spec = importlib.util.spec_from_file_location('report', repo/'stage3/meter/report.py')
module = importlib.util.module_from_spec(spec); spec.loader.exec_module(module)
report = json.loads((run/'report.json').read_text())
main = report['trees']['main']
owners = json.loads((repo/'stage3/meter/owners.json').read_text())
with gzip.open(run/'main/census.jsonl.gz', 'rt') as source:
    records = [json.loads(line) for line in source]
groups = {}
ordinary_errors = [record for record in records if len(record['roots']) != 1 and record['kind'] == 'error']
for record in records:
    if len(record['roots']) != 1: continue
    name = module.site_where(record['roots'][0])
    if pathlib.Path(name).suffix not in ('.ts','.a'):
        if record['kind'] == 'error': ordinary_errors.append(dict(record, file=name))
        continue
    if record['kind'] == 'accepted': continue
    reason = record.get('reason', record['kind'])
    key = (record['kind'],reason)
    group = groups.setdefault(key, {'kind':key[0], 'reason':reason, 'owner':module.reason_owner(reason,owners),'count':0,'files':[]})
    group['count'] += 1
    group['files'].append({'file':name, 'where':module.site_where(record.get('where','')), 'fix':record.get('fix','')})
    if record['kind'] == 'error': ordinary_errors.append(dict(record, file=name))
summary = {
 'measured_commit':main['tree_commit'], 'measurement_sources':json.loads((run/'compiler-mode.json').read_text())['measurement_sources'],
 'whole_program_diagnostics':main['whole_program_diagnostics'],
 'tsc_entry_diagnostics':main['tsc_entry']['whole_program_diagnostics'],
 'own_file_diagnostics_sum':sum(row['own_file_diagnostics'] for row in main['files']),
 'totals':main['totals'], 'first_lowering_stops':sorted(groups.values(),key=lambda row:(-row['count'],row['kind'],row['reason'])),
 'latent_compiler_totals':main['latent_lowering']['totals'],
 'latent_entry_totals':main['tsc_entry']['lowering_census']['totals'],
 'latent_compiler_panics':main['latent_lowering']['panic_sites'],
 'latent_compiler_errors':main['latent_lowering']['error_sites'],
 'latent_entry_panics':main['tsc_entry']['lowering_census']['panic_sites'],
 'latent_entry_errors':main['tsc_entry']['lowering_census']['error_sites'],
 'ordinary_errors':ordinary_errors,
 'timing':json.loads((run/'timing.json').read_text()),
 'resources':json.loads(pathlib.Path('/workspace/scratch/meter-resources.json').read_text()),
}
(run/'measurement-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k not in ('first_lowering_stops','ordinary_errors','latent_compiler_panics','latent_compiler_errors','latent_entry_panics','latent_entry_errors')},indent=2))
for group in summary['first_lowering_stops']:
 print(group['count'], group['kind'], group['reason'], '|',group['owner'])
 print(', '.join(row['file'] for row in group['files']))

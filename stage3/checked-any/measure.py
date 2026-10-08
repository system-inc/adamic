"""Measure the pinned stock entry without converting fixture success into site coverage."""
import hashlib
import json
import pathlib
import subprocess
import sys

root = pathlib.Path(sys.argv[1]).resolve()
output = pathlib.Path(sys.argv[2]).resolve()
source_pin = '050880ce59e30b356b686bd3144efe24f875ebc8'
assert subprocess.check_output(['git', '-C', str(root), 'rev-parse', 'HEAD'], text=True).strip() == source_pin
read = lambda ref: json.loads(subprocess.check_output(['git', 'show', ref], text=True))
sites = read('ea1b2359:stage3/fixtures/checked-casts/outside-stock-casts.json')['stock_non_casts']
assert len(sites) == 271
hashes = read('855bcfaa:stage3/step09-ledger/files.json')['stock']
for item in hashes:
    assert hashlib.sha256((root / item['file']).read_bytes()).hexdigest() == item['sha256'], item['file']
output.mkdir(parents=True, exist_ok=True)
with (output / 'stock.stdout.log.txt').open('wb') as stdout, (output / 'stock.stderr.log.txt').open('wb') as stderr:
    result = subprocess.run(['go', 'run', './cmd/adamic', 'c', str(root / 'src/tsc/tsc.ts')], stdout=stdout, stderr=stderr)
stdout = (output / 'stock.stdout.log.txt').read_bytes()
stderr = (output / 'stock.stderr.log.txt').read_text()
first = stderr.splitlines()[0] if stderr else ''
assert result.returncode != 0 and not stdout and ': error TS' in first, 'update measurement classification if the first stop moves'
record = {'source_commit': source_pin, 'inventory_commit': 'ea1b2359', 'ledger_commit': '855bcfaa',
          'source_hashes_verified': len(hashes), 'sites': len(sites), 'compiler_exit': result.returncode,
          'emitted_c_bytes': len(stdout), 'verified_checked_sites_in_whole_entry': 0,
          'independent_site_lowering': 'not measured; project loader returned no program',
          'first_project_stop': first, 'records': []}
for site in sites:
    record['records'].append(dict(site, status='project_checker_stopped_before_lowering'))
(output / 'coverage-entry-before.json').write_text(json.dumps(record, indent=2) + '\n')
print(json.dumps({key: value for key, value in record.items() if key != 'records'}, indent=2))

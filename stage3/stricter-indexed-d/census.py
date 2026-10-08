"""Hold witness ownership to the exact pinned ledger filter."""
import csv
import io
import json
from pathlib import Path
import subprocess

files = {
    'src/compiler/' + file for file in (
        'transformers/es2015.ts', 'transformers/esnext.ts',
        'transformers/module/esnextAnd2015.ts', 'transformers/esDecorators.ts',
        'transformers/jsx.ts', 'program.ts', 'commandLineParser.ts',
        'moduleNameResolver.ts', 'sourcemap.ts', 'transformers/declarations.ts',
    )
}
ref = 'a1a16427a46149435e25f847c45ad136f1f1e55c'
ledger = subprocess.check_output([
    'git', 'show', ref + ':stage3/ledger/checker-259/rows-2026-10-08.csv',
], text=True)
rows = {
    row['id']: row for row in csv.DictReader(io.StringIO(ledger))
    if row['option'] == 'noUncheckedIndexedAccess' and row['file'] in files
}
sites = json.loads(Path(__file__).with_name('sites.json').read_text())
assert len(sites) == len({site['id'] for site in sites}), 'duplicate witness ID'
assert {site['id'] for site in sites} == rows.keys(), 'missing or extra witness ID'
for site in sites:
    row = rows[site['id']]
    assert site['file'] == row['file']
    assert site['line'] == int(row['line'])
    assert site['expression'] == row['source_expression']
    assert site['cause'] == row['cause']
print(f'{len(sites)} ledger rows, 0 missing, 0 extra, exact source metadata')
assert len(sites) == 27, 'expected all 27 rows after corrected ownership split'
print('Corrected ten-file option filter contains all 27 assigned rows.')

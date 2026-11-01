# Counts own-file diagnostics exactly as stage3/meter/report.py does: its report() over the census JSONL.
import json, sys, importlib.util, tempfile, shutil
from pathlib import Path
spec = importlib.util.spec_from_file_location('report', '/home/user/adamic/stage3/meter/report.py')
report = importlib.util.module_from_spec(spec); spec.loader.exec_module(report)
tree, jsonl = Path(sys.argv[1]), Path(sys.argv[2])
with tempfile.TemporaryDirectory() as run:
    shutil.copy(jsonl, Path(run) / 'census.jsonl')
    result = report.report(tree, Path(run), 'local')
rows = {row['file']: row for row in result['files']}
print('own file totals:', result['totals']['checker_own_file'], '/', result['totals']['source_files'])
for name in ('src/compiler/transformers/es2017.ts', 'src/compiler/transformers/classFields.ts'):
    print(name, 'own_file_diagnostics', rows[name]['own_file_diagnostics'], 'own_file', 'pass' if rows[name]['checker_own_file'] else 'fail')
located = set()
for line in jsonl.read_text().splitlines():
    for d in json.loads(line).get('diagnostics', []):
        if '/transformers/es2017.ts:' in d.split(' error')[0] or '/transformers/classFields.ts:' in d.split(' error')[0]:
            located.add(d.splitlines()[0])
for d in sorted(located): print('  ', d)

import json, sys
from pathlib import Path
sys.path.insert(0, '/home/user/adamic/stage3/meter')
sys.dont_write_bytecode = True
import report
def counts(tree, run):
    r = report.report(Path(tree), Path(run), 'probe', compiler_commit='0'*40)
    return {row['file']: row['own_file_diagnostics'] for row in r['files'] if row['source']}, r['totals']
pairs = [a.split('=') for a in sys.argv[1:]]
results = {label: counts(tree, run) for label, tree, run in [(p[0], *p[1].split(',')) for p in pairs]}
labels = list(results)
base = results[labels[0]][0]
print('es2015:', {l: results[l][0]['src/compiler/transformers/es2015.ts'] for l in labels})
print('own file:', {l: f"{results[l][1]['checker_own_file']}/{results[l][1]['source_files']}" for l in labels})
for l in labels[1:]:
    moved = {f: (base[f], results[l][0][f]) for f in base if base[f] != results[l][0][f]}
    print(l, 'moved vs', labels[0], moved)
for l in labels:
    print(l, 'nonzero', sum(1 for v in results[l][0].values() if v))

from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
import argparse
import json
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('before')
parser.add_argument('after')
parser.add_argument('output')
args = parser.parse_args()
root = Path(__file__).resolve().parents[3]
output = Path(args.output).resolve()
output.mkdir(parents=True, exist_ok=True)
paths = sorted(str(p.relative_to(root)) for directory in ['stage1', 'stage3/fixtures'] for p in (root / directory).rglob('*') if p.suffix in ['.a', '.ts'] and not p.name.endswith('.d.ts'))

def check(path):
    row = {'file': path}
    for tag, compiler in [('before', args.before), ('after', args.after)]:
        result = subprocess.run([compiler, 'c', str(root / path)], cwd=root, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True)
        row[tag] = result.returncode
        row[tag + '_diagnostic'] = result.stderr
    return row

with ThreadPoolExecutor(max_workers=4) as pool:
    rows = list(pool.map(check, paths))
(output / 'inventory.json').write_text(json.dumps(rows, indent=2))
(output / 'inventory.tsv').write_text('file\tbefore_exit\tafter_exit\n' + ''.join(f"{r['file']}\t{r['before']}\t{r['after']}\n" for r in rows))
changed = [r for r in rows if r['before'] == 0 and r['after'] != 0]
print('scanned', len(rows), 'baseline accepted', sum(r['before'] == 0 for r in rows), 'newly refused', len(changed))
for row in changed:
    print(row['file'], row['after_diagnostic'])

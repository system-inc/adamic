#!/usr/bin/env python3
"""Reconstruct batch 8 snapshots in scratch; program output always goes to files."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('directory', type=Path)
parser.add_argument('typescript', type=Path)
args = parser.parse_args()
repo = Path(__file__).resolve().parents[4]
base = args.directory.resolve()
snapshot = base / 'batch8'
snapshot.mkdir(parents=True, exist_ok=True)
pin = subprocess.check_output(['git', '-C', str(args.typescript), 'rev-parse', 'HEAD'], text=True).strip()
if pin != '050880ce59e30b356b686bd3144efe24f875ebc8':
    raise RuntimeError('wrong TypeScript corpus pin')

def historical(name):
    return subprocess.check_output(['git', 'show', '4189abd:' + name], cwd=repo)

names = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', '4189abd', 'stage1/cohere/lint/batch8'], cwd=repo, text=True).splitlines()
for name in names:
    if not name.endswith('.ts'):
        continue
    text = historical(name).decode().replace('../../../typescript', str(repo / 'stage1/typescript'))
    text = re.sub(r"(from '\./[^']+)\.ts'", r"\1.a'", text)
    (snapshot / (Path(name).stem + '.a')).write_text(text)
text = (snapshot / 'main.a').read_text()
assert text.count('    visit(context, root);') == 1
(snapshot / 'parse.a').write_text(text.replace('    visit(context, root);', '    // Parse-only measurement: no rule traversal.'))
files = sorted((args.typescript.resolve() / 'src/compiler').rglob('*.ts'))
if len(files) != 77:
    raise RuntimeError('wrong compiler file count')
(base / 'compiler.txt').write_text(''.join(str(p) + '\n' for p in files))
(base / 'corpus.sha256').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest() + '  ' + p.name + '\n' for p in files))
(base / 'batch8_oracle.go').write_bytes(historical('stage1/cohere/lint/testdata/batch8_oracle.go'))
(base / 'go_parse.go').write_bytes((Path(__file__).parent / 'testdata/go_parse.go').read_bytes())
for name, virtual, source in [('overlay.json', 'adamic_batch8_oracle.go', 'batch8_oracle.go'), ('parse-overlay.json', 'adamic_parse.go', 'go_parse.go')]:
    (base / name).write_text(json.dumps({'Replace': {str(repo / 'cohere' / virtual): str(base / source)}}) + '\n')
print(f'Prepared {len(files)} files, batch 8 full and parse-only .a drivers in {base}')

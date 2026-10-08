#!/usr/bin/env python3
"""Replay final landing costs. Children write stdout and stderr to files."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys

sys.dont_write_bytecode = True
arguments = argparse.ArgumentParser(description=__doc__)
arguments.add_argument('repo', type=Path)
arguments.add_argument('artifacts', type=Path)
arguments.add_argument('--valgrind', default='valgrind')
options = arguments.parse_args()
repo, root = options.repo.resolve(), options.artifacts.resolve()
manifest = root / 'compiler.manifest'
spec = importlib.util.spec_from_file_location('profile', repo / 'stage1/typescript/scanner/profile.py')
profile = importlib.util.module_from_spec(spec)
spec.loader.exec_module(profile)


def run(command, label):
    print('COMMAND', command, flush=True)
    with (root / (label + '.stdout')).open('wb') as out, (root / (label + '.stderr')).open('wb') as err:
        subprocess.run(command, cwd=repo, env=os.environ.copy(), stdout=out, stderr=err, check=True)


mutant = root / 'unit1-extra-parse/stage1/typescript'
for part in ['parser', 'scanner']:
    target = mutant / part
    target.mkdir(parents=True, exist_ok=True)
    for suffix in ['*.ts', '*.a']:
        for source in (repo / 'stage1/typescript' / part).glob(suffix):
            shutil.copyfile(source, target / source.name)
source = mutant / 'parser/main.ts'
text = source.read_text()
anchor = 'const parser = new Parser(source.text, path);'
assert text.count(anchor) == 1
source.write_text(text.replace(anchor, anchor + '\n        const extra = new Parser(source.text, path);\n        extra.file();'))
rows = {}
for label in ['base', 'before']:
    rows[label] = dict(instructions=profile.summarize(root / (label + '.callgrind'))['total'], nodes=887803)
for label, entry in [('after-final', repo / 'stage1/typescript/parser/main.ts'), ('extra-parse-final', mutant / 'parser/main.ts')]:
    binary = root / (label + '-native')
    run(['go', 'run', './cmd/adamic', 'build', str(entry), '-o', str(binary)], label + '-build')
    run([str(binary), '--manifest', str(manifest), '--whole'], label + '-tree')
    tree = (root / (label + '-tree.stdout')).read_bytes()
    assert tree == (root / 'go-tree.stdout').read_bytes()
    assert not (root / (label + '-tree.stderr')).stat().st_size
    print(label, 'tree bytes', len(tree), 'sha256', hashlib.sha256(tree).hexdigest(), flush=True)
    run([options.valgrind, '--tool=callgrind', '--callgrind-out-file=' + str(root / (label + '.callgrind')), str(binary), '--manifest', str(manifest), '--whole', '--count'], label)
    assert (root / (label + '.stdout')).read_text().strip() == '887803'
    rows[label] = dict(instructions=profile.summarize(root / (label + '.callgrind'))['total'], nodes=887803)
    print(label, rows[label], flush=True)
assert rows['after-final']['instructions'] < rows['before']['instructions'] < rows['extra-parse-final']['instructions']
print('actual extra parse caught by the unchanged before-recovery-cost budget', flush=True)
text = (root / 'after-final.callgrind').read_text()
mutated = re.sub(r'summary: (\d+)', lambda match: 'summary: ' + str(int(match[1]) + 1), text, count=1)
path = root / 'accounting-mutant-final.callgrind'
path.write_text(mutated)
try:
    profile.summarize(path)
except RuntimeError as error:
    print('accounting mutant caught:', error, flush=True)
else:
    raise AssertionError('accounting mutant survived')
(root / 'final-measurements.json').write_text(json.dumps(rows, indent=2) + '\n')
print(json.dumps(rows, indent=2), flush=True)

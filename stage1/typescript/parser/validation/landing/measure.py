#!/usr/bin/env python3
"""Final measurement replay. Every child writes stdout and stderr to files."""
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
sys.dont_write_bytecode = True
import argparse
arguments = argparse.ArgumentParser(description='Verify landing AST bytes and recovery instruction budget; baseline profiles must already exist.')
arguments.add_argument('repo', type=Path)
arguments.add_argument('artifacts', type=Path)
arguments.add_argument('--valgrind', default='valgrind')
options = arguments.parse_args()
repo = options.repo.resolve()
root = options.artifacts.resolve()
manifest = root / 'compiler.manifest'
vg = options.valgrind
env = os.environ.copy()
spec = importlib.util.spec_from_file_location('profile', repo / 'stage1/typescript/scanner/profile.py')
profile = importlib.util.module_from_spec(spec)
spec.loader.exec_module(profile)
def run(command, label):
    print('COMMAND', command, flush=True)
    with (root / (label + '.stdout')).open('wb') as out, (root / (label + '.stderr')).open('wb') as err:
        subprocess.run(command, cwd=repo, env=env, stdout=out, stderr=err, check=True)
mutant = root / 'final-extra-parse/stage1/typescript'
for part in ['parser', 'scanner']:
    target = mutant / part
    target.mkdir(parents=True, exist_ok=True)
    for source in (repo / 'stage1/typescript' / part).glob('*.ts'):
        shutil.copyfile(source, target / source.name)
source = mutant / 'parser/main.ts'
text = source.read_text()
anchor = 'const parser = new Parser(source.text, path);'
assert text.count(anchor) == 1
source.write_text(text.replace(anchor, anchor + '\n        const extra = new Parser(source.text, path);\n        extra.file();'))
for label, entry in [('final-landing', repo / 'stage1/typescript/parser/main.ts'), ('final-extra-parse', mutant / 'parser/main.ts')]:
    binary = root / (label + '-native')
    run(['go', 'run', './cmd/adamic', 'build', str(entry), '-o', str(binary)], label + '-build')
    run([str(binary), '--manifest', str(manifest), '--whole'], label + '-tree')
    tree = (root / (label + '-tree.stdout')).read_bytes()
    want = (root / 'go-tree.stdout').read_bytes()
    assert tree == want and not (root / (label + '-tree.stderr')).stat().st_size
    print(label, 'identical tree bytes', len(tree), 'sha256', hashlib.sha256(tree).hexdigest(), flush=True)
    run([vg, '--tool=callgrind', '--callgrind-out-file=' + str(root / (label + '.callgrind')), str(binary), '--manifest', str(manifest), '--whole', '--count'], label)
    assert (root / (label + '.stdout')).read_text().strip() == '887803'
    result = profile.summarize(root / (label + '.callgrind'))
    print(label, 'Ir', result['total'], 'nodes', 887803, flush=True)
rows = {}
for label in ['base', 'area', 'recovery', 'final-landing', 'final-extra-parse']:
    rows[label] = dict(instructions=profile.summarize(root / (label + '.callgrind'))['total'], nodes=887803)
added = rows['recovery']['instructions'] - rows['base']['instructions']
budget = rows['area']['instructions'] + added
assert rows['final-landing']['instructions'] <= budget, (rows, budget)
assert rows['final-extra-parse']['instructions'] > budget, (rows, budget)
result = dict(rows=rows, recovery_added=added, landing_added=rows['final-landing']['instructions'] - rows['area']['instructions'], budget=budget, margin=budget-rows['final-landing']['instructions'], mutant_excess=rows['final-extra-parse']['instructions'] - budget)
(root / 'final-measurements.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result, indent=2), flush=True)

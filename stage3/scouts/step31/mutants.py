#!/usr/bin/env python3
"""Mutate actual binder bodies in disposable trees and require observable Node differences."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('tree', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
if args.output.exists():
    parser.error('output must be new')
args.output.mkdir(parents=True)
request = ROOT / 'fixtures/projects.json'

def run(tree, name):
    with (args.output / (name + '.stdout')).open('wb') as out, (args.output / (name + '.stderr')).open('wb') as err:
        result = subprocess.run(['node', str(ROOT / 'source.mjs'), str(tree), str(request)], stdout=out, stderr=err, timeout=120)
    (args.output / (name + '.exit')).write_text(str(result.returncode) + '\n')
    if result.returncode != 0 or (args.output / (name + '.stderr')).read_bytes():
        raise RuntimeError(name + ': execution failure does not count as a semantic kill')
    return (args.output / (name + '.stdout')).read_bytes()

baseline = run(args.tree.resolve(), 'baseline')
if baseline != (ROOT / 'fixtures/golden.jsonl').read_bytes():
    raise RuntimeError('real binder source does not match stock golden')
mutants = [
    ('function-flags', 'bindFunctionDeclaration', 'SymbolFlags.Function, SymbolFlags.FunctionExcludes', 'SymbolFlags.BlockScopedVariable, SymbolFlags.FunctionExcludes'),
    ('variable-flags', 'bindVariableDeclarationOrBindingElement', 'SymbolFlags.BlockScopedVariable, SymbolFlags.BlockScopedVariableExcludes', 'SymbolFlags.FunctionScopedVariable, SymbolFlags.BlockScopedVariableExcludes'),
    ('declaration-flags', 'addDeclarationToSymbol', 'symbol.flags |= symbolFlags;', 'symbol.flags = 0;'),
]
report = []
for name, function, before, after in mutants:
    tree = args.output / name
    shutil.copytree(args.tree / 'src', tree / 'src')
    source = tree / 'src/compiler/binder.ts'
    text = source.read_text()
    start = text.index('    function ' + function + '(')
    end = text.find('\n    function ', start + 1)
    if end < 0:
        raise RuntimeError('missing next function')
    body = text[start:end]
    if before not in body:
        raise RuntimeError('mutant anchor changed')
    # Function flags occur in the strict and nonstrict arms; both are one semantic mutation.
    source.write_text(text[:start] + body.replace(before, after) + text[end:])
    actual = run(tree, name)
    if actual == baseline:
        raise RuntimeError('mutant survived: ' + name)
    left = [json.loads(line) for line in baseline.splitlines()]
    right = [json.loads(line) for line in actual.splitlines()]
    if len(left) != len(right):
        raise RuntimeError('record population changed')
    changed = [a['path'] for a, b in zip(left, right) if a != b]
    row = {'mutant': name, 'function': function, 'before': before, 'after': after,
        'exit': 0, 'stderr_bytes': 0, 'caught_by': 'binder dump bytes versus independent stock golden',
        'changed_files': changed, 'baseline_sha256': hashlib.sha256(baseline).hexdigest(),
        'mutant_sha256': hashlib.sha256(actual).hexdigest()}
    report.append(row)
    print(name + ': caught, successful Node execution, changed ' + ', '.join(changed), flush=True)
(args.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')

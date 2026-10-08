#!/usr/bin/env python3
"""Require an effective no-tags parser mutant and a JSDoc-only diagnostic mutant."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

p = argparse.ArgumentParser()
p.add_argument('tree', type=Path)
p.add_argument('inputs', type=Path)
p.add_argument('output', type=Path)
p.add_argument('--baseline', type=Path)
a = p.parse_args()
a.tree, a.inputs, a.output = a.tree.resolve(), a.inputs.resolve(), a.output.resolve()
assert not a.output.is_relative_to(a.tree), 'mutant output must be outside parser tree'
a.output.mkdir(exist_ok=False)
here = Path(__file__).resolve().parent
source = (here / 'main.a').read_text()
manifest = a.output / 'manifest'
manifest.write_text(''.join(path.relative_to(a.inputs / 'src/compiler').as_posix() + '\n'
                          for path in sorted((a.inputs / 'src/compiler').rglob('*')) if path.is_file()))
env = dict(os.environ, PARSER_RUNTIME=str(here.parents[2] / 'oracle/adamic.mjs'))
assert env.get('PARSER_TYPESCRIPT'), 'set PARSER_TYPESCRIPT to stock 6.0.3'

def execute(entry, inputs, manifest, stem):
    with (a.output / (stem + '.dump')).open('wb') as stdout, (a.output / (stem + '.stderr')).open('wb') as stderr:
        result = subprocess.run(['node', '--disable-warning=ExperimentalWarning', str(here / 'node.mjs'),
                                 str(entry), str(inputs), str(manifest)], env=env, stdout=stdout, stderr=stderr)
    assert result.returncode == 0, (stem, result.returncode)
    assert (a.output / (stem + '.stderr')).stat().st_size == 0, stem
    return a.output / (stem + '.dump')

def observe(path):
    data = path.read_bytes()
    lines = data.decode().splitlines()
    return {'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest(), 'exit': 0,
            'jsdoc_nodes': sum(line.startswith('JSDoc ') for line in lines),
            'tag_nodes': sum('\ttagName ' in line for line in lines),
            'type_expressions': sum(line.startswith('JSDocTypeExpression ') for line in lines),
            'jsdoc_diagnostics': sum(line.startswith('jsDocDiagnostic ') for line in lines)}

def differs(before, after, stem):
    with (a.output / (stem + '.log')).open('wb') as log:
        result = subprocess.run(['cmp', str(before), str(after)], stdout=log, stderr=log)
    assert result.returncode == 1, (stem, result.returncode)
    return result.returncode

mutant = a.output / 'dropped-tags-tree'
shutil.copytree(a.tree, mutant, ignore=shutil.ignore_patterns('.git', 'node_modules', 'built', 'build'))
(mutant / 'parser-proof-main.a').write_text(source)
shutil.copyfile(here / 'kinds.a', mutant / 'kinds.a')
parser = mutant / 'src/compiler/parser.ts'
text = parser.read_bytes()
needle = b'const tagsArray = tags && createNodeArray(tags, tagsPos, tagsEnd);'
assert text.count(needle) == 1, 'JSDoc return-site drift'
parser.write_bytes(text.replace(needle, b'const tagsArray = undefined;'))
# Same mutation as the demonstrated legacy blind spot: consume tags, return none.
baseline = a.baseline.resolve() if a.baseline else execute(a.tree / 'parser-proof-main.a', a.inputs, manifest, 'control')
changed = execute(mutant / 'parser-proof-main.a', a.inputs, manifest, 'dropped-tags')
normal, dropped = observe(baseline), observe(changed)
assert normal['tag_nodes'] > 0 and dropped['tag_nodes'] == 0
assert dropped['jsdoc_nodes'] == normal['jsdoc_nodes'] > 0
report = {'dropped_tags': {'control': normal, 'mutant': dropped,
                         'comparison_exit': differs(baseline, changed, 'dropped-tags-comparison'),
                         'caught_by': 'complete extended dump; both Node runs exit zero, same JSDoc count but zero tags'}}

# The TS compiler corpus has empty jsDocDiagnostics. Real JS input is required.
cases = a.output / 'directed-inputs'
(cases / 'src/compiler').mkdir(parents=True)
for name, case_text in json.loads((here / 'jsdoc-inputs.json').read_text()).items():
    (cases / 'src/compiler' / name).write_text(case_text)
case_manifest = a.output / 'directed-manifest'
case_manifest.write_text(''.join(name + '\n' for name in sorted(json.loads((here / 'jsdoc-inputs.json').read_text()))))
control = execute(a.tree / 'parser-proof-main.a', cases, case_manifest, 'directed-control')
needle = '    dump(file);\n'
assert source.count(needle) == 1
mutation = "    const firstDocDiagnostic = file.jsDocDiagnostics?.[0];\n    if (firstDocDiagnostic !== undefined) { firstDocDiagnostic.code += 1; }\n"
entry = mutant / 'parser-proof-jsdoc-diagnostic-mutant.a'
entry.write_text(source.replace(needle, mutation + needle))
# Restore parser: this mutant changes only the actual diagnostic's code.
parser.write_bytes(text)
changed = execute(entry, cases, case_manifest, 'jsdoc-diagnostic')
before, after = control.read_text().splitlines(), changed.read_text().splitlines()
assert len(before) == len(after)
differences = [(index + 1, x, y) for index, (x, y) in enumerate(zip(before, after)) if x != y]
assert len(differences) == 1
line, old, new = differences[0]
x, y = old.split('\t')[0].split(), new.split('\t')[0].split()
assert x[0] == y[0] == 'jsDocDiagnostic' and x[1] == y[1] and int(y[2]) == int(x[2]) + 1
assert x[3:] == y[3:] and old.split('\t')[1:] == new.split('\t')[1:]
report['jsdoc_diagnostic'] = {'control': observe(control), 'mutant': observe(changed),
                            'comparison_exit': differs(control, changed, 'jsdoc-diagnostic-comparison'),
                            'line': line, 'before': old, 'after': new,
                            'caught_by': 'only JSDoc diagnostic code changes; trees and parse diagnostics remain identical'}
(a.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report, indent=2))

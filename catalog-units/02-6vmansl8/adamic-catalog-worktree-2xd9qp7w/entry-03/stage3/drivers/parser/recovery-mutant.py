#!/usr/bin/env python3
"""Suppress a real parser recovery diagnostic; require the case reference to reject it."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

p = argparse.ArgumentParser()
p.add_argument('slice', type=Path)
p.add_argument('compiler_inputs', type=Path)
p.add_argument('compiler_run', type=Path)
p.add_argument('case_run', type=Path)
p.add_argument('output', type=Path)
a = p.parse_args()
for name in vars(a): setattr(a, name, getattr(a, name).resolve())
assert not a.output.is_relative_to(a.slice)
a.output.mkdir(exist_ok=False)
here = Path(__file__).resolve().parent
reference = json.loads((here / 'cases-reference.json').read_text())
mutant = a.output / 'mutant-tree'
shutil.copytree(a.slice, mutant, ignore=shutil.ignore_patterns('.git', 'node_modules', 'build', 'built'))
shutil.copyfile(here / 'main.a', mutant / 'parser-proof-main.a')
shutil.copyfile(here / 'kinds.a', mutant / 'kinds.a')
parser = mutant / 'src/compiler/parser.ts'
source = parser.read_bytes()
needle = b'parseErrorAtCurrentToken(Diagnostics.or_JSX_element_expected);'
assert source.count(needle) == 1, 'recovery call site drift'
parser.write_bytes(source.replace(needle, b'void 0;'))
env = dict(os.environ, PARSER_RUNTIME=str(here.parents[2] / 'oracle/adamic.mjs'))
assert env.get('PARSER_TYPESCRIPT')
def run(inputs, manifest, label):
    dump, stderr = a.output / (label + '.dump'), a.output / (label + '.stderr')
    with dump.open('wb') as out, stderr.open('wb') as error:
        result = subprocess.run(['node', '--disable-warning=ExperimentalWarning', str(here / 'node.mjs'),
                                 str(mutant / 'parser-proof-main.a'), str(inputs), str(manifest)],
                                env=env, stdout=out, stderr=error)
    assert result.returncode == 0 and stderr.stat().st_size == 0, (label, result.returncode)
    return dump
old = run(a.compiler_inputs, a.compiler_run / 'manifest', 'compiler-mutant')
old_bytes = old.read_bytes()
old_ref = json.loads((here / 'reference.json').read_text())
assert len(old_bytes) == old_ref['bytes'] and hashlib.sha256(old_bytes).hexdigest() == old_ref['sha256']
with (a.output / 'old-comparison.log').open('wb') as log:
    old_cmp = subprocess.run(['cmp', str(a.compiler_run / 'node.dump'), str(old)], stdout=log, stderr=log)
assert old_cmp.returncode == 0
changed = run(a.case_run / 'inputs', a.case_run / 'input-manifest', 'cases-mutant')
def written(name):
    data = name.encode('utf-16-le', 'surrogatepass')
    return ''.join(chr(n) if 32 <= n <= 126 and n != 92 else '\\u' + format(n, '04x')
                   for n in (int.from_bytes(data[i:i+2], 'little') for i in range(0, len(data), 2)))
expected = {written(r['input_path']): r for r in reference['cases']}
observed = {}
key = None
with changed.open('rb') as stream:
    for line in stream:
        if line.startswith(b'file\t'):
            if key is not None: row['sha256'] = digest.hexdigest()
            key = line[5:].decode('ascii').rstrip('\n')
            assert key in expected and key not in observed
            row = {'diagnostic_rows': 0, 'node_count': 0}
            observed[key] = row
            digest = hashlib.sha256()
        assert key is not None
        digest.update(line)
        if line.startswith(b'diagnostic '): row['diagnostic_rows'] += 1
        elif not line.startswith((b'file\t', b'diagnostics ', b'jsDocDiagnostic')): row['node_count'] += 1
    if key is not None: row['sha256'] = digest.hexdigest()
assert len(observed) == len(expected)
differences = [{'path': expected[k]['path'], 'control': expected[k]['node'], 'mutant': row}
               for k, row in observed.items() if row['sha256'] != expected[k]['node']['sha256']]
assert differences, 'the error recovery mutant must change the case reference'
assert any(r['control']['diagnostic_rows'] != r['mutant']['diagnostic_rows'] for r in differences)
with (a.output / 'case-comparison.log').open('wb') as log:
    cmp = subprocess.run(['cmp', str(a.case_run / 'slice-node.dump'), str(changed)], stdout=log, stderr=log)
assert cmp.returncode == 1
report = {'mutation': 'parseJsxAttributeValue omits parseErrorAtCurrentToken(Diagnostics.or_JSX_element_expected)',
          'node_exit': 0, 'stderr_bytes': 0,
          'old_reference': {'bytes': len(old_bytes), 'sha256': hashlib.sha256(old_bytes).hexdigest(),
                            'comparison_exit': old_cmp.returncode, 'caught': False},
          'case_reference': {'cases': len(expected), 'changed_cases': len(differences),
                             'diagnostic_rows_before': sum(r['node']['diagnostic_rows'] for r in expected.values()),
                             'diagnostic_rows_after': sum(r['diagnostic_rows'] for r in observed.values()),
                             'node_count_after': sum(r['node_count'] for r in observed.values()),
                             'comparison_exit': cmp.returncode, 'caught': True},
          'differences': differences}
(a.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps({k: v for k, v in report.items() if k != 'differences'}, indent=2))

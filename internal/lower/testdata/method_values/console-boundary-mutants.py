#!/usr/bin/env python3
"""Kill the console proof's omission and unsafe-admission mutants; evidence stays external."""
import argparse
import json
from pathlib import Path
import subprocess

repo = Path(__file__).resolve().parents[4]
parser = argparse.ArgumentParser()
parser.add_argument('--output', required=True, type=Path)
args = parser.parse_args()
out = args.output.resolve()
assert out != repo and repo not in out.parents, 'run output must be outside the checkout'
out.mkdir(parents=True, exist_ok=True)
source_path = repo / 'internal/lower/library_method_values.go'
source = source_path.read_text()
body = source.index('{', source.index('func (l *lowering) libraryMethodConsoleScalar('))
mutants = {
    'drop': '\nreturn false\n}\n',
    'unsafe': '''
outer := node
for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression { outer = outer.Parent }
parent := outer.Parent
return parent != nil && parent.Kind == ast.KindCallExpression && l.isConsole(parent.AsCallExpression().Expression)
}
''',
}
for name, replacement in mutants.items():
    mutated = out / (name + '.go')
    mutated.write_text(source[:body + 1] + replacement)
    (out / (name + '.json')).write_text(json.dumps({'Replace': {str(source_path): str(mutated)}}))


def run(command, label):
    result = subprocess.run(command, cwd=repo, capture_output=True)
    (out / (label + '.stdout')).write_bytes(result.stdout)
    (out / (label + '.stderr')).write_bytes(result.stderr)
    (out / (label + '.exit')).write_text(str(result.returncode) + '\n')
    return result


for name in ('baseline', 'drop'):
    command = ['go', 'build']
    if name == 'drop':
        command += ['-overlay', str(out / 'drop.json')]
    command += ['-o', str(out / name), './cmd/adamic']
    built = run(command, name + '-build')
    assert built.returncode == 0, built.stderr.decode()

unsafe = run(['go', 'test', '-overlay', str(out / 'unsafe.json'), './internal/lower',
              '-run', '^TestLibraryMethodConsoleScalarBoundaryRefusals$', '-count=1'], 'unsafe')
assert unsafe.returncode != 0 and b'unproven alias result or general nullable view was admitted' in unsafe.stdout, unsafe.stdout.decode()
assert b"String.prototype.at;_console.log(f.call('x',_9))" in unsafe.stdout, unsafe.stdout.decode()

positive = run(['go', 'test', '-overlay', str(out / 'drop.json'), './internal/lower',
                '-run', '^TestLibraryMethodConsoleScalarBoundary$', '-count=1'], 'drop-boundary-tests')
assert positive.returncode != 0 and b'a delayed library result whose contextual representation is not proven' in positive.stdout, positive.stdout.decode()

affected = []
for fixture in sorted((repo / 'internal/oracle/testdata').glob('method_coverage*.a')):
    path = str(fixture.relative_to(repo))
    baseline = run([str(out / 'baseline'), 'js', path], fixture.stem + '-baseline')
    assert baseline.returncode == 0, (path, baseline.stderr.decode())
    dropped = run([str(out / 'drop'), 'js', path], fixture.stem + '-drop')
    if dropped.returncode:
        assert b'a delayed library result whose contextual representation is not proven' in dropped.stderr, (path, dropped.stderr.decode())
        affected.append(path)
        print(path + ': dropped boundary refuses again', flush=True)
assert {'internal/oracle/testdata/method_coverage_string_trim.a',
        'internal/oracle/testdata/method_coverage_string_normalize.a',
        'internal/oracle/testdata/method_coverage_string_padend.a'} <= set(affected)
(out / 'affected-fixtures.json').write_text(json.dumps(affected, indent=2) + '\n')
print(f'{len(affected)} affected const-alias fixtures; unsafe nullable admission mutant killed', flush=True)

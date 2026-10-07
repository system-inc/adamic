#!/usr/bin/env python3
"""Numeric declarations compared with independently parsed Go production listeners."""
import argparse
import json
import pathlib
import re
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--repository', type=pathlib.Path, required=True)
parser.add_argument('--artifacts', type=pathlib.Path, required=True)
parser.add_argument('--compiler', type=pathlib.Path, required=True)
parser.add_argument('--checker', type=pathlib.Path, required=True)
args = parser.parse_args()
root = args.repository.resolve()
base = root / 'stage1/cohere/typeaware'
out = args.artifacts.resolve()
out.mkdir(parents=True, exist_ok=True)
sequence = 0

def run(label, command, cwd=root):
    global sequence
    sequence += 1
    label = f'{sequence:03d}-{label}'
    with (out / (label + '.stdout')).open('wb') as stdout, (out / (label + '.stderr')).open('wb') as stderr:
        result = subprocess.run([str(part) for part in command], cwd=cwd, stdout=stdout, stderr=stderr, timeout=300)
    if result.returncode != 0:
        raise RuntimeError(f'{label}: exit {result.returncode}; see logs')
    return (out / (label + '.stdout')).read_bytes(), (out / (label + '.stderr')).read_bytes()

virtual = root / 'cohere/wave20_listener_oracle.go'
overlay = out / 'oracle-overlay.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(base / 'react-jsx-fragments/testdata/listener_oracle.go')}}))
oracle = out / 'oracle'
run('oracle-build', ['go', 'build', '-overlay', overlay, '-o', oracle, virtual], cwd=root / 'cohere')
truth, errors = run('oracle', [oracle, root / 'cohere'])
assert not errors
entry = base / 'react-jsx-fragments/listener_probe.a'

def build(label, source, sanitized=False):
    binary = out / label
    command = [args.compiler, 'build', source, '-o', binary, '--tsgo', args.checker]
    if sanitized:
        command.append('--sanitize')
    run(label + '-build', command)
    return binary

for label, sanitized in [('native', False), ('native-asan', True)]:
    binary = build(label, entry, sanitized)
    got, errors = run(label, [binary])
    assert not errors and got == truth, label + ': declaration bytes or sanitizer stderr differ'
    print(label, 'PASS', len(got), 'identical production listener bytes', flush=True)

imports = re.findall(r"import \{ syntaxKinds as (\w+)Kinds \} from '([^']+)';", entry.read_text())
assert len(imports) == 3
mutants = []
for label, path in imports:
    original = (entry.parent / path).resolve()
    source = original.read_text()
    source, count = re.subn(r'(export const syntaxKinds: readonly number\[\] = \[)(\d+)', lambda match: match[1] + str(int(match[2]) + 1), source, count=1)
    assert count == 1
    source = re.sub(r"from '([^']+)'", lambda match: "from '" + str((original.parent / match[1]).resolve()) + "'" if match[1].startswith('.') else match[0], source)
    mutation = out / (label + '-mutant.a')
    mutation.write_text(source)
    driver = entry.read_text()
    driver = re.sub(r"from '([^']+)'", lambda match: "from '" + str(mutation if (entry.parent / match[1]).resolve() == original else (entry.parent / match[1]).resolve()) + "'", driver)
    changed_entry = out / (label + '-entry.a')
    changed_entry.write_text(driver)
    binary = build(label + '-mutant', changed_entry)
    got, errors = run(label + '-mutant', [binary])
    assert not errors and got != truth, label + ': mutation not killed solely by declaration bytes'
    byte = next(at for at, pair in enumerate(zip(got, truth)) if pair[0] != pair[1])
    mutants.append({'rule': label, 'byte': byte, 'exit': 0, 'stderr_bytes': 0})
    print('KILLED', mutants[-1], flush=True)
    binary.unlink()
(out / 'mutants.json').write_text(json.dumps(mutants, indent=2) + '\n')
print('PASS three declarations, normal/sanitizer agreement, three compiling declaration mutants', flush=True)

records = []
rows = [('react/jsx-fragments', 'react-jsx-fragments'), ('react/jsx-no-constructed-context-values', 'react-jsx-no-constructed-context-values'), ('react/jsx-no-undef', 'react-jsx-no-undef')]
for name, directory in rows:
    obj = json.loads((base / directory / 'rule.json').read_text())
    assert set(obj) == {'kinds'} and all(type(kind) is int for kind in obj['kinds'])
    records.append((name, obj['kinds']))
def serialize(records):
    return ''.join(name + ' ' + ','.join(map(str, kinds)) + '\n' for name, kinds in records).encode()
assert serialize(records) == truth
for index, (name, kinds) in enumerate(records):
    mutation = [(label, values[:]) for label, values in records]
    mutation[index][1][0] += 1
    assert serialize(mutation) != truth
    print('KILLED manifest', name, 'first kind plus one', flush=True)
print('PASS three JSON manifests matched production registrations', flush=True)

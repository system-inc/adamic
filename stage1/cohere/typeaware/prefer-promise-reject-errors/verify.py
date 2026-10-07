#!/usr/bin/env python3
"""Isolated rule gate. Never changes the shared harness or registration generator."""
import argparse
import hashlib
import json
import os
import pathlib
import re
import statistics
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('--repository', type=pathlib.Path, required=True)
parser.add_argument('--artifacts', type=pathlib.Path, required=True)
parser.add_argument('--compiler', type=pathlib.Path, required=True)
parser.add_argument('--cases', type=pathlib.Path, required=True)
args = parser.parse_args()
repository = args.repository.resolve()
artifacts = args.artifacts.resolve()
artifacts.mkdir(parents=True, exist_ok=True)
base = repository / 'stage1/cohere/typeaware'
names = ['prefer-promise-reject-errors', 'prefer-regex-literals', 'prefer-rest-params']
sequence = 0

def run(label, command, cwd=repository, environment=None, expected=0):
    global sequence
    sequence += 1
    label = f'{sequence:03d}-{label}'
    started = time.monotonic()
    with (artifacts / (label + '.stdout')).open('wb') as stdout, (artifacts / (label + '.stderr')).open('wb') as stderr:
        result = subprocess.run([str(part) for part in command], cwd=cwd, env=environment, stdout=stdout, stderr=stderr, timeout=900)
    elapsed = time.monotonic() - started
    output = (artifacts / (label + '.stdout')).read_bytes()
    errors = (artifacts / (label + '.stderr')).read_bytes()
    if expected is not None and result.returncode != expected:
        raise RuntimeError(f'{label}: exit {result.returncode}; see artifact logs')
    return output, errors, result.returncode, elapsed

def overlay(label, original, replacement):
    path = artifacts / (label + '.json')
    path.write_text(json.dumps({'Replace': {str(original): str(replacement)}}))
    return path

def archive(label, mutation=None, sanitize=False):
    path = artifacts / (label + '.a')
    command = ['go', 'build', '-buildmode=c-archive', '-o', path]
    if mutation:
        command += ['-overlay', mutation]
    command += ['./bridge/tsgo/archive']
    environment = os.environ.copy()
    if sanitize:
        environment.update(CC='clang', CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
    run(label, command, environment=environment)
    return path

def build(label, entry, checker, sanitize=False):
    path = artifacts / label
    command = [stage0, 'build', entry, '-o', path, '--tsgo', checker]
    if sanitize:
        command += ['--sanitize']
    run(label, command)
    return path

def flags(manifest):
    path = pathlib.Path(manifest).parent / 'flags.json'
    return (json.loads(path.read_text()) or []) if path.exists() else []

def compare(label, native, config, manifest):
    truth = run(label + '-go', [oracle, config, manifest, *flags(manifest)])
    got = run(label + '-native', [native, config, manifest, *flags(manifest)])
    if got[1] or got[0] != truth[0]:
        raise RuntimeError(f'{label}: bytes or sanitizer stderr differ')
    print(label, 'identical bytes', len(got[0]), got[0].splitlines()[-1].decode(), flush=True)
    return truth[0]

stage0 = artifacts / 'adamic'
run('stage0', ['go', 'build', '-o', stage0, './cmd/adamic'])
checker = archive('checker')
entry = base / 'prefer-promise-reject-errors/suite.a'
native = build('suite', entry, checker)
oracle = artifacts / 'oracle'
virtual = repository / 'cohere/wave20_next_oracle.go'
oracle_overlay = overlay('oracle', virtual, base / 'prefer-promise-reject-errors/testdata/oracle.go')
run('oracle-build', ['go', 'build', '-overlay', oracle_overlay, '-o', oracle, virtual], cwd=repository / 'cohere')

sanitized_checker = archive('checker-asan', sanitize=True)
san = build('suite-asan', entry, sanitized_checker, sanitize=True)
cases = sorted(args.cases.glob('case-*'))
assert len(cases) >= 682, 'missing upstream fixture projects'
control_truth = {}
positive = set()
for case in cases:
    truth = compare(case.name, native, case / 'tsconfig.json', case / 'manifest')
    compare(case.name + '-asan', san, case / 'tsconfig.json', case / 'manifest')
    control_truth[case] = truth
    for name in names:
        if ('\t' + name + '\t').encode() in truth:
            positive.add(name)
assert positive == set(names), 'vacuous rule controls'

rule_mutations = [
    ('promise', 'prefer-promise-reject-errors/rule.a', '&& !globalUndefined', '&& globalUndefined'),
    ('regex-suggestion-span', 'prefer-regex-literals/rule.a', 'this.rules.byte(node.end), replacement', 'this.rules.byte(node.end) + 1, replacement'),
    ('rest', 'prefer-rest-params/rule.a', 'symbol.declarations.length !== 0', 'symbol.declarations.length === 0'),
]
question_mutations = [
    ('binding-origin', 'bridge/tsgo/checker/binding_origin.go', 'out.yes(source.IsDeclarationFile)', 'out.yes(!source.IsDeclarationFile)'),
]
mutant_results = []

def kill(label, mutant):
    for case in cases:
        got = run(label + '-' + case.name, [mutant, case / 'tsconfig.json', case / 'manifest', *flags(case / 'manifest')])
        if got[1]:
            raise RuntimeError(label + ': stderr kill is not a byte-oracle kill')
        if got[0] != control_truth[case]:
            truth = control_truth[case]
            byte = next((at for at, (left, right) in enumerate(zip(got[0], truth)) if left != right), min(len(got[0]), len(truth)))
            result = {'mutant': label, 'case': case.name, 'byte': byte, 'exit': got[2], 'stderr_bytes': len(got[1])}
            mutant_results.append(result)
            print('KILLED', json.dumps(result), flush=True)
            return
    raise RuntimeError(label + ': mutant survived')

for label, file, before, after in rule_mutations:
    directory = artifacts / (label + '-source')
    for name in names:
        destination = directory / name
        destination.mkdir(parents=True, exist_ok=True)
        for original in (base / name).glob('*.a'):
            source = original.read_text()
            if str(original.relative_to(base)) == file:
                assert source.count(before) == 1
                source = source.replace(before, after, 1)
            source = source.replace("'../no-process-exit-after-output/", "'" + str(base / 'no-process-exit-after-output') + '/')
            source = source.replace("'../../../typescript/", "'" + str(repository / 'stage1/typescript') + '/')
            source = re.sub(r"'\.\./([a-z_]+\.ts)'", lambda match: "'" + str(base / match.group(1)) + "'", source)
            (destination / original.name).write_text(source)
    mutant = build(label + '-mutant', directory / 'prefer-promise-reject-errors/suite.a', checker)
    kill(label, mutant)
    mutant.unlink()

for label, file, before, after in question_mutations:
    source = (repository / file).read_text()
    assert source.count(before) == 1
    replacement = artifacts / (label + '.go')
    replacement.write_text(source.replace(before, after, 1))
    change = overlay(label, repository / file, replacement)
    changed_checker = archive(label + '-checker', change)
    mutant = build(label + '-mutant', entry, changed_checker)
    kill(label, mutant)
    mutant.unlink()
    changed_checker.unlink()

corpora = []
for name, root, config in [('repository', repository, repository / 'tsconfig.json'), ('compiler', args.compiler.resolve(), args.compiler.resolve() / 'src/compiler/tsconfig.json')]:
    paths = [(root / path).resolve() for path in (base / ('validation-coverage/' + name + '.manifest')).read_text().splitlines() if path]
    manifest = artifacts / (name + '.manifest')
    manifest.write_text(''.join(str(path) + '\n' for path in paths))
    compare(name, native, config, manifest)
    compare(name + '-asan', san, config, manifest)
    corpora.append((name, config, manifest))

probe = artifacts / 'probe.a'
probe.write_text('x();\n')
config = artifacts / 'tsconfig.json'
config.write_text('{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"sourceExtensions":[".a"]}')
(artifacts / 'fixture.d.ts').write_text('export {};\n')
released = build('released', base / 'prefer-promise-reject-errors/released.a', checker)
for question, _, _, _ in question_mutations:
    got = run('released-' + question, [released, config, probe, question], expected=70)
    assert b'invalid or released checker handle' in got[1]
    print('released', question, 'panic 70', flush=True)
original = repository / 'bridge/tsgo/archive/main.go'
source = original.read_text()
before = 'delete(programs.live, uint64(handle))'
assert source.count(before) == 1
replacement = artifacts / 'released-registry.go'
replacement.write_text(source.replace(before, '// Mutant retains the released handle.', 1))
changed_checker = archive('released-registry', overlay('released-registry', original, replacement))
mutant = build('released-mutant', base / 'prefer-promise-reject-errors/released.a', changed_checker)
for question, _, _, _ in question_mutations:
    run('released-mutant-' + question, [mutant, config, probe, question], expected=0)
    print('KILLED released-registry', question, 'required panic catches exit 0', flush=True)
mutant.unlink()
changed_checker.unlink()

# No builds or tests overlap these alternating rounds.
measurements = {}
for name, config, manifest in corpora:
    rounds = {'native': [], 'go': []}
    for index in range(3):
        pair = {}
        for engine in (['native', 'go'] if index % 2 == 0 else ['go', 'native']):
            environment = os.environ.copy()
            environment['ADAMIC_TSGO_TIMING'] = '1'
            got = run(f'bench-{name}-{index + 1}-{engine}', [native if engine == 'native' else oracle, config, manifest, *flags(manifest)], environment=environment)
            pair[engine] = got[0]
            rounds[engine].append({'seconds': got[3], 'bytes': len(got[0]), 'sha256': hashlib.sha256(got[0]).hexdigest(), 'phases': got[1].decode()})
        assert pair['native'] == pair['go']
    measurements[name] = {'rounds': rounds, 'medians': {engine: statistics.median(row['seconds'] for row in rows) for engine, rows in rounds.items()}}
    print('TIMING', name, measurements[name]['medians'], flush=True)
(artifacts / 'bench.json').write_text(json.dumps(measurements, indent=2) + '\n')
(artifacts / 'mutants.json').write_text(json.dumps(mutant_results, indent=2) + '\n')
print('PASS', len(cases), 'projects (662 upstream, 20 additional), normal and ASan/UBSan/LSan; both corpora; every mutant and released query', flush=True)

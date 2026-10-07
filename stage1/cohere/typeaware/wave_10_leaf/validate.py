"""Isolated supplied-fact controls. This is not the live native lint harness."""
import argparse
import json
import subprocess
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('--stage0', required=True)
parser.add_argument('--artifacts', required=True)
args = parser.parse_args()
owned = Path(__file__).resolve().parent
repository = owned.parents[3]
artifacts = Path(args.artifacts).resolve()
artifacts.mkdir(parents=True, exist_ok=True)
sources = artifacts / 'sources'
sources.mkdir(exist_ok=True)
fixtures = json.loads((owned / 'testdata/controls.json').read_text())
for name, text in fixtures.items():
    (artifacts / name).write_text(text)
for label in ['symbol_description', 'react_hook_no_any_type', 'require_await']:
    paths = sorted(str(artifacts / name) for name in fixtures if name.startswith(label + '-'))
    (artifacts / (label + '.manifest')).write_text('\n'.join(paths) + '\n')
(artifacts / 'tsconfig.json').write_text('{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","ESNext.Disposable"]}}\n')
for source in owned.rglob('*.a'):
    target = sources / source.relative_to(owned)
    target.parent.mkdir(parents=True, exist_ok=True)
    text = source.read_text().replace('/workspace/wave-10-leaf-controls', str(artifacts))
    text = text.replace("'../../diagnostic.ts'", "'" + str(repository / 'stage1/cohere/typeaware/diagnostic.ts') + "'")
    target.write_text(text)


def run(label, command, directory=repository):
    result = subprocess.run(command, cwd=directory, capture_output=True)
    (artifacts / (label + '.stdout')).write_bytes(result.stdout)
    (artifacts / (label + '.stderr')).write_bytes(result.stderr)
    assert result.returncode == 0, (label, result.returncode, result.stderr)
    return result


oracles = {}
for label, name in [('symbol', 'symbol_description'), ('hook', 'react_hook_no_any_type'), ('await', 'require_await')]:
    virtual = repository / 'cohere' / ('adamic_wave10_leaf_' + name + '.go')
    overlay = artifacts / (name + '.overlay.json')
    overlay.write_text(json.dumps({'Replace': {str(virtual): str(owned / 'testdata' / ('oracle_' + name + '.go'))}}))
    oracle = artifacts / (label + '-oracle')
    run(label + '-oracle-build', ['go', 'build', '-overlay', str(overlay), '-o', str(oracle), str(virtual)], repository / 'cohere')
    oracles[label] = oracle
catalog = run('raw-catalog', [str(oracles['hook']), str(artifacts / 'tsconfig.json'), str(artifacts / 'react_hook_no_any_type.manifest'), '--catalog'])
(artifacts / 'catalog.txt').write_bytes(catalog.stdout)
truths = {}
for label, name in [('symbol', 'symbol_description'), ('hook', 'react_hook_no_any_type'), ('await', 'require_await')]:
    truth = run(label + '-go', [str(oracles[label]), str(artifacts / 'tsconfig.json'), str(artifacts / (name + '.manifest'))]).stdout
    if label == 'await':
        rows = []
        path = ''
        positive = False
        for line in truth.decode().splitlines():
            if line.startswith('file\t'):
                if path:
                    rows.append(f'case\t{path}\t{int(positive)}')
                path = line[5:]
                positive = False
            elif '\tmissingAwait\t' in line:
                positive = True
        if path:
            rows.append(f'case\t{path}\t{int(positive)}')
        truth = ('\n'.join(rows) + '\n').encode()
    truths[label] = truth
    (artifacts / (label + '-expected.stdout')).write_bytes(truth)
    for sanitize in [False, True]:
        binary = artifacts / (label + ('-asan' if sanitize else '-native'))
        command = [args.stage0, 'build', str(sources / (name + '_controls.a')), '-o', str(binary)]
        if sanitize:
            command.append('--sanitize')
        run(binary.name + '-build', command)
        command = [str(binary)]
        if label == 'hook':
            command.append(str(artifacts / 'catalog.txt'))
        observed = run(binary.name + '-run', command)
        assert observed.stderr == b'' and observed.stdout == truth, (label, 'Go comparison')
    print(label, 'supplied-fact comparison and sanitizers PASS', len(truth), 'bytes', flush=True)

for label, name, before, after in [
    ('symbol', 'symbol_description', 'node.declarationFiles[0] !== true', 'node.declarationFiles[0] === true'),
    ('hook', 'react_hook_no_any_type', '(node.resultFlags & 1) === 0', '(node.resultFlags & 1) !== 0'),
    ('await', 'require_await', '(node.flags & 7) === 6', '(node.flags & 7) !== 6'),
]:
    mutant = artifacts / (label + '-mutant-source')
    for source in sources.rglob('*.a'):
        target = mutant / source.relative_to(sources)
        target.parent.mkdir(parents=True, exist_ok=True)
        text = source.read_text()
        if source == sources / name / 'rule.a':
            assert text.count(before) == 1
            text = text.replace(before, after, 1)
        target.write_text(text)
    binary = artifacts / (label + '-mutant')
    run(label + '-mutant-build', [args.stage0, 'build', str(mutant / (name + '_controls.a')), '-o', str(binary)])
    command = [str(binary)]
    if label == 'hook':
        command.append(str(artifacts / 'catalog.txt'))
    observed = run(label + '-mutant-run', command)
    assert observed.stderr == b'' and observed.stdout != truths[label], (label, 'mutant survived')
    byte = next((i for i, (a, b) in enumerate(zip(observed.stdout, truths[label])) if a != b), min(len(observed.stdout), len(truths[label])))
    print(label, 'mutant exits 0, empty stderr, comparison catches byte', byte, flush=True)
refusal = subprocess.run([str(artifacts / 'await-native'), '--refuse'], capture_output=True)
(artifacts / 'await-refusal.stdout').write_bytes(refusal.stdout)
(artifacts / 'await-refusal.stderr').write_bytes(refusal.stderr)
assert refusal.returncode == 70 and b'NotYet: require-await supplied-node adapter' in refusal.stderr
print('require-await candidate refuses explicitly: exit 70', flush=True)

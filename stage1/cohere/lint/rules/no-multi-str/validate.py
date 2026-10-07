#!/usr/bin/env python3
"""Owned complete-proposal comparison; no shared harness or generator edits."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

owned = Path(__file__).resolve().parent
repository = owned.parents[4]
parser = argparse.ArgumentParser()
parser.add_argument('--scratch', type=Path, required=True)
parser.add_argument('--typescript', type=Path, required=True)
parser.add_argument('--mutants', action='store_true')
parser.add_argument('--throughput', action='store_true')
parser.add_argument('--skip-build', action='store_true')
parser.add_argument('--fixtures-only', action='store_true')
args = parser.parse_args()
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
subjects = [
    ('no-multi-str', 'no-multi-str', 'core', 'TestNoMultiStr', "var x%d = 'a\\\n b';\n"),
    ('no-nonoctal-decimal-escape', 'no-nonoctal-decimal-escape', 'core', 'TestNoNonoctalDecimalEscape', "var x%d = '\\8';\n"),
    ('no-octal', 'no-octal', 'core', 'TestNoOctal', "var x%d = 0755;\n"),
]
runner = repository / 'oracle/node.mjs'
virtual_builder = repository / 'wave15_owned_build.go'
virtual_oracle = repository / 'cohere/wave15_owned.go'
builder_overlay = scratch / 'builder-overlay.json'
builder_overlay.write_text(json.dumps({'Replace': {str(virtual_builder): str(owned / 'build.go.txt')}}))
oracle_overlay = scratch / 'oracle-overlay.json'
oracle_overlay.write_text(json.dumps({'Replace': {str(virtual_oracle): str(owned / 'verify-oracle.go.txt')}}))

def run(command, name, cwd=repository):
    start = time.perf_counter()
    with (scratch / name).open('wb') as log:
        result = subprocess.run([str(x) for x in command], cwd=cwd, stdout=log, stderr=subprocess.PIPE)
    (scratch / (name + '.stderr')).write_bytes(result.stderr)
    if result.returncode or result.stderr:
        raise RuntimeError(f'{name}: exit={result.returncode} stderr={result.stderr[:1000]!r}')
    return (scratch / name).read_bytes(), time.perf_counter() - start

def build(entry, output, name):
    return run(['go', 'run', '-overlay=' + str(builder_overlay), virtual_builder, entry, output] + (['fast'] if name == 'port-build.log' else []), name)

entry = owned / 'verify.a'
if not args.skip_build:
    run(['go', 'build', '-overlay=' + str(oracle_overlay), '-o', scratch / 'oracle', virtual_oracle], 'oracle-build.log', repository / 'cohere')
    build(entry, scratch / 'port', 'port-build.log')

capture = scratch / 'capture'
capture.mkdir(exist_ok=True)
environment = os.environ.copy()
environment['COHERE_DOCS_CAPTURE'] = str(capture)
for package in ['core']:
    filters = '|'.join(x[3] for x in subjects if x[2] == package)
    with (scratch / (package + '-upstream.log')).open('wb') as log:
        subprocess.run(['go', 'test', './internal/lint/rules/' + package, '-count=1', '-v', '-timeout=10m', '-run', '^(' + filters + ')'], cwd=repository / 'cohere', env=environment, stdout=log, stderr=subprocess.STDOUT, check=True)

selected = {x[1] for x in subjects}
records = {}
for path in capture.glob('*.jsonl'):
    for line in path.read_text().splitlines():
        record = {key[:1].upper() + key[1:]: value for key, value in json.loads(line).items()}
        if record['Rule'] in selected:
            records[(record['Rule'], record['File'], record['Source'])] = record
excluded = [record for record in records.values() if (record['File'].endswith('.tsx') and '<' in record['Source']) or record['Source'] in ['var a = 01.5;', 'var a = 0777.5;', 'var a = 0755n;']]
(scratch / 'parser-excluded.json').write_text(json.dumps(excluded, ensure_ascii=False, indent=2))
records = {key: value for key, value in records.items() if value not in excluded}
print('JSX/recovered-syntax exclusions: ' + str(len(excluded)), flush=True)
rows = []
counts = {name: 0 for name in selected}
for position, (key, record) in enumerate(sorted(records.items())):
    path = scratch / ('case-' + str(position) + (Path(record['File']).suffix or '.ts'))
    path.write_text(record['Source'])
    rows.append(str(path) + '\t' + record['Rule'])
    counts[record['Rule']] += 1
print('Captured original asserted tests: ' + json.dumps(counts, sort_keys=True), flush=True)
# Separate filename-sensitive vectors and scope/Unicode/numeric probes omitted
# by upstream's docs capture are included explicitly, never by altering its hook.
extra = [('no-nonoctal-decimal-escape', "// 💎\nconst text = '\\0\\8\\9';", '.ts')]
for position, (name, source, suffix) in enumerate(extra):
    path = scratch / ('extra-' + str(position) + suffix)
    path.write_text(source)
    rows.append(str(path) + '\t' + name)
for slug, name, *_ in subjects:
    for witness in sorted((repository / 'stage1/cohere/lint/rules' / slug / 'testdata').glob('*.ts.txt')):
        rows.append(str(witness) + '\t' + name)
manifest = scratch / 'fixtures.manifest'
manifest.write_text('\n'.join(rows) + '\n')

def compare(manifest, label, candidate=entry, binary=None):
    binary = binary or scratch / 'port'
    flags = ['--proposals-only'] if label == 'corpus' else []
    expected, _ = run([scratch / 'oracle', manifest] + flags, label + '-Go.log')
    for side, command in [
        ('Node', ['node', '--disable-warning=ExperimentalWarning', runner, candidate, manifest]),
        ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(binary) + '.mjs', manifest]),
        ('native', [binary, manifest]),
    ]:
        actual, _ = run(command + flags, label + '-' + side + '.log')
        if actual != expected:
            import difflib
            difference = ''.join(difflib.unified_diff(expected.decode().splitlines(True), actual.decode().splitlines(True)))
            (scratch / (label + '-' + side + '-diff.log')).write_text(difference)
            raise RuntimeError(label + ': ' + side + ' differs, see diff log')
    print(f'{label}: Go/Node/emitted JavaScript/sanitized native identical, {len(expected)} bytes', flush=True)

# Reproduce every excluded input on the unchanged Go oracle and all backends.
# No rule can manufacture correct JSX parents or recovered numeric nodes.
blocked = []
for position, record in enumerate(excluded):
    path = scratch / ('blocked-' + str(position) + Path(record['File']).suffix)
    path.write_text(record['Source'])
    manifest_path = scratch / ('blocked-' + str(position) + '.manifest')
    manifest_path.write_text(str(path) + '\t' + record['Rule'] + '\n')
    run([scratch / 'oracle', manifest_path], f'blocked-{position}-Go.log')
    for side, command in [('Node', ['node', '--disable-warning=ExperimentalWarning', runner, entry, manifest_path]), ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(scratch / 'port') + '.mjs', manifest_path]), ('native', [scratch / 'port', manifest_path])]:
        with (scratch / f'blocked-{position}-{side}.log').open('wb') as output:
            result = subprocess.run(list(map(str, command)), cwd=repository, stdout=output, stderr=subprocess.PIPE)
        (scratch / f'blocked-{position}-{side}.stderr').write_bytes(result.stderr)
        if result.returncode != 70 or not (b'parser' in result.stderr or b'NotYet: no-multi-str requires JSX parent nodes' in result.stderr):
            raise RuntimeError(f'excluded input escaped explicit refusal: {position} {side}')
        blocked.append(dict(case=position, side=side, exit=result.returncode, stderr=result.stderr.decode()))
(scratch / 'blocked-summary.json').write_text(json.dumps(blocked, indent=2))
print('Nine excluded inputs explicitly refused on all three backends; Go succeeds.', flush=True)
# Exercise the current branch bridge without truncating suggestions or ranges.
bridge = []
for slug, name, *_ in subjects:
    witness = scratch / (slug + '-bridge.manifest')
    source = repository / 'stage1/cohere/lint/rules' / slug / 'testdata/raw.ts.txt'
    witness.write_text(str(source) + '\t' + name + '\n')
    for side, command in [('Node', ['node', '--disable-warning=ExperimentalWarning', runner, entry, witness, '--bridge']), ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(scratch / 'port') + '.mjs', witness, '--bridge']), ('native', [scratch / 'port', witness, '--bridge'])]:
        with (scratch / f'bridge-{slug}-{side}.log').open('wb') as output:
            result = subprocess.run(list(map(str, command)), cwd=repository, stdout=output, stderr=subprocess.PIPE)
        (scratch / f'bridge-{slug}-{side}.stderr').write_bytes(result.stderr)
        expected_exit = 70 if name == 'no-nonoctal-decimal-escape' else 0
        if result.returncode != expected_exit or (expected_exit == 70 and b'NotYet: shared lint repair serialization' not in result.stderr) or (expected_exit == 0 and result.stderr):
            raise RuntimeError('unexpected bridge result: ' + name + ' ' + side)
        bridge.append(dict(rule=name, side=side, exit=result.returncode, stderr=result.stderr.decode()))
(scratch / 'bridge-summary.json').write_text(json.dumps(bridge, indent=2))
compare(manifest, 'fixtures')
if args.fixtures_only:
    raise SystemExit(0)
corpus_rows = []
files = sorted(args.typescript.resolve().joinpath('src/compiler').rglob('*.ts'))
files += sorted(path for path in (repository / 'stage1').rglob('*') if path.suffix in ['.a', '.ts'])
for path in files:
    for name in sorted(selected):
        corpus_rows.append(str(path) + '\t' + name)
corpus = scratch / 'corpus.manifest'
corpus.write_text('\n'.join(corpus_rows) + '\n')
print(f'Corpus: {len(files)} files, {len(corpus_rows)} file/rule pairs', flush=True)
compare(corpus, 'corpus')

if args.mutants:
    for slug, name, *_ in subjects:
        mutant = json.loads((repository / 'stage1/cohere/lint/rules' / slug / 'mutant.json').read_text())
        tree = scratch / slug
        for source in (repository / 'stage1').rglob('*'):
            if source.suffix not in ['.ts', '.a']:
                continue
            target = tree / source.relative_to(repository)
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, target)
        target = tree / 'stage1/cohere/lint/rules' / slug / mutant['file']
        source = target.read_text()
        if source.count(mutant['from']) != 1:
            raise RuntimeError('nonunique mutant anchor: ' + slug)
        target.write_text(source.replace(mutant['from'], mutant['to']))
        candidate = tree / entry.relative_to(repository)
        binary = scratch / (slug + '-mutant')
        build(candidate, binary, slug + '-mutant-build.log')
        witness = scratch / (slug + '-mutant.manifest')
        witness.write_text('\n'.join(row for row in rows if row.split('\t')[1] == name) + '\n')
        expected, _ = run([scratch / 'oracle', witness], slug + '-mutant-Go.log')
        for side, command in [('Node', ['node', '--disable-warning=ExperimentalWarning', runner, candidate, witness]), ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(binary) + '.mjs', witness]), ('native', [binary, witness])]:
            actual, _ = run(command, slug + '-mutant-' + side + '.log')
            if actual == expected:
                raise RuntimeError('mutant survived: ' + slug + ' ' + side)
            print(f'{slug} mutant caught on {side}: successful compile and exit, differing output only', flush=True)

if args.throughput:
    for slug, name, package, test, pattern in subjects:
        path = scratch / (slug + '-stress.ts')
        path.write_text(''.join(pattern % ((i, i) if pattern.count('%d') == 2 else i) for i in range(1000)))
        rows = [str(file) + '\t' + name for file in sorted(args.typescript.resolve().joinpath('src/compiler').rglob('*.ts'))] + [str(path) + '\t' + name]
        manifest = scratch / (slug + '-throughput.manifest')
        manifest.write_text('\n'.join(rows) + '\n')
        expected = None
        best = {}
        for round in range(3):
            for side, command in [('Go', [scratch / 'oracle', manifest, '--count']), ('native', [str(scratch / 'port') + '-fast', manifest, '--count']), ('Node', ['node', '--disable-warning=ExperimentalWarning', runner, entry, manifest, '--count'])]:
                output, seconds = run(command, slug + f'-throughput-{round}-{side}.log')
                if expected is None: expected = output
                if output != expected: raise RuntimeError('throughput count mismatch')
                best[side] = min(best.get(side, float('inf')), seconds)
        findings = int(expected)
        print(f'{name}: {len(rows)} files, {findings} findings, best of 3 ' + ', '.join(f'{side} {findings/seconds:.2f} findings/s ({seconds:.6f}s)' for side, seconds in best.items()), flush=True)

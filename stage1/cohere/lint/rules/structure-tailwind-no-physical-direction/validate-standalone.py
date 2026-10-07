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
    ('next-no-assign-module-variable', '@next/next/no-assign-module-variable', 'next', 'TestNoAssignModuleVariable', '{ let module = %d; }\n'),
    ('typescript-default-param-last', '@typescript-eslint/default-param-last', 'typescript', 'TestDefaultParamLast', 'function f%d(a = 1, b: number) {}\n'),
    ('structure-tailwind-no-physical-direction', 'structure/tailwind-no-physical-direction', 'tailwind', 'TestNoPhysicalDirection', "const c%d = 'ml-4 mr-4 rtl:ml-4';\n"),
]
runner = repository / 'oracle/node.mjs'
virtual_builder = repository / 'wave15_owned_build.go'
virtual_oracle = repository / 'cohere/wave15_owned.go'
builder_overlay = scratch / 'builder-overlay.json'
builder_overlay.write_text(json.dumps({'Replace': {str(virtual_builder): str(owned / 'standalone-build.go.txt')}}))
oracle_overlay = scratch / 'oracle-overlay.json'
oracle_overlay.write_text(json.dumps({'Replace': {str(virtual_oracle): str(owned / 'standalone-oracle.go.txt')}}))

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

entry = owned / 'standalone.a'
if not args.skip_build:
    run(['go', 'build', '-overlay=' + str(oracle_overlay), '-o', scratch / 'oracle', virtual_oracle], 'oracle-build.log', repository / 'cohere')
    build(entry, scratch / 'port', 'port-build.log')

capture = scratch / 'capture'
capture.mkdir(exist_ok=True)
environment = os.environ.copy()
environment['COHERE_DOCS_CAPTURE'] = str(capture)
for package in ['next', 'typescript', 'tailwind']:
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
excluded = [record for record in records.values() if record['File'].endswith('.tsx') and '<' in record['Source']]
(scratch / 'parser-excluded.json').write_text(json.dumps(excluded, ensure_ascii=False, indent=2))
if len(excluded) != 1:
    raise RuntimeError('expected the one pinned JSX parser gap, got ' + str(len(excluded)))
records = {key: record for key, record in records.items() if record not in excluded}
print('One upstream JSX case excluded and independently probed below.', flush=True)
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
extra = [
    ('@next/next/no-assign-module-variable', "// 💎\nlet module, moduleName; for (let module of values) {} const {module} = object; let other, module = {};", '.ts'),
    ('@typescript-eslint/default-param-last', 'function f(a: Map<string, number> = new Map(), b: number) {} class X { constructor(public a /* = comment */ : number, private b?: number, c: number) {} } declare function ambient(a = 1, b: number): void; const f2 = (a = 1, b: number) => b;', '.ts'),
    ('structure/tailwind-no-physical-direction', "const classes = 'ml-4\\u0085mr-4 md:hover:pl-2 rtl:mr-3 -border-l-2 flex ml-4;'; const template = `flex ml-2 ${other} mr-2 ${third} text-left`; const bare = 'ml-4;'; const dollars = 'flex ml-$& mr-$` pl-$$ pr-$1'; const placeholders = 'flex ml-{{replacement}}';", '.ts'),
    ('structure/tailwind-no-physical-direction', "const classes = 'ml-4';", '.tsx'),
    ('structure/tailwind-no-physical-direction', "const classes = 'ml-4';", '.js'),
    ('structure/tailwind-no-physical-direction', "const classes = 'ml-4';", '.a'),
    ('structure/tailwind-no-physical-direction', "const classes = 'ml-4';", '.mts'),
    ('structure/tailwind-no-physical-direction', "const classes = 'ml-4';", '.cts'),
]
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

for position, record in enumerate(excluded):
    path = scratch / ('blocked-' + str(position) + Path(record['File']).suffix)
    path.write_text(record['Source'])
    blocked = scratch / ('blocked-' + str(position) + '.manifest')
    blocked.write_text(str(path) + '\t' + record['Rule'] + '\n')
    expected, _ = run([scratch / 'oracle', blocked], 'blocked-Go.log')
    if b'useLogicalClass' not in expected:
        raise RuntimeError('Go JSX probe lost its finding')
    for side, command in [('Node', ['node', '--disable-warning=ExperimentalWarning', runner, entry, blocked]), ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(scratch / 'port') + '.mjs', blocked]), ('native', [scratch / 'port', blocked])]:
        with (scratch / ('blocked-' + side + '.log')).open('wb') as log:
            result = subprocess.run(list(map(str, command)), cwd=repository, stdout=log, stderr=subprocess.PIPE)
        (scratch / ('blocked-' + side + '.stderr')).write_bytes(result.stderr)
        if result.returncode != 70 or b'parser slice' not in result.stderr:
            raise RuntimeError('JSX probe escaped explicit parser refusal on ' + side)
print('Go JSX finding reproduced; source Node, emitted JavaScript and sanitized native explicitly refuse it.', flush=True)
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

if args.mutants:
    slug = 'structure-tailwind-no-physical-direction'
    tree = scratch / slug
    target = tree / 'stage1/cohere/lint/rules' / slug / 'rule.a'
    shutil.copyfile(repository / 'stage1/cohere/lint/rules' / slug / 'rule.a', target)
    message = tree / 'stage1/cohere/lint/rules' / slug / 'messages.a'
    original = '.split("{{original}}").join(original).split("{{replacement}}").join(replacement)'
    changed = '.replaceAll("{{original}}", original).replaceAll("{{replacement}}", replacement)'
    text = message.read_text()
    if text.count(original) != 1: raise RuntimeError('nonunique interpolation mutant')
    message.write_text(text.replace(original, changed))
    candidate = tree / entry.relative_to(repository)
    binary = scratch / 'interpolation-mutant'
    build(candidate, binary, 'interpolation-mutant-build.log')
    witness = scratch / 'interpolation-mutant.manifest'
    witness.write_text('\n'.join(row for row in rows if row.split('\t')[1] == 'structure/tailwind-no-physical-direction') + '\n')
    expected, _ = run([scratch / 'oracle', witness], 'interpolation-mutant-Go.log')
    for side, command in [('Node', ['node', '--disable-warning=ExperimentalWarning', runner, candidate, witness]), ('JavaScript', ['node', '--disable-warning=ExperimentalWarning', runner, str(binary) + '.mjs', witness]), ('native', [binary, witness])]:
        actual, _ = run(command, 'interpolation-mutant-' + side + '.log')
        if actual == expected: raise RuntimeError('interpolation mutant survived on ' + side)
        print('Literal interpolation mutant caught on ' + side + ': successful compile and exit, differing output only', flush=True)

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

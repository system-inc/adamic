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
    ('typescript-no-unnecessary-type-constraint', '@typescript-eslint/no-unnecessary-type-constraint', 'typescript', 'TestNoUnnecessaryTypeConstraint', 'function f%d<T extends unknown>() {}\n'),
    ('typescript-prefer-as-const', '@typescript-eslint/prefer-as-const', 'typescript', 'TestPreferAsConst', "let value%d: 'bar' = 'bar';\n"),
    ('typescript-prefer-enum-initializers', '@typescript-eslint/prefer-enum-initializers', 'typescript', 'TestPreferEnumInitializers', 'enum D%d { Up }\n'),
    ('nexus-import-require-node-namespace', 'nexus/import-require-node-namespace', 'nexus', 'TestImportRequireNodeNamespace|TestTheNodePrefixFix|TestExpectedAlias', "import * as fs%d from 'fs';\n"),
    ('structure-network-no-invalidate-cache-literal-key', 'structure/network-no-invalidate-cache-literal-key', 'structure', 'TestNetworkNoInvalidateCacheLiteralKey', "client.cache.invalidate('users%d');\n"),
    ('structure-network-no-string-literal-query', 'structure/network-no-string-literal-query', 'structure', 'TestNetworkNoStringLiteralQuery', "const Document%d = 'query { id }'; client.useGraphQlQuery(Document%d);\n"),
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
for package in ['typescript', 'nexus', 'structure']:
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
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends any /* , */>() => {};', '.tsx'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends any /* comment */,>() => {};', '.tsx'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends any /* comment */>() => {};', '.tsx'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <Element extends unknown>() => {};', '.tsx'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends any>() => {};', '.ts'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends any>() => {};', '.mts'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends any>() => {};', '.cts'),
    ('nexus/import-require-node-namespace', "import NodePath from 'node:path';\n", '.ts'),
    ('nexus/import-require-node-namespace', "import * as NodeFileSystem from 'node:fs/promises';\n", '.ts'),
    ('nexus/import-require-node-namespace', "import * as NodeChildProcess from 'child_process';\n", '.ts'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends unknown>() => {};', '.mts'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends unknown>() => {};', '.cts'),
    ('@typescript-eslint/no-unnecessary-type-constraint', 'const data = <T extends unknown>() => {};', '.ts'),
    ('@typescript-eslint/no-unnecessary-type-constraint', '// 💎\nfunction data<T, U extends unknown>() {}', '.ts'),
    ('@typescript-eslint/prefer-as-const', "let value: 0x10 = 16; let other: 'a\\u0062' = 'ab'; class C { static x: 2 = 2; }", '.ts'),
    ('@typescript-eslint/prefer-enum-initializers', "enum D { A = 10, B, 'quoted' }", '.ts'),
    ('nexus/import-require-node-namespace', "import type * as fs from 'fs'; import * as System from 'node:os'; import 'node:path';", '.ts'),
    ('structure/network-no-invalidate-cache-literal-key', "(client.cache).invalidate<string>((['users', (`posts-${id}`), , ...others]));", '.ts'),
    ('structure/network-no-string-literal-query', "const first = second; const second = first; client.graphQlRequest(first); { const Document = 'query'; client.graphQlRequest((Document as string)); }", '.ts'),
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

"""Verify owned mutations on all runtimes without changing the shared harness."""
import argparse
import json
from pathlib import Path
import re
import shutil
import subprocess
import tempfile

SLUGS = [
    'nexus-import-require-module-alias',
    'typescript-eslint-no-confusing-non-null-assertion',
    'typescript-eslint-no-duplicate-enum-values',
    'typescript-eslint-no-dynamic-delete',
    'typescript-eslint-no-extra-non-null-assertion',
    'typescript-eslint-no-misused-new',
    'typescript-eslint-no-unnecessary-parameter-property-assignment',
    'no-lone-blocks', 'no-lonely-if', 'no-loss-of-precision',
]
IMPORT = re.compile(r'''(?m)(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])''')


def run(args, log, cwd=None):
    with Path(str(log) + '.stdout').open('wb') as out, Path(str(log) + '.stderr').open('wb') as err:
        result = subprocess.run([str(a) for a in args], cwd=cwd, stdout=out, stderr=err, timeout=600)
    output = Path(str(log) + '.stdout').read_bytes()
    diagnostic = Path(str(log) + '.stderr').read_bytes()
    if result.returncode != 0 or diagnostic:
        raise RuntimeError(f'{args[0]} failed: exit={result.returncode}, stderr bytes={len(diagnostic)}; logs: {log}')
    return output


def rewrite(file, source, lint):
    def replace(match):
        specifier = match[3]
        if not specifier.startswith('.'):
            return match[0]
        absolute = (file.parent / specifier).resolve()
        if absolute.is_relative_to(lint):
            return match[0]
        return match[1] + match[2] + absolute.as_posix() + match[4]
    return IMPORT.sub(replace, source)


def check():
    parser = argparse.ArgumentParser()
    parser.add_argument('--repo', type=Path, required=True)
    parser.add_argument('--records', type=Path, required=True)
    parser.add_argument('--oracle', type=Path, required=True)
    parser.add_argument('--builder', type=Path, required=True)
    parser.add_argument('--archive', type=Path, required=True)
    parser.add_argument('--logs', type=Path, required=True)
    args = parser.parse_args()
    args.logs.mkdir(parents=True, exist_ok=True)
    lint = (args.repo / 'stage1/cohere/lint').resolve()
    records = json.loads(args.records.read_text())
    with tempfile.TemporaryDirectory(prefix='wave14-native-', dir='/tmp/adamic-gate') as temporary:
        root = Path(temporary)
        copy = root / 'lint'
        for file in lint.rglob('*'):
            if not file.is_file() or '.generated' in file.relative_to(lint).parts:
                continue
            module = file.suffix in ('.a', '.ts')
            witness = any(file.name.endswith(ext + '.txt') for ext in ('.ts', '.tsx', '.js', '.jsx'))
            if not (module or witness or file.name in ('rule.json', 'mutant.json', 'oracle.go') or file.name.endswith('.options.json')):
                continue
            target = copy / file.relative_to(lint)
            target.parent.mkdir(parents=True, exist_ok=True)
            if module:
                target.write_text(rewrite(file, file.read_text(), lint))
            else:
                shutil.copyfile(file, target)
        manifests = []
        for slug in SLUGS:
            descriptor = json.loads((lint / 'rules' / slug / 'rule.json').read_text())
            name = descriptor['name']
            cases = [r for r in records if r['rule'] == name]
            for witness in sorted((lint / 'rules' / slug / 'testdata').rglob('*.txt')):
                if not any(witness.name.endswith(ext + '.txt') for ext in ('.ts', '.tsx', '.js', '.jsx')):
                    continue
                sidecar = witness.with_name(witness.name.rsplit('.', 2)[0] + '.options.json')
                cases.append({'source': witness.read_text(), 'file': witness.name[:-4],
                              'options': json.loads(sidecar.read_text()) if sidecar.exists() else None})
            rows = []
            for index, case in enumerate(cases):
                relative = Path(case['file'].replace('\\', '/').lstrip('/'))
                if not relative.parts or '..' in relative.parts:
                    relative = Path(relative.name or 'source.ts')
                path = root / 'cases' / slug / str(index) / relative
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text(case['source'])
                options = '' if case.get('options') is None else json.dumps(case['options'], separators=(',', ':'), ensure_ascii=False)
                rows.append(f'{path}\t{name}\t\t\tfalse\t{options}\trecovery')
            for corpus in (args.repo / 'stage1', Path('/tmp/adamic-typescript-6.0.3/src/compiler')):
                for path in sorted(corpus.rglob('*')):
                    if path.is_file() and path.suffix in ('.a', '.ts'):
                        rows.append(f'{path.resolve()}\t{name}\t\t\tfalse\t\trecovery')
            manifest = root / (slug + '.txt')
            manifest.write_text('\n'.join(rows) + '\n')
            want = run([args.oracle, '--manifest', manifest], args.logs / (slug + '-go'))
            manifests.append((slug, manifest, want, len(cases), len(rows)))
            rule = copy / 'rules' / slug
            mutation = json.loads((rule / 'mutant.json').read_text())
            file = rule / mutation.get('file', 'rule.a')
            original = file.read_text()
            if original.count(mutation['from']) != 1:
                raise RuntimeError(f'{slug}: mutant anchor is not unique')
            file.write_text(original.replace(mutation['from'], mutation['to'], 1))
            binary = root / 'mutant'
            run([args.builder, copy, binary, args.archive], args.logs / (slug + '-build'))
            runtimes = [
                ('Node', ['node', '--disable-warning=ExperimentalWarning', args.repo / 'oracle/node.mjs', copy / 'main.ts']),
                ('emitted-JavaScript', ['node', '--disable-warning=ExperimentalWarning', args.repo / 'oracle/node.mjs', str(binary) + '.js']),
                ('sanitized-native', [binary]),
            ]
            for runtime, command in runtimes:
                got = run(command + ['--manifest', manifest], args.logs / (slug + '-' + runtime))
                if got == want:
                    raise RuntimeError(f'{slug}: mutant survived on {runtime}')
                print(f'{slug}: {mutation["name"]} caught only by Go byte comparison on {runtime}; exit zero, stderr empty; {len(cases)} upstream/witness inputs and {len(rows)} total inputs', flush=True)
            file.write_text(original)
        binary = root / 'restored'
        run([args.builder, copy, binary, args.archive], args.logs / 'restored-build')
        for slug, manifest, want, cases, count in manifests:
            for runtime, command in [
                ('Node', ['node', '--disable-warning=ExperimentalWarning', args.repo / 'oracle/node.mjs', copy / 'main.ts']),
                ('emitted-JavaScript', ['node', '--disable-warning=ExperimentalWarning', args.repo / 'oracle/node.mjs', str(binary) + '.js']),
                ('sanitized-native', [binary]),
            ]:
                got = run(command + ['--manifest', manifest], args.logs / (slug + '-restored-' + runtime))
                if got != want:
                    raise RuntimeError(f'{slug}: restored {runtime} differs from Go; see logs')
            print(f'{slug}: restored source matches Go on all three runtimes; {count} inputs', flush=True)


if __name__ == '__main__':
    check()

#!/usr/bin/env python3
"""Import a deterministic, restricted compiler-test corpus, without running tsc."""
import collections
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys

PIN = '050880ce59e30b356b686bd3144efe24f875ebc8'
ROOT = Path(__file__).resolve().parent
HEADER = re.compile(r'^//\s*@(\w+)\s*:\s*([^\r\n]*)')
OPTIONS = '''target module moduleResolution strict strictNullChecks strictFunctionTypes strictBindCallApply strictPropertyInitialization noImplicitAny noImplicitThis alwaysStrict noUnusedLocals noUnusedParameters noImplicitReturns noFallthroughCasesInSwitch noUncheckedIndexedAccess exactOptionalPropertyTypes useUnknownInCatchVariables noPropertyAccessFromIndexSignature allowUnreachableCode allowUnusedLabels experimentalDecorators emitDecoratorMetadata useDefineForClassFields downlevelIteration jsx jsxFactory jsxFragmentFactory noLib lib skipLibCheck skipDefaultLibCheck noErrorTruncation allowSyntheticDefaultImports esModuleInterop isolatedModules verbatimModuleSyntax moduleDetection preserveConstEnums resolveJsonModule allowJs checkJs allowImportingTsExtensions noResolve ignoreDeprecations'''.split()
CANONICAL = {s.lower(): s for s in OPTIONS}
METADATA = {'noemit', 'noemitonerror', 'declaration', 'sourcemap', 'inlinesourcemap', 'inlinesources', 'removecomments', 'noemithelpers', 'importhelpers'}
FEATURES = [
    ('jsx', r'jsx|tsx'), ('decorators', r'decorator'),
    ('modules', r'import|export|module'), ('namespaces', r'namespace'),
    ('enums', r'enum'), ('classes', r'class|constructor|private|protected|super'),
    ('generics', r'generic|typeparameter|constraint|infer'),
    ('unions', r'union|intersection|narrow|discriminat'),
    ('functions', r'function|arrow|overload|parameter|return'),
    ('arrays', r'array|tuple|spread|rest'), ('objects', r'object|interface|property|indexsignature'),
    ('types', r'typealias|mapped|conditional|keyof|typeof'),
    ('async', r'async|await|promise|generator|yield'),
    ('operators', r'operator|assignment|binary|unary|logical'),
    ('control', r'control|switch|loop|while|for|catch|finally'),
    ('literals', r'literal|template|string|number|boolean'),
    ('syntax', r'parse|syntax|scanner|unicode'),
]

def materialize(raw):
    text = raw.decode('utf-8-sig')
    options = {}
    content = ''
    for line in re.split(r'\r\n|\n|\r', text):
        match = HEADER.match(line)
        if match:
            name, value = match[1].lower(), match[2].strip()
            if name == 'declaration' and value != 'false':
                raise ValueError('declaration baseline requires emission')
            if name in METADATA:
                if value not in ('true', 'false'):
                    raise ValueError('variant metadata')
                continue
            if name not in CANONICAL or (',' in value and name != 'lib'):
                raise ValueError('unsupported header or variant')
            key = CANONICAL[name]
            options[key] = [v.strip() for v in value.split(',')] if name == 'lib' else {'true': True, 'false': False}.get(value, value)
        else:
            if content:
                content += '\n'
            content += line
    return content, options

def baseline_output(raw):
    text = raw.decode('utf-8-sig').replace('\r\n', '\n')
    summary = text.split('\n\n', 1)[0]
    if not re.match(r'^\S+\(\d+,\d+\): error TS\d+:', summary):
        raise ValueError('non-file diagnostic baseline')
    return (summary + '\n').encode()

def choose(tree, stock_api):
    candidates = []
    reference = tree / 'tests/baselines/reference'
    variants = {p.name.split('(', 1)[0] for p in reference.glob('*(*).errors.txt')}
    cases = tree / 'tests/cases/compiler'
    for source in sorted(cases.glob('*.ts')):
        raw = source.read_bytes()
        if len(raw) > 8192 or len(raw.splitlines()) > 80:
            continue
        try:
            content, options = materialize(raw)
        except ValueError:
            continue
        if re.search(r'///\s*<reference|\b(?:import|export)\b[^\n]*[\'"]\.', content):
            continue
        baseline = reference / (source.stem + '.errors.txt')
        if source.stem in variants:
            continue
        if baseline.exists():
            try:
                expected = baseline_output(baseline.read_bytes())
            except ValueError:
                continue
            if re.search(rb'^[^\n]*\(\d+,\d+\):', expected, re.M) and any(
                m[1].decode() != source.name for m in re.finditer(rb'^([^\n]+?)\(\d+,\d+\):', expected, re.M)
            ):
                continue
        else:
            expected = b''
        if any(int(c) < 2000 or int(c) == 18027 for c in re.findall(rb'error TS(\d+):', expected)):
            continue
        feature = next((name for name, pattern in FEATURES if re.search(pattern, source.stem, re.I)), 'other')
        candidates.append({'source': source, 'baseline': baseline if baseline.exists() else None,
                           'options': options, 'feature': feature,
                           'codes': sorted(set(map(int, re.findall(rb'error TS(\d+):', expected)))),
                           'rank': hashlib.sha256(source.name.encode()).hexdigest(), 'expected': expected})
    inputs = [{'name': c['source'].name, 'content': materialize(c['source'].read_bytes())[0]} for c in candidates]
    rejected = set(json.loads(subprocess.check_output(
        ['node', str(ROOT / 'syntax.cjs'), str(Path(stock_api).resolve())],
        input=json.dumps(inputs), text=True)))
    candidates = [c for c in candidates if c['source'].name not in rejected]
    chosen = []
    seen = set()
    for clean, quota in [(True, 60), (False, 240)]:
        pool = [c for c in candidates if (not c['codes']) == clean]
        counts = collections.Counter()
        for _ in range(quota):
            if not pool:
                raise RuntimeError('not enough eligible cases')
            best = min(pool, key=lambda c: (-len(set(c['codes']) - seen), counts[c['feature']], c['rank'], c['source'].name))
            pool.remove(best)
            chosen.append(best)
            counts[best['feature']] += 1
            seen.update(best['codes'])
    return chosen, len(candidates)

def main():
    if len(sys.argv) != 3:
        sys.exit('usage: corpus.py <pinned-tree> <stock-typescript-lib>')
    tree = Path(sys.argv[1]).resolve()
    actual = subprocess.check_output(['git', '-C', str(tree), 'rev-parse', 'HEAD'], text=True).strip()
    if actual != PIN:
        raise RuntimeError('upstream pin mismatch')
    output = ROOT / 'corpus'
    output.mkdir()  # Never replace a reviewed corpus.
    selected, eligible = choose(tree, sys.argv[2])
    manifest = []
    for index, case in enumerate(selected, 1):
        source = case['source']
        name = f'{index:03d}_{source.stem}'
        folder = output / name
        folder.mkdir()
        raw = source.read_bytes()
        (folder / source.name).write_bytes(raw)
        if case['baseline']:
            shutil.copyfile(case['baseline'], folder / 'reference.errors.txt')
        (folder / 'expected.stdout').write_bytes(case['expected'])
        (folder / 'expected.stderr').write_bytes(b'')
        (folder / 'expected.exit').write_text('2\n' if case['codes'] else '0\n')
        manifest.append({'id': name, 'source': 'tests/cases/compiler/' + source.name,
                         'path': str(Path('corpus') / name / source.name),
                         'source_sha256': hashlib.sha256(raw).hexdigest(),
                         'baseline': 'tests/baselines/reference/' + case['baseline'].name if case['baseline'] else None,
                         'baseline_sha256': hashlib.sha256(case['baseline'].read_bytes()).hexdigest() if case['baseline'] else None,
                         'feature': case['feature'], 'codes': case['codes'], 'options': case['options']})
    (ROOT / 'selection.json').write_text(json.dumps({'upstream_commit': PIN, 'eligible': eligible, 'cases': manifest}, indent=2) + '\n')
    shutil.copyfile(tree / 'LICENSE.txt', ROOT / 'LICENSE-TypeScript.txt')
    print(f'selected {len(selected)} from {eligible}; clean=60; codes={len(set(c for row in manifest for c in row["codes"]))}')

if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""Select baseline cases from inputs, never from the tested CLI's results."""
import collections
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parent
DRIVER = ROOT.parent / 'drivers/tsc'
sys.path.insert(0, str(DRIVER))
from corpus import PIN, HEADER, CANONICAL, METADATA, materialize, baseline_output


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def select(tree, api):
    if subprocess.check_output(['git', '-C', str(tree), 'rev-parse', 'HEAD'], text=True).strip() != PIN:
        raise ValueError('upstream pin mismatch')
    reference = tree / 'tests/baselines/reference'
    variants = {p.name.split('(', 1)[0].lower() for p in reference.glob('*(*).errors.txt')}
    baselines = {p.name.lower(): p for p in reference.glob('*.errors.txt')}
    sources = sorted(p for suite in ('compiler', 'conformance')
                     for p in (tree / 'tests/cases' / suite).rglob('*')
                     if p.suffix in ('.ts', '.tsx'))
    selected, excluded = [], []
    for source in sources:
        raw = source.read_bytes()
        path = source.relative_to(tree).as_posix()
        reason = None
        try:
            text = raw.decode('utf-8-sig')
        except UnicodeDecodeError:
            excluded.append({'source': path, 'source_sha256': digest(raw),
                             'reason': 'non-UTF8 source encoding outside materializer scope'})
            continue
        headers = [(m[1].lower(), m[2].strip()) for m in
                   (HEADER.match(line) for line in text.splitlines()) if m]
        for name, value in headers:
            if name == 'filename':
                reason = 'filename: virtual units and root-file selection are not implemented'
                break
            if name not in CANONICAL and name not in METADATA:
                reason = f'unsupported directive @{name}: {value}'
                break
            if ',' in value and name != 'lib':
                reason = f'option variants @{name}: {value}'
                break
            if name == 'declaration' and value != 'false':
                reason = 'declaration diagnostics require emit, outside noEmit subset'
                break
        if reason is None and source.stem.lower() in variants:
            reason = 'variant errors baselines: configuration expansion is not implemented'
        if reason is None and (len(raw) > 8192 or len(raw.splitlines()) > 80):
            reason = 'initial scope bound: more than 80 lines or 8192 bytes'
        try:
            content, options = materialize(raw) if reason is None else ('', {})
        except ValueError as error:
            reason = str(error)
        if reason is None and re.search(r'///\s*<reference|\b(?:import|export)\b[^\n]*[\'"].', content):
            # Any external module specifier can depend on a virtual host or mounted packages.
            reason = 'references or external module specifiers require host resolution'
        baseline = baselines.get((source.stem + '.errors.txt').lower())
        expected = b''
        if reason is None and baseline:
            try:
                expected = baseline_output(baseline.read_bytes())
            except ValueError:
                reason = 'global or non-file diagnostics: CLI configuration context differs'
            names = re.findall(rb'^([^\n]+?)\(\d+,\d+\):', expected, re.M)
            if any(name.decode() != source.name for name in names):
                reason = 'diagnostics outside the single source require virtual host paths'
            codes = list(map(int, re.findall(rb'error TS(\d+):', expected)))
            if any(code < 2000 for code in codes):
                reason = 'syntax diagnostic baseline: API collects errors that CLI suppresses'
            if 18027 in codes:
                reason = 'TS18027 is an emit-resolver diagnostic outside noEmit'
        row = {'source': path, 'source_sha256': digest(raw)}
        if reason:
            excluded.append({**row, 'reason': reason})
        else:
            selected.append({**row, 'options': options,
                             'baseline': baseline.relative_to(tree).as_posix() if baseline else None,
                             'baseline_sha256': digest(baseline.read_bytes()) if baseline else None,
                             'expected_sha256': digest(expected), 'name': source.name})
    inputs = [{'name': row['source'], 'content': materialize((tree / row['source']).read_bytes())[0]}
              for row in selected]
    rejected = set(json.loads(subprocess.check_output(
        ['node', str(DRIVER / 'syntax.cjs'), str(api)], input=json.dumps(inputs), text=True)))
    eligible = []
    for row in selected:
        if row['source'] in rejected:
            excluded.append({'source': row['source'], 'source_sha256': row['source_sha256'],
                             'reason': 'stock parser syntax errors: API and CLI diagnostic collection differ'})
        else:
            eligible.append(row)
    return {'upstream_commit': PIN, 'total': len(sources), 'selected': len(eligible),
            'excluded': len(excluded), 'reasons': dict(sorted(collections.Counter(
                row['reason'] for row in excluded).items())), 'cases': eligible,
            'exclusions': sorted(excluded, key=lambda row: row['source'])}


if __name__ == '__main__':
    tree, api, destination = map(Path, sys.argv[1:])
    manifest = select(tree.resolve(), api.resolve())
    destination.write_text(json.dumps(manifest, indent=2) + '\n')
    print(json.dumps({key: value for key, value in manifest.items() if key not in ('cases', 'exclusions')}))

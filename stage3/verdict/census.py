#!/usr/bin/env python3
"""Select baseline cases from inputs, never from the tested CLI's results."""
import collections
import hashlib
import json
from pathlib import Path
import posixpath
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parent
DRIVER = ROOT.parent / 'drivers/tsc'
sys.path.insert(0, str(DRIVER))
from corpus import PIN
from cases import parse, configurations, exhaustive_emit, DECLARATION_ERRORS, virtual_name


DIAGNOSTIC_SOURCES = re.compile(rb'^([^\n]+?)\((?:\d+|--),(?:\d+|--)\):', re.M)


def diagnostic_sources(summary):
    plain = plain_summary(summary)
    return [name.decode() for name in DIAGNOSTIC_SOURCES.findall(plain)]


def plain_summary(summary):
    plain = re.sub(rb'\x1b\[[0-9;]*m', b'', summary)
    return re.sub(rb'^([^\n]+?):(\d+):(\d+) - ', rb'\1(\2,\3): ', plain, flags=re.M)


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def summary_bytes(raw):
    text = raw.decode('utf-8-sig').replace('\r\n', '\n')
    pretty = text.startswith('\x1b[')
    summary = re.split(r'\n\n(?=====|!!!)', text, maxsplit=1)[0] if pretty else text.split('\n\n', 1)[0]
    if not re.search(rb'(?:error|message) TS\d+:', plain_summary(summary.encode())):
        raise ValueError('diagnostic summary requires harness-specific formatting')
    return summary.encode() if pretty else (summary + '\n').encode()


def select(tree, api):
    if subprocess.check_output(['git', '-C', str(tree), 'rev-parse', 'HEAD'], text=True).strip() != PIN:
        raise ValueError('upstream pin mismatch')
    reference = tree / 'tests/baselines/reference'
    baselines = {p.name.lower(): p for p in reference.glob('*.errors.txt')}
    sources = sorted(p for suite in ('compiler', 'conformance')
                     for p in (tree / 'tests/cases' / suite).rglob('*')
                     if p.suffix in ('.ts', '.tsx'))
    selected, excluded, configuration_exclusions = [], [], []
    projects = []
    for source in sources:
        raw = source.read_bytes()
        path = source.relative_to(tree).as_posix()
        reason = None
        try:
            units, settings, roots = parse(raw, source.name)
            configs = configurations(settings)
        except ValueError as error:
            reason = str(error)
        row = {'source': path, 'source_sha256': digest(raw)}
        rows = []
        if reason is None:
            if settings['__project']:
                projects.append({'source': path, 'units': units, 'project': settings['__project']})
            for suffix, options in configs:
                config_reason = None
                name = source.stem + (f'({suffix})' if suffix else '') + '.errors.txt'
                baseline = baselines.get(name.lower())
                expected = b''
                if baseline:
                    try:
                        expected = summary_bytes(baseline.read_bytes())
                    except ValueError as error:
                        config_reason = str(error)
                allowed_names = [u['name'] for u in units]
                allowed_names += [virtual_name(u['name'], settings) for u in units]
                library_names = sorted(set(re.findall(rb'^(lib\.[^/\n]+?\.d\.ts)\(--,--\):', plain_summary(expected), re.M)))
                allowed_names += [name.decode() for name in library_names]
                if any(name not in allowed_names for name in diagnostic_sources(expected)):
                    config_reason = 'diagnostics outside materialized units require harness-mounted paths or library placeholders'
                plain = plain_summary(expected)
                codes = list(map(int, re.findall(rb'error TS(\d+):', plain)))
                exhaustive = exhaustive_emit(options, expected)
                if not exhaustive and re.search(rb'^error TS\d+:', plain, re.M) and diagnostic_sources(expected):
                    config_reason = 'global/options diagnostics suppress semantic diagnostics on CLI; API collects both'
                project_name = settings['__project']
                if project_name and not exhaustive:
                    diagnostic_headers = re.findall(rb'^([^\n]+?)\(\d+,\d+\): error TS(\d+):', plain, re.M)
                    option_errors = [name.decode() == project_name and 5000 <= int(code) < 7000
                                     for name, code in diagnostic_headers]
                    if any(option_errors) and not all(option_errors):
                        config_reason = 'config option diagnostics suppress semantic diagnostics on CLI; API collects both'
                global_codes = set(map(int, re.findall(rb'^error TS(\d+):', plain, re.M)))
                if (not exhaustive and any(5000 <= code < 7000 for code in global_codes)
                        and any(code < 5000 or code >= 7000 for code in global_codes)):
                    config_reason = 'option diagnostics suppress global diagnostics on CLI; API collects both'
                js_syntax = set(map(int, re.findall(rb'^.*\.(?:[cm]?js|jsx)\(\d+,\d+\): error TS(80\d\d):', plain, re.M)))
                if not exhaustive and js_syntax and set(codes) != js_syntax:
                    config_reason = 'JavaScript syntax errors suppress semantic diagnostics on CLI; API collects both'
                if b'error TS-1:' in expected:
                    config_reason = 'pre/post emit consistency diagnostics exist only in the API harness'
                if ((options.get('noEmit') or options.get('noEmitOnError'))
                        and set(codes) & DECLARATION_ERRORS and set(codes) - DECLARATION_ERRORS):
                    config_reason = 'semantic errors suppress declaration diagnostics in CLI noEmit/noEmitOnError; API collects both'
                if config_reason:
                    configuration_exclusions.append({**row, 'configuration': suffix, 'reason': config_reason})
                    continue
                rows.append({**row, 'options': options, 'configuration': suffix,
                             'project': settings['__project'],
                             **({'library_placeholders': [name.decode() for name in library_names]} if library_names else {}),
                             'baseline': baseline.relative_to(tree).as_posix() if baseline else None,
                             'baseline_sha256': digest(baseline.read_bytes()) if baseline else None,
                             'expected_sha256': digest(expected), 'name': source.name})
        if reason or not rows:
            reason = reason or 'all configurations need API diagnostic collection or harness formatting'
            excluded.append({**row, 'reason': reason})
        else:
            selected.extend(rows)
    project_options = json.loads(subprocess.check_output(
        ['node', str(ROOT / 'project_options.cjs'), str(api)], input=json.dumps(projects), text=True))
    filtered = []
    for row in selected:
        if row['project']:
            project = project_options[row['source']]
            if project['errors']:
                configuration_exclusions.append({**row, 'reason': 'config parse diagnostics are discarded by API compiler runner but reported by CLI'})
                continue
            row['effective_options'] = {**project['options'], **row['options']}
            if parse((tree / row['source']).read_bytes(), row['name'])[1].get('__namespace'):
                row['project_files'] = project['files']
        else:
            row['effective_options'] = row['options']
        filtered.append(row)
    inputs = [{'name': row['source'] + ':' + unit['name'], 'content': unit['content']}
              for row in {row['source']: row for row in filtered}.values()
              for unit in parse((tree / row['source']).read_bytes(), row['name'])[0]
              if Path(unit['name']).suffix in ('.ts', '.tsx', '.mts', '.cts', '.js', '.jsx', '.cjs', '.mjs', '.json')]
    parsed = json.loads(subprocess.check_output(
        ['node', str(ROOT / 'parser_scope.cjs'), str(api)], input=json.dumps(inputs), text=True))
    parser_codes = collections.defaultdict(set)
    for name, codes in parsed.items():
        parser_codes[name.rsplit(':', 1)[0]].update(codes)
    eligible = []
    for row in filtered:
        codes = parser_codes[row['source']]
        expected = summary_bytes((tree / row['baseline']).read_bytes()) if row['baseline'] else b''
        expected_codes = set(map(int, re.findall(rb'error TS(\d+):', plain_summary(expected))))
        if codes and expected_codes != codes and not exhaustive_emit(row['effective_options'], expected):
            configuration_exclusions.append({'source': row['source'], 'source_sha256': row['source_sha256'],
                'configuration': row['configuration'],
                'reason': 'parser errors suppress other diagnostics on CLI; API collects both'})
        else:
            eligible.append(row)
    eligible_sources = {row['source'] for row in eligible}
    missing = {row['source']: row for row in selected if row['source'] not in eligible_sources}
    for row in missing.values():
        excluded.append({'source': row['source'], 'source_sha256': row['source_sha256'],
                         'reason': 'all configurations need API diagnostic collection or harness formatting'})
    return {'upstream_commit': PIN, 'total': len(sources), 'selected': len({row['source'] for row in eligible}),
            'configurations': len(eligible),
            'configuration_exclusions': configuration_exclusions,
            'excluded': len(excluded), 'reasons': dict(sorted(collections.Counter(
                row['reason'] for row in excluded).items())), 'cases': eligible,
            'exclusions': sorted(excluded, key=lambda row: row['source'])}


if __name__ == '__main__':
    tree, api, destination = map(Path, sys.argv[1:])
    manifest = select(tree.resolve(), api.resolve())
    destination.write_text(json.dumps(manifest, indent=2) + '\n')
    print(json.dumps({key: value for key, value in manifest.items()
                      if key not in ('cases', 'exclusions', 'configuration_exclusions')}))

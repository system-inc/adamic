#!/usr/bin/env python3
"""Validate census coverage and emit the morning/evening checker/lowering meter."""
import json
import re
from pathlib import Path
import subprocess
import sys


def report(tree, run, stamp):
    compiler = tree / 'src/compiler'
    expected = {str(path.resolve()): path for path in compiler.rglob('*') if path.is_file()}
    records = [json.loads(line) for line in (run / 'census.jsonl').read_text().splitlines()]
    # A root imports other files. Attribute diagnostics to their primary location,
    # never to the root that happened to load them or to an elaboration line.
    located = {}
    unlocated = set()
    external = set()
    for record in records:
        for diagnostic in record.get('diagnostics', []):
            match = re.match(r'^(.+?):[0-9]+:[0-9]+: error TS[0-9]+:', diagnostic)
            if match is None:
                if not diagnostic.startswith('error TS'):
                    raise ValueError(f'unrecognized diagnostic format: {diagnostic}')
                unlocated.add(diagnostic)
                continue
            name = str(Path(match[1]).resolve())
            if name in expected:
                located.setdefault(name, set()).add(diagnostic)
            else:
                external.add(diagnostic)
    rows = {}
    whole = None
    sources = sorted(name for name in expected if Path(name).suffix in ('.ts', '.a'))
    for record in records:
        roots = [str(Path(root).resolve()) for root in record['roots']]
        if len(roots) != 1:
            if whole is not None or sorted(roots) != sources:
                raise ValueError('invalid or duplicate whole-program census record')
            whole = record
            continue
        name = roots[0]
        if name not in expected or name in rows:
            raise ValueError(f'unexpected or duplicate census file: {name}')
        kind = record['kind']
        if kind not in ('accepted', 'checker', 'Refused', 'NotYet', 'error'):
            raise ValueError(f'unknown census kind: {kind}')
        checked = kind in ('accepted', 'Refused', 'NotYet')
        rows[name] = {
            'file': str(expected[name].relative_to(tree)),
            'source': Path(name).suffix in ('.ts', '.a'),
            'checker': checked,
            'checker_whole_program': checked,
            'checker_own_file': Path(name).suffix in ('.ts', '.a') and kind != 'error' and name not in located,
            'own_file_diagnostics': len(located.get(name, set())),
            'lowering': kind == 'accepted',
            'lowering_attempted': checked,
            'outcome': kind,
        }
    if set(rows) != set(expected) or whole is None:
        raise ValueError('census coverage incomplete: per-file or whole-program rows missing')
    files = [rows[name] for name in sorted(rows)]
    totals = {
        'files': len(files),
        'source_files': sum(row['source'] for row in files),
        'checker': sum(row['checker'] for row in files),
        'lowering': sum(row['lowering'] for row in files),
        'checker_whole_program': sum(row['checker_whole_program'] for row in files),
        'checker_own_file': sum(row['checker_own_file'] for row in files),
    }
    result = {
        'timestamp_utc': stamp,
        'adamic_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
        'profile': 'unmodified stage 0 checker options and prelude',
        'adapted_tree': str(tree),
        'unlocated_diagnostics': len(unlocated),
        'external_diagnostics': len(external),
        'files': files,
        'totals': totals,
        'whole_program': {key: value for key, value in whole.items() if key != 'diagnostics'},
        'whole_program_diagnostics': len(whole.get('diagnostics', [])),
    }
    (run / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
    lines = [f"Whole program: {totals['checker_whole_program']}/{totals['source_files']}",
             f"Own file: {totals['checker_own_file']}/{totals['source_files']}", '',
             f'Stage 3 meter {stamp}', '', '| File | Whole program | Own file | Lowering |',
             '| --- | --- | --- | --- |']
    for row in files:
        lowering = 'pass' if row['lowering'] else 'fail' if row['lowering_attempted'] else 'blocked'
        lines.append(f"| {row['file']} | {'pass' if row['checker'] else 'fail'} | {'pass' if row['checker_own_file'] else 'fail'} | {lowering} |")
    lines += ['', f"Checker: {totals['checker']}/{totals['source_files']} source files. "
              f"Lowering: {totals['lowering']}/{totals['source_files']} source files.",
              f"All files: {totals['files']}; JSON inputs fail the source-extension gate.",
              'Blocked means lowering was not attempted because the checker/input gate failed.', '']
    (run / 'report.md').write_text('\n'.join(lines))
    return result


def report_pair(main_tree, area_tree, run, stamp, main_commit, area_commit):
    trees = {}
    for label, tree, commit in [('main', main_tree, main_commit), ('area', area_tree, area_commit)]:
        trees[label] = report(tree, run / label, stamp)
        trees[label]['tree_ref'] = 'origin/main' if label == 'main' else 'origin/area/stage3'
        trees[label]['tree_commit'] = commit
        (run / label / 'report.json').write_text(json.dumps(trees[label], indent=2) + '\n')
    # Preserve the existing area's single-tree fields for JSON consumers.
    result = dict(trees['area'], trees=trees)
    (run / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
    lines = []
    for key, title in [('checker_whole_program', 'Whole program'), ('checker_own_file', 'Own file')]:
        values = [f"{label}: {trees[label]['totals'][key]}/{trees[label]['totals']['source_files']}" for label in trees]
        lines.append(f"{title}: " + '; '.join(values))
    lines += ['', f'Stage 3 meter {stamp}', '',
              f'Main: origin/main at {main_commit}. Area: origin/area/stage3 at {area_commit}.',
              f"Both measured with Adamic {result['adamic_commit']} and ordinary stage 0 options.", '',
              '| File | Main whole program | Main own file | Area whole program | Area own file |',
              '| --- | --- | --- | --- | --- |']
    rows = {label: {row['file']: row for row in tree['files']} for label, tree in trees.items()}
    for name in sorted(set(rows['main']) | set(rows['area'])):
        cells = []
        for label in trees:
            row = rows[label].get(name)
            for key in ['checker_whole_program', 'checker_own_file']:
                cells.append('absent' if row is None else 'pass' if row[key] else 'fail')
        lines.append('| ' + ' | '.join([name, *cells]) + ' |')
    lines += ['', 'Whole program includes imported diagnostics. Own file uses the primary diagnostic location.',
              'Global and external diagnostics are counted separately in JSON; they have no compiler-file location.',
              'Non-source inputs fail the extension gate and are excluded from source denominators.', '']
    (run / 'report.md').write_text('\n'.join(lines))
    return result


if __name__ == '__main__':
    if len(sys.argv) == 4:
        report(Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve(), sys.argv[3])
    else:
        report_pair(Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve(),
                    Path(sys.argv[3]).resolve(), *sys.argv[4:7])

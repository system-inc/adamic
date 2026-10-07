#!/usr/bin/env python3
"""Validate census coverage and emit the morning/evening checker/lowering meter."""
import json
from pathlib import Path
import subprocess
import sys


def report(tree, run, stamp):
    compiler = tree / 'src/compiler'
    expected = {str(path.resolve()): path for path in compiler.rglob('*') if path.is_file()}
    records = [json.loads(line) for line in (run / 'census.jsonl').read_text().splitlines()]
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
    }
    result = {
        'timestamp_utc': stamp,
        'adamic_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
        'profile': 'unmodified stage 0 checker options and prelude',
        'adapted_tree': str(tree),
        'files': files,
        'totals': totals,
        'whole_program': {key: value for key, value in whole.items() if key != 'diagnostics'},
        'whole_program_diagnostics': len(whole.get('diagnostics', [])),
    }
    (run / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
    lines = [f'Stage 3 meter {stamp}', '', '| File | Checker | Lowering |', '| --- | --- | --- |']
    for row in files:
        lowering = 'pass' if row['lowering'] else 'fail' if row['lowering_attempted'] else 'blocked'
        lines.append(f"| {row['file']} | {'pass' if row['checker'] else 'fail'} | {lowering} |")
    lines += ['', f"Checker: {totals['checker']}/{totals['source_files']} source files. "
              f"Lowering: {totals['lowering']}/{totals['source_files']} source files.",
              f"All files: {totals['files']}; JSON inputs fail the source-extension gate.",
              'Blocked means lowering was not attempted because the checker/input gate failed.', '']
    (run / 'report.md').write_text('\n'.join(lines))


if __name__ == '__main__':
    report(Path(sys.argv[1]).resolve(), Path(sys.argv[2]).resolve(), sys.argv[3])

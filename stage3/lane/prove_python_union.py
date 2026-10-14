#!/usr/bin/env python3
"""Compare actual Go-wrapped case records with fresh executions of the old Python CLIs."""
import argparse
from collections import Counter
import json
import os
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[2]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('measurements', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=False)
    manifest = json.loads((ROOT / 'stage3/lane/python-units.json').read_text())
    cache = Path(os.environ['ADAMIC_BUILD_CACHE_DIR']) / 'local'
    prepared = [path / 'products' for path in cache.iterdir() if path.is_dir() and (path / 'products').is_dir()]
    if len(prepared) != 1:
        raise ValueError('expected exactly one prepared apply archive')
    env = dict(os.environ, STAGE3_PRODUCT_STORE=str(prepared[0]), ADAMIC_BUILD_CACHE='read', PYTHONDONTWRITEBYTECODE='1')
    summaries = []
    for family, command, cwd in [
            ('apply', ['python3', 'stage3/test_apply.py'], ROOT),
            ('lane', ['python3', '-m', 'unittest', '-v', 'test_check', 'test_table'], ROOT / 'stage3/lane')]:
        log = args.output / (family + '-native.log')
        with log.open('w') as stream:
            result = subprocess.run(command, cwd=cwd, env=env, stdout=stream, stderr=subprocess.STDOUT, timeout=90)
        if result.returncode:
            raise ValueError(f'old {family} command failed: {result.returncode}')
        old = [name.removeprefix('__main__.') for name in re.findall(r'^test_\w+ \(([^)]+)\) \.\.\.', log.read_text(), re.MULTILINE)]
        measured = []
        for unit in manifest['units']:
            if unit['family'] != family:
                continue
            records = json.loads((args.measurements / 'python' / (unit['name'] + '.json')).read_text())
            if not records['successful'] or any(row['status'] != 'pass' for row in records['cases']):
                raise ValueError('nonpassing observed Python unit: ' + unit['name'])
            measured.extend(row['case'] for row in records['cases'])
        if Counter(old) != Counter(measured) or Counter(old) != Counter(manifest['old_commands'][family]):
            raise ValueError(f'{family} Python case union mismatch: native={old}, Go={measured}')
        summaries.append(dict(command=command, exit=result.returncode, family=family,
                              native_cases=sorted(old), go_cases=sorted(measured), count=len(old), exact_union=True))
    report = dict(status='pass', commands=summaries, total_cases=sum(row['count'] for row in summaries))
    (args.output / 'python-comparison.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()

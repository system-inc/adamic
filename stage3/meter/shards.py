#!/usr/bin/env python3
"""Run independent meter or census test cases with a stable, exhaustive shard map."""
import argparse
import hashlib
import importlib
import json
import os
from pathlib import Path
import sys
import unittest

ROOT = Path(__file__).resolve().parents[2]
GROUPS = {
    'meter': ['stage3/meter/report_test.py', 'stage3/meter/entry_test.py',
              'stage3/meter/compiler_test.py'],
    'census': ['stage3/census/latent/replay/replay_test.py'],
}


def cases(suite):
    for item in suite:
        if isinstance(item, unittest.TestSuite):
            yield from cases(item)
        else:
            yield item


def assignment(name, count):
    return int.from_bytes(hashlib.sha256(name.encode()).digest()[:8], 'big') % count


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('group', choices=GROUPS)
    parser.add_argument('--list', action='store_true')
    parser.add_argument('--plant-failure', help='exact case ID; append a planted dispatch failure after its unchanged test')
    parser.add_argument('--result', type=Path)
    args = parser.parse_args()
    try:
        index, count = map(int, os.environ.get('ADAMIC_TEST_SHARD', '0/1').split('/'))
        if count < 1 or not 0 <= index < count:
            raise ValueError()
    except ValueError:
        parser.error('ADAMIC_TEST_SHARD must be i/n with 0 <= i < n')
    if not args.list:
        import subprocess
        products = json.loads(subprocess.check_output(
            ['python3', str(ROOT / 'stage3/meter/build-probes.py')], cwd=ROOT, text=True))
        os.environ.update(CENSUS_BINARY=products['census'],
                          LATENT_CENSUS_BINARY=products['latent'],
                          ENTRY_CENSUS_BINARY=products['entry'])
    tests = []
    for relative in GROUPS[args.group]:
        path = ROOT / relative
        sys.path.insert(0, str(path.parent))
        module = importlib.import_module(path.stem)
        tests.extend(cases(unittest.defaultTestLoader.loadTestsFromModule(module)))
    names = [test.id() for test in tests]
    if len(names) != len(set(names)) or not names:
        parser.error('case inventory must be nonempty and unique')
    if args.plant_failure and args.plant_failure not in names:
        parser.error('planted case does not exist')
    mapping = {name: assignment(name, count) for name in sorted(names)}
    if args.list:
        print(json.dumps(mapping, indent=2))
        return 0
    selected = [test for test in tests if mapping[test.id()] == index]
    result = unittest.TextTestRunner(verbosity=2).run(unittest.TestSuite(selected))
    for test in selected:
        if test.id() == args.plant_failure:
            try:
                raise AssertionError('planted shard failure: ' + test.id())
            except AssertionError:
                result.addFailure(test, sys.exc_info())
                print('planted shard failure: ' + test.id(), file=sys.stderr)
    if args.result:
        args.result.write_text(json.dumps(dict(cases=[test.id() for test in selected],
            failures=[test.id() for test, _ in result.failures],
            errors=[test.id() for test, _ in result.errors],
            skipped=[test.id() for test, _ in result.skipped], ran=result.testsRun), indent=2) + '\n')
    print('Shard result: ' + ('PASS' if result.wasSuccessful() else 'FAIL'), file=sys.stderr)
    return 0 if result.wasSuccessful() else 1


if __name__ == '__main__':
    raise SystemExit(main())

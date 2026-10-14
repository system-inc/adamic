#!/usr/bin/env python3
"""Inventory and run the exact unittest suites behind the three old lane commands."""
import argparse
import importlib.util
import json
from pathlib import Path
import sys
import unittest
sys.dont_write_bytecode = True
LANE = Path(__file__).resolve().parent
sys.path[:0] = [str(LANE), str(LANE.parent)]


def apply_module():
    spec = importlib.util.spec_from_file_location('stage3_test_apply', LANE.parent / 'test_apply.py')
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def flatten(suite):
    for test in suite:
        if isinstance(test, unittest.TestSuite):
            yield from flatten(test)
        else:
            yield test


def identity(test):
    return test.id().removeprefix('stage3_test_apply.')


def inventory():
    apply = apply_module()
    rows = []
    for index in range(17):
        rows.append(dict(name=f'TestStage3ApplyShard{index:02d}', family='apply', selector=str(index),
                         cases=sorted(identity(test) for test in apply.selected_suite(index, 17))))
    for module, cls in [('test_check', 'LandingLaneTests'), ('test_table', 'LaneTableTests')]:
        suite = unittest.defaultTestLoader.loadTestsFromName(module + '.' + cls)
        rows.append(dict(name='TestStage3Lane' + cls, family='lane', selector=module + '.' + cls,
                         cases=sorted(identity(test) for test in flatten(suite))))
    old = dict(apply=sorted(identity(test) for test in apply.suite_cases()), lane=sorted(
        identity(test) for test in flatten(unittest.defaultTestLoader.loadTestsFromNames(['test_check', 'test_table']))))
    for family in old:
        actual = sorted(case for row in rows if row['family'] == family for case in row['cases'])
        if actual != old[family] or len(actual) != len(set(actual)):
            raise ValueError('Python unit partition differs from old command discovery: ' + family)
    return dict(version=1, old_commands=old, units=rows)


def run(family, selector, output):
    suite = (apply_module().selected_suite(int(selector), 17) if family == 'apply' else
             unittest.defaultTestLoader.loadTestsFromName(selector))
    records = []
    class Result(unittest.TextTestResult):
        def startTest(self, test):
            records.append(dict(case=identity(test), status='started'))
            super().startTest(test)
        def addSuccess(self, test):
            records[-1]['status'] = 'pass'; super().addSuccess(test)
        def addFailure(self, test, err):
            records[-1]['status'] = 'fail'; super().addFailure(test, err)
        def addError(self, test, err):
            records[-1]['status'] = 'error'; super().addError(test, err)
        def addSkip(self, test, reason):
            records[-1]['status'] = 'skip'; super().addSkip(test, reason)
    result = unittest.TextTestRunner(verbosity=2, resultclass=Result).run(suite)
    output.write_text(json.dumps(dict(cases=records, successful=result.wasSuccessful())) + '\n')
    return 0 if result.wasSuccessful() else 1


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['inventory', 'apply', 'lane'])
    parser.add_argument('selector', nargs='?')
    parser.add_argument('output', nargs='?', type=Path)
    args = parser.parse_args()
    if args.mode == 'inventory':
        print(json.dumps(inventory(), indent=2, sort_keys=True))
    else:
        raise SystemExit(run(args.mode, args.selector, args.output))

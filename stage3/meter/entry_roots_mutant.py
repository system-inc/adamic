"""Restore literal entry-path comparison; equivalent-path fixtures must reject it."""
from pathlib import Path
import sys
import unittest

import report
from entry_test import EntryReportTests

if __name__ == '__main__':
    log = Path(sys.argv[1])
    report.entry_root_matches = lambda record, entry: record.get('roots') == [str(entry)]
    suite = unittest.TestSuite([EntryReportTests(
        'test_equivalent_symlink_entry_roots_are_accepted')])
    with log.open('w') as stream:
        result = unittest.TextTestRunner(stream=stream, verbosity=2).run(suite)
    if len(result.errors) != 2 or result.failures or any(
            'invalid tsc entry census roots or coverage' not in text for _, text in result.errors):
        raise SystemExit('literal-path mutant was not caught by both equivalent-path cases')
    print('Caught literal-entry-path mutant for checker-clean and checker-rejected trees')

"""Witness that omitting adaptation provenance fails the report fixture."""
from pathlib import Path
import sys
import unittest

import report
from report_test import ReportTests

if __name__ == '__main__':
    log = Path(sys.argv[1])
    report.ADAPTATION_ROOTS = {}
    suite = unittest.TestSuite([ReportTests(
        'test_adaptation_created_root_keeps_total_and_tsc_counts_separate')])
    with log.open('w') as stream:
        result = unittest.TextTestRunner(stream=stream, verbosity=2).run(suite)
    if len(result.failures) != 1 or result.errors:
        raise SystemExit('provenance mutant was not caught by the intended fixture')
    print('Caught missing-adaptation provenance mutant; witness: ' + str(log))

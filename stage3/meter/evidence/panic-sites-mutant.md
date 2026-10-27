# Mutant evidence for the named panic and error sites in stage3/meter/report.py
# Test: report_test.LatentReportTests.test_panic_and_error_sites_are_named_from_src
# Run from stage3/meter: python3 -E -S -B -m unittest report_test.LatentReportTests.test_panic_and_error_sites_are_named_from_src

## With panic_sites (expected pass)
Ran 1 test in 0.002s

OK
exit=0

## Mutant 1: the 'panic_sites' key dropped from lowering_summary (expected red)
KeyError: 'panic_sites'
FAILED (errors=1)
exit=1

## Mutant 2: 'panic_sites' always the empty list (expected red)
AssertionError: Lists differ: [] != [{'where': 'src/compiler/main.a:3:1', 'tex[96 chars]ce'}]
FAILED (failures=1)
exit=1

## Restored (expected pass): python3 -E -S -B -m unittest report_test entry_test
Ran 34 tests in 0.209s

OK (skipped=4)

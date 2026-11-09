# Mutant evidence for the tsc entry pass's relation budget (stage3/meter, #ny0c3gb item 5)
# Run on Codex at e0d63ae7 (go1.27.1 linux/amd64), from stage3/meter: python3 -m unittest compiler_test report_test -v

## Fixed (expected pass)
Ran 63 tests, OK (skipped=2: optional real-binary report probes).
Overlay builds: make_overlay.py then go build -overlay of stage3/census/latent/tool, exit 0; through entry_overlay.py, exit 0.
Overlaid census tests pass: TestLatentCensusOverflowMarkReachesProgramAndEveryAttempt,
TestLatentCensusEndpointOverflowMarksAllAttemptsAndExplainsOnset, TestLatentStreamCutoffRetainsHeaderWithoutEndAnswer,
TestLatentStreamCompletesWithRelationTrailerAfterFiles; stage3/census/latent/overlaytext, exit 0.
A real run on a two-line program wrote cache_sizes
{"assignable":230,"comparable":0,"enum":0,"identity":0,"strictSubtype":4,"subtype":0}.

## Mutant (a): report.py's threshold compare never fires (expected red)
FAILED (failures=3)
test_relation_budget_red_day_keeps_its_record_and_exits_nonzero: 0 != 3
RelationBudgetTests.test_threshold_is_ok_and_one_past_it_is_red: ('ok', 1000001) != ('red', 1000001)
ReportTests.test_relation_budget_reaches_report_json_and_heads_the_entry_section: status ok != red

## Mutant (b): twice-daily.sh's last step no longer exits 3 (expected red)
FAILED (failures=2)
test_relation_budget_red_day_keeps_its_record_and_exits_nonzero: 0 != 3
test_relation_budget_unknown_day_without_trailer_exits_nonzero: 0 != 3

## Mutant (c): the script stops right after report.py, before gzip, timing and README (expected red)
FAILED (failures=8): every measure() test 3 != 0, and both red-day tests
"timing.json missing: the record stopped before the exit".

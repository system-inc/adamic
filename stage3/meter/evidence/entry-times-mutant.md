# Mutant evidence for the entry census writing tsc/latent-times.jsonl (stage3/meter/twice-daily.sh, #20pb346)
# Test: compiler_test.CompilerSelectionTests, whose measure() reads every copied stream, tsc/latent-times.jsonl.gz among them
# Run on Codex at 819116b9 (go1.27.1 linux/amd64, Python 3.12.14), from stage3/meter: python3 -m unittest compiler_test -v

## With the entry census timed (expected pass)
fixed exit=0
Ran 9 tests in 11.111s
OK

## Mutant: twice-daily.sh from main 7e07230bfb, the entry census without LATENT_TIMES (expected red)
## METER_SCRIPT_UNDER_TEST=<main's twice-daily.sh> python3 -m unittest compiler_test -v
mutant exit=1
Ran 9 tests in 11.260s
FAILED (errors=5)
ERROR test_default_uses_checkout_compiler_for_both_trees (compiler_test.py:180, self.measure())
ERROR test_explicit_pin_uses_requested_commit (compiler_test.py:270, self.measure('per-ref'))
ERROR test_legacy_latent_mode_and_stale_full_claim (compiler_test.py:230, self.measure('per-ref'))
ERROR test_missing_local_dependency_transport_retries_recorded_origins (compiler_test.py:284, self.measure('per-ref'))
ERROR test_per_ref_builds_and_runs_each_pinned_compiler (compiler_test.py:193, self.measure('per-ref'))
Each ends in FileNotFoundError: [Errno 2] No such file or directory for its run's main/tsc/latent-times.jsonl.gz.

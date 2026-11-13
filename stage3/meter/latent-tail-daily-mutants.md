# Daily scoped tail mutant evidence (#sx8wdg7)

The daily meter now runs both latent passes, the entry pass over tsc.ts and the
per-file pass over src/compiler, with `LATENT_TAIL=1`, so each unit's scoped
post-lowering tail runs on its snapshot (#hcxsh1k). Body findings stay
byte-identical; tail stops are reported on their own line and never move the
body totals. The two checker censuses are unchanged.

The fake drivers in `compiler_test.py` append each invocation's `LATENT_TAIL`
value, and the case below reads all four. Python cases are named by class and
case only.

| Mutant | Line changed | Case it turns red | Expected result |
| --- | --- | --- | --- |
| Drop the tail from the entry pass | In `twice-daily.sh`, remove `LATENT_TAIL=1 ` before `"$binaries/entry-census"`. | `CompilerSelectionTests.test_daily_runs_both_latent_passes_with_scoped_tail_sx8wdg7` | Fail: three empty tail values and one `tail=1`. |
| Drop the tail from the per-file pass | In `twice-daily.sh`, remove `LATENT_TAIL=1 ` before `"$binaries/latent-census"`. | `CompilerSelectionTests.test_daily_runs_both_latent_passes_with_scoped_tail_sx8wdg7` | Fail, the same way. |
| Leak the tail into a checker census | In `twice-daily.sh`, prefix the first `"$binaries/census"` call with `LATENT_TAIL=1 `. | `CompilerSelectionTests.test_daily_runs_both_latent_passes_with_scoped_tail_sx8wdg7` | Fail: three `tail=1` values. |

# The pos family's meter pin (#aryr7ea)

`report_test.LatentReportTests.test_recorded_pos_family_reads_zero_once_node_ranges_are_writable`
reads the run of record 20261122T131040Z.tNkys1 (main 178b45d3, which holds adaptations 80 and 85)
and finds 0 reasons whose readonly field `pos` or `end` becomes writable, in both the latent and the
tsc entry summaries. The Nov 15 record 20261115T131735Z.PG4H0R, before either adaptation, holds 173
such reasons and 550 sites in each, so the pattern is shown to match the family's real wording.

Proved in Loom job J-b6zdzjc at 1f4ebb89 (sealed, azure-d64a-1122-1718-1):

| Run | Change | Result |
| --- | --- | --- |
| pin | none | ok |
| suite | `python3 -m unittest discover -s stage3/meter -p '*_test.py'` | 110 tests ok, 8 skipped |
| mutant 1 | the pin reads the Nov 15 record as today's | red: `('7816a8d8…', [(173, 550), (173, 550)])` against `('178b45d3…', [(0, 0), (0, 0)])` |
| mutant 2 | the family's pattern names `start` in place of `pos` or `end` | red: the Nov 15 guard reads `[(0, 0), (0, 0)]` against `[(173, 550), (173, 550)]` |

Each mutant was applied with sed, counted as applied, and restored from a copy; `git diff` read clean after.

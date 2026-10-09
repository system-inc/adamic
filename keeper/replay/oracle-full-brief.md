New unit for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt): rp:oracle-full. Your workspace is warm, so skip setup if /workspace/adamic-tools/env.sh works. Check `df -h /tmp /workspace` first, and clear earlier units' scratch under /tmp if space is low (never /workspace/adamic, your tools or ~/.cache). Write evidence outside the repository, then copy it in at the end.

## What this settles
The internal/oracle deletion-set replay (test-defend/deletion-set/internal-oracle, commit a8add762, base main 7b9d4272c28f59530ab13daa5c49067e47933b06) stopped each mutant at its first failing test. Identical diffs got different first catchers (b11-024 and b11-033, for example), so its "still caught by" names only one catcher. The reviewer judges deletions by a stricter rule: a test may go only if each of its mutants is caught by another test on the same kind of fixture. So this pass needs every catcher.

## Do exactly this
Check out 7b9d4272 detached, with submodules at its pins. Take the diffs from review/test-defend/deletion-set/internal-oracle/diffs/ on that evidence branch:
b02-003-M14, b02-005-M8, b05-013-M10, b06-014-M4, b07-017-M04, b07-018-M05, b10-021-M06, b11-023-D02, b12-041-D4, b14-043-D3, b14-050-D5, b14-051-D6, b14-052-D7, b16-058-D02, b17-065-D4, b17-066-D5, b17-067-D6, b21-074-D2, b21-076-D4

For each one:
- apply it, and run the whole of ./internal/oracle once with `-count=1 -timeout 60m -json`. Don't use -failfast, and let it run to the end.
- use -skip for the 22 deletion candidates (the same expression the set replay used, from its report.json) and also for TestCountsAreRecorded, plus every test whose name contains Mutant, Mutants or Witness. Those plant their own mutants or pin the counts table, so they never count as a catcher.
- list every top-level test that FAILs. For a panic, name the test that panicked and the first frame inside internal/ or stage1/, and keep going, rerunning with that test skipped until the run completes.
- revert the diff.
Run two at a time if the machine allows. Clean the build cache under /tmp between mutants if disk runs low.

Push the logs and a summary JSON on test-defend/deletion-set/internal-oracle-full under review/test-defend/deletion-set/internal-oracle-full/, and push that branch only. Restore the tree.

Reply in two or three sentences of prose, then this JSON:
{"base":"7b9d4272","skipped_extra":["TestCountsAreRecorded","..."],"mutants":[{"diff":"b05-013-M10","file_line":"...","all_catchers":["..."],"panics":[{"test":"...","frame":"..."}],"completed":true}]}

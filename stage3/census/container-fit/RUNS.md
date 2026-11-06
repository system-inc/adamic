# Resuming latent pieces

The statecopy branch is one commit on A: 62ddd47969e5168e5a0e74a23498b85d144b2361. The logging sibling plus the same patch is local 29de73556749a0374da691bb9599e0e024d1214a. Neither optimizer branch nor main is pushed.

The frozen plan is evidence/plan/INDEX.json. It covers 79 files and 4,557 top-level units. p00 is the whole utilitiesPublic.ts file. p01 through p14 partition utilities.ts into disjoint contiguous top-level chunks. p15 through p19 cover all remaining files in disjoint groups. Cost estimates use historical function names, not measurements for this task. The adapted input is the same completed tree used for the ordinary runs; source-manifest.json is checked before every child.

Build scripts and all binaries are pinned under evidence/latent-build. Original and scheduling-only selector binaries are in /workspace/container-fit-build-v2/{base,head,log}/. The selector audit passed mode-off identity for all three, fixture split first stops, and lost/duplicate/moved-work controls. No production lowering source changes are overlaid beyond the normal meter.

Run each piece as its own foreground tool process. Never wrap the entire plan in a long-running process. Source /workspace/adamic-tools/env.sh first. For example:

```sh
python3 stage3/census/container-fit/run_piece.py latent base p00 @/workspace/adamic/stage3/census/container-fit/evidence/plan/p00.json
```

Use label log for the timed logging head and head only for the p00 byte-identity check. Each invocation stops its child by 1,500 seconds, allows at most 10 seconds for termination, compresses raw evidence, commits only the piece directory, and ordinarily pushes to typescript/container-fit-census. It returns only after publication succeeds or an error has been recorded. Do not start another piece after a push failure. Do not rerun a completed observation: inspect evidence/latent/{label}/{piece}/metrics.json, verify its raw hashes and its recorded binary SHA, then continue with the next missing piece. A directory without completed metrics must be preserved as interrupted evidence before a retry.

Start p00 base, log, then head. Compare the decompressed log/head records.jsonl bytes exactly. Other pieces need base and log only. All measured children run sequentially so timing does not include another census process. Preserve the log append overhead in head times and label all timing results as one sample on this box.

The previous ordinary 79-root records are unchanged, because the new code only changes measurement state generation. Their results and proof are in ordinary-summary.json and STOPPED-REPORT.md. The old summarize.py/finish.py describe that initial stopped report; do not use them to overwrite the completed latent report. Final latent pairing must use main stage3/census/pair and distinguish changed stops in already blocked units from true regressions.

## Revised public schedule

Initial base p00 timed out at 1,500.120056 seconds and was pushed at 258190b0c. Do not pair or rerun it. INDEX-initial.json preserves the initial plan. Current INDEX.json replaces p00 with q00 onward public chunks; every one needs base, log, and head, with a byte comparison after its head run. Other p01 through p19 remain unchanged. The current index lists exactly one occurrence of every unit.

API orchestration may invoke the independent per-piece calls in order; it does not create a long-lived supervisor process on the box. Every runner still exits after its own push. After the labels for a piece complete, run check_piece.py PIECE as a separate short process to pair it, verify public bytes and push the live index. Stop orchestration at any nonzero exit.

## Supplemental full-file check

After the primary partition is complete, r00 measures the whole utilitiesPublic file for head and log only. It is a duplicate validation sample, excluded from primary walk totals and LATENT_TIMES comparisons. Observed public unit sums plus setup predict about 27 minutes. Each invocation uses a 1,740 second child ceiling, at most 10 seconds for termination, and a 15 second publication timeout. An outer `timeout --kill-after=5s 1780` bounds every box process below 30 minutes, including the supervisor. Stop and retry publication separately if it fails. Never include r00 in the primary partition index.

Supplemental r00 outcome: head intentionally stopped at 1740.113 seconds with 193 completed timer records (exact count is in whole-public-status.json). Whole-file log run was not started. The completed primary partition and seven exact public chunk comparisons are unchanged.

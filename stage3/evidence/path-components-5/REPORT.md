# PathPathComponents lane metadata repair

Source commit: `b6f384ebe9f9cf18777f6733ab632ea1cc7bd98d`, one commit on main `4bba06cb2dc4a0046c11fc97e9953ec85ed0e638`.
Branch: `typescript/path-components-interface-5`.

Regenerated API metadata from the repository root at 65a42f81, then verified regeneration on the main-tip cuts. Only sanctioned-api.json proofs[1].sha256 (40's class-rules.json) and shard-plan.json runtime change. No sanction row is added; 267 normalized changes remain; test_check.py is byte-identical. generate_shards.py and --check pass.

The unfiltered stage3/lane package passes all 430 top-level tests, including all 398 shards and TestLaneSanctionedApiFreshnessPython. Actual Go exit code: 0. Process duration: 1709.67 seconds. Full apply rebuilt all adapters successfully at 529705a1e (716 seconds). The final main advancement changes only internal/lower files: every apply and adapted-tree input is unchanged. Ordinary stage3/apply.sh on the final cut restores that identical keyed product successfully. API regeneration and shard freshness pass again on the final cut.

Main advanced twice during validation. The complete lane also passed at the a93ac0b94 and 529705a1e cuts before the final run on 4bba06cb2. Preserve all ledger entries in each recut.

## Run history and exit reporting

The initial cold lane attempt was stopped with its real signal status because cold prerequisite work consumed the bounded process budget. A parallel=4 attempt then failed on 16 deadline/time-limit cases; retain its complete output and failure details. An unfiltered parallel=3 run passed at each subsequent cut. No test deadlines or source checks were weakened. Every ordinary wrapper records its child's own exit code and enforces a process limit below 30 minutes.

During the final unfiltered run the workspace restarted. The Python wrapper died, but Go survived as an orphan and continued writing its original JSON output. recover-exit.py verified the PID and kernel start time, retained the original 1780-second process limit, and recovered Go's actual raw wait status 0 from /proc/PID/stat field 52 after it became a zombie. The final JSON package-pass record, all 430 test passes, and that kernel exit status agree. This is the original whole-package run, not a filtered retry.

Logs, status files, exact metadata comparisons, wrappers and recovery monitor accompany this report. Go snapshots use .go.txt. Product caches and adapted TypeScript trees are excluded. Earlier type/JavaScript/census evidence remains in path-components-3 on this evidence branch.

Starting commit: 09769cb5ddc8067d2fa052a8c760d813894c8919.

The M01-M10 diffs are standalone production changes without a selector. Apply one to that commit, run go vet ./internal/lower/, and use a unique ADAMIC_BUILD_CACHE_DIR for the replay. W01-W03 are explicitly authorized witness harness weakenings, not production mutants. They passed go vet ./internal/oracle/. P01 is an empty-answer probe, not a mutant; its standalone diff selects ADAMIC_MUTANT=P01 and passed go vet ./internal/lower/.

Every diff passed git apply --check against the restored starting source. Switched production sources are fixtures under switched-source/, not active Go files. run.py records the original scoped matrix and witness runs; family-replay.py corrects the slash-sensitive fixture filter. runs.jsonl and family-runs.jsonl preserve exact command argv, exit status and command wall time. matrix.json distinguishes production kills from witness-precondition failures. W01-standalone.log proves the standalone comparator diff.

The 35 additional fixture inputs appear in family-members.json. All 13 scoped names appear in names.json. No individual-row mutant first-catch run was needed; every scoped production column completed. Probe rows were run individually because nil program results can panic in backend preparation.

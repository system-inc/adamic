Mutant evidence for proofs/oracle-comparison.json and results.json (#jk7tph7 B1).

Mutant: append one extra line to the actual completed baseline.diff while keeping
both API snapshots unchanged. Command: python3
stage3/adapt/92-tty-host-results/proofs/oracles.py. Its baseline mutant calls
stage3/lane/check.py's api_check directly; the acceptance predicate is unchanged.
Result with the mutant: the byte checker rejects it with the real message:

    baseline.diff does not exactly describe the API snapshots (extra, missing, or changed diff lines)

The runner requires that rejection and records baseline-mutant.json. It restores
the exact diff bytes in finally. Restored command:

    python3 stage3/lane/check.py /tmp/tty-adaptation/lane-after

Result: exit 0, PASS stage3 landing lane [causes: timeout=0, baseline-content=1,
baseline-missing=0, exception=0, other=0]. The one baseline-content observation is
the same sanctioned API test as main, not a new accepted failure. Both canonical
full suites, both failed resource attempts, and the restored CLI outputs are
retained as compressed raw logs. No mutant remains in source or baseline data.

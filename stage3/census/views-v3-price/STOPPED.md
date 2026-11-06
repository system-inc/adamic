# Exact closure census stopped at the baseline budget

The unmodified main full-latent meter baseline for the group containing
binder.ts, _namespaces/ts.ts and _namespaces/ts.performance.ts exceeded its
23-minute limit. subprocess.TimeoutExpired reports 1379.998793027 seconds.
Its stream contains only the complete entry header, with no completed file
records. The measurement counter was not running in that child. This group
contributes no counts or units to the report.

budget-stop.tar.gz preserves the exact runner error, partial stream, process
log, selected files, timeout metadata and the runner source loaded when this
legacy group started. The Python traceback displayed a line from the subsequently
updated on-disk runner; loaded-runner.py preserves the actual source version.

The smaller semver piece completed independently. Four of 81 files are paired;
closure totals and a full top-ten IR ranking remain unavailable. All completed
raw binary runs were committed and plainly pushed; their paired archives retain
the exact raw streams and timings. Both entry programs are checker-clean, with
zero diagnostics. The two own apply products have identical bytes in all 746
src files. The source upper bound is complete.

The original whole-group timer was superseded by separately checkpointing each
binary run within 23 minutes. Its outer controller was stopped, while its tracked
baseline child retained the original deadline. That child timed out and its
runner exited; no measurement process is left running.

To resume, first prepare finer file or unit scheduling for the timed-out group
and prove its scheduling-only behavior and meter identity. Check whether an
individual giant enclosing unit also exceeds the budget. Do not treat the
unfinished baseline or the four completed files as an 81-file total.

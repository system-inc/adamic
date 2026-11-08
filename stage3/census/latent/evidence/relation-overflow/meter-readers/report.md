# Meter relation-report readers: task 8dbaea5

The compiler-directory meter stream already used the latent driver and its
trailer. The entry overlay replaced the driver and lost the returned report;
entry.go.txt now retains it and relabels its findings. report.py validates and
separates the trailer, counts its program findings in ordinary NotYet totals,
and explicitly reports a missing trailer as an incomplete run. Root-order
credit is blocked by program-level relation findings too.

All final checks passed:

| Check | Child exit |
| --- | ---: |
| cloud/setup.sh | 0 |
| npm ci --prefix stage3/api | 0 |
| go test -count=1 -v ./stage3/meter (all five wrappers) | 0 |
| All meter Python suites with the three real binaries (70 tests, no skips) | 0 |
| go test -count=1 -json ./stage3/census/... (whole pattern) | 0 |
| Complete latent suite with the measurement overlay | 0 |
| Entry, latent and ordinary census builds | 0 each |
| Directory census basic audit | 0 |
| Changed split/closure reader fixture checks | 0 |
| Trailer-as-file mutant | 1, KeyError: file |
| Dropped-trailer-findings mutant | 1, expected NotYet 1, observed 0 |
| Mutant-result assertions | 0 |

The final four trailer tests cover file counts, findings/reasons/events,
missing-trailer output, and malformed trailer placement/endpoints. The separate
root-order regression proves that an unlocated program relation finding cannot
be discarded when awarding lowering credit. Existing real entry probes now
assert a relabeled trailer with both explicit endpoint answers.

An extra initial audit invocation used the entry binary with a directory-only
audit: its child exited 2 and the audit wrapper exited 1. This was an invocation
error, corrected by running the same audit on the latent binary (exit 0). Its
log is retained separately, not counted as a successful entry probe.

See readers.md for every git-grep reader match, commands.md for reproduction,
validation.json for child exits, and the adjacent logs for assertions. No new
Go tests were added. No native output or full compiler gate claim is made.

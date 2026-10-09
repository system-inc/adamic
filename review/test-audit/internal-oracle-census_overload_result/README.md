Unit u056, origin 68db8ddd145281a62655452496bdc32ef848bdf3.

REPORT.md is the deliverable. rows.json, matrix.json, manifest.json, diagnostics.json and timings.json are machine-readable evidence. Test stdout/stderr are log files. Every standalone diff in diffs/ applies independently to this origin commit and passed Go vet; standalone-validation.json records each command. No audited test, oracle or fixture was edited.

Replay each diff in a clean checkout of the origin commit. Source /workspace/adamic-tools/env.sh and use ADAMIC_GATE_UNCACHED=1 plus ADAMIC_BUILD_CACHE_DIR=/tmp/u056/cache/<mutant>. Run the bounded commands recorded in matrix-commands.json. Standalone diffs do not need ADAMIC_MUTANT. The family command runs two fixture members only. For switched replay, apply switched-source.diff and copy switched-helper.go.txt to internal/lower/audit_u056.go; select with ADAMIC_MUTANT.

P01 returns nil from Lower. P02 returns an empty IR program; it is the empty-answer probe used for vacuity. Neither counts as a production mutant. The independent M09 recognition adapter is saved separately and is only an observation probe. Copy recognition-adapter.go.txt to internal/lower/audit_u056_probe.go and recognition-main.go.txt to cmd/u056-audit-probe/main.go in the switched checkout, build the probe using survivor-commands.json, and remove those scratch files afterwards.

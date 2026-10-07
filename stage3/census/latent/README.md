# Latent census tooling

This is a partial measurement tool, compiled using a scratch Go overlay. Its
binary disables `lower.Lower` and exposes only a finding-returning measurement
API. No production compiler file is edited. The ordinary Go driver deliberately
panics unless replaced through the overlay.

For each accepted project, the overlay walks every refusal-scan node, recording
location, reason and complete diagnostic text instead of stopping at the first
refusal. It then attempts each top-level function and statement with fresh
lowering state. A function's lowering error is recorded and the next unit runs.
Findings are deduplicated by kind, location, reason and diagnostic text.

Limits: checker rejection still blocks the project. The driver checks all source
roots together, then measures individual files and units; checker counts are
unique diagnostic locations, not repeated per-entry runs. Lowering still stops
at the first error within each unit. Fresh state includes the unit's declaration,
not sibling globals or function signatures. Consequently isolation errors can
reflect missing sibling registration, and these counts must not be treated as
exhaustive production blockers. Generic declarations are attempted without
inventing concrete instantiations. Module order, ownership, specialization and
backend correctness are not measured. Panics and ordinary errors are recorded
separately from NotYet and Refused. The corpus run in REPORT.md reached none of
this lowering code, because the checker rejected the adapted project.

Build and prove the overlay, from the Adamic root:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/latent-overlay > /tmp/latent-overlay.log 2>&1
gofmt -w /tmp/latent-overlay/*.go
go build -buildvcs=false -overlay=/tmp/latent-overlay/overlay.json -o /tmp/latent-census ./stage3/census/latent/tool > /tmp/latent-build.log 2>&1
python3 stage3/census/latent/audit.py /tmp/latent-census > /tmp/latent-audit.log 2>&1
/tmp/latent-census /path/to/adapted/src/compiler /tmp/latent.jsonl > /tmp/latent-run.log 2>&1
```

`run_comparisons.py REPOSITORY ADAPTED NEW_SCRATCH [BRANCH_PREFIX]` creates five isolated,
never-pushed scratch branches: baseline and each named feature individually.
It imports the base pipeline, original census, and adaptations 10 and 20 before
building. It stops a configuration on merge or build failure. It uses the
repository's initialized cohere checkout through a scratch symlink; the measured
branches here all pin that same dependency. VCS stamping is disabled because
Git does not recognize that symlink as an initialized worktree submodule.
Use a new scratch directory and branch prefix for a later run; the fixed
`scratch/latent-compare-*` names intentionally fail if already present.

`summarize.py SCRATCH ADAPTED` records per-file and per-reason observed counts,
source hashes, branch SHAs, and unknown feature-lowering deltas in REPORT.json.
Compressed JSONL retains every raw diagnostic. A blocked file has null findings;
that is different from an accepted file with an empty finding list.

`LATENT_MUTANT_FUNCTION` plants an extra NotYet at the selected function only in
the overlay. The audit compares two source files and requires a delta of exactly
one in the selected file and zero in the other. It also requires both failing
functions and both refusal sites to be observed, and proves the binary cannot
return output IR. These synthetic probes are generated as scratch `.a` files;
they are not TypeScript-derived feature fixtures.

Validated merged implementation: c4d6d6c693fa7502cfa191099957f08bf0d97003.
Merged origin/main: e8ba3d5d81de4d3773c723914fccd4c76248b965, using merge commits, without rebase.
Green: CLI, loader, lowerer, entire uncached Node oracle, regenerated counts, vet and formatting.
Mutants: all 19 in mutants.py were caught again on the merged implementation.
Limitations: full gate stopped in stage-1 packages; native tsc and cycle-worker integration remain unvalidated.

# Branch boundaries

Only codex/multi-root is pushed. No push to main or any area/ branch, and no
merge into either. Integration owns main and area/compiler. The final response
reports the pushed feature tip, which additionally contains this validation report.

The first merge incorporated e011f8f. During its gate, main advanced with call
target routing and devirtualization changes. That superseded run was stopped,
and e8ba3d5 was merged into the feature branch. Both merges were clean; no runtime
implementation edits or count corrections were needed to resolve them.

# Re-green

All commands sourced /workspace/adamic-tools/env.sh and wrote test output to logs.

Final scoped gate:
`ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/load ./internal/lower ./internal/oracle -count=1 -timeout 30m`.

Exit 0. CLI 1.086s, loader 1.796s, lowerer 22.403s, entire oracle package 87.141s.
This runs all Node oracle comparisons, including main's new borrowed-element,
call-target and devirtualization fixtures. Log: /tmp/adamic-land-scoped-final.log.

`go vet ./...` and `gofmt -l cmd internal` passed without findings.
Logs: /tmp/adamic-land-vet-final.log and /tmp/adamic-land-format-final.log.
`git diff --check` also passed.

The broader `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...` was
attempted after the second merge. It passed the checker bridge, CLI tools, flow,
freshness, IR, loader, lowerer, native backend, entire oracle package and regexp.
It was stopped after roughly nine minutes while unrelated stage-1 suites were
still running, using the unit's expressly permitted scoped fallback. That run
is not a full-gate pass. Its log is /tmp/adamic-land-gate-final.log.
The final scoped gate above provides the complete successful exit.

# Counts

`go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`
passed in 39.698s, producing no diff against the merged table.
Log: /tmp/adamic-land-counts-final.log.

Relative to the merged main, the branch adds only these two rows:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| multi_root/tsconfig.json | 4 | 4 | 2 | 8 | 1 | 0 |
| project_entry/tsconfig.json | 0 | 0 | 0 | 0 | 0 | 0 |

The first row measures the explicit three-entry API, retaining its original
order and shared-module behavior; the config path identifies the fixture.
The second measures only the selected project entry. Its unimported configured
source is checked but never executed. Existing main rows, including its changed
borrowed-element and devirtualization counts, are preserved exactly.

# Mutants

`python3 stage3/multi-root/mutants.py /tmp/adamic-land-mutants-final` exited 0,
meaning each of its 19 selected test runs exited nonzero with an actual failing
test and without a Go build failure. Summary: /tmp/adamic-land-mutants-final.log;
individual overlay files and failing logs: /tmp/adamic-land-mutants-final.

Caught names: config-roots-as-entries, unchecked-unimported-config-files,
outside-config-entry, drop-direct-declarations, declaration-only-guard,
missing-adamic-glob-aliases, unchecked-javascript, wrong-root-order,
repeated-shared-modules, weakened-required-options, enabled-checker-bypass,
wrong-module, wrong-module-detection, wrong-module-resolution, old-target,
wrong-class-fields, unsound-project-json, ambiguous-source-alias,
missing-fallback-console.

mutants.py records the exact corresponding test for each mutant.
REPORT.md and ENTRY_REPORT.md explain their observed failures. In particular,
config-roots-as-entries is caught by native output disagreeing with Node on
the unimported side-effect source, rather than by compilation failure.

# Remaining boundaries

The tsc smoke observations and module-cycle integration boundaries in
ENTRY_REPORT.md remain unchanged. The upstream tsc config is refused for project
references, and the compiler config is refused for weakened strictBindCallApply.
The cycle worker's module traversal and cross-module linking have not been merged
into this branch. No claim of native tsc, full stage-1 validation, or the combined
cycle implementation is made.

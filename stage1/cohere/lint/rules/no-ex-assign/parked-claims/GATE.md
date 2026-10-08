# All-input checker audit gate

Branch: codex/typeaware-wave-01-checker, from area c4bdc23fa.
No executable rule changes were made: all 12 claimed rules stopped at the
exact missing questions/helpers documented in the sibling notes. Each has
0 newly certified upstream cases and 0 new mutant executions. The successful
area package gate does not certify those blocked rules.

Command after sourcing /workspace/adamic-tools/env.sh:
GOMAXPROCS=4 GOFLAGS=-buildvcs=false GOPROXY=https://proxy.golang.org\|direct
ADAMIC_LINT_BENCH=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-01-typescript
ADAMIC_LINT_PROFILE_DIR=/workspace/unpark-wave-checker-gate/profile
ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/unpark-wave-checker-gate/profile
go test -json -count=1 -timeout=60m ./stage1/cohere/lint

The profile directory was fresh. The TypeScript checkout was clean before
and after, at 050880ce59e30b356b686bd3144efe24f875ebc8. Every optional input
was supplied. Native, Node and emitted-JavaScript comparisons were retained.

Result: exit 0; top-level 34 pass, 0 fail, 1 skip.
Wall time: 1864.483594 seconds; nproc 5, quota 4 cores.
One-minute load min/median/max: 1.000000 / 1.545410 / 8.588867.
Full samples and input settings: evidence/result.json.
Only skip: TestCheckerBridgeRefusalPending, checker_pending_test.go:49,
awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from
the C error buffer. No guard, skip or expected output was changed.

All 75 registered area mutants were caught. Their raw log lines are in
mutant-lines.txt and lint.jsonl. They are inherited checks, not mutants for
the blocked claims. The complete package includes compiler/repository corpus,
profile, throughput, shards, owned witnesses, options and checker replay/cache
controls. The repository-wide gate is outside this report.

Setup finished in 194.701 seconds. An initial compile-only probe failed
because the restarted environment lacked /tmp/adamic-gate; it was created
with mode 1777 before the all-input run. Obsolete owned scratch ELF files
and archives were removed to recover disk space while retaining all source
and logs. No private checker, question implementation or shared helper copy
was built. Only owned rule-directory notes/evidence are committed.

Landing first: both previous local branches were merged with area, without
rebasing. no-ex-assign is already on area. The distinct typeaware-wave-01
merge remains local and is not newly published or represented as certified
by this fresh area-based review-branch run.

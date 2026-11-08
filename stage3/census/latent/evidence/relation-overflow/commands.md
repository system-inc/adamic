# Relation overflow validation commands

Base: `6f68dd0fc5ee78c0b1230482d2c2089f02b1856b` (origin/main fetched at task start).
Cost-measurement checkpoint: `726a71afdc5d9fdcd5a32d6791d7814c19c043bf`.
Final pre-run checkpoint: `b2bbfc3f56167c7c1ecafbe8110da154f7a67cb1`, pushed to
`typescript/latent-relation-overflow-checkpoint` before the tsc run.

Every shell wrapper saves logs, captures `$?` immediately, prints that child's
exit code, and exits with it. No command output is piped into head or tail.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/latent-setup.log 2>&1
npm ci --prefix stage3/api > /tmp/latent-npm.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/latent-typescript
# /tmp/latent-typescript is checked out at the stage3/source.json pin:
# 050880ce59e30b356b686bd3144efe24f875ebc8
bash stage3/apply.sh /tmp/latent-adapted > /tmp/latent-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/latent-overflow-endpoint-overlay > /tmp/latent-final-overlay.log 2>&1
go test -count=1 -v -overlay=/tmp/latent-overflow-endpoint-overlay/overlay.json ./stage3/census/latent/... > /tmp/latent-overlay-tests.log 2>&1
go test -count=1 -json ./stage3/census/... > /tmp/latent-census-final-tests.jsonl 2> /tmp/latent-census-final-tests.stderr
go build -buildvcs=false -overlay=/tmp/latent-overflow-endpoint-overlay/overlay.json -o /tmp/latent-endpoint-census ./stage3/census/latent/tool > /tmp/latent-final-build.log 2>&1
go test -count=1 ./cmd/adamic-gate/parallelcensus -run '^TestEveryTest(HasOneDescriptiveName|IsParallelOrSaysWhy)$' > /tmp/latent-gate-checks.log 2>&1
python3 stage3/census/latent/audit.py /tmp/latent-endpoint-census > /tmp/latent-basic-audit.log 2>&1
LATENT_FULL=0 LATENT_SPECULATIVE=0 LATENT_RELATION_PER_UNIT=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/tmp/latent-endpoint-tsc-units.jsonl LATENT_RELATION_TIMES=/tmp/latent-endpoint-tsc-relations.jsonl GOMEMLIMIT=8GiB python3 stage3/census/speculative/timed_run.py 1800 10240 /tmp/latent-endpoint-tsc-metrics.json /tmp/latent-endpoint-tsc.log /tmp/latent-endpoint-census /tmp/latent-adapted/src/tsc/tsc.ts /tmp/latent-endpoint-tsc.jsonl > /tmp/latent-endpoint-tsc-wrapper.log 2>&1
```

`package-exits.json` records separate complete `go test -count=1 PACKAGE`
invocations for every package returned by `go list ./stage3/census/...`, in
addition to the successful whole-pattern invocation. The overlay suite includes
all latent packages and the tool's overlay-only in-process tests.

For the removed-call mutant, copy the generated overlay JSON and replace only
the generated `internal/lower/latent_relation_overflow.go` with a scratch copy
where `answer := program.RelationOverflow(typeChecker)` becomes `answer := ""`.
Keep the final overlay's other mappings, including its tool tests, unchanged.
Run:

```sh
go test -count=1 -v -overlay=/tmp/latent-overflow-mutant-overlay.json ./stage3/census/latent/tool -run '^TestLatentCensusOverflowMarkReachesProgramAndEveryAttempt$' > /tmp/latent-mutant.log 2>&1
```

The child exits 1 with `planted mark disappeared`. The unchanged test exits 0
and logs the planted program and unit JSONL findings.

The optional `full_audit.py` exits 1 on both this change and the unchanged base
census. It still expects string and number conditions to be refused; both now
produce only `var`. To isolate the change, a second scratch overlay used the
base's `lower.go.txt`, `load.go.txt`, `units.go.txt`, `speculative.go.txt`, and
`main.go.txt`, retained the overlay generator's enum registration and import
cleaning, and omitted the new overflow helper and proof test mappings. The base
binary builds with exit 0 and has the same audit failure. For each of the full,
first-error-mutant and rollback-mutant runs, removing the new overflow fields
and normalizing source-directory paths yields byte-identical original
observations. The hashes are in `full-audit-baseline-comparison.json`.

The initial cost experiment used the per-unit checkpoint on this same entry.
It completed checker.ts but was deliberately stopped before the run end because
query time was excessive. Its supervised process received SIGTERM (exit -15);
the existing timed_run.py wrapper exposed that as shell exit 241. Both values
are retained in per-unit-cost-summary.json. It is not a completed census and
provides no end-of-run overflow conclusion. To reproduce that query mode with
the final driver, set LATENT_RELATION_PER_UNIT=1 explicitly.

Code checkpoint on the requested branch: `e6bfbce5db390af989f4a3c0aec7e469fab0d6f7`.
The final endpoint attempt exited 124 at its 1800-second bound before the end
query. A follow-up run needs a longer wall allowance. The absence of an end
answer is not interpreted as an empty answer.

Built: re-cut task #3zvr9mq's four-commit census generic-deferral delta onto current main as one commit.
Commits: source d2d64b35..76057b92 (7cdb389f, fe6481d5, bcab1e1d, 76057b92); delivery parent bd90a433bf636b8325a591ab3e7487d457a69ab4; delivery SHA is in the final response.
Commands and outputs: six oracle leaves PASS, five Linux counts rows regenerated, census audits PASS, reader checks and vet PASS.
Mutants: original runtime, premature-lowering, inventory, report, byte-coverage, snapshot, identity and isolation mutants caught; explicit no-caller false-witness mutant caught.
Not covered: whole packages, full gate, whole-compiler census, new readiness totals, all concrete specializations, production instantiation refusals, PR or Loom submission.

The net change was applied with `git diff --binary --full-index d2d64b35 76057b92` and `git apply -3`, excluding `internal/oracle/counts.md`. The first diff omitted binary payloads, so apply rejected it atomically; the binary diff then applied without conflicts. No source branch was merged. The delivery is one commit, with the required `Supersedes Loom chg_1cnr73q8ygf84bwa0wr6wfrfw5.` line.

**TypeScript census hunks changed: none.** All 49 `stage3/census` paths in the source delta are byte-for-byte equal to `76057b92`; `typescript-hunks.json` records the check. Unrelated files already added on current main remain intact. Historical source evidence is retained as originally delivered, and this run's evidence is under `review/compiler/loom-census-deferral`.

Main forced one compatibility hunk outside TypeScript's census files: `internal/oracle/instantiation_proof_test.go` now registers five `countCase` values, each preparing the current `(runs, measure)` API, instead of appending the obsolete `func(*testing.T) []string` counter. Fixture programs and oracle comparisons remain unchanged. This is also why the old counts hunk was not carried.

I read `1f07ce2d` before applying the change, including its counts-product migration. Main's counts writer documents partial refresh: under `-update-counts`, it replaces measured rows and preserves other rows. Linux's house clang is 20.1.8. The five rows were measured again through that writer, without copying them from the source diff. They happen to equal the old numerical observations:

| Fixture suffix | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| find | 13 | 13 | 28 | 31 | 7 | 0 |
| first | 13 | 13 | 8 | 20 | 6 | 0 |
| last | 9 | 9 | 9 | 15 | 7 | 0 |
| callback | 10 | 10 | 15 | 23 | 4 | 0 |
| array | 10 | 10 | 1 | 10 | 6 | 0 |

Only those five rows were added to counts.md. No digest column, toolchain heading, existing count, or predicate row changed. `TestCountsAreRecordedTable` passes. These are standalone oracle/proof leaves, not entries in `internal/lower`'s native-agreement row table; its registration test was run anyway and passes.

Setup passed on its first invocation in this unit:

```text
export GOPROXY='https://proxy.golang.org|direct'
timeout 300 bash cloud/setup.sh
Node ready 0.022s; Go ready 0.031s; Markdown dependencies ready 0.075s;
submodules ready 0.114s; clang ready 0.275s; shared cache ready 1.738s;
go build ready 46.443s; test binaries deferred 46.554s;
build cache warm 46.556s; done 46.587s.
nproc=5; cpu.max=400000 100000 (4 CPU quota).
Go 1.27.1; clang 20.1.8; Node v24.19.0.
source /workspace/adamic-tools/env.sh
```

The checked-in `stage3/api` npm lock was installed with `timeout 90 npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund`. Stock AST audits used TypeScript 6.0.3 via `NODE_PATH=/workspace/adamic/stage3/api/node_modules`.

Commands run, with every test's output redirected directly to a retained log:

```sh
# Prepare/build only, then execute the six separate oracle leaves.
timeout 120 go test -c -o /tmp/loom-census-oracle.test ./internal/oracle
# From internal/oracle:
ADAMIC_GATE_UNCACHED=1 timeout 90 /tmp/loom-census-oracle.test \
  -test.run '^TestInstantiation' -test.timeout 90s -test.parallel 4 -test.v
# PASS: Find 3.29s, First 0.48s, Last 3.27s, Callback 3.25s,
# Array 0.43s, PresenceMutant 3.17s. Each includes its setup and is below 60s.
# Source Node, native sanitizers, JS backend, release native, leak checks.

# From internal/oracle, current main's update writer, only the five new rows:
timeout 90 /tmp/loom-census-oracle.test \
  -test.run '^TestCountsAreRecorded_[0-9]+$/internal/oracle/testdata/instantiation_' \
  -test.timeout 90s -test.parallel 4 -test.v -update-counts
# PASS, five rows added.

timeout 90 go test ./internal/oracle -run '^TestCountsAreRecordedTable$' \
  -timeout 60s -count=1 -v
# PASS, package 0.005s.

timeout 90 python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/loom-census-overlay
timeout 90 go build -overlay=/tmp/loom-census-overlay/overlay.json \
  -o /tmp/loom-census ./stage3/census/latent/tool
# Both exit 0; no TypeScript census adaptation needed.

timeout 90 python3 stage3/census/latent/audit_deferred.py /tmp/loom-census \
  review/compiler/loom-census-deferral/deferred
# PASS all three modes, concrete later-file caller, unused/diagnosed generics,
# nested reached/never-reached bodies, independent attempted counts and byte masks.
timeout 60 python3 stage3/census/latent/audit.py /tmp/loom-census
# PASS continuation, body/signature eligibility, source attribution, no-output guards.
timeout 60 python3 stage3/census/speculative/audit.py /tmp/loom-census \
  review/compiler/loom-census-deferral/speculative
# PASS exact depths 0/1/2, sibling depth 0, no-stubs exact full JSON, AST visitation.

timeout 30 node stage3/census/speculative/stock.cjs \
  review/compiler/loom-census-deferral/deferred/source \
  review/compiler/loom-census-deferral/deferred/speculative.jsonl \
  review/compiler/loom-census-deferral/deferred/stock.json
timeout 30 python3 stage3/census/speculative/report.py \
  /workspace/adamic/review/compiler/loom-census-deferral/deferred/source \
  review/compiler/loom-census-deferral/deferred/speculative.jsonl \
  review/compiler/loom-census-deferral/deferred/stock.json \
  review/compiler/loom-census-deferral/deferred/report
timeout 30 python3 stage3/census/speculative/verify.py \
  review/compiler/loom-census-deferral/deferred/speculative.jsonl \
  review/compiler/loom-census-deferral/deferred/report/RESULT.json \
  /workspace/adamic/review/compiler/loom-census-deferral/deferred/source
# PASS: 20 independently matched source sites, all six artifact mutants caught.

timeout 120 python3 stage3/census/speculative/prove_identity.py /workspace/adamic \
  /tmp/loom-census-overlay /tmp/loom-census review/compiler/loom-census-deferral/identity
# PASS exact AST identity and token-position mutant.

timeout 300 python3 review/compiler/loom-census-deferral/build-mutants.py
timeout 300 python3 review/compiler/loom-census-deferral/run-mutants.py
# Each build/run also has its own 90s/60s child limit.
# PASS no-caller, plain-map, missing-inventory and premature-lowering assertion kills.

timeout 60 python3 review/compiler/loom-census-deferral/full-report-control.py
# PASS independent stock spans, original full-report count/gap audits,
# missing no-caller-ledger mutant and rerun accounting.

timeout 120 python3 review/compiler/loom-census-deferral/production-identity.py \
  /workspace/adamic review/compiler/loom-census-deferral/production-identity bd90a433
# PASS: main, ordinary compiler and flag-on production have byte-identical output:
# C 6907 bytes, JS 9930 bytes. Both leaked-output artifact mutants caught.
# This is the original isolation.py with only its diff preflight adjusted to
# exclude new internal/oracle tests/data/counts; no production source was excluded.

timeout 120 go test ./stage3/census/latent/statecopy ./stage3/census/latent/statementrewrite \
  ./stage3/census/latent/refusalrewrite \
  -run '^(TestUnknownMutableContainerFailsLoudly|TestUnknownForeignPointerFailsLoudly|TestPublishedCopyModeIsPrivate|TestMissingStatementFailsWithMethodName|TestCompiler41231d51Shape|TestMissingFunctionMutantFailsLoudly|TestChangedVisitorMutantFailsLoudly|TestVisitorsCollectContinueAndSkipDiagnosedBodies|TestFullOwnerGuardRequiresFunctionDeclaration|TestRewriteAcceptsTwoOuterRefusalReturns)$' \
  -timeout 90s -count=1 -v
# PASS: statecopy 3.874s, statementrewrite 3.828s, refusalrewrite 3.297s.
# Ten selected leaves; longest leaf 3.83s, including tool-product setup.

timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s -count=1
# PASS, package 34.413s.
timeout 90 go test ./internal/lower -run '^TestNativeAgreementRowsNameRegisteredPrograms$' \
  -timeout 90s -count=1 -v
# PASS, leaf 0.01s, package 0.021s.
timeout 90 go vet ./internal/oracle ./stage3/census/...
timeout 90 go vet -overlay=/tmp/loom-census-overlay/overlay.json \
  ./internal/lower ./internal/load ./stage3/census/latent/tool
# Both exit 0, no diagnostics.

timeout 90 go test -p 2 -run '^$' -timeout 90s \
  ./internal/oracle ./stage3/fixtures ./stage3/census/...
# PASS: all 11 affected tooling/harness packages compile. This is a filtered
# compilation check, not a claim that their whole behavior suites ran.
```

Every carried mutant family was exercised:

| Mutant | Catcher observed in this run |
|---|---|
| Erase one present generic object return | Original TestInstantiationPresenceMutant: clean native and JS stdout disagreements with Node |
| Premature independent generic lowering | Original census-check.py: expected deferred unit has status attempted |
| Remove firstOrUndefined inventory | Original census-check.py: missing-inventory assertion |
| Claim a concrete body witness with no caller | audit_deferred.py: unused generic must not be instantiated |
| Missing never-reached lists, legacy/full/speculative | Independent deferred-coverage ledger audit, all three modes |
| Count a deferred body as attempted, all three modes | Independent status-membership recount |
| Credit deferred bytes as examined, legacy/full | Independent byte-mask audit, despite balanced aggregate subtraction |
| Missing full-report no-caller list | full_report.audit_counts: deferred coverage gap |
| Speculative headline depth, top-20 count, missing file, false examined-byte share | speculative/verify.py's four original artifact checks |
| Speculative missing never-reached list and deferred-as-examined | speculative/verify.py's two deferred artifact checks |
| Restore plain-map optional-parameter snapshot | Successfully built mutant fails original speculative depth/continuation audit |
| No-stubs/first-error continuation and zeroed depths | speculative/audit.py exact nested-depth and continuation controls |
| Attribute by token position | Original prove_identity.py: async root wrongly depth 1 while await remains depth 1 |
| Leak stub text into C or JS | Independent production-output byte comparisons, both backends |
| Include signatures in body eligibility | latent/audit.py signature/body assertion |
| Misattribute an extra NotYet to sibling file | latent/audit.py unchanged-sibling equality |
| Unknown mutable/foreign state, missing statement/function, changed visitor | Original focused statecopy/statementrewrite/refusalrewrite guard tests |

`build-mutants.py` creates independent overlays and saves mutant source only as `.go.txt`. The original `premature-lowering-mutant.patch` remains unmodified. The original full-report helper initially received an empty owner table in my scratch control; it correctly rejected that mismatch. The control now loads the real stage3/meter/owners.json, as TypeScript's report builder does, and passes. No census source changed for this correction.

Conservative scope: deferral is inventory/coverage accounting, not compiler admission or proof of every specialization. The original restriction to generic function declarations and nested dependent declaration units is retained exactly. Unused deferred bodies remain coverage gaps. Production compiler sources and behavior did not change. No code was copied from cohere; baseline builds reference its existing submodule.

The source delta contains 19 existing whitespace warnings in two historical evidence artifacts (`review/compiler/instantiation-proof/premature-lowering-mutant.patch` and `stage3/census/latent/deferred-generics/speculative-controls.log.txt`). They were preserved to keep the original mutant and all TypeScript files exact; they are not changes to executable Go formatting. The staged whitespace check also repeats six warnings from verbatim new apply/audit logs, for 25 total diagnostics. No PR was opened or Loom change submitted.

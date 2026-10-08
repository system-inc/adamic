# Lowering NotYet measurement evidence

Built a table of 10,426 unique lowering NotYet sites, with exact reasons, per-file counts, examples, context flags and dispositions.
Compiler base is 44583d3283fdd8674085a7ddcce040cf2a73a94e; the clean replay merge is 58c687c676554cd2de17d9a15f5edb72ea162264.
The full guarded entry census exited 0; accounting, AST-ledger validation and the top-reason replay all passed.
Continuation, rollback, no-output, replay-selection and 14 accounting/ledger corruption mutants were caught; details follow.
Checked non-null was skipped under the authorized fallback; census entries' backend semantics and a full repository gate are not covered.

## Base and inputs

Resolved the newest requested refusal-table branch to `44583d32`. The required merge of `c17bf216` produced 38 conflicts and was aborted after the user's correction. Cherry-picking only `c17bf216` produced eight conflicts, named in TABLE.md. The base lacks the `--explain-checks` command path and predicate-site API on which that commit's report controls depend. Resolving only the changed hunks would not supply those dependencies. The user's under-20-minute fallback was used rather than introducing dependency compiler work. Both operations were aborted; no hunks were resolved or retained. `8a364b4c3ae356a79a8233fa1c4b9dd6a0f54c00` was fetched through the updated branch and inspected, but not cherry-picked after the fallback.

Merged `9a1f14c5d994aa855625e7cfa295677060348fec` cleanly. This brings the full-census implementation and replay command from the requested lineage. `git diff --name-only 44583d32 -- internal cmd cohere` was empty: the production compiler, tests, counts and submodule pin are identical to the measured base. No compiler work was introduced. [metadata.json](metadata.json) records these pins.

Prepared the input using the previous latent-full branch's own adaptation snapshot:

```sh
git archive origin/codex/stage3-latent-full stage3 | tar -x -C /tmp/stage3-notyet-input-snapshot
source /workspace/adamic-tools/env.sh
bash /tmp/stage3-notyet-input-snapshot/stage3/apply.sh /tmp/stage3-notyet-adapted > /tmp/stage3-notyet-apply.log 2>&1
```

Apply exited 0. All 81 source paths, byte lengths and SHA-256 hashes match `stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json`; [source-manifest.json](source-manifest.json) preserves the new manifest. The saved stock TypeScript 6.0.3 declaration ledger independently validates coverage, spans, parents and direct diagnostic ownership on these identical bytes.

This compiler observed 260 checker diagnostics, 10,560 attempted units and 138 `split_checker_body` units. Counts are **measured on a checker-rejected entry-root program**. The raw stream retains 16,602 Boundary, 1,054 Refused, 51 panic, 27 error and 14 SkippedDependency lowering observations separately; these are raw observations, not deduplicated totals, and none enters the NotYet table.

## Setup

Ran `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`, writing output to a log. The first run failed because the pending merge left markers in `internal/lower/census_panics_test.go` and `internal/oracle/nested_functions_test.go`: both reported `5:1: missing import path`. Its completed timings were Node 0.327s, Go 0.355s, clang 0.953s, markdown dependencies 1.936s, submodules 22.947s; `nproc` printed 5. [Initial setup output](evidence/setup-initial.log.txt).

After aborting the conflicting operations and merging replay, setup exited 0. Timings: Node 0.025s, Go 0.030s, markdown dependency validation step 0.012s and ready 0.083s, submodules 0.085s, clang 0.179s, Go build 187.344s, test binaries deferred 187.443s, build cache warm 187.445s, done 187.482s. `nproc=5`, cgroup quota `400000 100000`; Go 1.27.1, clang 20.1.8 and Node v24.19.0. Sourced the printed `/workspace/adamic-tools/env.sh`. [Setup output](evidence/setup.log.txt).

The initial entry census stopped with `load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node`. `npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund` installed the lockfile-pinned dependencies, exited 0, and resolved it. No tracked manifest or lockfile changed. [Dependency output](evidence/node-types.log.txt).

## Census and table commands

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/stage3-notyet-overlay > /tmp/stage3-notyet-overlay.log 2>&1
python3 stage3/meter/entry_overlay.py /tmp/stage3-notyet-overlay /tmp/stage3-notyet-entry-overlay > /tmp/stage3-notyet-entry-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/stage3-notyet-entry-overlay/overlay.json -o /tmp/stage3-notyet-census ./stage3/census/latent/tool > /tmp/stage3-notyet-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/stage3-notyet-census /tmp/stage3-notyet-adapted/src/tsc/tsc.ts /tmp/stage3-notyet-full.jsonl > /tmp/stage3-notyet-census.log 2>&1
python3 stage3/notyet-table/build-table.py /tmp/stage3-notyet-full.jsonl /tmp/stage3-notyet-adapted stage3/notyet-table > /tmp/stage3-notyet-table-build.log 2>&1
python3 stage3/notyet-table/audit.py stage3/notyet-table /tmp/stage3-notyet-adapted > /tmp/stage3-notyet-accounting.log 2>&1
```

All exited 0. The census wrote one header and 81 source records. [Census log](evidence/census.log.txt), [accounting log](evidence/accounting.log.txt), [raw CSV](raw.csv), [unfiltered JSONL](full.jsonl.gz), [summary](summary.json) and [table](TABLE.md) preserve the observations and derived artifacts. CSV contains all 14,704 matching lowering NotYet observations, including repeated attempts; the table deduplicates `(kind, where, reason, text)` to 10,426 sites. Context sensitivity is exactly the previous report's `reason.startswith('reading ')` flag.

Coverage: top 5 = 4,335/10,426 (41.58%); top 10 = 5,258/10,426 (50.43%); top 20 = 6,252/10,426 (59.97%). There are 1,488 reasons, 1,292 with fewer than five sites, and 4,323 context-sensitive sites.

Disposition is an inference, separate from observed counts. The supplied non-null, logical-assignment and comma rulings override older documentation. Writable variance is an adaptation: the observed `a writable-slot checked view requiring source contract certification` comes from `cast_proof.go` rejecting a `widened(target, source)` relation. Explicit any storage/calls/returns, debugger and void-operator observations are adaptations under the language's stated design refusals. Other NotYet observations are conservatively compiler lessons, including isolated reads that still need individual interpretation. A void-returning call used as a value is a representation lesson, distinct from the refused void operator.

## Top non-context-sensitive replay

The first-ranked reason, with 1,761 sites, is the structural-signature method-call limitation. Replayed its first table example using the complete tsc project:

```sh
source /workspace/adamic-tools/env.sh
go run ./stage3/census/latent/replay -project /tmp/stage3-notyet-adapted/src/tsc/tsc.ts -where /tmp/stage3-notyet-adapted/src/compiler/binder.ts:579:18 -kind NotYet -reason 'a method call through a structural signature in a program with statics; use typeof the declaring class' > /tmp/stage3-notyet-replay-candidate.json 2> /tmp/stage3-notyet-replay-candidate.log
```

Exit 0. Stderr:

```text
replay load/register/lower: 6.968941302s
reproduced NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class at /tmp/stage3-notyet-adapted/src/compiler/binder.ts:579:18
replay total (including overlay and Go build): 18.081668236s
```

[Exact argument array](evidence/replay-command.json), [stdout record](evidence/replay.json), [stderr](evidence/replay.log.txt). The selected unit is `src/compiler/binder.ts:571:5`. Its ten lowering findings equal the full census's selected-unit findings in order, excluding only the intentionally different entry/replay measurement labels. `python3 stage3/notyet-table/verify-replay.py > /tmp/stage3-notyet-replay-accounting.log 2>&1` exited 0. [Verification output](evidence/replay-accounting.log.txt).

Running the same real-project replay command with `LATENT_REPLAY_MUTANT_PARENT_SIBLING=1` exited 1. It selected `binder.ts:330:1` and printed `replay signature did not reproduce` for the requested exact signature. [Mutant stdout](evidence/replay-real-mutant.json) and [stderr](evidence/replay-real-mutant.log.txt). This is a signature failure, not a compiler build failure.

## Focused validation and mutants

Every test's output was written directly to a log; no test output was piped. Commands:

```sh
go build -buildvcs=false -overlay=/tmp/stage3-notyet-overlay/overlay.json -o /tmp/stage3-notyet-directory-census ./stage3/census/latent/tool > /tmp/stage3-notyet-directory-build.log 2>&1
python3 stage3/notyet-table/validate-tooling.py full /tmp/stage3-notyet-directory-census /tmp/stage3-notyet-audit-final > /tmp/stage3-notyet-full-audit-final.log 2>&1
python3 stage3/notyet-table/validate-tooling.py replay > /tmp/stage3-notyet-replay-tests-final.log 2>&1
LATENT_REPLAY_MUTANT_PARENT_SIBLING=1 python3 stage3/notyet-table/validate-tooling.py replay ReplayTests.test_nested_findings_match_full_census_in_order > /tmp/stage3-notyet-replay-mutant-final.log 2>&1
python3 stage3/census/latent/audit.py /tmp/stage3-notyet-directory-census > /tmp/stage3-notyet-legacy-audit.log 2>&1
python3 stage3/census/latent/audit_output_guards.py "$PWD" /tmp/stage3-notyet-overlay /tmp/stage3-notyet-output-guards > /tmp/stage3-notyet-output-guards.log 2>&1
go test ./stage3/census/latent/refusalrewrite ./stage3/census/latent/statecopy ./stage3/census/latent/statementrewrite -run 'TestCompiler41231d51Shape|TestMissingFunctionMutantFailsLoudly|TestChangedVisitorMutantFailsLoudly|TestVisitorsCollectContinueAndSkipDiagnosedBodies|Test' -count=1 > /tmp/stage3-notyet-overlay-tests.log 2>&1
go vet ./stage3/census/latent/replay ./stage3/census/latent/replay/worker > /tmp/stage3-notyet-vet.log 2>&1
```

The positive commands exited 0; replay tests printed `Ran 2 tests in 4.422s`, `OK`. The deliberate replay-selection test command exited 1 at its reproduction assertion (`AssertionError: 1 != 0`). The AST-helper tests passed in 0.977s, 1.125s and 0.368s. Focused vet produced no output. [Full audit](evidence/full-audit.log.txt), [replay tests](evidence/replay-tests.log.txt), [selection mutant](evidence/replay-mutant.log.txt), [legacy audit](evidence/legacy-audit.log.txt), [output guards](evidence/output-guards.log.txt), [AST helpers](evidence/overlay-tests.log.txt).

The inherited full/replay probes initially failed because they expected truthy string/number conditions to be refused, which this base already supports. [Inherited full failure](evidence/inherited-full-audit-failure.log.txt) and [replay failure](evidence/inherited-replay-failure.log.txt). `validate-tooling.py` changes only synthetic witnesses: debugger and numeric == produce actual-lowering NotYet beside var's Refused; the diagnosed grandchild is a sibling so this base does not stop the middle function at a dependency-signature boundary. It preserves continuation, rollback, nested checker eligibility, complete ordered replay equality, non-BMP column conversion and exact-signature assertions. No census/replay implementation or compiler source is changed.

| Mutant | Check that caught it |
| --- | --- |
| First-error-only lowering | Three distinct actual-lowering failures in one unit; mutant observes only the first |
| Keep failed initializer state | Missing `reading poison` rollback finding |
| Return non-nil IR | `measurement returned usable IR` output guard |
| Expose permissive ordinary loader | `measurement loader exposed an output program` guard |
| Replay parent's sibling, synthetic and real tsc site | Exact-node signature reproduction assertion |
| Include signatures in body eligibility | Signature/body range assertion in legacy audit |
| Plant an extra NotYet at target function | Exactly one-site delta and unchanged sibling file assertions |
| Misattribute planted finding to sibling file | Unchanged-file attribution assertion |
| Missing refusal visitor/function or changed child walk | Refusal AST transformer tests fail loudly |
| Unknown mutable state container or foreign pointer | Typed state-copy generator tests fail loudly |
| Missing statement method | Statement wrapper generator test fails loudly |
| Remove real nested declaration | Independent stock AST coverage |
| Change body byte span | Independent stock AST span |
| Forge direct checker ownership | Independent diagnostic ownership |
| Drop raw observation or change phase | Exact filtered multiset |
| Change total or reason count | Deduplicated totals and sorted rank/count |
| Change diagnostic file or example | Independent file attribution and example membership |
| Change context flag | Exact reading-X convention |
| Change disposition, including writable variance | Independent ruling checks |
| Change top-ten share or under-five tail | Independent coverage/tail recount |

All 14 accounting/ledger corruption mutants were rejected. Their exact assertion messages are preserved in [accounting.log.txt](evidence/accounting.log.txt).

## Delivery ancestry

After the census finished, origin/main advanced to `6f16a1693ff41bc102c4d9bfac83b6330277c479`. Merged it cleanly in `913397e39fd15e7090626204e11a5baeb1c29668` for delivery. Its changes are stage-1 and gate-sampling tooling. All cmd/adamic, load, lower, IR, native, JavaScript, oracle, flow/fresh and cohere inputs remain identical to the measured compiler base; no census rebuild or recount is required. The accounting/ledger audit, saved replay comparison and positive replay integration tests were rerun after this merge and passed.

Delivery validation commands and preserved outputs:

```sh
python3 stage3/notyet-table/audit.py stage3/notyet-table /tmp/stage3-notyet-adapted > /tmp/stage3-notyet-accounting-final.log 2>&1
python3 stage3/notyet-table/verify-replay.py > /tmp/stage3-notyet-replay-accounting-final.log 2>&1
python3 stage3/notyet-table/validate-tooling.py replay > /tmp/stage3-notyet-replay-tests-delivery.log 2>&1
```

All exit 0. Replay integration tests printed `Ran 2 tests in 7.872s`, `OK`. [Accounting](evidence/accounting-delivery.log.txt), [saved replay comparison](evidence/replay-accounting-delivery.log.txt), [integration tests](evidence/replay-tests-delivery.log.txt). `git diff --check` passed.

The refusal-table branch also advanced during the measurement to `d35a81d36fdafccf827bad0f572d311b2a0d4deb`, incorporating its own resolved non-null merge. It is not the pin measured here. This report preserves the selected `44583d32` fallback and does not claim counts on that later, substantially different compiler. The late upstream observations are recorded in metadata.json.

## Recorded-counts limitation

```sh
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m > /tmp/stage3-notyet-counts.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$/^fixtures$/^internal/oracle/testdata/process_exit.a$' -count=1 -timeout 5m > /tmp/stage3-notyet-counts-baseline-pin.log 2>&1
```

Both exited 1. The full recorded-count check took 83.161s and had 42 failed fixture subtests, including existing predicate-contract refusals, process.exit-as-value NotYet sites and graph-region executions ending with `free(): invalid pointer`. The focused unchanged-base pin reproduces the process_exit.a NotYet at `2:44` in 0.095s. [Complete counts log](evidence/counts.log.txt) (display trailing whitespace normalized; [original bytes](evidence/counts.raw.log.gz)) and [focused pin](evidence/counts-baseline-pin.log.txt). Production sources, fixture registrations and counts are unchanged from `44583d32`; these failures were not introduced by the table/replay merge. No oracle fixture was added and no recorded count was refreshed or changed.

The report's own census, ledger, accounting and replay validations pass. No whole compiler package or full repository gate was run. The census/replay invokes no backend; the separate recorded-counts check runs native fixtures. This report does not establish successful whole-program lowering, exhaustive expression-level coverage, final ownership correctness or native/JavaScript semantics of tsc.

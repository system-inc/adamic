# Compiler proof #39wtqdj — lower-39w-2


The proof is not fully green: step 3a and the full lower suite fail two branded-supertype diagnostic expectations at HEAD2. These are compiler/test expectation mismatches, not environment failures. Admission passes with no classification delta. All three requested mutants are caught by new red cases beyond the two HEAD2 failures.


HEAD: `94969d02eb643d9f57edb8196835f484772d99ce`

Main: `d43437079a7e7ca723168e67c5efb5bf7ebee216`

HEAD2: `6d64f3f567755dda398b62809aa87b2973c962a9`


| Step | Result | Exit |
|---|---|---|
| 1 formatting/vet | PASS; gofmt printed nothing | 0 |
| 2 oracle writer | PASS; exactly the three expected count changes | 0 |
| 3a lower pins | FAIL; two branded diagnostic expectations | 1 |
| 3b uncached oracle pins | PASS | 0 |
| 3c explain-checks pins | PASS | 0 |
| M1 3a / 3b | caught / caught | 1 / 1 |
| M2 3a / 3b | caught / caught | 1 / 1 |
| M3 3a / 3b | caught / passes | 1 / 0 |
| 5 full lower | FAIL; same two branded expectations | 1 |
| 5b full IR | PASS | 0 |
| 6 full oracle | FAIL; 13 environment failures, all reproduced at main | 1 |
| 7 admission | PASS, non-void | 0 |


## Decisive HEAD2 failures


```text
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.74s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.05s)
        predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u924662774/001/main.ts:16:56: stage 0 can't lower a read of i narrowed on the false side of named(i) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u924662774/001/main.ts:16:13, a call through a predicate value whose false answer isn't checked yet
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.12s)
        predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin2623058715/001/main.ts:14:131: stage 0 can't lower a read of x narrowed on the false side of either(x) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin2623058715/001/main.ts:14:62, a call through a predicate value whose false answer isn't checked yet
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1523.860s
FAIL
```


## Environment and execution discipline


Read the cloud-environment runtime skill and loaded `/workspace/adamic-tools/env.sh` before repository commands. No AGENTS.md was found in the checkout. Created `/tmp/adamic-gate` outside the repository because env.sh sets TMPDIR there. Heavy commands ran sequentially; no other test, install, build, or agent ran alongside admission. No sub-agents were used. All requested post-count proof runs used clean HEAD2, except each expressly requested temporary mutant. Main comparison used an isolated external detached worktree.


Pinned upstream was provisioned at `/tmp/lower-39w-proof/typescript` with:

```sh
git init /tmp/lower-39w-proof/typescript
git -C /tmp/lower-39w-proof/typescript remote add origin https://github.com/microsoft/TypeScript.git
git -C /tmp/lower-39w-proof/typescript fetch --depth=1 origin 050880ce59e30b356b686bd3144efe24f875ebc8
git -C /tmp/lower-39w-proof/typescript checkout --detach FETCH_HEAD
```


`git rev-parse HEAD` there returned `050880ce59e30b356b686bd3144efe24f875ebc8`. Copied stage3/api/package.json and package-lock.json to `/tmp/lower-39w-proof/api`, then ran `npm ci --ignore-scripts --prefix /tmp/lower-39w-proof/api`. All installed packages remained outside the repository.


```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lower-39w-proof/typescript
export PATH=/tmp/lower-39w-proof/api/node_modules/.bin:$PATH
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```


```text
go version go1.27.1 linux/amd64
Node.js v24.19.0
Version 6.0.3
```


For admission, an external `/tmp/lower-39w-proof/bin/npm` wrapper was first on PATH. The gate calls npm through toolchain.Env, which filters Node/npm CA variables. The wrapper re-exported the three settings above and execed `/opt/codex/runtimes/codex-primary-runtime/dependencies/node/bin/npm "$@"`, preserving them for npm and its descendants. The invoking gate also exported all three settings. No TLS verification was disabled.


## Step 1


```sh
git fetch origin compiler/predicate-false-read-2
git checkout --detach 94969d02eb643d9f57edb8196835f484772d99ce
git submodule update --init --recursive
gofmt -l internal/lower internal/oracle cmd/adamic
go vet ./internal/lower ./internal/oracle ./cmd/adamic
```


Exit 0. `gofmt -l` printed nothing; vet emitted no diagnostics. Submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.


## Step 2 — oracle-owned re-recording


```sh
go test ./internal/oracle -run 'TestCountsAreRecordedTable|TestCountsAreRecorded_[0-9]+/internal/lower/testdata/predicates_checked/true\.a_' -count=1 -timeout 40m -args -update-counts
```


Exit 0. No selector adjustment was necessary. The second alternative matches only the runtime key `internal/lower/testdata/predicates_checked/true.a (.ts checked mode)` after Go sanitizes spaces to underscores. That directory contains one true.a runtime record, and the writer rejects zero runtime measurements. The runtime record did not move. TestCountsAreRecordedTable measured all predicate records.


```text
ok  	github.com/system-inc/adamic/internal/oracle	41.608s
```


`git status --porcelain`, verbatim:

```text
 M internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts
 M internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts
 M internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts
```


`git diff`, verbatim:

```text
diff --git a/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts b/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts
index 9b5ba49a5..cf58e73a0 100644
--- a/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts
+++ b/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts
@@ -1 +1 @@
-| internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 2 | 0 |
+| internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 1 | 1 |
diff --git a/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts b/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts
index b5160fba8..631474e3a 100644
--- a/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts
+++ b/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts
@@ -1 +1 @@
-| internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 2 | 0 |
+| internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 1 | 1 |
diff --git a/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts b/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts
index 31d9548b6..48a413742 100644
--- a/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts
+++ b/internal/oracle/testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts
@@ -1 +1 @@
-| internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 2 | 0 |
+| internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 1 | 1 |
```


Before committing, a script asserted the exact three changed paths and that each new file equaled its HEAD content with only `| 1 | 0 | 2 | 0 |` replaced by `| 1 | 0 | 1 | 1 |`. No other record or file changed. Committed exactly those three .counts files, message `counts: re-recorded by -update-counts for #39wtqdj`, and pushed only `proof/lower-39w-2-counts`.


HEAD2 is `6d64f3f567755dda398b62809aa87b2973c962a9`. `git status --porcelain` printed nothing after the push.


## Step 3 — pins at HEAD2


### 3a — exit 1


```sh
go test ./internal/lower -run 'TestCheckedPredicateFalseReadDirections|TestCheckedPredicateBranches|TestCheckedPredicatePendingContracts|TestPredicateUseRegions|TestPredicateUsesBelongToEachCall|TestPredicateOverloadRuntime|TestBrandedSupertypeIdentityNarrowingLowers|TestKindPredicateMatchesTheWrittenComparison|TestBrandedSupertypeIsJudgedWhereItIsLowered' -count=1 -v -timeout 40m
```


PASS/FAIL per test and subtest:

```text
--- PASS: TestPredicateUsesBelongToEachCall (1.93s)
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (3.09s)
--- PASS: TestCheckedPredicatePendingContracts (3.39s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (1.91s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.15s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.75s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.09s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.12s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.09s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.09s)
--- PASS: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.04s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.05s)
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.05s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.06s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.04s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.05s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/replacement (0.04s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/later_read (0.04s)
    --- PASS: TestPredicateUseRegions/loop_read (0.06s)
    --- PASS: TestPredicateUseRegions/computed_index (0.05s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.07s)
    --- PASS: TestPredicateUseRegions/closure_read (0.07s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.05s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.04s)
    --- PASS: TestPredicateUseRegions/both (0.06s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/true (0.05s)
    --- PASS: TestPredicateUseRegions/false (0.07s)
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (3.56s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.06s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.08s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.04s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.05s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.10s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.06s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.12s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (32.11s)
--- PASS: TestCheckedPredicateBranches (43.40s)
    --- PASS: TestCheckedPredicateBranches/true (35.57s)
    --- PASS: TestCheckedPredicateBranches/false (4.04s)
    --- PASS: TestCheckedPredicateBranches/false_read (1.99s)
    --- PASS: TestCheckedPredicateBranches/asserts (1.09s)
    --- PASS: TestCheckedPredicateBranches/valid (0.70s)
--- PASS: TestPredicateOverloadRuntime (51.31s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (34.05s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (4.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (1.00s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.79s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.88s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.67s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (1.00s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.76s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.62s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.65s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.57s)
```


Complete verbose transcript, preserving every got/want diagnostic verbatim:

```text
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestPredicateUsesBelongToEachCall
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateBranches/true
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestCheckedPredicatePendingContracts/open_tag
--- PASS: TestPredicateUsesBelongToEachCall (1.93s)
=== CONT  TestPredicateUseRegions
=== RUN   TestPredicateUseRegions/unused
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestPredicateUseRegions/both
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestPredicateUseRegions/const_result_alias
=== PAUSE TestPredicateUseRegions/const_result_alias
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (3.09s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
--- PASS: TestCheckedPredicatePendingContracts (3.39s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (1.91s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.15s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.75s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.09s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.12s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.09s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.09s)
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestPredicateUseRegions/unused
--- PASS: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.04s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.05s)
=== CONT  TestPredicateUseRegions/indexed_argument
=== CONT  TestPredicateUseRegions/comma_continuation
=== CONT  TestPredicateUseRegions/switch_true_default
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== CONT  TestPredicateUseRegions/true_inequality
=== CONT  TestPredicateUseRegions/loose_false_comparison
=== CONT  TestPredicateUseRegions/false_comparison
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/closure_read
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.05s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.06s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.04s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.05s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/replacement (0.04s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/later_read (0.04s)
    --- PASS: TestPredicateUseRegions/loop_read (0.06s)
    --- PASS: TestPredicateUseRegions/computed_index (0.05s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.07s)
    --- PASS: TestPredicateUseRegions/closure_read (0.07s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.05s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.04s)
    --- PASS: TestPredicateUseRegions/both (0.06s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/true (0.05s)
    --- PASS: TestPredicateUseRegions/false (0.07s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u2382327862/001/main.ts:16:56: stage 0 can't lower a read of i narrowed on the false side of named(i) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u2382327862/001/main.ts:16:13, a call through a predicate value whose false answer isn't checked yet
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== NAME  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin1963960557/001/main.ts:14:131: stage 0 can't lower a read of x narrowed on the false side of either(x) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin1963960557/001/main.ts:14:62, a call through a predicate value whose false answer isn't checked yet
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (3.56s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.06s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.08s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.04s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.05s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.10s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.06s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.12s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (32.11s)
=== RUN   TestCheckedPredicateBranches/false
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestPredicateOverloadRuntime/overload_some
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== RUN   TestCheckedPredicateBranches/valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
--- PASS: TestCheckedPredicateBranches (43.40s)
    --- PASS: TestCheckedPredicateBranches/true (35.57s)
    --- PASS: TestCheckedPredicateBranches/false (4.04s)
    --- PASS: TestCheckedPredicateBranches/false_read (1.99s)
    --- PASS: TestCheckedPredicateBranches/asserts (1.09s)
    --- PASS: TestCheckedPredicateBranches/valid (0.70s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (51.31s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (34.05s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (4.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (1.00s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.79s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.88s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.67s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (1.00s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.76s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.62s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.65s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.57s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	53.464s
FAIL
```


### 3b — exit 0


```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestPredicateDirectionCountsAreRecorded|TestPredicateCountsRegenerateDeterministically' -count=1 -v -timeout 40m
```


PASS/FAIL per test and subtest:

```text
--- PASS: TestPredicateCountsRegenerateDeterministically (0.86s)
--- PASS: TestPredicateDirectionCountsAreRecorded (0.88s)
--- PASS: TestCheckedPredicateCountsAreRecorded (10.75s)
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/renarrowed (255.29s)
    --- PASS: TestPredicateFalseRead/spread (245.41s)
    --- PASS: TestPredicateFalseRead/again (255.38s)
    --- PASS: TestPredicateFalseRead/no_read (256.29s)
    --- PASS: TestPredicateFalseRead/getter (0.80s)
    --- PASS: TestPredicateFalseRead/early_exit (1.59s)
    --- PASS: TestPredicateFalseRead/capture (1.56s)
    --- PASS: TestPredicateFalseRead/getter_capture (1.72s)
    --- PASS: TestPredicateFalseRead/else_read (1.13s)
    --- PASS: TestPredicateFalseRead/false_read (0.93s)
--- PASS: TestCheckedPredicateOracle (276.82s)
    --- PASS: TestCheckedPredicateOracle/true (257.30s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.42s)
    --- PASS: TestCheckedPredicateOracle/false (0.81s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.90s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.95s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.17s)
    --- PASS: TestCheckedPredicateOracle/valid (0.96s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.82s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.07s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.49s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.90s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.99s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.09s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.14s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.12s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.18s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.01s)
```


Complete verbose transcript, preserving every got/want diagnostic verbatim:

```text
=== RUN   TestPredicateCountsRegenerateDeterministically
=== PAUSE TestPredicateCountsRegenerateDeterministically
=== RUN   TestPredicateDirectionCountsAreRecorded
=== PAUSE TestPredicateDirectionCountsAreRecorded
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateCountsRegenerateDeterministically
=== CONT  TestPredicateFalseRead
=== CONT  TestPredicateDirectionCountsAreRecorded
=== CONT  TestCheckedPredicateOracle
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== RUN   TestCheckedPredicateOracle/true
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
--- PASS: TestPredicateCountsRegenerateDeterministically (0.86s)
=== CONT  TestPredicateFalseRead/renarrowed
--- PASS: TestPredicateDirectionCountsAreRecorded (0.88s)
=== CONT  TestPredicateFalseRead/again
--- PASS: TestCheckedPredicateCountsAreRecorded (10.75s)
=== CONT  TestPredicateFalseRead/spread
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 243.91
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 miss 249.19
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 249.20
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 249.20
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 249.19
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.03
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.04
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.02
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.02
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.02
=== NAME  TestCheckedPredicateOracle/true_unused
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/renarrowed (255.29s)
    --- PASS: TestPredicateFalseRead/spread (245.41s)
    --- PASS: TestPredicateFalseRead/again (255.38s)
    --- PASS: TestPredicateFalseRead/no_read (256.29s)
    --- PASS: TestPredicateFalseRead/getter (0.80s)
    --- PASS: TestPredicateFalseRead/early_exit (1.59s)
    --- PASS: TestPredicateFalseRead/capture (1.56s)
    --- PASS: TestPredicateFalseRead/getter_capture (1.72s)
    --- PASS: TestPredicateFalseRead/else_read (1.13s)
    --- PASS: TestPredicateFalseRead/false_read (0.93s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.05
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.12
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.02
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 217e9da50375 hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (276.82s)
    --- PASS: TestCheckedPredicateOracle/true (257.30s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.42s)
    --- PASS: TestCheckedPredicateOracle/false (0.81s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.90s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.95s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.17s)
    --- PASS: TestCheckedPredicateOracle/valid (0.96s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.82s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.07s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.49s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.90s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.99s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.09s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.14s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.12s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.18s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.01s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=68
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	276.973s
```


### 3c — exit 0


```sh
go test ./cmd/adamic -run 'TestExplainChecksOutput' -count=1 -v -timeout 30m
```


PASS/FAIL per test and subtest:

```text
--- PASS: TestExplainChecksOutput (0.00s)
    --- PASS: TestExplainChecksOutput/overload_some_true_only (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/js (0.10s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/c (0.37s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/build (4.21s)
    --- PASS: TestExplainChecksOutput/overload_some_empty (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/js (0.27s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/c (0.34s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/build (4.25s)
    --- PASS: TestExplainChecksOutput/overload_erased (0.00s)
        --- PASS: TestExplainChecksOutput/overload_erased/js (0.14s)
        --- PASS: TestExplainChecksOutput/overload_erased/c (0.17s)
        --- PASS: TestExplainChecksOutput/overload_erased/build (3.89s)
    --- PASS: TestExplainChecksOutput/overload_some_false_read (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/c (0.31s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/js (0.11s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/build (3.63s)
    --- PASS: TestExplainChecksOutput/overload_assertion (0.00s)
        --- PASS: TestExplainChecksOutput/overload_assertion/c (0.26s)
        --- PASS: TestExplainChecksOutput/overload_assertion/js (0.11s)
        --- PASS: TestExplainChecksOutput/overload_assertion/build (3.74s)
```


Complete verbose transcript, preserving every got/want diagnostic verbatim:

```text
=== RUN   TestExplainChecksOutput
=== PAUSE TestExplainChecksOutput
=== CONT  TestExplainChecksOutput
=== RUN   TestExplainChecksOutput/overload_some_empty
=== PAUSE TestExplainChecksOutput/overload_some_empty
=== RUN   TestExplainChecksOutput/overload_some_false_read
=== PAUSE TestExplainChecksOutput/overload_some_false_read
=== RUN   TestExplainChecksOutput/overload_some_true_only
=== PAUSE TestExplainChecksOutput/overload_some_true_only
=== RUN   TestExplainChecksOutput/overload_erased
=== PAUSE TestExplainChecksOutput/overload_erased
=== RUN   TestExplainChecksOutput/overload_assertion
=== PAUSE TestExplainChecksOutput/overload_assertion
=== CONT  TestExplainChecksOutput/overload_some_empty
=== CONT  TestExplainChecksOutput/overload_erased
=== CONT  TestExplainChecksOutput/overload_assertion
=== RUN   TestExplainChecksOutput/overload_some_empty/c
=== PAUSE TestExplainChecksOutput/overload_some_empty/c
=== RUN   TestExplainChecksOutput/overload_some_empty/js
=== PAUSE TestExplainChecksOutput/overload_some_empty/js
=== RUN   TestExplainChecksOutput/overload_some_empty/build
=== PAUSE TestExplainChecksOutput/overload_some_empty/build
=== CONT  TestExplainChecksOutput/overload_some_empty/c
=== CONT  TestExplainChecksOutput/overload_some_true_only
=== CONT  TestExplainChecksOutput/overload_some_empty/build
=== RUN   TestExplainChecksOutput/overload_assertion/c
=== PAUSE TestExplainChecksOutput/overload_assertion/c
=== RUN   TestExplainChecksOutput/overload_some_true_only/c
=== PAUSE TestExplainChecksOutput/overload_some_true_only/c
=== RUN   TestExplainChecksOutput/overload_some_true_only/js
=== PAUSE TestExplainChecksOutput/overload_some_true_only/js
=== RUN   TestExplainChecksOutput/overload_erased/c
=== RUN   TestExplainChecksOutput/overload_some_true_only/build
=== PAUSE TestExplainChecksOutput/overload_some_true_only/build
=== PAUSE TestExplainChecksOutput/overload_erased/c
=== RUN   TestExplainChecksOutput/overload_erased/js
=== CONT  TestExplainChecksOutput/overload_some_false_read
=== PAUSE TestExplainChecksOutput/overload_erased/js
=== RUN   TestExplainChecksOutput/overload_erased/build
=== PAUSE TestExplainChecksOutput/overload_erased/build
=== RUN   TestExplainChecksOutput/overload_assertion/js
=== CONT  TestExplainChecksOutput/overload_some_empty/js
=== PAUSE TestExplainChecksOutput/overload_assertion/js
=== RUN   TestExplainChecksOutput/overload_some_false_read/c
=== PAUSE TestExplainChecksOutput/overload_some_false_read/c
=== RUN   TestExplainChecksOutput/overload_some_false_read/js
=== PAUSE TestExplainChecksOutput/overload_some_false_read/js
=== RUN   TestExplainChecksOutput/overload_some_false_read/build
=== PAUSE TestExplainChecksOutput/overload_some_false_read/build
=== RUN   TestExplainChecksOutput/overload_assertion/build
=== CONT  TestExplainChecksOutput/overload_some_true_only/c
=== PAUSE TestExplainChecksOutput/overload_assertion/build
=== CONT  TestExplainChecksOutput/overload_some_true_only/build
=== CONT  TestExplainChecksOutput/overload_some_true_only/js
=== CONT  TestExplainChecksOutput/overload_erased/c
=== CONT  TestExplainChecksOutput/overload_erased/js
=== CONT  TestExplainChecksOutput/overload_erased/build
=== CONT  TestExplainChecksOutput/overload_some_false_read/c
=== CONT  TestExplainChecksOutput/overload_assertion/c
=== CONT  TestExplainChecksOutput/overload_some_false_read/js
=== CONT  TestExplainChecksOutput/overload_some_false_read/build
=== CONT  TestExplainChecksOutput/overload_assertion/js
=== CONT  TestExplainChecksOutput/overload_assertion/build
--- PASS: TestExplainChecksOutput (0.00s)
    --- PASS: TestExplainChecksOutput/overload_some_true_only (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/js (0.10s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/c (0.37s)
        --- PASS: TestExplainChecksOutput/overload_some_true_only/build (4.21s)
    --- PASS: TestExplainChecksOutput/overload_some_empty (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/js (0.27s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/c (0.34s)
        --- PASS: TestExplainChecksOutput/overload_some_empty/build (4.25s)
    --- PASS: TestExplainChecksOutput/overload_erased (0.00s)
        --- PASS: TestExplainChecksOutput/overload_erased/js (0.14s)
        --- PASS: TestExplainChecksOutput/overload_erased/c (0.17s)
        --- PASS: TestExplainChecksOutput/overload_erased/build (3.89s)
    --- PASS: TestExplainChecksOutput/overload_some_false_read (0.00s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/c (0.31s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/js (0.11s)
        --- PASS: TestExplainChecksOutput/overload_some_false_read/build (3.63s)
    --- PASS: TestExplainChecksOutput/overload_assertion (0.00s)
        --- PASS: TestExplainChecksOutput/overload_assertion/c (0.26s)
        --- PASS: TestExplainChecksOutput/overload_assertion/js (0.11s)
        --- PASS: TestExplainChecksOutput/overload_assertion/build (3.74s)
PASS
ok  	github.com/system-inc/adamic/cmd/adamic	4.751s
```


## Step 4 — mutants, one at a time


Every run asserted clean HEAD2 and exactly one occurrence of the old text, saved the diff, applied one edit, ran 3a then 3b using the exact commands above, restored the original bytes in a finally block, and ran `git diff --exit-code`. The .counts files were never changed by a mutant.


### M1
The condition-only arm restored: extracted the fenced Go block from internal/lower/predicate_false_read.mutants.md and inserted it immediately before the uniquely identified final return directions in checkedPredicateDirections.


```text
old text occurs exactly once
```


Exact applied diff:

```text
diff --git a/internal/lower/predicates_checked.go b/internal/lower/predicates_checked.go
index 4052c4d27..9bc68ad5d 100644
--- a/internal/lower/predicates_checked.go
+++ b/internal/lower/predicates_checked.go
@@ -62,6 +62,54 @@ func (l *lowering) checkedPredicateDirections(call *ast.CallExpression, claim *c
 	if l.predicateUseDirections(call)&predicateFalse != 0 || l.predicateFalseRead(call, argument) {
 		return predicateEither
 	}
+	conditionNarrows := func(condition *ast.Node, truth bool) bool {
+		if condition == nil || l.predicateConditionDirection(condition, call.AsNode(), truth, 0)&predicateFalse == 0 {
+			return false
+		}
+		flags := ast.FlowFlagsFalseCondition
+		if truth {
+			flags = ast.FlowFlagsTrueCondition
+		}
+		branch := &ast.FlowNode{Flags: flags, Node: condition, Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
+		narrowed := predicateFlowType(l.checker, argument, source, source, nil, branch)
+		return !checker.Checker_isTypeIdenticalTo(l.checker, source, narrowed)
+	}
+	var visit ast.Visitor
+	visit = func(node *ast.Node) bool {
+		var condition *ast.Node
+		switch node.Kind {
+		case ast.KindIfStatement:
+			condition = node.AsIfStatement().Expression
+		case ast.KindConditionalExpression:
+			condition = node.AsConditionalExpression().Condition
+		case ast.KindWhileStatement:
+			condition = node.AsWhileStatement().Expression
+		case ast.KindDoStatement:
+			condition = node.AsDoStatement().Expression
+		case ast.KindForStatement:
+			condition = node.AsForStatement().Condition
+		case ast.KindBinaryExpression:
+			binary := node.AsBinaryExpression()
+			if binary.OperatorToken.Kind == ast.KindBarBarToken && conditionNarrows(binary.Left, false) {
+				directions |= predicateFalse
+			}
+		case ast.KindSwitchStatement:
+			statement := node.AsSwitchStatement()
+			if ast.SkipParentheses(statement.Expression).Kind == ast.KindTrueKeyword {
+				for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
+					if clause.Kind == ast.KindCaseClause && conditionNarrows(clause.Expression(), false) {
+						directions |= predicateFalse
+					}
+				}
+			}
+		}
+		if conditionNarrows(condition, true) || conditionNarrows(condition, false) {
+			directions |= predicateFalse
+		}
+		node.ForEachChild(visit)
+		return false
+	}
+	ast.GetSourceFileOfNode(call.AsNode()).AsNode().ForEachChild(visit)
 	return directions
 }
 
```


3a: exit 1. Red tests/subtests:

```text
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/empty_branches (0.06s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.15s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.05s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/semver's_toString (0.06s)
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (2.17s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.04s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.05s)
--- FAIL: TestCheckedPredicateBranches (9.20s)
    --- FAIL: TestCheckedPredicateBranches/false (0.48s)
--- FAIL: TestPredicateOverloadRuntime (17.99s)
    --- FAIL: TestPredicateOverloadRuntime/overload_some_true_only (0.06s)
    --- FAIL: TestPredicateOverloadRuntime/overload_checked (0.05s)
```


Full transcript, including every red subtest and its complete got diagnostics:

```text
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== CONT  TestCheckedPredicateFalseReadDirections
=== CONT  TestPredicateUseRegions
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestPredicateUseRegions/unused
=== CONT  TestCheckedPredicateBranches
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestPredicateOverloadRuntime/parser_every_result
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateBranches/true
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestPredicateUseRegions/both
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestPredicateUseRegions/const_result_alias
=== PAUSE TestPredicateUseRegions/const_result_alias
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestPredicateUsesBelongToEachCall
--- PASS: TestPredicateUsesBelongToEachCall (0.19s)
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestCheckedPredicatePendingContracts/open_tag
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (1.40s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
--- PASS: TestCheckedPredicatePendingContracts (2.16s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.19s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.12s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.32s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.64s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.17s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.51s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.09s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestPredicateUseRegions/unused
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/closure_read
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
=== CONT  TestPredicateUseRegions/loose_false_comparison
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestPredicateUseRegions/indexed_argument
=== NAME  TestCheckedPredicateFalseReadDirections/empty_branches
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestPredicateUseRegions/false_comparison
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestPredicateUseRegions/comma_continuation
=== CONT  TestPredicateUseRegions/switch_true_default
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== CONT  TestPredicateUseRegions/true_inequality
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
--- PASS: TestPredicateUseRegions (0.00s)
    --- PASS: TestPredicateUseRegions/unused (0.05s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.05s)
    --- PASS: TestPredicateUseRegions/closure_read (0.07s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.11s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.06s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.11s)
    --- PASS: TestPredicateUseRegions/both (0.05s)
    --- PASS: TestPredicateUseRegions/false (0.06s)
    --- PASS: TestPredicateUseRegions/true (0.05s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.06s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.06s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.04s)
    --- PASS: TestPredicateUseRegions/replacement (0.05s)
    --- PASS: TestPredicateUseRegions/later_read (0.04s)
    --- PASS: TestPredicateUseRegions/loop_read (0.06s)
    --- PASS: TestPredicateUseRegions/computed_index (0.05s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.05s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.05s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.75s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.76s)
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== NAME  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== NAME  TestCheckedPredicateFalseReadDirections/semver's_toString
    predicates_false_read_test.go:89: false side checked true, want false
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/empty_branches (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.22s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.57s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.36s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.59s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.78s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.30s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.27s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.16s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.05s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/semver's_toString (0.06s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== NAME  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u1815888303/001/main.ts:16:56: stage 0 can't lower a read of i narrowed on the false side of named(i) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u1815888303/001/main.ts:16:13, a call through a predicate value whose false answer isn't checked yet
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== NAME  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin675431728/001/main.ts:14:131: stage 0 can't lower a read of x narrowed on the false side of either(x) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin675431728/001/main.ts:14:62, a call through a predicate value whose false answer isn't checked yet
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (2.17s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.05s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.04s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.04s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.05s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.05s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.05s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.05s)
=== RUN   TestCheckedPredicateBranches/false
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.51s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.08s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (4.73s)
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== NAME  TestCheckedPredicateBranches/false
    predicates_checked_test.go:90: /tmp/adamic-gate/TestCheckedPredicateBranchesfalse2825438461/002/native: exit 70 stdout "" stderr "adamic: panic: predicate isNumber at /tmp/adamic-gate/TestCheckedPredicateBranchesfalse2825438461/001/main.ts:2:50 false branch: source string | number, target number\n"; want 0 "continued\n" ""
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestPredicateOverloadRuntime/overload_some
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== RUN   TestCheckedPredicateBranches/valid
--- FAIL: TestCheckedPredicateBranches (9.20s)
    --- PASS: TestCheckedPredicateBranches/true (6.88s)
    --- FAIL: TestCheckedPredicateBranches/false (0.48s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.60s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.66s)
    --- PASS: TestCheckedPredicateBranches/valid (0.57s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_true_only85295232/001/main.ts:7:9 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly number[] | undefined -> readonly number[]} {Direction:false Status:checked Reason:readonly number[] | undefined -> readonly number[]}]}]}, want [0 1 1] (proven, checked, unobservable)
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
    predicates_overload_test.go:89: predicate counts: {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_checked2736610597/001/main.ts:3:50 Function:lie Overload:1 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}, want [0 1 1] (proven, checked, unobservable)
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- FAIL: TestPredicateOverloadRuntime (17.99s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (7.13s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (0.62s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (0.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (1.41s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (1.03s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.57s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.75s)
    --- FAIL: TestPredicateOverloadRuntime/overload_some_true_only (0.06s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.59s)
    --- FAIL: TestPredicateOverloadRuntime/overload_checked (0.05s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.60s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.66s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.65s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.73s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	18.104s
FAIL
```


3b: exit 1. Red tests/subtests:

```text
--- FAIL: TestPredicateCountsRegenerateDeterministically (0.98s)
--- FAIL: TestPredicateDirectionCountsAreRecorded (1.06s)
--- FAIL: TestCheckedPredicateCountsAreRecorded (12.49s)
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/no_read (7.82s)
--- FAIL: TestCheckedPredicateOracle (54.10s)
    --- FAIL: TestCheckedPredicateOracle/true (7.71s)
    --- FAIL: TestCheckedPredicateOracle/false (0.33s)
    --- FAIL: TestCheckedPredicateOracle/false_alias (0.28s)
```


Full transcript, including every red subtest and its complete got diagnostics:

```text
=== RUN   TestPredicateCountsRegenerateDeterministically
=== PAUSE TestPredicateCountsRegenerateDeterministically
=== RUN   TestPredicateDirectionCountsAreRecorded
=== PAUSE TestPredicateDirectionCountsAreRecorded
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== CONT  TestPredicateDirectionCountsAreRecorded
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestPredicateFalseRead/no_read
=== CONT  TestPredicateCountsRegenerateDeterministically
=== PAUSE TestPredicateFalseRead/no_read
=== CONT  TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
=== NAME  TestPredicateCountsRegenerateDeterministically
    count_files_test.go:197: regenerated counts differ in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts:
        recorded: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 1 | 1 |
        regenerated: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 2 | 0 |
    count_files_test.go:197: regenerated counts differ in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts:
        recorded: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 1 | 1 |
        regenerated: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 2 | 0 |
    count_files_test.go:197: regenerated counts differ in testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts:
        recorded: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 1 | 1 |
        regenerated: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 2 | 0 |
--- FAIL: TestPredicateCountsRegenerateDeterministically (0.98s)
=== CONT  TestPredicateFalseRead/again
=== NAME  TestPredicateDirectionCountsAreRecorded
    predicate_counts_test.go:100: predicate counts changed in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 1 | 1 |
        measured: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 2 | 0 |
    predicate_counts_test.go:100: predicate counts changed in testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 1 | 1 |
        measured: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 2 | 0 |
    predicate_counts_test.go:100: predicate counts changed in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 1 | 1 |
        measured: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 2 | 0 |
--- FAIL: TestPredicateDirectionCountsAreRecorded (1.06s)
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestCheckedPredicateOracle/true
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue948721414/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:80: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/getter_capture
=== NAME  TestCheckedPredicateCountsAreRecorded
    predicates_checked_test.go:280: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
    predicates_checked_test.go:280: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
--- FAIL: TestCheckedPredicateCountsAreRecorded (12.49s)
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 24.23
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 24.23
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 24.23
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 miss 24.23
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 19.91
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.03
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse4212466191/001/false.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/false_alias
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.03
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.02
=== NAME  TestCheckedPredicateOracle/false_alias
    predicates_checked_test.go:101: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse_alias3144931103/001/false_alias.ts:2:61 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/no_read (7.82s)
    --- PASS: TestPredicateFalseRead/getter (0.23s)
    --- PASS: TestPredicateFalseRead/getter_capture (25.64s)
    --- PASS: TestPredicateFalseRead/again (32.49s)
    --- PASS: TestPredicateFalseRead/spread (32.43s)
    --- PASS: TestPredicateFalseRead/renarrowed (20.80s)
    --- PASS: TestPredicateFalseRead/capture (1.22s)
    --- PASS: TestPredicateFalseRead/early_exit (1.35s)
    --- PASS: TestPredicateFalseRead/else_read (1.36s)
    --- PASS: TestPredicateFalseRead/false_read (1.37s)
=== NAME  TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.02
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.08
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 66813bbcf1e5 hit 0.02
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- FAIL: TestCheckedPredicateOracle (54.10s)
    --- FAIL: TestCheckedPredicateOracle/true (7.71s)
    --- PASS: TestCheckedPredicateOracle/true_unused (26.53s)
    --- FAIL: TestCheckedPredicateOracle/false (0.33s)
    --- FAIL: TestCheckedPredicateOracle/false_alias (0.28s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.75s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.12s)
    --- PASS: TestCheckedPredicateOracle/valid (0.75s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.91s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (0.93s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.33s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.90s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.38s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.86s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.65s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.64s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.37s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (2.51s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.65s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.51s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=63
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	54.268s
FAIL
```


Restore check:

```text
git diff --exit-code: 0
```


### M2
Types decide alone: removed only `|| l.predicateFalseRead(call, argument)` from the specified if condition.


```text
old text occurs exactly once
```


Exact applied diff:

```text
diff --git a/internal/lower/predicates_checked.go b/internal/lower/predicates_checked.go
index 4052c4d27..9cf2abf76 100644
--- a/internal/lower/predicates_checked.go
+++ b/internal/lower/predicates_checked.go
@@ -59,7 +59,7 @@ func (l *lowering) checkedPredicateDirections(call *ast.CallExpression, claim *c
 	// merge that the false side still narrows; predicateFalseRead adds every read the false side reaches
 	// before they merge, by the flow graph alone. A false side nothing reads makes no claim the program can
 	// see: its check is dropped, and recordPredicateSite counts the direction unobservable.
-	if l.predicateUseDirections(call)&predicateFalse != 0 || l.predicateFalseRead(call, argument) {
+	if l.predicateUseDirections(call)&predicateFalse != 0 {
 		return predicateEither
 	}
 	return directions
```


3a: exit 1. Red tests/subtests:

```text
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.05s)
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (0.87s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.15s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.07s)
```


Full transcript, including every red subtest and its complete got diagnostics:

```text
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== CONT  TestCheckedPredicateBranches
=== CONT  TestPredicateUseRegions
=== RUN   TestCheckedPredicateBranches/true
=== CONT  TestPredicateUsesBelongToEachCall
=== RUN   TestPredicateUseRegions/unused
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== CONT  TestCheckedPredicatePendingContracts
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestPredicateUseRegions/both
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestPredicateUseRegions/const_result_alias
=== PAUSE TestPredicateUseRegions/const_result_alias
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== RUN   TestPredicateOverloadRuntime/parser_every_result
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicatePendingContracts/open_tag
--- PASS: TestPredicateUsesBelongToEachCall (0.40s)
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.57s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
--- PASS: TestCheckedPredicatePendingContracts (1.02s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.39s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.14s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.13s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.08s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.07s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.06s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.08s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.06s)
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
    predicates_false_read_test.go:89: false side checked false, want true
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestPredicateUseRegions/unused
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.03s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.04s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.05s)
=== CONT  TestPredicateUseRegions/comma_continuation
=== CONT  TestPredicateUseRegions/switch_true_default
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== CONT  TestPredicateUseRegions/true_inequality
=== CONT  TestPredicateUseRegions/loose_false_comparison
=== CONT  TestPredicateUseRegions/false_comparison
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestPredicateUseRegions/indexed_argument
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/closure_read
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
--- PASS: TestPredicateUseRegions (0.02s)
    --- PASS: TestPredicateUseRegions/unused (0.04s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.05s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.07s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.06s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.06s)
    --- PASS: TestPredicateUseRegions/later_read (0.06s)
    --- PASS: TestPredicateUseRegions/replacement (0.08s)
    --- PASS: TestPredicateUseRegions/loop_read (0.06s)
    --- PASS: TestPredicateUseRegions/computed_index (0.06s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.05s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.06s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.06s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.04s)
    --- PASS: TestPredicateUseRegions/closure_read (0.12s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.10s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.14s)
    --- PASS: TestPredicateUseRegions/false (0.06s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.15s)
    --- PASS: TestPredicateUseRegions/true (0.08s)
    --- PASS: TestPredicateUseRegions/both (0.06s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.06s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== NAME  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u899354230/001/main.ts:16:56: stage 0 can't lower a read of i narrowed on the false side of named(i) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u899354230/001/main.ts:16:13, a call through a predicate value whose false answer isn't checked yet
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== NAME  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin1610715800/001/main.ts:14:131: stage 0 can't lower a read of x narrowed on the false side of either(x) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin1610715800/001/main.ts:14:62, a call through a predicate value whose false answer isn't checked yet
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (0.87s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.18s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.15s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.11s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.07s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.07s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.04s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.01s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.08s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.06s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (2.89s)
=== RUN   TestCheckedPredicateBranches/false
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestPredicateOverloadRuntime/overload_some
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== RUN   TestCheckedPredicateBranches/valid
--- PASS: TestCheckedPredicateBranches (6.40s)
    --- PASS: TestCheckedPredicateBranches/true (3.97s)
    --- PASS: TestCheckedPredicateBranches/false (0.72s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.58s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.55s)
    --- PASS: TestCheckedPredicateBranches/valid (0.58s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (15.73s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (4.26s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (0.79s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (0.71s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.64s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.56s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.57s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.75s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.59s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.60s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.75s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.61s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (1.02s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.63s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.96s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	15.967s
FAIL
```


3b: exit 1. Red tests/subtests:

```text
--- FAIL: TestPredicateFalseRead (0.01s)
    --- FAIL: TestPredicateFalseRead/getter_capture (0.27s)
```


Full transcript, including every red subtest and its complete got diagnostics:

```text
=== RUN   TestPredicateCountsRegenerateDeterministically
=== PAUSE TestPredicateCountsRegenerateDeterministically
=== RUN   TestPredicateDirectionCountsAreRecorded
=== PAUSE TestPredicateDirectionCountsAreRecorded
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateCountsRegenerateDeterministically
=== CONT  TestCheckedPredicateOracle
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestCheckedPredicateOracle/true
=== CONT  TestPredicateFalseRead
=== CONT  TestPredicateDirectionCountsAreRecorded
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
--- PASS: TestPredicateDirectionCountsAreRecorded (0.91s)
=== CONT  TestPredicateFalseRead/again
--- PASS: TestPredicateCountsRegenerateDeterministically (0.92s)
=== CONT  TestPredicateFalseRead/spread
--- PASS: TestCheckedPredicateCountsAreRecorded (9.72s)
=== CONT  TestPredicateFalseRead/getter_capture
    predicate_false_read_test.go:80: some's false direction at 38:7: "unobservable", want "checked", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/getter_capture.ts:38:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec miss 22.80
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 22.80
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 22.80
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 22.80
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 18.50
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.03
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
--- FAIL: TestPredicateFalseRead (0.01s)
    --- FAIL: TestPredicateFalseRead/getter_capture (0.27s)
    --- PASS: TestPredicateFalseRead/getter (0.27s)
    --- PASS: TestPredicateFalseRead/again (28.65s)
    --- PASS: TestPredicateFalseRead/spread (28.64s)
    --- PASS: TestPredicateFalseRead/renarrowed (19.33s)
    --- PASS: TestPredicateFalseRead/no_read (29.59s)
    --- PASS: TestPredicateFalseRead/capture (1.08s)
    --- PASS: TestPredicateFalseRead/early_exit (1.11s)
    --- PASS: TestPredicateFalseRead/false_read (1.14s)
    --- PASS: TestPredicateFalseRead/else_read (1.20s)
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.02
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.03
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.02
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.02
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.03
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.02
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.02
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ca8a83b232ec hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (58.82s)
    --- PASS: TestCheckedPredicateOracle/true (30.13s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.24s)
    --- PASS: TestCheckedPredicateOracle/false (0.72s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.95s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.85s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.05s)
    --- PASS: TestCheckedPredicateOracle/valid (0.98s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.88s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.80s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (3.56s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.81s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.09s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.89s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.58s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (5.05s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (3.87s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.09s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.24s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.05s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=67
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	58.983s
FAIL
```


Restore check:

```text
git diff --exit-code: 0
```


### M3
A merge counts: replaced `return falseMet && (!trueMet || reduced)` with exactly `return falseMet || trueMet && reduced && false`.


```text
old text occurs exactly once
```


Exact applied diff:

```text
diff --git a/internal/lower/predicates_checked.go b/internal/lower/predicates_checked.go
index 4052c4d27..6c40c67d8 100644
--- a/internal/lower/predicates_checked.go
+++ b/internal/lower/predicates_checked.go
@@ -134,7 +134,7 @@ func (l *lowering) predicateFalseReaches(flow *ast.FlowNode, call, reference *as
 		}
 	}
 	walk(flow)
-	return falseMet && (!trueMet || reduced)
+	return falseMet || trueMet && reduced && false
 }
 
 // checkedPredicateUnclassifiedUse names where the call's result goes when the direction walk
```


3a: exit 1. Red tests/subtests:

```text
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.05s)
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (0.77s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.04s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.10s)
```


Full transcript, including every red subtest and its complete got diagnostics:

```text
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== CONT  TestCheckedPredicateFalseReadDirections
=== CONT  TestPredicateUsesBelongToEachCall
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== CONT  TestCheckedPredicateBranches
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateBranches/true
=== CONT  TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicatePendingContracts/open_tag
--- PASS: TestPredicateUsesBelongToEachCall (0.21s)
=== CONT  TestPredicateUseRegions
=== RUN   TestPredicateUseRegions/unused
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestPredicateUseRegions/both
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestPredicateUseRegions/const_result_alias
=== PAUSE TestPredicateUseRegions/const_result_alias
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.58s)
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
--- PASS: TestCheckedPredicatePendingContracts (0.74s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.19s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.12s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.08s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.09s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.07s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.06s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.07s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.06s)
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== NAME  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestPredicateUseRegions/unused
=== CONT  TestPredicateUseRegions/comma_continuation
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.03s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.04s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.04s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.04s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.04s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.05s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.13s)
=== CONT  TestPredicateUseRegions/switch_true_default
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== CONT  TestPredicateUseRegions/true_inequality
=== CONT  TestPredicateUseRegions/loose_false_comparison
=== CONT  TestPredicateUseRegions/false_comparison
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestPredicateUseRegions/indexed_argument
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/closure_read
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
--- PASS: TestPredicateUseRegions (0.00s)
    --- PASS: TestPredicateUseRegions/unused (0.15s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.05s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.06s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.06s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.05s)
    --- PASS: TestPredicateUseRegions/replacement (0.04s)
    --- PASS: TestPredicateUseRegions/later_read (0.04s)
    --- PASS: TestPredicateUseRegions/loop_read (0.05s)
    --- PASS: TestPredicateUseRegions/computed_index (0.05s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.05s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.05s)
    --- PASS: TestPredicateUseRegions/closure_read (0.04s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.07s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.06s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.06s)
    --- PASS: TestPredicateUseRegions/both (0.04s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.04s)
    --- PASS: TestPredicateUseRegions/true (0.04s)
    --- PASS: TestPredicateUseRegions/false (0.05s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u448519006/001/main.ts:16:56: stage 0 can't lower a read of i narrowed on the false side of named(i) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2_over_u448519006/001/main.ts:16:13, a call through a predicate value whose false answer isn't checked yet
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== NAME  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
    predicates_branded_test.go:830: want NotYet with "checked predicate membership over an open tag domain", got /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin3455081460/001/main.ts:14:131: stage 0 can't lower a read of x narrowed on the false side of either(x) at /tmp/adamic-gate/TestBrandedSupertypeIsJudgedWhereItIsLoweredorPredicate2s_lyin3455081460/001/main.ts:14:62, a call through a predicate value whose false answer isn't checked yet
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (0.77s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.04s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.05s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.05s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.05s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.12s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.04s)
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.01s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.06s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.07s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (2.09s)
=== RUN   TestCheckedPredicateBranches/false
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestPredicateOverloadRuntime/overload_some
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== RUN   TestCheckedPredicateBranches/valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
--- PASS: TestCheckedPredicateBranches (5.56s)
    --- PASS: TestCheckedPredicateBranches/true (2.89s)
    --- PASS: TestCheckedPredicateBranches/false (0.64s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.60s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.68s)
    --- PASS: TestCheckedPredicateBranches/valid (0.74s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (15.46s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (2.91s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (0.77s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (0.75s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.82s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.71s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.78s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.85s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.73s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.98s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.71s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.69s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.83s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.89s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.67s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.80s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	15.785s
FAIL
```


3b: exit 0. Red tests/subtests:

```text
(none)
```


Full transcript, including every red subtest and its complete got diagnostics:

```text
=== RUN   TestPredicateCountsRegenerateDeterministically
=== PAUSE TestPredicateCountsRegenerateDeterministically
=== RUN   TestPredicateDirectionCountsAreRecorded
=== PAUSE TestPredicateDirectionCountsAreRecorded
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateCountsRegenerateDeterministically
=== CONT  TestCheckedPredicateOracle
=== CONT  TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== CONT  TestPredicateDirectionCountsAreRecorded
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
--- PASS: TestPredicateDirectionCountsAreRecorded (0.82s)
=== CONT  TestPredicateFalseRead/again
--- PASS: TestPredicateCountsRegenerateDeterministically (0.86s)
=== CONT  TestPredicateFalseRead/spread
--- PASS: TestCheckedPredicateCountsAreRecorded (9.91s)
=== CONT  TestPredicateFalseRead/getter_capture
=== NAME  TestPredicateFalseRead/again
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 24.51
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 miss 24.51
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 24.51
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 20.54
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 24.52
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.02
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.03
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.02
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.02
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/getter_capture (21.42s)
    --- PASS: TestPredicateFalseRead/spread (30.50s)
    --- PASS: TestPredicateFalseRead/again (30.56s)
    --- PASS: TestPredicateFalseRead/no_read (31.39s)
    --- PASS: TestPredicateFalseRead/getter (0.32s)
    --- PASS: TestPredicateFalseRead/capture (0.98s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.10s)
    --- PASS: TestPredicateFalseRead/early_exit (1.16s)
    --- PASS: TestPredicateFalseRead/else_read (1.10s)
    --- PASS: TestPredicateFalseRead/false_read (0.78s)
=== NAME  TestCheckedPredicateOracle/true_unused
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.07
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4c1470db1ed5 hit 0.01
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (50.00s)
    --- PASS: TestCheckedPredicateOracle/true (31.86s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.30s)
    --- PASS: TestCheckedPredicateOracle/false (0.73s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.75s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.97s)
    --- PASS: TestCheckedPredicateOracle/asserts (0.89s)
    --- PASS: TestCheckedPredicateOracle/valid (0.90s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.79s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.19s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (0.99s)
    --- PASS: TestCheckedPredicateOracle/condition_once (0.80s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.09s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.82s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.17s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.01s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.36s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.18s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.02s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.15s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=68
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	50.144s
```


Restore check:

```text
git diff --exit-code: 0
```


M1 adds lower false-read failures for empty branches, replacement, merged reads, semver, and the false-branch/overload checks; it also adds oracle no_read, direction/count, and runtime-record failures. For the M1 oracle diagnostics that print only got direction counts (`true`, `false`, `false_alias`), the table wants Proven=0, Checked=1, Unobservable=1, and one Site; got is printed in full above as Proven=0, Checked=2, Unobservable=0. The two branded failures in every mutant 3a run were already red at unmodified HEAD2. M2 adds the late-assigned-let capture lower failure and getter_capture oracle failure. M3 adds the merged-read lower failure; its oracle pins pass. Thus each mutant is distinguished by a new failing assertion beyond HEAD2’s existing reds.


## Step 5 — full lower


```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```


Exit 1; wall seconds 1528.2397313650001. Every failure and its first error line (neither is environment-related) appears in the full log at the beginning of this report. The parent TestBrandedSupertypeIsJudgedWhereItIsLowered fails because of those two children.


Timing note: the initial /usr/bin/time wrapper exited 127 because that executable is absent; Go did not run in that attempt. The command above then ran under Python subprocess with a monotonic clock. The reported exit is Go’s exit and the wall measurement includes process launch and build time.


## Step 5b — full IR


```sh
go test ./internal/ir -count=1 -timeout 30m
```


Exit 0. Includes TestCallTargetReadersLower and the IR field guards. No failures.


```text
ok  	github.com/system-inc/adamic/internal/ir	135.645s
```


## Step 6 — full oracle


```sh
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```


Exit 1; wall seconds 3785.13100523. All 13 failures below are ENVIRONMENT failures. Every one also failed at main d43437079a7e7ca723168e67c5efb5bf7ebee216 for the same reason.


- ENVIRONMENT — --- FAIL: TestAssignmentProofWiderTargetMutant (0.07s): assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofWiderTarget (0.04s): assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofCompoundMutant (0.05s): assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofCompound (0.04s): assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofChainedMutant (0.09s): assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofChained (0.10s): assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofReturnDefinedMutant (0.05s): assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofReturnDefined (0.08s): assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofIfNarrowingMutant (0.11s): assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofIfNarrowing (0.08s): assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofScannerKeywordMutant (0.06s): assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestAssignmentProofScannerKeyword (0.10s): assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520

- ENVIRONMENT — --- FAIL: TestStdoutColumnsRefusesWASI (0.87s): library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT

For all twelve assignment-proof tests, the first error line is followed by `Error: Cannot find module 'typescript'`. Their helper overrides NODE_PATH to `<repository>/stage3/api/node_modules`. The pinned external install makes tsc 6.0.3 available on PATH, but that does not satisfy this hardcoded CommonJS lookup. The WASI test gets `native: wasm32-wasi requires WASI_SYSROOT` instead of its expected typed refusal.


Full original oracle output, verbatim:

```text
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestAssignmentProofWiderTargetMutant (0.07s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.04s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.05s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.04s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.09s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.10s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.05s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.08s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.11s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.08s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.06s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.10s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestStdoutColumnsRefusesWASI (0.87s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
FAIL
gate cache: native hits=3052 misses=6508
gate cache: node hits=1745 misses=6532
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	3774.929s
FAIL
```


### Main comparison


Created `/tmp/lower-39w-proof/main` as a detached worktree at the exact main SHA and initialized the same submodules using local clones. The primary checkout stayed clean at HEAD2. Ran:

```sh
go test ./internal/oracle -run 'TestAssignmentProof|TestStdoutColumnsRefusesWASI' -count=1 -v -timeout 40m
```


Exit 1. All thirteen tests fail for the same respective environment reasons:

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== RUN   TestStdoutColumnsRefusesWASI
=== PAUSE TestStdoutColumnsRefusesWASI
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChainedMutant
=== CONT  TestAssignmentProofReturnDefined
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.14s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.14s)
=== CONT  TestAssignmentProofIfNarrowingMutant
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.14s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.14s)
=== CONT  TestStdoutColumnsRefusesWASI
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.15s)
=== CONT  TestAssignmentProofWiderTargetMutant
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.05s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTargetMutant (0.06s)
=== CONT  TestAssignmentProofCompoundMutant
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.07s)
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.10s)
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.06s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.08s)
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/tmp/lower-39w-proof/main/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.09s)
=== NAME  TestStdoutColumnsRefusesWASI
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestStdoutColumnsRefusesWASI (0.50s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.738s
FAIL
```


### Focused external-package follow-up at HEAD2


Attempted `export NODE_PATH=/tmp/lower-39w-proof/api/node_modules`, then the command below. It still failed all twelve tests because assignmentProofTypecheck explicitly appends its own NODE_PATH after the inherited setting. This follow-up does not replace or erase the original full-suite result. No repository node_modules directory or link was added.


```sh
go test ./internal/oracle -run 'TestAssignmentProof' -count=1 -v -timeout 40m
```


Exit 1.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofIfNarrowingMutant
=== CONT  TestAssignmentProofReturnDefinedMutant
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.07s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.07s)
=== CONT  TestAssignmentProofCompoundMutant
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.09s)
=== CONT  TestAssignmentProofWiderTargetMutant
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.10s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.12s)
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTargetMutant (0.07s)
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.11s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.11s)
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.08s)
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.11s)
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.06s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.258s
FAIL
```


## Step 7 — admission delta


```sh
python3 cloud/admission-corpus/manifest.py --sha HEAD > /tmp/lower-39w-proof/manifest.json
go run ./cmd/adamic-admission-delta --base d43437079a7e7ca723168e67c5efb5bf7ebee216 --head 6d64f3f567755dda398b62809aa87b2973c962a9 --workers 2 --manifest /tmp/lower-39w-proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json
```


Ran alone, after all test/build processes had exited. Exit 0; wall seconds 1354.7222497379998.


```json
{
  "negative_witness_count": 50,
  "accepted_witnesses": 50,
  "node_agreements": 13,
  "runtime_notyet_count": 52,
  "ruled_runtime_stops": 52,
  "ruled_runtime_stop_paths": [
    "internal/oracle/testdata/intl_collator/in_compare.a",
    "internal/oracle/testdata/intl_collator/locale_en.a",
    "internal/oracle/testdata/intl_collator/locale_malformed.a",
    "internal/oracle/testdata/intl_collator/locale_variable.a",
    "internal/oracle/testdata/intl_collator/own_keys.a",
    "internal/oracle/testdata/intl_collator/tsc_shape.a",
    "internal/oracle/testdata/intl_collator/user_and_host.a",
    "internal/oracle/testdata/intl_collator_compare_value/p01_tsc_shape.a",
    "internal/oracle/testdata/intl_collator_compare_value/p01_tsc_shape.ts",
    "internal/oracle/testdata/reflect_get/stop_error_stack.a",
    "internal/oracle/testdata/reflect_get/stop_function_name.a",
    "internal/oracle/testdata/reflect_get/stop_hasownproperty_host.a",
    "internal/oracle/testdata/reflect_get/stop_host_collator.a",
    "internal/oracle/testdata/reflect_get/stop_host_dirent.a",
    "internal/oracle/testdata/reflect_get/stop_host_error_prototype.a",
    "internal/oracle/testdata/reflect_get/stop_host_fs_error_prototype.a",
    "internal/oracle/testdata/reflect_get/stop_host_groups.a",
    "internal/oracle/testdata/reflect_get/stop_host_handle.a",
    "internal/oracle/testdata/reflect_get/stop_host_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_host_iterator_next.a",
    "internal/oracle/testdata/reflect_get/stop_host_memory_field.a",
    "internal/oracle/testdata/reflect_get/stop_host_memory_heap.a",
    "internal/oracle/testdata/reflect_get/stop_host_performance_mark.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_ctime.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_date.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_field.a",
    "internal/oracle/testdata/reflect_get/stop_host_timeout.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_error_code.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_internal.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_literal.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_method.a",
    "internal/oracle/testdata/reflect_get/stop_keys_host.a",
    "internal/oracle/testdata/reflect_get/stop_keys_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_keys_memory.a",
    "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "internal/oracle/testdata/reflect_get/stop_method.a",
    "internal/oracle/testdata/reflect_get/stop_primitive.a",
    "internal/oracle/testdata/reflect_get/stop_private.a",
    "internal/oracle/testdata/reflect_get/stop_proto.a",
    "internal/oracle/testdata/reflect_get/stop_prototype_class.a",
    "internal/oracle/testdata/reflect_get/stop_string_host.a",
    "internal/oracle/testdata/timers_source/stop_foreign_generic.a",
    "internal/oracle/testdata/timers_source/stop_foreign_handle.a",
    "internal/oracle/testdata/timers_source/stop_foreign_intersection.a",
    "internal/oracle/testdata/timers_source/stop_foreign_record.a",
    "internal/oracle/testdata/to_primitive/stop_boolean_array.a",
    "internal/oracle/testdata/to_primitive/stop_record_function.a",
    "internal/oracle/testdata/to_primitive/stop_scalar_array.a",
    "internal/oracle/testdata/to_primitive/stop_static_class.a",
    "internal/oracle/testdata/to_primitive/stop_undefined_array.a",
    "internal/oracle/testdata/to_primitive/stop_union_numbers.a"
  ],
  "stale_runtime_stops": [],
  "pending_runtime_stops": [],
  "backend_divergence_count": 12,
  "ruled_backend_divergences": 0,
  "ruled_backend_divergence_paths": [],
  "corpus_filter": "",
  "shard": "",
  "complete_corpus": true,
  "phase_seconds": {
    "base_build": 51.694304566,
    "checkout": 17.069096905,
    "checkout_and_provision": 18.134815666,
    "classification": 822.02708035,
    "head_build": 25.439440935,
    "manifest_and_verification": 23.268308559,
    "provision": 1.065690879,
    "runtime": 409.821862637,
    "total": 1350.400538601
  },
  "workers": 2,
  "classification": "lowering-only",
  "diff_count": 11,
  "diff_ts_unmatched": [],
  "compile_timeout_seconds": 45,
  "runtime_timeout_seconds": 10,
  "base": "d43437079a7e7ca723168e67c5efb5bf7ebee216",
  "head": "6d64f3f567755dda398b62809aa87b2973c962a9",
  "generator_revision": "6d64f3f567755dda398b62809aa87b2973c962a9",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "757cbdbfb73376f7b2d10fbce6b423543492a5b4",
  "admitted": 0,
  "verdict": "pass",
  "void": false,
  "sampling_seed": "948b926c581e7141b7a9dbde365e29f0513b330a231c85c0c2dba9baa3d43587",
  "seed_given": false,
  "sampling_size": 115,
  "omitted": 0,
  "budget_seconds": 0,
  "budget_used_seconds": 409.821862637,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0,
  "class_counts": {
    "refused-by-both": 1953,
    "accepted-by-both": 3291
  },
  "program_count": 5244,
  "newly_refused": 0,
  "classification_stderr_differences": 0
}
```


Verdict PASS; void=false; no top-level or per-program void reasons. diff_count=11; admitted=0; newly refused=0. Every classification stderr was compared byte-for-byte between base_compile.stderr and compile.stderr: zero differ, so there are no base/head stderr text pairs to list. This classification gate does not replace the checked-mode runtime and direction pins in steps 3–4.


Changed programs:

```text
internal/lower/testdata/predicates_checked/false_read.a
internal/oracle/testdata/predicate_false_read/again.ts
internal/oracle/testdata/predicate_false_read/capture.ts
internal/oracle/testdata/predicate_false_read/early_exit.ts
internal/oracle/testdata/predicate_false_read/else_read.ts
internal/oracle/testdata/predicate_false_read/false_read.ts
internal/oracle/testdata/predicate_false_read/getter.ts
internal/oracle/testdata/predicate_false_read/getter_capture.ts
internal/oracle/testdata/predicate_false_read/no_read.ts
internal/oracle/testdata/predicate_false_read/renarrowed.ts
internal/oracle/testdata/predicate_false_read/spread.ts
```


Corpora:

```json
[
  {
    "name": "diff",
    "programs": 11
  },
  {
    "name": "witnesses",
    "programs": 2
  },
  {
    "name": "fixtures",
    "programs": 4296
  },
  {
    "name": "gaps",
    "programs": 105
  },
  {
    "name": "review",
    "programs": 20
  },
  {
    "name": "fuzz",
    "programs": 0
  },
  {
    "name": "stage3",
    "programs": 821
  }
]
```


Classification timeout errors: none at either SHA. Runtime timed-out runs=0, retried=0, still timed out=0. No timeout reruns were necessary.


```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 11
```


## Artifact integrity and publication


Only the three oracle-written .counts files were committed on the counts branch. This report is the sole file in the report commit, on new branch `proof/lower-39w-2-report` made directly from HEAD2. The only push destinations are `proof/lower-39w-2-counts` and `proof/lower-39w-2-report`. No source fixes, test edits, package installs, mutant edits, logs, or external worktrees are included in either commit.


External raw evidence SHA-256 values (the relevant full transcripts are embedded above):

- `/tmp/lower-39w-proof/admission.json`: `e209e35f7653cff1206784b7ff26b91b1aa82a07d93cb08abeacc8a2be7f5369`

- `/tmp/lower-39w-proof/manifest.json`: `c26fd0b817fe6de5c1b9e394118ea4aac4a671d703faba03de65725f8a70287f`

- `/tmp/lower.log`: `dd8ba2ef9047dbcf1136e8cc998e2201cc4957eea77fc4d3c1a219b0c3f14521`

- `/tmp/oracle.log`: `6d604170b38dfb41651d258025afd0df9abb5347d2f6cf057429fb401782338b`

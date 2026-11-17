# getProperties landing — #ev14ch9

Pushed `typescript/get-properties-1` at `6870a814639a6ae2745da20ce791483271a2d6ea`. Its single parent is the fetched origin/main tip `932a6344af63dc7c8728fffac05584c176d0475f`. The landing contains only the four files in stage3/adapt/99-correctness-get-properties and one appended stage3/upstream/LEDGER.md line (1 added, 0 removed). No evidence, census dump or Go file is in that commit.

The true overload claimed initialized elements no caller needs. Its true caller tests array nonemptiness; the false caller already consumes ordinary PropertyDeclaration values. Widening only the true overload result preserves the implementation, array identity and all runtime values. The adapter pins the complete three-declaration owner before and after, refuses drift and half-adapted/duplicate owners, preserves line endings, replays with zero writes and checks byte-identical commented and comment-free JavaScript.

The first ownership search found only an esnext.ts import in 99-correctness-receiver-transformers/owners.json. No adaptation pins getProperties whole, so this is a new directory rather than an amendment. The original source pin remains TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

## Required proof lines

| Proof | Result | Receipt |
| --- | --- | --- |
| stage3/apply.sh whole | PASS | whole-apply.json, exit 0, 626.21 s |
| TestProduct_Stage3AdaptedTree | PASS | adapted-tree-product.json, exit 0, 652.95 s; independent stock compiler/harness build |
| stage3/lane whole, unfiltered | PASS | whole-lane.json, exit 0, 625.08 s; all runners, eight workers, no filter |
| Restoring mutant red | PASS | mutant-red.json; all three original overload-refusal replays restored, exact baseline findings |
| Public typescript.d.ts byte-identical | PASS | public-declarations.json; actual adapted-base and candidate npm ci / Hereby dts builds |

The whole lane has 106,366 passing tests, one sanctioned Public APIs failure and zero pending. Its unchanged declaration guard passes. Complete apply/install/build/test logs, reports, baseline diff, failure details and verdict are in lane-results/. The ordinary lane was run without --artifact, without a filter, and with unchanged test expectations/timeouts. The wrapper stores each child's own return code.

Public typescript.d.ts SHA-256: `129e05979b0c292c75322d2446906a5bff055a00e10b5efbb1bed1f8bac38873` on both sides. tsserverlibrary.d.ts also matches, SHA-256 `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd`. Both complete before/after bundles are retained.

## Compiler ruling and fresh full latent census

The original 0-new gate correctly held the candidate. Compiler subsequently accepted it under the #ev14ch9 delegation: its units move on to the pre-existing transformers/utilities.ts:590:12 cast and binder.ts:2258:13 checked-field write. These sites are reached, not created. Only the true overload's type annotation changes; the two existing statements and every other compiler source are byte-identical. No cast, write, new shape or copy was introduced.

| File | Before | After | Newly reached events | Gone events | Changed events |
| --- | ---: | ---: | ---: | ---: | ---: |
| transformers/utilities.ts | 146 | 146 | 6 | 6 | 0 |
| transformers/ts.ts | 658 | 654 | 0 | 4 | 0 |
| transformers/classFields.ts | 633 | 624 | 2 | 11 | 1 |

Counts include Boundary bookkeeping. The owner has three declaration-unit replays of the reached cast. classFields.ts also has a changed Boundary that now reaches the existing core.ts filter-overload wall. The only newly reached production refusal sites are the two named by Compiler; analyze.py enforces that boundary. All new, gone and changed records are retained in delta-utilities.json, delta-ts.json and delta-classFields.json. Full raw before/after streams are latent-{base,candidate}-*.jsonl. Only raw start/end byte offsets are excluded from semantic event identity; original offsets remain in the streams.

Every run uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, the selected absolute LATENT_ONLY_FILE and CommonJSBundle. All streams finish with relation_report trailers and zero checker diagnostics. The fresh measurement binary was built from the fetched main's lowering, not reused from the old base. source-manifests.json records all 79 source inputs and its binary digest. The restoring mutant restores all 79 baseline sources exactly. The complete apply-order output matches all 79 measured candidate sources byte-for-byte.

## Shards and publication audit

No shard regeneration was needed. The runtime hash and actual installed stock-parser API digest match shard-plan.json; generate_shards.py --check passes. shard-plan.json, shards.json and shards_generated_test.go retain their original bytes. Their hashes and the input audit are in shard-inputs.json. The adaptation changes the dynamically keyed adapted-tree product; it does not change the plan's runtime/API/unit definitions.

publication-audit.json verifies the exact tested commit hashes, one-commit ancestry, permitted five-file diff, ledger +1/-0, all five successful proofs and a clean landing worktree. landing-push.json records the successful normal push; no force push was used. CHECKPOINTS.md retains the in-progress record, and the original gate-hold report remains in the parent evidence directory with a ruling banner.

Nothing remains blocked. No proof or acceptance predicate was weakened.

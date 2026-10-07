Built: rebased the helper branch onto current main, retained its published history, and preserved the blocked rule-rebase witness; no new helper claims.
SHAs: tested main e8ba3d5d; rebased helper commits ad6d3a75, published-history merge cb92e747; unchanged rule branch a7ff8948; final evidence is in this report commit.
Checks: owned helper oracle PASS 160.618s; options foundation PASS 113.350s; comments foundation PASS 132.302s; uncached input oracle PASS 46.611s; vet clean.
Mutants: all nineteen owned semantic mutants and ten foundation/adapter mutants caught again; original nineteen witnesses remain unchanged and are listed in REPORT.md.
Limits: rule rebase conflicts in four shared files outside this unit, and original six-consumer validation remains blocked by missing Tailwind/Kirk inputs; the landing cap remains closed.

## Branch audit and rebase

Fetched every origin head, then checked both published branches against origin/main. Neither published tip was landed nor based on current main e8ba3d5d81de4d3773c723914fccd4c76248b965. The helper branch had seventeen non-merge commits to replay, including required helper/inventory foundations absent on main. Its rebase completed cleanly, retaining current-main compiler and linter code. No compiler, harness or registration source was manually changed.

Local archive refs preserve the two published pre-rebase tips; they are not pushed. Because CLAUDE.md prohibits force-push, an explicit history-preserving merge of the previously published helper tip was added after the rebase. Its result tree is identical to the tested rebased tree; git diff HEAD^ HEAD was empty. Both current main and the old published helper tip are ancestors, allowing a normal fast-forward push to the same private branch. The rebased replay is therefore preserved without overwriting published history. Only own report/evidence files are added afterward.

The rule branch rebase in a separate worktree stopped immediately at the registration foundation's first commit, 29175443. Conflicts are in:

- stage1/cohere/lint/README.md
- stage1/cohere/lint/lint.ts
- stage1/cohere/lint/lint_test.go
- stage1/cohere/lint/testdata/oracle.go

Current main has additional monolithic rules and option handling while that foundation converts the driver to directory registration. Taking either side wholesale would discard one side's behavior. Ahra explicitly restricted this unit to its own rule directories and prohibited shared generator/harness edits. The complete conflict patch and original rebase output are preserved in evidence/landing/. The rebase was aborted cleanly, restoring a7ff8948 and preserving its published source/evidence. No rule-branch update is represented as rebased or green on current main.

The registration/harness owner or integration must reconcile that shared foundation with current main before this unit can replay the rule branch safely. The worker did not push main or any area branch. No subsequent helper is claimed while this cap is closed.

## Fresh external validation

All prior owned helper comparisons were rerun on current-main compiler code: 3,461 declaration-node cases, 3,463 private property-sort cases and 10,230 value-parser cases. All 17,154 match actual Go, source Node, emitted JavaScript and ASan/UBSan native. Every one of the nineteen independent semantic mutants compiles, executes with zero exit and no stderr, then fails only the external comparison. The allocation-alias, traversal-order and malformed-parser witnesses still catch their intended bugs.

Retained options/policy foundations and comment helpers also pass their unchanged oracles and all ten mutants, including the JSX adapter guard's required refusal. The filtered uncached input oracle covers six fixtures with six probe misses and zero cache hits. Vet output is empty. Full repository gate, whole-rule findings/fixes and original consumer parity are not claimed.

The exact original six consumer commands were rerun through the oracle-only capture overlay on the rebased branch. Canonical, class-order, conflicting and unknown suites exit 1 through fixture guards; variant-order and shorthand exit 0 with skips. All three helper call counts are zero in every consumer. The missing installed tailwindcss and /Users/kirkouimet/Projects/ahra/app/_theme/styles/theme.css blockers remain. This is not a live consumer pass and retains zero confirmed readiness credit.

The separate file-loader gap probe remains successful as a gap detector: valid ASCII/Unicode agree, invalid bytes 80/81/ff become ef bf bd on all three Adamic paths while actual Go theme ingestion preserves the original byte. This does not certify a file-loader implementation; its feasibility reservation remains withdrawn.

## Commands and artifacts

```sh
git fetch origin '+refs/heads/*:refs/remotes/origin/*' > /tmp/wave109-landing-fetch.log 2>&1
git rebase origin/main > /tmp/wave109-helper-rebase.log 2>&1
bash cloud/setup.sh > /tmp/wave109-landing-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_09 -count=1 -v -timeout=15m > /tmp/wave109-landing-helper-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/comments -count=1 -v -timeout=20m > /tmp/wave109-landing-foundation-oracles.log 2>&1
go vet ./stage1/cohere/lint/helpers/... > /tmp/wave109-landing-helper-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/wave109-landing-input-oracle.log 2>&1
python3 stage1/cohere/lint/helpers/from_wave1_09/testdata/capture_consumers.py /tmp/wave109-landing-live > /tmp/wave109-landing-live-capture.log 2>&1
git fetch origin main > /tmp/wave109-landing-final-main-fetch.log 2>&1
```

Setup timing: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 163s, total 163s; nproc 5. Go 1.27.1, clang 20.1.8, Node 24.19.0. Timing includes concurrent setup/test CPU contention and is not a throughput benchmark. Final fetch confirmed the same tested main SHA. Complete logs, conflict witness and machine-readable status are in evidence/landing/. The untracked shared .generated cache remains excluded from commits.

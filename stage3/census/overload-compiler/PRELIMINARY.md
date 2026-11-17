# Evidence disposition

REPORT.md is the final decision and proof matrix. final-callers-decoded.json is the authoritative completed-tree caller audit; its changedDuringAudit list is empty.

upstream-callers.json is an empty output from an aborted preliminary upstream audit. upstream-audit-error.stderr and upstream-audit-error2.stderr retain checker failures. adapted-callers.json and src-callers.json were initial audits while baseline application was in progress. final-callers.json used an incorrect UTF-8-only consistency comparison and flags seven encoded fixtures; it is superseded by the decoder-correct audit. None of these preliminary products is acceptance proof.

builder-trial.json, builder-trial2.json and extensions-trial.json are initial 77-root upstream trials before generated inputs. Their preparation diagnostics are retained, not treated as passing 79-root checks. Completed-tree stock-*.stdout receipts contain the 79-root diagnostic counts. tsc-cli-extensions-union.* is the separate actual tsc CLI receipt.

latent-extensions-*.jsonl streams are deliberately incomplete following failed scan gates. Each has a direct-child return code -15 in its run receipt and no relation-report trailer. They must not be interpreted as empty/full censuses. stopped-measurements.json records why they ended.

builder-candidate.diff and its owner/mutant files describe a rejected, uncommitted worktree candidate. No landing commit or ledger entry exists. The candidate remains locally reviewable at /workspace/scratch/ev14ch9/builder-work, whose HEAD is the requested base.

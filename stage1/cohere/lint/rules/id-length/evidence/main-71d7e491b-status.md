Rebased cleanly onto origin/main 71d7e491b. Area remains bb2ece564; dedup decisions unchanged. Zero shared-file conflict hunks or new shared edits.

Ran `go test ./stage1/cohere/lint -run '^TestMutants/debugger_fix_suppressed$' -count=1` with the configured environment. FAIL in 3.273s before mutant execution: restricted-types Go decoder rejects foreign `allowLoop` from the generated all-rule options row. Adjacent log preserves the reproducer. No mutant credit, full gate unverified, no new helper claims. Stopped under the existing shared-blocker instruction.

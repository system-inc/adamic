Built: rebased the existing deduplicated rule branch onto origin/area/stage1-lint at 7481e032 after reading its unchanged DEDUP_LEDGER.md; no new claims or shared edits.
SHAs: area 7481e0324e34a2537aafa9db7eeacda50405611b; current main 39638d9e278d38bb5aeae887f46d55a70e47aaad; tested rule history merge 749e45325c654e042124692c12123a4ca6b5bf37.
Checks: registry PASS 0.066s; uncached TestRulesAgree FAIL 3.125s at typescript-no-empty-object-type/rule.ts:27:58, nonconstant RegExp pattern.
Mutants: registry descriptor rejection checks passed; rule semantic mutants were not rerun because native compilation is refused.
Limits: compiler gap prevents green or parking; no new helper claims, whole-rule parity, or full repository gate.

The landed area is newer than the requested 50a5f105 and contains the harness 41eb6eab2 plus current main. The four losing ports remain removed in favour of the ledger winners; six owned candidates remain. The ledger is byte-identical to the previously read ledger and assigns no batch-only rule specifically to this unit. The seven-argument report contract is unchanged; no shared finding, context, main, generator, oracle, comparison harness or compiler implementation was edited.

Ran go test ./stage1/cohere/lint/registry -count=1 -v and ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=15m. Logs are committed in evidence/area-7481. The compiler refusal is unchanged despite the harness landing; the option regex already uses new RegExp(pattern, 'u') with no matcher fallback.

Owned commits were replayed onto the area, then the existing dedup cleanup was replayed. An explicit ours merge retains the prior published tip without changing the rebased tree; push is a normal fast-forward to the same private branch. Current main is an ancestor. No main or area branch is pushed. The temporary cohere dependency link used for tests was restored before committing.

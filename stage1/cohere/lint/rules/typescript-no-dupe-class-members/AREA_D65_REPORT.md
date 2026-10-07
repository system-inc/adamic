Built: replayed the deduplicated existing rules onto latest lint area d65a8f93, retaining its runtime changes and unchanged ledger winners; no new claims.
SHAs: area d65a8f931c98655936ae04c6899f38f14862b73e; main 39638d9e; helper branch unchanged 649214b04; previous published rule tip 609bbf0d0 retained in history.
Checks: registry PASS 0.072s; uncached TestRulesAgree FAIL 3.085s at empty-object-type/rule.ts:27:58, nonconstant RegExp pattern.
Mutants: registry descriptor rejection checks passed; rule semantic mutants were not rerun after compilation was refused.
Limits: compiler gap still blocks landing and parking; no new helper claim, whole-rule parity or full gate.

Ran go test ./stage1/cohere/lint/registry -count=1 -v and ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=15m, with output preserved in evidence/area-d65. The dedup ledger has no differences from 7481e032; the four losing ports remain removed and six owned candidates remain. Runtime-profile area changes are retained without editing shared files. Options still use new RegExp(pattern, 'u'), with no matcher fallback. The rebased tree retains old published history by an explicit ours merge and is pushed normally to the same owned branch. No main or area branch is pushed.

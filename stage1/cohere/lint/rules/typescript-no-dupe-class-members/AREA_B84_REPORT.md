Built: rebased the existing deduplicated rules onto current lint area b84a9d93, preserving its proof-lowering and runtime changes; no new claims or shared source edits.
SHAs: area b84a9d9314b65d3d0261ee017e233287b4f071da; main c7991b900362796aefd111474e65eb5398e91953; previous published rule tip 3fd4e449a retained in history.
Checks: registry PASS 0.069s; uncached TestRulesAgree FAIL 6.107s at empty-object-type/rule.ts:27:58, nonconstant RegExp pattern.
Mutants: registry descriptor rejection checks passed; rule semantic mutants were not rerun after compilation refusal.
Limits: compiler gap still blocks landing; original consumer fixtures remain unavailable; no full gate or new helper claims.

Ran go test ./stage1/cohere/lint/registry -count=1 -v and ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=15m, with logs preserved in evidence/area-b84. The ledger is unchanged and its four losing ports remain removed. No check was skipped, relaxed or removed to obtain a pass. The option pattern still uses new RegExp(pattern, 'u'); no matcher fallback was added. An explicit ours merge retains the published history without changing the rebased tree and permits normal fast-forward push. No main or area branch is pushed.

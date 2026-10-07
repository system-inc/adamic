Built: rebased existing deduplicated rules onto lint area b4691483 after its legacy registry migration; no new claims or shared source edits.
SHAs: area b46914832d70e00847d82d5d221ab7bb24040c53; main c7991b90; helper branch remains f00d469d9; prior rule tip 158bd57de retained in history.
Checks: registry PASS 0.093s; uncached TestRulesAgree FAIL 6.948s at empty-object-type/rule.ts:27:58, nonconstant RegExp pattern.
Mutants: registry descriptor rejection checks passed; rule semantic mutants not rerun after compilation refusal.
Limits: compiler gap still prevents landing; consumer fixtures remain unavailable; no new helper claims or full gate.

Ran go test ./stage1/cohere/lint/registry -count=1 -v and ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=15m. Logs are preserved in evidence/area-b469. The dedup ledger is unchanged; the four losing ports remain removed. All landed legacy registry, harness and runtime changes are retained. No shared finding, context, main, generator, oracle, comparison harness or compiler source was manually changed. The option regex remains new RegExp(pattern, 'u'), with no matcher fallback.

Published history is retained by an explicit ours merge without changing the replayed tree, allowing a normal fast-forward push to the owned rule branch. No main or area branch is pushed. No correctness check was skipped, relaxed or deleted to obtain a pass. The helper branch's main and compiler revision are unchanged; its previously recorded 17,154 four-way cases and 29 mutant checks were not rerun this turn.

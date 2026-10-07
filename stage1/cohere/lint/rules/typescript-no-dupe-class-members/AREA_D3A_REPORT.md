Built: rebased existing deduplicated rules onto current lint area d3a37422, retaining the typeof-null landing; no new claims or shared edits.
SHAs: area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898; main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06; old rule tip eaf526191 retained in history.
Checks: registry PASS 0.097s; uncached TestRulesAgree FAIL 7.908s at empty-object-type/rule.ts:27:58, nonconstant RegExp pattern.
Mutants: registry descriptor rejection checks passed; rule semantic mutants not rerun after compilation refusal.
Limits: compiler gap still prevents landing; consumer fixtures remain unavailable; no full gate or new claims.

Ran go test ./stage1/cohere/lint/registry -count=1 -v and ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=15m; all output is recorded in evidence/area-d3a. The ledger is unchanged and the four losing ports remain removed. Configurable adapters decode field 5; the two optionless upstream adapters return nil. No option guard was bypassed. The required option RegExp remains new RegExp(pattern, 'u'), with no fallback matcher.

Shared compiler, registry, finding, context, main and oracle implementation files were not manually changed. No check was weakened, skipped or deleted to produce a pass. The prior published tip is retained by an explicit ours merge without a replayed-tree change, so push is a normal fast-forward to the private branch. No main or area branch is pushed.

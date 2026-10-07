Built: rebased existing helpers onto current main after typeof-null fixes; no new claims or shared source edits.
SHAs: main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06; rules pushed 317ccb0a2 on lint area d3a37422; prior helper tip f00d469d9 retained in history.
Checks: uncached owned helper oracle PASS 67.342s; options foundation PASS 49.708s; comments foundation PASS 101.354s; helper vet clean.
Mutants: all 19 owned semantic mutants and 10 foundation/adapter mutants caught again by the external comparisons.
Limits: rule dynamic RegExp lowering and original consumer fixture gaps remain; no full gate, new claim or readiness credit.

Ran ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_09 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/comments -count=1 -v -timeout=20m and go vet ./stage1/cohere/lint/helpers/..., with logs recorded in evidence/landing-b6b. All 17,154 owned cases match actual Go, source Node, emitted JavaScript and ASan/UBSan native. The invalid-byte loader probe passes as a gap detector, not loader parity.

Main's typeof-null changes are retained unchanged. Existing owned commits were replayed; an explicit ours merge retains published history without changing the replayed tree, enabling normal fast-forward push to the private branch. No shared compiler or harness source was manually changed. No main or area branch is pushed.

The rule branch also rebased onto current lint area, with the unchanged dedup ledger applied: registry PASS 0.097s, TestRulesAgree FAIL 7.908s at empty-object-type/rule.ts:27:58 because dynamic RegExp patterns cannot lower. Configurable adapters already decode field 5, and no guard was bypassed. No check was weakened, skipped or deleted to obtain a pass. Original six-consumer fixture failures were not rerun or counted as passes; their confirmed readiness credit remains zero. New helper claims remain paused.

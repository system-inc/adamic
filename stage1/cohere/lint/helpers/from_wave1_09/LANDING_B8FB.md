Built: replayed the existing helper branch onto current main after the inherited static-field landing; no new claims or shared source edits.
SHAs: main b8fb957aa839a9e8cb0b54279dd9864fa317bd30; tested helper history merge 91e61eee35702ce244b152eb82ac24042ce67c23; rules remain published at ea18917c0.
Checks: uncached owned helper oracle PASS 65.733s; options foundation PASS 52.554s; comments foundation PASS 101.331s; helper vet clean.
Mutants: all 19 owned semantic mutants and 10 foundation/adapter mutants caught again by external comparisons.
Limits: dynamic RegExp lowering and missing consumer fixtures remain; no full gate, new rule oracle, or new readiness credit.

Ran ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_09 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/comments -count=1 -v -timeout=20m and go vet ./stage1/cohere/lint/helpers/..., with output recorded in evidence/landing-b8fb. All 17,154 owned cases match actual Go, source Node, emitted JavaScript and ASan/UBSan native. The invalid-byte file-loader probe remains a gap detector, not loader parity.

The helper replay retains every current-main change and has no compiler-file diff against main. An explicit ours merge of the previously published tip preserves history without changing the replayed tree, allowing a normal fast-forward push. No main or area branch is pushed.

The rules were already rebased onto the named unified harness ab70f38d4 and pushed at ea18917c0. Its uncached rule oracle remains blocked at empty-object-type/rule.ts:27:58 by nonconstant RegExp lowering, which is still refused in current main. The option path uses new RegExp(pattern, 'u') and no matcher fallback. This is outside the harness-only parking exception. No new helper claim is made while that cap is closed. Original six-consumer fixture blockers remain documented in LANDING_REPORT.md; confirmed fully unblocked rule count remains zero.

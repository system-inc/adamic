Built: replayed the existing helper branch onto current main and preserved its published history; no new claims.
SHAs: main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; tested helper history merge c2f31b4653ff2267f53bce47b4abebc8730a3774; unchanged rules a7ff8948.
Checks: uncached owned helper oracle PASS 61.561s; options PASS 48.903s; comments PASS 97.276s; helper vet clean.
Mutants: all 19 owned semantic mutants and 10 foundation/adapter mutants caught again by the existing external comparisons.
Limits: dynamic RegExp lowering, missing consumer fixtures and four shared rule-rebase conflicts remain; parking is not certified.

Command: ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_09 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/comments -count=1 -v -timeout=20m. Test output and rebase logs are preserved in evidence/landing-c019. Also ran go vet ./stage1/cohere/lint/helpers/.... No full gate or new filtered input run.

The 17,154 helper cases agree with actual Go on source Node, emitted JavaScript and ASan/UBSan native. The invalid-byte loader probe still passes as a gap detector, not loader parity. The original six live consumer fixture blockers remain unchanged; those tests were not rerun, and no new readiness credit is claimed.

The newly fetched origin/codex/lint-regex documents the same nonconstant RegExp compiler refusal in regex/gaps.md. The owned empty-object-type option path already uses new RegExp(pattern, 'u'); no matcher fallback was added. The parking exception requires the shared harness to be the only landing obstacle, which is not the case here. Rule rebase was attempted against current main, hit README.md, lint.ts, lint_test.go and testdata/oracle.go conflicts, and was aborted without shared edits. Shared node-handoff integration remains pending.

Replayed the prior clean helper chain, then its owned report commits, and retained the previous published tip with an explicit ours merge whose result tree is unchanged. Push is a normal fast-forward to the owned helper branch. Main's landed changes are retained without modification; no main or area branch is pushed. Each helper remains a candidate prerequisite for six better-tailwindcss rules, with zero confirmed fully unblocked rules.

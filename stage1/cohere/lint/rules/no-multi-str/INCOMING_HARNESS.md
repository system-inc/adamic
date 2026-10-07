# Announced harness follow-up

Current main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Both owned pushed branches already contain that main and their recorded oracle checks remain unchanged. Announced harness ab70f38d47de1d4974082b38f84a56af2368b7af on origin/lint-rules/harness is not an ancestor of main yet. No shared file was edited on either owned branch, and neither main nor an area branch was pushed.

The announced harness includes parser changes beyond registration. Restored its 36 changed stage1/typescript paths only into the existing scratch comparison checkout. All eighteen owned rule sources, current-main compiler/runtime and owned Go comparison adapter remain unchanged. This probe is forward-looking scratch evidence, not integration certification of the new harness or a rebase onto it.

```
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/wave08-park-overlay/overlay.json ./stage1/cohere/lint -run '^(TestWave08Shapes|TestOriginalJSX|TestSixthRecovery)$/^TSX$' -count=1 -v -timeout=10m > /tmp/wave08-incoming-parser-gaps.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/wave08-park-overlay/overlay.json ./stage1/cohere/lint -run '^(TestOriginalFull|TestSixthFull|TestSixthRecovery)$' -count=1 -v -timeout=15m > /tmp/wave08-incoming-full-gaps.log 2>&1
```

The first command PASS in 83.218s. The JSX gating witness matches Go on source Node, emitted JavaScript and ASan/UBSan native: 649 equal bytes. The Tailwind JSX shape matches on those same executions: 481 equal bytes. The slash filter selects no recovery child cases in this first command; its top-level recovery PASS is not recovery coverage.

The second command runs those recovery children without filtering and exits 1 in 96.365s. The full original batch captures 224 source/rule/options cases, but source Node exits 70 on the recovered computed key `({ ['x' });`: parser slice expected CloseBracketToken, got CloseBraceToken at 8. The full escape/octal batch captures 83 cases, but source Node exits 70 on a malformed legacy number. All three explicit probes fail on source Node:

- `var a = 01.5;`: expected semicolon at 10.
- `var a = 0777.5;`: expected semicolon at 12.
- `var a = 0755n;`: expected semicolon at 12.

These are observable parser refusals, not rule findings disagreements or silent successes. Comparison stops on the Node refusal, so no emitted/native parity is claimed for these failing inputs. The earlier supported-corpus and eighteen comparison-only mutant checks remain under PARKING.md; no new rule logic or check was introduced and no new mutant result is claimed here.

Full upstream coverage still has parser blockers beyond missing shared registration. Consequently the shared-harness-only parking exception is not asserted, and no helper is claimed. Fixing the parser is outside this unit's ownership. Once the shared harness lands on main, rebase and rerun before taking new work.

Read the helper README and readiness.json from origin/codex/lint-helpers and inspected all origin helper claim files after fetching all heads. Audit data is recorded below. No availability assertion or claim is made while the landing cap remains unresolved. The regex translation table on origin/codex/lint-regex has no Go regex site for any of these eighteen rule names. No regex translation or handwritten regex matcher was added in this follow-up. Future actual Go regex sites must use its literal translations, and dynamic option patterns use new RegExp(pattern, 'u').

Claim audit: 20 origin branches with helper claims; 20 claim file occurrences. All origin references were enumerated by the audit, including branches with no helper claims.

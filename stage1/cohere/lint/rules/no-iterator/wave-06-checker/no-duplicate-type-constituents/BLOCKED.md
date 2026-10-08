# @typescript-eslint/no-duplicate-type-constituents

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`. The checker questions needed by this rule are available; this is a shared fixer/parser blocker.

Reproducer: `blocked.ts.txt`. Go's case is at `cohere/internal/lint/rules/typescript/no_duplicate_type_constituents_test.go:383`. Its initial edits produce `blocked-fixed.ts.txt`. While converging, `Linter.fixed` reparses this intermediate at `stage1/cohere/lint/lint.ts:198` and runs it at line 213. The parser refuses at `stage1/typescript/parser/parser.ts:499`: `expected CloseParenToken, got SemicolonToken at 31`. This refusal is observed in `proof/checker-before-parking/gate.jsonl` under `TestRulesAgree`; it is not an inference that the original input is invalid.

Shared checker call `ctx.TypeChecker.GetTypeAtLocation(constituent)` at `cohere/internal/lint/rules/typescript/no_duplicate_type_constituents.go:280` is correctly served by `raw-shape`. The native witness and 41 upstream combinations matched Go, Node and emitted JavaScript; native parity runs used sanitizers. The witness's text-verdict mutant compiled and exited successfully, then disagreed with Go on native, Node and emitted JavaScript: `proof/partial-mutant.log`.

All 102 unique upstream combinations were captured, but only the first 41 completed comparison. No all-case parity is claimed. The pending descriptor is archived as `rule.pending.json` and the rule is not registered. Shared parser, fixer, checker, harness and registry code were not edited.

Separate observation: the fixer's next Linter is constructed without a checker at `lint.ts:200-210`; any further typed fixing support needs shared review. This was not the first refusal and no workaround was attempted.

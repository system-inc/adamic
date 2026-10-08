The no-useless-rename port uses the shared comments.forFile implementation and the existing RuleContext token, parent and reporting adapters. Its oracle adapter returns the unchanged pinned Go rule with NoUselessRenameSettings.

170 unique captured upstream source/rule/options combinations and three owned witnesses match Go on source Node, emitted JavaScript and sanitized native. Witnesses cover declarations, assignments, imports, exports, options, contained comments, parenthesized defaults, nested targets and for-in/for-of/for-await targets. The inherited compiler/stage1 corpus and profile snapshots also match.

Go behavior preserved by name: `export {'foo' as foo} from 'bar'` keeps the quoted first name; `import {'foo' as foo} from 'foo'` keeps the second name. `({foo: (foo) = a} = obj)` reports without a fix.

Validation and evidence:

- [Registry](evidence/registry.log): `go run ./cmd/lint-registry`.
- [Selected tests](evidence/selected.log): `go test ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/no-useless-rename' -count=1 -v -timeout 30m`; PASS, 142.099s.
- [Mutant](evidence/selected.log): the compiling import-side mutation disagrees with Go on Node (line 575), emitted JavaScript (line 587), and native (line 599).
- [Captured cases](evidence/upstream.jsonl): 170 unique cases.
- [Complete package events](evidence/whole.jsonl) and [machine summary](evidence/whole-summary.json): PASS; 114 passing test/subtest events, 0 failures, 1 skip (34 top-level passes, 0 failures, 1 skip); wall 1681.034s; nproc 5; load before 0.16/0.92/1.59, after 2.52/2.35/2.30.
- [Inputs](evidence/inputs.sh): clean TypeScript v6.0.3 checkout at 050880ce59e30b356b686bd3144efe24f875ebc8, WASI SDK 27 from cloud/setup.sh --wasi-sdk, benchmark enabled, and one fresh directory for both profile variables. Setup environment: /workspace/adamic-tools/env.sh.

The sole skip is the inherited unconditional TestCheckerBridgeRefusalPending at ../../checker_pending_test.go:49, awaiting codex/tsgo-errors-as-values. No rule-specific helper or language blocker remains. The completed whole-package run used unchanged final source; earlier incomplete runs were interrupted for the for-await correction and by an execution-server reset.

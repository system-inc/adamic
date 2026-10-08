# no-throw-literal shared typed recovery gap

Observed on d845dccde with all 48 captured upstream combinations. TestRulesAgree fails before comparing `function f() { throw; }`: Go oracle panics `invalid corpus ... [Expression expected.]`. The source was captured with unsupported-recovery metadata; the typed branch at lint_test.go:377-382 passes it into compareWithJavaScript without the syntax branch's recoveryRows/checkRecoveryRefusal classification. The rule's missing-expression guard is no_throw_literal.go:92-97. Its symbol questions are answered and the native port compiles, but all-case parity is blocked on shared typed recovery handling. No harness change or skipped case was introduced.

Reproducer: `blocked.ts.txt`. Source and unregistered descriptor are archived here. This rule is excluded from the landing registry.

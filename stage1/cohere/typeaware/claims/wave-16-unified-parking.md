# Wave 16 unified landing boundary

The private typeaware ports are not certified by the unified lint harness.
They remain on codex/typeaware-wave-16 as drafts, not on the new area-based
prefer-destructuring landing branch. No TestOwnedWitnesses/TestMutants/
TestRulesAgree pass is claimed for these private ports. Step 1 is skipped:
their checker-backed Rules/Context adapters cannot be used as RuleContext;
RuleContext has no program/checker handle or private fact transport.

Private-adapter integration parked: no-import-assign, prefer-exponentiation-operator,
use-isnan, no-global-listener-target-assertion, no-leaked-number-render,
no-mock-on-module-namespace, no-interpolated-shell-command,
no-interpolated-sql-string, no-alert, no-new-func,
no-new-native-nonconstructor, no-new-wrappers, no-throw-literal,
no-useless-backreference, prefer-arrow-callback.
Their existing private adapter imports Rules from typeaware/rules.ts and
requires its program, path, parser/scanner and checker-fact APIs. They need
unified context adapters before any unified green claim; their private parity
is preserved in AREA_D3_LANDING.md, not substituted for certification.

Parser/checker integration parked: react/jsx-fragments, react/jsx-no-undef,
react/jsx-no-constructed-context-values. Exact full-arena reproducer:

```
/workspace/wave16-artifacts/seventh-gate/adamic build  /workspace/wave16-artifacts/seventh-gate/full-arena-source/suite.a  -o /workspace/wave16-artifacts/seventh-gate/foreign-gap  --tsgo /workspace/wave16-artifacts/seventh-gate/checker.a
```

Observed exit 1, stage1/typescript/parser/parser.ts:41:20: escaping this
before every field is set. The standalone foreign-arena probe exits 0.
The complete raw reproducer and owned draft Context live in
wave16_seventh/context_foreign_gap.a and gaps/foreign_arena.a; validate.py
constructs the full graph. Cause is unestablished.

Previously parked: react-hooks/set-state-in-effect, set-state-in-render,
static-components; native HIR/SSA/capture analysis required, as recorded in
wave-16.md. No private descriptor is added to unified lint discovery.

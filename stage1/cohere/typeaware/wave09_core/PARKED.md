# Wave 09 parking for unified-harness landing

No owned rule is certified on the unified harness yet, so step 1's landing
branch is skipped. The isolated native comparisons remain evidence only;
they do not substitute for TestOwnedWitnesses/TestMutants/TestRulesAgree.
The separate syntax-only nexus port starts from origin/area/stage1-lint and
contains none of these implementations.

- no-invalid-regexp: dynamic RegExp lowering and shared checker integration.
  Reproducer: invalid_regexp/dynamic_regexp_gap.a, line 3. Native build
  refuses `RegExp with a nonconstant pattern`; source Node returns true.
- no-misleading-character-class: the same dynamic pattern dependency and
  unfinished constructor visitor/tracker. Reproducer: the same dynamic
  constructor probe; its supported literal slice is not full certification.
- no-label-var: the unified RuleContext does not expose the native checker
  scope-value-symbols question. Its isolated driver imports tsgoProgram and
  tsgoInspect; source Node cannot import those from oracle/adamic.mjs.
  Reproducer: label_var/main.a through `node oracle/node.mjs` with the
  manifest/config from label_var/verify.py.
- prefer-const, radix and @typescript-eslint/non-nullable-type-assertion-style:
  their native checker driver has the same unified checker boundary.
  Reproducer: stage1/cohere/typeaware/wave_09_suite.a through
  `node oracle/node.mjs`, using the controls config/manifest emitted by
  TestWave09AgreementAndMutants. Its tsgoProgram import is absent from the
  Node runtime. Typed/scope/global judgments must not be replaced by syntax.

See LANDING_TYPEOF_REPORT.md and FLAG_OPTIONS_REPORT.md for passing native
scopes, mutants, sanitizers and limits. No blocked rule is placed in the new
unified-harness branch, and no shared checker/runtime source is edited.

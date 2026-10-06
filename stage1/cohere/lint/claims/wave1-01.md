# Lint wave 1, slot 01

Branch: codex/lint-wave1-01. Base: origin/main d090af5.

The helper report links the ordered 46-rule handoff in ../HELPERS.md.
Positions 1, 2 and 3 are:

1. @typescript-eslint/consistent-type-assertions: claimed for investigation.
2. @typescript-eslint/consistent-type-definitions: skipped, already implemented in
   origin/codex/stage1-lint-batch4-typescript at
   63782c53678711717f9fd4f12762ce3d103b9a3d,
   stage1/cohere/lint/consistent_type_definitions.ts.
3. @typescript-eslint/init-declarations: skipped, already implemented on that
   branch in stage1/cohere/lint/init_declarations.ts, also on
   origin/codex/stage1-lint-batch4 at d486b03a15f3202acc317dc81c40cc21818e21f7
   in stage1/cohere/lint/ts_init_declarations.ts.

No rule implementation has been written before this claim commit and push.
Existing branch implementations are observations, not parity certifications by
this worker. The remaining rule is not declared ported.

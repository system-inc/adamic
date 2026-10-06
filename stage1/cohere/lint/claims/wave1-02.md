# Lint wave 1, slot 02

Branch: `codex/lint-wave1-02`.

The helper report links the ordered list in `stage1/cohere/lint/HELPERS.md`.
Positions 4, 5 and 6 in that list are:

1. `@typescript-eslint/no-explicit-any`: skipped, already ported on
   `origin/codex/stage1-lint-batch5` (`dcef3eebfddcb297464bdce398d850775852587d`),
   in `stage1/cohere/lint/no_explicit_any.ts`.
2. `@typescript-eslint/no-inferrable-types`: skipped, already ported on
   `origin/codex/stage1-lint-batch4-typescript`
   (`63782c53678711717f9fd4f12762ce3d103b9a3d`),
   in `stage1/cohere/lint/no_inferrable_types.ts`.
3. `@typescript-eslint/no-restricted-types`: claimed by slot 02; no matching
   stage1 port filename found across fetched origin branches.

This claim is published before any rule implementation.

## Foundation blockers observed before implementation

Base: `origin/main` at `d090af531216ddd3c25a0dede6b82d7c0a6edf76`.
Registration: `48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8`, merged successfully.
Helpers: `5d13f5baaecaf11d4ea62de693426f69a1f41bba`, attempted merge conflicts
in README.md, lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go
under stage1/cohere/lint. That merge was aborted, preserving both foundations.

The registration generator requires rule.ts and rejects mutant files without
a .ts suffix. This conflicts with the unit's requirement that new Adamic files
are .a and with the rule-directory-only ownership boundary. No shared generator
or dispatch file is changed by this unit. docs/parallel-work.md is absent on
main and on both requested foundation refs. Implementation is blocked pending
compatible foundations; this claim does not assert a completed port.

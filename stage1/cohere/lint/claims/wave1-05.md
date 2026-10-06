# Lint wave 1 slot 05

Branch: codex/lint-wave1-05.

Positions 13, 14 and 15 in the helper handoff list in HELPERS.md, linked from helpers/REPORT.md:

- max-depth: skipped, already ported as max_depth.ts on origin/codex/stage1-lint-batch4.
- max-nested-callbacks: skipped, already ported as max_nested_callbacks.ts on origin/codex/stage1-lint-batch4.
- no-cond-assign: claimed for this worker, under rules/no-cond-assign/.

The registration foundation merged cleanly. The helper foundation conflicted in six shared lint files; the directory registration versions were retained, with helper-owned files brought in. docs/parallel-work.md is absent; docs/lint-registration.md supplies the registration contract.

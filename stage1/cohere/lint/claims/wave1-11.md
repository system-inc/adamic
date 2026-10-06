# Lint wave 1 slot 11

Branch: codex/lint-wave1-11.

The helper report points to the ordered list in ../HELPERS.md. Assigned positions:

31. `adamic/no-definite-assignment`: claimed for a new directory port.
32. `base/boundary-no-global-container`: skipped, already ported on origin/codex/stage1-lint-batch4 in base_boundary_no_global_container.ts.
33. `base/consistency-no-hand-built-declared-error`: skipped, already ported on origin/codex/stage1-lint-batch4 in base_consistency_no_hand_built_declared_error.ts.

Existing-port evidence: batch 4 implementation 7bd94a2f92cc17d12bb69bbd78bfe2c4ac4ca63d and its BATCH4.md report. This claim is pushed before new rule code.

Foundation merge note: helpers conflicts with registration in six shared files. Resolution retains directory registration and imports the helper package, inventory and HELPERS.md; it does not restore the helper branch's legacy shared dispatcher. docs/parallel-work.md is absent in main and both foundation branches; docs/lint-registration.md provides the available contract.

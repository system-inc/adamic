# Lint wave 1 slot 14

Branch: codex/lint-wave1-14. Positions are the 30 option-ready rules followed by
16 policy-ready rules in HELPERS.md, as linked by helpers/REPORT.md.

40. nexus/consistency-no-stuttering-name: skipped, already ported on origin/codex/stage1-lint-batch4 (nexus_consistency_no_stuttering_name.ts).
41. nexus/consistency-no-utils-folder: skipped, already ported on origin/codex/stage1-lint-batch4 (nexus_consistency_no_utils_folder.ts).
42. nexus/import-require-module-alias: claimed for this worker. No implementation found on fetched origin branches.

Foundation merge note: lint-registration merged cleanly; lint-helpers had six
shared-driver conflicts from its scanner-branch ancestry. The resolution retains
the directory registry versions of those shared files and all helper additions.
The requested docs/parallel-work.md is absent on main and both foundations; read
its committed version at 7f958de on origin/codex/no-shared-lists.

## Next three rules

Fetched all 305 origin refs after pushing the existing branch. All 46 helper-ready
rules are represented in origin claim records. Continuing in inventory order from
the syntax ready for AST/API adaptation list at origin/codex/lint-inventory.

- @typescript-eslint/no-confusing-non-null-assertion: claimed.
- @typescript-eslint/no-duplicate-enum-values: claimed.
- @typescript-eslint/no-dynamic-delete: claimed.

These are the first three entries neither implemented on origin/main ef3d907e
nor named by any direct claim Markdown file on fetched origin branches.
This update is committed and pushed before implementation.

### Next-three outcome

No-duplicate-enum-values and no-dynamic-delete are implemented as directory-owned
.a ports and passed four-way fixture/corpus parity and compiling semantic mutants.
No-confusing-non-null-assertion stays claimed but blocked: the shared Finding and
oracle serializer cannot carry its suggestion ranges, arrays and multiple edits.
The exact independent probes and limits are recorded in
../rules/typescript-eslint-no-duplicate-enum-values/REPORT.md.

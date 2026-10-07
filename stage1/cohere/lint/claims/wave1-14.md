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

## Third batch

All helper-ready names remain represented in fetched origin claims. The next
three syntax-ready inventory rules, absent from main and all origin claim files,
are claimed here before implementation:

- @typescript-eslint/no-extra-non-null-assertion
- @typescript-eslint/no-misused-new
- @typescript-eslint/no-unnecessary-parameter-property-assignment

### Third-batch outcome

No-misused-new is implemented in owned .a modules and passed complete upstream
fixtures, compiler/stage1 corpus parity, all-rule dispatch and a compiling semantic
mutant on Node source, emitted JavaScript and sanitized native against Go cohere.
No-extra-non-null-assertion and no-unnecessary-parameter-property-assignment remain
claimed but blocked on the shared serializer and independent repair-range contract.
Valid one-finding Go probes exit 2 with unexpected fix shape and unexpected
suggestion shape respectively. Reproduction and measured throughput are in
../rules/typescript-eslint-no-misused-new/REPORT.md.

## Fourth batch

After pushing all prior work and fetching all origin heads, helper-ready rules
remain claimed. The first three unported and unclaimed syntax-ready entries are
claimed here before implementation:

- @typescript-eslint/no-unnecessary-type-constraint
- @typescript-eslint/prefer-as-const
- @typescript-eslint/prefer-enum-initializers

### Fourth-batch outcome

All three remain claimed but blocked by the shared repair contract. Independent
Go probes first count one valid finding and then refuse serialization: constraint
removal has a distinct suggestion range, as-const annotations have two fix edits,
and enum initializers have three ordered suggestions. No partial port is registered.
The complete original Go tests pass. Reproduction and exact limits are recorded
in wave1-14-fourth-report.md and wave1-14-fourth-evidence/.

## Complete repair follow-up

The six previously blocked claims now have owned .a implementations with complete
repair data and four-way fixture/corpus comparisons. No-confusing-non-null-assertion,
no-extra-non-null-assertion and no-unnecessary-parameter-property-assignment join
the three fourth-batch rules. Shared integration is still explicitly refused until
the .a registry and Finding repair contract land. One modified-destructuring parser
fixture is excluded; split-UTF-8 repairs are explicitly refused. Full evidence,
all semantic mutants and native/Node/Go findings rates are recorded in
../rules/typescript-eslint-prefer-as-const/COMPLETE_REPORT.md.

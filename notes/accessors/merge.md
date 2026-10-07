# Accessor merge receipt

The branch merges current main e011f8f under the standing self-landing rule.
Integration 16 had not reached main at the time of this merge. Coverage commit
d6b9675 is preserved in ancestry. Main's descriptor implementation replaces the
branch's earlier nominal method implementation, leaving one accessorSymbol in
class_accessors.go. Main's static, literal and structural accessor support stays
in place. Earlier implementation tests are replaced by main's accessor tests
and the focused inferred-array regression.

The remaining compiler delta is in invariance.go. A fresh data literal cannot
enter an accessor-bearing inferred view, including either ordering of
[new A(), { total: 2 }], and reports adamic/accessor-field-view. Nominal accessor
writes are checked through setter override parameters rather than the getter
result's storage invariance. Recursive checks still prove getter result contents.

All 26 branch fixtures remain registered. Twenty-four execute against Node,
JavaScript and native sanitizer builds. The original scalars and union_setter
programs are NotYet probes because main has no optional boolean accessor slot
and requires one native representation for both descriptor halves. Four fixtures
are adapted without changing their intended Node output: unrelated setter/getter
names no longer collide with main's conservative same-name dispatch, and optional
reference reads use receiver helpers before narrowing local results.

Counts were regenerated from native executions. Compared with main's 305 rows,
24 rows are added, zero removed and zero changed. The two NotYet programs have no
native execution and therefore no count row. New rows measure main's descriptor
calls and ordinary ownership, replacing the historical branch's method counts.

Validation, with test output sent to logs:

- Full internal/lower passed in 8.203 seconds.
- Uncached accessor, class feature, inheritance and borrowed-element oracles
  passed in 20.307 seconds.
- Counts regeneration passed in 15.475 seconds.
- Final restored checks passed: lowering in 1.948 seconds and uncached selected
  oracles plus counts in 19.258 seconds.
- go vet ./... and gofmt -l cmd internal passed with no output.

Mutants are restored after each run. Restoring the fresh-literal exemption makes
both element-order refusal tests fail with got nil. Treating the getter result as
a writable field again makes accessors_coverage_override_types.a fail lowering
at its valid Derived-to-Base view. Both fail for the intended check, not C warnings.

The complete 30-minute workspace gate was not run for this merge. Scope is the
lowering package, affected Node oracles, and the complete counts table.

# Mutants executed

All listed delivered-topic mutants must trip their intended check; compile failures are excluded unless the check is explicitly the pre-backend IR representation invariant. Raw passing self-mutant tests and runner output are under [evidence](evidence/). Failed controls from skipped topics are recorded separately and are never mutant kills.

## Group 1

- Checked non-null: remove the undefined check (required panic/exit and stdout pin); look through narrowed union storage (return representation assertion); remove literal null/undefined initializer checks (harmless fallback output pin).
- Namespace reads: remove readiness for early and direct-early cyclic reads (Node exit 70 versus mutant exit 0); freeze mutable export at zero (Node stdout).
- Namespace receiver, class registration, callable namespace: substitute receiver flags (Node stdout), omit constructor registration (required initialization trap), redirect attached call to root function (Node stdout).
- Migration supplements: seven ReadinessMutants variants, uninitialized-is-nullish, lazy initializer, and freed Weak diagnostic all preserve the assertion-site checks; dropping/misrepresenting them fails the pinned panic/output. Existing interface-cast and closed-nested constituent mutants in affected test files also passed. Exact subtest names and catchers are in the verbose supplement log.

## Group 2

- NaN-as-truthy and null-as-truthy: source Node stdout in both backends.
- Missing historical condition sample: independent source/site ledger rejects a missing or duplicate sample.

## Group 3

### area-stack-group3-final-element-presence-mutants.log

```text
missing-presence-check caught /tmp/element-access-presence-mutant-missing-presence-check.log
```

### area-stack-group3-final-element-read-mutants.log

```text
wrong-dispatch caught /tmp/element-access-read-mutant-wrong-dispatch.log
unrelated-getter-name caught /tmp/element-access-read-mutant-unrelated-getter-name.log
missing-absence caught /tmp/element-access-read-mutant-missing-absence.log
repeated-receiver caught /tmp/element-access-read-mutant-repeated-receiver.log
```

### area-stack-group3-final-element-tuple-mutants.log

```text
fractional-index caught /tmp/element-access-tuple-mutant-fractional-index.log
missing-slot caught /tmp/element-access-tuple-mutant-missing-slot.log
out-of-range caught /tmp/element-access-tuple-mutant-out-of-range.log
repeated-receiver caught /tmp/element-access-tuple-mutant-repeated-receiver.log
```

### area-stack-group3-final-optional-index-mutants.log

```text
lowering-optional-flag: caught by exit codes differ
native-eager-key: caught by stdout differs
native-repeat-receiver: caught by stdout differs
```

### area-stack-group3-final-overload-mutants.log

```text
additional-binder-refusal: caught by TestNativeAgreesWithNode/internal/oracle/testdata/(census_overload_binders|overload_(array_to_|ancestor_directory|leading_comment_range|trailing_comment_range|original_node|mutate_map|resolve_type_names|sort_deduplicate)); exit 1
additional-binder-refusal: all 11 executable per-kind fixtures caught it
parameter-inference: caught by TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders; exit 1
bottom-witness: caught by TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders; exit 1
constraint: caught by TestCensusOverloadBinderGuards/constraint; exit 1
parameter: caught by TestCensusOverloadBinderGuards/parameter; exit 1
parameter: also caught by 1 Node-held refusal fixtures
result: caught by TestCensusOverloadBinderGuards/result; exit 1
result: also caught by 2 Node-held refusal fixtures
union-inference: caught by TestOverloadInferenceWitnesses/one_compatible_binder; exit 1
union-inference: also caught by 3 oracle fixtures
known-union-member: caught by TestOverloadInferenceWitnesses/incompatible_known_member; exit 1
multiple-union-binders: caught by TestOverloadInferenceWitnesses/multiple_unknown_binders; exit 1
optional-rigid-binder: caught by TestOverloadInferenceWitnesses/optional_rigid_binder; exit 1
optional-rigid-binder: also caught by 1 oracle fixtures
```

### area-stack-group3-final-statements-mutants.log

```text
prefix-write: caught by stdout
prefix-result: caught by stdout
nonnull-check: caught by exit codes differ
nonnull-write: caught by stdout differs
nonnull-receiver: caught by stdout differs
throw-identity: caught by stdout
error-provenance: caught by stdout differs
parameter-order: caught by stdout
template-object: caught by want preserved stop
nonnull-refusal: caught by want preserved stop
initializing-capture: caught by want preserved stop
```

### Inline syntax mutants

Seven substr mutants (relative-start, truncate, NaN, undefined-length, length-not-end, length-clamp, evaluation-order) and two namespace syntax mutants (initializer, scoped read) fail only source Node stdout comparison in both backends. These self-mutant tests passed in the final verbose oracle log.

## Group 4

Omit non-direct argument escape: TestMethodKeepsArgument rejects acceptance of the cycle-closing push; the independent oracle executes Node and native, both printing 1, and reproduces 217 leaked bytes in four allocations under LeakSanitizer. Treat length-only targets as changing: both throughVirtual and throughClosure borrowing-plan assertions fail. See group-four overlay mutation logs.

## Catalog after every group

The eleven active catalog undo patches are independent: shared-slice-append, liveness-throw, defined-lent, borrowed-array-move, spread-method-reuse, constructor-capture-region, borrowed-element-reads, narrowed-number-field, literal-undefined-field, refuse-definite-assignment, refuse-suppression-directives. Each baseline passes and each mutant fails its exact recorded named test/diagnostic. Entries 10 and 13-16 remain historical catalog skips. No undo patch drift was found in groups 1-3. Group-four results are reported with the final push.

## Individual failure diagnostics

### area-stack-group4-borrow-proof-mutant.log

```text
--- FAIL: TestDevirtualizeBorrowDocClaim (0.05s)
element_borrow_test.go:142: throughVirtual: length-only targets prevented borrowing
element_borrow_test.go:142: throughClosure: length-only targets prevented borrowing
```

### area-stack-group4-method-escape-mutant.log

```text
--- FAIL: TestMethodKeepsArgument (0.04s)
fresh_test.go:85: method-kept argument closes a cycle: want refusal, got <nil>
```

### area-stack-group4-method-escape-oracle-mutant.log

```text
--- FAIL: TestFreshWriteProbesStayRefused (0.00s)
--- FAIL: TestFreshWriteProbesStayRefused/devirt_fresh_method_keeps_argument.a (0.42s)
```

### element-access-presence-mutant-missing-presence-check.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_presence.a (0.48s)
oracle_test.go:758: the release build: exit codes differ
oracle_test.go:774: exit codes differ
```

### element-access-read-mutant-missing-absence.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_reads.a (0.58s)
oracle_test.go:774: exit codes differ
```

### element-access-read-mutant-repeated-receiver.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_reads.a (0.56s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

### element-access-read-mutant-unrelated-getter-name.log

```text
--- FAIL: TestNativeAgreesWithNode (0.01s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_reads.a (0.04s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/element_access_reads.a:19:137: stage 0 can't lower a computed key without an own data field yet
```

### element-access-read-mutant-wrong-dispatch.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_reads.a (0.57s)
oracle_test.go:774: exit codes differ
oracle_test.go:780: JavaScript backend: exit codes differ
```

### element-access-tuple-mutant-fractional-index.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_tuple.a (0.61s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

### element-access-tuple-mutant-missing-slot.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_tuple.a (0.57s)
oracle_test.go:758: the release build: exit codes differ
oracle_test.go:774: exit codes differ
```

### element-access-tuple-mutant-out-of-range.log

```text
--- FAIL: TestNativeAgreesWithNode (0.01s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_tuple.a (0.47s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

### element-access-tuple-mutant-repeated-receiver.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/element_access_tuple.a (0.51s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

### notyet-overloads-mutants/additional-binder-refusal.log

```text
--- FAIL: TestNativeAgreesWithNode (0.06s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/overload_trailing_comment_range.a (0.07s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/overload_trailing_comment_range.a:1:1: stage 0 can't lower overload 1 of forEachTrailingCommentRange with additional implementation type parameters yet
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders.a (0.07s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/census_overload_binders.a:2:1: stage 0 can't lower overload 1 of token with additional implementation type parameters yet
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/overload_ancestor_directory.a (0.04s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/overload_ancestor_directory.a:2:1: stage 0 can't lower overload 1 of forEachAncestorDirectory with additional implementation type parameters yet
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/overload_leading_comment_range.a (0.05s)
```

### notyet-overloads-mutants/bottom-witness.log

```text
--- FAIL: TestNativeAgreesWithNode (0.01s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders.a (0.04s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/census_overload_binders.a:23:17: Adamic 0.1 refuses overload 1 of erased parameter callback cannot be served by implementation parameter callback; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

### notyet-overloads-mutants/constraint.log

```text
--- FAIL: TestCensusOverloadBinderGuards (0.04s)
--- FAIL: TestCensusOverloadBinderGuards/constraint (0.04s)
census_overload_binders_test.go:23: want "cannot satisfy implementation parameter", got <nil>
```

### notyet-overloads-mutants/known-union-member.log

```text
--- FAIL: TestOverloadInferenceWitnesses (0.04s)
--- FAIL: TestOverloadInferenceWitnesses/incompatible_known_member (0.00s)
```

### notyet-overloads-mutants/multiple-union-binders.log

```text
--- FAIL: TestOverloadInferenceWitnesses (0.04s)
--- FAIL: TestOverloadInferenceWitnesses/multiple_unknown_binders (0.00s)
```

### notyet-overloads-mutants/optional-rigid-binder-oracle.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/overload_ancestor_directory.a (0.03s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/overload_ancestor_directory.a:2:55: Adamic 0.1 refuses overload 1 of forEachAncestorDirectory parameter callback cannot be served by implementation parameter callback; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

### notyet-overloads-mutants/optional-rigid-binder.log

```text
--- FAIL: TestOverloadInferenceWitnesses (0.04s)
--- FAIL: TestOverloadInferenceWitnesses/optional_rigid_binder (0.00s)
```

### notyet-overloads-mutants/parameter-contracts.log

```text
--- FAIL: TestOverloadContractRulings (0.18s)
--- FAIL: TestOverloadContractRulings/serializer (0.17s)
overload_contract_test.go:32: want refusal "overload 2 of setSerializerContextAnd parameter cb", got <nil>
```

### notyet-overloads-mutants/parameter-inference.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders.a (0.04s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/census_overload_binders.a:6:22: Adamic 0.1 refuses overload 1 of ancestor parameter directory cannot be served by implementation parameter directory; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

### notyet-overloads-mutants/parameter.log

```text
--- FAIL: TestCensusOverloadBinderGuards (0.04s)
--- FAIL: TestCensusOverloadBinderGuards/parameter (0.04s)
census_overload_binders_test.go:23: want "cannot be served by implementation parameter", got <nil>
```

### notyet-overloads-mutants/result-contracts.log

```text
--- FAIL: TestOverloadContractRulings (0.39s)
--- FAIL: TestOverloadContractRulings/token (0.23s)
overload_contract_test.go:32: want refusal "overload 1 of createToken result", got <nil>
--- FAIL: TestOverloadContractRulings/trampoline (0.17s)
overload_contract_test.go:32: want refusal "overload 1 of createBinaryExpressionTrampoline result", got /workspace/area-stack-review/internal/oracle/testdata/overload_contracts/trampoline.a:3:19: stage 0 can't lower a value of type Outer yet
```

### notyet-overloads-mutants/result.log

```text
--- FAIL: TestCensusOverloadBinderGuards (0.04s)
--- FAIL: TestCensusOverloadBinderGuards/result (0.04s)
census_overload_binders_test.go:23: want "cannot be served by implementation result", got /workspace/adamic-scratch/TestCensusOverloadBinderGuardsresult1444850424/001/main.a:3:16: stage 0 can't lower an overload result requiring another representation yet
```

### notyet-overloads-mutants/union-inference-oracle.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/overload_array_to_numeric_map.a (0.05s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/overload_array_to_numeric_map.a:5:32: stage 0 can't lower an array of number | U yet
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/overload_array_to_map.a (0.05s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/overload_array_to_map.a:3:20: stage 0 can't lower a Map of V1 | V2 yet
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/overload_array_to_multimap.a (0.04s)
oracle_test.go:751: Lower: /workspace/area-stack-review/internal/oracle/testdata/overload_array_to_multimap.a:8:38: stage 0 can't lower an array of never yet
```

### notyet-overloads-mutants/union-inference.log

```text
--- FAIL: TestOverloadInferenceWitnesses (0.05s)
--- FAIL: TestOverloadInferenceWitnesses/one_compatible_binder (0.00s)
```

### statements-small-mutants/error-provenance.log

```text
--- FAIL: TestStatementsSmallRulings (0.50s)
--- FAIL: TestStatementsSmallRulings/structural_error (0.50s)
statements_small_rulings_test.go:36: want preserved stop "throwing an Error that isn't made where it's thrown or caught by the catch around it"; admitted program: native stdout differs; JavaScript stdout differs
```

### statements-small-mutants/initializing-capture.log

```text
--- FAIL: TestStatementsSmallRulings (0.25s)
--- FAIL: TestStatementsSmallRulings/capture (0.25s)
statements_small_rulings_test.go:45: want preserved stop "a function value that captures the variable its own initializer declares", got /workspace/area-stack-review/internal/oracle/testdata/statements_small_stopped/capture.a:2:11: Adamic 0.1 refuses 'map', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)
```

### statements-small-mutants/nonnull-check.log

```text
--- FAIL: TestStatementsSmallTypeScriptIncrement (0.35s)
--- FAIL: TestStatementsSmallTypeScriptIncrement/missing.a (0.35s)
statements_small_typescript_test.go:78: exit codes differ: got exit 0 stdout "before\nafter\n" stderr ""; want exit 70 stdout "before\n" stderr "adamic: panic: non-null assertion failed at /workspace/adamic-scratch/TestStatementsSmallTypeScriptIncrementmissing.a1079311314/001/missing.ts:3:1: state.index! is null or undefined\n"
```

### statements-small-mutants/nonnull-receiver.log

```text
--- FAIL: TestStatementsSmallTypeScriptIncrement (0.44s)
--- FAIL: TestStatementsSmallTypeScriptIncrement/nonnull.a (0.44s)
statements_small_typescript_test.go:78: stdout differs: got exit 0 stdout "2\n1\n-1\n1:4\nNaN\n" stderr ""; want exit 0 stdout "2\n1\n-1\n1:2\nNaN\n" stderr ""
```

### statements-small-mutants/nonnull-refusal.log

```text
--- FAIL: TestStatementsSmallRulings (0.48s)
--- FAIL: TestStatementsSmallRulings/nonnull (0.48s)
statements_small_rulings_test.go:36: want preserved stop "the non-null assertion !"; admitted program: native ; JavaScript
```

### statements-small-mutants/nonnull-write.log

```text
--- FAIL: TestStatementsSmallTypeScriptIncrement (0.43s)
--- FAIL: TestStatementsSmallTypeScriptIncrement/nonnull.a (0.43s)
statements_small_typescript_test.go:78: stdout differs: got exit 0 stdout "1\n1\n0\n0:2\nNaN\n" stderr ""; want exit 0 stdout "2\n1\n-1\n1:2\nNaN\n" stderr ""
```

### statements-small-mutants/parameter-order.log

```text
--- FAIL: TestNativeAgreesWithNode (0.01s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_parameters.a (0.59s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

### statements-small-mutants/prefix-result.log

```text
--- FAIL: TestNativeAgreesWithNode (0.02s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_prefix.a (0.54s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

### statements-small-mutants/prefix-write.log

```text
--- FAIL: TestNativeAgreesWithNode (0.01s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_prefix.a (0.48s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

### statements-small-mutants/template-object.log

```text
--- FAIL: TestStatementsSmallRulings (0.43s)
--- FAIL: TestStatementsSmallRulings/template (0.43s)
statements_small_rulings_test.go:36: want preserved stop "a template interpolating an object, an array, a map, a function or undefined"; admitted program: native stdout differs; JavaScript stdout differs
```

### statements-small-mutants/throw-identity.log

```text
--- FAIL: TestNativeAgreesWithNode (0.01s)
--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_throw.a (0.45s)
oracle_test.go:774: stdout differs
oracle_test.go:780: JavaScript backend: stdout differs
```

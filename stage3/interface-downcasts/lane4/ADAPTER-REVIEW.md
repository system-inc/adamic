# Primitive array adapter review

This patch is a review artifact, not applied production code. Automatic approval
review rejected the cross-layer adapter on memory-safety and silent-miscompile
risk, stating that it exceeded the accepted probe-only scope. No rejected edit
ran. The independent accepted heap and object probes passed release/sanitizer
checks and five semantic release mutants. They do not prove this array adapter.

The patch preserves the concrete rejected array snapshot/conversion helpers,
primitive descriptors and the intended source oracle. Admission and dispatcher
hooks are deliberately absent until the adapter passes its own runtime tests.
The source fixture group is currently a green gap test: Node prints each actual
member/wrong value, and Adamic refuses array lowering before execution. No pair
or candidate read is completed by this artifact.

After approval, the intended named hooks are:

1. PrimitiveViewMembers rejects unsupported, object and recursive member graphs.
2. adamic_view_primitive_array_snapshot calls shared array bounds/sparse lookup
   and normalizes actual storage. It does not scan or rewrite the source array.
3. emitViewPrimitiveArrayRead in each backend selects a member before producing
   the result. Native number boxes are owned; borrowed strings are retained.
4. Shared array read dispatch calls that adapter only for ir.Union. JavaScript
   program assembly includes MixedUnionRuntime once. No object member admission.
5. Lowering's elementType admits only complete primitive ir.Union elements and
   interns their declared array/element contracts before final read dispatch.
   Existing whole-program flow is reused. Unknown provenance retains a check.
   Unsupported union consumers remain named compile refusals.

Before source admission, runtime tests must cover boxed and unboxed numbers,
booleans, strings, unknown storage, missing indices, holes, relative indices and
undefined. Source tests must cover ordinary helpers as well as viewed arrays;
a normal numeric array passed to a union-array helper must convert or refuse,
never read numeric bits as a pointer. Alias writes must retain existing storage
certification checks. Each valid member runs source Node, native release,
JavaScript and sanitized native with leak checking. Wrong values pin exit 70,
field expression, declared union and found category. Selector skip/first-member,
unknown-storage trust and conversion ownership mutants must fail semantically.
Only then can the 6849:<element> pair / eight candidate reads be removed.

Approval requested: applying and validating this staged primitive array adapter,
then enabling the minimal named hooks only after its runtime/source gates pass.
This is not approval to accept wrong output, bypass diagnostics or push a failing
source-admission change. Remaining candidate pairs are nine / twenty-seven reads. Real CompilerOptions
has a string index signature, so its two keyof pairs still require string/number/
undefined selection. The earlier finite-key fixtures do not complete those pairs.

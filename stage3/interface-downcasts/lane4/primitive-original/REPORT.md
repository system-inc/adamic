Certified five original primitive pairs, 70 candidate reads, including EmitNode destructuring.
Commit: this checkpoint on codex/views-mixed-unions-2, after 527b7557.
Checks: uncached original C/JS/Node oracle PASS 100.134s; filtered lower PASS 0.872s; vet passed.
Mutants: five member-check bypasses and five untested-first-member substitutions; 20 backend kills.
Uncovered: remaining primitive shapes, arrays and dictionary selectors; no exact tsc reachability claim.

| Original pair | Candidate reads | Status |
| --- | ---: | --- |
| EvaluatorResult<string \| number \| undefined>.value | 61 | Certified |
| NodeLinks.isExhaustive | 4 | Certified, including rejection of number 1 where only 0 is allowed |
| EvaluatorResult<string \| number \| undefined> \| undefined, value | 3 | Certified |
| EmitNode \| undefined, constantValue | 1 | Certified |
| EmitNode.constantValue | 1 | Certified original destructuring |

Fixtures import the complete original declarations from pinned TypeScript
050880ce59e30b356b686bd3144efe24f875ebc8. Preparation verifies all 78 declaration
hashes, original source read locations and complete receiver field sets. Required
undefined and optional missing values remain distinct. Each valid member is held
to Node in both backends; positive native runs use sanitizers and leak checks.
Wrong boolean, null, wrong numeric literal and wrong string values have named
refusal pins. Helpers receive viewed values through the shared flow.

Run: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewOriginal(EvaluatorPrimitivePairs|FinitePrimitiveFields)$'
-count=1 -v. Output is logs/oracle.log. The integrated nullable primitive selector supplies the runtime checks.


The additional destructuring hook is in internal/lower/view_primitive_reads.go,
with minimal named calls from collections.go, interface_cast.go and readiness.go.
It requires a complete declared primitive contract. The helper can lower before
the cast, so bindings intern their descriptor before checking admissibility.
The ordinary binding control exposed removal of a required primitive-to-box
conversion; primitiveBindingConversion retains it even without a viewed value.
The three ordinary members now match Node in both backends and sanitizers.
Object unions and missing/unsupported descriptors retain refusal. Unknown-flow
guards are unchanged. No runtime ABI or emitter change is needed.

The complete original EmitNode graph exposed a separate callable producer-scan
panic on an unrelated binding pattern. viewCallableProducerDeclaration in
view_callables.go restricts name extraction to the four existing producer kinds.
The fixture catches the former panic; callable signature tests pass unchanged.
Both failures before their fixes are retained as diagnostic logs, not green gates.

Current command: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^Test(CheckedViewOriginal(EvaluatorPrimitivePairs|FinitePrimitiveFields)|PrimitiveOrdinaryBinding)$'
-count=1 -v. Exit 0, 100.134s, native cache misses 22, Node misses 80, no hits;
logs/bindings-oracle.log. Lower command: go test ./internal/lower -run
'^Test(Primitive(ViewDestructuringAdmission|BindingConversionRequiresCompleteDeclaration)|ViewFallbackScopesRetainWiderHelperGuard|ViewAggregateNullishConstantsAreNotAllocations|ViewUnionTargetAdmission|LazyViewDemandUsesSharedFlow|LazyViewArrayDemand|ViewCallable.*|CheckedViewCallable.*|DestructuredMethodsCannotLoadOwnSlots)$'
-count=1 -v. Exit 0, 0.872s, logs/bindings-lower.log. go vet
./internal/lower ./internal/oracle exits 0. No full repository gate is claimed;
the nine full lower-package baseline failures recorded at 6ed3ced0 are unchanged
and were not rerun for this focused checkpoint.

Each wrong-member read bypass executes valid release code in C and JS and
violates the named exit-70 pin. Each first-member substitution replaces the
undefined result with the first present primitive (string or numeric zero),
executes valid release code and disagrees with the original Node control.
The five groups contribute ten kills of each mutation class. This family has
no object member whose transitive check could be independently dropped.

Original mixed-primitive column now has 39 / 621 certified and 24 / 69 remaining.
One remaining intersection pair / one read belongs to lane 7. Nonbrand remainder
is 23 / 68, using candidate counts. __String is delivered early; mixed primitives
retain October 13, 2026, 17:00 MDT as the estimate. Whole-tsc exact reachability
still depends on a checker-clean program. Continue most-read primitive groups;
there is no overnight time stop.

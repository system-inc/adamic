# Remaining compiler resolutions

These are concrete proposals for held topics. Automatic approval review rejected the attempted resolutions before they ran; the cherry-picks were aborted. The incoming code is absent from compiler/area-stack. No test pass is claimed for these topics.

## case-declaration

Files: internal/javascript/javascript.go, internal/lower/locals.go, internal/lower/object.go, internal/native/emit_locals.go, and the topic's switch_declarations.go versus the stack's switch_bindings.go.

Stack: captured bindings carry readiness, global declarations are recorded, function declarations are hoisted, and switch scrutinees are evaluated outside case scope. Topic: case bindings and optional or literal-undefined declarations acquire readiness through a second switch declaration helper.

Proposal: retain switch_bindings.go and its captured/global readiness and hoisting rules; extend its per-binding initializer proof for the topic's cases. Keep every unproved-read refusal. Do not replace the counted closure or object layouts. Run every case fixture and readiness mutant. The attempted broad conflict resolver was rejected because it could discard switch scope behavior.

## void-or-undefined

Files: internal/lower/expression.go and internal/native/emit_objects.go.

Stack: spreads, defaults, counted arguments and FunctionType are carried through call lowering; method thunks retain ClosureTargets and packed slot proofs, and unsupported union methods refuse. Topic: void-or-undefined result handling and union method parameters.

Proposal: add the optional-void result only alongside the existing concrete void result handling; preserve all call metadata and evaluation order. Admit union parameters only through the existing boxed reference slots for a known single nonnullable signature. Keep union-return and MaybeBoolean method refusals and StructuralMethodThunks/ClosureTargets proofs. Run its own positive fixtures in both backends and its mutants. Automatic review rejected the attempted combined call and method lowering as an unverified semantic rewrite.

## for-in

Files: internal/fresh/library_language.go, internal/native/emit_functions.go, internal/native/emit_objects.go, internal/native/emit_statements.go, internal/native/reuse.go, and the incoming for_in.go/runtime/for_in.c.

Stack: counted argument fitting, ignored argument effects, Node Buffer freshness, object readiness/accessor metadata and ownership are already established. Topic: enumerated hidden keys and a manual object copier.

Proposal: retain existing argument and freshness rules; add enumeration only within proven fitted positions. Copy objects through adamic_object_copy_checked, preserving field metadata, accessors, readiness and undefined-spread shape. Release the retained copied hidden-key array and clone that array only; do not manually clone object slots. Preserve all private, field type, readiness and data-write refusals. Run its Node fixtures and ownership mutants. Automatic review rejected the proposed custom key copying as unverified memory ownership.

## Existing pending proposals

Method values: METHOD-VALUES-PROPOSAL.md gives the counted dispatch and receiver closure adaptation. Binary: REPORT.md records the five affected files and the retained assignment, freshness and ownership proofs. Namespace: PARSER-REVIEW.md records the qualified primitive-union proof and shared namespace var-slot proposals; these still gate enum and live exported bindings.

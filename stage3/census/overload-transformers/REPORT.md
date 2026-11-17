# Transformer overloads — #ev14ch9

Base: cf468df0747075e50d326a97ff03e8530d7a1a88. Upstream: TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

Initial report checkpoint, before adaptation construction. Three sets, three reported result sites; one proposed build, two stops. No landing branch exists yet and no proof is claimed passed.

| Set | Resolved callers | Road | Planned branch |
| --- | ---: | --- | --- |
| utilities.ts getProperties | 2 | (i) Widen only the true overload result to readonly PropertyDeclaration[] | typescript/overload-transformers-get-properties-1 |
| namedEvaluation.ts:173 injectClassNamedEvaluationHelperBlockIfMissing | 3 | (iii) Stop: kind-preserving generic result needed by a caller; current proof cannot follow member tests or reassignment | none |
| namedEvaluation.ts:469 transformNamedEvaluation | 23 | (iii) Stop: 23 assignments retain node-specific views; widening loses required kinds; current proof cannot follow switch/member tests | none |

## Read and caller census

Read all of internal/lower/census_small.go, overload_predicate_result.go, census_overload_proof.go and overload_result_check.go before choosing these roads. Exact upstream overloads, full implementations and all resolved call statements/types are in upstream-callers.json. callers.cjs uses stock TypeScript 6.0.3 getResolvedSignature and compares the resolved declaration's identity/location, including aliases. Identifier references are also recorded to detect non-call uses. It inventories the entire checkout: 81,468 files, 21,232 TypeScript files, all 709 src TypeScript roots (including harness and testRunner), and all four additional test/baseline TypeScript files containing a target spelling. Text only selects extra roots; it never identifies a caller. The 807-file program resolves 28 calls to these exact owners; none is in harness, testRunner or tests. Unrelated getProperties methods do not count. Upstream uses CRLF.

Ownership search covered every stage3/adapt owners.json, targets.json and adapt.cjs. Only 99-correctness-receiver-transformers/owners.json matched an exact target identifier, in the esnext.ts import. It pins none of the three function owners. getProperties therefore needs a new house-shape directory, ordered after existing dependencies.

### getProperties

Two overloads, one implementation. One true argument at ts.ts:858 feeds some solely for array nonemptiness (one caller); one false argument at classFields.ts:2312 binds instanceProperties for ordinary PropertyDeclaration processing (one caller). No caller requires InitializedPropertyDeclaration. The true result promises initialized elements, but the implementation's readonly PropertyDeclaration[] does not serve that promise under censusRelated. There is no presence difference, callback predicate or void promise. Its reached return has a cast, which censusReturnProof refuses; narrowOverload never accepts arrays.

Proposed road (i): widen the first overload's result to the already-served readonly PropertyDeclaration[]. Preserve both overloads, all parameters, the body, documentation and runtime values. No public API baseline declares this internal function. Byte-identical commented and comment-free emits will be asserted by the adapter, and the whole source/declaration checks remain required.

### injectClassNamedEvaluationHelperBlockIfMissing

Three callers: namedEvaluation.ts:250 passes a ClassExpression and immediately casts/checks the result as ClassExpression (one); esDecorators.ts:676 passes ClassLikeDeclaration and reassigns its union (one); esDecorators.ts:1062 passes ClassDeclaration and reassigns the ClassDeclaration variable (one). The last needs the narrow result. Body and overload are retained in upstream-callers.json in full.

censusRelated cannot serve Extract<ClassLikeDeclaration, Pick<T, "kind">> from ClassLikeDeclaration for rigid T. No callback/presence/void/phantom acceptance applies. The body proof stops on classHasExplicitlyAssignedName(node), on member-based tests, and on the assignment to node; its final local/parameter flow cannot prove the promised Extract. The implementation is non-generic, but its unresolved generic Extract promise is not an ordinary checkable literal-tag object here (the recorded refusal remains the task's starting evidence).

Candidate (i), dropping the overload or widening its result to ClassLikeDeclaration, loses the ClassDeclaration caller's kind. A generic implementation annotation requires proving the conditional result, not asserting it with a new cast. Candidate (ii), merely branching on node.kind or isClassDeclaration(node), is outside censusTrustedTest's bare-parameter leaves. Adding wrappers, assertions, a new shape or moving the problem into callers needs a ruling. Stop under (iii); no adaptation or ledger line is built.

### transformNamedEvaluation

23 callers, all reassigning the result: classFields.ts has 10, esDecorators.ts 9, esnext.ts 4. By argument kind/use: PropertyAssignment 2, ShorthandPropertyAssignment 2, VariableDeclaration 3, ParameterDeclaration 2, BindingElement 2, PropertyDeclaration 3 (including one private-name case whose Extract instantiates as never), ExportAssignment 2, assignment expressions 7. These are node-specific variables, not discarded results. Each full statement and instantiated result is recorded.

The implementation's inferred union of factory-update results does not serve the rigid Extract<NamedEvaluation, Pick<T, ...>> promise. There is no predicate callback, presence-only difference or void result. The body proof cannot read switch(node.kind), and a generic implementation would disable narrowOverload. The promised Extract includes intersection/name/operator restrictions rather than one ordinary checkable literal-tag node result.

Candidate (i), widening or dropping the overload, loses node-specific views at the assignment callers. Annotating the implementation with the generic Extract would need unproved casts. Candidate (ii), expanding the switch to if tests on node.kind, still fails censusTrustedTest. New type shapes, assertions or specialized wrappers require a ruling. Stop under (iii); no adaptation or ledger line is built.

## Proof status at initial checkpoint

For getProperties: scan-all, full latent census (utilities.ts, ts.ts, classFields.ts), restoring mutant, stock 79-root diagnostics, Adamic types, apply.sh, upstream npm ci/Hereby dts declaration comparison, TestProduct_Stage3AdaptedTree and unfiltered lane: NOT RUN. No before/after finding delta is claimed yet.

For both stopped sets: every build proof above is NOT RUN; no source edit was built. Node witnesses, lane value counter and body-restoring mutant are NOT RUN because no road (ii) is proposed.

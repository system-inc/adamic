Built: constraint-backed TNode storage and four .a witnesses toward roadmap step 30.
Base: dcdbb9098f77f30ad41790c56df1bd63ad462b63 on codex/hidden-06-generic-tnode; final delivery SHA is reported after push.
Checks: exact replay reproduced; focused lower/oracle and counts passed; assigned hidden intersection 11,417 -> 3,369 bytes, revealing 8,048 bytes.
Mutants: eight independent compiler mutations were caught; the default-return mutant printed 0 where Node printed 7.
Uncovered: visitNodes overload refusal, later statement boundaries and the enclosing ImmediatelyInvokedArrowFunction boundary remain; no corpus-wide reveal claim.

The original head still stops at esDecorators.ts:1239:9 on the requested area base. Concrete checker mappings and existing substitutions retain priority. A supported scalar constraint supplies scalar storage. A constraint represented as an object supplies tagged Union storage, preserving structural subtypes' runtime brands instead of assuming an object layout. Unknown, any, unconstrained binders, collection layouts and structural constraints admitting primitive values remain unsupported without specialization. Existing write lowering refuses writes through the tagged value. Generic function values remain NotYet.

The area-base overlay oracle independently passed the requested visit(7) fixture before this change; evidence/base-witness.log records that observation. The additional constraints fixture checks object identity, its narrower extra field, distinct string/array specializations, and a live constraint-backed read through an object field that inference leaves unmapped. Removing the fallback makes that emitted-body oracle fail with a value-of-type-TNode NotYet. Source Node, the JavaScript backend, sanitized native, release native and leak checks agree for both positive fixtures. The source witness prints `7\n`, with exit 0 and empty stderr. The constraints witness prints `true:7:node\n3:2\nnode\n`. Both negative fixtures are refused.

Commands and observations:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/hidden-06-setup-current.log 2>&1
source /workspace/adamic-tools/env.sh
# Pin worktree: /tmp/hidden-census-source at 388096e6a83a4e9d287fb827f793c599ba1bf0ad.
bash /tmp/hidden-census-source/stage3/apply.sh /tmp/hidden-adapted > /tmp/hidden-06-apply.log 2>&1
# Prepared with stage3/census/latent/make_overlay.py and built with -overlay.
/tmp/hidden-06-base-replay -project /tmp/hidden-adapted/src/compiler -where /tmp/hidden-adapted/src/compiler/transformers/esDecorators.ts:1239:9 -kind NotYet -reason 'a value of type TNode' > /tmp/hidden-boundary-06-replay.log 2>&1
# Exit 0: exact position, kind and reason reproduced.
/tmp/hidden-06-fix-replay -project /tmp/hidden-adapted/src/compiler -where /tmp/hidden-adapted/src/compiler/transformers/esDecorators.ts:1239:9 -kind NotYet -reason 'a value of type TNode' > /tmp/hidden-06-next-replay.log 2>&1
# Exit 1 expected: the old signature no longer reproduces; records the next boundaries.
go test ./internal/lower -run '^TestHiddenTNode' -count=1 > /tmp/hidden-06-lower.log 2>&1
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_generic_tnode' -count=1 -v > /tmp/hidden-06-oracle.log 2>&1
go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/fixtures/taste/17_binder_flow.a$' -count=1 -v > /tmp/hidden-06-binder-flow.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/hidden-06-counts-final.log 2>&1
# The original-expression overlay also passed the focused visit(7) witness.
# All final checks above passed. No whole-package test or full gate was run.
```

Setup succeeded. Reported cumulative timing lines: Node 0.120s, Go 0.194s, clang 0.631s, markdown dependencies 1.308s (npm step 0.971s), submodules 3.389s, Go build 298.632s, build cache 299.103s, done 299.146s. `nproc` is 5; cgroup CPU quota is 4. Go 1.27.1, clang 20.1.8, Node 24.19.0. The environment file is /workspace/adamic-tools/env.sh. Complete setup output is preserved in evidence/setup.log.gz.

Every mutant ran through a separate scratch Go overlay; production files were never left mutated. Each test command returned exit 1, with the intended assertion failing:

| Mutation | Intended catcher | Observation |
|---|---|---|
| Return numeric default for the specialized node argument | Focused hidden_boundary_generic_tnode.a oracle | Node stdout 7; native and JavaScript backend stdout 0; all exits 0 and stderr empty |
| Use Object instead of tagged Union for object constraints | TestHiddenTNodeConstraintRepresentation and TestHiddenTNodeConstraintMutationRemainsNotYet | Wrong representation; unsafe mutation accepted with nil error |
| Accept unconstrained/unknown/any binders as numbers | TestHiddenTNodeConstraintRepresentation | All three refusal probes fail |
| Accept an array layout from an unmapped constraint | TestHiddenTNodeConstraintRepresentation | Array-layout refusal probe fails |
| Remove constraint fallback | TestHiddenTNodeConstraintRepresentation and focused constraints fixture oracle | Positive constraint probes fail; emitted-body oracle stops at TNode |
| Ignore concrete checker mapper | TestHiddenTNodeConstraintRepresentation | Concrete mapping assertions fail |
| Ignore existing substitution | TestHiddenTNodeConstraintRepresentation | Concrete substitution assertions fail |
| Accept a generic function value as an undefined closure | TestHiddenTNodeGenericValueRemainsNotYet | Refusal replaced by nil error |

Mutation patches and logs are under evidence/. To reproduce one, apply its patch through a scratch Go overlay mapping internal/lower/expression.go, then run its named selector with `go test -overlay=<overlay.json> ... -count=1`, redirecting output to a log. The default-return selector is `^TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_generic_tnode.a$`; the emitted constraint selector is the corresponding `_constraints.a$` name.

Counts were refreshed from the final fixtures. Every changed row:

- New hidden_boundary_generic_tnode.a: allocations/frees 2/2, retains/releases 0/2, peak 2, regions 0. These are the requested specialized identity witness's observed counts.
- New hidden_boundary_generic_tnode_constraints.a: 8/8, 6/14, peak 5, regions 0. This includes identity, extra-field and string/array specialization checks plus the tagged constraint-backed field read. All allocated values are freed.
- logical_and_reference_maybe.a moved to its regenerated registration position; its numbers remain 8/8, 11/22, peak 4, regions 0.
- The stale 17_binder_flow.a row was removed: the area base already registers this fixture with lowers=false for optional own-field presence. Its focused oracle passed the existing refusal at 23:13, `writing a possibly absent optional own field`. This change does not alter that behavior.
- Negative generic-value and mutation fixtures are registered as lowers=false and do not receive counts rows. No existing numeric count changed.

The hidden measurement preserves all 82 adapted file hashes from RESULT.json at 388096e6. The assigned file SHA-256 is 3d77006c23ed173fb28a1399b0404a6cf51fe9976993747a22fb14d8a6059633. The selected half-open byte range is [61324, 72741).

The complete-project census was stopped before reaching the assigned file; a whole-file attempt was also replaced by a focused region census. The successful measurement retains the entire loaded project and all dependency/sibling registration, but attempts every independent declaration whose stock span intersects the assigned region. The pinned stock catalog proves there are exactly two: the enclosing transformESDecorators declaration at 295:1 and partialTransformClassElement at 1236:5. The scratch-only scope patches are preserved; the file refusal scan is omitted because its findings supply no hidden Boundary coverage. Statement rollback/continuation, signature refusals, checker guards and the no-output assertions are unchanged.

The pinned hidden.py `calculate` union/subtraction algorithm was applied to these records and stock metadata, and its output was intersected with the assigned range. This is an exact region measurement: the enclosing declaration still blocks its entire body on both sides; both overlapping declaration attempts are present and checker-eligible. The only additional stock body overlapping the region is the inline arrow at 1351:81 [69259, 69290), already contained in a remaining hidden statement span [69170, 69305). Thus skipped-dependency additions from other files cannot change this intersection, and no other independent declaration span can subtract bytes from it. Partial-file or corpus totals are not reported.

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-06-base-head-census /tmp/hidden-adapted/src/compiler /tmp/hidden-06-base-head.jsonl > /tmp/hidden-06-base-head.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-06-fix-head-census /tmp/hidden-adapted/src/compiler /tmp/hidden-06-fix-head.jsonl > /tmp/hidden-06-fix-head.log 2>&1
# Each record contains exactly the two intersecting declaration attempts.
# The pinned hidden.py calculate API computes boundary union minus independent coverage.
```

The historical and area-base intersections are both 11,417 bytes. The fixed intersection is 3,369 bytes, in 15 ranges preserved in evidence/fix-intersection.json. Difference: 8,048 bytes revealed toward step 30. The next boundary is esDecorators.ts:1250:13, bytes [61665, 61755), caused by visitorPublic.ts:196:5 refusing `overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor`. The enclosing ImmediatelyInvokedArrowFunction blocker at esDecorators.ts:670:14 remains, together with subsequent statement failures. Those are left refused.

The delivery base stays the explicitly requested origin/compiler/area-next-fixtures tip dcdbb909. Current origin/main 45487a80 diverges with stage 3 work; it was not merged, following this unit's requirement that the delivery carry only this worker's commits on the area tip. No main or area branch was changed or pushed. No checker options were changed and no cohere implementation was copied.

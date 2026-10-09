# Placeholder merge onto the corrected integration stack

This merge carries roadmap step 09 placeholders onto c79c75726375c937f0b18e18040e6769d916ed06. Delivery starts at c8f858b979e7461762b1742d0982dd73567b9a74; rehearsal starts at 678d94f95c07e0572e61cf2193fd5014bcbe1d1e. Both obsolete e3380f23 merges were uncommitted and were aborted. Neither was pushed. No main or area branch was modified or pushed.

## Conflict resolutions

| File | Resolution |
| --- | --- |
| internal/ir/ir.go | Keep placeholder provenance and checks together with checked-view origins and contracts. |
| internal/javascript/javascript.go | Keep nullish placeholder payloads and ready flags, ordinary uninitialized fields, the async readiness guard and checked-view dispatch. The automatic object-literal merge also needs the unset exception to preserve observable null. |
| internal/native/emit_objects.go | Keep stack physical-slot indexing and presence, plus observable placeholder storage. |
| internal/native/reuse.go | Keep physical-slot indexing and insertion presence for reused objects, plus placeholder storage readiness. |
| internal/native/view_fields.go | Keep the checked-view adapter for ordinary reads and the placeholder reader for unset-capable fields; present arms share the physical-kind check. |
| internal/oracle/counts.md | Regenerate from counted native execution; do not merge numerical rows by hand. |
| stage3/fixtures/host/status.json | Keep the captured callback-cycle refusal at line 13, matching incoming source headers. |

The corrected stack packs slot caches into one atomic word. Placeholder structural reads cannot recover an owner from the former cache index. Expose the runtime's existing owner-aware lookup, borrow its slot and owner, and derive metadata offsets with adamic_slot_index. Static inheritance continues to read the physical parent's metadata. Spread overrides and reused objects likewise derive insertion presence from the returned slot. These are compatibility resolutions, with no new admission or removed refusal.

## Mutants

| Independent compiler mutation | Intended catcher |
| --- | --- |
| Skip placeholder flow checks | Saved undefined and null reach a T parameter, printing missing or present with exit 0 instead of the pinned exit 70. |
| Evaluate loose nullish operands twice | Call count becomes 5 instead of Node's 3. |
| Retain stale alias readiness | The alias-reset negative prints missing with exit 0 instead of checking the typed use. |
| Treat JSON literal placeholders as assertions | Serialization exits 70 instead of completing with Node's preserved null and omitted undefined. |
| Erase null in merged JavaScript literal emission | JavaScript prints an empty object and an empty nested object instead of preserving null; native still agrees with Node. |
| Admit Weak placeholder storage | The retained NotYet test lowers successfully. |
| Remove standalone refusal bookkeeping initialization | TestOptionalWideningSpreadOverwrite panics in the compiler instead of returning its retained result. This is a compiler regression catcher, not a semantic or sanitizer proof. |

The focused oracle group also independently exercises seven dropped typed-flow guards, null-tag collapse, spread readiness, seven ordinary readiness mutants, ordinary nonliteral assertions, namespace readiness and nested checked-view guards. Completed semantic runs compare Node, native and JavaScript; native uses ASan/UBSan and leak checks. Sanitizer failures are not counted as intended mutation catchers.

## Limits

Only the requested focused checks and stage3 package are run, not the full gate. The scanner witness is a reduction of scanner.ts:4097, not a full scanner replay or a native run over all 301 measured projects. Linux counts are used. Deliberate checked stops report allocations at the stop, not completion-time leak freedom. Rehearsal remains build-ahead only.

## Count rows against the corrected stack

Columns are allocations, frees, retains, releases, peak live and region releases. These 40 measured changes are the same placeholder changes recorded on the pre-merge branch: 19 new rows and 21 later-check rows.

| Fixture | Before | After | Reason |
| --- | --- | --- | --- |
| internal/oracle/testdata/placeholder_nonnull_scanner.a | new | 4, 4, 9, 8, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_local.a | new | 12, 12, 11, 24, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_field.a | new | 9, 9, 15, 23, 5, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_observe.a | new | 27, 27, 48, 79, 9, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_equality.a | new | 4, 4, 10, 16, 2, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_loose.a | new | 10, 10, 23, 35, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_truthiness.a | new | 1, 1, 7, 10, 1, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_typeof.a | new | 3, 3, 9, 11, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_coalesce.a | new | 8, 8, 15, 27, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_optional.a | new | 6, 6, 16, 25, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_copy.a | new | 26, 26, 102, 121, 7, 0 | New fixture; copy, save/restore and spread preserve nullish payloads and release their storage. |
| internal/oracle/testdata/placeholder_nonnull_null_json.a | new | 5, 5, 0, 6, 3, 0 | New fixture; exact null/undefined JSON temporary fields use existing schema arms and release their objects and output strings. |
| internal/oracle/testdata/placeholder_nonnull_before_use.a | new | 0, 0, 1, 1, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_saved_leak.a | new | 1, 0, 3, 4, 1, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_before_use.a | new | 0, 0, 2, 1, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a | new | 1, 0, 4, 4, 1, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_assignment_result.a | new | 0, 0, 2, 3, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_return_assignment.a | new | 0, 0, 2, 3, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_alias_reset.a | new | 2, 1, 6, 7, 2, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/non_null_initialized.ts | 0, 0, 0, 0, 0, 0 | 55, 55, 50, 109, 18, 0 | Completing writes now execute; scalar placeholder slots box their values and captured saves own their cells. |
| internal/oracle/testdata/non_null_uninitialized_field.ts | 1, 0, 0, 0, 1, 0 | 1, 0, 3, 2, 1, 0 | Allocate the object, transfer its nullish field and stop at the typed property use. |
| internal/oracle/testdata/non_null_uninitialized_capture.ts | 0, 0, 0, 0, 0, 0 | 2, 0, 2, 0, 2, 0 | Create the captured cell and closure, then stop when the callback uses the unset value as T. |
| internal/oracle/testdata/non_null_uninitialized_exception.ts | 0, 0, 0, 0, 0, 0 | 1, 1, 3, 2, 1, 0 | Execute catch handling and its string output before the later typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_loop.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 1, 0, 0, 0 | Carry unset into the loop exit and stop at the typed use, after its reference transfer. |
| internal/oracle/testdata/non_null_uninitialized_const.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 2, 1, 0, 0 | Save the nullish reference then stop at the property receiver. |
| internal/oracle/testdata/non_null_static_initialized.ts | 1, 0, 0, 1, 1, 0 | 9, 9, 12, 22, 6, 0 | Complete the static writes; scalar boxes and strings are released on the successful path. |
| internal/oracle/testdata/non_null_static_uninitialized.ts | 1, 0, 0, 1, 1, 0 | 1, 0, 3, 2, 1, 0 | Allocate static storage and transfer its nullish value before the property-use check. |
| internal/oracle/testdata/non_null_uninitialized_optional.ts | 1, 0, 0, 0, 1, 0 | 3, 3, 2, 6, 3, 0 | Observe unset through optional handling and complete rather than stopping at initialization. |
| internal/oracle/testdata/non_null_uninitialized_spread.ts | 1, 0, 0, 0, 1, 0 | 2, 0, 5, 5, 2, 0 | Copy the actual nullish field, allocate the copy and stop only at its later typed use. |
| internal/oracle/testdata/non_null_literal_assignment.ts | 0, 0, 0, 0, 0, 0 | 1, 1, 2, 2, 1, 0 | Execute the reset, later completing write and observation; own and release the resulting value. |
| internal/oracle/testdata/non_null_uninitialized_iteration.ts | 1, 0, 0, 1, 1, 0 | 6, 1, 10, 8, 6, 0 | Allocate captured iteration cells and closures before the callback typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_interface.ts | 1, 0, 1, 0, 1, 0 | 1, 0, 2, 1, 1, 0 | Read the nullish field through the checked structural view before stopping at use as T. |
| internal/oracle/testdata/non_null_uninitialized_map_entry.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 1, 1, 0, 0 | Transfer the nullish reference before the map argument boundary stops. |
| stage3/interface-downcasts/default-staged.ts | 0, 0, 0, 0, 0, 0 | 2, 2, 6, 7, 2, 0 | Complete staged placeholder field writes and the structural-view reads. |
| stage3/interface-downcasts/default-boxed-write.ts | 0, 0, 0, 0, 0, 0 | 5, 5, 5, 12, 4, 0 | Complete scalar field writes using union boxes and release their values. |
| stage3/interface-downcasts/default-read-before-set.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 4, 2, 1, 0 | Allocate the placeholder object and check its typed structural field read. |
| stage3/interface-downcasts/readiness-identifier.ts | 0, 0, 0, 0, 0, 0 | 2, 2, 6, 8, 2, 0 | Complete reference field assignment and checked structural-view transfer. |
| stage3/interface-downcasts/readiness-identifier-uninitialized.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 4, 3, 1, 0 | Transfer the unset reference through a structural view and stop at use as T. |
| stage3/interface-downcasts/readiness-number.ts | 0, 0, 0, 0, 0, 0 | 5, 5, 5, 10, 5, 0 | Box completed numeric placeholder fields and release the boxes after structural reads. |
| stage3/interface-downcasts/readiness-number-uninitialized.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 5, 3, 1, 0 | Preserve the numeric field nullish payload and stop at its typed structural read. |

## Count rows against the pre-merge delivery and rehearsal

Every existing numerical row is unchanged. The 90 rows below are incoming stack fixtures absent from the original branch table; their additions record those fixtures without changing existing counting behavior. No rows are removed.

| Incoming fixture | Measured counts | Reason |
| --- | --- | --- |
| internal/oracle/testdata/literal_units_walk.a | 3, 3, 1, 5, 3, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/moves/accepted/objects.a | 2056, 2056, 6145, 4104, 2053, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_plain.a | 21, 21, 32, 36, 13, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_unions.a | 27, 27, 60, 49, 12, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_reject_empty.a | 5, 4, 9, 7, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_parameters.a | 17, 17, 18, 22, 12, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_typeof.a | 15, 15, 24, 26, 9, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_values.a | 66, 66, 126, 121, 16, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_discard.a | 25, 25, 42, 43, 10, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_reject_eager.a | 12, 9, 18, 16, 10, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_coverage_reject_nested.a | 18, 15, 31, 26, 17, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_typeof.a | 15, 15, 24, 26, 9, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_three.a | 20, 20, 25, 29, 11, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_nested.a | 22, 22, 38, 41, 12, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/async_throw.a | 12, 10, 16, 15, 10, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/hidden_boundary_generic_tnode.a | 2, 2, 0, 2, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a | 8, 8, 6, 14, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/hidden_boundary_never_array.a | 3, 3, 0, 4, 3, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/hidden_boundary_never_array_observations.a | 17, 17, 8, 33, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/optional_field_write.a | 2, 2, 0, 2, 1, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/optional_field_presence.a | 45, 45, 40, 59, 11, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/optional_field_construction.a | 85, 85, 59, 130, 25, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/optional_field_unknown.a | 22, 22, 15, 44, 7, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/optional_field_alias.a | 10, 10, 9, 20, 6, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/optional_field_alias_variants.a | 69, 69, 32, 99, 16, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/optional_field_checked_copy.a | 6, 2, 6, 9, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/output_edges/fsize.a | 8, 8, 3, 12, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/output_edges/fsize_out.a | 60000, 60000, 0, 60000, 3, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/output_edges/closed.a | 0, 0, 0, 0, 0, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/output_edges/panic_surrogate.a | 3, 1, 0, 2, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/output_edges/usr1.a | 2, 2, 0, 2, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_callback_served.a | 11, 11, 14, 20, 6, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_field_hatch/kind-valid.ts | 14, 14, 7, 20, 6, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_field_hatch/value-valid.ts | 9, 9, 17, 23, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_field_hatch/value-undefined.ts | 9, 9, 19, 23, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_results_transform.a | 8, 8, 9, 11, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_results_evaluate.a | 9, 9, 13, 22, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_results_binding.a | 6, 6, 14, 16, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_results_parameter.a | 2, 2, 5, 6, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_results_scalar.a | 3, 3, 11, 14, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_structural_block.a | 6, 6, 3, 9, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_structural_evaluator.a | 4, 4, 9, 14, 3, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_structural_fields.a | 4, 4, 16, 18, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_structural_factory.a | 9, 9, 8, 18, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_values/proven.a | 7, 7, 12, 19, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_values/returned-valid.ts | 13, 13, 26, 37, 6, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_values/narrow-valid.ts | 7, 7, 17, 23, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_values/order-valid.ts | 6, 6, 12, 17, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_values/module-valid.ts | 4, 4, 8, 14, 3, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_visitors/instantiations.ts | 11, 11, 11, 21, 9, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_visitors/original.a | 5, 5, 11, 17, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_visitors/original.ts | 5, 5, 11, 17, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_visitors/helper.a | 5, 5, 11, 17, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_visitors/helper.ts | 5, 5, 11, 17, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_visitors/overloaded-helper.a | 5, 5, 11, 17, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/overload_visitors/overloaded-helper.ts | 5, 5, 11, 17, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/fresh.a | 20, 20, 10, 25, 12, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/global_capture.a | 10, 10, 17, 21, 10, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/identity.a | 8, 8, 13, 19, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/large.a | 4105, 4105, 6154, 6164, 4102, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/map_capture.a | 8, 8, 10, 9, 8, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/nested.a | 20, 20, 6, 17, 13, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/numbers.a | 6, 6, 3, 9, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/readonly_forms.a | 12, 12, 8, 11, 10, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/readonly_function.a | 9, 9, 7, 9, 9, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/records.a | 18, 18, 19, 23, 15, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/recursive_tree.a | 10, 10, 9, 15, 10, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/strings.a | 13, 13, 10, 16, 9, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/concurrency/accepted/throw_middle.a | 9, 9, 6, 13, 9, 0 | New incoming stack fixture; native counted execution recorded. |
| bench/parallel_files.a | 3952651, 3952651, 8151045, 8654930, 8839, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/predicate_refusals/oct8_predicates_p14_overload_optional_chain_container.a | 3, 3, 5, 6, 3, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/predicate_refusals/oct8_predicates_p27_overload_parameter_rebound.a | 2, 1, 5, 6, 1, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/predicate_refusals/representation_controls.a | 14, 14, 16, 32, 7, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/regexp_replace/effects.a | 21, 21, 27, 33, 11, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/regexp_replace/move_effect.a | 11, 11, 16, 19, 8, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/regexp_replace/reentrant.a | 161, 161, 151, 184, 30, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/typed_arrays_uint16array.a | 173, 173, 50, 225, 16, 0 | New incoming stack fixture; native counted execution recorded. |
| internal/oracle/testdata/typed_arrays_uint16_stop.a | 5, 4, 1, 6, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/interfaces-good.a | 5, 5, 6, 12, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/interfaces-missing-inherited.a | 2, 0, 5, 3, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/interfaces-uninitialized-object.a | 2, 0, 6, 4, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/interfaces-wrong-inherited.a | 2, 0, 5, 3, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/objects-good.a | 5, 5, 6, 12, 4, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/objects-missing-nested.a | 2, 0, 5, 3, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/objects-uninitialized-nested.a | 2, 0, 6, 4, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/objects-untagged-good.a | 6, 6, 3, 10, 5, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/objects-untagged-wrong.a | 2, 0, 3, 3, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/objects-wrong-nested.a | 2, 0, 5, 3, 2, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/optional-read-control.a | 1, 1, 6, 6, 1, 0 | New incoming stack fixture; native counted execution recorded. |
| stage3/interface-downcasts/lane1/flag-downcast-wrong.a | 3, 0, 2, 3, 3, 0 | New incoming stack fixture; native counted execution recorded. |

## Verification

Both trees pass:

- go build ./cmd/adamic and go vet ./internal/....
- gofmt on resolved Go files and git diff --check against c79c7572. The complete merge diff against the old parent reports whitespace already present in incoming disassembly and evidence; those artifacts are preserved.
- The fast gate Gate.aCheck against c79c7572: 20 changed .a files pass, including the pinned ordinary nonliteral refusal. The harness uses the gate implementation recorded in fbac28c62493f27a788edc02a18bc8edb68de5da; no dropped integration test split is restored.
- go test ./internal/lower -run 'Placeholder|Readiness|NonNull|Uninitialized|View|OptionalPresence' -count=1.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(placeholder_nonnull|non_null_)|TestPlaceholder|Test.*Readiness|Test.*ViewField|Test.*NonNull|TestLiteralDefaultPlaceholder|TestCheckedViewObjects|TestCheckedViewInterfaces|TestCheckedViewOptionalReadBoundary' -count=1 -v. Each tree records 229 native misses and 248 Node misses.
- go test ./internal/oracle -run '^TestDefaultTaggedSourceViews$' -count=1 -v.
- go test ./cmd/adamic -run Placeholder -count=1, including checked-site reporting.
- Seven independently compiled Go overlay mutants per tree, each failing at its intended catcher.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts, then the same test without update. The regenerated tables are identical.
- go test ./stage3/fixtures, green after the two verified stage0 record updates. Initial runs failed only those two metadata comparisons, with Node and native comparisons passing. No Node disagreement was concealed.

Toolchain setup on the resolved corrected tree reports node ready 0.088s, Go ready 0.097s, markdown ready 0.221s, submodules ready 0.245s, clang ready 0.817s, build ready 97.633s, test binaries deferred 99.718s, cache warm 99.725s, done 99.933s. nproc is 5; cgroup CPU quota is 4. An initial setup against the obsolete unresolved merge failed on IR conflict markers; resolving the merge and running setup on the corrected tree succeeded. GOPROXY is https://proxy.golang.org|direct; the setup environment is sourced from /workspace/adamic-tools/env.sh.

Logs and the temporary harnesses are compressed in stack-merge-evidence. Go build cache and abandoned temporary build directories were pruned to avoid disk exhaustion, excluding active imports. Source files were preserved. The rehearsal also incorporates the delivery merge so the build-ahead tip contains the exact delivery history.

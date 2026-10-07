# Re-land Linux counts

Baseline: `a37ebdb0913eca9d3d6e8adbe8f01bb733da52eb`. Measured final table: 820 rows; 87 existing rows move, 39 are added, none are removed. Tuples are allocations, frees, retains, releases, peak, regions.

Attribution uses the source constructs, merged IR and measured tables; it is not a runtime profile assigning each retain to a particular stack frame. Source error paths include real library, lexical initialization and invalidated-narrowing failures.

| Program | Batch 3 | Re-land | Cause |
| --- | --- | --- | --- |
| `internal/load/testdata/0.1/compile/09_tree.ts` | `33, 33, 86, 89, 16, 0` | `19, 19, 48, 89, 16, 0` | Proven reference reads retain their invariant checks without throw edges, restoring recursive Perceus reuse and reducing retains. |
| `internal/oracle/testdata/047cb0d_n_arrayindex.a` | `2, 1, 4, 4, 1, 0` | `4, 4, 12, 15, 3, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/047cb0d_n_element.a` | `5, 3, 9, 10, 4, 0` | `7, 6, 17, 21, 4, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/047cb0d_n_element_method.a` | `5, 3, 10, 12, 4, 0` | `7, 6, 18, 23, 4, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/047cb0d_n_element_plain.a` | `5, 3, 8, 10, 4, 0` | `7, 6, 16, 21, 4, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/4e1649c_generic_receiver.a` | `2, 2, 6, 8, 1, 0` | `2, 2, 5, 7, 1, 0` | The executed nullable String.prototype receiver failure now constructs a nominal TypeError directly instead of constructing Error and rewriting its name. |
| `internal/oracle/testdata/4e1649c_generic_receiver_charat.a` | `2, 2, 6, 8, 1, 0` | `2, 2, 5, 7, 1, 0` | The executed nullable String.prototype receiver failure now constructs a nominal TypeError directly instead of constructing Error and rewriting its name. |
| `internal/oracle/testdata/4e1649c_generic_receiver_number.a` | `2, 2, 5, 7, 1, 0` | `2, 2, 4, 6, 1, 0` | The executed nullable String.prototype receiver failure now constructs a nominal TypeError directly instead of constructing Error and rewriting its name. |
| `internal/oracle/testdata/4e1649c_generic_receiver_order.a` | `2, 2, 6, 8, 1, 0` | `2, 2, 5, 7, 1, 0` | The executed nullable String.prototype receiver failure now constructs a nominal TypeError directly instead of constructing Error and rewriting its name. |
| `internal/oracle/testdata/borrow_chain_coverage_finally.a` | `13, 13, 9, 13, 6, 0` | `13, 13, 8, 12, 6, 0` | Executed try { throw new Error(root.child.text); }; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/borrow_element_throw.a` | `38, 38, 31, 51, 8, 0` | `38, 38, 26, 46, 8, 0` | Executed throw new Error(caught${items.length});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/borrow_loop.a` | `128, 126, 113, 184, 14, 2` | `128, 126, 111, 182, 14, 2` | Executed throw new Error(throw${items.length});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/borrow_loop_calls.a` | `263, 261, 197, 363, 25, 2` | `263, 261, 196, 362, 25, 2` | Executed throw new Error(stop:${item.label});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/class_as_interface.a` | `372, 372, 386, 558, 60, 0` | `372, 372, 406, 578, 60, 0` | Executed throw new Error(built('negative', value));; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/class_features_accessors.a` | `67, 67, 53, 110, 21, 0` | `67, 67, 51, 108, 21, 0` | Executed get text(): string { throw new Error(getter${1}); }; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/class_features_static.a` | `110, 110, 101, 208, 28, 0` | `110, 110, 98, 205, 28, 0` | Executed class StaticFailure { static text = mark('q'); static fail(): string { throw new Error(this.text + 'fail'); } static get b...; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/class_features_static_private.a` | `42, 42, 119, 160, 12, 0` | `42, 42, 114, 155, 12, 0` | Actual private-brand failures construct nominal TypeErrors directly, omitting legacy name-store handoffs. |
| `internal/oracle/testdata/class_inheritance_conditional.a` | `164, 164, 289, 392, 35, 0` | `164, 164, 285, 388, 35, 0` | Executed missing/repeated super paths construct nominal ReferenceErrors; source Error throws also use direct prefix initialization. |
| `internal/oracle/testdata/class_inheritance_exceptions.a` | `29, 29, 27, 47, 8, 0` | `29, 29, 26, 46, 8, 0` | Executed throw new Error('base constructor');; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/class_instance_key_throw.a` | `6, 5, 10, 15, 5, 1` | `6, 5, 9, 14, 5, 1` | Executed override read(): string { throw new Error(bad${this.value.n}); }; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/class_static_alias_tdz.a` | `1, 0, 1, 2, 1, 0` | `4, 3, 11, 12, 4, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `internal/oracle/testdata/class_static_tdz.a` | `0, 0, 0, 0, 0, 0` | `3, 3, 10, 10, 3, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `internal/oracle/testdata/closures_throw.a` | `418, 418, 541, 776, 142, 0` | `418, 418, 1553, 1788, 142, 0` | Executed throw new Error(word('map stopped at ', index));; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/closures_throw_uncaught.a` | `31, 29, 15, 31, 13, 0` | `33, 33, 14, 33, 13, 0` | Executed throw new Error(${item} is too loud);; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/dead_zone.a` | `0, 0, 0, 0, 0, 0` | `3, 3, 10, 10, 3, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `internal/oracle/testdata/exceptions.a` | `133, 133, 167, 242, 20, 0` | `133, 133, 166, 241, 20, 0` | Executed throw new Error(empty input after trimming '${text}');; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/exceptions_empty.a` | `2, 1, 6, 4, 2, 0` | `3, 2, 7, 7, 2, 0` | Executed throw new Error();; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/exceptions_uncaught.a` | `3, 1, 3, 5, 2, 0` | `5, 5, 2, 7, 4, 0` | Executed throw new Error(${label} failed);; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/finally_leaves.a` | `157, 157, 59, 174, 10, 0` | `157, 157, 47, 162, 10, 0` | Executed throw new Error(word('lost ', index));; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/fresh_writes.a` | `330, 330, 447, 563, 132, 0` | `330, 330, 446, 562, 132, 0` | Executed throw new Error(too many: ${count});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/generic_method_return.a` | `23, 15, 34, 42, 5, 6` | `25, 19, 32, 43, 5, 6` | Executed throw new Error(missing ${key});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/library_array_flat_map.a` | `50, 50, 64, 113, 17, 0` | `50, 50, 63, 112, 17, 0` | Executed try { values.flatMap((value: number): number[] => { throw new Error(${value}); }); } catch (error) { if (error instanceof ...; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/library_array_with.a` | `68, 68, 70, 136, 10, 0` | `68, 68, 100, 161, 10, 0` | Invalid with indices construct nominal RangeErrors and retain cleanup-safe operands on the protected calls. |
| `internal/oracle/testdata/library_function_expressions.a` | `56, 56, 47, 94, 19, 0` | `56, 56, 50, 97, 19, 0` | Executed [1].forEach(function (): void { throw new Error('expression throw'); });; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_arrays.a` | `995, 995, 1128, 1551, 95, 0` | `995, 995, 1127, 1550, 95, 0` | Executed throw new Error(deleted ${show(current)} ${show(item)});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_closures.a` | `816, 816, 1057, 1290, 87, 0` | `816, 816, 1056, 1289, 87, 0` | Executed throw new Error(deleted ${current()} ${item()});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_fnexpr.a` | `44, 44, 64, 100, 15, 0` | `44, 44, 63, 99, 15, 0` | Executed throw new Error(cleared ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_keep.a` | `516, 516, 559, 865, 35, 0` | `516, 516, 624, 930, 35, 0` | Executed throw new Error(deleted ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_keys.a` | `391, 391, 303, 614, 25, 0` | `391, 391, 302, 613, 25, 0` | Executed throw new Error(deleted ${current} ${value});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_named_closures.a` | `117, 117, 197, 237, 29, 0` | `117, 117, 196, 236, 29, 0` | Executed throw new Error(deleted ${current()} ${item()});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_named_more.a` | `395, 395, 620, 870, 36, 0` | `395, 395, 619, 869, 36, 0` | Executed throw new Error(deleted ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_named_numbers.a` | `117, 117, 93, 197, 12, 0` | `117, 117, 92, 196, 12, 0` | Executed throw new Error(cleared ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_named_objects.a` | `165, 165, 213, 284, 31, 0` | `165, 165, 212, 283, 31, 0` | Executed throw new Error(deleted ${current.label} ${item.label});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_numbers.a` | `373, 373, 155, 478, 14, 0` | `373, 373, 154, 477, 14, 0` | Executed throw new Error(deleted ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/map_foreach_objects.a` | `732, 732, 702, 944, 63, 0` | `732, 732, 701, 943, 63, 0` | Executed throw new Error(deleted ${current.label} ${item.label});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/move_throw.a` | `42, 42, 22, 50, 9, 0` | `42, 42, 17, 45, 9, 0` | Executed throw new Error(failed${1});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/named_function_values.a` | `116, 116, 109, 204, 25, 0` | `116, 116, 111, 206, 25, 0` | Executed throw new Error(failing at ${value});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/narrowed_fields.a` | `5, 3, 7, 8, 4, 0` | `7, 6, 15, 19, 4, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/narrowed_methods.a` | `4, 3, 7, 9, 3, 0` | `6, 6, 15, 20, 3, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/narrowed_reads.a` | `6, 5, 6, 9, 4, 0` | `8, 8, 14, 20, 4, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/narrowed_writes.a` | `1, 1, 2, 4, 1, 0` | `4, 4, 13, 16, 3, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/nbody_runtime_fields.a` | `8, 8, 8, 15, 4, 0` | `8, 8, 7, 14, 4, 0` | Nonthrowing dynamic new Error initialization is direct, preserving its name/message/cause prefix and removing a constructor handoff. |
| `internal/oracle/testdata/nested_destructured_tdz.a` | `2, 0, 1, 0, 2, 0` | `5, 5, 11, 12, 3, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `internal/oracle/testdata/nested_tdz.a` | `2, 0, 1, 0, 2, 0` | `5, 5, 11, 12, 3, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `internal/oracle/testdata/nested_tdz_write.a` | `2, 0, 1, 0, 2, 0` | `5, 5, 11, 12, 3, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `internal/oracle/testdata/normalize_form.a` | `3, 2, 0, 3, 3, 0` | `6, 5, 10, 14, 4, 0` | The invalid normalization form constructs a nominal RangeError before entering the runtime. |
| `internal/oracle/testdata/optional_after_call_propagation.a` | `18, 18, 13, 20, 7, 0` | `18, 18, 29, 36, 7, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/optional_after_call_required.a` | `3, 3, 5, 7, 2, 0` | `3, 3, 13, 15, 2, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/param_assigned_in_try.a` | `46, 46, 27, 62, 12, 0` | `46, 46, 28, 63, 12, 0` | Executed throw new Error(word('too big ', value));; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/precision_range.a` | `4, 3, 1, 4, 2, 0` | `7, 6, 11, 15, 4, 0` | The executed invalid toPrecision count constructs a nominal RangeError and unwinds. |
| `internal/oracle/testdata/radix_range.a` | `3, 3, 0, 3, 1, 0` | `6, 6, 10, 14, 3, 0` | The executed radix 37 constructs a nominal RangeError instead of stopping in the runtime. |
| `internal/oracle/testdata/records_narrowed_reference.a` | `4, 1, 10, 10, 3, 0` | `6, 4, 18, 21, 5, 0` | Deleting the record entry invalidates its narrowed reference; the failed name read now raises a nominal TypeError through the shared guard. |
| `internal/oracle/testdata/regexp_null_narrowed.a` | `5, 4, 11, 7, 3, 0` | `7, 7, 19, 18, 3, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/regions_throw.a` | `124, 95, 9, 99, 31, 29` | `124, 95, 50, 140, 31, 29` | Executed throw new Error(budget spent at depth ${depth});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/reuse_narrowed.a` | `8, 5, 9, 10, 5, 0` | `10, 8, 17, 21, 5, 0` | An executed invalidated narrowing or absent receiver raises a nominal TypeError and unwinds through its guard rather than a structural error or panic. |
| `internal/oracle/testdata/reuse_throw.a` | `26, 26, 24, 41, 8, 0` | `26, 26, 21, 38, 8, 0` | Executed throw new Error(refused at ${next.count});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/route_targets_try_loop.a` | `68, 66, 50, 94, 9, 2` | `68, 66, 47, 91, 9, 2` | Executed throw new Error(stopped${value.label});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_arrays.a` | `516, 516, 650, 839, 58, 0` | `516, 516, 649, 838, 58, 0` | Executed throw new Error(deleted ${show(current)} ${show(item)});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_closures.a` | `411, 411, 640, 711, 39, 0` | `411, 411, 639, 710, 39, 0` | Executed throw new Error(deleted ${current()} ${item()});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_keep.a` | `290, 290, 385, 520, 18, 0` | `290, 290, 384, 519, 18, 0` | Executed throw new Error(deleted ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_keys.a` | `285, 285, 182, 417, 16, 0` | `285, 285, 181, 416, 16, 0` | Executed throw new Error(deleted ${current} ${value});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_named_closures.a` | `95, 95, 175, 207, 28, 0` | `95, 95, 174, 206, 28, 0` | Executed throw new Error(deleted ${current()} ${item()});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_named_more.a` | `179, 179, 363, 444, 26, 0` | `179, 179, 362, 443, 26, 0` | Executed throw new Error(deleted ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_named_objects.a` | `87, 87, 119, 158, 22, 0` | `87, 87, 118, 157, 22, 0` | Executed throw new Error(cleared ${current.label} ${item.label});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_numbers.a` | `214, 214, 140, 317, 11, 0` | `214, 214, 139, 316, 11, 0` | Executed throw new Error(deleted ${current} ${item});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/set_foreach_objects.a` | `368, 368, 461, 532, 30, 0` | `368, 368, 460, 531, 30, 0` | Executed throw new Error(deleted ${current.label} ${item.label});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/string_too_long.a` | `0, 0, 0, 0, 0, 0` | `3, 3, 11, 12, 3, 0` | The oversized repeat executes its length guard and constructs a nominal RangeError. |
| `internal/oracle/testdata/taste_void.a` | `39, 39, 8, 44, 4, 0` | `39, 39, 7, 43, 4, 0` | Executed function fail(): number { count += 1; throw new Error(void throws ${count}); }; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/throw_global_move.a` | `18, 18, 12, 26, 6, 0` | `18, 18, 9, 23, 6, 0` | Executed throw new Error(inner${2});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/throw_in_writes.a` | `22, 22, 8, 22, 8, 0` | `22, 22, 6, 20, 8, 0` | Executed throw new Error(failed${1});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/throw_keeps_old_value.a` | `18, 18, 9, 21, 6, 0` | `18, 18, 7, 19, 6, 0` | Executed throw new Error(failed${1});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/throw_keeps_old_value_variants.a` | `26, 26, 11, 32, 6, 0` | `26, 26, 8, 29, 6, 0` | Executed throw new Error(failed${1});; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/user_iterators.a` | `669, 669, 624, 1073, 67, 0` | `669, 669, 617, 1066, 67, 0` | Executed throw new Error('next failed');; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `internal/oracle/testdata/user_iterators_rest_tdz.a` | `5, 0, 3, 2, 5, 0` | `8, 3, 13, 12, 8, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `internal/oracle/testdata/writes_in_try.a` | `30, 30, 33, 45, 6, 0` | `30, 30, 32, 44, 6, 0` | Executed throw new Error(['th', 'rown'].join(''));; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |
| `stage3/drivers/scanner/probes/cyclic-premature-value/main.a` | `0, 0, 0, 0, 0, 0` | `3, 3, 10, 10, 3, 0` | The actual read or write before lexical initialization constructs a nominal ReferenceError and unwinds its owned frame. |
| `stage3/fixtures/taste/24_proportional_conditions.a` | `43, 43, 143, 185, 14, 0` | `43, 43, 137, 179, 14, 0` | Executed const Debug = {fail: (message: string): void => {throw new Error(message);}};; nominal prefix ownership, direct construction and exceptional cleanup replace the legacy structural-error path. |

## Added rows

- `internal/oracle/testdata/047cb0d_narrowed_in_try.a`
- `internal/oracle/testdata/4ddd17f_opt_3.a`
- `internal/oracle/testdata/57f2d04_with_frozen.a`
- `internal/oracle/testdata/9984394_defined_in_try.a`
- `internal/oracle/testdata/9984394_lib_dispatch.a`
- `internal/oracle/testdata/catchability-limits/d96d304_try_finally_concat.a`
- `internal/oracle/testdata/catchability-limits/d96d304_try_pad.a`
- `internal/oracle/testdata/catchability-limits/d96d304_try_repeat.a`
- `internal/oracle/testdata/catchability-limits/d96d304_try_stack.a`
- `internal/oracle/testdata/coverage_error_argument_order.a`
- `internal/oracle/testdata/coverage_error_array_with.a`
- `internal/oracle/testdata/coverage_error_bounds_mutation.a`
- `internal/oracle/testdata/coverage_error_callback_library.a`
- `internal/oracle/testdata/coverage_error_catch_finally.a`
- `internal/oracle/testdata/coverage_error_cause_null.a`
- `internal/oracle/testdata/coverage_error_causes.a`
- `internal/oracle/testdata/coverage_error_codepoint_uncaught.a`
- `internal/oracle/testdata/coverage_error_interface_throw.a`
- `internal/oracle/testdata/coverage_error_kinds.a`
- `internal/oracle/testdata/coverage_error_nested_rethrow.a`
- `internal/oracle/testdata/coverage_error_normalize_forms.a`
- `internal/oracle/testdata/coverage_error_nullable_cause.a`
- `internal/oracle/testdata/coverage_error_number_bounds.a`
- `internal/oracle/testdata/coverage_error_number_prototypes.a`
- `internal/oracle/testdata/coverage_error_optional_messages.a`
- `internal/oracle/testdata/coverage_error_padding_dispatch.a`
- `internal/oracle/testdata/coverage_error_prototype.a`
- `internal/oracle/testdata/coverage_error_prototype_null.a`
- `internal/oracle/testdata/coverage_error_repeat_unicode.a`
- `internal/oracle/testdata/coverage_error_tdz_interface.a`
- `internal/oracle/testdata/coverage_error_uncaught_surrogate.a`
- `internal/oracle/testdata/error_checks.a`
- `internal/oracle/testdata/error_classes.a`
- `internal/oracle/testdata/error_classes_uncaught.a`
- `internal/oracle/testdata/error_classes_uncaught_empty.a`
- `internal/oracle/testdata/error_classes_uncaught_name.a`
- `internal/oracle/testdata/error_record_narrowing_identity.a`
- `internal/oracle/testdata/error_host_uncaught.a`
- `internal/oracle/testdata/node_fs_file_nominal.a`

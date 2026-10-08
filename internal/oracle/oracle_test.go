// Package oracle holds Adamic's differential test: every fixture runs twice, from source on Node and
// as a native binary stage 0 built, and the two must agree byte for byte on stdout and stderr and
// exactly on the exit code. A disagreement is a miscompile until proven otherwise.
package oracle

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../.."

// fixtures is every program the oracle knows, and whether stage 0 lowers it yet. A program that
// doesn't lower yet must still be refused loudly, with where and what, so all of them are run every
// time and the list says exactly how far stage 0 has come.
var fixtures = []struct {
	path   string
	lowers bool

	// checked is a program whose inserted check fires: there, Node running the source doesn't have
	// the check, so the native binary is held to the JavaScript backend, which does.
	checked bool
}{
	{"internal/oracle/testdata/string_views_lifetime.a", true, false},
	{"internal/oracle/testdata/string_views_holders.a", true, false},
	{"internal/oracle/testdata/string_views_throw.a", true, false},
	{"internal/oracle/testdata/string_views_loops.a", true, false},
	{"internal/oracle/testdata/string_views_policy.a", true, false},
	{"internal/oracle/testdata/string_views_methods.a", true, false},
	{"internal/oracle/testdata/string_views_characters.a", true, false},
	{"internal/oracle/testdata/string_views_calls.a", true, false},
	{"internal/oracle/testdata/string_views_surrogates.a", true, false},
	{"internal/oracle/testdata/non_null.ts", true, false},
	{"internal/oracle/testdata/non_null_map.ts", true, true},
	{"internal/oracle/testdata/non_null_catch.ts", true, true},
	{"internal/oracle/testdata/non_null_field.ts", true, true},
	{"internal/oracle/testdata/non_null_capture.ts", true, true},
	{"internal/oracle/testdata/non_null_array.ts", true, true},
	{"internal/oracle/testdata/non_null_union.ts", true, true},
	{"internal/oracle/testdata/non_null_boolean.ts", true, true},
	{"internal/oracle/testdata/non_null_null.ts", true, true},
	{"internal/oracle/testdata/typeof_null.a", true, false},
	{"internal/oracle/testdata/route_targets_callbacks.a", true, false},
	{"internal/oracle/testdata/route_targets_virtual_fresh.a", true, false},
	{"internal/oracle/testdata/route_targets_unknown.a", true, false},
	{"internal/oracle/testdata/route_targets_bound_method.a", true, false},
	{"internal/oracle/testdata/route_targets_recursive.a", true, false},
	{"internal/oracle/testdata/route_targets_try_loop.a", true, false},
	{"internal/oracle/testdata/route_targets_accessor.a", true, false},
	{"internal/oracle/testdata/route_targets_sort_values.a", true, false},
	{"internal/oracle/testdata/map_foreach_named_keys.a", true, false},
	{"internal/oracle/testdata/set_foreach_named_keys.a", true, false},
	{"internal/oracle/testdata/map_foreach_keys.a", true, false},
	{"internal/oracle/testdata/set_foreach_keys.a", true, false},
	{"internal/oracle/testdata/map_foreach_keep.a", true, false},
	{"internal/oracle/testdata/set_foreach_keep.a", true, false},
	{"internal/oracle/testdata/map_foreach_objects.a", true, false},
	{"internal/oracle/testdata/set_foreach_objects.a", true, false},
	{"internal/oracle/testdata/map_foreach_closures.a", true, false},
	{"internal/oracle/testdata/set_foreach_closures.a", true, false},
	{"internal/oracle/testdata/map_foreach_numbers.a", true, false},
	{"internal/oracle/testdata/set_foreach_numbers.a", true, false},
	{"internal/oracle/testdata/map_foreach_arrays.a", true, false},
	{"internal/oracle/testdata/set_foreach_arrays.a", true, false},
	{"internal/oracle/testdata/map_foreach_named_more.a", true, false},
	{"internal/oracle/testdata/set_foreach_named_more.a", true, false},
	{"internal/oracle/testdata/map_foreach_named_objects.a", true, false},
	{"internal/oracle/testdata/set_foreach_named_objects.a", true, false},
	{"internal/oracle/testdata/map_foreach_named_closures.a", true, false},
	{"internal/oracle/testdata/set_foreach_named_closures.a", true, false},
	{"internal/oracle/testdata/map_foreach_fnexpr.a", true, false},
	{"internal/oracle/testdata/set_foreach_fnexpr.a", true, false},
	{"internal/oracle/testdata/map_foreach_named_numbers.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_guards.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_callbacks.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_exclusions.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_spread.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_scope.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_depths.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_values.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_return.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_method.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_finally.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_parents.a", true, false},
	{"internal/oracle/testdata/borrow_chain_coverage_bounded.a", true, false},
	{"internal/oracle/testdata/borrow_chain_override.a", true, false},
	{"internal/oracle/testdata/borrow_chain_unknown.a", true, false},
	{"internal/oracle/testdata/borrow_chain_argument.a", true, false},
	{"internal/oracle/testdata/borrow_chain_walk.a", true, false},
	{"internal/oracle/testdata/borrow_chain_listener.a", true, false},
	{"internal/oracle/testdata/borrow_chain_write.a", true, false},
	{"internal/oracle/testdata/borrow_chain_reassigned.a", true, false},
	{"internal/oracle/testdata/borrow_chain_capture.a", true, false},
	{"internal/oracle/testdata/borrow_chain_store.a", true, false},
	{"internal/oracle/testdata/moves/accepted/objects.a", true, false},
	{"internal/oracle/testdata/async_plain.a", true, false},
	{"internal/oracle/testdata/async_coverage_unions.a", true, false},
	{"internal/oracle/testdata/async_coverage_reject_empty.a", true, false},
	{"internal/oracle/testdata/async_coverage_parameters.a", true, false},
	{"internal/oracle/testdata/async_coverage_typeof.a", true, false},
	{"internal/oracle/testdata/async_coverage_values.a", true, false},
	{"internal/oracle/testdata/async_coverage_discard.a", true, false},
	{"internal/oracle/testdata/async_coverage_reject_eager.a", true, false},
	{"internal/oracle/testdata/async_coverage_reject_nested.a", true, false},
	{"internal/oracle/testdata/async_typeof.a", true, false},
	{"internal/oracle/testdata/async_three.a", true, false},
	{"internal/oracle/testdata/async_nested.a", true, false},
	{"internal/oracle/testdata/async_throw.a", true, false},
	{"internal/oracle/testdata/call_targets_element.a", true, false},
	{"internal/oracle/testdata/call_targets_region.a", true, false},
	{"internal/oracle/testdata/call_targets_reuse.a", true, false},
	{"internal/oracle/testdata/call_targets_closure.a", true, false},
	{"internal/oracle/testdata/call_targets_sort.a", true, false},
	{"internal/oracle/testdata/input_spread_local.a", true, false},
	{"internal/oracle/testdata/input_spread_ordinary.a", true, false},
	{"internal/oracle/testdata/library_object_keys.a", true, false},
	{"internal/oracle/testdata/library_object_is.a", true, false},
	{"internal/oracle/testdata/library_object_has_own.a", true, false},
	{"internal/oracle/testdata/library_object_assign.a", true, false},
	{"internal/oracle/testdata/library_object_freeze.a", true, false},
	{"internal/oracle/testdata/library_object_freeze_write.a", true, false},
	{"internal/oracle/testdata/library_object_order.a", true, false},
	{"internal/oracle/testdata/library_object_assign_fields.a", true, false},
	{"internal/oracle/testdata/library_object_freeze_alias.a", true, false},
	{"internal/oracle/testdata/library_object_freeze_assign.a", true, false},
	{"internal/oracle/testdata/library_object_same.a", true, false},
	{"internal/oracle/testdata/library_object_own.a", true, false},
	{"dedication/dedication.a", true, false},
	{"cmd/adamic/testdata/wasi/request.a", true, false},
	// October 6 coverage: deeper dispatch, field order, ownership and subclass holders.
	{"internal/oracle/testdata/class_oct6_deep.a", true, false},
	{"internal/oracle/testdata/class_oct6_parameters.a", true, false},
	{"internal/oracle/testdata/class_oct6_release.a", true, false},
	{"internal/oracle/testdata/class_oct6_subclass_holder.a", true, false},

	{"internal/oracle/testdata/library_array_join.a", true, false},
	{"internal/oracle/testdata/library_array_iterators.a", true, false},
	{"internal/oracle/testdata/library_array_metadata.a", true, false},
	{"internal/oracle/testdata/library_array_with.a", true, false},
	{"internal/oracle/testdata/library_array_flat_map.a", true, false},
	{"internal/oracle/testdata/library_array_flat.a", true, false},
	{"internal/oracle/testdata/library_array_spliced.a", true, false},
	{"internal/oracle/testdata/library_array_copy_within.a", true, false},
	{"internal/oracle/testdata/library_array_search.a", true, false},
	{"internal/oracle/testdata/library_array_copy.a", true, false},
	{"internal/oracle/testdata/library_array_find_last.a", true, false},
	{"internal/oracle/testdata/json_stringify_scalars.a", true, false},
	{"internal/oracle/testdata/json_stringify_values.a", true, false},
	{"internal/oracle/testdata/json_stringify_options.a", true, false},
	{"internal/oracle/testdata/json_stringify_escapes.a", true, false},
	{"internal/oracle/testdata/json_stringify_numbers.a", true, false},
	{"internal/oracle/testdata/json_stringify_undefined.a", true, false},
	{"internal/oracle/testdata/json_stringify_indent.a", true, false},
	{"internal/oracle/testdata/json_stringify_keys.a", true, false},
	{"internal/oracle/testdata/json_stringify_replacer.a", true, false},
	{"internal/oracle/testdata/library_function_expressions.a", true, false},
	{"internal/oracle/testdata/library_fnexpr_recurse.a", true, false},
	{"internal/oracle/testdata/library_fnexpr_store.a", true, false},
	{"internal/oracle/testdata/library_fnexpr_loops.a", true, false},
	{"internal/oracle/testdata/library_for_in.a", true, false},
	{"internal/oracle/testdata/library_for_in_keys.a", true, false},
	{"internal/oracle/testdata/library_for_in_live.a", true, false},
	{"internal/oracle/testdata/library_globals.a", true, false},
	{"internal/oracle/testdata/library_globals_typeof.a", true, false},
	{"internal/load/testdata/0.1/compile/01_hello.ts", true, false},
	{"internal/load/testdata/0.1/compile/02_fizzbuzz.ts", true, false},
	{"internal/load/testdata/0.1/compile/03_shapes.ts", true, false},
	{"internal/load/testdata/0.1/compile/04_closures.ts", true, false},
	{"internal/load/testdata/0.1/compile/05_wordcount.ts", true, false},
	{"internal/load/testdata/0.1/compile/06_stack.ts", true, false},
	{"internal/load/testdata/0.1/compile/07_modules/main.ts", true, false},
	{"internal/load/testdata/0.1/compile/08_results.ts", true, false},
	{"internal/load/testdata/0.1/compile/09_tree.ts", true, false},
	{"internal/load/testdata/0.1/compile/10_unicode.ts", true, false},
	{"internal/oracle/testdata/strings.a", true, false},
	{"internal/oracle/testdata/string_build_caches.a", true, false},
	{"internal/oracle/testdata/string_build_padding.a", true, false},
	{"internal/oracle/testdata/string_build_join.a", true, false},
	{"internal/oracle/testdata/string_build_repeat.a", true, false},
	{"internal/oracle/testdata/string_build_cached_reads.a", true, false},
	{"internal/oracle/testdata/string_build_boundaries.a", true, false},
	{"internal/oracle/testdata/string_build_calls.a", true, false},
	{"internal/oracle/testdata/numbers.a", true, false},
	{"internal/oracle/testdata/bitwise_sweep.a", true, false},
	{"internal/oracle/testdata/loops.a", true, false},
	{"internal/oracle/testdata/booleans.a", true, false},
	{"internal/oracle/testdata/shadowing.a", true, false},
	{"internal/oracle/testdata/functions.a", true, false},
	{"internal/oracle/testdata/effects.a", true, false},
	{"internal/oracle/testdata/dead_zone.a", true, false},
	{"internal/oracle/testdata/objects.a", true, false},
	{"internal/oracle/testdata/modules/main.a", true, false},
	{"internal/oracle/testdata/panic.a", true, false},
	{"internal/oracle/testdata/maps_and_text.a", true, false},
	{"internal/oracle/testdata/lone_surrogates.a", true, false},
	{"internal/oracle/testdata/sorting.a", true, false},
	{"internal/oracle/testdata/classes.a", true, false},
	{"internal/oracle/testdata/closures.a", true, false},
	// A local console is the program's, and lowering it as the prelude's would print what the program
	// never asked to print.
	{"internal/oracle/testdata/local_console.a", true, false},
	{"internal/oracle/testdata/indexing.a", true, false},
	{"internal/oracle/testdata/writes.a", true, false},
	{"internal/oracle/testdata/writes_past_end.a", true, true},
	{"internal/oracle/testdata/casts.a", true, false},
	{"internal/oracle/testdata/cast_fails.a", true, true},
	{"internal/oracle/testdata/updates.a", true, false},
	{"internal/oracle/testdata/read_order.a", true, false},
	{"internal/oracle/testdata/string_index.a", true, false},
	{"internal/oracle/testdata/visits.a", true, false},
	{"internal/oracle/testdata/searches.a", true, false},
	{"internal/oracle/testdata/spreads.a", true, false},
	{"internal/oracle/testdata/maybe_numbers.a", true, false},
	{"internal/oracle/testdata/defaults.a", true, false},
	{"internal/oracle/testdata/search_halves.a", true, false},
	{"internal/oracle/testdata/runtime_last_index_of.a", true, false},
	{"internal/oracle/testdata/lint_runtime_release_chain.a", true, false},
	{"internal/oracle/testdata/lint_runtime_release_shared.a", true, false},
	{"internal/oracle/testdata/lint_runtime_search_boundaries.a", true, false},
	{"internal/oracle/testdata/lint_runtime_search_calls.a", true, false},
	{"internal/oracle/testdata/lint_runtime_equal_headers.a", true, false},
	{"internal/oracle/testdata/strings_more.a", true, false},
	{"internal/oracle/testdata/number_parsing.a", true, false},
	{"internal/oracle/testdata/library_math_number_math.a", true, false},
	{"internal/oracle/testdata/library_math_number_convert.a", true, false},
	{"internal/oracle/testdata/library_math_number_prototype.a", true, false},
	// Every Math function ported from V8, printed in full, so a last bit that differs from Node shows.
	{"internal/oracle/testdata/navigation.a", true, false},
	{"internal/oracle/testdata/number_formats.a", true, false},
	{"internal/oracle/testdata/precision_range.a", true, false},
	{"internal/oracle/testdata/radixes.a", true, false},
	{"internal/oracle/testdata/radix_range.a", true, false},
	{"internal/oracle/testdata/optional_numbers.a", true, false},
	{"internal/oracle/testdata/map_iteration.a", true, false},
	{"internal/oracle/testdata/sorts.a", true, false},
	{"internal/oracle/testdata/sort_releases.a", true, false},
	{"internal/oracle/testdata/timsort.a", true, false},
	{"internal/oracle/testdata/unused_parameters.a", true, false},
	{"internal/oracle/testdata/map_shrinks.a", true, true},
	{"internal/oracle/testdata/find_shrinks.a", true, true},
	{"internal/oracle/testdata/find_index_shrinks.a", true, true},
	{"internal/oracle/testdata/self_assignments.a", true, false},
	{"internal/oracle/testdata/method_closures.a", true, false},
	{"internal/oracle/testdata/splice_empty.a", true, false},
	// A narrowing outlives a call that assigns the variable again: what was narrowed away is checked.
	{"internal/oracle/testdata/narrowed_reads.a", true, false},
	{"internal/oracle/testdata/narrowed_writes.a", true, false},
	{"internal/oracle/testdata/narrowed_methods.a", true, false},
	{"internal/oracle/testdata/narrowed_fields.a", true, false},
	{"internal/oracle/testdata/narrowed_numbers.a", true, true},
	{"internal/oracle/testdata/narrowed_compared.a", true, false},
	// The same sorts at the top level, where only the globals' release at exit lets the leak check see.
	{"internal/oracle/testdata/sort_top_level.a", true, false},
	{"internal/oracle/testdata/splices.a", true, false},
	{"internal/oracle/testdata/fills.a", true, false},
	{"internal/oracle/testdata/fill_length.a", true, false},
	{"internal/oracle/testdata/array_from.a", true, false},
	{"internal/oracle/testdata/array_from_length.a", true, false},
	{"internal/oracle/testdata/array_from_undefined.a", true, false},
	{"internal/oracle/testdata/weak_parent.a", true, false},
	{"internal/oracle/testdata/doubly_linked.a", true, false},
	// The cycle finder's relaxation for fresh writes: a parser pushing fresh children into the node it
	// parses, and every other kind of write it proves (fresh_test.go holds the probes it refuses).
	{"internal/oracle/testdata/fresh_parser.a", true, false},
	{"internal/oracle/testdata/fresh_writes.a", true, false},
	// The holder proof across calls: a helper's push judged where it's called.
	{"internal/oracle/testdata/fresh_calls.a", true, false},
	{"internal/oracle/testdata/weak_narrowed.a", true, false},
	{"internal/oracle/testdata/exceptions.a", true, false},
	{"internal/oracle/testdata/exceptions_uncaught.a", true, false},
	{"internal/oracle/testdata/exceptions_empty.a", true, false},
	{"internal/oracle/testdata/closures_throw.a", true, false},
	{"internal/oracle/testdata/closures_throw_uncaught.a", true, false},
	{"internal/oracle/testdata/finally_leaves.a", true, false},
	{"internal/oracle/testdata/reuse_foreach_global.a", true, false},
	{"internal/oracle/testdata/param_assigned_in_try.a", true, false},
	{"internal/oracle/testdata/named_function_values.a", true, false},
	{"internal/oracle/testdata/panic_in_try.a", true, false},
	{"internal/oracle/testdata/invariance_readonly.a", true, false},
	{"internal/oracle/testdata/tuples_kept.a", true, false},
	{"internal/oracle/testdata/undefined_keys.a", true, false},
	{"internal/oracle/testdata/undefined_strings.a", true, false},
	{"internal/oracle/testdata/maybe_booleans.a", true, false},
	{"internal/oracle/testdata/maybe_boolean_panic.a", true, false},
	{"internal/oracle/testdata/unions.a", true, false},
	{"internal/oracle/testdata/maybe_number_slots.a", true, false},
	{"internal/oracle/testdata/case_mapping.a", true, false},
	{"internal/oracle/testdata/undefined_elements.a", true, false},
	{"internal/oracle/testdata/map_zero_keys.a", true, false},
	{"internal/oracle/testdata/string_limits.a", true, false},
	{"internal/oracle/testdata/string_too_long.a", true, false},
	{"internal/oracle/testdata/pad_too_long.a", true, false},
	{"internal/oracle/testdata/stack_overflow.a", true, false},
	{"internal/oracle/testdata/adversarial_order.a", true, false},
	{"internal/oracle/testdata/adversarial_exits.a", true, false},
	{"internal/oracle/testdata/adversarial_iteration.a", true, false},
	{"internal/oracle/testdata/number_edges.a", true, false},
	{"internal/oracle/testdata/long_chain.a", true, false},
	{"internal/oracle/testdata/write_after_shrink.a", true, true},
	{"internal/oracle/testdata/normalize.a", true, false},
	{"internal/oracle/testdata/normalize_form.a", true, false},
	{"internal/oracle/testdata/string_positions.a", true, false},
	{"internal/oracle/testdata/long_literals.a", true, false},
	{"internal/oracle/testdata/class_layouts.a", true, false},
	{"internal/oracle/testdata/field_access_paths.a", true, false},
	{"internal/oracle/testdata/nbody_field_values.a", true, false},
	{"internal/oracle/testdata/nbody_slot_kinds.a", true, false},
	{"internal/oracle/testdata/nbody_runtime_fields.a", true, false},
	{"internal/oracle/testdata/nbody_narrowed_receiver.a", true, false},
	{"internal/oracle/testdata/nbody_base_writes.a", true, false},
	{"internal/oracle/testdata/nbody_collection_fields.a", true, false},
	{"internal/oracle/testdata/nbody_write_read_order.a", true, false},
	{"internal/oracle/testdata/nbody_static_collision.a", true, false},
	{"internal/oracle/testdata/nbody_optional_references.a", true, false},
	{"internal/oracle/testdata/nbody_spread_fields.a", true, false},
	{"internal/oracle/testdata/field_write_paths.a", true, false},
	{"internal/oracle/testdata/inherited_static_field_read.a", true, false},
	{"internal/oracle/testdata/ascii_scan.a", true, false},
	{"internal/oracle/testdata/size_class_churn.a", true, false},
	// Borrowed parameters: a reassigned one has to stay owned, and so does a closure's, which map hands
	// an element it may overwrite. Each breaks under ASan if it's borrowed.
	{"internal/oracle/testdata/borrow_reassigned.a", true, false},
	// A narrowed read (ir.Defined) of a global, a field and a captured variable, lent to a call whose
	// later argument writes the place it was read from (integration's reading of aa17d3c).
	{"internal/oracle/testdata/borrow_defined_lent.a", true, false},
	{"internal/oracle/testdata/borrow_defined_lent_field.a", true, false},
	{"internal/oracle/testdata/writes_in_try.a", true, false},
	{"internal/oracle/testdata/class_as_interface.a", true, false},
	{"internal/oracle/testdata/optional_class_method.a", true, false},
	{"internal/oracle/testdata/set_undefined.a", true, false},
	{"internal/oracle/testdata/borrow_map_overwrite.a", true, false},
	// A spread is read before its fields' values, as JavaScript reads it.
	{"internal/oracle/testdata/spread_snapshot.a", true, false},
	// Reuse in place: taken only where nothing can tell, and where something could, never.
	{"internal/oracle/testdata/reuse.a", true, false},
	// Reuse for arrays: a map in place, a spread appended to, a splice with nothing to return.
	{"internal/oracle/testdata/reuse_arrays.a", true, false},
	// Reviewer R's probes of integration 5, each a way reuse in place was once seen: a borrowed
	// parameter moved on, a global moved out while another argument reads it, and a value a Weak
	// reaches taken over, during the spread and after it.
	{"internal/oracle/testdata/reuse_forward.a", true, false},
	{"internal/oracle/testdata/reuse_global_sibling.a", true, false},
	{"internal/oracle/testdata/reuse_weak_during_spread.a", true, false},
	{"internal/oracle/testdata/reuse_weak_after_reuse.a", true, false},
	// Regions: a statement's fresh values let go of together, and every way one could escape kept off it.
	{"internal/oracle/testdata/regions.a", true, false},
	// Reviewer R's round 8: a throw out of a statement with a region ends the region on its way out.
	{"internal/oracle/testdata/regions_throw.a", true, false},
	// A constructor whose object a closure captures, kept in a global (integration's reading of
	// fa49e43): the object outlives its statement, so no region.
	{"internal/oracle/testdata/regions_constructor_capture.a", true, false},
	// A variable borrowed from an array, beside every way the array could lose the element while it lives.
	{"internal/oracle/testdata/borrow_element.a", true, false},
	{"internal/oracle/testdata/borrow_loop.a", true, false},
	{"internal/oracle/testdata/borrow_loop_calls.a", true, false},
	{"internal/oracle/testdata/borrow_global_call.a", true, false},
	{"internal/oracle/testdata/borrow_element_throw.a", true, false},
	{"internal/oracle/testdata/borrow_element_virtual_store.a", true, false},
	// A variable borrowed from an array, then the array moved into a consumed parameter of a
	// function that only reads it, through a virtual call and through super (integration's reading
	// of aa17d3c): the array is never moved while something borrows from it.
	{"internal/oracle/testdata/borrow_element_virtual_move.a", true, false},
	{"internal/oracle/testdata/borrow_element_super_move.a", true, false},
	// Reviewer R's round 7: a throw between a move or an in-place spread and a catch that reads what
	// was moved or spread.
	{"internal/oracle/testdata/move_throw.a", true, false},
	// An assignment that throws never gives its variable a value, so the statement before it can't
	// take that variable's old value while a catch, a finally or the code after a swallowing catch
	// reads it; and a field or element write whose value throws.
	{"internal/oracle/testdata/throw_keeps_old_value.a", true, false},
	{"internal/oracle/testdata/throw_keeps_old_value_variants.a", true, false},
	{"internal/oracle/testdata/throw_in_writes.a", true, false},
	{"internal/oracle/testdata/throw_global_move.a", true, false},
	// Reviewer R's round 8b: a spread of a value that may be undefined is {} with the literal's fields.
	{"internal/oracle/testdata/spread_undefined.a", true, false},
	// A read lent without a count, beside a call that reassigns what was read.
	{"internal/oracle/testdata/lent_reads.a", true, false},
	// Output beyond native's stdout buffer, and the points where it must be flushed (adamic.c).
	{"internal/oracle/testdata/large_output.a", true, false},
	{"internal/oracle/testdata/output_then_panic.a", true, false},
	{"internal/oracle/testdata/interleaved.a", true, false},
	// sin, cos and tan where reducing by pi / 2 cancels the most bits, as no sweep input does.
	{"internal/oracle/testdata/trig_reduction.a", true, false},
	{"internal/oracle/testdata/sets.a", true, false},
	{"internal/oracle/testdata/set_maybe_numbers.a", true, false},
	{"internal/oracle/testdata/library_map_set.a", true, false},
	{"internal/oracle/testdata/library_map_set_keys.a", true, false},
	{"internal/oracle/testdata/library_map_set_iterators.a", true, false},
	{"internal/oracle/testdata/library_map_set_construct.a", true, false},
	{"internal/oracle/testdata/library_map_set_group_by.a", true, false},
	{"internal/oracle/testdata/maybe_collections.a", true, false},
	// Tail calls keep their frames, so recursion runs out of stack as on Node (native.go).
	{"internal/oracle/testdata/stack_tail_call.a", true, false},
	{"internal/oracle/testdata/stack_forever.a", true, false},
	{"internal/oracle/testdata/optional_strings.a", true, false},
	{"internal/oracle/testdata/concat_too_long.a", true, false},
	{"internal/oracle/testdata/replace_all_large.a", true, false},
	// A file read or written after console.log: what was printed is out first. output_test.go also runs
	// these with both streams on one pipe and stdin answered after the prompt; killed_after_output.a
	// is only there, since it runs until a signal stops it.
	{"internal/oracle/testdata/write_stdout_order.a", true, false},
	{"internal/oracle/testdata/write_stderr_order.a", true, false},
	{"internal/oracle/testdata/prompt_then_read.a", true, false},
	{"internal/oracle/testdata/collections.a", true, false},
	{"internal/oracle/testdata/gaps.a", true, false},
	// A run of marks longer than any the normalize sweep has, for canonical ordering.
	{"internal/oracle/testdata/normalize_long_marks.a", true, false},
	{"internal/oracle/testdata/normalize_coverage_quick.a", true, false},
	{"internal/oracle/testdata/normalize_coverage_repeat.a", true, false},
	{"internal/oracle/testdata/normalize_coverage_stream.a", true, false},
	{"internal/oracle/testdata/normalize_coverage_long.a", true, false},
	{"internal/oracle/testdata/normalize_coverage_cache.a", true, false},
	{"internal/oracle/testdata/normalize_coverage_limit.a", true, false},
	// String() of the powers of two where the closest digits of the shortest length aren't the ones that
	// read back (reviewer R, from the first night's number.c).
	{"internal/oracle/testdata/power_of_two_string.a", true, false},
	// Signatures before bodies: calls to functions and methods declared later, mutual recursion, and
	// declared functions as values.
	{"internal/oracle/testdata/declared_later.a", true, false},
	// return panic('why'), in functions of each kind of result and in an arrow, where it doesn't run
	// and where it does.
	{"internal/oracle/testdata/return_panic.a", true, false},
	{"internal/oracle/testdata/return_panic_fires.a", true, false},
	// String.fromCharCode over ToUint16's edges, and String.fromCodePoint, surrogates joined into pairs
	// as JavaScript joins them, and its RangeError.
	{"internal/oracle/testdata/from_codes.a", true, false},
	{"internal/oracle/testdata/from_code_point_fails.a", true, false},
	// &, |, ^, ~, <<, >> and >>>, swept over every pair of 36 edge values against Node.
	{"internal/oracle/testdata/bitwise.a", true, false},
	// Tuples as values: returned, passed, held, destructured and read by index.
	{"internal/oracle/testdata/tuple_values.a", true, false},
	// Generic functions per instantiation, undefined where a reference goes, ?.length, and spread
	// arguments.
	{"internal/oracle/testdata/generic_functions.a", true, false},
	{"internal/oracle/testdata/generic_method_return.a", true, false},
	{"internal/oracle/testdata/generic_values.a", true, false},
	{"internal/oracle/testdata/undefined_references.a", true, false},
	{"internal/oracle/testdata/spread_calls.a", true, false},
	// A string's UTF-8 read in place, and past its end.
	{"internal/oracle/testdata/utf8_view.a", true, false},
	{"internal/oracle/testdata/utf8_view_fails.a", true, false},
	// indexOf and includes from a position.
	{"internal/oracle/testdata/search_from.a", true, false},
	{"internal/oracle/testdata/shared_slices.a", true, false},
	{"internal/oracle/testdata/string_append.a", true, false},
	{"internal/oracle/testdata/shared_slice_append.a", true, false},
	{"internal/oracle/testdata/search_from_sweep.a", true, false},
	// Numbers as text around the integer fast path: every power of two and of ten, their neighbors
	// and negatives, -0, and the safe range's edges.
	{"internal/oracle/testdata/integer_format.a", true, false},
	// Reuse meeting exceptions, narrowing and lending (integration 7): a throw after a move, a narrowed
	// field emptied in place, and a lent global moved under its borrower.
	{"internal/oracle/testdata/reuse_throw.a", true, false},
	{"internal/oracle/testdata/reuse_narrowed.a", true, false},
	{"internal/oracle/testdata/reuse_lent_global.a", true, false},
	// A method called on a spread's source inside the literal runs code with the source as this
	// (integration's reading of aa17d3c): the source is not only read there, so it isn't reused.
	{"internal/oracle/testdata/reuse_spread_method.a", true, false},
	{"internal/oracle/testdata/reuse_spread_method_alias.a", true, false},
	// Assignments inside a try (integration 9): a parameter assigned there, a counter assigned there,
	// and a counted loop inside one.
	{"internal/oracle/testdata/try_assignments.a", true, false},
	{"internal/oracle/testdata/library_string_conversion.a", true, false},
	{"internal/oracle/testdata/library_string_prototype.a", true, false},
	{"internal/oracle/testdata/library_string_indices.a", true, false},
	{"internal/oracle/testdata/library_string_raw.a", true, false},
	{"internal/oracle/testdata/library_string_existing.a", true, false},
	// Map and Set visits, SameValueZero keys, ES2025 set arguments, groupBy collisions, and next() after done.
	{"internal/oracle/testdata/library_map_set_visit.a", true, false},
	{"internal/oracle/testdata/library_map_set_setops.a", true, false},
	{"internal/oracle/testdata/library_map_set_zeros.a", true, false},
	{"internal/oracle/testdata/library_map_set_groupby_keys.a", true, false},
	{"internal/oracle/testdata/library_map_set_next.a", true, false},
	// Object.prototype.hasOwnProperty is on every object's type and not in its shape.
	{"internal/oracle/testdata/has_own.a", true, false},
	{"internal/oracle/testdata/regexp.a", true, false},
	{"internal/oracle/testdata/sweeps/regexp_methods.a", true, false},
	{"internal/oracle/testdata/regexp_matchall_nonglobal.a", true, false},
	{"internal/oracle/testdata/regexp_replaceall_nonglobal.a", true, false},
	{"internal/oracle/testdata/regexp_null_narrowed.a", true, false},
	{"internal/oracle/testdata/regexp_replace.a", true, false},
	{"internal/oracle/testdata/regexp_split.a", true, false},
	{"internal/oracle/testdata/regexp_exec.a", true, false},
	{"internal/oracle/testdata/regexp_match.a", true, false},
	{"internal/oracle/testdata/regexp_search.a", true, false},
	{"internal/oracle/testdata/regexp_unicode.a", true, false},
	{"internal/oracle/testdata/regexp_split_pair_pattern.a", true, false},
}

// run is one execution's observable behavior: what the oracle compares.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

func execute(t *testing.T, name string, arguments ...string) run {
	t.Helper()
	return executeWith(t, nil, name, arguments...)
}

// leakSanitizer asks AddressSanitizer to check for leaks where it can. macOS's has no leak detector
// and aborts when asked for one; there the leak check is internal/leakcheck's counted build.
func leakSanitizer() string {
	if runtime.GOOS == "darwin" {
		return "ASAN_OPTIONS=detect_leaks=0"
	}
	return "ASAN_OPTIONS=detect_leaks=1"
}

// executeWith runs a command with environment added to the test's own; a later value for the same
// name wins, so a sanitizer setting here can't be overridden by one inherited from the shell.
func executeWith(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := bounded(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	result := run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
	rememberRun(t, result)
	return result
}

// bounded is a command that can't outlive its test: it has a deadline, it runs in a process group of
// its own, and when the deadline passes or the test ends, the whole group is killed. A fixture that
// loops, or a child left with nowhere to write, is stopped instead of orphaned.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	return command
}

// onNode runs a program's source on Node: the oracle.
func onNode(t *testing.T, path string) run {
	t.Helper()
	return cachedNode(t, path, func() run {
		return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
	})
}

// onJavaScriptBackend runs a lowered program through the JavaScript backend, on Node.
func onJavaScriptBackend(t *testing.T, program *ir.Program) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}

// lowered checks and lowers a program, or returns stage 0's refusal.
func lowered(t *testing.T, path string) (*ir.Program, error) {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range program.Files() {
		rememberImports(absolute, file.Text(), program)
	}
	return lower.Lower(context.Background(), program)
}

// nativeVariant is a native build a fixture runs as, each with its own runtime library (identity's
// libraries, in this order) and its own cached result.
type nativeVariant int

const (
	sanitizedBuild nativeVariant = iota
	releaseBuild
	countedBuild
	slabsBuild
)

// natively builds a lowered program under the sanitizers and runs it. It returns the binary too, so
// the leak check on Linux can run the same one again.
//
// On Linux, ASan carries LeakSanitizer and runs it at exit by default. This run is the comparison, and
// a program that panics exits 70 holding what it held, which isn't a leak, so leak detection is off
// here and the leak check is a run of its own. macOS's ASan has no leak detection to turn off.
// released builds a lowered program as a user's build is made, without sanitizers, and runs it.
func released(t *testing.T, program *ir.Program) run {
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		identity(t).cache.misses[nativeResults].Add(1)
		return releasedUncached(t, program)
	}
	return cachedNative(t, program, releaseBuild).Run.run()
}

// slabbed builds a lowered program under the sanitizers with the size-class allocator kept on (heap.c),
// and runs it. The release build runs the classes too, but with nothing to catch a slot read past its
// class or a class index past the table; here ASan and UBSan watch them. A leak into a chunk can't be
// seen here (the chunk stays reachable), which is why the comparison build stays on malloc.
func slabbed(t *testing.T, program *ir.Program) run {
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		identity(t).cache.misses[nativeResults].Add(1)
		return slabbedUncached(t, program)
	}
	return cachedNative(t, program, slabsBuild).Run.run()
}
func natively(t *testing.T, program *ir.Program) (run, string) {
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		identity(t).cache.misses[nativeResults].Add(1)
		return nativelyUncached(t, program)
	}
	return cachedNative(t, program, sanitizedBuild).Run.run(), ""
}
func leaks(t *testing.T, program *ir.Program, sanitized string) string {
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		identity(t).cache.misses[nativeResults].Add(1)
		return leaksUncached(t, program, sanitized)
	}
	return string(cachedNative(t, program, sanitizedBuild).LeakReport)
}

func releasedUncached(t *testing.T, program *ir.Program) run {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "release")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	return execute(t, binary)
}

func slabbedUncached(t *testing.T, program *ir.Program) run {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "slabs")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true, Slabs: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return executeWith(t, environment, binary)
}

func nativelyUncached(t *testing.T, program *ir.Program) (run, string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return executeWith(t, environment, binary), binary
}

// leaks returns a report of everything a finished program never let go of, or "" when it let go of
// everything: the leak check every test of a native program runs (internal/leakcheck), with each
// command run as the oracle runs its own. Only programs Node finishes with exit 0 are asked.
func leaksUncached(t *testing.T, program *ir.Program, sanitized string) string {
	t.Helper()
	return leakChecked(t, native.C(program), sanitized)
}

// leakChecked is the leak check for a program's C and the sanitized binary built from it.
func leakChecked(t *testing.T, code string, sanitized string) string {
	t.Helper()
	report, err := leakcheck.Check(leakcheck.Program{
		C:         code,
		Sanitized: sanitized,
		Counted:   filepath.Join(t.TempDir(), "counted"),
		Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
			return leakRun(executeWith(t, environment, name, arguments...))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

// leakRun is a run as the leak check reads it.
func leakRun(result run) leakcheck.Run {
	return leakcheck.Run{Stdout: result.stdout, Stderr: result.stderr, ExitCode: result.exitCode}
}

// disagreement says how two runs differ, or "" when they don't.
func disagreement(oracle run, native run) string {
	switch {
	case oracle.exitCode != native.exitCode:
		return "exit codes differ"
	case !bytes.Equal(oracle.stdout, native.stdout):
		return "stdout differs"
	case !bytes.Equal(oracle.stderr, native.stderr):
		return "stderr differs"
	}
	return ""
}

func TestNativeAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range fixtures {
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if !fixture.lowers {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("want stage 0 to refuse with where and what, got %v", err)
				}
				t.Logf("not yet: %v", notYet)
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			oracle, backend := onNode(t, path), onJavaScriptBackend(t, program)
			if usesParallelMap(program) {
				if difference := disagreement(oracle, backend); difference != "" {
					t.Fatalf("JavaScript backend: %s", difference)
				}
				checkParallelVariants(t, program, oracle)
				return
			}
			native, sanitized := natively(t, program)
			// The build a user gets (clang -O2, no sanitizers, heap values from the size-class
			// allocator rather than malloc) must say exactly what the sanitized one did.
			if released := released(t, program); disagreement(native, released) != "" {
				t.Errorf("the release build: %s\nsanitized: exit %d, stdout %q, stderr %q\nrelease:   exit %d, stdout %q, stderr %q",
					disagreement(native, released), native.exitCode, native.stdout, native.stderr, released.exitCode, released.stdout, released.stderr)
			}
			// The size classes under the sanitizers must say exactly what malloc did too.
			if slabbed := slabbed(t, program); disagreement(native, slabbed) != "" {
				t.Errorf("the sanitized build with the size classes: %s\nsanitized: exit %d, stdout %q, stderr %q\nslabs:     exit %d, stdout %q, stderr %q",
					disagreement(native, slabbed), native.exitCode, native.stdout, native.stderr, slabbed.exitCode, slabbed.stdout, slabbed.stderr)
			}
			if fixture.checked {
				// The check fires, so the source on Node goes on where Adamic stops: hold native to the
				// backend that carries the same check, and make sure the check really did fire.
				if difference := disagreement(backend, native); difference != "" {
					t.Errorf("%s\nbackend: exit %d, stdout %q, stderr %q\nnative:  exit %d, stdout %q, stderr %q",
						difference, backend.exitCode, backend.stdout, backend.stderr, native.exitCode, native.stdout, native.stderr)
				}
				if native.exitCode != 70 || oracle.exitCode == 70 {
					t.Errorf("want the inserted check to fire natively (exit 70) where the source on Node runs on: native %d, Node %d", native.exitCode, oracle.exitCode)
				}
				return
			}
			if difference := disagreement(oracle, native); difference != "" {
				t.Errorf("%s\nnode:   exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q",
					difference, oracle.exitCode, oracle.stdout, oracle.stderr, native.exitCode, native.stdout, native.stderr)
			}
			// The JavaScript backend runs the same IR the native one compiled, with nothing of C, so
			// where it differs from the source the fault is in lowering.
			if difference := disagreement(oracle, backend); difference != "" {
				t.Errorf("JavaScript backend: %s\nnode:    exit %d, stdout %q, stderr %q\nbackend: exit %d, stdout %q, stderr %q",
					difference, oracle.exitCode, oracle.stdout, oracle.stderr, backend.exitCode, backend.stdout, backend.stderr)
			}
			// A program that panicked stopped where it stood, as Node's does, so what it held then
			// isn't a leak; every program that finishes must have let go of everything.
			if oracle.exitCode == 0 {
				if leaked := leaks(t, program, sanitized); leaked != "" {
					t.Errorf("leaks:\n%s", leaked)
				}
			}
		})
	}
}

// The oracle has to be able to fail. One byte added to the dedication's string, after lowering, must
// read as a disagreement.
func TestTheOracleCatchesOneByte(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "dedication", "dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	program.Strings[0] += "!"
	native, _ := natively(t, program)
	if difference := disagreement(onNode(t, path), native); difference != "stdout differs" {
		t.Errorf("got %q, want the mutant caught as \"stdout differs\"", difference)
	}
}

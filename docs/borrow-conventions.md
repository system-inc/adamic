# Borrowing calling conventions

The feature branch is `codex/borrow-conventions`. Main and area branches were
never written or merged into. The routing base is
`f64641f388f32d877522140e174dd0b1dcb6d13e`; the initial main base was `e011f8f`.
The three implementation pushes were `746b216`, `7845b01`, and `41220c4`.
Current main `e8ba3d5d81de4d3773c723914fccd4c76248b965` was merged into the feature
branch without rebasing or force pushing. No pull request was opened.

The convention, loop and return lifetime rules and the visitor's per-node table
are in [memory.md](memory.md#calling-conventions-through-call-targets-october-7-2026).
Read-only visitor traversal retains/releases per node: **3.984127 -> 2.984127 ->
1 -> 0**. The final merged compiler's four- and eight-round runs both count
346 retains and 352 releases. The tree has 63 nodes, and the eight-round walk
visits 1,008 nodes through two interface listeners. Node outputs agree.
These observations measure counts; runtime's stage-1 batch-8 timing was not run
by this worker.

The main merge needed two reconciliations. Structural exact-receiver dispatch
keeps devirtualization and the conditional absent-closure retain/release. A
nominal devirtualized call keeps the virtual borrowing ABI: a consuming target
acquires its own argument counts even when its table lookup is eliminated.
`borrow_devirtualized.a` holds this boundary with a live caller value and a
reusing field-spread implementation. `borrow_return.a` also checks both accessor
and ordinary data-field paths of the getter wrapper.

Setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 1s,
build cache warm 151s, done 151s. `nproc`: 5; cgroup quota: 4 CPUs.
Go 1.27.1, clang 20.1.8, Node 24.19.0. Every build shell sourced
`/workspace/adamic-tools/env.sh`.

## Mutants

| Mutant | Check that caught it | Log |
|---|---|---|
| Force storing callback `item` to borrow | ASan heap-use-after-free after replacing its last source owner, at the following push | `/tmp/borrow-step1-mutant.log` |
| Join only the first virtual implementation | `TestBorrowConventionJoinsEveryTarget` rejects a storing override | `/tmp/borrow-step1-mutant-first_target.log` |
| Make Unknown function targets borrowed | Same test rejects Unknown | `/tmp/borrow-step1-mutant-unknown_borrowed.log` |
| Omit a virtual adapter's consumed count | ASan heap-use-after-free in inheritance/reuse oracle | `/tmp/borrow-step1-mutant-adapter_count.log` |
| Force keeping/mutating array loop to borrow | ASan heap-use-after-free after source replacement, at the following push | `/tmp/borrow-step2-mutant.log` |
| Remove caller-scope field-preservation guard | ASan heap-use-after-free across `Parser.replace` | `/tmp/borrow-step3-mutant.log` |
| Omit devirtualized consuming argument count | ASan heap-use-after-free in `borrow_devirtualized.a` | `/tmp/borrow-merge-mutant-direct_boundary.log` |
| Omit ordinary-field fallback from borrowed getter wrapper | Node succeeds; native panics on missing field, oracle rejects exit/output mismatch | `/tmp/borrow-merge-mutant-getter_fallback.log` |

All mutants were restored. A push alone takes its own count and cannot expose a
lifetime violation while the source array remains stable; both storing probes
first replace the last source owner. The first devirtualization mutant attempt
on older polymorphic fixtures survived because those receivers were not proven
exact. The dedicated exact-receiver fixture above then killed it with ASan.
The first fallback probe had a console argument type error and was rejected by
loading; that was not counted as a killed mutant. After converting numeric
console arguments to strings, the restored fixture passed and the fallback
mutant reached native execution and failed the Node comparison.

## Count changes after integrating main

These are every changed numeric row relative to main `e8ba3d5`, not just the new
visitor. All other numeric rows are unchanged. Existing rows show signed deltas;
new rows show absolute counts. Table order changes from fixture registration are
not numeric changes. The complete final counts are in
[internal/oracle/counts.md](../internal/oracle/counts.md).

Paired decreases come from proved read-only parameters, stable borrowed array
loops, field-return specialization, and absent closure counts. Paired increases
come from refusing the earlier lowering flag for stored, returned, captured,
mutated, or unbounded-call arguments. This makes the convention stricter than
the old borrow optimization. A tree-walk improvement therefore does not imply
a count improvement for every unrelated program. The table records the net of
these rules; it does not claim an isolated causal measurement for each rule in
each row.

Two rows allocate one more value: `call_targets_reuse.a` and
`borrow_element_virtual_move.a`. The virtual adapter boundary now preserves the
caller's count, including at exact direct dispatch; a consuming spread/map
therefore cannot take over that caller's uniquely held object/array. The extra
copy accounts for the extra allocation, free and peak-live value. Virtual
caller moves are deliberately conservative. The remaining unequal retain and
release deltas occur on reuse/forwarding paths, structural method dispatch with
conditional absent-closure cleanup, or programs that panic before ordinary
scope cleanup. They are raw operation counts, including NULL calls and nested
field/element destruction, not a conservation equation for live references.
Their exact observations are shown rather than inferred from a paired total.

147 numeric rows changed, including 5 new fixtures.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| `call_targets_element.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `call_targets_region.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `call_targets_reuse.a` | +1 | +1 | +3 | 0 | +1 | 0 |
| `call_targets_closure.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `call_targets_sort.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `library_object_is.a` | 0 | 0 | +8 | +8 | 0 | 0 |
| `class_oct6_deep.a` | 0 | 0 | +14 | +14 | 0 | 0 |
| `class_oct6_parameters.a` | 0 | 0 | +12 | +12 | 0 | 0 |
| `class_oct6_release.a` | 0 | 0 | +117 | +117 | 0 | 0 |
| `class_oct6_subclass_holder.a` | 0 | 0 | +45 | +45 | 0 | 0 |
| `library_array_with.a` | 0 | 0 | +14 | +14 | 0 | 0 |
| `library_array_flat_map.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `library_array_flat.a` | 0 | 0 | -16 | -16 | 0 | 0 |
| `library_array_spliced.a` | 0 | 0 | +20 | +20 | 0 | 0 |
| `library_array_copy_within.a` | 0 | 0 | +104 | +104 | 0 | 0 |
| `library_array_copy.a` | 0 | 0 | -9 | -9 | 0 | 0 |
| `library_array_find_last.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `json_stringify_scalars.a` | 0 | 0 | +5 | +5 | 0 | 0 |
| `json_stringify_escapes.a` | 0 | 0 | +22 | +22 | 0 | 0 |
| `library_function_expressions.a` | 0 | 0 | -2 | -2 | 0 | 0 |
| `library_fnexpr_store.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `internal/load/testdata/0.1/compile/03_shapes.ts` | 0 | 0 | +3 | +3 | 0 | 0 |
| `internal/load/testdata/0.1/compile/06_stack.ts` | 0 | 0 | +32 | +32 | 0 | 0 |
| `internal/load/testdata/0.1/compile/07_modules/main.ts` | 0 | 0 | +1 | +1 | 0 | 0 |
| `internal/load/testdata/0.1/compile/09_tree.ts` | 0 | 0 | +19 | +19 | 0 | 0 |
| `objects.a` | 0 | 0 | -5 | -5 | 0 | 0 |
| `classes.a` | 0 | 0 | +13 | +13 | 0 | 0 |
| `closures.a` | 0 | 0 | -4 | -4 | 0 | 0 |
| `local_console.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `updates.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `visits.a` | 0 | 0 | -28 | -28 | 0 | 0 |
| `searches.a` | 0 | 0 | -13 | -13 | 0 | 0 |
| `spreads.a` | 0 | 0 | -2 | -2 | 0 | 0 |
| `defaults.a` | 0 | 0 | +15 | +15 | 0 | 0 |
| `search_halves.a` | 0 | 0 | -6 | -6 | 0 | 0 |
| `library_math_number_convert.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `sorts.a` | 0 | 0 | -66 | -66 | 0 | 0 |
| `sort_releases.a` | 0 | 0 | -28 | -28 | 0 | 0 |
| `timsort.a` | 0 | 0 | -2316 | -2316 | 0 | 0 |
| `method_closures.a` | 0 | 0 | +3 | +3 | 0 | 0 |
| `narrowed_methods.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `sort_top_level.a` | 0 | 0 | -28 | -28 | 0 | 0 |
| `splices.a` | 0 | 0 | +32 | +32 | 0 | 0 |
| `fills.a` | 0 | 0 | -3 | -3 | 0 | 0 |
| `array_from.a` | 0 | 0 | -34 | -34 | 0 | 0 |
| `array_from_length.a` | 0 | 0 | -3 | -3 | 0 | 0 |
| `array_from_undefined.a` | 0 | 0 | -4 | -4 | 0 | 0 |
| `weak_parent.a` | 0 | 0 | +35 | +35 | 0 | 0 |
| `doubly_linked.a` | 0 | 0 | +37 | +37 | 0 | 0 |
| `fresh_parser.a` | 0 | 0 | +23 | +23 | 0 | 0 |
| `fresh_writes.a` | 0 | 0 | +75 | +75 | 0 | 0 |
| `fresh_calls.a` | 0 | 0 | +86 | +86 | 0 | 0 |
| `exceptions.a` | 0 | 0 | +2 | +2 | 0 | 0 |
| `closures_throw.a` | 0 | 0 | -1025 | -1025 | 0 | 0 |
| `closures_throw_uncaught.a` | 0 | 0 | -5 | -5 | 0 | 0 |
| `reuse_foreach_global.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `param_assigned_in_try.a` | 0 | 0 | -2 | -2 | 0 | 0 |
| `named_function_values.a` | 0 | 0 | -18 | -18 | 0 | 0 |
| `invariance_readonly.a` | 0 | 0 | -6 | -6 | 0 | 0 |
| `tuples_kept.a` | 0 | 0 | +3 | +3 | 0 | 0 |
| `undefined_keys.a` | 0 | 0 | -18 | -18 | 0 | 0 |
| `undefined_strings.a` | 0 | 0 | +11 | +11 | 0 | 0 |
| `maybe_booleans.a` | 0 | 0 | +3 | +3 | 0 | 0 |
| `unions.a` | 0 | 0 | +23 | +23 | 0 | 0 |
| `maybe_number_slots.a` | 0 | 0 | -3 | -3 | 0 | 0 |
| `case_mapping.a` | 0 | 0 | +24 | +24 | 0 | 0 |
| `adversarial_order.a` | 0 | 0 | +3 | +3 | 0 | 0 |
| `adversarial_exits.a` | 0 | 0 | -16 | -16 | 0 | 0 |
| `normalize.a` | 0 | 0 | +10 | +10 | 0 | 0 |
| `class_layouts.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `writes_in_try.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `class_as_interface.a` | 0 | 0 | +82 | +89 | 0 | 0 |
| `optional_class_method.a` | 0 | 0 | -5 | +1 | 0 | 0 |
| `spread_snapshot.a` | 0 | 0 | +3 | +3 | 0 | 0 |
| `reuse.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `reuse_arrays.a` | 0 | 0 | -6 | -6 | 0 | 0 |
| `reuse_forward.a` | 0 | 0 | 0 | +1 | 0 | 0 |
| `regions.a` | 0 | 0 | +6 | +6 | 0 | 0 |
| `borrow_element.a` | 0 | 0 | +3 | +3 | 0 | 0 |
| `borrow_element_throw.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `borrow_element_virtual_store.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `borrow_element_virtual_move.a` | +1 | +1 | -2 | -5 | +1 | 0 |
| `borrow_element_super_move.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `sets.a` | 0 | 0 | -5 | -5 | 0 | 0 |
| `library_map_set.a` | 0 | 0 | +32 | +32 | 0 | 0 |
| `library_map_set_keys.a` | 0 | 0 | -3 | -3 | 0 | 0 |
| `library_map_set_construct.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `library_map_set_group_by.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `maybe_collections.a` | 0 | 0 | -4 | -4 | 0 | 0 |
| `optional_strings.a` | 0 | 0 | +8 | +8 | 0 | 0 |
| `collections.a` | 0 | 0 | -3 | -3 | 0 | 0 |
| `gaps.a` | 0 | 0 | -8 | -8 | 0 | 0 |
| `declared_later.a` | 0 | 0 | -2 | -2 | 0 | 0 |
| `return_panic.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `return_panic_fires.a` | 0 | 0 | +2 | +1 | 0 | 0 |
| `tuple_values.a` | 0 | 0 | -13 | -13 | 0 | 0 |
| `generic_functions.a` | 0 | 0 | +25 | +25 | 0 | 0 |
| `generic_method_return.a` | 0 | 0 | +6 | +6 | 0 | 0 |
| `generic_values.a` | 0 | 0 | +14 | +14 | 0 | 0 |
| `undefined_references.a` | 0 | 0 | +8 | +8 | 0 | 0 |
| `utf8_view.a` | 0 | 0 | +12 | +12 | 0 | 0 |
| `shared_slices.a` | 0 | 0 | +5 | +5 | 0 | 0 |
| `string_append.a` | 0 | 0 | -5 | -5 | 0 | 0 |
| `reuse_spread_method.a` | 0 | 0 | +2 | +2 | 0 | 0 |
| `reuse_spread_method_alias.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `library_string_conversion.a` | 0 | 0 | +6 | +6 | 0 | 0 |
| `library_string_raw.a` | 0 | 0 | +8 | +8 | 0 | 0 |
| `library_map_set_setops.a` | 0 | 0 | +7 | +7 | 0 | 0 |
| `library_map_set_zeros.a` | 0 | 0 | -3 | -3 | 0 | 0 |
| `regexp.a` | 0 | 0 | +14 | +14 | 0 | 0 |
| `sweeps/regexp_methods.a` | 0 | 0 | +2480 | +2480 | 0 | 0 |
| `regexp_exec.a` | 0 | 0 | +10 | +10 | 0 | 0 |
| `regexp_match.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `regexp_unicode.a` | 0 | 0 | +6 | +6 | 0 | 0 |
| `borrow_visitor.a (new)` | 192 | 192 | 346 | 352 | 161 | 0 |
| `borrow_target_store.a (new)` | 17 | 17 | 16 | 25 | 11 | 0 |
| `borrow_for_of_store.a (new)` | 11 | 11 | 9 | 16 | 7 | 0 |
| `borrow_return.a (new)` | 31 | 31 | 19 | 39 | 8 | 0 |
| `borrow_devirtualized.a (new)` | 7 | 7 | 4 | 9 | 6 | 0 |
| `class_features_static.a` | 0 | 0 | +30 | +30 | 0 | 0 |
| `class_features_static_private.a` | 0 | 0 | +49 | +49 | 0 | 0 |
| `class_features_private.a` | 0 | 0 | +8 | +8 | 0 | 0 |
| `class_features_accessors.a` | 0 | 0 | +4 | +4 | 0 | 0 |
| `class_features_twice.a` | 0 | 0 | +15 | +15 | 0 | 0 |
| `class_features_retained.a` | 0 | 0 | +39 | +39 | 0 | 0 |
| `class_features_distinct.a` | 0 | 0 | +15 | +15 | 0 | 0 |
| `class_inheritance.a` | 0 | 0 | +11 | +11 | 0 | 0 |
| `class_inheritance_exceptions.a` | 0 | 0 | +8 | +8 | 0 | 0 |
| `class_inheritance_order.a` | 0 | 0 | +14 | +14 | 0 | 0 |
| `class_inheritance_memory.a` | 0 | 0 | +3 | +3 | 0 | 0 |
| `class_inheritance_generic.a` | 0 | 0 | +104 | +104 | 0 | 0 |
| `class_inheritance_interface.a` | 0 | 0 | +1 | +9 | 0 | 0 |
| `class_inheritance_conditional.a` | 0 | 0 | +143 | +143 | 0 | 0 |
| `devirtualize.a` | 0 | 0 | -1 | -1 | 0 | 0 |
| `user_iterators.a` | 0 | 0 | +161 | +161 | 0 | 0 |
| `user_iterators_rest_tdz.a` | 0 | 0 | -2 | -1 | 0 | 0 |
| `literal_optional_shapes.a` | 0 | 0 | +11 | +11 | 0 | 0 |
| `e4eec87_f1_field_narrowed.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `e4eec87_f1_class_narrowed.a` | 0 | 0 | +2 | +1 | 0 | 0 |
| `e4eec87_f1_field_present.a` | 0 | 0 | +1 | +1 | 0 | 0 |
| `object_prototype.a` | 0 | 0 | -4 | -4 | 0 | 0 |
| `internal/fresh/testdata/regexp_tree.ts` | 0 | 0 | +4 | +4 | 0 | 0 |
| `regexp_cycle_fields.a` | 0 | 0 | +11 | +11 | 0 | 0 |
| `regexp_cycle_closures.a` | 0 | 0 | +2 | +2 | 0 | 0 |
| `read_files.a` | 0 | 0 | +11 | +11 | 0 | 0 |
| `write_files.a` | 0 | 0 | +32 | +32 | 0 | 0 |
| `walk.a` | 0 | 0 | +6 | +6 | 0 | 0 |

## Validation

Step 1: focused native/IR checks passed, call-target reader guard passed, and
uncached filtered Node oracle passed (3.778s). Full native package passed
(191.940s); the initial IR reader guard failure was corrected and rerun.
Counts regeneration passed (27.597s).

Step 2: focused native checks passed (0.257s), uncached borrowing/call-target
Node oracle passed (3.115s), and counts regeneration passed (21.923s).

Step 3 before the main merge: full native/IR packages passed (147.561s and
0.843s), the uncached filtered oracle passed (3.150s), and counts regeneration
passed (21.259s). The merge's restored focused checks passed (0.236s), the two
expanded fixtures passed (0.751s), and final counts regeneration passed
(21.668s). After resolving the main merge, the full native package passed in
218.491s, IR in 2.532s, and the complete uncached oracle in 207.039s. The oracle
includes source Node behavior, the JavaScript backend, native release builds,
ASan/UBSan and Linux LeakSanitizer. Vet exited 0 with no diagnostics; formatting
and diff checks printed nothing. Every test command wrote directly to a log file.

Final commands:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/ir ./internal/oracle -count=1 -timeout 30m > /tmp/borrow-final-gate.log 2>&1
go vet ./... > /tmp/borrow-final-vet.log 2>&1
gofmt -l cmd internal > /tmp/borrow-final-format.log
git diff --check > /tmp/borrow-final-diff.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/borrow-merge-counts.log 2>&1
```

The gate log contains:

```text
ok github.com/system-inc/adamic/internal/native 218.491s
ok github.com/system-inc/adamic/internal/ir 2.532s
ok github.com/system-inc/adamic/internal/oracle 207.039s
```

Final merged visitor logs are `/tmp/borrow-final-eight-count.log` and
`/tmp/borrow-final-four-count.log`. Both have 192 allocations, 192 frees,
346 retains, 352 releases, peak live 161, and regions 0. Their outputs are
`1008 visits: 1872` and `504 visits: 936`.

Not covered: a whole-repository gate, runtime's batch-8 stage-1 benchmark,
arbitrary computed borrowed returns or inline result borrowing, mutable or
captured receiver roots, unknown targets, and for-of patterns/maps/regex/string
iterators. Getter targets with different fields or effects stay owned. Current
main's lowering restriction on property-style function views of class prototype
methods remains; the visitor uses an interface method signature. No cohere code
was copied.

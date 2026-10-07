Built both requested merge commits and the open numeric enum ruling with shared never checks.
Merge SHAs: flag 246ecc073993de6a5f1bcb64346cde7938deec43; checked downcasts 982e2118f968788464256005e840b80e9b193254.
Validation: lower/load, full uncached oracle, counts, vet and formatting; final results below.
Mutants: eleven ruling mutants plus existing enum semantic, name enumeration and leak mutants.
Remaining: fixture 06 generic mutable write, 08/11 checker errors, 07 pinned sparse-array stop; conservative enum literal/tag boundary.

No main or area branch was merged into or pushed. Only codex/flag-enums and codex/checked-downcasts were pushed. The ruling belongs to codex/flag-enums.

## Integration

The flag merge had no textual conflicts. Its parents are 4097384 and e011f8f. Both existing features and the new main ownership layout survived. Lower/load passed in 25.393s/2.050s, full uncached oracle in 136.783s, counts in 37.282s. Vet and formatting were clean.

Checked downcasts had two textual conflicts. In internal/lower/refusals.go, retain checked-cast validation and definite-assignment assertion checks as successive complete blocks, including the generic string-widening exception. In internal/oracle/counts.md, retain cast rows and main's regexp/class rows, selecting the current regexp_unicode counts. Two nominalAncestor calls in cast_proof.go also needed main's visited-map argument. The enum cleanup mutant exposed stale conservative LeakSanitizer stack/register roots; disable these roots for this deliberately leaking test, retaining global roots. The final lower/load run passed in 40.485s/1.980s, full uncached oracle in 116.777s, counts verification in 13.562s. Vet and formatting were clean.

Counts regeneration moved these seven rows after the cast rows, without numerical changes: class_features_static, class_features_static_private, class_features_private, class_features_accessors, class_features_twice, class_features_retained, class_features_distinct. Other changes are inherited from the merged branches and main, listed exactly below. Tuple order is the existing counts header order.

### flag_merge

| Row | Before | After |
| --- | --- | --- |
| internal/oracle/testdata/borrow_element_virtual_move.a | [26, 24, 10, 22, 12, 2] | [25, 23, 11, 26, 11, 2] |

Inherited new rows:

- `internal/oracle/testdata/borrow_element_throw.a`: `38, 38, 34, 54, 8, 0`
- `internal/oracle/testdata/borrow_element_virtual_store.a`: `13, 11, 9, 15, 7, 2`

### checked_merge

| Row | Before | After |
| --- | --- | --- |
| internal/oracle/testdata/class_oct6_subclass_holder.a | [76, 76, 49, 110, 12, 0] | [76, 76, 55, 116, 12, 0] |
| internal/load/testdata/0.1/compile/09_tree.ts | [19, 19, 82, 123, 16, 0] | [19, 19, 85, 126, 16, 0] |
| internal/oracle/testdata/casts.a | [9, 9, 16, 24, 6, 0] | [9, 9, 17, 25, 6, 0] |
| internal/oracle/testdata/visits.a | [95, 95, 124, 208, 23, 0] | [95, 95, 125, 209, 23, 0] |
| internal/oracle/testdata/narrowed_reads.a | [5, 5, 0, 5, 4, 0] | [5, 5, 3, 7, 4, 0] |
| internal/oracle/testdata/narrowed_methods.a | [3, 3, 1, 5, 3, 0] | [3, 3, 3, 6, 3, 0] |
| internal/oracle/testdata/narrowed_fields.a | [4, 3, 2, 5, 4, 0] | [4, 3, 4, 6, 4, 0] |
| internal/oracle/testdata/fills.a | [24, 24, 27, 50, 11, 0] | [24, 24, 28, 51, 11, 0] |
| internal/oracle/testdata/fresh_writes.a | [330, 330, 407, 523, 132, 0] | [330, 330, 409, 525, 132, 0] |
| internal/oracle/testdata/fresh_calls.a | [147, 147, 266, 321, 39, 0] | [147, 147, 268, 323, 39, 0] |
| internal/oracle/testdata/weak_narrowed.a | [26, 26, 50, 69, 11, 0] | [26, 26, 53, 72, 11, 0] |
| internal/oracle/testdata/undefined_keys.a | [122, 122, 171, 236, 24, 0] | [122, 122, 173, 238, 24, 0] |
| internal/oracle/testdata/undefined_strings.a | [15, 15, 20, 42, 6, 0] | [15, 15, 21, 43, 6, 0] |
| internal/oracle/testdata/undefined_references.a | [17, 17, 36, 51, 11, 0] | [17, 17, 37, 52, 11, 0] |
| internal/oracle/testdata/reuse_narrowed.a | [7, 5, 4, 7, 5, 0] | [7, 5, 6, 8, 5, 0] |
| internal/oracle/testdata/reuse_lent_global.a | [10, 10, 7, 14, 7, 0] | [10, 10, 8, 15, 7, 0] |
| internal/oracle/testdata/regexp.a | [455, 455, 355, 408, 62, 0] | [455, 455, 369, 422, 62, 0] |
| internal/oracle/testdata/regexp_null_narrowed.a | [4, 4, 7, 5, 3, 0] | [4, 4, 8, 5, 3, 0] |
| internal/oracle/testdata/regexp_exec.a | [179, 179, 81, 159, 18, 0] | [179, 179, 82, 160, 18, 0] |
| internal/oracle/testdata/regexp_unicode.a | [108, 108, 99, 86, 21, 0] | [108, 108, 101, 88, 21, 0] |
| internal/oracle/testdata/class_inheritance_generic.a | [64, 64, 74, 121, 24, 0] | [89, 89, 104, 167, 39, 0] |

Inherited new rows:

- `internal/oracle/testdata/library_fnexpr_recurse.a`: `29, 29, 69, 99, 11, 0`
- `internal/oracle/testdata/library_fnexpr_store.a`: `43, 43, 30, 71, 17, 0`
- `internal/oracle/testdata/library_fnexpr_loops.a`: `182, 182, 197, 269, 82, 0`
- `internal/oracle/testdata/library_for_in_keys.a`: `38, 38, 110, 79, 34, 0`
- `internal/oracle/testdata/library_for_in_live.a`: `39, 39, 63, 80, 21, 0`
- `internal/oracle/testdata/library_globals_typeof.a`: `14, 14, 3, 20, 1, 0`
- `internal/oracle/testdata/borrow_defined_lent.a`: `8, 8, 1, 8, 6, 0`
- `internal/oracle/testdata/borrow_defined_lent_field.a`: `19, 19, 4, 17, 11, 0`
- `internal/oracle/testdata/regions_constructor_capture.a`: `15, 15, 12, 22, 8, 0`
- `internal/oracle/testdata/borrow_element_throw.a`: `38, 38, 34, 54, 8, 0`
- `internal/oracle/testdata/borrow_element_virtual_store.a`: `13, 11, 9, 15, 7, 2`
- `internal/oracle/testdata/borrow_element_virtual_move.a`: `25, 23, 11, 26, 11, 2`
- `internal/oracle/testdata/borrow_element_super_move.a`: `10, 9, 4, 9, 8, 1`
- `internal/oracle/testdata/throw_keeps_old_value.a`: `18, 18, 9, 21, 6, 0`
- `internal/oracle/testdata/throw_keeps_old_value_variants.a`: `26, 26, 11, 32, 6, 0`
- `internal/oracle/testdata/throw_in_writes.a`: `22, 22, 8, 22, 8, 0`
- `internal/oracle/testdata/throw_global_move.a`: `18, 18, 12, 26, 6, 0`
- `internal/oracle/testdata/shared_slice_append.a`: `200, 200, 20, 206, 8, 0`
- `internal/oracle/testdata/reuse_spread_method.a`: `10, 10, 7, 12, 8, 0`
- `internal/oracle/testdata/reuse_spread_method_alias.a`: `8, 8, 8, 15, 6, 0`
- `internal/oracle/testdata/regexp_split_pair_pattern.a`: `40, 40, 48, 78, 18, 0`
- `internal/oracle/testdata/class_features_static.a`: `110, 110, 84, 191, 28, 0`
- `internal/oracle/testdata/class_features_static_private.a`: `42, 42, 71, 112, 12, 0`
- `internal/oracle/testdata/class_features_private.a`: `47, 47, 37, 65, 15, 0`
- `internal/oracle/testdata/class_features_accessors.a`: `67, 67, 53, 110, 21, 0`
- `internal/oracle/testdata/class_features_twice.a`: `26, 26, 6, 35, 6, 0`
- `internal/oracle/testdata/class_features_retained.a`: `48, 48, 57, 88, 18, 0`
- `internal/oracle/testdata/class_features_distinct.a`: `48, 48, 50, 81, 14, 0`
- `internal/oracle/testdata/class_inheritance_conditional.a`: `164, 164, 148, 251, 35, 0`
- `internal/oracle/testdata/enums_flags.a`: `44, 44, 23, 64, 11, 0`
- `internal/oracle/testdata/enums_flags_never_default.a`: `2, 2, 6, 6, 2, 0`
- `internal/oracle/testdata/enums_flags_modules/main.a`: `9, 9, 6, 12, 5, 0`
- `stage3/fixtures/enums/01_token_range.a`: `9, 9, 1, 10, 3, 0`
- `stage3/fixtures/enums/02_node_range.a`: `7, 7, 1, 8, 3, 0`
- `stage3/fixtures/enums/03_jsdoc_range.a`: `10, 10, 1, 11, 4, 0`
- `stage3/fixtures/enums/04_parse_tree_mask.a`: `10, 10, 1, 11, 4, 0`
- `stage3/fixtures/enums/10_string_enum.a`: `32, 32, 80, 74, 6, 0`
- `stage3/fixtures/enums/12_diagnostic_reverse_lookup.a`: `34, 34, 29, 60, 4, 0`
- `stage3/fixtures/enums/13_regex_map_keys.a`: `6, 6, 0, 8, 3, 0`
- `internal/oracle/testdata/enums_names.a`: `24, 24, 35, 38, 12, 0`
- `internal/oracle/testdata/user_iterators.a`: `669, 669, 466, 915, 67, 0`
- `internal/oracle/testdata/user_iterators_rest_tdz.a`: `5, 0, 5, 3, 5, 0`
- `internal/oracle/testdata/e4eec87_u02_optional_absent.a`: `6, 6, 3, 13, 4, 0`
- `internal/oracle/testdata/literal_optional_shapes.a`: `36, 36, 40, 68, 9, 0`
- `internal/oracle/testdata/e4eec87_u03_discriminated_undefined.a`: `10, 10, 14, 20, 6, 0`
- `internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a`: `4, 4, 3, 10, 3, 0`
- `internal/oracle/testdata/e4eec87_f1_field_narrowed.a`: `1, 0, 2, 3, 1, 0`
- `internal/oracle/testdata/e4eec87_f1_class_narrowed.a`: `1, 0, 1, 1, 1, 0`
- `internal/oracle/testdata/e4eec87_f1_alias_narrowed.a`: `1, 0, 3, 3, 1, 0`
- `internal/oracle/testdata/e4eec87_f1_field_present.a`: `6, 6, 2, 9, 3, 0`
- `internal/fresh/testdata/regexp_tree.ts`: `97, 97, 69, 85, 35, 0`
- `internal/oracle/testdata/regexp_cycle_fields.a`: `61, 61, 134, 132, 44, 0`
- `internal/oracle/testdata/regexp_cycle_collections.a`: `31, 31, 86, 90, 23, 0`
- `internal/oracle/testdata/regexp_cycle_closures.a`: `11, 11, 25, 26, 10, 0`
- `internal/oracle/testdata/regexp_cycle_weak.a`: `4, 4, 12, 14, 4, 0`

## Numeric enum ruling

Whole numeric enums accept numbers, arithmetic, shifts, signed bit masks and number assertions. String enums retain identity and closure. Reverse lookup stays string | undefined. Numeric never reads and exhaustive-switch unmatched paths panic with exit 70, the value and enum name in both backends. Only actual immutable member origins permit erasure. Switch values are evaluated once.

Nine new oracle sources cover normal completion and seven pinned never paths. All original stage3 sources are unchanged. Matrix: before 7 Node matches, 0 checked stops, 4 refusals, 2 checker failures; after 9 matches, 1 checked stop, 1 refusal, 2 checker failures. Fixtures 05/09 now agree byte for byte, including signed SymbolFlags. Fixture 07 compiles but stops on index 1 of an empty array, which existing array safety forbids. Fixture 06 cannot prove the generic Mutable<T> field's write contract. 08/11 retain TS2532 plus unchecked array/any contracts. Name enumeration support and its original mutant remain in place.

An extra probe found an actual payload miscompile when an open numeric enum was treated as a fixed object tag: Node text1 versus native 1. Member-specific literal assignments and object union refinements with unproven open enum tags therefore retain explicit soundness refusals. Whole singleton enum values stay open. Named lower tests cover those refusals and genuine member-tagged unions.

Ruling counts add twelve rows, with no changes to previous numerical rows: nine enums_open sources, stage3 fixtures 05/09, and checked-stop fixture 07.

Mutant details are in mutants.json and the replay script mutants.py. Every mutant is restored after its test, and build failures do not count as catches. The first enum-object mutant rerun was masked by the new independent literal check; its probe now assigns the correct member constant, isolating the runtime-object mutation rule. The final replay must catch all eleven.

The first final full gate found a nullish-coalescing regression in enums_flags.a.
A focused lower test reproduced it; the whole-enum comparison now checks defined
members while preserving the existing separate absence rules. An additional
array never probe was NotYet because indexed reads have no field symbol. The
small stored-element-type helper and pinned index fixture cover that path.
These fixes are included in the final validation below.

## Final commands and outputs

All commands use `source /workspace/adamic-tools/env.sh`. Test output goes to the linked log files.

| Command | Result |
| --- | --- |
| go test ./internal/lower ./internal/load -count=1 | PASS 13.017s / 1.088s |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout=30m | PASS 129.919s |
| go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts | PASS 28.850s |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=30m | PASS 34.483s |
| go vet ./... | PASS, empty output |
| gofmt -l internal cmd | Empty output |
| git diff --check | Empty output |
| Fresh CLI build and thirteen original stage3 source comparisons | 9 Node matches, 1 pinned stop, 1 Refused, 2 Checker |

The full oracle includes the enum bucket, both backends, release and sanitized native, ASan/UBSan, leak checks, counts and the permanent enum mutants. The additional source mutants are listed below. Ruling counts rows and tuples are recorded exactly in ruling-counts.json; no prior numerical row changes.

| Mutant | Named test | Caught |
| --- | --- | --- |
| remove-never-check | ^TestNumericEnumNeverPinned$ | True |
| erase-without-proof | ^TestNumericEnumNeverPinned$ | True |
| close-numeric-domain | ^TestNumericEnumsAreOpen/arithmetic$ | True |
| invert-object-presence | TestNativeAgreesWithNode/stage3/fixtures/enums/05 | True |
| permit-enum-object-write | ^TestEnumSlotViews/object_assign$ | True |
| repeat-switch-effects | ^TestNumericEnumNeverPathsPinned/implicit$ | True |
| omit-never-increment-check | ^TestNumericEnumNeverPathsPinned/update$ | True |
| permit-member-number | ^TestNumericEnumLiteralPromises/member_tag$ | True |
| trust-exclusion-literal | ^TestNumericEnumLiteralPromises/narrowed_member$ | True |
| trust-open-object-tag | ^TestNumericEnumLiteralPromises/open_tag$ | True |
| omit-array-enum-storage | ^TestNumericEnumNeverPathsPinned/index$ | True |

Proof erasure is intentionally limited to immutable aliases of actual members.
No erasure from mutable fields, array slots or general control-flow equality
proofs is attempted. Dynamic property and mixed-enum union never paths were not
specifically exercised in this unit. The original source fixtures and status
files are unchanged; open-results.json records the new observations.

Final logs: [lower/load](enums-open-packages-final.log),
[uncached full oracle](enums-open-oracle-final.log),
[counts](enums-open-counts-final.log),
[mutant replay](enums-open-mutants.log),
[stage3 matrix](enums-open-matrix.log).
Individual mutant logs use the basenames recorded in mutants.json.

The compiler implementation commit is 287a5ec543b10b51d7944b3855815f0fa8d96445. A follow-up trims three trailing spaces from the matrix summary log; source comparisons and JSON observations are unchanged.

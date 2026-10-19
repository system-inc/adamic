Re-cut checked writes onto newest fetched main, preserving slice 2 Error semantics and checked-write shape metadata.
Source 2dad6ae36c3ba6aeaeefc636e05f4b2956a6c4f6; base 561aeb4b851ca0820f39440a09dc5fb3fd27ba40; delivery SHA is in the final response.
Counts regenerated once on Linux clang 20.1.8; full package, guard, vet, formatting and a-check results follow below.
Mutants: all 19 original overlays are rebased onto current source; five exceptions.c mutants and a new Error-shape preservation mutant are exercised.
Not covered: full repository gate, full-tree a-check, ahra/Loom submission, other platforms, or the opt-in full WASI matrix.

Source and base

Fetch origin/main and the named source branch; fetched main is exactly the requested minimum 561aeb4b851ca0820f39440a09dc5fb3fd27ba40. Record git merge-tree --write-tree --merge-base=2dad6ae36c3b~1 origin/main 2dad6ae36c3b, then apply only that commit's parent-to-tip change with cherry-pick --no-commit. The source's house-check result is supplied by the user, not rerun or claimed here. The previous source message's WIP statements describe its historical worker run; this report records this re-cut's independent results. No older main-to-source diff is applied.

Conflicts and runtime clearance

Only exceptions.c and counts.md conflict. Main's slice 2 adds absent_message_shape with private #message storage, an immortal empty fallback, message-undefined constructor selection, string representation tags and adamic_error_field_name's public lookup alias. All remain byte-for-byte from main. Checked writes extends adamic_shape with the contracts pointer, so both runtime-owned Error shapes must supply a fifth NULL initializer. Their compiler allocation contracts are absent; a NULL pointer is the conservative existing-runtime convention. This is the sole resolution assumption. No main behavior is reverted. The existing Node/ASan/UBSan own-property and checked-read witness passes for new Error(), new Error(''), new Error('m') and new Error(undefined).

Exact signed exceptions.c lines against main (complete behavioral-source delta for that file):

```diff
-static const adamic_shape error_shape = {2, error_names, error_references, NULL};
+static const adamic_shape error_shape = {2, error_names, error_references, NULL, NULL};
-static const adamic_shape absent_message_shape = {2, absent_message_names, error_references, NULL};
+static const adamic_shape absent_message_shape = {2, absent_message_names, error_references, NULL, NULL};
```

The error_shape pair is exactly the original commit's own runtime pair. The absent_message_shape pair is the only additional pair anywhere in internal/native/runtime/*.c or *.h beyond the original commit's own signed runtime lines. No original runtime signed line is missing. runtime-comparison.json, source-runtime.patch, recut-runtime.patch and exceptions-against-main.patch retain the exact clearance evidence. No cohere source is copied.

Counts

Take main's entire table; regenerate exactly once after compilation with timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -json -timeout=570s -args -update-counts. PASS, 90.782s package. No hand-merged row. Table grows from 1,110 to 1,230 rows: 95 existing rows change, 120 are added and none are removed. Of 215 changed rows, 212 match the original source's exact counter delta or new value.

Three additional rows are fixtures newly on main, each an effect of this change. An emission baseline using main's Go compiler sources is built with an overlay restoring every changed production Go file to origin/main and deleting source-added Go files only in the overlay (32 files; the checkout is untouched). Its emitted C is compared with the real branch compiler on each unchanged fixture. The checked-write ArrayPush emitter retains a borrowed receiver before evaluating the argument and releases that hold after the statement. callback_widening/number-generic.a and object-generic.a each push three elements: exactly +3 retains/+3 releases. miscompile_2a/controls.a has two collection pushes: exactly +2/+2. The other C differences are temporary renumbering and five-field shape declarations; allocations/frees, peak live and regions are unchanged. main-overlay.json, additional-count-row-evidence.json, both emitted C files per fixture and their emission patches preserve the causal evidence. These fixtures also run against Node and sanitizers in the full oracle suite.

Every changed counts row follows. Counter order: allocations / frees / retains / releases / peak live / in regions.

| Fixture | Main A/F/R/L/P/G | Re-cut A/F/R/L/P/G | Delta |
|---|---|---|---|
| bench/parallel_files.a | 3952651/3952651/8151045/8654930/8839/0 | 3952651/3952651/11034629/11538514/8839/0 | 0/0/2883584/2883584/0/0 |
| internal/load/testdata/0.1/compile/06_stack.ts | 39/39/50/76/7/0 | 39/39/60/86/7/0 | 0/0/10/10/0/0 |
| internal/load/testdata/0.1/compile/09_tree.ts | 19/19/28/69/16/0 | 19/19/37/78/16/0 | 0/0/9/9/0/0 |
| internal/oracle/testdata/arguments.a | 77/77/14/68/27/0 | 77/77/70/124/27/0 | 0/0/56/56/0/0 |
| internal/oracle/testdata/array_literal_empty_first.a | 25/25/17/39/11/0 | 25/25/21/43/11/0 | 0/0/4/4/0/0 |
| internal/oracle/testdata/array_literal_empty_last.a | 28/28/21/43/14/0 | 28/28/25/47/14/0 | 0/0/4/4/0/0 |
| internal/oracle/testdata/array_literal_empty_typed.a | 17/17/13/31/10/0 | 17/17/14/32/10/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/bitwise.a | 9304/9304/1369/9342/45/0 | 9304/9304/2665/10638/45/0 | 0/0/1296/1296/0/0 |
| internal/oracle/testdata/borrow_chain_coverage_return.a | 16/16/7/13/5/0 | 16/16/8/14/5/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/borrow_element.a | 113/113/67/128/15/0 | 113/113/90/151/15/0 | 0/0/23/23/0/0 |
| internal/oracle/testdata/borrow_element_throw.a | 38/38/30/50/8/0 | 38/38/35/55/8/0 | 0/0/5/5/0/0 |
| internal/oracle/testdata/borrow_loop.a | 128/126/110/181/14/2 | 128/126/134/205/14/2 | 0/0/24/24/0/0 |
| internal/oracle/testdata/borrow_loop_calls.a | 263/261/176/342/25/2 | 263/261/216/382/25/2 | 0/0/40/40/0/0 |
| internal/oracle/testdata/callback_widening/number-generic.a | 13/13/8/18/8/0 | 13/13/11/21/8/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/callback_widening/object-generic.a | 13/13/20/27/10/0 | 13/13/23/30/10/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/case_mapping.a | 203/203/19/216/16/0 | 203/203/271/468/16/0 | 0/0/252/252/0/0 |
| internal/oracle/testdata/census_append_overload.a | 10/10/3/16/3/0 | 10/10/4/17/3/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/census_small_rest.a | 20/20/35/42/8/0 | 20/20/36/43/8/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/checked-runtime-fields_refused/directory.a (TypeScript adapter) | new | 5/5/7/10/5/0 | new |
| internal/oracle/testdata/checked-runtime-fields_refused/fs.a (TypeScript adapter) | new | 4/4/7/9/4/0 | new |
| internal/oracle/testdata/checked-runtime-fields_refused/map.a (TypeScript adapter) | new | 9/9/19/21/9/0 | new |
| internal/oracle/testdata/checked-runtime-fields_refused/regex.a (TypeScript adapter) | new | 15/15/23/24/13/0 | new |
| internal/oracle/testdata/checked-runtime-fields_refused/set.a (TypeScript adapter) | new | 10/10/19/22/9/0 | new |
| internal/oracle/testdata/checked-runtime-fields_refused/status.a (TypeScript adapter) | new | 4/4/6/10/4/0 | new |
| internal/oracle/testdata/checked-write-messages_refused/02_shared-empty_out.a (TypeScript adapter) | new | 2/0/1/1/2/0 | new |
| internal/oracle/testdata/checked-write-messages_refused/03_flow-node_out.a (TypeScript adapter) | new | 3/0/4/3/3/0 | new |
| internal/oracle/testdata/checked-write-messages_refused/13_declaration-array_out.a (TypeScript adapter) | new | 3/0/3/3/3/0 | new |
| internal/oracle/testdata/checked-write-messages_refused/15_detached-diagnostic_out.a (TypeScript adapter) | new | 2/0/2/2/2/0 | new |
| internal/oracle/testdata/checked-write-messages_refused/16_flow-assignment-union_out.a (TypeScript adapter) | new | 3/0/4/3/3/0 | new |
| internal/oracle/testdata/class_as_interface.a | 372/372/303/468/60/0 | 372/372/324/489/60/0 | 0/0/21/21/0/0 |
| internal/oracle/testdata/closures_throw.a | 418/418/1566/1801/142/0 | 418/418/1572/1807/142/0 | 0/0/6/6/0/0 |
| internal/oracle/testdata/closures_throw_uncaught.a | 31/29/20/36/13/0 | 31/29/28/44/13/0 | 0/0/8/8/0/0 |
| internal/oracle/testdata/collections.a | 362/362/392/550/103/0 | 362/362/415/573/103/0 | 0/0/23/23/0/0 |
| internal/oracle/testdata/doubly_linked.a | 90/90/525/558/49/0 | 90/90/558/591/49/0 | 0/0/33/33/0/0 |
| internal/oracle/testdata/exceptions.a | 133/133/163/229/20/0 | 133/133/168/234/20/0 | 0/0/5/5/0/0 |
| internal/oracle/testdata/finally_leaves.a | 157/157/59/174/10/0 | 157/157/67/182/10/0 | 0/0/8/8/0/0 |
| internal/oracle/testdata/fresh_calls.a | 113/113/216/273/37/0 | 113/113/258/315/37/0 | 0/0/42/42/0/0 |
| internal/oracle/testdata/fresh_parser.a | 145/145/280/450/34/0 | 145/145/316/486/34/0 | 0/0/36/36/0/0 |
| internal/oracle/testdata/fresh_writes.a | 330/330/369/485/132/0 | 330/330/386/502/132/0 | 0/0/17/17/0/0 |
| internal/oracle/testdata/from_codes.a | 408/408/56/413/12/0 | 408/408/106/463/12/0 | 0/0/50/50/0/0 |
| internal/oracle/testdata/gaps.a | 97/97/85/154/47/0 | 97/97/87/156/47/0 | 0/0/2/2/0/0 |
| internal/oracle/testdata/generic_functions.a | 35/35/32/55/8/0 | 35/35/40/63/8/0 | 0/0/8/8/0/0 |
| internal/oracle/testdata/host_map_union.a | 61/61/124/163/12/0 | 61/61/127/166/12/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/library_array_copy.a | 77/77/55/117/17/0 | 77/77/56/118/17/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/library_array_find_last.a | 43/43/26/75/10/0 | 43/43/27/76/10/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/library_array_flat.a | 33/33/41/57/19/0 | 33/33/60/76/19/0 | 0/0/19/19/0/0 |
| internal/oracle/testdata/library_array_flat_map.a | 50/50/60/109/17/0 | 50/50/79/128/17/0 | 0/0/19/19/0/0 |
| internal/oracle/testdata/library_array_iterators.a | 57/57/30/81/13/0 | 57/57/31/82/13/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/library_array_spliced.a | 72/72/40/114/12/0 | 72/72/60/134/12/0 | 0/0/20/20/0/0 |
| internal/oracle/testdata/library_array_with.a | 68/68/43/109/10/0 | 68/68/44/110/10/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/library_fnexpr_recurse.a | 29/29/69/99/11/0 | 29/29/74/104/11/0 | 0/0/5/5/0/0 |
| internal/oracle/testdata/library_map_set_construct.a | 180/180/225/271/52/0 | 180/180/236/282/52/0 | 0/0/11/11/0/0 |
| internal/oracle/testdata/library_map_set_group_by.a | 115/115/117/201/36/0 | 115/115/137/221/36/0 | 0/0/20/20/0/0 |
| internal/oracle/testdata/library_map_set_groupby_keys.a | 31/31/42/68/12/0 | 31/31/53/79/12/0 | 0/0/11/11/0/0 |
| internal/oracle/testdata/library_map_set_iterator_exhausted.a | 20230/20230/30328/50539/18/0 | 20230/20230/30330/50541/18/0 | 0/0/2/2/0/0 |
| internal/oracle/testdata/library_map_set_keys.a | 76/76/64/129/15/0 | 76/76/68/133/15/0 | 0/0/4/4/0/0 |
| internal/oracle/testdata/library_map_set_setops.a | 590/590/336/655/17/0 | 590/590/427/746/17/0 | 0/0/91/91/0/0 |
| internal/oracle/testdata/lint_runtime_release_chain.a | 405034/405034/505776/303790/300003/0 | 405034/405034/506032/304046/300003/0 | 0/0/256/256/0/0 |
| internal/oracle/testdata/lint_runtime_release_shared.a | 1008/1008/8478/6396/263/0 | 1008/1008/9249/7167/263/0 | 0/0/771/771/0/0 |
| internal/oracle/testdata/map_foreach_arrays.a | 995/995/932/1355/95/0 | 995/995/1028/1451/95/0 | 0/0/96/96/0/0 |
| internal/oracle/testdata/map_foreach_closures.a | 816/816/1044/1277/87/0 | 816/816/1140/1373/87/0 | 0/0/96/96/0/0 |
| internal/oracle/testdata/map_foreach_keep.a | 516/516/625/931/35/0 | 516/516/628/934/35/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/map_foreach_objects.a | 732/732/834/1076/63/0 | 732/732/930/1172/63/0 | 0/0/96/96/0/0 |
| internal/oracle/testdata/maybe_number_slots.a | 484/484/304/797/16/0 | 484/484/305/798/16/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/method_coverage_array_mutations.a | 18/18/47/58/10/0 | 18/18/49/60/10/0 | 0/0/2/2/0/0 |
| internal/oracle/testdata/miscompile_2a/controls.a | 33/32/34/56/10/1 | 33/32/36/58/10/1 | 0/0/2/2/0/0 |
| internal/oracle/testdata/nested_array.a | 17/17/17/27/9/0 | 17/17/21/31/9/0 | 0/0/4/4/0/0 |
| internal/oracle/testdata/normalize.a | 205/205/104/274/11/0 | 205/205/439/609/11/0 | 0/0/335/335/0/0 |
| internal/oracle/testdata/normalize_long_marks.a | 8693/8693/8521/8769/398/0 | 8693/8693/16898/17146/398/0 | 0/0/8377/8377/0/0 |
| internal/oracle/testdata/number_edges.a | 74/74/17/94/13/0 | 74/74/22/99/13/0 | 0/0/5/5/0/0 |
| internal/oracle/testdata/overload_array_to_multimap.a | 11/11/33/39/8/0 | 11/11/36/42/8/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/overload_sort_deduplicate.a | 10/10/39/38/3/0 | 10/10/42/41/3/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/read_files.a | 58/58/33/76/7/0 | 58/58/130/173/7/0 | 0/0/97/97/0/0 |
| internal/oracle/testdata/regexp_exec.a | 163/163/93/155/17/0 | 163/163/120/182/17/0 | 0/0/27/27/0/0 |
| internal/oracle/testdata/reuse_arrays.a | 78/78/35/94/11/0 | 78/78/45/104/11/0 | 0/0/10/10/0/0 |
| internal/oracle/testdata/route_targets_recursive.a | 19/18/6/23/7/1 | 19/18/7/24/7/1 | 0/0/1/1/0/0 |
| internal/oracle/testdata/route_targets_unknown.a | 36/36/18/49/7/0 | 36/36/21/52/7/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/search_from.a | 292/292/159/320/22/0 | 292/292/287/448/22/0 | 0/0/128/128/0/0 |
| internal/oracle/testdata/searches.a | 106/106/81/173/19/0 | 106/106/83/175/19/0 | 0/0/2/2/0/0 |
| internal/oracle/testdata/set_foreach_arrays.a | 516/516/524/713/58/0 | 516/516/564/753/58/0 | 0/0/40/40/0/0 |
| internal/oracle/testdata/set_foreach_closures.a | 411/411/634/705/39/0 | 411/411/674/745/39/0 | 0/0/40/40/0/0 |
| internal/oracle/testdata/set_foreach_objects.a | 368/368/515/586/30/0 | 368/368/555/626/30/0 | 0/0/40/40/0/0 |
| internal/oracle/testdata/size_class_churn.a | 675173/675173/270038/675212/8005/0 | 675173/675173/405038/810212/8005/0 | 0/0/135000/135000/0/0 |
| internal/oracle/testdata/sort_releases.a | 30/30/59/72/11/0 | 30/30/68/81/11/0 | 0/0/9/9/0/0 |
| internal/oracle/testdata/splice_empty.a | 14/14/6/23/10/0 | 14/14/10/27/10/0 | 0/0/4/4/0/0 |
| internal/oracle/testdata/splices.a | 74/74/29/96/12/0 | 74/74/44/111/12/0 | 0/0/15/15/0/0 |
| internal/oracle/testdata/string_append.a | 248/248/140/278/48/0 | 248/248/260/398/48/0 | 0/0/120/120/0/0 |
| internal/oracle/testdata/string_build_boundaries.a | 2229/2229/342/2347/11/0 | 2229/2229/6749/8754/11/0 | 0/0/6407/6407/0/0 |
| internal/oracle/testdata/string_build_calls.a | 47/47/4/52/7/0 | 47/47/420/468/7/0 | 0/0/416/416/0/0 |
| internal/oracle/testdata/string_build_join.a | 161/161/81/193/10/0 | 161/161/281/393/10/0 | 0/0/200/200/0/0 |
| internal/oracle/testdata/string_build_padding.a | 759/759/126/877/7/0 | 759/759/1154/1905/7/0 | 0/0/1028/1028/0/0 |
| internal/oracle/testdata/string_build_repeat.a | 240/240/32/264/8/0 | 240/240/196/428/8/0 | 0/0/164/164/0/0 |
| internal/oracle/testdata/string_positions.a | 327/327/368/606/97/0 | 327/327/396/634/97/0 | 0/0/28/28/0/0 |
| internal/oracle/testdata/string_views_characters.a | 281/281/1542/1825/8/0 | 281/281/2310/2593/8/0 | 0/0/768/768/0/0 |
| internal/oracle/testdata/string_views_holders.a | 45/45/72/79/13/0 | 45/45/75/82/13/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/string_views_loops.a | 159/159/45/167/21/0 | 159/159/63/185/21/0 | 0/0/18/18/0/0 |
| internal/oracle/testdata/string_views_surrogates.a | 188/188/83/255/13/0 | 188/188/98/270/13/0 | 0/0/15/15/0/0 |
| internal/oracle/testdata/string_views_throw.a | 55/55/62/76/14/0 | 55/55/65/79/14/0 | 0/0/3/3/0/0 |
| internal/oracle/testdata/throw_keeps_old_value_variants.a | 26/26/11/32/6/0 | 26/26/12/33/6/0 | 0/0/1/1/0/0 |
| internal/oracle/testdata/timsort.a | 2769/2769/3778/6231/308/0 | 2769/2769/48733/51186/308/0 | 0/0/44955/44955/0/0 |
| internal/oracle/testdata/undefined_keys.a | 122/122/196/261/24/0 | 122/122/208/273/24/0 | 0/0/12/12/0/0 |
| internal/oracle/testdata/user_iterators.a | 663/663/455/904/64/0 | 663/663/487/936/64/0 | 0/0/32/32/0/0 |
| internal/oracle/testdata/utf8_sweep.a | 38498/38498/7701/38501/7706/0 | 38498/38498/36235/67035/7706/0 | 0/0/28534/28534/0/0 |
| internal/oracle/testdata/utf8_view.a | 59/59/33/79/7/0 | 59/59/105/151/7/0 | 0/0/72/72/0/0 |
| internal/oracle/testdata/write_files.a | 132/132/52/167/8/0 | 132/132/184/299/8/0 | 0/0/132/132/0/0 |
| stage3/checked-writes/boolean-fit.ts | new | 1/1/1/3/1/0 | new |
| stage3/checked-writes/boolean-misfit.ts | new | 1/0/1/1/1/0 | new |
| stage3/checked-writes/branch-bottom-number-fit.ts | new | 4/4/5/11/3/0 | new |
| stage3/checked-writes/branch-bottom-number-misfit.ts | new | 3/0/3/3/3/0 | new |
| stage3/checked-writes/branch-bottom-string-fit.ts | new | 6/6/7/13/5/0 | new |
| stage3/checked-writes/branch-bottom-string-misfit.ts | new | 4/0/4/4/4/0 | new |
| stage3/checked-writes/branch-file-fit.ts | new | 4/4/6/11/4/0 | new |
| stage3/checked-writes/branch-file-misfit.ts | new | 4/0/4/4/4/0 | new |
| stage3/checked-writes/branch-length-fit.ts | new | 5/5/4/10/5/0 | new |
| stage3/checked-writes/branch-length-misfit.ts | new | 4/0/4/4/4/0 | new |
| stage3/checked-writes/branch-start-fit.ts | new | 5/5/4/10/5/0 | new |
| stage3/checked-writes/branch-start-misfit.ts | new | 4/0/4/4/4/0 | new |
| stage3/checked-writes/compound-fit.ts | new | 2/2/3/7/2/0 | new |
| stage3/checked-writes/compound-misfit.ts | new | 2/0/3/2/2/0 | new |
| stage3/checked-writes/container-alias-fit.ts | new | 3/3/2/5/3/0 | new |
| stage3/checked-writes/container-alias-misfit.ts | new | 2/0/2/2/2/0 | new |
| stage3/checked-writes/container-boolean-fit.ts | new | 1/1/1/3/1/0 | new |
| stage3/checked-writes/container-boolean-misfit.ts | new | 1/0/1/1/1/0 | new |
| stage3/checked-writes/container-fill-fit.ts | new | 2/2/1/4/2/0 | new |
| stage3/checked-writes/container-fill-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/container-map-fit.ts | new | 2/2/2/4/2/0 | new |
| stage3/checked-writes/container-map-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/container-map-object-fit.ts | new | 6/6/3/8/6/0 | new |
| stage3/checked-writes/container-map-object-misfit.ts | new | 5/0/1/3/5/0 | new |
| stage3/checked-writes/container-map-values-fit.ts | new | 10/10/11/17/8/0 | new |
| stage3/checked-writes/container-map-values-misfit.ts | new | 10/7/11/13/8/0 | new |
| stage3/checked-writes/container-nested-fit.ts | new | 4/4/2/8/4/0 | new |
| stage3/checked-writes/container-nested-misfit.ts | new | 3/0/1/3/3/0 | new |
| stage3/checked-writes/container-number-fit.ts | new | 2/2/1/4/2/0 | new |
| stage3/checked-writes/container-number-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/container-object-fit.ts | new | 6/6/2/8/6/0 | new |
| stage3/checked-writes/container-object-misfit.ts | new | 5/0/1/3/5/0 | new |
| stage3/checked-writes/container-slice-fit.ts | new | 3/3/2/7/3/0 | new |
| stage3/checked-writes/container-slice-misfit.ts | new | 3/0/2/3/3/0 | new |
| stage3/checked-writes/container-splice-fit.ts | new | 2/2/1/4/2/0 | new |
| stage3/checked-writes/container-splice-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/container-string-fit.ts | new | 1/1/4/4/1/0 | new |
| stage3/checked-writes/container-string-misfit.ts | new | 1/0/2/1/1/0 | new |
| stage3/checked-writes/diagnostic-alias-fit.ts | new | 3/3/5/10/3/0 | new |
| stage3/checked-writes/diagnostic-alias-misfit.ts | new | 3/0/5/5/3/0 | new |
| stage3/checked-writes/diagnostic-fit.ts | new | 3/3/4/7/3/0 | new |
| stage3/checked-writes/diagnostic-inner-fit.ts | new | 4/4/2/6/4/0 | new |
| stage3/checked-writes/diagnostic-inner-misfit.ts | new | 4/0/2/2/4/0 | new |
| stage3/checked-writes/diagnostic-misfit.ts | new | 3/0/2/2/3/0 | new |
| stage3/checked-writes/diagnostic-nested-fit.ts | new | 10/10/4/13/10/0 | new |
| stage3/checked-writes/diagnostic-nested-misfit.ts | new | 7/0/2/3/7/0 | new |
| stage3/checked-writes/diagnostic-proof-fit.ts | new | 12/12/6/15/12/0 | new |
| stage3/checked-writes/diagnostic-proof-misfit.ts | new | 9/0/2/3/9/0 | new |
| stage3/checked-writes/diagnostic-related-fit.ts | new | 5/5/2/6/5/0 | new |
| stage3/checked-writes/diagnostic-related-misfit.ts | new | 4/0/2/2/4/0 | new |
| stage3/checked-writes/diagnostic-rich-fit.ts | new | 27/27/11/32/23/0 | new |
| stage3/checked-writes/diagnostic-rich-misfit.ts | new | 6/0/2/2/6/0 | new |
| stage3/checked-writes/emit-auto-fit.ts | new | 6/6/4/11/6/0 | new |
| stage3/checked-writes/emit-auto-misfit.ts | new | 6/0/2/4/6/0 | new |
| stage3/checked-writes/emit-callback-fit.ts | new | 7/7/7/16/7/0 | new |
| stage3/checked-writes/emit-callback-misfit.ts | new | 7/0/6/5/7/0 | new |
| stage3/checked-writes/emit-comment-fit.ts | new | 2/2/4/8/2/0 | new |
| stage3/checked-writes/emit-comment-misfit.ts | new | 3/0/4/2/3/0 | new |
| stage3/checked-writes/emit-node-fit.ts | new | 6/6/3/10/6/0 | new |
| stage3/checked-writes/emit-node-misfit.ts | new | 6/0/2/4/6/0 | new |
| stage3/checked-writes/emit-resolution-fit.ts | new | 2/2/6/10/2/0 | new |
| stage3/checked-writes/emit-resolution-misfit.ts | new | 2/0/5/4/2/0 | new |
| stage3/checked-writes/flags-fit.ts | new | 5/5/127/132/4/0 | new |
| stage3/checked-writes/flags-misfit.ts | new | 5/2/4/4/4/0 | new |
| stage3/checked-writes/flow-array-fit.ts | new | 1/1/1/3/1/0 | new |
| stage3/checked-writes/flow-array-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/flow-node-fit.ts | new | 4/4/6/9/4/0 | new |
| stage3/checked-writes/flow-node-misfit.ts | new | 4/0/4/3/4/0 | new |
| stage3/checked-writes/flow-undefined-fit.ts | new | 1/1/1/3/1/0 | new |
| stage3/checked-writes/flow-undefined-misfit.ts | new | 2/0/2/1/2/0 | new |
| stage3/checked-writes/generic-site-fit.ts | new | 2/2/2/5/2/0 | new |
| stage3/checked-writes/generic-site-misfit.ts | new | 2/0/2/1/2/0 | new |
| stage3/checked-writes/never-index-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/never-nullable-fit.ts | new | 5/5/4/10/3/0 | new |
| stage3/checked-writes/never-nullable-misfit.ts | new | 3/0/2/2/3/0 | new |
| stage3/checked-writes/never-number-fit.ts | new | 5/5/2/8/3/0 | new |
| stage3/checked-writes/never-number-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/never-object-fit.ts | new | 7/7/2/9/3/0 | new |
| stage3/checked-writes/never-object-misfit.ts | new | 3/0/1/1/3/0 | new |
| stage3/checked-writes/never-splice-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/never-string-fit.ts | new | 6/6/2/9/3/0 | new |
| stage3/checked-writes/never-string-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/never-wide-fit.ts | new | 6/6/2/11/4/0 | new |
| stage3/checked-writes/nullable-number-fit.ts | new | 2/2/2/5/2/0 | new |
| stage3/checked-writes/nullable-number-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/number-fit.ts | new | 6/6/2/10/6/0 | new |
| stage3/checked-writes/number-misfit.ts | new | 2/0/1/1/2/0 | new |
| stage3/checked-writes/optional-boolean-fit.ts | new | 1/1/4/6/1/0 | new |
| stage3/checked-writes/optional-boolean-misfit.ts | new | 1/0/1/1/1/0 | new |
| stage3/checked-writes/optional-false-fit.ts | new | 1/1/2/4/1/0 | new |
| stage3/checked-writes/optional-false-misfit.ts | new | 1/0/1/1/1/0 | new |
| stage3/checked-writes/overload-callback-fit.ts | new | 5/5/4/11/5/0 | new |
| stage3/checked-writes/overload-callback-misfit.ts | new | 5/0/3/4/5/0 | new |
| stage3/checked-writes/overload-result-fit.ts | new | 4/4/3/8/4/0 | new |
| stage3/checked-writes/overload-result-misfit.ts | new | 4/0/2/3/4/0 | new |
| stage3/checked-writes/parent-fit.ts | new | 4/4/8/11/4/0 | new |
| stage3/checked-writes/parent-misfit.ts | new | 4/0/4/2/4/0 | new |
| stage3/checked-writes/proven-containers.a | new | 6/6/3/10/6/0 | new |
| stage3/checked-writes/proven-emit.a | new | 7/7/3/11/6/0 | new |
| stage3/checked-writes/proven-generic.a | new | 2/2/1/4/2/0 | new |
| stage3/checked-writes/proven-number.a | new | 4/4/2/6/4/0 | new |
| stage3/checked-writes/range-fit.ts | new | 5/5/2/8/5/0 | new |
| stage3/checked-writes/range-misfit.ts | new | 2/0/2/2/2/0 | new |
| stage3/checked-writes/receiver-fit.ts | new | 6/6/4/13/6/0 | new |
| stage3/checked-writes/reference-structure.ts | new | 3/0/3/2/3/0 | new |
| stage3/checked-writes/required-boolean-fit.ts | new | 1/1/2/4/1/0 | new |
| stage3/checked-writes/required-boolean-misfit.ts | new | 1/0/1/1/1/0 | new |
| stage3/checked-writes/string-fit.ts | new | 1/1/4/5/1/0 | new |
| stage3/checked-writes/string-misfit.ts | new | 2/0/2/1/2/0 | new |
| stage3/fixtures/taste/14_relative_complement.a | 10/10/22/34/7/0 | 10/10/25/37/7/0 | 0/0/3/3/0/0 |

Toolchain and resources

GOPROXY=https://proxy.golang.org|direct, ADAMIC_GOCACHE_OFF=1, GOMAXPROCS=4; source /workspace/adamic-tools/env.sh in each Go shell and unset GOCACHEPROG. The external cache remains disabled from this worker's earlier environment workaround. bash cloud/setup.sh --wasi-sdk passes in 34.392s: Go build 34.223s, deferred tests 34.363s, warm cache 34.364s; setup.log has every timing line. Native clang is house 20.1.8, WASI SDK clang is 20.1.8-wasi-sdk, Go is 1.27.1 and Node is 24.19.0. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). The existing declared WASI and command-runtime products are prepared before consuming CLI tests, with ADAMIC_BUILD_CACHE_DIR selecting the existing placeless cache when CLI tests vary XDG_CACHE_HOME.

Full IR export guards increased disk usage. Link only completed immutable export archives with identical full SHA256, size, owner and mode; replace each duplicate atomically with a hard link to the canonical archive. No product is deleted, no path or byte changes, and no building directory is touched. export-deduplication.json records every verified pair and recovered capacity. The full IR package passes; no test is weakened or skipped for capacity.

A-check

Compile all 29 touched .a files anywhere in the source change, including review evidence, with the actual branch compiler. Six checked, one NotYet (SpreadElement), 22 refused, zero type errors. Every whole file is byte-identical to 2dad6ae36c3b: no new header, body or line-position edits. The unchanged pinned Gate.aCheck accepts all. Removed-header and wrong-reason mutants are caught by replaying captured actual diagnostics through the same predicate, with an in-memory header override instead of changing checkout files during parallel suites. Restored predicate passes. a-check-status.json includes per-file statuses and full reasons; a-check.json and header-mutants.json retain the checks. Predicate source and SHA256 are recorded in predicate-provenance.json.

Mutant methodology

Rebase all 19 recorded source overlays by applying only their original mutation hunks to current source, with unique context assertions. This preserves newer main code that a stale whole-file overlay would discard. Run each original named catcher with timeout 120 go test -overlay <overlay> <package> -run <selector> -count=1 -parallel=4 -v -timeout=90s; write one log per mutant. No production source is changed. A caught mutant must fail its intended test, compile cleanly and avoid timeout; compiler/warning kills are excluded. mutant-results.json lists all exact commands and outcomes.

The full native package and runtime-focused.log explicitly run TestOrdinaryToPrimitiveMutants and TestOrdinaryToPrimitiveProtocolGuards against exceptions.c's wrong-order, missing-TypeError, leaked-object-result, terminal-invalid-hint and preserve-pending-error mutants. Node output/traces catch wrong order and missing TypeError; LeakSanitizer catches the omitted rejected-result release; expected terminal panic catches an omitted hint check; preserved pending-error output catches running conversion with a pending exception. All five are observed caught.

A new current-source exceptions.c overlay swaps the constructor's two shape choices without creating a compile warning. TestErrorUndefinedMessageHasNoOwnProperty fails its comparison against Node, with sanitizer-clean runtime execution, proving the main behavior retained by the resolution can fail independently of compilation. swap-error-message-shapes-result.json, .patch, .c.txt and .log preserve the evidence. The full oracle suite also runs its inherited embedded checked-write mutants.

Environment prerequisite follow-up

The first complete native run finishes with 499 of 500 top-level roots passing; only TestLongUnicodePortabilityGCC fails because GCC 14.2.0 lacks sanitizer/allocator_interface.h while its libasan exports the required APIs. No leaf exceeds 60s (maximum 18.89s). Exposing the LLVM public headers through C_INCLUDE_PATH does not fix it because main intentionally strips that ambient variable in toolchain.Env. Do not weaken that policy. Instead place a thin gcc wrapper in the existing ADAMIC_TOOLS/bin, executing the real /usr/bin/gcc with only an explicit -isystem for a sibling directory exposing the house LLVM sanitizer headers. No compiler/test source changes and no compiler substitution: --version remains Debian GCC 14.2.0-19, and real GCC compiles and links the entire runtime and program. The targeted GCC portability test then passes in 8.36s with source Node, emitted JavaScript, ASan, UBSan and leak observations agreeing. gcc-wrapper.sh.txt records this worker-only toolchain configuration. The full native package is rerun with the prerequisite fixed. Clang for counts and normal native builds remains 20.1.8.

Mutant baseline correction

The first historical plain-map overlay also removed ParallelMap.Moved, an unrelated main field absent when that overlay was generated. Its build failure is retained as plain-checked-writes.initial.log/.go.txt/.json and is explicitly not a catch. Audit each overlay against its actual 5aa205ec generation baseline. Sixteen overlays have identical target baselines in 5aa205ec and 2dad6ae; the overload overlay additionally carried stale invariance-cache changes and the two map overlays carried the stale field removal. Rebase and rerun those three from 5aa205ec so only their intended mutations remain. All nineteen final mutants are caught by actual assertions, with zero compiler-error or timeout kills (mutant-audit.json). Both plain-map mutants fail TestEveryMapAProgramHoldsIsSorted and TestAProgramIsTheSameBytesEveryTimeGobWritesIt. The overload mutant fails TestCheckedSignatureEveryReceivingOverload while preserving main's target-query cache. Initial results are retained separately; no failed build is counted green.

Final standalone mutant catchers

| Mutant | Catcher selector | Package | Result |
|---|---|---|---|
| revert-native-reference-fields | ^TestCheckedRuntime | ./internal/oracle | caught, exit 1 |
| revert-js-reference-fields | ^TestCheckedRuntime | ./internal/oracle | caught, exit 1 |
| revert-array-field-schema | ^TestCheckedRuntime(Regex|Directory)Fields$ | ./internal/oracle | caught, exit 1 |
| drop-runtime-array-contract | TestRuntimeCheckedReferenceArrayKeepsContract | ./internal/native | caught, exit 1 |
| drop-string-heap-kind | TestRuntimeCheckedReferenceArrayKeepsContract | ./internal/native | caught, exit 1 |
| accept-runtime-array-literal | TestRuntimeCheckedReferenceArrayRejectsLiteral | ./internal/native | caught, exit 1 |
| accept-weaker-array-contract | TestRuntimeCheckedReferenceArrayRejectsWeakerContract | ./internal/native | caught, exit 1 |
| restore-number-view-native | ^TestCheckedWriteMessage02$ | ./internal/oracle | caught, exit 1 |
| restore-number-view-js | ^TestCheckedWriteMessage02$ | ./internal/oracle | caught, exit 1 |
| drop-js-runtime-array-contract | ^TestCheckedRuntimeReferenceArrayLaterWriter$ | ./internal/oracle | caught, exit 1 |
| drop-js-string-heap-kind | ^TestCheckedRuntimeReferenceArrayLaterWriter$ | ./internal/oracle | caught, exit 1 |
| prior-literal | ^TestRuntimeCheckedLibraryPushRejectsLiteral$ | ./internal/native | caught, exit 1 |
| drop-checked-write-metadata | ^TestObjectMetadataCheckedWritesIsPerEmission$ | ./internal/native | caught, exit 1 |
| skip-signature-instantiation | ^TestCheckedSignatureInstantiation$ | ./internal/lower | caught, exit 1 |
| disable-mapped-identity | ^TestCheckedSignatureInstantiation$ | ./internal/lower | caught, exit 1 |
| accept-optional-mapped-identity | ^TestCheckedSignatureInstantiation$ | ./internal/lower | caught, exit 1 |
| skip-receiving-overloads | ^TestCheckedSignatureEveryReceivingOverload$ | ./internal/lower | caught, exit 1 |
| plain-checked-writes | ^Test(EveryMapAProgramHoldsIsSorted|AProgramIsTheSameBytesEveryTimeGobWritesIt)$ | ./internal/ir | caught, exit 1 |
| plain-contract-type-names | ^Test(EveryMapAProgramHoldsIsSorted|AProgramIsTheSameBytesEveryTimeGobWritesIt)$ | ./internal/ir | caught, exit 1 |

The source's prior lowering-resolution mutants are also re-cut onto current invariance.go without changing the checkout. uncached-and fails TestImpliedTargetAndQueriesOnce with "queried 0 logical expressions, want at least 255" (0.81s recorded test), proving the query bypass is observed. widen-and-left fails TestCheckedLogicalAndDoesNotWidenTruthTestedLeft with the readonly widening refusal. The initial extra evidence-runner diagnostic predicate expected a different query-count wording and rejected the already failing test; it is corrected to the actual diagnostic, and that captured execution is audited rather than rerun again. Both semantic mutants compile and fail the intended assertion; prior-resolution-mutants.json and their overlays retain exact commands and observations.

No new Adamic witness is written in this re-cut. Historical .ts fixtures and all existing source evidence are carried verbatim as part of the explicitly requested original change. The only production edit beyond that change plus newer-main integration is the extra absent-message shape initializer required by the ABI. Every changed row is this change's measured effect.

Successful setup timing lines:

```text
setup: node ready (0.021s)
setup: go ready (0.022s)
setup: submodules ready (0.069s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.010s
setup: markdown dependencies ready (0.075s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.185s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk, linking with /workspace/adamic-tools/bin/clang) (0.300s)
setup: shared cache off (ADAMIC_GOCACHE_OFF=1) (0.302s)
setup: go build ready (34.223s)
setup: test binaries deferred (use --warm-tests) (34.363s)
setup: build cache warm (34.364s)
setup: build-flags commit=561aeb4b851ca0820f39440a09dc5fb3fd27ba40 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=off load-before=0.09 0.03 0.14 1/841 411573 load-after=3.89 0.99 0.46 1/844 412227
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (34.392s)
```

Final verification

Each command sources the printed toolchain environment, sets GOMAXPROCS=4 and -parallel=4, and writes all test output directly to its named log. Complete direct packages are run, not a selected-test substitute. Native initial GCC failures and the corrected full rerun are retained separately. All individual completed final leaves stay below 60s; no tests are deleted or skipped to make this change pass.

| Package | Selection | Package seconds | Slowest leaf seconds | Result |
|---|---|---:|---:|---|
| internal/ir | all tests | 45.804 | 20.87 | PASS |
| internal/lower | all tests | 75.084 | 12.94 | PASS |
| internal/javascript | all tests | 0.165 | 0.06 | PASS |
| internal/native | all tests | 197.849 | 18.33 | PASS |
| internal/flow | all tests | 67.095 | 16.16 | PASS |
| internal/fresh | all tests | 13.316 | 2.1 | PASS |
| internal/regexp | all tests | 9.105 | 3.57 | PASS |
| cmd/adamic | ^TestProduct_CommandRuntime | 2.065 | 2.06 | PASS |
| cmd/adamic | all tests | 13.517 | 3.51 | PASS |
| stage1/cohere/values/oracles | ^TestProduct_ValuesGoOracle$ | 0.558 | 0.55 | PASS |
| stage1/cohere/values | all tests | 19.111 | 4.25 | PASS |
| internal/oracle | all tests | 536.069 | 32.73 | PASS |

final-package-results.json retains every exact command and outcome; all-test-seconds.json retains all terminal test events; changed-test-seconds.json gives every new/changed declaration its measured seconds. Complete oracle includes the inherited embedded checked-write mutants and holds admitted fixtures to source Node, backend Node and release/sanitized native outcomes.

Vet on all ten changed Go packages passes (timeout 180 go vet with each package explicitly named). Gofmt -l over every changed Go file prints nothing; git diff --check origin/main -- cmd internal passes. Mandatory guard command (timeout 120 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run '^(TestCallTargetReaders.*|TestEveryMapAProgramHoldsIsSorted|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestEveryTestIsParallelOrSaysWhy)$' -count=1 -parallel=4 -json -timeout=90s) passes. No new compiler-consuming package or runtime feature definition is introduced. No product/test gains a build directory or machine path; the worker-only GCC SDK configuration is evidence, not a repository product change.

All 19 standalone mutation logs are independently audited: assertions catch every mutant, with no compiler-error or timeout kills counted. The first stale map overlay failure is excluded. Five exceptions.c mutants, the Error-shape swap and both unchanged header-predicate mutants are also caught.

Delivery is one commit with parent 561aeb4b851ca0820f39440a09dc5fb3fd27ba40, the original message retained plus the requested re-cut line and explicit conservative assumption. Push only compiler/loom-r-checked-writes-ac-m2, without force. No ahra, PR or Loom submission. Final response supplies full SHA and remote verification.

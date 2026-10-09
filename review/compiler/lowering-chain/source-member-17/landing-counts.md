Counts regenerated on the step 21 landing compiler against main 7a10c877.
16 rows added, 91 existing rows changed, no rows removed.
A/F/R/L/P/G mean allocations, frees, retains, releases, peak live and values in regions.

| Fixture | Main | Landing | Delta | Cause |
|---|---|---|---|---|
| `internal/oracle/testdata/array_from_length.a` | 5/3/3/7/3/0 | 6/6/6/12/3/0 | 1/3/3/5/0/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/borrow_chain_coverage_finally.a` | 13/13/9/13/6/0 | 13/13/12/16/6/0 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/borrow_element_throw.a` | 38/38/30/50/8/0 | 38/38/36/56/8/0 | 0/0/6/6/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/borrow_loop.a` | 128/126/110/181/14/2 | 128/126/116/187/14/2 | 0/0/6/6/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/borrow_loop_calls.a` | 263/261/176/342/25/2 | 263/261/179/345/25/2 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/class_as_interface.a` | 372/372/303/468/60/0 | 372/372/304/469/60/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/class_features_accessors.a` | 67/67/47/104/21/0 | 67/67/49/106/21/0 | 0/0/2/2/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/class_features_static.a` | 110/110/55/162/28/0 | 110/110/58/165/28/0 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/class_features_static_private.a` | 42/42/53/94/12/0 | 42/42/73/114/12/0 | 0/0/20/20/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/class_inheritance_conditional.a` | 164/164/146/249/35/0 | 164/164/157/260/35/0 | 0/0/11/11/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/class_inheritance_exceptions.a` | 29/29/19/39/8/0 | 29/29/23/43/8/0 | 0/0/4/4/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/class_instance_key_throw.a` | 6/5/5/10/5/1 | 6/5/7/12/5/1 | 0/0/2/2/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/closures_throw.a` | 418/418/1566/1801/142/0 | 418/418/1585/1820/142/0 | 0/0/19/19/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/closures_throw_uncaught.a` | 31/29/20/36/13/0 | 31/31/20/37/13/0 | 0/2/0/1/0/0 | Owned tagged payload and cleanup on the catch or uncaught exit-1 path. |
| `internal/oracle/testdata/debugger_fail.a` | 2/2/4/4/2/0 | 2/2/5/5/2/0 | 0/0/1/1/0/0 | Unknown catch storage and checked nominal Error narrowing add one balanced pair. |
| `internal/oracle/testdata/devirtualize.a` | 43/39/22/62/10/4 | 43/39/24/64/10/4 | 0/0/2/2/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/entries_record_alias.a` | 65/65/89/118/28/0 | 65/65/90/119/28/0 | 0/0/1/1/0/0 | The dynamic repeat call now uses the checked library wrapper, adding one balanced pair. |
| `internal/oracle/testdata/exceptions.a` | 133/133/163/229/20/0 | 133/133/186/252/20/0 | 0/0/23/23/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/exceptions_empty.a` | 2/1/6/4/2/0 | 2/2/8/8/2/0 | 0/1/2/4/0/0 | Owned tagged payload and cleanup on the catch or uncaught exit-1 path. |
| `internal/oracle/testdata/exceptions_uncaught.a` | 3/1/3/5/2/0 | 3/3/3/6/2/0 | 0/2/0/1/0/0 | Owned tagged payload and cleanup on the catch or uncaught exit-1 path. |
| `internal/oracle/testdata/fallthrough_exceptions.a` | 26/26/10/29/5/0 | 26/26/12/31/5/0 | 0/0/2/2/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/fill_length.a` | 2/2/0/2/2/0 | 3/3/3/5/2/0 | 1/1/3/3/0/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/finally_leaves.a` | 157/157/59/174/10/0 | 157/157/64/179/10/0 | 0/0/5/5/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/fresh_writes.a` | 330/330/369/485/132/0 | 330/330/370/486/132/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/generic_method_return.a` | 23/15/28/36/5/6 | 23/17/30/39/5/6 | 0/2/2/3/0/0 | Owned tagged payload and cleanup on the catch or uncaught exit-1 path. |
| `internal/oracle/testdata/library_array_flat_map.a` | 50/50/60/109/17/0 | 50/50/61/110/17/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/library_array_spliced.a` | 72/72/40/114/12/0 | 72/72/41/115/12/0 | 0/0/1/1/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/library_array_with.a` | 68/68/43/109/10/0 | 68/68/64/130/10/0 | 0/0/21/21/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/library_function_expressions.a` | 56/56/49/96/19/0 | 56/56/50/97/19/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/library_map_set_group_by.a` | 115/115/117/201/36/0 | 115/115/118/202/36/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/library_object_freeze.a` | 16/16/10/27/5/0 | 16/16/10/28/5/0 | 0/0/0/1/0/0 | Validation wrapper releases its argument/result temporaries, including absent results; allocations unchanged. |
| `internal/oracle/testdata/library_object_freeze_alias.a` | 4/1/7/8/3/0 | 5/5/9/14/4/0 | 1/4/2/6/1/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/library_object_freeze_assign.a` | 4/2/6/8/2/0 | 10/10/10/18/8/0 | 6/8/4/10/6/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/library_object_freeze_write.a` | 2/1/5/6/2/0 | 3/3/7/11/2/0 | 1/2/2/5/0/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/long_literals.a` | 6117/6117/85/14610/8/0 | 6117/6117/86/14611/8/0 | 0/0/1/1/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_arrays.a` | 995/995/932/1355/95/0 | 995/995/933/1356/95/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_closures.a` | 816/816/1044/1277/87/0 | 816/816/1045/1278/87/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_fnexpr.a` | 44/44/64/100/15/0 | 44/44/65/101/15/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_keep.a` | 516/516/625/931/35/0 | 516/516/626/932/35/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_keys.a` | 391/391/363/674/25/0 | 391/391/364/675/25/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_named_closures.a` | 117/117/179/219/29/0 | 117/117/180/220/29/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_named_more.a` | 395/395/569/819/36/0 | 395/395/570/820/36/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_named_numbers.a` | 117/117/73/177/12/0 | 117/117/74/178/12/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_named_objects.a` | 165/165/207/278/31/0 | 165/165/208/279/31/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_numbers.a` | 373/373/155/478/14/0 | 373/373/156/479/14/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/map_foreach_objects.a` | 732/732/834/1076/63/0 | 732/732/835/1077/63/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/method_coverage_format_failure.a` | 1/0/0/1/1/0 | 2/2/4/6/2/0 | 1/2/4/5/1/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/method_coverage_string_repeat.a` | 2/2/3/6/2/0 | 2/2/5/8/2/0 | 0/0/2/2/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/move_throw.a` | 42/42/22/50/9/0 | 42/42/24/52/9/0 | 0/0/2/2/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/named_function_values.a` | 116/116/127/222/25/0 | 116/116/128/223/25/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/node_buffer_finalized.a` | 4/2/4/6/3/0 | 5/5/7/10/3/0 | 1/3/3/4/0/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/node_fs_file_close.a` | 27/27/13/30/7/0 | 27/27/16/33/7/0 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/node_fs_file_mkdir.a` | 61/61/24/81/10/0 | 61/61/29/86/10/0 | 0/0/5/5/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/node_fs_file_write_buffer.a` | 34/34/11/45/12/0 | 34/34/15/49/12/0 | 0/0/4/4/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/node_fs_file_write_file.a` | 31/31/11/36/9/0 | 31/31/13/38/9/0 | 0/0/2/2/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/normalize_form.a` | 3/2/0/3/3/0 | 4/4/3/7/3/0 | 1/2/3/4/0/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/number_formats.a` | 225/225/55/262/18/0 | 225/225/97/304/18/0 | 0/0/42/42/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/param_assigned_in_try.a` | 46/46/29/64/12/0 | 46/46/31/66/12/0 | 0/0/2/2/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/precision_range.a` | 4/3/1/4/2/0 | 5/5/4/9/2/0 | 1/2/3/5/0/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/radix_range.a` | 3/3/0/3/1/0 | 4/4/3/6/1/0 | 1/1/3/3/0/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/regexp.a` | 437/437/382/417/58/0 | 437/437/381/416/58/0 | 0/0/-1/-1/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/regexp_matchall_nonglobal.a` | 1/0/2/0/1/0 | 2/2/5/4/2/0 | 1/2/3/4/1/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/regexp_replace.a` | 625/625/206/336/113/0 | 625/625/205/335/113/0 | 0/0/-1/-1/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/regexp_replace/throw.a` | 13/13/11/14/11/0 | 13/13/12/15/11/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/regexp_replaceall_nonglobal.a` | 1/0/2/0/1/0 | 2/2/5/4/2/0 | 1/2/3/4/1/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/regions_throw.a` | 124/95/9/99/31/29 | 124/95/15/105/31/29 | 0/0/6/6/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/reuse_throw.a` | 26/26/24/41/8/0 | 26/26/27/44/8/0 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/route_targets_try_loop.a` | 68/66/41/85/9/2 | 68/66/44/88/9/2 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/search_from_sweep.a` | 133/133/33/154/10/0 | 133/133/34/155/10/0 | 0/0/1/1/0/0 | Library argument ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_arrays.a` | 516/516/524/713/58/0 | 516/516/525/714/58/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_closures.a` | 411/411/634/705/39/0 | 411/411/635/706/39/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_keep.a` | 290/290/385/520/18/0 | 290/290/386/521/18/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_keys.a` | 285/285/242/477/16/0 | 285/285/243/478/16/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_named_closures.a` | 95/95/155/187/28/0 | 95/95/156/188/28/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_named_more.a` | 179/179/297/378/26/0 | 179/179/298/379/26/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_named_objects.a` | 87/87/110/149/22/0 | 87/87/111/150/22/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_numbers.a` | 214/214/140/317/11/0 | 214/214/141/318/11/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/set_foreach_objects.a` | 368/368/515/586/30/0 | 368/368/516/587/30/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/statements_small_throw.a` | 5/5/9/13/4/0 | 5/5/16/20/4/0 | 0/0/7/7/0/0 | Saved Error payloads use tagged catch storage and checked nominal field reads; both dynamic repeat calls use checked library wrappers. Seven balanced pairs are added. |
| `internal/oracle/testdata/step21_builtin_narrow_terminal.a` | new | 6/4/8/9/5/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_catch_callback.a` | new | 17/17/12/24/6/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_dynamic.a` | new | 65/65/49/100/6/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_dynamic_uncaught.a` | new | 0/0/1/2/0/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_error_subclasses.a` | new | 16/16/37/45/5/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_finally_callback.a` | new | 19/19/18/34/8/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_finally_completion.a` | new | 17/17/6/19/5/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_library_failures.a` | new | 35/35/56/82/4/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_library_host.a` | new | 25/25/31/48/7/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_library_types.a` | new | 22/22/36/51/9/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_liveness.a` | new | 12/12/9/17/5/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_object_uncaught.a` | new | 3/3/2/4/3/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_region_payload.a` | new | 3/3/3/5/3/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_rethrow.a` | new | 13/13/22/32/4/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_saved_error.a` | new | 20/20/37/47/7/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/step21_soundness_terminal.a` | new | 3/2/3/5/3/0 | new | New fixture for the exception ruling; terminal fixtures stop with their live values. |
| `internal/oracle/testdata/string_too_long.a` | 0/0/0/0/0/0 | 1/1/3/3/1/0 | 1/1/3/3/1/0 | Nominal library error construction and cleanup on the now catchable/exit-1 path. |
| `internal/oracle/testdata/string_views_throw.a` | 55/55/62/76/14/0 | 55/55/65/79/14/0 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/taste_comma.a` | 51/51/5/54/4/0 | 51/51/6/55/4/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/taste_void.a` | 39/39/8/44/4/0 | 39/39/9/45/4/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/typed_arrays_order.a` | 34/34/12/50/8/0 | 34/34/12/51/8/0 | 0/0/0/1/0/0 | Validation wrapper releases its argument/result temporaries, including absent results; allocations unchanged. |
| `internal/oracle/testdata/typed_arrays_stats.a` | 54/54/14/61/10/0 | 54/54/14/66/10/0 | 0/0/0/5/0/0 | Validation wrapper releases its argument/result temporaries, including absent results; allocations unchanged. |
| `internal/oracle/testdata/typed_arrays_views.a` | 83/83/40/111/13/0 | 83/83/40/115/13/0 | 0/0/0/4/0/0 | Validation wrapper releases its argument/result temporaries, including absent results; allocations unchanged. |
| `internal/oracle/testdata/typed_arrays_workers_stats.a` | 513/513/82/584/23/0 | 513/513/82/600/23/0 | 0/0/0/16/0/0 | Validation wrapper releases its argument/result temporaries, including absent results; allocations unchanged. |
| `internal/oracle/testdata/unknown_narrowing.a` | 31/31/42/63/3/0 | 31/31/43/64/3/0 | 0/0/1/1/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/user_iterators.a` | 663/663/455/904/64/0 | 663/663/466/915/64/0 | 0/0/11/11/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |
| `internal/oracle/testdata/walk.a` | 222/222/117/288/36/0 | 222/222/131/302/36/0 | 0/0/14/14/0/0 | Library argument ownership; balanced retain/release change. |
| `stage3/fixtures/taste/24_proportional_conditions.a` | 43/43/89/131/14/0 | 43/43/92/134/14/0 | 0/0/3/3/0/0 | Tagged catch and throwing-call ownership; balanced retain/release change. |

# Catch values, step 1 counts

Base: `390af985caf2a81edb239cdf5e3238add50a7fcd`. Linux, clang 20.1.8.
A/F/R/L/P/G are allocations, frees, retains, releases, peak live, and regions.

| Fixture | Before A/F/R/L/P/G | After A/F/R/L/P/G | Cause |
|---|---|---|---|
| `internal/oracle/testdata/route_targets_try_loop.a` | 68/66/41/85/9/2 | 68/66/44/88/9/2 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_keys.a` | 391/391/363/674/25/0 | 391/391/364/675/25/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_keys.a` | 285/285/242/477/16/0 | 285/285/243/478/16/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_keep.a` | 516/516/625/931/35/0 | 516/516/626/932/35/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_keep.a` | 290/290/385/520/18/0 | 290/290/386/521/18/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_objects.a` | 732/732/834/1076/63/0 | 732/732/835/1077/63/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_objects.a` | 368/368/515/586/30/0 | 368/368/516/587/30/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_closures.a` | 816/816/1044/1277/87/0 | 816/816/1045/1278/87/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_closures.a` | 411/411/634/705/39/0 | 411/411/635/706/39/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_numbers.a` | 373/373/155/478/14/0 | 373/373/156/479/14/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_numbers.a` | 214/214/140/317/11/0 | 214/214/141/318/11/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_arrays.a` | 995/995/932/1355/95/0 | 995/995/933/1356/95/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_arrays.a` | 516/516/524/713/58/0 | 516/516/525/714/58/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_named_more.a` | 395/395/569/819/36/0 | 395/395/570/820/36/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_named_more.a` | 179/179/297/378/26/0 | 179/179/298/379/26/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_named_objects.a` | 165/165/207/278/31/0 | 165/165/208/279/31/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_named_objects.a` | 87/87/110/149/22/0 | 87/87/111/150/22/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_named_closures.a` | 117/117/179/219/29/0 | 117/117/180/220/29/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/set_foreach_named_closures.a` | 95/95/155/187/28/0 | 95/95/156/188/28/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_fnexpr.a` | 44/44/64/100/15/0 | 44/44/65/101/15/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/map_foreach_named_numbers.a` | 117/117/73/177/12/0 | 117/117/74/178/12/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/borrow_chain_coverage_finally.a` | 13/13/9/13/6/0 | 13/13/12/16/6/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/library_array_with.a` | 68/68/43/109/10/0 | 68/68/63/129/10/0 | Array.with generated errors are inspected through a truthful Error guard. |
| `internal/oracle/testdata/library_array_flat_map.a` | 50/50/60/109/17/0 | 50/50/61/110/17/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/library_function_expressions.a` | 56/56/49/96/19/0 | 56/56/50/97/19/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/fresh_writes.a` | 330/330/369/485/132/0 | 330/330/370/486/132/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/exceptions.a` | 133/133/154/229/20/0 | 133/133/177/252/20/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/exceptions_empty.a` | 2/1/6/4/2/0 | 2/1/8/6/2/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/closures_throw.a` | 418/418/1566/1801/142/0 | 418/418/1585/1820/142/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/finally_leaves.a` | 157/157/59/174/10/0 | 157/157/64/179/10/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/param_assigned_in_try.a` | 46/46/29/64/12/0 | 46/46/31/66/12/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/named_function_values.a` | 116/116/127/222/25/0 | 116/116/128/223/25/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/class_as_interface.a` | 372/372/303/468/60/0 | 372/372/304/469/60/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/regions_throw.a` | 124/95/9/99/31/29 | 124/95/15/105/31/29 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/borrow_loop.a` | 128/126/110/181/14/2 | 128/126/116/187/14/2 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/borrow_loop_calls.a` | 263/261/176/342/25/2 | 263/261/179/345/25/2 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/borrow_element_throw.a` | 38/38/30/50/8/0 | 38/38/36/56/8/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/move_throw.a` | 42/42/22/50/9/0 | 42/42/24/52/9/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/library_map_set_group_by.a` | 115/115/117/201/36/0 | 115/115/118/202/36/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/generic_method_return.a` | 23/15/28/36/5/6 | 23/15/30/38/5/6 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/reuse_throw.a` | 26/26/24/41/8/0 | 26/26/27/44/8/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/catch_values.a` | new | 17/17/16/31/4/0 | New thrown-value and undefined-finally oracle fixture. |
| `internal/oracle/testdata/class_features_static.a` | 110/110/55/162/28/0 | 110/110/58/165/28/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/class_features_static_private.a` | 42/42/53/94/12/0 | 42/42/73/114/12/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/class_features_accessors.a` | 67/67/47/104/21/0 | 67/67/49/106/21/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/class_inheritance_exceptions.a` | 29/29/19/39/8/0 | 29/29/23/43/8/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/class_inheritance_conditional.a` | 164/164/146/249/35/0 | 164/164/157/260/35/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/class_instance_key_throw.a` | 6/5/5/10/5/1 | 6/5/7/12/5/1 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/devirtualize.a` | 43/39/22/62/10/4 | 43/39/24/64/10/4 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/fallthrough_exceptions.a` | 26/26/10/29/5/0 | 26/26/12/31/5/0 | The catch now checks Error identity and retains its narrowed value during the read. |
| `internal/oracle/testdata/user_iterators.a` | 663/663/455/904/64/0 | 663/663/466/915/64/0 | The catch now checks Error identity and retains its narrowed value during the read. |

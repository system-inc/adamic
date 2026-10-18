Every regenerated row is attributed in counts-attribution.json. A/F/R/L/P/G denote allocations, frees, retains, releases, peak and regions.

| Fixture | Owner | Before | After | Cause |
| --- | --- | --- | --- | --- |
| internal/oracle/testdata/regexp.a | generators-main | 437/437/382/417/58/0 | 437/437/383/418/58/0 | iterator result value reads own boxed field values; respectively one or four retain/release pairs |
| internal/oracle/testdata/notyet_element_access/nodearray.a | iteration-main | added | 0/0/0/0/0/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| internal/oracle/testdata/notyet_element_access/template.a | iteration-main | added | 0/0/0/0/0/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| internal/oracle/testdata/notyet_element_access/sorted.a | iteration-main | added | 0/0/0/0/0/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| internal/oracle/testdata/generators/basic.a | generators-main | added | 30/30/85/101/14/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/cancellation.a | generators-main | added | 23/23/65/74/14/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/throw.a | generators-main | added | 16/16/75/74/12/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/defaults.a | generators-main | added | 29/29/104/107/21/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/close.a | generators-main | added | 27/27/57/59/13/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/delegate.a | generators-main | added | 89/89/316/332/51/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/delegate_throw.a | generators-main | added | 28/28/158/151/22/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/reentrant.a | generators-main | added | 14/14/53/54/11/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/owned.a | generators-main | added | 32/32/132/130/21/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/captures.a | generators-main | added | 31/31/95/95/24/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/default_error.a | generators-main | added | 13/13/31/30/12/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/map_iterator.a | generators-main | added | 24/24/95/103/15/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/checker.a | generators-main | added | 20/20/104/93/16/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/first_reference.a | generators-main | added | 17/17/73/72/12/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generators/completions.a | generators-main | added | 59/59/166/175/38/0 | new owned synchronous generator frame fixture |
| internal/oracle/testdata/generic_body_read.a | generic-body-relations | added | 10/10/0/10/3/0 | new dependent read and indexed mapper witness |
| internal/oracle/testdata/generic_body_indexed.a | generic-body-relations | added | 2/2/0/2/2/0 | new dependent read and indexed mapper witness |
| stage3/fixtures/iteration-dispatch/class.a | iteration-main | added | 22/22/11/22/7/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/object.a | iteration-main | added | 26/26/14/28/11/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/break.a | iteration-main | added | 20/20/11/18/10/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/return.a | iteration-main | added | 21/21/11/18/13/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/labeled_continue.a | iteration-main | added | 44/44/22/38/12/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/body_throw.a | iteration-main | added | 22/22/17/24/12/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/next_throw.a | iteration-main | added | 12/12/13/16/7/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/close_throw_body.a | iteration-main | added | 22/22/20/26/12/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/close_throw_break.a | iteration-main | added | 20/20/16/21/10/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/cached_next.a | iteration-main | added | 29/29/17/34/12/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/completion.a | iteration-main | added | 20/20/10/20/11/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/done_throw.a | iteration-main | added | 13/13/14/16/10/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/value_throw.a | iteration-main | added | 13/13/14/16/10/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration-dispatch/method_getters.a | iteration-main | added | 26/26/16/24/15/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a | iteration-main | added | 57/57/31/49/9/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a | iteration-main | added | 28/28/16/20/8/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| internal/oracle/testdata/user_iterators.a | iteration-main | 663/663/455/904/64/0 | 1050/1050/677/994/97/0 | runtime protocol dispatch retains cached method closures, represented step fields and close lookup |
| internal/oracle/testdata/user_iterators_rest_tdz.a | iteration-main | 5/0/5/3/5/0 | 11/3/7/3/8/0 | runtime protocol dispatch retains cached method closures, represented step fields and close lookup |
| stage3/project-references-source/fixture/two/app/main.a | project-references-main | added | 2/2/0/2/2/0 | new executable project reference fixture |
| stage3/project-references-source/fixture/chain/app/main-print.a | project-references-main | added | 2/2/0/2/2/0 | new executable project reference fixture |
| internal/fresh/testdata/regexp_tree.ts | generators-main | 91/91/81/85/35/0 | 91/91/85/89/35/0 | iterator result value reads own boxed field values; respectively one or four retain/release pairs |
| stage3/fixtures/generics/01_identity.a | generics-scout-main | added | 3/3/3/7/3/0 | new generic fixture registration |
| stage3/fixtures/generics/02_optional_return.a | generics-scout-main | added | 5/5/3/9/3/0 | new generic fixture registration |
| stage3/fixtures/generics/03_callback_return.a | generics-scout-main | added | 8/8/3/11/3/0 | new generic fixture registration |
| stage3/fixtures/generics/05_nested_class.a | generics-scout-main | added | 8/8/5/16/7/0 | new generic fixture registration |
| stage3/fixtures/generics/09_recursive_optional.a | generics-scout-main | added | 3/3/1/4/3/0 | new generic fixture registration |
| stage3/fixtures/generics/10_identifier_multimap.a | generics-scout-main | added | 11/11/27/36/9/0 | new generic fixture registration |
| stage3/fixtures/generics/14_optional_literal_union.a | generics-scout-main | added | 3/3/0/3/3/0 | new generic fixture registration |
| stage3/fixtures/generics/15_array_callback.a | generics-scout-main | added | 18/18/7/27/4/0 | new generic fixture registration |
| stage3/fixtures/iteration/collections.a | iteration-main | added | 35/35/24/47/11/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration/strings.a | iteration-main | added | 20/20/14/27/9/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration/object_iteration.a | iteration-main | added | 5/5/1/6/3/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration/array_view_stress.a | iteration-main | added | 34/34/50/72/16/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration/test262_array_views.a | iteration-main | added | 11/11/25/38/11/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration/array_view_weak.a | iteration-main | added | 5/5/8/13/4/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration/user_forwarding.a | iteration-main | added | 22/22/7/17/14/0 | runtime protocol dispatch and saved methods with owned closures; new view/dispatch witness |
| stage3/fixtures/iteration/generator.a | generators-main | added | 14/14/33/35/13/0 | new owned synchronous generator frame fixture |
| stage3/fixtures/iteration/delegated_generator.a | generators-main | added | 20/20/70/74/15/0 | new owned synchronous generator frame fixture |

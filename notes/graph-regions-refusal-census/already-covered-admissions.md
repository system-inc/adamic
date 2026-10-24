# Recorded file-backed refusals already handled in batch 8

488 programs compared on exact main and batch 8; 56 become compiled code, all previously cycle-capable Refused, none NotYet. These records already use the graph-admission checks in 749e668e and require no further edit. Each proof compares native ASan/UBSan and JavaScript with Node and checks leaks. Full diagnostic and Node bytes are in already-covered-admissions.json.

| Program (under internal/oracle/testdata/) | Proof test | Node stdout |
| --- | --- | --- |
| fresh_refused/alias_fill.a | TestFreshClobberAliasFill | `100\n` |
| fresh_refused/alias_mapset.a | TestFreshClobberAliasMapset | `200\n` |
| fresh_refused/alias_reverse.a | TestFreshClobberAliasReverse | `100\n` |
| fresh_refused/alias_setadd.a | TestFreshClobberAliasSetadd | `200\n` |
| fresh_refused/alias_sort.a | TestFreshClobberAliasSort | `100\n` |
| fresh_refused/another_view.a | TestFreshWriteProbesStayRefused/another_view.a | `2\n` |
| fresh_refused/array_from.a | TestFreshWriteProbesStayRefused/array_from.a | `1\n` |
| fresh_refused/call_clobbers.a | TestFreshWriteProbesStayRefused/call_clobbers.a | `1\n` |
| fresh_refused/call_comparator.a | TestFreshWriteProbesStayRefused/call_comparator.a | `2\n` |
| fresh_refused/call_escaped_parent.a | TestFreshWriteProbesStayRefused/call_escaped_parent.a | `1\n` |
| fresh_refused/call_grandchild.a | TestFreshWriteProbesStayRefused/call_grandchild.a | `1\n` |
| fresh_refused/call_leaks_parent.a | TestFreshWriteProbesStayRefused/call_leaks_parent.a | `1\n` |
| fresh_refused/call_one_bad_site.a | TestFreshWriteProbesStayRefused/call_one_bad_site.a | `2\n` |
| fresh_refused/call_owner.a | TestFreshWriteProbesStayRefused/call_owner.a | `1\n` |
| fresh_refused/call_returns_argument.a | TestFreshWriteProbesStayRefused/call_returns_argument.a | `1\n` |
| fresh_refused/call_same.a | TestFreshWriteProbesStayRefused/call_same.a | `1\n` |
| fresh_refused/call_targets_virtual.a | TestFreshWriteProbesStayRefused/call_targets_virtual.a | `1\n` |
| fresh_refused/call_then_throws.a | TestFreshWriteProbesStayRefused/call_then_throws.a | `10\n` |
| fresh_refused/call_twice_deferred.a | TestFreshWriteProbesStayRefused/call_twice_deferred.a | `2\n` |
| fresh_refused/call_wrapper.a | TestFreshWriteProbesStayRefused/call_wrapper.a | `1\n` |
| fresh_refused/captured_before.a | TestFreshWriteProbesStayRefused/captured_before.a | `2\n` |
| fresh_refused/caught.a | TestFreshWriteProbesStayRefused/caught.a | `1\n` |
| fresh_refused/clobber_link_loop.a | TestFreshClobberClobberLinkLoop | `100\n` |
| fresh_refused/closure_captures.a | TestFreshWriteProbesStayRefused/closure_captures.a | `1\n` |
| fresh_refused/concat.a | TestFreshWriteProbesStayRefused/concat.a | `1\n` |
| fresh_refused/constructor_tie.a | TestFreshWriteProbesStayRefused/constructor_tie.a | `2\n` |
| fresh_refused/ctor_readonly.a | TestFreshWriteProbesStayRefused/ctor_readonly.a | `total 6\n` |
| fresh_refused/ctor_set.a | TestFreshWriteProbesStayRefused/ctor_set.a | `2\n` |
| fresh_refused/ctor_wrapped.a | TestFreshWriteProbesStayRefused/ctor_wrapped.a | `2\n` |
| fresh_refused/devirt_fresh_method_keeps_argument.a | TestFreshWriteProbesStayRefused/devirt_fresh_method_keeps_argument.a | `1\n` |
| fresh_refused/escaped_before.a | TestFreshClobberEscapedBefore | `100\n` |
| fresh_refused/escapes_after_return.a | TestFreshWriteProbesStayRefused/escapes_after_return.a | `1\n` |
| fresh_refused/fill.a | TestFreshWriteProbesStayRefused/fill.a | `1\n` |
| fresh_refused/global_then_function.a | TestFreshWriteProbesStayRefused/global_then_function.a | `1\n` |
| fresh_refused/index_write.a | TestFreshWriteProbesStayRefused/index_write.a | `1\n` |
| fresh_refused/map_callback.a | TestFreshWriteProbesStayRefused/map_callback.a | `1\n` |
| fresh_refused/map_entries_pattern.a | TestFreshClobberMapEntriesPattern | `100\n` |
| fresh_refused/map_iterator.a | TestFreshWriteProbesStayRefused/map_iterator.a | `1\n` |
| fresh_refused/map_iterator_parenthesized.a | TestFreshWriteProbesStayRefused/map_iterator_parenthesized.a | `1\n` |
| fresh_refused/map_key.a | TestFreshWriteProbesStayRefused/map_key.a | `1\n` |
| fresh_refused/map_set.a | TestFreshWriteProbesStayRefused/map_set.a | `1\n` |
| fresh_refused/next_iteration.a | TestFreshWriteProbesStayRefused/next_iteration.a | `2\n` |
| fresh_refused/older_instance.a | TestFreshWriteProbesStayRefused/older_instance.a | `0\n` |
| fresh_refused/one_call.a | TestFreshWriteProbesStayRefused/one_call.a | `1\n` |
| fresh_refused/returns_argument.a | TestFreshWriteProbesStayRefused/returns_argument.a | `1\n` |
| fresh_refused/returns_global.a | TestFreshWriteProbesStayRefused/returns_global.a | `1\n` |
| fresh_refused/set_add.a | TestFreshWriteProbesStayRefused/set_add.a | `1\n` |
| fresh_refused/set_iterator.a | TestFreshWriteProbesStayRefused/set_iterator.a | `1\n` |
| fresh_refused/splice.a | TestFreshWriteProbesStayRefused/splice.a | `1\n` |
| fresh_refused/spread_array.a | TestFreshWriteProbesStayRefused/spread_array.a | `1\n` |
| fresh_refused/spread_old.a | TestFreshWriteProbesStayRefused/spread_old.a | `1\n` |
| fresh_refused/spread_outside.a | TestFreshClobberSpreadOutside | `1 1\n` |
| fresh_refused/throw_keeps_old.a | TestFreshClobberThrowKeepsOld | `caught\n1\n` |
| fresh_refused/wraps_argument.a | TestFreshWriteProbesStayRefused/wraps_argument.a | `1\n` |
| memory_examples/refused/closure.a | TestMemoryExampleRefusedClosure | `3\n` |
| memory_examples/refused/tree.a | TestMemoryExampleRefusedTree | `6\n` |

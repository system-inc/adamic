# Final subset counts audit

Compared with pushed prefix 1eaed07b7. Values: allocations / frees / retains / releases / peak / regions. Generated on Linux by TestCountsAreRecorded, not edited by hand. The final regeneration passes in 80.222 seconds. An earlier regeneration (142.959 seconds) exposed the hoisted namespace-var representation guard: that guard was restored to terminal exit 70 and the final regeneration leaves its existing row unchanged. No unexplained row remains.

## 6 existing rows

Lexical read: checkReadyRead/readinessError now allocates one ReferenceError, retains its immortal name and message, and releases the pending Error at uncaught exit; this replaces a terminal panic which allocated nothing.

| Fixture | Old | New |
| --- | --- | --- |
| internal/oracle/testdata/dead_zone.a | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 1 / 2 / 1 / 1 / 0 |
| stage3/drivers/scanner/probes/cyclic-premature-value/main.a | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 1 / 2 / 1 / 1 / 0 |
| internal/oracle/testdata/method_coverage_empty_alias_tdz.a | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 1 / 2 / 1 / 1 / 0 |
| internal/oracle/testdata/library_method_values_dead_zone.a | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 1 / 2 / 1 / 1 / 0 |
| internal/oracle/testdata/module_namespace_reads/early.a | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 1 / 2 / 1 / 1 / 0 |
| internal/oracle/testdata/module_namespace_reads/direct_early.a | 0 / 0 / 0 / 0 / 0 / 0 | 1 / 1 / 2 / 1 / 1 / 0 |

## 1 existing rows

Iterator rest binding lexical read: ReferenceError construction replaces terminal panic; the newly modelled throw edge cleans up the live rest/iterator owners and releases the pending Error.

| Fixture | Old | New |
| --- | --- | --- |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 11 / 3 / 7 / 3 / 8 / 0 | 12 / 12 / 9 / 10 / 9 / 0 |

## 5 existing rows

Namespace TypeError: namespaceReadyReads now throws a classed Error through ir.Throw. The name and message are retained, and exception-path ownership cleanup frees the callable alias and thrown Error instead of abandoning them in adamic_panic.

| Fixture | Old | New |
| --- | --- | --- |
| internal/oracle/testdata/namespaces_unknown_before.a | 2 / 1 / 2 / 3 / 2 / 0 | 3 / 3 / 5 / 8 / 2 / 0 |
| internal/oracle/testdata/namespaces_unknown_function_before.a | 2 / 1 / 2 / 3 / 2 / 0 | 3 / 3 / 5 / 8 / 2 / 0 |
| internal/oracle/testdata/namespaces_unknown_write_before.a | 2 / 1 / 2 / 3 / 2 / 0 | 3 / 3 / 5 / 8 / 2 / 0 |
| internal/oracle/testdata/namespaces_unknown_void_before.a | 2 / 1 / 2 / 3 / 2 / 0 | 3 / 3 / 5 / 8 / 2 / 0 |
| internal/oracle/testdata/namespaces_unknown_enum_before.a | 2 / 1 / 2 / 3 / 2 / 0 | 3 / 3 / 5 / 8 / 2 / 0 |

## 3 existing rows

Captured lexical read or write: checkReadyRead now constructs ReferenceError. flow.CanThrow and readinessExceptions propagate the effect, so closure/cell owners are released on the exception path; the pending Error is released at uncaught exit.

| Fixture | Old | New |
| --- | --- | --- |
| internal/oracle/testdata/nested_tdz.a | 2 / 0 / 1 / 0 / 2 / 0 | 3 / 3 / 3 / 3 / 3 / 0 |
| internal/oracle/testdata/nested_tdz_write.a | 2 / 0 / 1 / 0 / 2 / 0 | 3 / 3 / 3 / 3 / 3 / 0 |
| internal/oracle/testdata/nested_destructured_tdz.a | 2 / 0 / 1 / 0 / 2 / 0 | 3 / 3 / 3 / 3 / 3 / 0 |

## 4 existing rows

Switch lexical read or write: checkReadyRead constructs ReferenceError and checkThrown unwinds scope owners. flow.CanThrow keeps cleanup valid on the newly catchable read/write edge; adamic_uncaught releases the pending Error.

| Fixture | Old | New |
| --- | --- | --- |
| internal/oracle/testdata/switch_case_declarations/dead_zone.a | 8 / 5 / 7 / 8 / 5 / 0 | 9 / 9 / 9 / 13 / 5 / 0 |
| internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a | 3 / 3 / 3 / 5 / 2 / 0 | 4 / 4 / 5 / 7 / 2 / 0 |
| internal/oracle/testdata/switch_case_declarations/dead_zone_initializer.a | 3 / 0 / 3 / 1 / 3 / 0 | 4 / 4 / 5 / 6 / 4 / 0 |
| internal/oracle/testdata/switch_case_declarations/dead_zone_write.a | 5 / 1 / 3 / 2 / 5 / 0 | 6 / 6 / 5 / 7 / 5 / 0 |

## Added rows

| Fixture | Counts | Cause |
| --- | --- | --- |
| internal/oracle/testdata/lowering_chain_namespace_catch.a | 6 / 6 / 9 / 15 / 3 / 0 | New caught namespace/TDZ witness. |
| internal/oracle/testdata/lowering_chain_tdz_catch.a | 4 / 4 / 6 / 8 / 2 / 0 | New caught namespace/TDZ witness. |
| internal/oracle/testdata/statements_small_stopped/structural_error.a | 2 / 2 / 5 / 5 / 2 / 0 | Step 21 already admits structural Error throws; replace the stale stop with Node agreement. |
| stage3/fixtures/iteration/generator.a | 14 / 14 / 33 / 35 / 13 / 0 | The generators member already admits this fixture; replace the stale outcome and register ordinary Node agreement. |
| stage3/fixtures/iteration/delegated_generator.a | 20 / 20 / 70 / 74 / 15 / 0 | The generators member already admits this fixture; replace the stale outcome and register ordinary Node agreement. |

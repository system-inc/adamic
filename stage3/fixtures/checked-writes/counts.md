# Counts

Local stage-3 corpus, not registered in internal/oracle.

| Family | Class-d sites | Node fixtures passing | Input mutants caught | .ts runtime contracts passing |
|---|---:|---:|---:|---:|
| DiagnosticWithLocation | 144 | 2 | 1 | 2 |
| never[] | 82 | 2 | 1 | 0 |
| FlowNode | 26 | 2 | 1 | 0 |
| ResolvedType | 17 | 2 | 1 | 0 |
| T | 17 | 2 | 1 | 2 |
| NodeBuilderContext | 11 | 2 | 1 | 0 |
| DiagnosticWithLocation[] | 9 | 2 | 1 | 0 |
| GeneratedIdentifier | 8 | 2 | 1 | 0 |
| Identifier | 8 | 2 | 1 | 2 |
| Expression | 7 | 2 | 1 | 2 |
| Node | 7 | 2 | 1 | 2 |
| CapturedThis | 6 | 2 | 1 | 0 |
| Declaration[] | 6 | 2 | 1 | 0 |
| NodeArray<Statement> | 6 | 2 | 1 | 0 |
| DiagnosticWithDetachedLocation | 5 | 2 | 1 | 0 |
| FlowArrayMutation \| FlowAssignment | 5 | 2 | 1 | 0 |
| GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node | 5 | 2 | 1 | 0 |
| SyntheticSuper | 5 | 2 | 1 | 0 |
| Mutable<GeneratedIdentifier> | 4 | 2 | 1 | 0 |
| Diagnostic | 3 | 2 | 1 | 2 |
| **Total** | **381 / 515** | **40** | **20** | **12 / 40** |

Allocation rows below are release counted builds; panic rows stop at exit 70.
Finished positives also pass ASan/UBSan and leak detection.
No internal/oracle registry or counts rows were added by this unit.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| 01_located-diagnostic_in.a | 3 | 3 | 8 | 12 | 3 | 0 |
| 01_located-diagnostic_out.a | 2 | 0 | 3 | 2 | 2 | 0 |
| 05_generic-range_in.a | 3 | 3 | 2 | 6 | 3 | 0 |
| 05_generic-range_out.a | 2 | 0 | 1 | 1 | 2 | 0 |
| 09_identifier-flags_in.a | 3 | 3 | 2 | 5 | 3 | 0 |
| 09_identifier-flags_out.a | 2 | 0 | 2 | 1 | 2 | 0 |
| 10_expression-range_in.a | 3 | 3 | 3 | 6 | 3 | 0 |
| 10_expression-range_out.a | 2 | 0 | 2 | 1 | 2 | 0 |
| 11_node-flags_in.a | 3 | 3 | 1 | 5 | 3 | 0 |
| 11_node-flags_out.a | 2 | 0 | 1 | 1 | 2 | 0 |
| 20_diagnostic-union_in.a | 3 | 3 | 8 | 12 | 3 | 0 |
| 20_diagnostic-union_out.a | 2 | 0 | 3 | 2 | 2 | 0 |

Uncompiled compatibility inputs have no native allocation row.

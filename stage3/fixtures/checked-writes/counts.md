# Counts

Local stage-3 corpus, not registered in internal/oracle.

| Family | Class-d sites | Node fixtures passing | Input mutants caught | .ts runtime contracts passing |
|---|---:|---:|---:|---:|
| DiagnosticWithLocation | 144 | 2 | 1 | 2 |
| never[] | 82 | 2 | 1 | 2 |
| FlowNode | 26 | 2 | 1 | 2 |
| ResolvedType | 17 | 2 | 1 | 2 |
| T | 17 | 2 | 1 | 2 |
| NodeBuilderContext | 11 | 2 | 1 | 2 |
| DiagnosticWithLocation[] | 9 | 2 | 1 | 0 |
| GeneratedIdentifier | 8 | 2 | 1 | 2 |
| Identifier | 8 | 2 | 1 | 2 |
| Expression | 7 | 2 | 1 | 2 |
| Node | 7 | 2 | 1 | 2 |
| CapturedThis | 6 | 2 | 1 | 2 |
| Declaration[] | 6 | 2 | 1 | 2 |
| NodeArray<Statement> | 6 | 2 | 1 | 0 |
| DiagnosticWithDetachedLocation | 5 | 2 | 1 | 2 |
| FlowArrayMutation \| FlowAssignment | 5 | 2 | 1 | 2 |
| GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node | 5 | 2 | 1 | 2 |
| SyntheticSuper | 5 | 2 | 1 | 2 |
| Mutable<GeneratedIdentifier> | 4 | 2 | 1 | 2 |
| Diagnostic | 3 | 2 | 1 | 2 |
| **Total** | **381 / 515** | **40** | **20** | **36 / 40** |

Allocation rows below are release counted builds; panic rows stop at exit 70.
Finished positives also pass ASan/UBSan and leak detection.
No internal/oracle registry or counts rows were added by this unit.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| 01_located-diagnostic_in.a | 3 | 3 | 8 | 12 | 3 | 0 |
| 01_located-diagnostic_out.a | 2 | 0 | 3 | 2 | 2 | 0 |
| 02_shared-empty_in.a | 3 | 3 | 0 | 4 | 3 | 0 |
| 02_shared-empty_out.a | 2 | 0 | 1 | 1 | 2 | 0 |
| 03_flow-node_in.a | 4 | 4 | 5 | 9 | 4 | 0 |
| 03_flow-node_out.a | 3 | 0 | 4 | 3 | 3 | 0 |
| 04_resolved-members_in.a | 3 | 3 | 4 | 7 | 3 | 0 |
| 04_resolved-members_out.a | 2 | 0 | 3 | 2 | 2 | 0 |
| 05_generic-range_in.a | 3 | 3 | 2 | 6 | 3 | 0 |
| 05_generic-range_out.a | 2 | 0 | 1 | 1 | 2 | 0 |
| 06_builder-tracker_in.a | 7 | 7 | 3 | 9 | 7 | 0 |
| 06_builder-tracker_out.a | 6 | 0 | 2 | 3 | 6 | 0 |
| 08_generated-identifier_in.a | 4 | 4 | 3 | 7 | 4 | 0 |
| 08_generated-identifier_out.a | 3 | 0 | 2 | 2 | 3 | 0 |
| 09_identifier-flags_in.a | 3 | 3 | 2 | 5 | 3 | 0 |
| 09_identifier-flags_out.a | 2 | 0 | 2 | 1 | 2 | 0 |
| 10_expression-range_in.a | 3 | 3 | 3 | 6 | 3 | 0 |
| 10_expression-range_out.a | 2 | 0 | 2 | 1 | 2 | 0 |
| 11_node-flags_in.a | 3 | 3 | 1 | 5 | 3 | 0 |
| 11_node-flags_out.a | 2 | 0 | 1 | 1 | 2 | 0 |
| 12_captured-this_in.a | 4 | 4 | 4 | 7 | 4 | 0 |
| 12_captured-this_out.a | 3 | 0 | 3 | 2 | 3 | 0 |
| 13_declaration-array_in.a | 4 | 4 | 6 | 10 | 4 | 0 |
| 13_declaration-array_out.a | 3 | 0 | 3 | 3 | 3 | 0 |
| 15_detached-diagnostic_in.a | 3 | 3 | 5 | 9 | 3 | 0 |
| 15_detached-diagnostic_out.a | 2 | 0 | 2 | 2 | 2 | 0 |
| 16_flow-assignment-union_in.a | 4 | 4 | 5 | 9 | 4 | 0 |
| 16_flow-assignment-union_out.a | 3 | 0 | 4 | 3 | 3 | 0 |
| 17_generated-node-union_in.a | 4 | 4 | 3 | 7 | 4 | 0 |
| 17_generated-node-union_out.a | 3 | 0 | 2 | 2 | 3 | 0 |
| 18_synthetic-super_in.a | 4 | 4 | 4 | 7 | 4 | 0 |
| 18_synthetic-super_out.a | 3 | 0 | 3 | 2 | 3 | 0 |
| 19_mutable-generated_in.a | 4 | 4 | 3 | 7 | 4 | 0 |
| 19_mutable-generated_out.a | 3 | 0 | 2 | 2 | 3 | 0 |
| 20_diagnostic-union_in.a | 3 | 3 | 8 | 12 | 3 | 0 |
| 20_diagnostic-union_out.a | 2 | 0 | 3 | 2 | 2 | 0 |

Uncompiled compatibility inputs have no native allocation row.

Current compiler head: `29275d4ee7c8269a10f12fa0b4aa5732f0d94f9d`. All 36 compiled inputs pass the strict contract in native sanitized, native counted and emitted JavaScript. Four inputs remain uncompiled (unshift and NodeArray intersection pairs).

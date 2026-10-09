| ID | origin/main file:line | Change | Failed rows in bounded matrix |
|---|---|---|---|
| M01 | internal/load/load.go:147 | `len(diagnostics) > 0` to `len(diagnostics) > 1` | TestNestedRebindingCheckerRefusal |
| M02 | internal/load/load.go:233 | `all = append(all, p.compiler.GetSemanticDiagnostics(ctx, nil)...)` to `// Semantic diagnostic collection dropped.` | TestNestedRebindingCheckerRefusal |
| M03 | internal/load/load.go:255 | `diagnostic.Code(), message` to `diagnostic.Code()+1, message` | TestNestedRebindingCheckerRefusal |
| M04 | internal/lower/cycles.go:105 | `!declared.Captured \|\| declared.Global` to `declared.Captured \|\| declared.Global` | TestNestedCycleRefusal |
| M05 | internal/lower/cycles.go:363 | `if node == target { 				return true` to `if node == target { 				return false` | TestNestedCycleRefusal |
| M06 | internal/lower/cycles.go:112 | `&& !l.closedFrameInput(local)` to `&& l.closedFrameInput(local)` | TestNestedCycleRefusal |
| M07 | internal/lower/typed_arrays.go:59 | `"typed array element type "+name` to `"typed array facility "+name` | TestNewExpressionUint16RemainsNotYet |
| M08 | internal/lower/typed_arrays.go:43 | `if node.Kind == ast.KindNewExpression {` to `if node.Kind == ast.KindCallExpression {` | TestNestedCallbackCarrierMutantIsCaught, TestNestedCycleRefusal, TestNestedSiblingCycleMutantIsCaught, TestNewExpressionCacheCaptureRemainsNotYet, TestNewExpressionUint16RemainsNotYet (incomplete outside six isolated unit rows) |
| M09 | internal/lower/new_class_value.go:166 | `if !closed {` to `if closed {` | TestNativeAgreesWithNode, TestNewExpressionCacheCaptureRemainsNotYet |
| M10 | internal/lower/new_class_value.go:163 | `closed && !initializer.Closure` to `closed && initializer.Closure` | TestNativeAgreesWithNode |
| M11 | internal/lower/new_class_value.go:160 | `len(value.Arguments) == 0 && len(targets) > 0` to `len(value.Arguments) <= 1 && len(targets) > 0` | [] |
| M12 | internal/lower/nested_functions.go:56 | `l.function.NestedFrame = true` to `l.function.NestedFrame = false` | TestNativeAgreesWithNode, TestNestedReferenceIdentityMutant, TestScannerNestedReferenceMutants, TestScannerNestedReferences |
| M13 | internal/lower/nested_functions.go:227 | `l.result.Locals[local].Preallocated = true 		return []ir.Statement` to `l.result.Locals[local].Preallocated = false 		return []ir.Statement` | TestNativeAgreesWithNode |
| M14 | internal/lower/nested_functions.go:190 | `for _, function := range l.closures[boundary+1:] { 		l.result.Functions[function].ForwardedNestedParent = parent 	}` to `// Single-statement registration loop dropped.` | TestNativeAgreesWithNode, TestNestedCallbackCarrierMutantIsCaught |
| M15 | internal/lower/nested_functions.go:140 | `for _, index := range functions { 		l.result.Functions[index].Environment = slices.Clone(environment) 	}` to `// Single-statement registration loop dropped.` | [] |
| M16 | internal/lower/diagnostics.go:33 | `What: what` to `What: "unsupported construct"` | TestNewExpressionCacheCaptureRemainsNotYet, TestNewExpressionUint16RemainsNotYet |
| E00 | internal/lower/lower.go:20 | `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {` to `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { return &ir.Program{}, nil;` | TestNestedCallbackCarrierMutantIsCaught, TestNestedCycleRefusal, TestNestedReferenceIdentityMutant, TestNestedSiblingCycleMutantIsCaught, TestNewExpressionCacheCaptureRemainsNotYet, TestNewExpressionUint16RemainsNotYet, TestTheOracleCatchesOneByte (incomplete outside six isolated unit rows) |
| W01 | internal/oracle/oracle_test.go:712 | `return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)` to `return ""` | TestNestedSiblingCycleMutantIsCaught |
| W02 | internal/oracle/oracle_test.go:716 | `func disagreement(oracle run, native run) string {` to `func disagreement(oracle run, native run) string { return "";` | TestNestedCallbackCarrierMutantIsCaught, TestNestedReferenceIdentityMutant, TestScannerNestedReferenceMutants, TestTheOracleCatchesOneByte |

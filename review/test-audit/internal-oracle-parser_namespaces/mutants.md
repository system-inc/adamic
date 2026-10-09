| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/namespace_receiver_scope.go:13 | node.Kind != ast.KindArrowFunction -> node.Kind == ast.KindArrowFunction | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M02 | internal/lower/namespaces.go:426 | return value at entry |  |
| M03 | internal/lower/namespaces.go:502 | body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: true}}) -> body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: false}}) | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M04 | internal/lower/namespaces.go:94 | if node.Kind == ast.KindElementAccessExpression { 		return nil, true -> if node.Kind == ast.KindPropertyAccessExpression { 		return nil, true | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M05 | internal/lower/namespace_callable.go:21 | if called(node) { -> if !called(node) { | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M06 | internal/lower/lower.go:43 | if err := lowering.namespaceInitialization(modules); err != nil { 		return nil, err 	}  -> |  |
| M07 | internal/lower/predicates_proof.go:1222 | counts.Checked++  -> | TestPredicateDirectionCountsAreRecorded, TestPredicateMiscompileRefusals |
| M08 | internal/lower/predicates_proof.go:1224 | "unobservable", "no narrowed read in the " -> "proven", "no narrowed read in the " | TestPredicateDirectionCountsAreRecorded |
| M09 | internal/lower/predicates_proof.go:1211 | name := "true" -> name := "false" | TestPredicateDirectionCountsAreRecorded |
| M10 | internal/lower/predicates_proof.go:703 | implementation.Body().ForEachChild(writes) 	if changed { -> implementation.Body().ForEachChild(writes) 	if !changed { | TestPredicateDirectionCountsAreRecorded, TestPredicateMiscompileRefusals |
| M11 | internal/lower/unknown.go:45 | "use a discriminant, or a Map" -> "use a discriminant" | TestInUnionNarrowingIsRefused |
| M12 | internal/lower/unknown.go:44 | objects > 1 -> objects > 2 | TestInUnionNarrowingIsRefused |
| M13 | internal/lower/predicates.go:37 | "a type predicate whose return is not proven (" -> "a predicate whose return is not proven (" | TestPredicateMiscompileRefusals |
| M14 | internal/lower/predicates_proof.go:753 | "overload %d of %s result: predicate %s is false" -> "overload %d of %s result: predicate %s failed" | TestPredicateMiscompileRefusals |

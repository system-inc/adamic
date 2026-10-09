| ID | Origin file:line | Change | Observed failing rows |
|---|---|---|---|
| M01 | internal/lower/typed_arrays.go:94 | FromArray: true -> FromArray: false |  |
| M02 | internal/lower/typed_arrays.go:113 | return ir.TypedArrayNew{Of: kind, Source: source, FromArray: true}, true, nil -> return ir.TypedArrayNew{Of: kind, Source: source, FromArray: false}, true, nil |  |
| M03 | internal/lower/typed_arrays.go:256 | Value: value, Element: ir.Number, Site: -> Value: value, Element: ir.String, Site: | TestTypedArraysLower |
| M04 | internal/lower/typed_arrays.go:208 | if value.Type() != kind { -> if value.Type() == kind { | TestTypedArrayGaps, TestTypedArraysLower |
| M05 | internal/lower/expression.go:286 | return fromKind != 0 && fromKind == toKind -> return true | TestTypedArrayViewsCannotChangeRepresentation |
| M06 | internal/lower/unknown.go:31 | strings.ContainsRune(key.Text(), 0) -> strings.ContainsRune(key.Text(), 1) | TestUnknownReflectionRefusals |
| M07 | internal/lower/unknown.go:51 | key.Text() == "stack" -> key.Text() == "stack_disabled" | TestUnknownReflectionRefusals |
| M08 | internal/lower/enums.go:183 | member.Name().Text() == "__proto__" -> member.Name().Text() == "__proto_disabled__" | TestEnumLimitsStayLoud, TestTasteRepresentationLimitsStayExplicit |
| M09 | internal/lower/interface_cast.go:77 | if of == ir.Object { -> if of != ir.Object { | TestDefaultTaggedInterfaceAdmission, TestViewObjectWritesNeedSourceCertificate |
| M10 | internal/lower/view_contracts.go:27 | Kind: ir.ViewUndefined, Name: "undefined", Undefined: true -> Kind: ir.ViewUndefined, Name: "undefined", Undefined: false | TestMixedUnionContractPhantomVoidIsUndefined |
| M11 | internal/lower/view_contracts.go:69 | contract.Kind = ir.ViewScalar -> contract.Kind = ir.ViewObject | TestMixedUnionContractPhantomBrandUsesPrimitiveBase, TestMixedUnionContractRecursiveMember, TestViewObjectContractsAreAvailableToEraser |
| M12 | internal/lower/view_contracts.go:90 | l.result.ViewContracts[int(id)-1] = contract -> (drop) | TestDefaultTaggedInterfaceAdmission, TestMixedUnionContractGraph, TestMixedUnionContractRecursiveMember, TestPredicateBodyProof, TestViewObjectContractsAreAvailableToEraser, TestViewObjectWritesNeedSourceCertificate |
| M13 | internal/lower/view_unions_mixed.go:42 | l.result.ViewContracts = l.result.ViewContracts[:start] -> (drop) | TestMixedUnionContractFailureDoesNotCertifyRetry, TestMixedUnionContractUnknownMemberFails |
| M14 | internal/lower/view_unions_mixed.go:43 | for key, value := range l.result.ViewContractTypes { 			if int(value) > start { 				delete(l.result.ViewContractTypes, key) 			} 		} -> (drop) | TestMixedUnionContractFailureDoesNotCertifyRetry, TestMixedUnionContractUnknownMemberFails |
| M15 | internal/lower/view_unions_mixed.go:56 | l.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown -> l.result.ViewContracts[int(child)-1].Kind == ir.ViewCallable | TestMixedUnionContractUnknownMemberFails |
| M16 | internal/lower/view_unions_mixed.go:61 | l.result.ViewContracts[int(id)-1] = contract -> (drop) | TestMixedUnionContractFailureDoesNotCertifyRetry, TestMixedUnionContractGraph, TestMixedUnionContractPhantomBrandUsesPrimitiveBase, TestMixedUnionContractRecursiveMember |
| M17 | internal/lower/view_unions_primitive_brands.go:69 | return flags&checker.TypeFlagsVoid != 0 \|\| optional && flags&checker.TypeFlagsUndefined != 0 -> return flags&checker.TypeFlagsVoid == 0 \|\| optional && flags&checker.TypeFlagsUndefined != 0 | TestMixedUnionContractPhantomBrandUsesPrimitiveBase, TestMixedUnionContractPhantomVoidIsUndefined |
| M18 | internal/lower/view_unions_untagged.go:51 | if field.Optional { -> if !field.Optional { | TestUntaggedViewMemberTags |
| M19 | internal/lower/view_unions_untagged.go:45 | if contract.Kind != ir.ViewObject \|\| contract.Of != ir.Object { -> if contract.Kind == ir.ViewObject \|\| contract.Of != ir.Object { | TestUntaggedViewMemberTags, TestUntaggedViewStructuralFallback |
| M20 supplemental | internal/lower/view_unions_untagged.go:66 | append([]ir.ViewLiteral(nil), tag.Allowed...) -> append(tag.Allowed[:0], tag.Allowed...) | TestUntaggedViewMemberTags |

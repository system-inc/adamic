Fixed plan before mutant outcomes. CODE UNDER TEST is Adamic Go lowering and relation analysis; ORACLES are handwritten assertions. All reached functions listed in reached-functions.txt. Sixteen mutations spread across seven files, plus four separate entry probes. Budget limits this below the twenty-mutant aim. Census input is the existing initializer fixture copied to a strict TypeScript project.
M01 internal/lower/non_null.go:106 change constant: ".ts" -> ".a"
M02 internal/lower/non_null.go:81 change constant: value.Type() == ir.Number || value.Type() == ir.Boolean -> value.Type() == ir.MaybeNumber || value.Type() == ir.Boolean
M03 internal/lower/non_null.go:96 change constant: "non-null assertion failed at " -> ""
M04 internal/lower/non_null.go:83 return early with constant: return value, nil -> return ir.NumberConstant{Value: 1}, nil
M05 internal/lower/refusals.go:28 change constant: "use a Map, which keeps keys in the order they were added" -> "use a Map"
M06 internal/lower/optional_indexing_map.go:99 return early: return hazard -> return false
M07 internal/lower/optional_indexing.go:35 change constant: index.Type() != ir.Number -> index.Type() != ir.String
M08 internal/lower/optional_widening.go:86 flip condition: !isClassInstance(source) -> isClassInstance(source)
M09 internal/lower/optional_widening.go:81 change option: skip[property.Name] -> false
M10 internal/lower/optional_widening.go:179 change option: target = l.impliedTarget(node) -> target = nil
M11 internal/lower/optional_widening.go:196 change constant: " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)" -> " on the source type, or build a fresh object with known fields (adamic/no-optional-view)"
M12 internal/lower/generic.go:259 off-by-one bound: len(present) == 1 && given.Flags()&checker.TypeFlagsUndefined == 0 -> len(present) == 0 && given.Flags()&checker.TypeFlagsUndefined == 0
M13 internal/lower/generic.go:269 flip condition: missing != nil -> missing == nil
M14 internal/lower/generic.go:288 flip condition: if _, isSet := into[declared]; !isSet { -> if _, isSet := into[declared]; isSet {
M15 internal/lower/class_inheritance.go:140 flip condition: checkABI && (!knownA || !knownB || a != b) -> checkABI && (!knownA || !knownB || a == b)
M16 internal/lower/class_inheritance.go:98 swap two arguments: !l.classAssignable(accepts, override) -> !l.classAssignable(override, accepts)
